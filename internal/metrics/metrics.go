package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Счётчики
	URLsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "crawler_urls_processed_total",
			Help: "Общее количество обработанных URL",
		},
	)

	URLsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "crawler_urls_failed_total",
			Help: "Общее количество URL, которые не удалось обработать",
		},
	)

	LinksExtracted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "crawler_links_extracted_total",
			Help: "Общее количество извлечённых ссылок",
		},
	)

	URLProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "crawler_url_processing_duration_seconds",
			Help:    "Время обработки одного URL (в секундах)",
			Buckets: prometheus.DefBuckets,
		},
	)

	ActiveWorkers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "crawler_active_workers",
			Help: "Количество активных воркеров",
		},
	)

	HTTPStatusCodes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crawler_http_requests_total",
			Help: "Количество HTTP-запросов по статус-коду",
		},
		[]string{"code"},
	)

	URLsByDomain = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crawler_urls_processed_by_domain_total",
			Help: "Количество обработанных URL по доменам",
		},
		[]string{"domain"},
	)

	URLsByDepth = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crawler_urls_processed_by_depth",
			Help: "Количество обработанных URL по глубине обхода",
		},
		[]string{"depth"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(URLsProcessed)
	prometheus.MustRegister(URLsFailed)
	prometheus.MustRegister(LinksExtracted)
	prometheus.MustRegister(URLProcessingDuration)
	prometheus.MustRegister(ActiveWorkers)
	prometheus.MustRegister(HTTPStatusCodes)
	prometheus.MustRegister(URLsByDomain)
	prometheus.MustRegister(URLsByDepth)
}
