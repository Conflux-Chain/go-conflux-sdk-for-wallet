# walletsdk
--
    import "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"


## Usage

#### type RichClient

```go
type RichClient struct {
        Client *sdk.Client
}
```

RichClient contains client, cfx-scan-backend server and contract-manager server

RichClient is the client for wallet, it's methods need request centralized
servers cfx-scan-backend and contract-manager in order to apply better
performance.

#### func  NewRichClient

```go
func NewRichClient(client *sdk.Client, configOption *ServerConfig) *RichClient
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

#### func (*RichClient) GetTokenByIdentifier

```go
func (rc *RichClient) GetTokenByIdentifier(tokenIdentifier types.Address) (*richtypes.Contract, error)
```
GetTokenByIdentifier returns token detail infomation of token identifier

#### func (*RichClient) GetTransactionsFromPool

```go
func (rc *RichClient) GetTransactionsFromPool() (*[]types.Transaction, error)
```
GetTransactionsFromPool returns all pending transactions in mempool of conflux
node.

#### type ServerConfig

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