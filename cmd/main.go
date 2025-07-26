package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/AlexSamarskii/web_crawler/internal/config"
	"github.com/AlexSamarskii/web_crawler/internal/crawler"
	"github.com/AlexSamarskii/web_crawler/internal/db"
	redisqueue "github.com/AlexSamarskii/web_crawler/internal/redis"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	maxWorkers := conf.Crawler.MaxWorkers

	redisClient := redisqueue.NewRedisClient("localhost:6379")

	database := db.NewDatabase(conf.Database.ConnStr)

	resultCh := make(chan string, maxWorkers)

	seedURLs := []string{
		"https://golang.org",
		"https://vk.com/",
		"https://habr.com/",
		"https://www.youtube.com/",
	}

	for _, url := range seedURLs {
		log.Printf("Пушим начальный URL: %s", url)
		redisClient.PushToSeedQueue(url)
	}

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			crawler.StartWorker(id, redisClient, database, resultCh)
		}(i)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for msg := range resultCh {
			fmt.Println(msg)
		}
	}()

	select {
	case <-stopChan:
	case <-time.After(30 * time.Second):
	}

	wg.Wait()
	redisClient.ClearQueues()
	close(resultCh)

	log.Println("Завершение работы сервера приложения")

}
