package connector

import (
	"fmt"
	"time"

	"github.com/AlexSamarskii/web_crawler/internal/config"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

func NewRedisConnection(cfg config.RedisConfig) (redis.Conn, error) {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	fmt.Println(address)
	conn, err := redis.Dial("tcp", address,
		redis.DialPassword(cfg.Password),
		redis.DialDatabase(cfg.DB),
		redis.DialConnectTimeout(5*time.Second),
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с Redis")

		return nil, fmt.Errorf("не удалось установить соединение с Redis: %w", err)
	}

	if _, err := conn.Do("PING"); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping Redis")

		return nil, fmt.Errorf("не удалось выполнить ping Redis: %w", err)
	}

	return conn, nil
}
