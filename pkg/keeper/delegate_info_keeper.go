package keeper

import (
	"sync"

	"github.com/quangkeu95/fantom-bot/pkg"
)

type DelegateInfoKeeper struct {
	listInfo []pkg.SFCDelegateInfo
	mu       sync.RWMutex
}

func NewDelegateInfoKeeper() *DelegateInfoKeeper {
	return &DelegateInfoKeeper{
		listInfo: make([]pkg.SFCDelegateInfo, 0),
		mu:       sync.RWMutex{},
	}
}

func (k *DelegateInfoKeeper) Add(info pkg.SFCDelegateInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listInfo = append(k.listInfo, info)
}

func (k *DelegateInfoKeeper) AddBatch(infos []pkg.SFCDelegateInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listInfo = append(k.listInfo, infos...)
}

func (k *DelegateInfoKeeper) GetLast() (pkg.SFCDelegateInfo, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if len(k.listInfo) == 0 {
		return pkg.SFCDelegateInfo{}, false
	}
	return k.listInfo[len(k.listInfo)-1], true
}
