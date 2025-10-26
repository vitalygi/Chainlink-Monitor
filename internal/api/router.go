package api


import (
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, redisClient *redis.Client) {
	feedHandler := NewFeedHandler(redisClient)

	e.GET("/", feedHandler.GetPrice)
}