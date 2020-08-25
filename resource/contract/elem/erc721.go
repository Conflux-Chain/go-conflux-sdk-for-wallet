package elem

import (
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
)

var erc721 []richtypes.ContractElem = []richtypes.ContractElem{
	{ElemName: "Transfer", ElemType: richtypes.TransferEvent},
	{ElemName: "name", ElemType: richtypes.NameFunction},
	{ElemName: "symbol", ElemType: richtypes.SymbolFunction},
	{ElemName: "safeTransferFrom", ElemType: richtypes.TransferFunction},
}
