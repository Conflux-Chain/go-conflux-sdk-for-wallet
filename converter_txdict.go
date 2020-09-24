package walletsdk

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/decoder"
	walletinterface "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/interface"
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/constants"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TxDictConverter contains methods for convert other types to TxDict.
type TxDictConverter struct {
	richClient walletinterface.RichClientOperator
	tokenCache map[types.Address]*richtypes.Token
	decoder    *decoder.ContractDecoder
}

// NewTxDictConverter creates a TxDictConverter instance.
func NewTxDictConverter(richClient walletinterface.RichClientOperator) (*TxDictConverter, error) {
	contractDecoder, err := decoder.NewContractDecoder()
	if err != nil {
		return nil, err
	}

	return &TxDictConverter{
		richClient: richClient,
		tokenCache: make(map[types.Address]*richtypes.Token),
		decoder:    contractDecoder,
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
			TokenCode:       tte.TokenSymbol,
			TokenDecimal:    tte.TokenDecimal,
			TokenIdentifier: &tte.ContractAddress,
		},
	}

	txDict.Outputs = []richtypes.TxUnit{
		{
			Value:           value,
			Address:         &tte.To,
			Sn:              0,
			TokenCode:       tte.TokenSymbol,
			TokenDecimal:    tte.TokenDecimal,
			TokenIdentifier: &tte.ContractAddress,
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

	txDict, err := tc.createTxDict(tx, revertRate, blockTime) //, &tx.From, tx.To, tx.Value)

	sn := uint64(0)
	tc.fillTxDictByTx(txDict, &tx.From, tx.To, tx.Value, &sn)

	if err != nil {
		msg := fmt.Sprintf("creat tx_dict with txHash:%v, blockHash:%v, from:%v, to:%v, value:%v error",
			tx.Hash, tx.BlockHash, tx.From, tx.To, tx.Value)
		return nil, types.WrapError(err, msg)
	}
	// fmt.Println("create txdict done")

	// wait tx be packed up to 5 seconds
	var receipit *types.TransactionReceipt
	for i := 0; i < 5; i++ {
		receipit, err = tc.richClient.GetClient().GetTransactionReceipt(tx.Hash)
		if err != nil {
			msg := fmt.Sprintf("get transaction receipt by hash %v error", tx.Hash)
			return nil, types.WrapError(err, msg)
		}
		if receipit != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if receipit == nil {
		return nil, errors.New("convert failed, transaction is not be packed with 5 seconds")
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

func (tc *TxDictConverter) createTxDict(tx *types.Transaction, revertRate *big.Float, blockTime *hexutil.Uint64) (*richtypes.TxDict, error) {

	// fmt.Println("start creat txdict")
	txDict := new(richtypes.TxDict)
	txDict.TxHash = tx.Hash
	txDict.BlockHash = tx.BlockHash
	txDict.Gas = tx.Gas.ToInt()
	txDict.GasPrice = tx.GasPrice.ToInt()
	txDict.Inputs = make([]richtypes.TxUnit, 0)
	txDict.Outputs = make([]richtypes.TxUnit, 0)

	client := tc.richClient.GetClient()
	if client == nil {
		msg := "could not GetBlockByHash because of client is nil"
		return nil, errors.New(msg)
	}

	if revertRate == nil && tx.BlockHash != nil {
		// fmt.Println("start get block revert rate by hash")
		var err error
		revertRate, err = client.GetBlockConfirmationRisk(*tx.BlockHash)
		if err != nil {
			msg := fmt.Sprintf("get block revert rate by hash %v error", tx.BlockHash)
			return nil, types.WrapError(err, msg)
		}
		// fmt.Println("get block revert rate by hash done")
	}
	txDict.RevertRate = revertRate

	if blockTime == nil && tx.BlockHash != nil {
		// fmt.Println("start get block by hash ", *blockhash)
		var err error
		block, err := client.GetBlockByHash(*tx.BlockHash)
		if err != nil {
			msg := fmt.Sprintf("get block by hash %v error", tx.BlockHash)
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
	//decode event logs
	logs := receipt.Logs
	if len(logs) == 0 {
		return nil
	}

	// sn := uint64(0)
	for _, log := range logs {

		// fmt.Println("start decode log")
		eventParams, err := tc.decoder.DecodeEvent(&log)
		if err != nil {
			msg := fmt.Sprintf("decode log %+v error", log)
			return types.WrapError(err, msg)
		}

		// fill fields to input and output of tx_dict
		if eventParams != nil {
			// fmt.Printf("gen input and output by eventParams %+v", eventParams)
			tokenInfo := tc.getTokenByIdentifier(&log, *receipt.To)
			// get amount or value, if nil that means not token transfer
			amount, err := getValueOrAmount(eventParams)
			if err != nil {
				return err
			}

			paramsV := reflect.ValueOf(eventParams).Elem()
			from := types.NewAddress(paramsV.FieldByName("From").Interface().(common.Address).String())
			to := types.NewAddress(paramsV.FieldByName("To").Interface().(common.Address).String())

			//fill to txdict inputs and outputs
			input := richtypes.TxUnit{
				Value:           amount,
				Address:         from,
				Sn:              *sn,
				TokenCode:       tokenInfo.TokenSymbol,
				TokenIdentifier: receipt.To,
				TokenDecimal:    tokenInfo.TokenDecimal,
			}
			output := richtypes.TxUnit{
				Value:           amount,
				Address:         to,
				Sn:              *sn,
				TokenCode:       tokenInfo.TokenSymbol,
				TokenIdentifier: receipt.To,
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
			// Address:      &contractAddress,
			// TokenType:    concrete.ContractType,
		}
	}

	return tc.tokenCache[contractAddress]
}

// ConvertByUnsignedTransaction converts types.UnsignedTransaction to TxDictBase.
func (tc *TxDictConverter) ConvertByUnsignedTransaction(tx *types.UnsignedTransaction) *richtypes.TxDictBase {
	txDictBase := new(richtypes.TxDictBase)
	if tx.Gas != nil {
		txDictBase.Gas = tx.Gas.ToInt()
	}

	if tx.GasPrice != nil {
		txDictBase.GasPrice = tx.GasPrice.ToInt()
	}

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

	// decode tx.Data to token transfer
	contractDecoder, err := decoder.NewContractDecoder()
	if err != nil {
		// fmt.Printf("NewContractDecoder err %v\n", err)
		return txDictBase
	}

	if tx.Data == nil || len(tx.Data) < 4 {
		return txDictBase
	}

	funcParams, err := contractDecoder.DecodeFunction(tx.Data)
	if err != nil {
		// fmt.Printf("decode function err %v\n", err)
		return txDictBase
	}

	// fmt.Printf("decode function done %+v\n\n", funcParams)
	amount, err := getValueOrAmount(funcParams)
	// fmt.Printf("get amount done %+v,err:%v\n\n", amount, err)
	if err == nil {
		paramsV := reflect.ValueOf(funcParams).Elem()
		to := types.NewAddress(paramsV.FieldByName("To").Interface().(common.Address).String())

		txDictBase.Inputs = append(txDictBase.Inputs, richtypes.TxUnit{

			Value:           amount,
			Address:         tx.From,
			Sn:              1,
			TokenIdentifier: tx.To,
		},
		)
		txDictBase.Outputs = append(txDictBase.Outputs, richtypes.TxUnit{
			Value:           amount,
			Address:         to,
			Sn:              1,
			TokenIdentifier: tx.To,
		})
	}

	return txDictBase
}

func getValueOrAmount(funcOrEventParams interface{}) (*big.Int, error) {

	// fmt.Printf("getValueOrAmount of %#v\n\n", funcOrEventParams)
	reflectValue := reflect.ValueOf(funcOrEventParams)
	// fmt.Printf("reflectValue is %+v\n\n", reflectValue)

	paramsV := reflectValue.Elem()
	fieldV := paramsV.FieldByName("Value")
	if (fieldV == reflect.Value{}) {
		fieldV = paramsV.FieldByName("Amount")
	}
	if (fieldV == reflect.Value{}) {
		return nil, fmt.Errorf("not found field amout or value from %+v", funcOrEventParams)
	}
	tokenValue := fieldV.Interface().(*big.Int)

	return tokenValue, nil
}
