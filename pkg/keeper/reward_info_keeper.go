package keeper

import (
	"sync"

	"github.com/quangkeu95/fantom-bot/pkg"
)

type RewardInfoKeeper struct {
	listRewardInfo []pkg.SFCRewardInfo
	mu             sync.RWMutex
}

func NewRewardInfoKeeper() *RewardInfoKeeper {
	return &RewardInfoKeeper{
		listRewardInfo: make([]pkg.SFCRewardInfo, 0),
		mu:             sync.RWMutex{},
	}
}

func (k *RewardInfoKeeper) Add(info pkg.SFCRewardInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listRewardInfo = append(k.listRewardInfo, info)
}

func (k *RewardInfoKeeper) AddBatch(infos []pkg.SFCRewardInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listRewardInfo = append(k.listRewardInfo, infos...)
}

func (k *RewardInfoKeeper) GetLast() (pkg.SFCRewardInfo, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if len(k.listRewardInfo) == 0 {
		return pkg.SFCRewardInfo{}, false
	}
	return k.listRewardInfo[len(k.listRewardInfo)-1], true
}
