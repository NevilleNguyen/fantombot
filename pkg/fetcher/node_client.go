package fetcher

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	etherCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/quangkeu95/fantom-bot/pkg"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type NodeClient struct {
	l      *zap.SugaredLogger
	client *ethclient.Client
	mu     sync.RWMutex

	chainID *big.Int
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

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &NodeClient{
		l:       zap.S(),
		client:  client,
		mu:      sync.RWMutex{},
		chainID: chainID,
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

func (c *NodeClient) GetListValidators(ctx context.Context) ([]pkg.SFCValidator, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *NodeClient) GetListFTMTransferByBlock(ctx context.Context, blockNumber uint64) ([]pkg.TransferLog, error) {
	block, err := c.client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		c.l.Warnw("get list FTM transfer by block error", "error", err, "block_number", blockNumber)
		return nil, err
	}

	var result = make([]pkg.TransferLog, 0)
	for _, tx := range block.Transactions() {
		data := etherCommon.Bytes2Hex(tx.Data())
		if data != "" {
			continue
		}

		var from string
		msg, err := tx.AsMessage(types.NewEIP155Signer(c.chainID))
		if err != nil {
			c.l.Warnw("get tx as message error", "error", err)
			continue
		}
		from = msg.From().Hex()

		result = append(result, pkg.TransferLog{
			BlockNumber: blockNumber,
			TxHash:      tx.Hash().Hex(),
			From:        from,
			To:          tx.To().Hex(),
			Amount:      pkg.WeiToFloat(tx.Value(), 18),
		})
	}
	return result, nil
}
