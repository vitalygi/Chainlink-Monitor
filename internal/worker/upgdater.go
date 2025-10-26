package worker

import (
	"github.com/vitalygi/chainlink-monitor/internal/config"
	"github.com/vitalygi/chainlink-monitor/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
)

// AnswerUpdated event hash
var eventSignature = common.HexToHash("0x0559884fd3a460db3073b7fc896cc77986f16e378210ded43186175bf646fc5f")

type Worker struct {
	ethClient   *ethclient.Client
	redisClient *redis.Client
	priceFeeds  map[common.Address]config.FeedConfig
	nodeURL     string
}

func NewWorker(cfg *config.Config, redisClient *redis.Client) *Worker {
	priceFeedsMap := make(map[common.Address]config.FeedConfig)
	for _, feed := range cfg.Feeds {
		priceFeedsMap[common.HexToAddress(feed.HexAddress)] = feed
	}

	return &Worker{
		nodeURL:     cfg.NodeUrl,
		redisClient: redisClient,
		priceFeeds:  priceFeedsMap,
	}
}

func (w *Worker) Run() {
	for {
		log.Println("Connecting to Ethereum node...")
		client, err := ethclient.Dial(w.nodeURL)
		if err != nil {
			log.Printf("[ERROR] Connection failed: %v. Retrying in 15 seconds...", err)
			time.Sleep(15 * time.Second)
			continue
		}
		w.ethClient = client

		log.Println("Connection established. Subscribing to price feed events...")
		err = w.monitorPrices()

		log.Printf("[WARN] Subscription lost: %v. Reconnecting...", err)
		client.Close()
		time.Sleep(5 * time.Second)
	}
}

func (w *Worker) monitorPrices() error {
	addresses := make([]common.Address, 0, len(w.priceFeeds))
	for addr := range w.priceFeeds {
		addresses = append(addresses, addr)
	}

	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    [][]common.Hash{{eventSignature}},
	}

	logs := make(chan types.Log)
	sub, err := w.ethClient.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			return err
		case vLog := <-logs:
			go w.handleLog(vLog)
		}
	}
}


func (w *Worker) handleLog(vLog types.Log) {
	feed, ok := w.priceFeeds[vLog.Address]
	if !ok {
		return
	}

	if len(vLog.Topics) < 2 || len(vLog.Data) < 32 {
		log.Printf("[WARN] Incomplete log received for %s, skipping.", feed.CurrencyPair)
		return
	}

	priceBigInt := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
	priceDecimal := decimal.NewFromBigInt(priceBigInt, -feed.Decimals)
	priceFloat, _ := priceDecimal.Float64()

	timestampBigInt := new(big.Int).SetBytes(vLog.Data[:32])
	timestamp := timestampBigInt.Int64()

	priceData := models.Price{
		CurrencyPair: feed.CurrencyPair,
		Price:        priceFloat,
		Timestamp:    timestamp,
	}

	jsonData, _ := json.Marshal(priceData)

	err := w.redisClient.Set(context.Background(), priceData.CurrencyPair, jsonData, 0).Err()
	if err != nil {
		log.Printf("[ERROR] Failed to save price to Redis for %s: %v", priceData.CurrencyPair, err)
		return
	}

	log.Printf("[PRICE] %s: %s", priceData.CurrencyPair, priceDecimal.String())
}