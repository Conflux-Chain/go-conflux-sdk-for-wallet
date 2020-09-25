package context

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path"
	"runtime"

	"github.com/BurntSushi/toml"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	walletsdk "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"
	exampletypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/example/context/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var client *sdk.Client
var (
	currentDir     string
	configPath     string
	erc20Contract  *sdk.Contract
	erc777Contract *sdk.Contract
	am             *sdk.AccountManager
	defaultAccount *types.Address
	nextNonce      *big.Int
)
var config exampletypes.Config

func Prepare() *exampletypes.Config {
	fmt.Println("=======start prepare config===========\n")
	getConfig()
	initClient()
	initRichClient()
	deployContracts()
	sendCfx()
	sendTokens()
	saveConfig()
	fmt.Println("=======prepare config done!===========\n")
	return &config
}

func getConfig() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("get current file path error")
	}
	currentDir = path.Join(filename, "../")
	configPath = path.Join(currentDir, "config.toml")
	// cp := make(map[string]string)
	config = exampletypes.Config{}
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("- to get config done: %+v\n", JsonFmt(config))
}

func initClient() {
	// url := "http://testnet-jsonrpc.conflux-chain.org:12537"
	var err error
	client, err = sdk.NewClient(config.NodeURL)
	if err != nil {
		panic(err)
	}

	am = sdk.NewAccountManager(path.Join(currentDir, "keystore"))
	client.SetAccountManager(am)
	defaultAccount, err = am.GetDefault()
	if err != nil {
		panic(err)
	}

	am.UnlockDefault("hello")
	nextNonce, err = client.GetNextNonce(*defaultAccount, nil)
	if err != nil {
		panic(err)
	}
	config.SetClient(client)
	fmt.Println("- to init client done")
}

func initRichClient() {

	// init rich client
	serverConfig := new(walletsdk.ServerConfig)

	// public test net (公共测试网)
	serverConfig.CfxScanBackendSchema = config.CfxScanBackendSchema
	serverConfig.CfxScanBackendAddress = config.CfxScanBackendAddress
	serverConfig.ContractManagerSchema = config.ContractManagerSchema
	serverConfig.ContractManagerAddress = config.ContractManagerAddress

	// main net
	// config.CfxScanBackendDomain = "47.102.164.229:8885"
	// config.ContractManagerDomain = "139.196.47.91:8886"

	// private test net (内部测试网)
	// config.CfxScanBackendAddress = "101.201.103.131:8885"
	// config.ContractManagerAddress = "101.201.103.131:8886"

	rc := walletsdk.NewRichClient(client, serverConfig)

	config.SetRichClient(rc)
	fmt.Println("- to init rich client done")

}

func deployContracts() {
	// check erc20 and erc777 address, if len !==42 or getcode error, deploy
	erc20Contract = deployIfNotExist(config.ERC20Address, path.Join(currentDir, "contract/erc20.abi"), path.Join(currentDir, "contract/erc20.bytecode"))
	erc777Contract = deployIfNotExist(config.ERC777Address, path.Join(currentDir, "contract/erc777.abi"), path.Join(currentDir, "contract/erc777.bytecode"))
	fmt.Println("- to deploy contracts if not exist done")
}

func sendCfx() {
	utx, err := client.CreateUnsignedTransaction(*defaultAccount, types.Address("0x10697db19a51514f83a7cc00cea2db0676724270"), types.NewBigInt(100), nil)
	if err != nil {
		panic(err)
	}
	txhash, err := client.SendTransaction(utx)
	config.NormalTransactions = []types.Hash{txhash}
	WaitPacked(txhash)
	fmt.Println("- to send normal txs done")
}

