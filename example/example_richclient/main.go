package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	richsdk "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"
	context "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/example/context"
	exampletypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/example/context/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

var rc *richsdk.RichClient
var config *exampletypes.Config

func init() {
	config = context.Prepare()
	rc = config.GetRichClient()
}

func main() {
	fmt.Println("\n=======start rich client example=======\n")
	testGetAccountTokenTransfers()
	testGetAccountTokens()
	testCreateSendTokenTransaction()
	testGetTransactionsByEpoch()
	testGetTxDictByTxHash()
	testGetContractInfo()
	testGetTransactionsFromPool()
	fmt.Println("\n=======rich client example done=======\n")
}

func readConfig() {
	var path string = "../config.toml"
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Fatal(err)
	}
}

func testGetAccountTokenTransfers() {
	start := time.Now()
	from := types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302")
	token := config.ERC20Address
	tteList, err := rc.GetAccountTokenTransfers(from, &token, 1, 50)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- get account %v token %v transers done, token transfer list is:\n%+v\nused time:%v\n\n",
		from, token, context.JsonFmt(tteList), time.Now().Sub(start))

	start = time.Now()
	tteList, err = rc.GetAccountTokenTransfers(from, nil, 1, 50)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- get account %v CFX coin transers done, token transfer list is:\n%+v\nused time:%v\n\n",
		from, context.JsonFmt(tteList), time.Now().Sub(start))
}

func testCreateSendTokenTransaction() {
	tx, err := rc.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &config.ERC20Address)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- create send erc20 token tx:%+v\n\n", context.JsonFmt(tx))

	tx, err = rc.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &config.ERC777Address)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- create send erc777 token tx:%+v\n\n", context.JsonFmt(tx))
}

func testGetAccountTokens() {
	ts, err := rc.GetAccountTokens(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("- address has tokens:\n%+v\n\n", context.JsonFmt(ts))

}

func testGetTransactionsFromPool() {
	txs, err := rc.GetTransactionsFromPool()
	if err != nil {
		panic(err)
	}
	fmt.Printf("- txs in pool is:\n%+v\n\n", context.JsonFmt(txs))
}

func testGetTransactionsByEpoch() {
	start := time.Now()
	// epochNum := big.NewInt(7906525)

	epochNum, err := rc.GetClient().GetEpochNumber()

	txdicts, err := rc.GetTxDictsByEpoch(types.NewEpochNumber(epochNum))
	if err != nil {
		panic(err)
	}
	fmt.Printf("- get txdicts of epoch %v done, txidcts is\n%+v, used time: %v\n\n", epochNum, context.JsonFmt(txdicts), time.Now().Sub(start))
}

func testGetTxDictByTxHash() {
	hash := types.Hash(config.ERC20Transactions[0])
	txdict, err := rc.GetTxDictByTxHash(hash)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- get txdict by txhash done\n%+v\n\n", context.JsonFmt(txdict))
}

func testGetContractInfo() {
	tokenInfo, err := rc.GetContractInfo(config.ERC20Address, true, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("- get token info of %s done\n%v\n\n", config.ERC20Address, context.JsonFmt(tokenInfo))
}
