package robots

import (
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

type RobotsChecker struct {
	cache  map[string]*robotstxt.RobotsData
	mu     sync.RWMutex
	client *http.Client
}

func NewRobotsChecker() *RobotsChecker {
	return &RobotsChecker{
		cache: make(map[string]*robotstxt.RobotsData),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (rc *RobotsChecker) GetRobotsTxt(domain string) (*robotstxt.RobotsData, error) {
	rc.mu.RLock()
	robots, exists := rc.cache[domain]
	rc.mu.RUnlock()

	if exists {
		return robots, nil
	}

	robotsURL := "https://" + domain + "/robots.txt"
	resp, err := rc.client.Get(robotsURL)
	if err != nil {
		robotsURL = "http://" + domain + "/robots.txt"
		resp, err = rc.client.Get(robotsURL)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	robots, err = robotstxt.FromBytes(body)
	if err != nil {
		return nil, err
	}

	rc.mu.Lock()
	rc.cache[domain] = robots
	rc.mu.Unlock()

	return robots, nil
}

// CanFetch проверяет, разрешён ли URL для данного User-Agent
func (rc *RobotsChecker) CanFetch(userAgent, urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	robots, err := rc.GetRobotsTxt(parsedURL.Host)
	if err != nil || robots == nil {
		return true // Если robots.txt недоступен — разрешаем
	}

	return robots.TestAgent(urlStr, userAgent)
}
