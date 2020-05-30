package walletsdk

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/decoder"
	walletinterface "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/interface"
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/constants"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TxDictBaseConverter contains methods for convert other types to TxDictBase.
type TxDictBaseConverter struct {
}

// TxDictConverter contains methods for convert other types to TxDict.
type TxDictConverter struct {
	richClient walletinterface.RichClientOperator
	tokenCache map[types.Address]*richtypes.Token
	decoder    *decoder.EventDecoder
}

// NewTxDictConverter creates a TxDictConverter instance.
func NewTxDictConverter(richClient walletinterface.RichClientOperator) (*TxDictConverter, error) {
	eventDecoder, err := decoder.NewEventDecoder()
	if err != nil {
		return nil, err
	}

	return &TxDictConverter{
		richClient: richClient,
		tokenCache: make(map[types.Address]*richtypes.Token),
		decoder:    eventDecoder,
	}, nil
}

// ConvertByTokenTransferEvent converts richtypes.TokenTransferEvent to TxDict.
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
			Value:           value,
			Address:         &tte.From,
			Sn:              0,
			TokenCode:       constants.CFXSymbol,
			TokenDecimal:    constants.CFXDecimal,
			TokenIdentifier: tte.Address,
		},
	}

	txDict.Outputs = []richtypes.TxUnit{
		{
			Value:           value,
			Address:         &tte.To,
			Sn:              0,
			TokenCode:       constants.CFXSymbol,
			TokenDecimal:    constants.CFXDecimal,
			TokenIdentifier: tte.Address,
		},
	}
	return txDict, nil
}

// ConvertByTransaction converts types.Transaction to TxDict.
func (tc *TxDictConverter) ConvertByTransaction(tx *types.Transaction, revertRate *big.Float, blockTime *hexutil.Uint64) (*richtypes.TxDict, error) {
	// fmt.Printf("start convert by tx, the blocktime is %#v\n", *blockTime)
	if tx == nil {
		return nil, errors.New("tx is nil")
	}

	txDict, err := tc.createTxDict(tx.Hash, tx.BlockHash, revertRate, blockTime) //, &tx.From, tx.To, tx.Value)

	sn := uint64(0)
	tc.fillTxDictByTx(txDict, &tx.From, tx.To, tx.Value, &sn)

	if err != nil {
		msg := fmt.Sprintf("creat tx_dict with txHash:%v, blockHash:%v, from:%v, to:%v, value:%v error",
			tx.Hash, tx.BlockHash, tx.From, tx.To, tx.Value)
		return nil, types.WrapError(err, msg)
	}
	// fmt.Println("create txdict done")

	receipit, err := tc.richClient.GetClient().GetTransactionReceipt(tx.Hash)
	if err != nil {
		msg := fmt.Sprintf("get transaction receipt by hash %v error", tx.Hash)
		return nil, types.WrapError(err, msg)
	}
	// fmt.Println("get tx receipt done")

	err = tc.fillTxDictByTxReceipt(txDict, receipit, &sn)
	// fmt.Printf("after fill by receipt: %+v\n\n", txDict)
	if err != nil {
		msg := fmt.Sprintf("fill tx_dict by tx receipt %v error", receipit)
		return nil, types.WrapError(err, msg)
	}
	return txDict, nil
}

