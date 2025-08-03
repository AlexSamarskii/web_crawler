package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/AlexSamarskii/web_crawler/internal/config"
	"github.com/AlexSamarskii/web_crawler/internal/crawler"
	"github.com/AlexSamarskii/web_crawler/internal/db"
	"github.com/AlexSamarskii/web_crawler/internal/limiter"
	"github.com/AlexSamarskii/web_crawler/internal/redis"
	"github.com/AlexSamarskii/web_crawler/internal/robots"
	"github.com/spf13/cobra"
)

var (
	seedURLs   []string
	maxWorkers int
	timeout    time.Duration
)

var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Запускает веб-краулер",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Не удалось загрузить конфиг: %v", err)
		}
		redisAddr := fmt.Sprintf("%s:%s", conf.Redis.Host, conf.Redis.Port)
		redisClient := redis.NewRedisClient(redisAddr)
		database := db.NewDatabase("mongodb://localhost:27017")
		domainLimiter := limiter.NewDomainLimiter()
		robotsChecker := robots.NewRobotsChecker()
		resultCh := make(chan string, maxWorkers)

		for _, raw := range seedURLs {
			url := strings.TrimSpace(raw)
			if url != "" {
				redisClient.PushToSeedQueue(url)
				log.Printf("Seed добавлен: %s", url)
			}
		}

		var wg sync.WaitGroup
		for i := 0; i < maxWorkers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				crawler.StartWorker(id, redisClient, database, resultCh, domainLimiter, robotsChecker)
			}(i)
		}

		go func() {
			for msg := range resultCh {
				log.Println("CRAWLER:", msg)
			}
		}()

		timer := time.After(timeout)
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Println("Краулер завершил работу")
		case <-timer:
			log.Println("Таймаут достигнут — остановка")
		}

		close(resultCh)
		redisClient.ClearQueues()
	},
}

func init() {
	crawlCmd.Flags().StringSliceVar(&seedURLs, "seed", []string{}, "Seed URL (можно несколько)")
	crawlCmd.Flags().IntVar(&maxWorkers, "workers", 3, "Количество воркеров")
	crawlCmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "Таймаут выполнения")
	rootCmd.AddCommand(crawlCmd)
}
