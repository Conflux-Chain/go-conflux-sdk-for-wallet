# API Reference
## Getting Started
The go-conflux-sdk-for-wallet module is a collection of packages which contain specific functionality for the wallet develop of conflux ecosystem.

- The package `walletsdk` offer complex APIs are provided through communicate with centralized server. currently, it is mainly for querying summary of user transactions and token transfer event.

## Installation
You can get Conflux Golang API For Wallet directly or use go module as below
```
go get github.com/Conflux-Chain/go-conflux-sdk-for-wallet
```
You can also add the Conflux Golang API For Wallet into vendor folder.
```
govendor fetch github.com/Conflux-Chain/go-conflux-sdk-for-wallet
```

After that you need to create a rich client instance with sdk.client and server config
```go
url:= "http://testnet-jsonrpc.conflux-chain.org:12537"
client, err := sdk.NewClient(url)
if err != nil {
	fmt.Println("new client error:", err)
	return
}
am := sdk.NewAccountManager("./keystore")
client.SetAccountManager(am)

config := new(richsdk.ServerConfig)
//main net
config.CfxScanBackendDomain = "47.102.164.229:8885"
config.ContractManagerDomain = "139.196.47.91:8886"
rc := richsdk.NewRichClient(client, config)
```
## package walletsdk
```
import "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"
```


### type RichClient

```go
type RichClient struct {
}
```

RichClient contains client, cfx-scan-backend server and contract-manager server

RichClient is the client for wallet, it's methods need request centralized
servers cfx-scan-backend and contract-manager in order to apply better
performance.

#### func  NewRichClient

```go
func NewRichClient(client sdk.ClientOperator, configOption *ServerConfig) *RichClient
```
NewRichClient create new rich client with client and server config.

The fields of config will use default value when it's empty

#### func (*RichClient) CreateSendTokenTransaction

```go
func (rc *RichClient) CreateSendTokenTransaction(from types.Address, to types.Address, amount *hexutil.Big, tokenIdentifier *types.Address) (*types.UnsignedTransaction, error)
```
CreateSendTokenTransaction creates unsigned transaction for sending token
according to input params, the tokenIdentifier represnets the token contract
address. It supports erc20, erc777, fanscoin at present

#### func (*RichClient) GetAccountTokenTransfers

```go
func (rc *RichClient) GetAccountTokenTransfers(address types.Address, tokenIdentifier *types.Address, pageNumber, pageSize uint) (*richtypes.TokenTransferEventList, error)
```
GetAccountTokenTransfers returns address releated transactions, the
tokenIdentifier represnets the token contract address and it is optional, when
tokenIdentifier is specicied it returns token transfer events related the
address, otherwise returns transactions about main coin.

#### func (*RichClient) GetAccountTokens

```go
func (rc *RichClient) GetAccountTokens(account types.Address) (*richtypes.TokenWithBlanceList, error)
```
GetAccountTokens returns coin balance and all token balances of specified
address

#### func (*RichClient) GetClient

```go
func (rc *RichClient) GetClient() sdk.ClientOperator
```
GetClient returns client

#### func (*RichClient) GetContractByIdentifier

```go
func (rc *RichClient) GetContractByIdentifier(tokenIdentifier types.Address) (*richtypes.Contract, error)
```
GetContractByIdentifier returns token detail infomation of token identifier

#### func (*RichClient) GetTransactionsFromPool

```go
func (rc *RichClient) GetTransactionsFromPool() (*[]types.Transaction, error)
```
GetTransactionsFromPool returns all pending transactions in mempool of conflux
node.

it only works on local conflux node currently.

#### func (*RichClient) GetTxDictByTxHash

```go
func (rc *RichClient) GetTxDictByTxHash(hash types.Hash) (*richtypes.TxDict, error)
```
GetTxDictByTxHash returns all cfx transfers and token transfers of transaction

#### func (*RichClient) GetTxDictsByEpoch

```go
func (rc *RichClient) GetTxDictsByEpoch(epoch *types.Epoch) ([]richtypes.TxDict, error)
```
GetTxDictsByEpoch returns all cfx transfers and token transfers of the epoch

### type ServerConfig

```go
type ServerConfig struct {
	CfxScanBackendSchema   string
	CfxScanBackendAddress  string
	ContractManagerSchema  string
	ContractManagerAddress string

	AccountBalancesPath    string
	AccountTokenTxListPath string
	TxListPath             string
	ContractQueryPath      string
}
```

ServerConfig represents cfx-scan-backend and contract-manager configurations,
because centralized servers maybe changed.

### type TxDictBaseConverter

```go
type TxDictBaseConverter struct {
}
```

TxDictBaseConverter contains methods for convert other types to TxDictBase.

#### func (*TxDictBaseConverter) ConvertByUnsignedTransaction

```go
func (tc *TxDictBaseConverter) ConvertByUnsignedTransaction(tx *types.UnsignedTransaction) *richtypes.TxDictBase
```
ConvertByUnsignedTransaction converts types.UnsignedTransaction to TxDictBase.

### type TxDictConverter

```go
type TxDictConverter struct {
}
```

TxDictConverter contains methods for convert other types to TxDict.

#### func  NewTxDictConverter

```go
func NewTxDictConverter(richClient walletinterface.RichClientOperator) (*TxDictConverter, error)
```
NewTxDictConverter creates a TxDictConverter instance.

#### func (*TxDictConverter) ConvertByTokenTransferEvent

```go
func (tc *TxDictConverter) ConvertByTokenTransferEvent(tte *richtypes.TokenTransferEvent) (*richtypes.TxDict, error)
```
ConvertByTokenTransferEvent converts richtypes.TokenTransferEvent to TxDict.

#### func (*TxDictConverter) ConvertByTransaction

```go
func (tc *TxDictConverter) ConvertByTransaction(tx *types.Transaction, revertRate *big.Float, blockTime *hexutil.Uint64) (*richtypes.TxDict, error)
```
ConvertByTransaction converts types.Transaction to TxDict.
