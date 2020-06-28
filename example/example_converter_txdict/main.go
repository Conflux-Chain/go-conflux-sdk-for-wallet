package main

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	walletsdk "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

var converter *walletsdk.TxDictConverter
var richClient *walletsdk.RichClient

var contractErc20Address = types.Address("0x8c3da77847b4efa454e6081dd4e898265d1787a2")
var contractErc777Address = types.Address("0x8726be94d7503b05f1738f026f00e74348c3d3eb")

func init() {

	//unlock account
	am := sdk.NewAccountManager("../keystore")
	err := am.TimedUnlockDefault("hello", 300*time.Second)
	if err != nil {
		panic(err)
	}

	//init client without retry and excute it
	//it doesn't work now, you could try later
	// url := "http://testnet-jsonrpc.conflux-chain.org:12537"
	url := "http://123.57.45.90:12537"

	client, err := sdk.NewClient(url)
	if err != nil {
		panic(err)
	}
	client.SetAccountManager(am)
	config := new(walletsdk.ServerConfig)

	// init rich client

	// main net
	// config.CfxScanBackendDomain = "47.102.164.229:8885"
	// config.ContractManagerDomain = "139.196.47.91:8886"

	// public test net (公共测试网)
	config.CfxScanBackendAddress = "testnet-scantest.confluxscan.io"
	config.ContractManagerAddress = "testnet-scantest.confluxscan.io/contract-manager"

	// private test net (内部测试网)
	// config.CfxScanBackendAddress = "101.201.103.131:8885"
	// config.ContractManagerAddress = "101.201.103.131:8886"

	richClient = walletsdk.NewRichClient(client, config)

	// c := walletsdk.NewTxDictConverter(rc, &walletsdk.TokenReaderByClient{Client: rc.GetClient()})
	// tr:=walletsdk.TokenReaderByClient
	converter, err = walletsdk.NewTxDictConverter(richClient)
	if err != nil {
		panic(err)
	}
}

func main() {
	testTxDictConverter()
}

func testTxDictConverter() {
	// tx, err := rc.GetClient().GetTransactionByHash("0xaf5156c4a6f1c9bc90af3f34da5c8d62497543ae49c107654d858d3b2c02a0f1")//对应的block 有 788 条交易，rpc client会返回错误，大bug
	tx, err := richClient.GetClient().GetTransactionByHash("0x86669c1d12e8d0882336b33f8d22d7b4bc7b4c92047a6c6c399c2ebbd16fb28e") //对应的block 有 45 条交易，rpc client会返回错误，大bug
	if err != nil {
		panic(err)
	}
	fmt.Printf("get transaction done: %+v\n\n", jsonFmt(tx))

	txdict, err := converter.ConvertByTransaction(tx, nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Convert tx \n%v\nto txdict done:\n%+v\n\n", jsonFmt(tx), jsonFmt(txdict))

	ttes, err := richClient.GetAccountTokenTransfers("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302", &contractErc20Address, 1, 10)
	if err != nil {
		panic(err)
	}
	// for _, tte := range ttes.List {
	tte := ttes.List[0]
	txdict, err = converter.ConvertByTokenTransferEvent(&tte)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Convert TokenTransferEvent \n%+v\nto txdict result:\n%+v\n\n", jsonFmt(tte), jsonFmt(txdict))
	// }

	unsignedTx, err := richClient.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &contractErc20Address)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("create send erc20 token tx:%+v\n\n", jsonFmt(unsignedTx))
	txdictBase := converter.ConvertByUnsignedTransaction(unsignedTx)
	fmt.Printf("Convert erc20 UnsignedTransaction \n%v\nto TxDictBase done:\n%+v\n\n", jsonFmt(unsignedTx), jsonFmt(txdictBase))

	unsignedTx, err = richClient.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &contractErc777Address)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("create send erc777 token tx:%+v\n\n", jsonFmt(unsignedTx))
	txdictBase = converter.ConvertByUnsignedTransaction(unsignedTx)
	fmt.Printf("Convert erc777 UnsignedTransaction \n%v\nto TxDictBase done:\n%+v\n\n", jsonFmt(unsignedTx), jsonFmt(txdictBase))
}

func jsonFmt(input interface{}) string {
	j, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	return string(j)
}
