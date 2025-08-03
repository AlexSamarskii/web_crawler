package crawler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AlexSamarskii/web_crawler/internal/db"
	"github.com/AlexSamarskii/web_crawler/internal/limiter"
	"github.com/AlexSamarskii/web_crawler/internal/metrics"
	redisqueue "github.com/AlexSamarskii/web_crawler/internal/redis"
	"github.com/AlexSamarskii/web_crawler/internal/robots"

	"github.com/PuerkitoBio/goquery"
)

func StartWorker(id int, redisClient *redisqueue.RedisClient, db *db.Database, result chan string, domainLimiter *limiter.DomainLimiter, robotsChecker *robots.RobotsChecker) {
	for {
		log.Printf("Воркер начал искать URL %d", id)
		urls := redisClient.PopNextURL()
		if urls == "" {
			log.Printf("Не найден URL воркером %d", id)
			break
		}

		parsedURL, err := url.Parse(urls)
		if err != nil {
			log.Printf("Ошибка парсинга URL %s: %v", urls, err)
			result <- "Ошибка парсинга: " + urls
			continue
		}

		limiter := domainLimiter.GetLimiter(parsedURL.Host)

		// Ждём, пока лимитер разрешит запрос
		for !limiter.Allow() {
			time.Sleep(100 * time.Millisecond) // Проверяем каждые 100мс
		}

		depths, err := redisClient.GetClient().HGet(redisClient.GetCtx(), "url_depths", urls).Result()

		if err != nil {
			log.Printf("%d, Ошибка определения глубины для %s, 0 по умолчанию", id, urls)
		}

		depth, _ := strconv.Atoi(depths)

		crawlURL(urls, depth, redisClient, result, db, robotsChecker)
	}
}

func fetchURL(urls string, robotsChecker *robots.RobotsChecker) (*http.Response, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.URLProcessingDuration.Observe(duration)
	}()

	parsedURL, err := url.Parse(urls)
	if err != nil {
		metrics.URLsFailed.Inc()
		return nil, err
	}

	userAgent := "CrawlerBot/1.0"

	if !robotsChecker.CanFetch(userAgent, urls) {
		log.Printf("Запрещено robots.txt: %s", urls)
		return nil, errors.New("запрещено robots.txt")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		metrics.URLsFailed.Inc()
		log.Printf("Ошибка запроса %s: %v", urls, err)
		return nil, err
	}

	statusCode := strconv.Itoa(resp.StatusCode)
	metrics.HTTPStatusCodes.WithLabelValues(statusCode).Inc()

	metrics.URLsByDomain.WithLabelValues(parsedURL.Host).Inc()

	log.Println("Запрос:", urls, "Статус:", resp.Status)
	return resp, nil
}

func crawlURL(urls string, depth int, redisClient *redisqueue.RedisClient, result chan string, db *db.Database, robotsChecker *robots.RobotsChecker) {
	resp, err := fetchURL(urls, robotsChecker)
	if err != nil {
		result <- fmt.Sprintf("Ошибка запроса по %s", urls)
		return
	}
	defer resp.Body.Close()

	metrics.URLsProcessed.Inc()
	metrics.URLsByDepth.WithLabelValues(strconv.Itoa(depth)).Inc()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result <- fmt.Sprintf("Ошибка чтения тела по %s", urls)
		return
	}
	bodyString := string(bodyBytes)

	links := extractLinks(urls, bodyString)
	fmt.Println("Найдено", len(links), "links on", urls)

	metrics.LinksExtracted.Add(float64(len(links)))

	if len(links) > 0 {
		db.StoreLinks(urls, links)
		for _, link := range links {
			if !redisClient.IsVisited(link) {
				redisClient.MarkVisited(link)
				redisClient.PushToURLQueueWithDepth(link, depth+1)
			}
		}
		result <- fmt.Sprintf("Crawled %s, найдено %d ссылок", urls, len(links))
	}
}

func extractLinks(baseURL string, body string) []string {
	var links []string

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Printf("Ошибка парсинга HTML по %s: %v", baseURL, err)
		return links
	}

	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Ошибка парсинга base URL %s: %v", baseURL, err)
		return links
	}

	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, exists := item.Attr("href")
		if exists && href != "" {
			parsedHref, err := parsedBase.Parse(href)
			if err == nil {
				links = append(links, parsedHref.String())
			}
		}
	})

	return links
}
