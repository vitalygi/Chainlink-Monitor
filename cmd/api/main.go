package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/vitalygi/chainlink-monitor/internal/api"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

    redisClient := redis.NewClient(&redis.Options{Addr: "redis:6379"})

	api.RegisterRoutes(e, redisClient)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Logger.Fatal(e.Start(":8080"))
}
