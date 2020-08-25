package elem

import (
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
)

var erc777 []richtypes.ContractElem = []richtypes.ContractElem{
	{ElemName: "Sent", ElemType: richtypes.TransferEvent},
	{ElemName: "name", ElemType: richtypes.NameFunction},
	{ElemName: "symbol", ElemType: richtypes.SymbolFunction},
	{ElemName: "decimals", ElemType: richtypes.DecimalsFunction},
	{ElemName: "send", ElemType: richtypes.TransferFunction},
}
