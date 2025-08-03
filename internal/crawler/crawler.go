package crawler

import (
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
	redisqueue "github.com/AlexSamarskii/web_crawler/internal/redis"

	"github.com/PuerkitoBio/goquery"
)

func StartWorker(id int, redisClient *redisqueue.RedisClient, db *db.Database, result chan string, domainLimiter *limiter.DomainLimiter) {
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

		crawlURL(urls, depth, redisClient, result, db)
	}
}

func fetchURL(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Ошибка запроса %s: %v", url, err)
		return nil, err
	}
	fmt.Println("Запрос:", url, "Статус:", resp.Status)
	return resp, nil
}

func crawlURL(url string, depth int, redisClient *redisqueue.RedisClient, result chan string, db *db.Database) {

	resp, err := fetchURL(url)
	if err != nil {
		result <- fmt.Sprintf("Ошибка запроса по %s", url)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result <- fmt.Sprintf("Ошибка чтения тела по %s", url)
		return
	}
	bodyString := string(bodyBytes)

	links := extractLinks(url, bodyString)
	fmt.Println("Найдено", len(links), "links on", url)

	if len(links) > 0 {
		db.StoreLinks(url, links)
		for _, link := range links {
			if !redisClient.IsVisited(link) {
				redisClient.MarkVisited(link)
				redisClient.PushToURLQueueWithDepth(link, depth+1)
			}
		}
		result <- fmt.Sprintf("Crawled %s, найдено %d ссылок", url, len(links))
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
