package core

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type WsClient struct {
	l      *zap.SugaredLogger
	client *ethclient.Client
	mu     sync.RWMutex
}

func NewWsClient() (*WsClient, error) {
	endpoint := viper.GetString("fantom_chain.ws_endpoint")
	if err := validation.Validate(endpoint, is.URL); err != nil {
		return nil, err
	}
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	return &WsClient{
		l:      zap.S(),
		client: client,
		mu:     sync.RWMutex{},
	}, nil
}

func (c *WsClient) GetETHClient() *ethclient.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}
