package fetcher

import (
	"context"

	"github.com/quangkeu95/fantom-bot/pkg"
)

type Fetcher interface {
	GetListValidators(ctx context.Context) ([]pkg.SFCValidator, error)
	GetListFTMTransferByBlock(ctx context.Context, blockNumber uint64) ([]pkg.TransferLog, error)
}
