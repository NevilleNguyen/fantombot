package core

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/quangkeu95/fantom-bot/lib/notification"
	"github.com/quangkeu95/fantom-bot/pkg"
	"github.com/quangkeu95/fantom-bot/pkg/keeper"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	MinStakingAmountFlag  = "min_staking_amount"
	MinClaimAmountFlag    = "min_claim_amount"
	MinTransferAmountFlag = "min_transfer_amount"
)

type Core struct {
	l             *zap.SugaredLogger
	sfcClient     *SFCClient
	graphqlClient *GraphqlAltClient

	validatorKeeper   *keeper.ValidatorsKeeper
	stakeInfoKeeper   *keeper.StakeInfoKeeper
	unstakeInfoKeeper *keeper.UnstakeInfoKeeper
	rewardInfoKeeper  *keeper.RewardInfoKeeper

	socialBot notification.SocialBot

	minStakingAmount  float64
	minClaimAmount    float64
	minTransferAmount float64
}

func New() (*Core, error) {
	l := zap.S()
	nodeClient, err := NewNodeClient()
	if err != nil {
		l.Errorw("error dial node client", "error", err)
		return nil, err
	}
	wsClient, err := NewWsClient()
	if err != nil {
		l.Errorw("error dial ws client", "error", err)
		return nil, err
	}

	sfcClient, err := NewSFCClient(nodeClient, wsClient)
	if err != nil {
		l.Errorw("error initial SFC client", "error", err)
		return nil, err
	}

	graphqlClient, err := NewGraphqlAltClient()
	if err != nil {
		l.Errorw("error initial GraphQL client", "error", err)
		return nil, err
	}

	telegramBot, err := notification.NewTelegramBot()
	if err != nil {
		l.Errorw("error initial social bot", "error", err)
		return nil, err
	}

	minStakingAmount := viper.GetFloat64(MinStakingAmountFlag)
	if err := validation.Validate(minStakingAmount, validation.Required); err != nil {
		l.Errorw("min staking amount config error", "error", err)
		return nil, err
	}
	minClaimAmount := viper.GetFloat64(MinClaimAmountFlag)
	if err := validation.Validate(minClaimAmount, validation.Required); err != nil {
		l.Errorw("min claim amount config error", "error", err)
		return nil, err
	}
	minTransferAmount := viper.GetFloat64(MinClaimAmountFlag)
	if err := validation.Validate(minClaimAmount, validation.Required); err != nil {
		l.Errorw("min claim amount config error", "error", err)
		return nil, err
	}

	return &Core{
		l:                 l,
		sfcClient:         sfcClient,
		graphqlClient:     graphqlClient,
		validatorKeeper:   keeper.NewValidatorsKeeper(),
		stakeInfoKeeper:   keeper.NewStakeInfoKeeper(),
		unstakeInfoKeeper: keeper.NewUnstakeInfoKeeper(),
		rewardInfoKeeper:  keeper.NewRewardInfoKeeper(),
		socialBot:         telegramBot,
		minStakingAmount:  minStakingAmount,
		minClaimAmount:    minClaimAmount,
		minTransferAmount: minTransferAmount,
	}, nil
}

func (c *Core) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.socialBot.SendMessage("fantom bot start")

	if err := c.initFetchValidators(ctx); err != nil {
		c.l.Warnw("fetch validators error", "error", err)
		return err
	}

	go c.watchCreatedValidators(ctx)
	go c.watchStakeEvent(ctx)
	go c.watchUnstakeEvent(ctx)
	go c.watchClaimRewardEvent(ctx)
	go c.watchFTMTransferEvent(ctx)

	for {
	}
	return nil
}

func (c *Core) initFetchValidators(ctx context.Context) error {
	validators, err := c.graphqlClient.GetListValidators(ctx)
	if err != nil {
		return err
	}
	c.l.Debugw("fetch validators", "length", len(validators), "last_validator_id", validators[len(validators)-1].ID)

	c.validatorKeeper.AddBatch(validators)
	return nil
}

func (c *Core) watchCreatedValidators(ctx context.Context) {
	var (
		validatorCh = make(chan pkg.SFCValidator)
		errCh       = make(chan error)
	)

	c.l.Info("watch created validators")

	go c.sfcClient.WatchCreatedValidator(ctx, validatorCh, errCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			c.l.Warnw("reset WatchCreatedValidator subscription", "error", err)
			<-ticker.C
			go c.sfcClient.WatchCreatedValidator(ctx, validatorCh, errCh)
		case validator := <-validatorCh:
			c.l.Debugw("new created validator", "validator_id", validator.ID)
			c.validatorKeeper.Add(validator)
			if err := c.socialBot.SendVariadicMessage("new created validator", validator.FormatListParams()...); err != nil {
				c.l.Debugw("bot send message error", "error", err)
			}
		}
	}

}

