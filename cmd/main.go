package main

import (
	"log"

	"github.com/AlexSamarskii/web_crawler/internal/config"
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

	// resultCh := make(chan string, maxWorkers)

	// seedURLs := []string{
	// 	"https://golang.org",
	// 	"https://www.w3schools.com/",
	// 	"https://www.reddit.com/",
	// 	"https://www.youtube.com/",
	// }

}
