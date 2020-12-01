// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package richtypes

import (
	"fmt"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/constants"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ContractType represents contract type
type ContractType string

// ContractElemType ...
type ContractElemType string

type ContractElem struct {
	ElemName string           `json:"elem_name"`
	ElemType ContractElemType `json:"elem_type"`
}

// ContractElemConcrete indicates contract element, such as events and methods
type ContractElemConcrete struct {
	// ElemName     string           `json:"elem_name"`
	Contract     *sdk.Contract `json:"contract,omitempty"`
	ContractType ContractType  `json:"contract_type"`
	// ElemType     ContractElemType `json:"elem_type"`
	ContractElem
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
	TransferFunction ContractElemType = "TransferFunction"
)

// Contract describe response contract information of scan rest api request
type Contract struct {
	Token
	ABI string `json:"abi"`

	// TypeCode      uint   `json:"typeCode"`
	// ContractName  string `json:"name"`

	// TokenSymbol   string `json:"tokenSymbol"`
	// TokenDecimals uint64 `json:"tokenDecimals"`

	// TokenName     string `json:"tokenName"`
}

// GetContractType return contract type
// func (c *Contract) GetContractType() ContractType {
// 	if c.TypeCode == 0 {
// 		return GENERAL
// 	}
// 	if c.TypeCode >= 100 && c.TypeCode < 200 {
// 		return ERC20
// 	}
// 	if c.TypeCode >= 200 && c.TypeCode < 300 {
// 		return ERC777
// 	}
// 	if c.TypeCode == 201 {
// 		return FANSCOIN
// 	}
// 	if c.TypeCode >= 500 && c.TypeCode < 600 {
// 		return ERC721
// 	}
// 	if c.TypeCode >= 1000 {
// 		return DEX
// 	}
// 	return UNKNOWN
// }

// GetContractTypeByABI acquires contract type by ABI
func (c *Contract) GetContractTypeByABI() ContractType {
	realContract, err := sdk.NewContract([]byte(c.ABI), nil, nil)
	if err != nil {
		return UNKNOWN
	}
	// method := realContract.ABI.Methods["0xa9059cbb"]
	erc20sign, _ := hexutil.Decode(constants.Erc20TransferFuncSign)
	method, err := realContract.ABI.MethodById(erc20sign)
	if err == nil && method != nil {
		return ERC20
	}

	erc777sign, _ := hexutil.Decode(constants.Erc777SendFuncSign)
	method, err = realContract.ABI.MethodById(erc777sign)
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

// DecodeEvent decodes log into instance of event params struct
func (contrete *ContractElemConcrete) DecodeEvent(log *types.LogEntry) (eventParmsPtr interface{}, err error) {
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

// DecodeFunction decodes packed data into instance of function params struct
func (contrete *ContractElemConcrete) DecodeFunction(data []byte) (functionParmsPtr interface{}, err error) {
	id := data[0:4]
	method, err := contrete.Contract.ABI.MethodById(id)
	if err != nil {
		return nil, err
	}

	switch contrete.ElemType {
	case TransferFunction:
		switch contrete.ContractType {
		case ERC20:
			params := ERC20TokenTransferFunctionParams{}
			err = method.Inputs.Unpack(&params, data[4:])
			functionParmsPtr = &params
			return
		case ERC777:
			params := ERC777TokenTransferFunctionParams{}
			err = method.Inputs.Unpack(&params, data[4:])
			functionParmsPtr = &params
			return
		}
	}
	return nil, fmt.Errorf("not found tuple type for contract type: %v, function type: %v", contrete.ContractType, contrete.ElemType)
}
