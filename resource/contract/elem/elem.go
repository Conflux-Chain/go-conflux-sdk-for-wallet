package elem

import (
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
)

var ContractType2ElemMetasMap map[richtypes.ContractType][]richtypes.ContractElem

func init() {
	ContractType2ElemMetasMap = make(map[richtypes.ContractType][]richtypes.ContractElem)
	ContractType2ElemMetasMap["ERC20"] = erc20
	ContractType2ElemMetasMap["ERC721"] = erc721
	ContractType2ElemMetasMap["ERC777"] = erc777
}

// GetContractElems ...
func GetContractElems(contractType richtypes.ContractType) []richtypes.ContractElem {
	return ContractType2ElemMetasMap[contractType]
}
