package pkg

import "math/big"

func WeiToFloat(amountWei *big.Int, decimal int) float64 {
	amountWeiFloat := new(big.Float).SetInt(amountWei)
	decimalBigFloat := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil))
	amount := new(big.Float).Quo(amountWeiFloat, decimalBigFloat)
	rawResult, _ := amount.Float64()
	return rawResult
}
