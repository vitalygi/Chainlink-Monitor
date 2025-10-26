package main

import (
	"github.com/vitalygi/chainlink-monitor/internal/config"
	"github.com/vitalygi/chainlink-monitor/internal/worker"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatal("redis connection failed", err)
	}

	appWorker := worker.NewWorker(cfg, rdb)
	appWorker.Run()
}