package core

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	etherCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/quangkeu95/fantom-bot/lib/contracts"
	"github.com/quangkeu95/fantom-bot/pkg"
	"github.com/quangkeu95/fantom-bot/pkg/fetcher"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	BlockRange = uint64(50000)
)

type SFCClient struct {
	l          *zap.SugaredLogger
	nodeClient *fetcher.NodeClient
	wsClient   *fetcher.WsClient

	sfcContract   *contracts.SFC
	sfcWsContract *contracts.SFC
}

func NewSFCClient(nodeClient *fetcher.NodeClient, wsClient *fetcher.WsClient) (*SFCClient, error) {
	l := zap.S()
	sfcAddr := viper.GetString("fantom_chain.sfc_contract_address")
	if err := validation.Validate(sfcAddr, validation.Required); err != nil {
		l.Errorw("sfc contract address is invalid", "error", err)
		return nil, err
	}
	sfcContract, err := contracts.NewSFC(etherCommon.HexToAddress(sfcAddr), nodeClient.GetETHClient())
	if err != nil {
		l.Warnw("init SFC contract error", "error", err)
		return nil, err
	}

	sfcWsContract, err := contracts.NewSFC(etherCommon.HexToAddress(sfcAddr), wsClient.GetETHClient())
	if err != nil {
		l.Warnw("init SFC ws contract error", "error", err)
		return nil, err
	}

	return &SFCClient{
		l:             l,
		nodeClient:    nodeClient,
		wsClient:      wsClient,
		sfcContract:   sfcContract,
		sfcWsContract: sfcWsContract,
	}, nil
}

func (c *SFCClient) GetCreatedValidatorByID(ctx context.Context, ids []uint64) ([]pkg.SFCValidator, error) {
	var validatorID = make([]*big.Int, 0)
	for _, id := range ids {
		validatorID = append(validatorID, new(big.Int).SetUint64(id))
	}

	opts := &bind.FilterOpts{
		Context: ctx,
	}
	iterator, err := c.sfcContract.FilterCreatedValidator(opts, validatorID, nil)
	if err != nil {
		c.l.Warnw("get created validator error", "error", err)
		return nil, err
	}

	var listValidators = make([]pkg.SFCValidator, 0)
	for iterator.Next() {
		if err := iterator.Error(); err != nil {
			c.l.Warnw("parse created validator error", "error", err)
			continue
		}
		v := pkg.ToSFCValidator(iterator.Event)
		listValidators = append(listValidators, v)
	}
	return listValidators, nil
}

func (c *SFCClient) GetCreatedValidatorByBlock(ctx context.Context, fromBlock uint64, toBlock *uint64) ([]pkg.SFCValidator, error) {
	var listValidators = make([]pkg.SFCValidator, 0)
	var (
		startBlock = fromBlock
		endBlock   uint64
	)
	if toBlock == nil {
		latestBlock, err := c.nodeClient.GetLatestBlock(ctx)
		if err != nil {
			return nil, err
		}
		endBlock = latestBlock
	} else {
		endBlock = *toBlock
	}

	for {
		if startBlock >= endBlock {
			break
		}
		nextBlock := startBlock + BlockRange
		if nextBlock > endBlock {
			nextBlock = endBlock
		}

		opts := &bind.FilterOpts{
			Context: ctx,
			Start:   startBlock,
			End:     &nextBlock,
		}
		iterator, err := c.sfcContract.FilterCreatedValidator(opts, nil, nil)
		if err != nil {
			c.l.Warnw("get created validator error", "error", err)
			return nil, err
		}

		for iterator.Next() {
			if err := iterator.Error(); err != nil {
				c.l.Warnw("parse created validator error", "error", err)
				continue
			}
			v := pkg.ToSFCValidator(iterator.Event)
			listValidators = append(listValidators, v)
		}
		startBlock = nextBlock + 1
	}

	sort.SliceStable(listValidators, func(i, j int) bool {
		return listValidators[i].ID < listValidators[j].ID
	})

	return listValidators, nil
}

func (c *SFCClient) WatchCreatedValidator(ctx context.Context, validatorCh chan<- pkg.SFCValidator, errCh chan<- error) {
	opts := &bind.WatchOpts{
		Context: ctx,
	}
	sink := make(chan *contracts.SFCCreatedValidator)

	sub, err := c.sfcWsContract.WatchCreatedValidator(opts, sink, nil, nil)
	if err != nil {
		c.l.Warnw("watch created validator error", "error", err)
		errCh <- err
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case validator := <-sink:
			validatorCh <- pkg.ToSFCValidator(validator)
		case err := <-sub.Err():
			errCh <- err
			return
		}
	}
}

func (c *SFCClient) GetValidatorByID(ctx context.Context, id uint64) (pkg.SFCValidator, error) {
	opts := &bind.CallOpts{
		Context: ctx,
	}
	res, err := c.sfcContract.GetValidator(opts, new(big.Int).SetUint64(id))
	if err != nil {
		c.l.Warnw("get validator by id error", "error", err)
		return pkg.SFCValidator{}, err
	}

	var isActive = res.Status.Cmp(new(big.Int)) == 0

	return pkg.SFCValidator{
		ID:               id,
		Address:          res.Auth.Hex(),
		CreatedTime:      res.CreatedTime.Uint64(),
		CreatedEpoch:     res.CreatedEpoch.Uint64(),
		DeactivatedTime:  res.DeactivatedTime.Uint64(),
		DeactivatedEpoch: res.DeactivatedEpoch.Uint64(),
		IsActive:         isActive,
	}, nil
}

