package richtypes

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// TokenTransferEventParams ...
type TokenTransferEventParams struct {
	From common.Address
	To   common.Address
}

// ERC20TokenTransferEventParams ...
type ERC20TokenTransferEventParams struct {
	TokenTransferEventParams
	Value *big.Int
}

// ERC777TokenTransferEventParams ...
type ERC777TokenTransferEventParams struct {
	TokenTransferEventParams
	Amount       *big.Int
	Operator     common.Address
	Data         []byte
	OperatorData []byte
}

// ERC721TokenTransferEventParams ...
type ERC721TokenTransferEventParams struct {
	TokenTransferEventParams
	TokenId *big.Int
}

// CreateEventParams ...
func CreateEventParams(contractType ContractType, eventType ContractElemType) (interface{}, error) {
	switch eventType {
	case TransferEvent:
		switch contractType {
		case ERC20:
			return &ERC20TokenTransferEventParams{}, nil
		case FANSCOIN:
			return &ERC20TokenTransferEventParams{}, nil
		case ERC721:
			return &ERC721TokenTransferEventParams{}, nil
		case ERC777:
			return &ERC777TokenTransferEventParams{}, nil
		}
	}
	return nil, fmt.Errorf("not found tuple type for contract type: %v, event type: %v", contractType, eventType)
}
