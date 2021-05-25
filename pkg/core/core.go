package core

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/quangkeu95/fantom-bot/lib/notification"
	"github.com/quangkeu95/fantom-bot/pkg"
	"github.com/quangkeu95/fantom-bot/pkg/fetcher"
	"github.com/quangkeu95/fantom-bot/pkg/keeper"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	MinStakingAmountFlag  = "min_staking_amount"
	MinClaimAmountFlag    = "min_claim_amount"
	MinTransferAmountFlag = "min_transfer_amount"
	ShiftBlock            = uint64(5)
)

type Core struct {
	l         *zap.SugaredLogger
	sfcClient *SFCClient
	fetchers  []fetcher.Fetcher

	validatorKeeper      *keeper.ValidatorsKeeper
	delegateInfoKeeper   *keeper.DelegateInfoKeeper
	undelegateInfoKeeper *keeper.UndelegateInfoKeeper
	rewardInfoKeeper     *keeper.RewardInfoKeeper

	socialBot notification.SocialBot

	minStakingAmount  float64
	minClaimAmount    float64
	minTransferAmount float64
}

func New() (*Core, error) {
	l := zap.S()
	nodeClient, err := fetcher.NewNodeClient()
	if err != nil {
		l.Errorw("error dial node client", "error", err)
		return nil, err
	}
	wsClient, err := fetcher.NewWsClient()
	if err != nil {
		l.Errorw("error dial ws client", "error", err)
		return nil, err
	}

	sfcClient, err := NewSFCClient(nodeClient, wsClient)
	if err != nil {
		l.Errorw("error initial SFC client", "error", err)
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

	fetchers := make([]fetcher.Fetcher, 0)
	graphqlClient, err := fetcher.NewGraphqlClient()
	if err != nil {
		l.Errorw("error initial GraphQL client", "error", err)
		return nil, err
	}
	fetchers = append(fetchers, graphqlClient, nodeClient)

	return &Core{
		l:                    l,
		sfcClient:            sfcClient,
		fetchers:             fetchers,
		validatorKeeper:      keeper.NewValidatorsKeeper(),
		delegateInfoKeeper:   keeper.NewDelegateInfoKeeper(),
		undelegateInfoKeeper: keeper.NewUndelegateInfoKeeper(),
		rewardInfoKeeper:     keeper.NewRewardInfoKeeper(),
		socialBot:            telegramBot,
		minStakingAmount:     minStakingAmount,
		minClaimAmount:       minClaimAmount,
		minTransferAmount:    minTransferAmount,
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
	go c.watchDelegateEvent(ctx)
	go c.watchUndelegateEvent(ctx)
	go c.watchClaimRewardEvent(ctx)
	go c.watchFTMTransferEvent(ctx)

	for {
	}
	return nil
}

func (c *Core) initFetchValidators(ctx context.Context) error {
	for _, f := range c.fetchers {
		validators, err := f.GetListValidators(ctx)
		if err != nil {
			continue
		}
		c.l.Debugw("fetch validators", "length", len(validators), "last_validator_id", validators[len(validators)-1].ID)

		c.validatorKeeper.AddBatch(validators)
		return nil
	}
	return fmt.Errorf("cannot fetch validators from any fetcher")
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
			c.sendCreatedValidatorMessage(validator)
		}
	}

}

func (c *Core) watchDelegateEvent(ctx context.Context) {
	var (
		delegateInfoCh = make(chan pkg.SFCDelegateInfo)
		errCh          = make(chan error)
	)

	c.l.Info("watch delegate event")

	go c.sfcClient.WatchDelegateEvent(ctx, delegateInfoCh, errCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			c.l.Warnw("reset WatchDelegateEvent subscription", "error", err)
			<-ticker.C
			go c.sfcClient.WatchDelegateEvent(ctx, delegateInfoCh, errCh)
		case info := <-delegateInfoCh:
			if info.Amount > c.minStakingAmount {
				c.l.Debugw("new delegate event", "tx_hash", info.TxHash)
				c.delegateInfoKeeper.Add(info)
				c.sendDelegateMessage(info)
			}
		}
	}
}

