package exampletypes

import (
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	walletsdk "github.com/Conflux-Chain/go-conflux-sdk-for-wallet"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

type Config struct {
	NodeURL                string
	CfxScanBackendSchema   string
	CfxScanBackendAddress  string
	ContractManagerSchema  string
	ContractManagerAddress string
	ERC20Address           types.Address
	ERC777Address          types.Address
	NormalTransactions     []types.Hash
	ERC20Transactions      []types.Hash
	ERC777Transactions     []types.Hash

	client     *sdk.Client
	richClient *walletsdk.RichClient
}

func (c *Config) SetClient(client *sdk.Client) {
	c.client = client
}

func (c *Config) GetClient() *sdk.Client {
	return c.client
}

func (c *Config) SetRichClient(richClient *walletsdk.RichClient) {
	c.richClient = richClient
}

func (c *Config) GetRichClient() *walletsdk.RichClient {
	return c.richClient
}