func (tc *TxDictConverter) createTxDict(txhash types.Hash, blockhash *types.Hash, revertRate *big.Float, blockTime *hexutil.Uint64) (*richtypes.TxDict, error) {

	// fmt.Println("start creat txdict")
	txDict := new(richtypes.TxDict)
	txDict.TxHash = txhash
	txDict.BlockHash = blockhash
	txDict.Inputs = make([]richtypes.TxUnit, 0)
	txDict.Outputs = make([]richtypes.TxUnit, 0)

	client := tc.richClient.GetClient()
	if client == nil {
		msg := "could not GetBlockByHash because of client is nil"
		return nil, errors.New(msg)
	}

	if revertRate == nil && blockhash != nil {
		// fmt.Println("start get block revert rate by hash")
		var err error
		revertRate, err = client.GetBlockRevertRateByHash(*blockhash)
		if err != nil {
			msg := fmt.Sprintf("get block revert rate by hash %v error", blockhash)
			return nil, types.WrapError(err, msg)
		}
		// fmt.Println("get block revert rate by hash done")
	}
	txDict.RevertRate = revertRate

	if blockTime == nil && blockhash != nil {
		// fmt.Println("start get block by hash ", *blockhash)
		var err error
		block, err := client.GetBlockByHash(*blockhash)
		if err != nil {
			msg := fmt.Sprintf("get block by hash %v error", blockhash)
			return nil, types.WrapError(err, msg)
		}
		blockTime = block.Timestamp
		// fmt.Println("get block by hash done")
	}
	txDict.TxAt = richtypes.JSONTime(*blockTime)
	return txDict, nil
}

func (tc *TxDictConverter) fillTxDictByTx(txDict *richtypes.TxDict, from *types.Address, to *types.Address, value *hexutil.Big, sn *uint64) {

	_value := big.Int(*value)

	input := richtypes.TxUnit{
		Value:        &_value,
		Address:      from,
		Sn:           *sn,
		TokenCode:    constants.CFXSymbol,
		TokenDecimal: constants.CFXDecimal,
	}

	output := richtypes.TxUnit{
		Value:        &_value,
		Address:      to,
		Sn:           *sn,
		TokenCode:    constants.CFXSymbol,
		TokenDecimal: constants.CFXDecimal,
	}
	(*sn)++

	txDict.Inputs = append(txDict.Inputs, input)
	txDict.Outputs = append(txDict.Outputs, output)
}

// fillTxDictByTxReceipt fills token transfers to txDict by analizing receipt
func (tc *TxDictConverter) fillTxDictByTxReceipt(txDict *richtypes.TxDict, receipt *types.TransactionReceipt, sn *uint64) error {
	// fmt.Println("start fillTxDictByTxReceipt")
	defer func() {
		// fmt.Println("fillTxDictByTxReceipt done")
	}()
	//decode event logs
	logs := receipt.Logs
	if len(logs) == 0 {
		return nil
	}

	// sn := uint64(0)
	for _, log := range logs {

		// fmt.Println("start decode log")
		eventParams, err := tc.decoder.Decode(&log)
		if err != nil {
			msg := fmt.Sprintf("decode log %+v error", log)
			return types.WrapError(err, msg)
		}

		// fill fields to input and output of tx_dict
		if eventParams != nil {
			// fmt.Printf("gen input and output by eventParams %+v", eventParams)
			tokenInfo := tc.getTokenByIdentifier(&log, *receipt.To)
			// get amount or value, if nil that means not token transfer
			paramsV := reflect.ValueOf(eventParams).Elem()
			fieldV := paramsV.FieldByName("Value")
			if (fieldV == reflect.Value{}) {
				fieldV = paramsV.FieldByName("Amount")
			}

			if (fieldV == reflect.Value{}) {
				msg := fmt.Sprintf("not found field amout or value from %+v", eventParams)
				return types.WrapError(err, msg)
			}
			value := fieldV.Interface().(*big.Int)

			from := types.NewAddress(paramsV.FieldByName("From").Interface().(common.Address).String())
			to := types.NewAddress(paramsV.FieldByName("To").Interface().(common.Address).String())

			//fill to txdict inputs and outputs
			input := richtypes.TxUnit{
				Value:           value,
				Address:         from,
				Sn:              *sn,
				TokenCode:       tokenInfo.TokenSymbol,
				TokenIdentifier: tokenInfo.Address,
				TokenDecimal:    tokenInfo.TokenDecimal,
			}
			output := richtypes.TxUnit{
				Value:           value,
				Address:         to,
				Sn:              *sn,
				TokenCode:       tokenInfo.TokenSymbol,
				TokenIdentifier: tokenInfo.Address,
				TokenDecimal:    tokenInfo.TokenDecimal,
			}
			txDict.Inputs = append(txDict.Inputs, input)
			txDict.Outputs = append(txDict.Outputs, output)
			(*sn)++
		}
	}
	return nil
}

