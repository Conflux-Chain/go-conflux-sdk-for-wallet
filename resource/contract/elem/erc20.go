package elem

import (
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
)

var erc20 []richtypes.ContractElem = []richtypes.ContractElem{
	{ElemName: "Transfer", ElemType: richtypes.TransferEvent},
	{ElemName: "name", ElemType: richtypes.NameFunction},
	{ElemName: "symbol", ElemType: richtypes.SymbolFunction},
	{ElemName: "decimals", ElemType: richtypes.DecimalsFunction},
	{ElemName: "transfer", ElemType: richtypes.TransferFunction},
}