func (c *Core) watchUndelegateEvent(ctx context.Context) {
	var (
		undelegateInfoCh = make(chan pkg.SFCUndelegateInfo)
		errCh            = make(chan error)
	)

	c.l.Info("watch unstake event")

	go c.sfcClient.WatchUndelegateEvent(ctx, undelegateInfoCh, errCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errCh:
			c.l.Warnw("reset WatchUndelegateEvent subscription", "error", err)
			<-ticker.C
			go c.sfcClient.WatchUndelegateEvent(ctx, undelegateInfoCh, errCh)
		case info := <-undelegateInfoCh:
			if info.Amount > c.minStakingAmount {
				c.l.Debugw("new undelegate event", "tx_hash", info.TxHash)
				c.undelegateInfoKeeper.Add(info)
				c.sendUndelegateMessage(info)
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
		case info := <-rewardInfoCh:
			if info.UnlockedReward > c.minClaimAmount {
				c.l.Debugw("new claim reward event", "tx_hash", info.TxHash)
				c.rewardInfoKeeper.Add(info)
				c.sendClaimRewardMessage(info)
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
			blockNumber = blockNumber - ShiftBlock
		L1:
			for _, f := range c.fetchers {
				logs, err := f.GetListFTMTransferByBlock(ctx, blockNumber)
				if err != nil {
					continue
				}

				for _, item := range logs {
					if item.Amount > c.minTransferAmount {
						c.l.Debugw("new big transfer event", "tx_hash", item.TxHash, "block_number", item.BlockNumber)
						c.sendBigTransferMessage(item)
					}
				}
				break L1
			}

		}
	}
}

func (c *Core) sendCreatedValidatorMessage(item pkg.SFCValidator) {
	explorerEndpoint := viper.GetString("fantom_chain.explorer_tx_endpoint")
	msg := fmt.Sprintf("A new <a href=\"%s/address/%s\">created validator</a> with ID <b>%v</b> ",
		explorerEndpoint, item.Address, item.ID)

	if err := c.socialBot.SendMessage(msg); err != nil {
		c.l.Debugw("bot send message error", "error", err)
	}
}

func (c *Core) sendBigTransferMessage(item pkg.TransferLog) {
	explorerEndpoint := viper.GetString("fantom_chain.explorer_tx_endpoint")
	msg := fmt.Sprintf("%v Big <a href=\"%s/tx/%s\">transfer</a> of <b>%f FTM</b> from <code>%s</code> to <code>%s</code>",
		notification.EmojiWhale, explorerEndpoint, item.TxHash, item.Amount, item.From, item.To)

	if err := c.socialBot.SendMessage(msg); err != nil {
		c.l.Debugw("bot send message error", "error", err)
	}
}

func (c *Core) sendDelegateMessage(item pkg.SFCDelegateInfo) {
	explorerEndpoint := viper.GetString("fantom_chain.explorer_tx_endpoint")
	msg := fmt.Sprintf("%v An <a href=\"%s/tx/%s\">delegation event</a> of <b>%f FTM</b> from <code>%s</code> to validator ID <code>%d</code>",
		notification.EmojiCheckMark, explorerEndpoint, item.TxHash, item.Amount, item.Delegator, item.ToValidatorID)

	if err := c.socialBot.SendMessage(msg); err != nil {
		c.l.Debugw("bot send message error", "error", err)
	}
}

func (c *Core) sendUndelegateMessage(item pkg.SFCUndelegateInfo) {
	explorerEndpoint := viper.GetString("fantom_chain.explorer_tx_endpoint")
	msg := fmt.Sprintf("%v An <a href=\"%s/tx/%s\">undelegation event</a> of <b>%f FTM</b> from <code>%s</code> to validator ID <code>%d</code>",
		notification.EmojiCrossMark, explorerEndpoint, item.TxHash, item.Amount, item.Delegator, item.ToValidatorID)

	if err := c.socialBot.SendMessage(msg); err != nil {
		c.l.Debugw("bot send message error", "error", err)
	}
}

func (c *Core) sendClaimRewardMessage(item pkg.SFCRewardInfo) {
	explorerEndpoint := viper.GetString("fantom_chain.explorer_tx_endpoint")
	msg := fmt.Sprintf("%v An <a href=\"%s/tx/%s\">reward claim event</a> of <b>%f FTM</b> from <code>%s</code> to validator ID <code>%d</code>",
		notification.EmojiStar, explorerEndpoint, item.TxHash, item.UnlockedReward, item.Delegator, item.ToValidatorID)

	if err := c.socialBot.SendMessage(msg); err != nil {
		c.l.Debugw("bot send message error", "error", err)
	}
}
