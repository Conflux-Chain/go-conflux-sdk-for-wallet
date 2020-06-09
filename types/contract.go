// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package richtypes

import (
	"fmt"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

// ContractType represents contract type
type ContractType string

// ContractElemType ...
type ContractElemType string

// ContractElemConcrete indicates contract element, such as events and methods
type ContractElemConcrete struct {
	ElemName     string           `json:"elem_name"`
	Contract     *sdk.Contract    `json:"contract,omitempty"`
	ContractType ContractType     `json:"contract_type"`
	ElemType     ContractElemType `json:"elem_type"`
}

const (
	UNKNOWN  ContractType = "UNKNOWN"
	GENERAL  ContractType = "GENERAL"
	ERC20    ContractType = "ERC20"
	ERC777   ContractType = "ERC777"
	FANSCOIN ContractType = "FANSCOIN"
	ERC721   ContractType = "ERC721"
	DEX      ContractType = "DEX"
)

const (
	TransferEvent    ContractElemType = "TransferEvent"
	NameFunction     ContractElemType = "NameFunction"
	SymbolFunction   ContractElemType = "SymbolFunction"
	DecimalsFunction ContractElemType = "DecimalsFunction"
)

// Contract describe response contract information of scan rest api request
type Contract struct {
	TypeCode      uint   `json:"typeCode"`
	ContractName  string `json:"name"`
	ABI           string `json:"abi"`
	TokenSymbol   string `json:"tokenSymbol"`
	TokenDecimals uint64 `json:"tokenDecimals"`
	TokenIcon     string `json:"tokenIcon"`
	TokenName     string `json:"tokenName"`
}

// GetContractType return contract type
func (c *Contract) GetContractType() ContractType {
	if c.TypeCode == 0 {
		return GENERAL
	}
	if c.TypeCode >= 100 && c.TypeCode < 200 {
		return ERC20
	}
	if c.TypeCode >= 200 && c.TypeCode < 300 {
		return ERC777
	}
	if c.TypeCode == 201 {
		return FANSCOIN
	}
	if c.TypeCode >= 500 && c.TypeCode < 600 {
		return ERC721
	}
	if c.TypeCode >= 1000 {
		return DEX
	}
	return UNKNOWN
}

// GetContractTypeByABI acquires contract type by ABI
func (c *Contract) GetContractTypeByABI() ContractType {
	realContract, err := sdk.NewContract([]byte(c.ABI), nil, nil)
	if err != nil {
		return UNKNOWN
	}
	// method := realContract.ABI.Methods["0xa9059cbb"]

	method, err := realContract.ABI.MethodById([]byte{0xa9, 0x05, 0x9c, 0xbb})
	if err == nil && method != nil {
		return ERC20
	}

	method, err = realContract.ABI.MethodById([]byte{0x9b, 0xd9, 0xbb, 0xc6})
	if err == nil && method != nil {
		return ERC777
	}
	return UNKNOWN
}

// // String implements the fmt.Stringer interface
// func (c ContractType) String() string {
// 	dic := make(map[ContractType]string)
// 	dic[UNKNOWN] = "unknown"
// 	dic[GENERAL] = "general"
// 	dic[ERC20] = "erc20"
// 	dic[ERC777] = "erc777"
// 	dic[FANSCOIN] = "fanscoin"
// 	dic[ERC721] = "erc721"
// 	dic[DEX] = "dex"
// 	return dic[c]
// }

// Decode decodes log into instance of event params struct
func (contrete *ContractElemConcrete) Decode(log *types.LogEntry) (eventParmsPtr interface{}, err error) {
	switch contrete.ElemType {
	case TransferEvent:
		switch contrete.ContractType {
		case ERC20:
			fallthrough
		case FANSCOIN:
			params := ERC20TokenTransferEventParams{}
			err = contrete.Contract.DecodeEvent(&params, contrete.ElemName, *log)
			eventParmsPtr = &params
			return
		case ERC721:
			params := ERC721TokenTransferEventParams{}
			err = contrete.Contract.DecodeEvent(&params, contrete.ElemName, *log)
			eventParmsPtr = &params
			return
		case ERC777:
			params := ERC777TokenTransferEventParams{}
			err = contrete.Contract.DecodeEvent(&params, contrete.ElemName, *log)
			eventParmsPtr = &params
			return
		}
	}

	return nil, fmt.Errorf("not found tuple type for contract type: %v, event type: %v", contrete.ContractType, contrete.ElemType)
}
