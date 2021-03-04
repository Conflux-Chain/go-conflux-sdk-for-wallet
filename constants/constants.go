package constants

import (
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
)

const (
	// RPCConcurrence represents rpc request concurrence with conflux full node
	RPCConcurrence = 10
)

var (
	// erc20, err = sdk.NewContract(nil, nil, nil)
	Erc20TransferFuncSign = "0xa9059cbb"
	Erc777SendFuncSign    = "0x9bd9bbc6"

	// TethysFcV1Address represents Tethys Fc Contract Address
	TethysFcV1Address   = cfxaddress.MustNewFromHex("0x8e2f2e68eb75bb8b18caafe9607242d4748f8d98", 1029)
	Erc777SentEventSign = types.Hash("0x06b541ddaa720db2b10a4d0cdac39b8d360425fc073085fac19bc82614677987")
)
