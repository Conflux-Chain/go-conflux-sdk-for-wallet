package walletsdk

import (
	"errors"
	"fmt"
	"math/big"

	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/constants"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

// TxDictConverter contains methods for convert other types to TxDict
type TxDictConverter struct {
	richClient RichClientOperator
}

// NewTxDictConverter creates a TxDictConverter instance
func NewTxDictConverter(rc RichClientOperator) *TxDictConverter {
	return &TxDictConverter{
		richClient: rc,
	}
}

// TxDictBaseConverter contains methods for convert other types to TxDictBase
type TxDictBaseConverter struct {
}

// ConvertByRichTransaction convert richtypes.Transaction to TxDict
func (tc *TxDictConverter) ConvertByRichTransaction(tx *richtypes.Transaction) (*richtypes.TxDict, error) {
	txDict := new(richtypes.TxDict)
	txDict.TxHash = tx.Hash
	txDict.BlockHash = &tx.BlockHash
	txDict.TxAt = tx.Timestamp

	client := tc.richClient.GetClient()
	if client == nil {
		msg := "could not GetBlockRevertRateByHash because of client is nil"
		return nil, errors.New(msg)
	}

	revertRate, err := client.GetBlockRevertRateByHash(tx.BlockHash)
	if err != nil {
		msg := fmt.Sprintf("get block revert rate by hash %v error", tx.BlockHash)
		return nil, types.WrapError(err, msg)
	}
	txDict.RevertRate = revertRate

	value, ok := new(big.Int).SetString(tx.Value, 0)
	if !ok {
		msg := fmt.Sprintf("Convert tx.Value %v to *big.Int fail", tx.Value)
		return nil, errors.New(msg)
	}

	txDict.Inputs = []richtypes.TxUnit{
		{
			Value:        value,
			Address:      &tx.From,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}

	txDict.Outputs = []richtypes.TxUnit{
		{
			Value:        value,
			Address:      &tx.To,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}

	return txDict, nil
}

// ConvertByTokenTransferEvent converts richtypes.TokenTransferEvent to TxDict
func (tc *TxDictConverter) ConvertByTokenTransferEvent(tte *richtypes.TokenTransferEvent) (*richtypes.TxDict, error) {
	txDict := new(richtypes.TxDict)
	txDict.TxHash = tte.TransactionHash
	txDict.BlockHash = &tte.BlockHash
	txDict.RevertRate = tte.RevertRate
	txDict.TxAt = tte.Timestamp

	value, ok := new(big.Int).SetString(tte.Value, 0)
	if !ok {
		msg := fmt.Sprintf("Convert TokenTransferEvent.Value %v to *big.Int fail", tte.Value)
		return nil, errors.New(msg)
	}

	txDict.Inputs = []richtypes.TxUnit{
		{
			Value:        value,
			Address:      &tte.From,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}

	txDict.Outputs = []richtypes.TxUnit{
		{
			Value:        value,
			Address:      &tte.To,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}
	return txDict, nil
}

// ConvertByTransaction converts types.Transaction to TxDict
func (tc *TxDictConverter) ConvertByTransaction(tx *types.Transaction) (*richtypes.TxDict, error) {
	txDict := new(richtypes.TxDict)
	txDict.TxHash = tx.Hash
	txDict.BlockHash = tx.BlockHash
	// txDict.RevertRate = tc.rc.Client.GetBlockRevertRateByHash(tx.BlockHash)

	client := tc.richClient.GetClient()
	if client == nil {
		msg := "could not GetBlockByHash because of client is nil"
		return nil, errors.New(msg)
	}

	if tx.BlockHash != nil {
		revertRate, err := client.GetBlockRevertRateByHash(*tx.BlockHash)
		if err != nil {
			msg := fmt.Sprintf("get block revert rate by hash %v error", tx.BlockHash)
			return nil, types.WrapError(err, msg)
		}
		txDict.RevertRate = revertRate
	}

	block, err := client.GetBlockByHash(*tx.BlockHash)
	if err != nil {

	}
	txDict.TxAt = uint64(*block.Timestamp)

	value := big.Int(*tx.Value)
	txDict.Inputs = []richtypes.TxUnit{
		{
			Value:        &value,
			Address:      &tx.From,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}

	txDict.Outputs = []richtypes.TxUnit{
		{
			Value:        &value,
			Address:      tx.To,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}

	return txDict, nil
}

// ConvertByUnsignedTransaction converts types.UnsignedTransaction to TxDictBase
func (tc *TxDictBaseConverter) ConvertByUnsignedTransaction(tx *types.UnsignedTransaction) *richtypes.TxDictBase {
	txDictBase := new(richtypes.TxDictBase)

	value := big.Int(*tx.Value)
	txDictBase.Inputs = []richtypes.TxUnit{
		{
			Value:        &value,
			Address:      tx.From,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}

	txDictBase.Outputs = []richtypes.TxUnit{
		{
			Value:        &value,
			Address:      tx.To,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}
	return txDictBase
}
