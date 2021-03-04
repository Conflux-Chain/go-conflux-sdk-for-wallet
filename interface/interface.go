package walletinterface

import (
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

//RichClientOperator represents rich client operator
type RichClientOperator interface {
	GetClient() sdk.ClientOperator
	GetAccountTokenTransfers(address types.Address, tokenIdentifier *types.Address, pageNumber, pageSize uint) (*richtypes.TokenTransferEventList, error)
	CreateSendTokenTransaction(from types.Address, to types.Address, amount *hexutil.Big, tokenIdentifier *types.Address) (*types.UnsignedTransaction, error)
	GetContractInfo(contractAddress types.Address, needABI, needIcon bool) (*richtypes.Contract, error)
	GetAccountTokens(account types.Address) (*richtypes.TokenWithBlanceList, error)
	GetTransactionsFromPool() (*[]types.Transaction, error)
}

// TokenReader ...
type TokenReader interface {
	GetTokenByIdentifier(contractAddress types.Address) (*richtypes.Token, error)
}

// EventDecoder represents interface for decoding event
type EventDecoder interface {
	Decode(log *types.Log) (interface{}, error)
}
