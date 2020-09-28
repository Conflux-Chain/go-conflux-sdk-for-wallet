package main

import (
	"encoding/hex"
	"fmt"

	walletsdk "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/example/context"
	exampletypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/example/context/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var converter *walletsdk.TxDictConverter
var richClient *walletsdk.RichClient
var config *exampletypes.Config

func init() {
	config = context.Prepare()
	richClient = config.GetRichClient()

	var err error
	converter, err = walletsdk.NewTxDictConverter(richClient)
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("=======strat txdict converter example=======")
	testConvertByTransaction(config.NormalTransactions[0])
	testConvertByTransaction(config.ERC20Transactions[0])

	testConvertByTokenTransferEvent()
	testConvertByUnsignedTransaction()
	testConvertByUnsignedTransactionWithoutNetwork()
	fmt.Println("=======txdict converter example end!=========")
}

func testConvertByTransaction(hash types.Hash) {
	tx, err := richClient.GetClient().GetTransactionByHash(hash)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get transaction done: %+v\n\n", context.JsonFmt(tx))

	txdict, err := converter.ConvertByTransaction(tx, nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- Convert tx \n%v\nto txdict done:\n%+v\n\n", context.JsonFmt(tx), context.JsonFmt(txdict))
}

func testConvertByTokenTransferEvent() {
	ttes, err := richClient.GetAccountTokenTransfers("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302", &config.ERC20Address, 1, 10)
	if err != nil {
		panic(err)
	}
	// for _, tte := range ttes.List {
	tte := ttes.List[0]
	txdict, err := converter.ConvertByTokenTransferEvent(&tte)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- Convert TokenTransferEvent \n%+v\nto txdict result:\n%+v\n\n", context.JsonFmt(tte), context.JsonFmt(txdict))
}

func testConvertByUnsignedTransaction() {

	unsignedTx, err := richClient.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &config.ERC20Address)
	if err != nil {
		panic(err)
	}
	unsignedTx.Gas = types.NewBigInt(1000)
	unsignedTx.GasPrice = types.NewBigInt(10000)

	txdictBase := converter.ConvertByUnsignedTransaction(unsignedTx)
	fmt.Printf("- Convert erc20 UnsignedTransaction \n%v\nto TxDictBase done:\n%+v\n\n", context.JsonFmt(unsignedTx), context.JsonFmt(txdictBase))

	unsignedTx, err = richClient.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &config.ERC777Address)
	if err != nil {
		panic(err)
	}

	txdictBase = converter.ConvertByUnsignedTransaction(unsignedTx)
	fmt.Printf("- Convert erc777 UnsignedTransaction \n%v\nto TxDictBase done:\n%+v\n\n", context.JsonFmt(unsignedTx), context.JsonFmt(txdictBase))
}

func testConvertByUnsignedTransactionWithoutNetwork() {
	data, _ := hex.DecodeString("a9059cbb0000000000000000000000001a6048c1d81190c9a3555d0a06d0699663c4ddf0000000000000000000000000000000000000000000000000000000000000000a")
	unsignedTx := &types.UnsignedTransaction{
		UnsignedTransactionBase: types.UnsignedTransactionBase{
			From:         types.NewAddress("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"),
			Nonce:        types.NewBigInt(0x9),
			GasPrice:     types.NewBigInt(0x3b9aca00),
			Gas:          types.NewBigInt(0x8fb1),
			Value:        types.NewBigInt(0x0),
			StorageLimit: types.NewUint64(0x40),
			EpochHeight:  types.NewUint64(0x1eb1ea),
			ChainID:      types.NewUint(0x1)},
		To:   types.NewAddress("0x8c3da77847b4efa454e6081dd4e898265d1787a2"),
		Data: hexutil.Bytes(data),
	}

	txdictBase := converter.ConvertByUnsignedTransaction(unsignedTx)
	fmt.Printf("- Convert erc20 UnsignedTransaction \n%v\nto TxDictBase done:\n%+v\n\n", context.JsonFmt(unsignedTx), context.JsonFmt(txdictBase))
}
