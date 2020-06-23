package richtypes

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type ERC20TokenTransferFunctionParams struct {
	To    common.Address
	Value *big.Int
}

type ERC777TokenTransferFunctionParams struct {
	To     common.Address
	Amount *big.Int
	Data   []byte
}
