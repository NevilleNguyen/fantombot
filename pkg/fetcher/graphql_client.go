package fetcher

import (
	"context"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/quangkeu95/fantom-bot/pkg"
	"github.com/shurcooL/graphql"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	GraphqlTimeout = 5 * time.Second
)

type Long string

type GraphqlClient struct {
	l      *zap.SugaredLogger
	client *graphql.Client
}

func NewGraphqlClient() (*GraphqlClient, error) {
	l := zap.S()
	endpoint := viper.GetString("fantom_chain.graphql_endpoint")
	if err := validation.Validate(endpoint, validation.Required); err != nil {
		l.Errorw("sfc contract address is invalid", "error", err)
		return nil, err
	}
	client := graphql.NewClient(endpoint, nil)
	return &GraphqlClient{
		l:      l,
		client: client,
	}, nil
}

func (c *GraphqlClient) GetListValidators(parentCtx context.Context) ([]pkg.SFCValidator, error) {
	var result = make([]pkg.SFCValidator, 0)
	var query struct {
		Stakers []struct {
			ID               graphql.String  `graphql:"id"`
			Address          graphql.String  `graphql:"stakerAddress"`
			IsActive         graphql.Boolean `graphql:"isActive"`
			IsOffline        graphql.Boolean `graphql:"isOffline"`
			CreatedTime      graphql.String  `graphql:"createdTime"`
			CreatedEpoch     graphql.String  `graphql:"createdEpoch"`
			DeactivatedTime  graphql.String  `graphql:"deactivatedTime"`
			DeactivatedEpoch graphql.String  `graphql:"deactivatedEpoch"`
		} `graphql:"stakers"`
	}

	ctx, cancel := context.WithTimeout(parentCtx, GraphqlTimeout)
	defer cancel()
	if err := c.client.Query(ctx, &query, nil); err != nil {
		c.l.Warnw("graphql get list validators error", "error", err)
		return nil, err
	}

	for _, item := range query.Stakers {
		id, err := hexutil.DecodeBig(string(item.ID))
		if err != nil {
			c.l.Debugw("parse id error", "error", err)
			continue
		}
		createdTime, err := hexutil.DecodeBig(string(item.CreatedTime))
		if err != nil {
			c.l.Debugw("parse created time error", "error", err)
			continue
		}
		createdEpoch, err := hexutil.DecodeBig(string(item.CreatedEpoch))
		if err != nil {
			c.l.Debugw("parse created epoch error", "error", err)
			continue
		}
		deactivatedTime, err := hexutil.DecodeBig(string(item.DeactivatedTime))
		if err != nil {
			c.l.Debugw("parse deactivated time error", "error", err)
			continue
		}
		deactivatedEpoch, err := hexutil.DecodeBig(string(item.DeactivatedEpoch))
		if err != nil {
			c.l.Debugw("parse deactivated epoch error", "error", err)
			continue
		}
		validator := pkg.SFCValidator{
			ID:               id.Uint64(),
			Address:          string(item.Address),
			IsActive:         bool(item.IsActive),
			IsOffline:        bool(item.IsOffline),
			CreatedTime:      createdTime.Uint64(),
			CreatedEpoch:     createdEpoch.Uint64(),
			DeactivatedTime:  deactivatedTime.Uint64(),
			DeactivatedEpoch: deactivatedEpoch.Uint64(),
		}
		result = append(result, validator)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (c *GraphqlClient) GetListFTMTransferByBlock(parentCtx context.Context, blockNumber uint64) ([]pkg.TransferLog, error) {
	var result = make([]pkg.TransferLog, 0)
	var query struct {
		Block struct {
			Hash             graphql.String `graphql:"hash"`
			TransactionCount graphql.Int    `graphql:"transactionCount"`
			TxList           []struct {
				Hash      graphql.String `graphql:"hash"`
				From      graphql.String `graphql:"from"`
				To        graphql.String `graphql:"to"`
				Value     graphql.String `graphql:"value"`
				InputData graphql.String `graphql:"inputData"`
			} `graphql:"txList"`
		} `graphql:"block(number: $blockNumber)"`
	}

	ctx, cancel := context.WithTimeout(parentCtx, GraphqlTimeout)
	defer cancel()
	params := map[string]interface{}{
		"blockNumber": Long(hexutil.EncodeUint64(blockNumber)),
	}
	if err := c.client.Query(ctx, &query, params); err != nil {
		c.l.Warnw("graphql get list transactions error", "error", err, "block_number", blockNumber)
		return nil, err
	}

	if int(query.Block.TransactionCount) == 0 {
		return []pkg.TransferLog{}, nil
	}
	for _, tx := range query.Block.TxList {
		if string(tx.InputData) != "0x" {
			continue
		}
		amount, err := hexutil.DecodeBig(string(tx.Value))
		if err != nil {
			c.l.Debugw("parse tx value error", "error", err)
			continue
		}
		result = append(result, pkg.TransferLog{
			BlockNumber: blockNumber,
			TxHash:      string(tx.Hash),
			From:        string(tx.From),
			To:          string(tx.To),
			Amount:      pkg.WeiToFloat(amount, 18),
		})
	}
	return result, nil
}
