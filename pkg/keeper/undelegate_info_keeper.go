package keeper

import (
	"sync"

	"github.com/quangkeu95/fantom-bot/pkg"
)

type UndelegateInfoKeeper struct {
	listInfo []pkg.SFCUndelegateInfo
	mu       sync.RWMutex
}

func NewUndelegateInfoKeeper() *UndelegateInfoKeeper {
	return &UndelegateInfoKeeper{
		listInfo: make([]pkg.SFCUndelegateInfo, 0),
		mu:       sync.RWMutex{},
	}
}

func (k *UndelegateInfoKeeper) Add(info pkg.SFCUndelegateInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listInfo = append(k.listInfo, info)
}

func (k *UndelegateInfoKeeper) AddBatch(infos []pkg.SFCUndelegateInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listInfo = append(k.listInfo, infos...)
}

func (k *UndelegateInfoKeeper) GetLast() (pkg.SFCUndelegateInfo, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if len(k.listInfo) == 0 {
		return pkg.SFCUndelegateInfo{}, false
	}
	return k.listInfo[len(k.listInfo)-1], true
}
