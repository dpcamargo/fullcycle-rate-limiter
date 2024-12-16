package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/dpcamargo/fullcycle-rate-limiter/internal/database"
	"github.com/dpcamargo/fullcycle-rate-limiter/internal/web/controller"
	"github.com/dpcamargo/fullcycle-rate-limiter/internal/web/middleware"
	"github.com/labstack/echo/v4"
)

type Config struct {
	IPLimit     int           `mapstructure:"ip_limit"`
	IPDuration  time.Duration `mapstructure:"IP_DURATION"`
	APIKey      string        `mapstructure:"API_KEY"`
	APILimit    int           `mapstructure:"API_LIMIT"`
	APIDuration time.Duration `mapstructure:"API_DURATION"`
	RedisAddr   string        `mapstructure:"REDIS_ADDR"`
}

func main() {
	ctx := context.Background()
	var config Config

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	viper.AutomaticEnv()
	viper.BindEnv("IP_LIMIT")
	viper.BindEnv("IP_DURATION")
	viper.BindEnv("API_KEY")
	viper.BindEnv("API_LIMIT")
	viper.BindEnv("API_DURATION")
	viper.BindEnv("REDIS_ADDR")
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	db := database.NewDatabase(config.RedisAddr)

	rateLimitConf := middleware.NewRateLimiter(db)
	rateLimitConf.AddTokenConf("ip", config.IPLimit, config.IPDuration)
	rateLimitConf.AddTokenConf(config.APIKey, config.APILimit, config.APIDuration)

	e := echo.New()
	e.Use(middleware.LoggingMiddleware)
	e.Use(middleware.RateLimiterMiddleware(ctx, rateLimitConf))

	e.GET("/", controller.GetIP)
	e.Logger.Fatal(e.Start(":8080"))
}
