package core

import (
	"context"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/machinebox/graphql"
	"github.com/quangkeu95/fantom-bot/pkg"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	GraphqlTimeout = 5 * time.Second
)

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
	client := graphql.NewClient(endpoint)
	return &GraphqlClient{
		l:      l,
		client: client,
	}, nil
}

func (c *GraphqlClient) GetListValidators(parentCtx context.Context) ([]pkg.SFCValidator, error) {
	var result = make([]pkg.SFCValidator, 0)
	var query = graphql.NewRequest(`
		query {
			stakers {
				id
				stakerAddress
				totalStake
				isActive
				ifOffline
				createdTime
				createdEpoch
				deactivatedTime
				deactivatedEpoch
			}
		}	  
	`)

	var res struct {
		Stakers []struct {
			ID               string `json:"id"`
			Address          string `json:"stakerAddress"`
			IsActive         bool   `json:"isActive"`
			IsOffline        bool   `json:"isOffline"`
			CreatedTime      string `json:"createdTime"`
			CreatedEpoch     string `json:"createdEpoch"`
			DeactivatedTime  string `json:"deactivatedTime"`
			DeactivatedEpoch string `json:"deactivatedEpoch"`
		} `json:"stakers"`
	}
	ctx, cancel := context.WithTimeout(parentCtx, GraphqlTimeout)
	defer cancel()
	if err := c.client.Run(ctx, query, &res); err != nil {
		c.l.Warnw("graphql get list validators error", "error", err)
		return nil, err
	}

	for _, item := range res.Stakers {
		id, err := hexutil.DecodeBig(item.ID)
		if err != nil {
			c.l.Debugw("parse id error", "error", err)
			continue
		}
		createdTime, err := hexutil.DecodeBig(item.CreatedTime)
		if err != nil {
			c.l.Debugw("parse created time error", "error", err)
			continue
		}
		createdEpoch, err := hexutil.DecodeBig(item.CreatedEpoch)
		if err != nil {
			c.l.Debugw("parse created epoch error", "error", err)
			continue
		}
		deactivatedTime, err := hexutil.DecodeBig(item.DeactivatedTime)
		if err != nil {
			c.l.Debugw("parse deactivated time error", "error", err)
			continue
		}
		deactivatedEpoch, err := hexutil.DecodeBig(item.DeactivatedEpoch)
		if err != nil {
			c.l.Debugw("parse deactivated epoch error", "error", err)
			continue
		}
		validator := pkg.SFCValidator{
			ID:               id.Uint64(),
			Address:          item.Address,
			IsActive:         item.IsActive,
			IsOffline:        item.IsOffline,
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
	var query = graphql.NewRequest(`
		query ($blockNumber: Long!) {
			block(number: $blockNumber) {
				hash
				transactionCount
				txList {
					hash
					from
					to
					value
					inputData
				}
			}
		}	  
	`)

	query.Var("blockNumber", hexutil.EncodeUint64(blockNumber))

	var res struct {
		Block struct {
			Hash             string `json:"hash"`
			TransactionCount int    `json:"transactionCount"`
			TxList           []struct {
				Hash      string `json:"hash"`
				From      string `json:"from"`
				To        string `json:"to"`
				Value     string `json:"value"`
				InputData string `json:"inputData"`
			} `json:"txList"`
		} `json:"block"`
	}
	ctx, cancel := context.WithTimeout(parentCtx, GraphqlTimeout)
	defer cancel()
	if err := c.client.Run(ctx, query, &res); err != nil {
		c.l.Warnw("graphql get list transactions error", "error", err)
		return nil, err
	}

	if res.Block.TransactionCount == 0 {
		return []pkg.TransferLog{}, nil
	}
	for _, tx := range res.Block.TxList {
		if tx.InputData != "0x" {
			continue
		}
		amount, err := hexutil.DecodeBig(tx.Value)
		if err != nil {
			c.l.Debugw("parse tx value error", "error", err)
			continue
		}
		result = append(result, pkg.TransferLog{
			BlockNumber: blockNumber,
			TxHash:      tx.Hash,
			From:        tx.From,
			To:          tx.To,
			Amount:      pkg.WeiToFloat(amount, 18),
		})
	}
	return result, nil
}
