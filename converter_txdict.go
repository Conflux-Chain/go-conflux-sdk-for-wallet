package walletsdk

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	richconstants "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/constants"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/decoder"
	walletinterface "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/interface"
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/constants"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

// TxDictConverter contains methods for convert other types to TxDict.
type TxDictConverter struct {
	richClient walletinterface.RichClientOperator
	tokenCache map[string]*richtypes.Token
	decoder    *decoder.ContractDecoder
	mutex      *sync.Mutex
	networkID  uint32
}

// NewTxDictConverter creates a TxDictConverter instance.
func NewTxDictConverter(richClient walletinterface.RichClientOperator) (*TxDictConverter, error) {
	contractDecoder, err := decoder.NewContractDecoder()
	if err != nil {
		return nil, err
	}

	networkID, err := richClient.GetClient().GetNetworkID()
	if err != nil {
		return nil, err
	}

	return &TxDictConverter{
		richClient: richClient,
		tokenCache: make(map[string]*richtypes.Token),
		decoder:    contractDecoder,
		mutex:      new(sync.Mutex),
		networkID:  networkID,
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

	if err != nil {
		msg := fmt.Sprintf("creat tx_dict with txHash:%v, blockHash:%v, from:%v, to:%v, value:%v error",
			tx.Hash, tx.BlockHash, tx.From, tx.To, tx.Value)
		return nil, errors.Wrap(err, msg)
	}

	sn := uint64(0)
	tc.fillTxDictByTx(txDict, &tx.From, tx.To, tx.Value, &sn)
	// fmt.Println("create txdict done")

	// no log will produced when transaction to is normal account or nil, so return
	if tx.To == nil {
		return txDict, nil
	}

	toType := tx.To.GetAddressType()
	if toType == cfxaddress.AddressTypeUser {
		return txDict, nil
	}

	// wait tx be packed up to 5 seconds
	var receipit *types.TransactionReceipt
	for i := 0; i < 5; i++ {
		receipit, err = tc.richClient.GetClient().GetTransactionReceipt(tx.Hash)
		if err != nil {
			return nil, errors.Wrapf(err, "get transaction receipt by hash %v error", tx.Hash)
		}
		if receipit != nil {
			break
		}
		time.Sleep(time.Second)
		fmt.Printf("receipt of %v : %+v\n", tx.Hash, receipit)
	}
	if receipit == nil {
		msg := fmt.Sprintf("convert failed, transaction %v is not be packed with in 5 seconds", tx.Hash)
		return nil, errors.New(msg)
	}
	// fmt.Println("get tx receipt done")

	err = tc.fillTxDictByTxReceipt(txDict, receipit, &sn)
	// fmt.Printf("after fill by receipt: %+v\n\n", txDict)
	if err != nil {
		return nil, errors.Wrapf(err, "fill tx_dict by tx receipt %v error", receipit)
	}
	return txDict, nil
}

func (tc *TxDictConverter) createTxDict(tx *types.Transaction, revertRate *big.Float, blockTime *hexutil.Uint64) (*richtypes.TxDict, error) {

	// fmt.Println("start creat txdict")
	txDict := new(richtypes.TxDict)
	txDict.TxHash = tx.Hash
	txDict.BlockHash = tx.BlockHash
	txDict.Extra.Gas = tx.Gas.ToInt()
	txDict.Extra.GasPrice = tx.GasPrice.ToInt()
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
			return nil, errors.Wrapf(err, "get block revert rate by hash %v error", tx.BlockHash)
		}
		// fmt.Println("get block revert rate by hash done")
	}
	txDict.RevertRate = revertRate

	if blockTime == nil && tx.BlockHash != nil {
		// fmt.Println("start get block by hash ", *blockhash)
		var err error
		block, err := client.GetBlockByHash(*tx.BlockHash)
		if err != nil {
			return nil, errors.Wrapf(err, "get block by hash %v error", tx.BlockHash)
		}
		blockTimeInU64 := hexutil.Uint64(block.Timestamp.ToInt().Uint64())
		blockTime = &blockTimeInU64
		// fmt.Println("get block by hash done")
	}
	if blockTime != nil {
		txDict.TxAt = richtypes.JSONTime(*blockTime)
	}

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
	// fmt.Printf("tc: %+v, txDict: %+v, receipt: %+v, sn: %+v\n", tc, txDict, receipt, sn)

	if txDict == nil || receipt == nil || sn == nil {
		return errors.New("all of txdict, receipt and sn could not be nil")
	}

	//decode event logs
	logs := receipt.Logs
	if logs == nil || len(logs) == 0 || receipt.To == nil {
		return nil
	}

	// sn := uint64(0)
	for _, log := range logs {

		// if to address is fc contract address, then only decode 1 log
		if reflect.DeepEqual(*receipt.To, richconstants.TethysFcV1Address) && log.Topics[0] == richconstants.Erc777SentEventSign {
			continue
		}

		// fmt.Println("start decode log")
		eventParams, err := tc.decoder.DecodeEvent(&log)
		if err != nil {
			return errors.Wrapf(err, "decode log %+v error", log)
		}

		// fill fields to input and output of tx_dict
		if eventParams != nil {
			// fmt.Printf("gen input and output by eventParams %+v", eventParams)
			// fmt.Printf("before getTokenByIdentifier, tc:%+v,log:%+v,receipt.To:%v", tc, log, receipt.To)
			tokenInfo := tc.getTokenByIdentifier(&log, *receipt.To)
			// get amount or value, if nil that means not token transfer
			amount, err := getValueOrAmount(eventParams)
			if err != nil {
				return err
			}

			paramsV := reflect.ValueOf(eventParams).Elem()
			from := cfxaddress.MustNewFromCommon(paramsV.FieldByName("From").Interface().(common.Address), tc.networkID)
			to := cfxaddress.MustNewFromCommon(paramsV.FieldByName("To").Interface().(common.Address), tc.networkID)

			//fill to txdict inputs and outputs
			input := richtypes.TxUnit{
				Value:           amount,
				Address:         &from,
				Sn:              *sn,
				TokenCode:       tokenInfo.TokenSymbol,
				TokenIdentifier: receipt.To,
				TokenDecimal:    tokenInfo.TokenDecimal,
			}
			output := richtypes.TxUnit{
				Value:           amount,
				Address:         &to,
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
func (tc *TxDictConverter) getTokenByIdentifier(log *types.Log, contractAddress types.Address) *richtypes.Token {

	if _, ok := tc.tokenCache[contractAddress.String()]; !ok {

		concrete, err := tc.decoder.GetTransferEventMatchedConcrete(log)

		tc.mutex.Lock()
		if err != nil || concrete == nil {
			tc.tokenCache[contractAddress.String()] = nil
			tc.mutex.Unlock()
			return nil
		}
		tc.mutex.Unlock()

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
		// TODO: Add --dubeg flag for print logs
		// if err != nil {
		// 	fmt.Printf("call function 'name' of contract address %+v which is %v error: %v\n\n", realContract.Address, concrete.ContractType, err)
		// }

		err = realContract.Call(nil, &symbol, "symbol")

		// the contract maybe not completely standard, so it is legal without symbol
		// TODO: Add --dubeg flag for print logs
		// if err != nil {
		// 	fmt.Printf("call function 'symbol' of contract address %+v which is %v error: %v\n\n", realContract.Address, concrete.ContractType, err)
		// }

		if _, ok := realContract.ABI.Methods["decimals"]; ok {
			err = realContract.Call(nil, &decimals, "decimals")

			// TODO: Add --dubeg flag for print logs
			// the contract maybe not completely standard, so it is legal without decimals
			// if err != nil {
			// 	fmt.Printf("call function 'decimals' of contract address %+v which is %v error: %v\n\n", realContract.Address, concrete.ContractType, err)
			// }
		}

		tc.mutex.Lock()
		tc.tokenCache[contractAddress.String()] = &richtypes.Token{
			TokenName:    string(name),
			TokenSymbol:  string(symbol),
			TokenDecimal: uint64(decimals),
			// Address:      &contractAddress,
			// TokenType:    concrete.ContractType,
		}
		tc.mutex.Unlock()
	}

	return tc.tokenCache[contractAddress.String()]
}

// ConvertByUnsignedTransaction converts types.UnsignedTransaction to TxDictBase.
func (tc *TxDictConverter) ConvertByUnsignedTransaction(tx *types.UnsignedTransaction) *richtypes.TxDictBase {
	txDictBase := new(richtypes.TxDictBase)
	if tx.Gas != nil {
		txDictBase.Extra.Gas = tx.Gas.ToInt()
	}

	if tx.GasPrice != nil {
		txDictBase.Extra.GasPrice = tx.GasPrice.ToInt()
	}

	value := tx.Value.ToInt()
	txDictBase.Inputs = []richtypes.TxUnit{
		{
			Value:        value,
			Address:      tx.From,
			Sn:           0,
			TokenCode:    constants.CFXSymbol,
			TokenDecimal: constants.CFXDecimal,
		},
	}

	txDictBase.Outputs = []richtypes.TxUnit{
		{
			Value:        value,
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
		to := cfxaddress.MustNewFromCommon(paramsV.FieldByName("To").Interface().(common.Address), tc.networkID)

		txDictBase.Inputs = append(txDictBase.Inputs, richtypes.TxUnit{

			Value:           amount,
			Address:         tx.From,
			Sn:              1,
			TokenIdentifier: tx.To,
		},
		)
		txDictBase.Outputs = append(txDictBase.Outputs, richtypes.TxUnit{
			Value:           amount,
			Address:         &to,
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
