package api

import (
	"context"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

type FeedHandler struct {
	redisClient *redis.Client
}

func NewFeedHandler(redisClient *redis.Client) *FeedHandler {
	return &FeedHandler{
		redisClient: redisClient,
	}
}

func (h *FeedHandler) GetPrice(c echo.Context) error {
	currencyPair := c.QueryParam("currency")

	if currencyPair == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "currency query parameter is required",
		})
	}

	ctx := context.Background()
	priceData, err := h.redisClient.Get(ctx, currencyPair).Result()

	if err != nil {
		if err == redis.Nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "price data not found for currency pair: " + currencyPair,
			})
		}

		c.Logger().Errorf("Redis error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "could not retrieve data from storage",
		})
	}

	return c.JSONBlob(http.StatusOK, []byte(priceData))
}