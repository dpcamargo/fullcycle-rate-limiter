package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	limit    = 3
	interval = 10 * time.Second
)

var apiKeyMap = map[string]APIKeyLimit{
	"api-key-1": {limit: 5, interval: 10 * time.Second},
	"api-key-2": {limit: 10, interval: 10 * time.Second},
}

type RateLimiter struct {
	apiKeyLimit map[string]APIKeyLimit
	visitCount  map[string]int
	limitation  int
	interval    time.Duration
	mu          sync.Mutex
}

type APIKeyLimit struct {
	limit    int
	count    int
	interval time.Duration
}

func NewRateLimiter(limit int, interval time.Duration, apiKeyMap map[string]APIKeyLimit) *RateLimiter {
	return &RateLimiter{
		visitCount:  make(map[string]int),
		limitation:  limit,
		interval:    interval,
		apiKeyLimit: apiKeyMap,
	}
}

func main() {
	e := echo.New()

	rl := NewRateLimiter(limit, interval, apiKeyMap)
	e.Use(LoggingMiddleware)
	e.Use(RateLimiterMiddleware(rl))

	e.GET("/", GetIP)
	e.Logger.Fatal(e.Start(":8080"))
}

func GetIP(c echo.Context) error {
	req := c.Request()

	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(req.RemoteAddr, ":")[0]
	} else {
		ip = strings.Split(ip, ",")[0]
	}
	return c.String(http.StatusOK, "Your IP address is: "+ip)
}

func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Printf("Request method: %s, path: %s\n", c.Request().Method, c.Request().URL.Path)
		return next(c)
	}
}

func RateLimiterMiddleware(rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("API_KEY")
			if token != "" {
				rl.mu.Lock()
				apiConf, ok := rl.apiKeyLimit[token]
				if !ok {
					rl.mu.Unlock()
					return c.String(http.StatusBadRequest, "Invalid API key")
				}
				if apiConf.count >= apiConf.limit {
					fmt.Println("API key rate limit exceeded for token:", token)
					rl.mu.Unlock()
					return c.String(http.StatusTooManyRequests, "you have reached the maximum number of requests or actions allowed within a certain time frame")
				}
				apiConf.count++
				rl.apiKeyLimit[token] = apiConf
				rl.mu.Unlock()
				go resetAPIRateLimit(token, rl, apiConf.interval)
				return next(c)
			}

			ip := c.RealIP()
			rl.mu.Lock()
			if rl.visitCount[ip] >= rl.limitation {
				fmt.Println("IP rate limit exceeded for IP:", ip)
				rl.mu.Unlock()
				return c.String(http.StatusTooManyRequests, "you have reached the maximum number of requests or actions allowed within a certain time frame")
			}
			rl.visitCount[ip]++
			rl.mu.Unlock()
			go resetIPRateLimit(ip, rl)
			return next(c)
		}
	}
}

func resetIPRateLimit(ip string, rl *RateLimiter) {
	time.Sleep(rl.interval)
	rl.mu.Lock()
	rl.visitCount[ip]--
	rl.mu.Unlock()
	fmt.Println("Decreased rate limit count for IP:", ip)
}

func resetAPIRateLimit(token string, rl *RateLimiter, sleepTimer time.Duration) {
	time.Sleep(sleepTimer)
	rl.mu.Lock()
	apiConf := rl.apiKeyLimit[token]
	apiConf.count--
	rl.apiKeyLimit[token] = apiConf
	rl.mu.Unlock()
	fmt.Println("Decreased rate limit count for API key:", token)
}
