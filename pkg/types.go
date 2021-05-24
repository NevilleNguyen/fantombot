package pkg

import (
	"fmt"

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

func (v SFCValidator) Format() string {
	return fmt.Sprintf("validator_id: %v | validator_address: %v | created_epoch: %v | created_time %v",
		v.ID, v.Address, v.CreatedEpoch, v.CreatedTime)
}

func (v SFCValidator) FormatMap() map[string]interface{} {
	return map[string]interface{}{
		"validator_id":      v.ID,
		"address":           v.Address,
		"created_time":      v.CreatedTime,
		"created_epoch":     v.CreatedEpoch,
		"deactivated_time":  v.DeactivatedTime,
		"deactivated_epoch": v.DeactivatedEpoch,
		"is_active":         v.IsActive,
		"is_offline":        v.IsOffline,
	}
}

func (v SFCValidator) FormatListParams() []interface{} {
	var result = make([]interface{}, 0)
	result = append(result, "validator_id", v.ID,
		"address", v.Address,
		"created_time", v.CreatedTime,
		"created_epoch", v.CreatedEpoch,
		"deactivated_time", v.DeactivatedTime,
		"deactivated_epoch", v.DeactivatedEpoch,
		"is_active", v.IsActive,
		"is_offline", v.IsOffline)
	return result
}

type SFCStakeInfo struct {
	Delegator   string
	ValidatorID uint64
	Amount      float64
	BlockNumber uint64
	TxHash      string
}

func ToSFCStakeInfo(v *contracts.SFCLockedUpStake) SFCStakeInfo {
	return SFCStakeInfo{
		Delegator:   v.Delegator.Hex(),
		ValidatorID: v.ValidatorID.Uint64(),
		Amount:      WeiToFloat(v.Amount, 18),
		BlockNumber: v.Raw.BlockNumber,
		TxHash:      v.Raw.TxHash.Hex(),
	}
}

func (v SFCStakeInfo) FormatMap() map[string]interface{} {
	return map[string]interface{}{
		"delegator":    v.Delegator,
		"validator_id": v.ValidatorID,
		"amount":       v.Amount,
		"block_number": v.BlockNumber,
		"tx_hash":      v.TxHash,
	}
}

func (v SFCStakeInfo) FormatListParams() []interface{} {
	var result = make([]interface{}, 0)
	result = append(result, "delegator", v.Delegator,
		"validator_id", v.ValidatorID,
		"amount", v.Amount,
		"block_number", v.BlockNumber,
		"tx_hash", v.TxHash)
	return result
}

type SFCUnstakeInfo struct {
	Delegator   string
	ValidatorID uint64
	Amount      float64
	Penalty     float64
	BlockNumber uint64
	TxHash      string
}

func ToSFCUnstakeInfo(v *contracts.SFCUnlockedStake) SFCUnstakeInfo {
	return SFCUnstakeInfo{
		Delegator:   v.Delegator.Hex(),
		ValidatorID: v.ValidatorID.Uint64(),
		Amount:      WeiToFloat(v.Amount, 18),
		Penalty:     WeiToFloat(v.Amount, 18),
		BlockNumber: v.Raw.BlockNumber,
		TxHash:      v.Raw.TxHash.Hex(),
	}
}

func (v SFCUnstakeInfo) FormatMap() map[string]interface{} {
	return map[string]interface{}{
		"delegator":    v.Delegator,
		"validator_id": v.ValidatorID,
		"amount":       v.Amount,
		"penalty":      v.Penalty,
		"block_number": v.BlockNumber,
		"tx_hash":      v.TxHash,
	}
}

func (v SFCUnstakeInfo) FormatListParams() []interface{} {
	var result = make([]interface{}, 0)
	result = append(result, "delegator", v.Delegator,
		"validator_id", v.ValidatorID,
		"amount", v.Amount,
		"penalty", v.Penalty,
		"block_number", v.BlockNumber,
		"tx_hash", v.TxHash)
	return result
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

func (v SFCRewardInfo) FormatMap() map[string]interface{} {
	return map[string]interface{}{
		"delegator":           v.Delegator,
		"to_validator_id":     v.ToValidatorID,
		"lockup_extra_reward": v.LockupExtraReward,
		"lockup_base_reward":  v.LockupBaseReward,
		"unlocked_reward":     v.UnlockedReward,
		"block_number":        v.BlockNumber,
		"tx_hash":             v.TxHash,
	}
}

func (v SFCRewardInfo) FormatListParams() []interface{} {
	var result = make([]interface{}, 0)
	result = append(result, "delegator", v.Delegator,
		"to_validator_id", v.ToValidatorID,
		"lockup_extra_reward", v.LockupExtraReward,
		"lockup_base_reward", v.LockupBaseReward,
		"unlocked_reward", v.UnlockedReward,
		"block_number", v.BlockNumber,
		"tx_hash", v.TxHash)
	return result
}

type TransferLog struct {
	BlockNumber uint64
	TxHash      string
	From        string
	To          string
	Amount      float64
}

func (v TransferLog) FormatMap() map[string]interface{} {
	return map[string]interface{}{
		"block_number": v.BlockNumber,
		"tx_hash":      v.TxHash,
		"from":         v.From,
		"to":           v.To,
		"amount":       v.Amount,
	}
}

func (v TransferLog) FormatListParams() []interface{} {
	var result = make([]interface{}, 0)
	result = append(result, "block_number", v.BlockNumber,
		"tx_hash", v.TxHash,
		"from", v.From,
		"to", v.To,
		"amount", v.Amount)
	return result
}