func sendTokens() {

	batchSend := func(contractType string, contract *sdk.Contract, txs []types.Hash) []types.Hash {
		// nextNonce := startNonce
		if txs == nil {
			txs = make([]types.Hash, 0)
		}
		for i := int64(0); i < 5; i++ {
			if len(txs) <= int(i) {
				txs = append(txs, types.Hash("0x"))
			}
			tx, err := client.GetTransactionByHash(txs[i])
			if err != nil || tx == nil || *tx.To != *contract.Address {
				to := types.Address("0x10697db19a51514f83a7cc00cea2db0676724270")
				options := &types.ContractMethodSendOption{
					Nonce: getNextNonceAndIncrease(),
				}
				var txhash *types.Hash
				switch contractType {
				case "ERC20":
					txhash, err = contract.SendTransaction(options, "transfer", to.ToCommonAddress(), big.NewInt(1))
				case "ERC777":
					txhash, err = contract.SendTransaction(options, "send", to.ToCommonAddress(), big.NewInt(1), []byte{})
				default:
					panic("unrecognized contract type:" + contractType)
				}

				// nextNonce = nextNonce.Add(nextNonce, big.NewInt(1))
				if err != nil {
					panic(err)
				}
				txs[i] = *txhash
				fmt.Printf("send %v transfer done: %v\n", contractType, txhash)
			}
		}
		return txs
	}

	// if ERC20Transaction not exist, send a erc20 transaction
	// nonce, err := client.GetNextNonce(*defaultAccount, nil)
	// if err != nil {
	// 	panic(err)
	// }
	config.ERC20Transactions = batchSend("ERC20", erc20Contract, config.ERC20Transactions)
	config.ERC777Transactions = batchSend("ERC777", erc777Contract, config.ERC777Transactions)
	WaitPacked(config.ERC777Transactions[4])
	fmt.Println("- to send tokens if tx not valid done")
}

func saveConfig() {
	f, err := os.OpenFile(configPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	config.ERC20Address = *erc20Contract.Address
	config.ERC777Address = *erc777Contract.Address
	encoder := toml.NewEncoder(f)
	err = encoder.Encode(config)
	if err != nil {
		panic(err)
	}
	fmt.Println("- to save config done")
}

func deployIfNotExist(contractAddress types.Address, abiFilePath string, bytecodeFilePath string) *sdk.Contract {
	isAddress := len(contractAddress) == 42 && (contractAddress)[0:2] == "0x"
	isCodeExist := false

	if isAddress {
		code, err := client.GetCode(contractAddress)
		// fmt.Printf("err: %v,code:%v\n", err, len(code))
		if err == nil && len(code) > 0 && code != "0x" {
			isCodeExist = true
		}
	}

	fmt.Printf("%v isAddress:%v, isCodeExist:%v\n", contractAddress, isAddress, isCodeExist)
	if isAddress && isCodeExist {
		abi, err := ioutil.ReadFile(abiFilePath)
		if err != nil {
			panic(err)
		}
		contract, err := client.GetContract(abi, &contractAddress)
		if err != nil {
			panic(err)
		}
		return contract
	}

	contract := deployContractWithConstroctor(abiFilePath, bytecodeFilePath, big.NewInt(100000), "biu", uint8(10), "BIU")
	return contract
}

func deployContractWithConstroctor(abiFile string, bytecodeFile string, params ...interface{}) *sdk.Contract {
	fmt.Println("start deploy contract with construcotr")
	abi, err := ioutil.ReadFile(abiFile)
	if err != nil {
		panic(err)
	}

	bytecodeHexStr, err := ioutil.ReadFile(bytecodeFile)
	if err != nil {
		panic(err)
	}

	bytecode, err := hex.DecodeString(string(bytecodeHexStr))
	if err != nil {
		panic(err)
	}

	option := types.ContractDeployOption{}
	option.Nonce = getNextNonceAndIncrease()
	result := client.DeployContract(&option, abi, bytecode, params...)

	_ = <-result.DoneChannel
	if result.Error != nil {
		panic(result.Error)
	}
	contract := result.DeployedContract
	fmt.Printf("deploy contract with abi: %v, bytecode: %v done\ncontract address: %+v\ntxhash:%v\n\n", abiFile, bytecodeFile, contract.Address, result.TransactionHash)

	return contract
}

func getNextNonceAndIncrease() *hexutil.Big {
	// println("current in:", nextNonce.String())
	currentNonce := big.NewInt(0).SetBytes(nextNonce.Bytes())
	nextNonce = nextNonce.Add(nextNonce, big.NewInt(1))
	// println("current out:", currentNonce.String())
	return types.NewBigIntByRaw(currentNonce)
}
