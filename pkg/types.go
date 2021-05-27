package pkg

import (
	"github.com/quangkeu95/fantom-bot/lib/contracts"
)

type SFCValidator struct {
	ID               uint64
	Address          string
	CreatedTime      uint64
	CreatedEpoch     uint64
	DeactivatedTime  uint64
	DeactivatedEpoch uint64
	IsActive         bool
	IsOffline        bool
}

func ToSFCValidator(v *contracts.SFCCreatedValidator) SFCValidator {
	return SFCValidator{
		ID:           v.ValidatorID.Uint64(),
		Address:      v.Auth.Hex(),
		CreatedTime:  v.CreatedTime.Uint64(),
		CreatedEpoch: v.CreatedEpoch.Uint64(),
	}
}

type SFCDelegateInfo struct {
	Delegator     string
	ToValidatorID uint64
	Amount        float64
	BlockNumber   uint64
	TxHash        string
}

func ToSFCDelegateInfo(v *contracts.SFCDelegated) SFCDelegateInfo {
	return SFCDelegateInfo{
		Delegator:     v.Delegator.Hex(),
		ToValidatorID: v.ToValidatorID.Uint64(),
		Amount:        WeiToFloat(v.Amount, 18),
		BlockNumber:   v.Raw.BlockNumber,
		TxHash:        v.Raw.TxHash.Hex(),
	}
}

type SFCUndelegateInfo struct {
	Delegator     string
	ToValidatorID uint64
	Amount        float64
	WrID          uint64
	BlockNumber   uint64
	TxHash        string
}

func ToSFCUndelegateInfo(v *contracts.SFCUndelegated) SFCUndelegateInfo {
	return SFCUndelegateInfo{
		Delegator:     v.Delegator.Hex(),
		ToValidatorID: v.ToValidatorID.Uint64(),
		Amount:        WeiToFloat(v.Amount, 18),
		WrID:          v.ToValidatorID.Uint64(),
		BlockNumber:   v.Raw.BlockNumber,
		TxHash:        v.Raw.TxHash.Hex(),
	}
}

type SFCRewardInfo struct {
	Delegator         string
	ToValidatorID     uint64
	LockupExtraReward float64
	LockupBaseReward  float64
	UnlockedReward    float64
	BlockNumber       uint64
	TxHash            string
}

func ToSFCRewardInfo(v *contracts.SFCClaimedRewards) SFCRewardInfo {
	return SFCRewardInfo{
		Delegator:         v.Delegator.Hex(),
		ToValidatorID:     v.ToValidatorID.Uint64(),
		LockupExtraReward: WeiToFloat(v.LockupExtraReward, 18),
		LockupBaseReward:  WeiToFloat(v.LockupBaseReward, 18),
		UnlockedReward:    WeiToFloat(v.UnlockedReward, 18),
		BlockNumber:       v.Raw.BlockNumber,
		TxHash:            v.Raw.TxHash.Hex(),
	}
}

type TransferLog struct {
	BlockNumber uint64
	TxHash      string
	From        string
	To          string
	Amount      float64
}

type SFCLockedUpStake struct {
	Delegator   string
	ValidatorID uint64
	Duration    uint64
	Amount      float64
	BlockNumber uint64
	TxHash      string
}

func ToSFCLockedUpStake(v *contracts.SFCLockedUpStake) SFCLockedUpStake {
	return SFCLockedUpStake{
		Delegator:   v.Delegator.Hex(),
		ValidatorID: v.ValidatorID.Uint64(),
		Duration:    v.Duration.Uint64(),
		Amount:      WeiToFloat(v.Amount, 18),
		BlockNumber: v.Raw.BlockNumber,
		TxHash:      v.Raw.TxHash.Hex(),
	}
}

type SFCUnlockedStake struct {
	Delegator   string
	ValidatorID uint64
	Amount      float64
	Penalty     float64
	BlockNumber uint64
	TxHash      string
}

func ToSFCUnlockedStake(v *contracts.SFCUnlockedStake) SFCUnlockedStake {
	return SFCUnlockedStake{
		Delegator:   v.Delegator.Hex(),
		ValidatorID: v.ValidatorID.Uint64(),
		Amount:      WeiToFloat(v.Amount, 18),
		Penalty:     WeiToFloat(v.Penalty, 18),
		BlockNumber: v.Raw.BlockNumber,
		TxHash:      v.Raw.TxHash.Hex(),
	}
}