func (c *Core) watchStakeEvent(ctx context.Context) {
	var (
		stakeInfoCh = make(chan pkg.SFCStakeInfo)
		errCh       = make(chan error)
	)

	c.l.Info("watch stake event")

	go c.sfcClient.WatchStakeEvent(ctx, stakeInfoCh, errCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			c.l.Warnw("reset WatchStakeEvent subscription", "error", err)
			<-ticker.C
			go c.sfcClient.WatchStakeEvent(ctx, stakeInfoCh, errCh)
		case stakeInfo := <-stakeInfoCh:
			if stakeInfo.Amount > c.minStakingAmount {
				c.l.Debugw("new stake event", "tx_hash", stakeInfo.TxHash)
				c.stakeInfoKeeper.Add(stakeInfo)
				if err := c.socialBot.SendVariadicMessage("new stake event", stakeInfo.FormatListParams()...); err != nil {
					c.l.Debugw("bot send message error", "error", err)
				}
			}
		}
	}
}

func (c *Core) watchUnstakeEvent(ctx context.Context) {
	var (
		unstakeInfoCh = make(chan pkg.SFCUnstakeInfo)
		errCh         = make(chan error)
	)

	c.l.Info("watch unstake event")

	go c.sfcClient.WatchUnstakeEvent(ctx, unstakeInfoCh, errCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			c.l.Warnw("reset WatchUnstakeEvent subscription", "error", err)
			<-ticker.C
			go c.sfcClient.WatchUnstakeEvent(ctx, unstakeInfoCh, errCh)
		case unstakeInfo := <-unstakeInfoCh:
			if unstakeInfo.Amount > c.minStakingAmount {
				c.l.Debugw("new unstake event", "tx_hash", unstakeInfo.TxHash)
				c.unstakeInfoKeeper.Add(unstakeInfo)
				if err := c.socialBot.SendVariadicMessage("new unstake event", unstakeInfo.FormatListParams()...); err != nil {
					c.l.Debugw("bot send message error", "error", err)
				}
			}

		}
	}
}

func (c *Core) watchClaimRewardEvent(ctx context.Context) {
	var (
		rewardInfoCh = make(chan pkg.SFCRewardInfo)
		errCh        = make(chan error)
	)

	c.l.Info("watch claim reward event")

	go c.sfcClient.WatchClaimRewardEvent(ctx, rewardInfoCh, errCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			c.l.Warnw("reset WatchClaimRewardEvent subscription", "error", err)
			<-ticker.C
			go c.sfcClient.WatchClaimRewardEvent(ctx, rewardInfoCh, errCh)
		case rewardInfo := <-rewardInfoCh:
			if rewardInfo.UnlockedReward > c.minClaimAmount {
				c.l.Debugw("new claim reward event", "tx_hash", rewardInfo.TxHash)
				c.rewardInfoKeeper.Add(rewardInfo)
				if err := c.socialBot.SendVariadicMessage("new claim reward event", rewardInfo.FormatListParams()...); err != nil {
					c.l.Debugw("bot send message error", "error", err)
				}
			}
		}
	}
}

func (c *Core) watchFTMTransferEvent(ctx context.Context) {
	var (
		headCh = make(chan *types.Header)
		errCh  = make(chan error)
	)

	go c.sfcClient.SubscribeNewHead(ctx, headCh, errCh)

	c.l.Info("watch FTM transfer event")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			c.l.Warnw("reset WatchFTMTransferEvent subscription", "error", err)
			<-ticker.C
			go c.sfcClient.SubscribeNewHead(ctx, headCh, errCh)
		case head := <-headCh:
			blockNumber := head.Number.Uint64()
			// c.l.Debugw("new block", "block_number", blockNumber)
			logs, err := c.graphqlClient.GetListFTMTransferByBlock(ctx, blockNumber)
			if err != nil {
				c.l.Warnw("get list ftm transfer by block error", "error", err)
				continue
			}

			for _, item := range logs {
				if item.Amount > c.minTransferAmount {
					c.l.Debugw("new big transfer event", "tx_hash", item.TxHash, "block_number", item.BlockNumber)
					if err := c.socialBot.SendVariadicMessage("new big transfer event", item.FormatListParams()...); err != nil {
						c.l.Debugw("bot send message error", "error", err)
					}
				}
			}
		}
	}
}
