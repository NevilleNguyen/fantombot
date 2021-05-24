package core

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type NodeClient struct {
	l      *zap.SugaredLogger
	client *ethclient.Client
	mu     sync.RWMutex
}

func NewNodeClient() (*NodeClient, error) {
	endpoint := viper.GetString("fantom_chain.rpc_endpoint")
	if err := validation.Validate(endpoint, is.URL); err != nil {
		return nil, err
	}
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	return &NodeClient{
		l:      zap.S(),
		client: client,
		mu:     sync.RWMutex{},
	}, nil
}

func (c *NodeClient) GetETHClient() *ethclient.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

func (c *NodeClient) GetLatestBlock(ctx context.Context) (uint64, error) {
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	return header.Number.Uint64(), nil
}
