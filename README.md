# Go Web Crawler

Многопоточный веб-краулер на языке Go, который безопасно и этично сканирует сайты, извлекает ссылки и хранит данные. Поддерживает `robots.txt`, ограничение частоты запросов, метрики и HTTP API для управления.

---

## Возможности

- Краулинг сайтов с контролем глубины и скорости
- Многопоточность (настраиваемое количество воркеров)
- Ограничение скорости запросов (rate limiting)
- Хранение URL в **Redis** (очереди и дедупликация)
- Сохранение результатов в **MongoDB**
- CLI-интерфейс на базе [Cobra](https://github.com/spf13/cobra)

---

## Запуск

```bash
go build -o crawler-cli
./crawler-cli crawl --seed https://golang.org --workers 5 --timeout 2m
```

```text
+----------------+ +--------+ +-----------+
| HTTP API | --> | Воркеры | --> | Redis |
+----------------+ +--------+ +-----------+
| |
v v
+-----------+ +-----------+
| MongoDB | | Prometheus|
+-----------+ +-----------+
```

- **Redis**: очередь URL, множество посещённых, хранение глубины.
- **MongoDB**: коллекция `urls` с полями `url` (источник) и `link` (найденная ссылка).

## Конфигурация

```yml
redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

database:
  conn_str: "mongodb://localhost:27017"

crawler:
  user_agent: "Go-Crawler-Bot"
  max_workers: 5
  rate_limit: 5
```

### Пример запуска

```bash
./crawler-cli crawl --seed https://example.com --workers 10 --timeout 5m
```

### Поддерживаемые флаги

- --seed — начальный URL
- --workers — количество параллельных воркеров
- --timeout — максимальное время выполнения
