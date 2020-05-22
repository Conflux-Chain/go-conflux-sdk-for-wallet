package main

import (
	"fmt"
	"math/big"
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	richsdk "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

var rc *richsdk.RichClient
var contractErc20Address = types.Address("0x8f1230f70d0984e29cb7b1d02547c361f85a93fa")
var contractErc777Address = types.Address("0x8726be94d7503b05f1738f026f00e74348c3d3eb")

func init() {

	//unlock account
	am := sdk.NewAccountManager("./keystore")
	err := am.TimedUnlockDefault("hello", 300*time.Second)
	if err != nil {
		panic(err)
	}

	//init client without retry and excute it
	url := "http://123.57.45.90:12537"

	//it doesn't work now, you could try later
	// url := "http://testnet-jsonrpc.conflux-chain.org:12537"
	client, err := sdk.NewClient(url)
	if err != nil {
		panic(err)
	}
	client.SetAccountManager(am)
	config := new(richsdk.ServerConfig)

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

	rc = richsdk.NewRichClient(client, config)
}

func main() {
	testGetAccountTokenTransfers()
	testGetAccountTokens()
	testCreateSendTokenTransaction()
	// testGetTransactionsFromPool()
	testGetTransactionsByEpoch()
	testGetTxDictByTxHash()
}

func testGetAccountTokenTransfers() {
	from := types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302")
	token := contractErc20Address
	tte, err := rc.GetAccountTokenTransfers(from, &token, 1, 50)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get account %v token %v transers done:\n%+v\n\n", from, token, tte)

	tte, err = rc.GetAccountTokenTransfers(from, nil, 1, 50)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get account %v main coin transers done:\n%+v\n\n", from, tte)
}

func testCreateSendTokenTransaction() {
	tx, err := rc.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &contractErc20Address)
	if err != nil {
		panic(err)
	}
	fmt.Printf("create send erc20 token tx:%+v\n\n", tx)

	tx, err = rc.CreateSendTokenTransaction(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"), types.Address("0x1a6048c1d81190c9a3555d0a06d0699663c4ddf0"), types.NewBigInt(10), &contractErc777Address)
	if err != nil {
		panic(err)
	}
	fmt.Printf("create send erc777 token tx:%+v\n\n", tx)
}

func testGetAccountTokens() {
	ts, err := rc.GetAccountTokens(types.Address("0x19f4bcf113e0b896d9b34294fd3da86b4adf0302"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("address has tokens:\n%+v\n\n", ts)

}

func testGetTransactionsFromPool() {
	txs, err := rc.GetTransactionsFromPool()
	if err != nil {
		panic(err)
	}
	fmt.Printf("txs in pool count is %+v\n\n", len(*txs))
}

func testGetTransactionsByEpoch() {
	start := time.Now()
	// epochNum := big.NewInt(1267420) //888 txs
	// epochNum := big.NewInt(2356824) //45 txs
	// epochNum := big.NewInt(2375610) //1 txs
	epochNum := big.NewInt(2478524) //4 txs

	txdicts, err := rc.GetTxDictsByEpoch(types.NewEpochNumber(epochNum))
	if err != nil {
		panic(err)
	}
	fmt.Printf("get txdicts of epoch %v is %+v, used time: %v\n\n", epochNum, txdicts, time.Now().Sub(start))
}

func testGetTxDictByTxHash() {
	hash := types.Hash("0xaed27380dcc0d96371d553d68811d6feffdbe3c2183c82128be53df7268a88a1")
	txdict, err := rc.GetTxDictByTxHash(hash)
	if err != nil {
		panic(err)
	}
	fmt.Printf("get txdict by txhash result: %+v\n\n", txdict)
}
