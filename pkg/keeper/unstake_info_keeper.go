package keeper

import (
	"sync"

	"github.com/quangkeu95/fantom-bot/pkg"
)

type UnstakeInfoKeeper struct {
	listUnstakeInfo []pkg.SFCUnstakeInfo
	mu              sync.RWMutex
}

func NewUnstakeInfoKeeper() *UnstakeInfoKeeper {
	return &UnstakeInfoKeeper{
		listUnstakeInfo: make([]pkg.SFCUnstakeInfo, 0),
		mu:              sync.RWMutex{},
	}
}

func (k *UnstakeInfoKeeper) Add(info pkg.SFCUnstakeInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listUnstakeInfo = append(k.listUnstakeInfo, info)
}

func (k *UnstakeInfoKeeper) AddBatch(infos []pkg.SFCUnstakeInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listUnstakeInfo = append(k.listUnstakeInfo, infos...)
}

func (k *UnstakeInfoKeeper) GetLast() (pkg.SFCUnstakeInfo, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if len(k.listUnstakeInfo) == 0 {
		return pkg.SFCUnstakeInfo{}, false
	}
	return k.listUnstakeInfo[len(k.listUnstakeInfo)-1], true
}