func (c *SFCClient) GetLastValidatorID(ctx context.Context) (uint64, error) {
	opts := &bind.CallOpts{
		Context: ctx,
	}
	res, err := c.sfcContract.LastValidatorID(opts)
	if err != nil {
		c.l.Warnw("get last validator id error", "error", err)
		return 0, err
	}
	return res.Uint64(), nil
}

func (c *SFCClient) GetDelegateInfoByBlock(ctx context.Context, fromBlock uint64, toBlock *uint64) ([]pkg.SFCDelegateInfo, error) {
	var result = make([]pkg.SFCDelegateInfo, 0)
	opts := &bind.FilterOpts{
		Context: ctx,
		Start:   fromBlock,
		End:     toBlock,
	}

	iterator, err := c.sfcContract.FilterDelegated(opts, nil, nil)
	if err != nil {
		c.l.Warnw("get delegate error", "error", err)
		return nil, err
	}

	for iterator.Next() {
		if err := iterator.Error(); err != nil {
			c.l.Debugw("parse delegate info error", "error", err)
			continue
		}
		result = append(result, pkg.ToSFCDelegateInfo(iterator.Event))
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})
	return result, nil
}

func (c *SFCClient) WatchDelegateEvent(ctx context.Context, delegateInfoCh chan<- pkg.SFCDelegateInfo, errCh chan<- error) {
	opts := &bind.WatchOpts{
		Context: ctx,
	}
	var (
		sink = make(chan *contracts.SFCDelegated)
	)
	sub, err := c.sfcWsContract.WatchDelegated(opts, sink, nil, nil)
	if err != nil {
		c.l.Warnw("watch delegate error", "error", err)
		errCh <- err
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case info := <-sink:
			delegateInfoCh <- pkg.ToSFCDelegateInfo(info)
		case err := <-sub.Err():
			errCh <- err
			return
		}
	}
}

func (c *SFCClient) GetUndelegateInfoByBlock(ctx context.Context, fromBlock uint64, toBlock *uint64) ([]pkg.SFCUndelegateInfo, error) {
	var result = make([]pkg.SFCUndelegateInfo, 0)
	opts := &bind.FilterOpts{
		Context: ctx,
		Start:   fromBlock,
		End:     toBlock,
	}

	iterator, err := c.sfcContract.FilterUndelegated(opts, nil, nil, nil)
	if err != nil {
		c.l.Warnw("get undelegate error", "error", err)
		return nil, err
	}

	for iterator.Next() {
		if err := iterator.Error(); err != nil {
			c.l.Debugw("parse undelegate error", "error", err)
			continue
		}
		result = append(result, pkg.ToSFCUndelegateInfo(iterator.Event))
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})
	return result, nil
}

func (c *SFCClient) WatchUndelegateEvent(ctx context.Context, undelegateInfoCh chan<- pkg.SFCUndelegateInfo, errCh chan<- error) {
	opts := &bind.WatchOpts{
		Context: ctx,
	}
	var (
		sink = make(chan *contracts.SFCUndelegated)
	)
	sub, err := c.sfcWsContract.WatchUndelegated(opts, sink, nil, nil, nil)
	if err != nil {
		c.l.Warnw("watch undelegate error", "error", err)
		errCh <- err
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case info := <-sink:
			undelegateInfoCh <- pkg.ToSFCUndelegateInfo(info)
		case err := <-sub.Err():
			errCh <- err
			return
		}
	}
}

func (c *SFCClient) WatchClaimRewardEvent(ctx context.Context, rewardInfoCh chan<- pkg.SFCRewardInfo, errCh chan<- error) {
	opts := &bind.WatchOpts{
		Context: ctx,
	}
	var (
		sink = make(chan *contracts.SFCClaimedRewards)
	)
	sub, err := c.sfcWsContract.WatchClaimedRewards(opts, sink, nil, nil)
	if err != nil {
		c.l.Warnw("watch claim reward error", "error", err)
		errCh <- err
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case info := <-sink:
			rewardInfoCh <- pkg.ToSFCRewardInfo(info)
		case err := <-sub.Err():
			errCh <- err
			return
		}
	}
}

func (c *SFCClient) SubscribeNewHead(ctx context.Context, headCh chan<- *types.Header, errCh chan<- error) {
	sub, err := c.wsClient.GetETHClient().SubscribeNewHead(ctx, headCh)
	if err != nil {
		c.l.Warnf("subscribe new head error", "error", err)
		errCh <- err
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-sub.Err():
			errCh <- err
			return
		}
	}
}

func (c *SFCClient) GetTxByHash(ctx context.Context, txHash string) error {
	tx, _, err := c.nodeClient.GetETHClient().TransactionByHash(ctx, etherCommon.HexToHash(txHash))
	if err != nil {
		c.l.Warnw("get tx by hash error", "error", err)
		return err
	}

	data := etherCommon.Bytes2Hex(tx.Data())
	log.Println(data)
	return nil
}

func FormatSFCCreatedValidator(v contracts.SFCCreatedValidator) string {
	return fmt.Sprintf("validator_id: %v | validator_address: %v | created_epoch: %v | created_time %v",
		v.ValidatorID.Int64(), v.Auth.Hex(), v.CreatedEpoch.Int64(), v.CreatedTime.Int64())
}
