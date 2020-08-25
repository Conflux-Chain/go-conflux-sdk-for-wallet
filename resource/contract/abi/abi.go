package abi

import richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"

var ABIJsonDic map[richtypes.ContractType]string = make(map[richtypes.ContractType]string)

func init() {
	ABIJsonDic = make(map[richtypes.ContractType]string)
	ABIJsonDic[richtypes.ERC20] = erc20
	ABIJsonDic[richtypes.ERC777] = erc777
	ABIJsonDic[richtypes.ERC721] = erc721
}

// GetABI ...
func GetABI(contractType richtypes.ContractType) string {
	return ABIJsonDic[contractType]
}
