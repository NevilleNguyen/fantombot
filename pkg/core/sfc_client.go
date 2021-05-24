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
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	BlockRange = uint64(50000)
)

type SFCClient struct {
	l          *zap.SugaredLogger
	nodeClient *NodeClient
	wsClient   *WsClient

	sfcContract   *contracts.SFC
	sfcWsContract *contracts.SFC
}

func NewSFCClient(nodeClient *NodeClient, wsClient *WsClient) (*SFCClient, error) {
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

func (c *SFCClient) GetStakeInfoByBlock(ctx context.Context, fromBlock uint64, toBlock *uint64) ([]pkg.SFCStakeInfo, error) {
	var result = make([]pkg.SFCStakeInfo, 0)
	opts := &bind.FilterOpts{
		Context: ctx,
		Start:   fromBlock,
		End:     toBlock,
	}

	iterator, err := c.sfcContract.FilterLockedUpStake(opts, nil, nil)
	if err != nil {
		c.l.Warnw("get locked up stake error", "error", err)
		return nil, err
	}

	for iterator.Next() {
		if err := iterator.Error(); err != nil {
			c.l.Debugw("parse locked up stake error", "error", err)
			continue
		}
		result = append(result, pkg.ToSFCStakeInfo(iterator.Event))
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})
	return result, nil
}

func (c *SFCClient) WatchStakeEvent(ctx context.Context, stakeInfoCh chan<- pkg.SFCStakeInfo, errCh chan<- error) {
	opts := &bind.WatchOpts{
		Context: ctx,
	}
	var (
		sink = make(chan *contracts.SFCLockedUpStake)
	)
	sub, err := c.sfcWsContract.WatchLockedUpStake(opts, sink, nil, nil)
	if err != nil {
		c.l.Warnw("watch stake error", "error", err)
		errCh <- err
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case info := <-sink:
			stakeInfoCh <- pkg.ToSFCStakeInfo(info)
		case err := <-sub.Err():
			errCh <- err
			return
		}
	}
}

func (c *SFCClient) GetUnstakeInfoByBlock(ctx context.Context, fromBlock uint64, toBlock *uint64) ([]pkg.SFCUnstakeInfo, error) {
	var result = make([]pkg.SFCUnstakeInfo, 0)
	opts := &bind.FilterOpts{
		Context: ctx,
		Start:   fromBlock,
		End:     toBlock,
	}

	iterator, err := c.sfcContract.FilterUnlockedStake(opts, nil, nil)
	if err != nil {
		c.l.Warnw("get locked up stake error", "error", err)
		return nil, err
	}

	for iterator.Next() {
		if err := iterator.Error(); err != nil {
			c.l.Debugw("parse locked up stake error", "error", err)
			continue
		}
		result = append(result, pkg.ToSFCUnstakeInfo(iterator.Event))
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].BlockNumber < result[j].BlockNumber
	})
	return result, nil
}

func (c *SFCClient) WatchUnstakeEvent(ctx context.Context, unstakeInfoCh chan<- pkg.SFCUnstakeInfo, errCh chan<- error) {
	opts := &bind.WatchOpts{
		Context: ctx,
	}
	var (
		sink = make(chan *contracts.SFCUnlockedStake)
	)
	sub, err := c.sfcWsContract.WatchUnlockedStake(opts, sink, nil, nil)
	if err != nil {
		c.l.Warnw("watch unlocked stake error", "error", err)
		errCh <- err
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case info := <-sink:
			unstakeInfoCh <- pkg.ToSFCUnstakeInfo(info)
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
	sub, err := c.wsClient.client.SubscribeNewHead(ctx, headCh)
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
