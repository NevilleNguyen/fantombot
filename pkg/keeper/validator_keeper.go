package keeper

import (
	"sync"

	"github.com/quangkeu95/fantom-bot/pkg"
)

type ValidatorsKeeper struct {
	listValidators map[uint64]pkg.SFCValidator
	mu             sync.RWMutex
}

func NewValidatorsKeeper() *ValidatorsKeeper {
	return &ValidatorsKeeper{
		listValidators: make(map[uint64]pkg.SFCValidator),
		mu:             sync.RWMutex{},
	}
}

func (k *ValidatorsKeeper) Add(v pkg.SFCValidator) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.listValidators[v.ID] = v
}

func (k *ValidatorsKeeper) AddBatch(listV []pkg.SFCValidator) {
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, v := range listV {
		k.listValidators[v.ID] = v
	}
}

func (k *ValidatorsKeeper) GetListValidators() map[uint64]pkg.SFCValidator {
	k.mu.RLock()
	defer k.mu.RUnlock()
	var result = make(map[uint64]pkg.SFCValidator)
	for key, value := range k.listValidators {
		result[key] = value
	}
	return result
}
