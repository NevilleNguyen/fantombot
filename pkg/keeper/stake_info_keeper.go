package keeper

import (
	"sync"

	"github.com/quangkeu95/fantom-bot/pkg"
)

type StakeInfoKeeper struct {
	listStakeInfo []pkg.SFCStakeInfo
	mu            sync.RWMutex
}

func NewStakeInfoKeeper() *StakeInfoKeeper {
	return &StakeInfoKeeper{
		listStakeInfo: make([]pkg.SFCStakeInfo, 0),
		mu:            sync.RWMutex{},
	}
}

func (k *StakeInfoKeeper) Add(info pkg.SFCStakeInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listStakeInfo = append(k.listStakeInfo, info)
}

func (k *StakeInfoKeeper) AddBatch(infos []pkg.SFCStakeInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listStakeInfo = append(k.listStakeInfo, infos...)
}

func (k *StakeInfoKeeper) GetLast() (pkg.SFCStakeInfo, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if len(k.listStakeInfo) == 0 {
		return pkg.SFCStakeInfo{}, false
	}
	return k.listStakeInfo[len(k.listStakeInfo)-1], true
}