// getTokenByIdentifier ...
func (tc *TxDictConverter) getTokenByIdentifier(log *types.LogEntry, contractAddress types.Address) *richtypes.Token {
	if _, ok := tc.tokenCache[contractAddress]; !ok {

		concrete, err := tc.decoder.GetMatchedConcrete(log)
		if err != nil {
			tc.tokenCache[contractAddress] = nil
			return nil
		}

		if concrete == nil {
			tc.tokenCache[contractAddress] = nil
			return nil
		}

		realContract := sdk.Contract{ABI: concrete.Contract.ABI, Client: tc.richClient.GetClient(), Address: &contractAddress}
		var (
			name     string
			symbol   string
			decimals uint8
		)

		// currently there is no confusion methods exist, so call by the method name directly,
		// if not we need use type_map file to identify the exactly contract type and method type.
		err = realContract.Call(nil, &name, "name")
		// the contract maybe not completely standard, so it is legal without name
		if err != nil {
			fmt.Printf("call function 'name' of contract address %+v which is %v error: %v\n\n", realContract.Address, concrete.ContractType, err)
		}

		err = realContract.Call(nil, &symbol, "symbol")
		// the contract maybe not completely standard, so it is legal without symbol
		if err != nil {
			fmt.Printf("call function 'symbol' of contract address %+v which is %v error: %v\n\n", realContract.Address, concrete.ContractType, err)
		}

		if _, ok := realContract.ABI.Methods["decimals"]; ok {
			err = realContract.Call(nil, &decimals, "decimals")
			// the contract maybe not completely standard, so it is legal without decimals
			if err != nil {
				fmt.Printf("call function 'decimals' of contract address %+v which is %v error: %v\n\n", realContract.Address, concrete.ContractType, err)
			}
		}

		tc.tokenCache[contractAddress] = &richtypes.Token{
			TokenName:    string(name),
			TokenSymbol:  string(symbol),
			TokenDecimal: uint64(decimals),
			Address:      &contractAddress,
			TokenType:    concrete.ContractType,
		}
	}

	return tc.tokenCache[contractAddress]
}

// ConvertByUnsignedTransaction converts types.UnsignedTransaction to TxDictBase.
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

// // ConvertByRichTransaction convert richtypes.Transaction to TxDict.
// func (tc *TxDictConverter) ConvertByRichTransaction(tx *richtypes.Transaction) (*richtypes.TxDict, error) {
// 	txDict := new(richtypes.TxDict)
// 	txDict.TxHash = tx.Hash
// 	txDict.BlockHash = &tx.BlockHash
// 	txDict.TxAt = tx.Timestamp

// 	client := tc.richClient.GetClient()
// 	if client == nil {
// 		msg := "could not GetBlockRevertRateByHash because of client is nil"
// 		return nil, errors.New(msg)
// 	}

// 	revertRate, err := client.GetBlockRevertRateByHash(tx.BlockHash)
// 	if err != nil {
// 		msg := fmt.Sprintf("get block revert rate by hash %v error", tx.BlockHash)
// 		return nil, types.WrapError(err, msg)
// 	}
// 	txDict.RevertRate = revertRate

// 	value, ok := new(big.Int).SetString(tx.Value, 0)
// 	if !ok {
// 		msg := fmt.Sprintf("Convert tx.Value %v to *big.Int fail", tx.Value)
// 		return nil, errors.New(msg)
// 	}

// 	txDict.Inputs = []richtypes.TxUnit{
// 		{
// 			Value:        value,
// 			Address:      &tx.From,
// 			Sn:           0,
// 			TokenCode:    constants.CFXSymbol,
// 			TokenDecimal: constants.CFXDecimal,
// 		},
// 	}

// 	txDict.Outputs = []richtypes.TxUnit{
// 		{
// 			Value:        value,
// 			Address:      &tx.To,
// 			Sn:           0,
// 			TokenCode:    constants.CFXSymbol,
// 			TokenDecimal: constants.CFXDecimal,
// 		},
// 	}

// 	return txDict, nil
// }
