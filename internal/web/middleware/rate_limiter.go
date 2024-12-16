package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dpcamargo/fullcycle-rate-limiter/internal/database"
	"github.com/labstack/echo/v4"
)

type RateLimiterConf struct {
	Limit    int
	Interval time.Duration
}

type RateLimiter struct {
	db   database.DatabaseInterface
	Conf map[string]RateLimiterConf
}

func NewRateLimiter(db database.DatabaseInterface) *RateLimiter {
	return &RateLimiter{
		db:   db,
		Conf: map[string]RateLimiterConf{},
	}
}

func (r *RateLimiter) AddTokenConf(token string, limit int, interval time.Duration) {
	r.Conf[token] = RateLimiterConf{
		Limit:    limit,
		Interval: interval,
	}
}

func (r *RateLimiter) CheckRateLimitCount(ctx context.Context, token string, ip bool) (bool, error) {
	count, err := r.db.GetCount(ctx, token)
	if err != nil {
		return false, err
	}
	if ip {
		token = "ip"
	}
	if count >= r.Conf[token].Limit {
		return false, nil
	}
	return true, nil
}

func (r *RateLimiter) IncreaseRateLimitCount(ctx context.Context, token string, ip bool) error {
	key := token
	if ip {
		token = "ip"
	}
	return r.db.IncrementCount(ctx, key, r.Conf[token].Interval)
}

func (r *RateLimiter) UpdateExpiration(ctx context.Context, token string, ip bool) error {
	key := token
	if ip {
		token = "ip"
	}
	return r.db.UpdateExpiration(ctx, key, r.Conf[token].Interval)
}

func RateLimiterMiddleware(ctx context.Context, rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var key string
			var ip bool
			token := c.Request().Header.Get("API_KEY")
			if token != "" {
				key = token
			} else {
				key = c.RealIP()
				ip = true
			}

			ok, err := RateLimiterCore(ctx, rl, key, ip)
			if err != nil {
				fmt.Println("rate limit error. key: ", key)
				c.String(http.StatusInternalServerError, "rate limit error")
				return err
			}
			if !ok {
				fmt.Println("rate limit exceeded. key ", key)
				c.String(http.StatusTooManyRequests, "you have reached the maximum number of requests or actions allowed within a certain time frame")
				return nil
			}
			return next(c)
		}
	}
}

func RateLimiterCore(ctx context.Context, rl *RateLimiter, key string, ip bool) (bool, error) {
	ok, err := rl.CheckRateLimitCount(ctx, key, ip)
	if err != nil {
		return false, err
	}

	if !ok {
		rl.UpdateExpiration(ctx, key, ip)
		return false, nil
	}

	err = rl.IncreaseRateLimitCount(ctx, key, ip)
	if err != nil {
		return false, err
	}
	return true, nil
}
