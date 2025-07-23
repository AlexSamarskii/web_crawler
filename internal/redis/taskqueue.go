package redis

import (
	"context"
	"log"

	"github.com/AlexSamarskii/web_crawler/pkg/connector"
	"github.com/go-redis/redis/v8"

	cfg "github.com/AlexSamarskii/web_crawler/internal/config"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

func (r *RedisClient) GetCtx() context.Context {
	return r.ctx
}

func NewRedisClient(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := connector.NewRedisConnection(cfg.RedisConfig{})
	if err != nil {
		log.Fatalf("не удалось установить соединение с Redis: %v", err)
	}
	log.Println("успешно установлено соединение с Redis")

	return &RedisClient{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (r *RedisClient) PushToSeedQueue(url string) {
	err := r.client.LPush(r.ctx, "seed_queue", url).Err()
	if err != nil {
		log.Printf("Error pushing seed URL %s to queue: %v", url, err)
	}
}

func (r *RedisClient) PushToURLQueue(url string) {
	if err := r.client.RPush(r.ctx, "url_queue", url).Err(); err != nil {
		log.Printf("Error pushing to url_queue: %v", err)
	}
}

func (r *RedisClient) PushToURLQueueWithDepth(url string, depth int) {
	if depth > 10 {
		return
	}
	r.client.LPush(r.ctx, "url_queue", url)
	r.client.HSet(r.ctx, "url_depths", url, depth)
}

func (r *RedisClient) PopNextURL() string {
	var url string

	url, _ = r.client.LPop(r.ctx, "seed_queue").Result()
	if url != "" {
		r.MarkVisited(url)
		return url
	}
	return ""
}

func (r *RedisClient) IsVisited(url string) bool {
	exists, err := r.client.SIsMember(r.ctx, "visited_urls", url).Result()
	if err != nil {
		log.Printf("Ошибка проверки посещенных URLов: %v", err)
		return false
	}
	return exists
}

func (r *RedisClient) MarkVisited(url string) {
	err := r.client.SAdd(r.ctx, "visited_urls", url).Err()
	if err != nil {
		log.Printf("Ошибка при пометке URLа посещенным: %v", err)
	} else {
		log.Printf("URL %s посещен", url)
	}
}

func (r *RedisClient) ClearQueues() {
	r.client.Del(r.ctx, "url_queue")
	r.client.Del(r.ctx, "seed_queue")
}
