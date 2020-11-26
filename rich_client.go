// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package walletsdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/constants"
	"github.com/Conflux-Chain/go-conflux-sdk/types"

	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// RichClient contains client, cfx-scan-backend server and contract-manager server
//
// RichClient is the client for wallet, it's methods need request centralized servers
// cfx-scan-backend and contract-manager in order to apply better performance.
type RichClient struct {
	cfxScanBackend  *scanServer
	contractManager *scanServer
	client          sdk.ClientOperator
}

// scanServer represents a centralized server
type scanServer struct {
	Scheme        string
	Address       string
	HTTPRequester sdk.HTTPRequester
}

// ServerConfig represents cfx-scan-backend and contract-manager configurations, because centralized servers maybe changed.
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

type blockAndRevertrate struct {
	block      *types.Block
	revertRate *big.Float
}

// default value of server config
var (
	accountTokensPath     = "/v1/token"       // "/api/account/token/list" //cfx scan backend
	tokenTransferListPath = "/v1/transfer"    // "/api/transfer/list"    //cfx scan backend
	txListPath            = "/v1/transaction" // "/api/transaction/list" //cfx scan backend
	contractQueryBasePath = "/v1/contract"    // "/api/contract/query" //contract manager
	tokenQueryBasePath    = "/v1/token"       //cfx scan backend

	cfxScanBackend = &scanServer{
		Scheme:        "http",
		Address:       "101.201.103.131:8885", //"testnet-jsonrpc.conflux-chain.org:18084",
		HTTPRequester: &http.Client{},
	}

	contractManager = &scanServer{
		Scheme:        "http",
		Address:       "101.201.103.131:8886", //"13.75.69.106:8886",
		HTTPRequester: &http.Client{},
	}
)

// NewRichClient create new rich client with client and server config.
//
// The fields of config will use default value when it's empty
func NewRichClient(client sdk.ClientOperator, configOption *ServerConfig) *RichClient {

	if configOption != nil {
		if configOption.CfxScanBackendSchema != "" {
			cfxScanBackend.Scheme = configOption.CfxScanBackendSchema
		}

		if configOption.CfxScanBackendAddress != "" {
			cfxScanBackend.Address = configOption.CfxScanBackendAddress
		}

		if configOption.ContractManagerSchema != "" {
			contractManager.Scheme = configOption.ContractManagerSchema
		}

		if configOption.ContractManagerAddress != "" {
			contractManager.Address = configOption.ContractManagerAddress
		}

		if configOption.AccountBalancesPath != "" {
			accountTokensPath = configOption.AccountBalancesPath
		}

		if configOption.AccountTokenTxListPath != "" {
			tokenTransferListPath = configOption.AccountTokenTxListPath
		}

		if configOption.TxListPath != "" {
			txListPath = configOption.TxListPath
		}

		if configOption.ContractQueryPath != "" {
			contractQueryBasePath = configOption.ContractQueryPath
		}
	}

	richClient := RichClient{
		cfxScanBackend,
		contractManager,
		client,
	}

	return &richClient
}

// GetClient returns client
func (rc *RichClient) GetClient() sdk.ClientOperator {
	return rc.client
}

// setHTTPRequester for unit test
func (rc *RichClient) setHTTPRequester(requester sdk.HTTPRequester) {
	rc.cfxScanBackend.HTTPRequester = requester
	rc.contractManager.HTTPRequester = requester
}

// URL returns url build by schema, host, path and params
func (s *scanServer) URL(path string, params map[string]interface{}) string {
	q := url.Values{}
	for key, val := range params {
		q.Add(key, fmt.Sprintf("%+v", val))
	}
	encodedParams := q.Encode()
	result := fmt.Sprintf("%+v://%+v%+v?%+v", s.Scheme, s.Address, path, encodedParams)
	return result
}

// Get sends a "Get" request and fill the unmarshaled value of field "Result" in response to unmarshaledResult
func (s *scanServer) Get(path string, params map[string]interface{}, unmarshaledResult interface{}) error {
	client := s.HTTPRequester
	// fmt.Println("request url:", s.URL(path, params))
	rspBytes, err := client.Get(s.URL(path, params))
	if err != nil {
		return err
	}

	defer func() {
		err := rspBytes.Body.Close()
		if err != nil {
			//fmt.Println("close rsp error", err)
		}
	}()

	body, err := ioutil.ReadAll(rspBytes.Body)
	if err != nil {
		return err
	}
	// fmt.Printf("body:%+v\n\n", string(body))

	// check if error response
	var rsp richtypes.ErrorResponse
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		return err
	}
	// fmt.Printf("unmarshaled body: %+v\n\n", rsp)

	if rsp.Code != 0 {
		msg := fmt.Sprintf("code:%+v, message:%+v", rsp.Code, rsp.Message)
		return errors.New(msg)
	}

	// rstBytes, err := json.Marshal(rsp.Result)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("marshaled result: %+v\n\n", string(rstBytes))

	// unmarshl to result
	err = json.Unmarshal(body, unmarshaledResult)
	if err != nil {
		return err
	}
	// fmt.Printf("unmarshaled result: %+v\n\n", unmarshaledResult)
	return nil
}

// GetAccountTokenTransfers returns address releated transactions,
// the tokenIdentifier represnets the token contract address and it is optional,
// when tokenIdentifier is specicied it returns token transfer events related the address,
// otherwise returns transactions about main coin.
func (rc *RichClient) GetAccountTokenTransfers(address types.Address, tokenIdentifier *types.Address, pageNumber, pageSize uint) (*richtypes.TokenTransferEventList, error) {
	params := make(map[string]interface{})
	params["accountAddress"] = address
	params["skip"] = pageNumber
	params["limit"] = pageSize
	params["txType"] = "all"

	var tteList *richtypes.TokenTransferEventList
	blockhashes := []types.Hash{}
	// when tokenIdentifier is not nil return transfer events of the token
	if tokenIdentifier != nil {
		var tts richtypes.TokenTransferEventList
		params["address"] = *tokenIdentifier
		err := rc.cfxScanBackend.Get(tokenTransferListPath, params, &tts)
		if err != nil {
			msg := fmt.Sprintf("get result of CfxScanBackend server and path {%+v}, params: {%+v} error", tokenTransferListPath, params)
			return nil, types.WrapError(err, msg)
		}
		tteList = &tts
		// fmt.Printf("%+v", tteList)

		// batch get blockhash through getTransactionByHash
		blockhashes = make([]types.Hash, 0, len(tteList.List))
		txhashes := make([]types.Hash, len(tteList.List))
		tokenAddressToTokenInfoMap := make(map[types.Address]*richtypes.Token)

		for i := range tteList.List {
			txhashes[i] = tteList.List[i].TransactionHash
		}

		// set block hash
		txhashToTxMap, err := rc.client.BatchGetTxByHashes(txhashes)
		if err != nil {
			msg := fmt.Sprintf("batch get txs by tx hashes %v error", txhashes)
			return nil, types.WrapError(err, msg)
		}

		for i := range tteList.List {
			hash := tteList.List[i].TransactionHash
			tx := txhashToTxMap[hash]
			if tx != nil && tx.BlockHash != nil {
				tteList.List[i].BlockHash = *txhashToTxMap[hash].BlockHash
			}
		}

		for _, th := range txhashes {
			if txhashToTxMap[th] != nil && txhashToTxMap[th].BlockHash != nil {
				blockhashes = append(blockhashes, *txhashToTxMap[th].BlockHash)
			}
		}

		// set token info
		for i := range tteList.List {
			tokenAddress := tteList.List[i].ContractAddress
			if _, ok := tokenAddressToTokenInfoMap[tokenAddress]; !ok {
				contract, err := rc.GetContractInfo(tokenAddress, true, false)
				if err != nil {
					msg := fmt.Sprintf("get token info of %v error", tokenAddress)
					return nil, types.WrapError(err, msg)
				}
				tokenAddressToTokenInfoMap[tokenAddress] = &contract.Token
			}
			tteList.List[i].Token = *tokenAddressToTokenInfoMap[tokenAddress]
		}

	} else {
		// when tokenIdentifier is nil return transaction of main coin
		var txs richtypes.TransactionList
		err := rc.cfxScanBackend.Get(txListPath, params, &txs)
		if err != nil {
			msg := fmt.Sprintf("get result of CfxScanBackend server and path {%+v}, params: {%+v} error", txListPath, params)
			return nil, types.WrapError(err, msg)
		}

		tteList = txs.ToTokenTransferEventList()

		// set blockhashes
		blockhashes = make([]types.Hash, len(txs.List))
		for i := range txs.List {
			blockhashes[i] = txs.List[i].BlockHash
		}
	}

	// use batch call instead of concurrency
	blkhashToRateMap, err := rc.client.BatchGetBlockConfirmationRisk(blockhashes)
	// fmt.Printf("blkhashToRateMap: %+v\n\n", blkhashToRateMap)
	if err != nil {
		msg := fmt.Sprintf("batch get block revert of blockhashes %v error", blockhashes)
		return nil, types.WrapError(err, msg)
	}
	for i, tte := range tteList.List {
		rate := blkhashToRateMap[tte.BlockHash]
		tteList.List[i].RevertRate = rate
	}
	return tteList, nil
}

// CreateSendTokenTransaction creates unsigned transaction for sending token according to input params,
// the tokenIdentifier represnets the token contract address.
// It supports erc20, erc777, fanscoin at present
func (rc *RichClient) CreateSendTokenTransaction(from types.Address, to types.Address, amount *hexutil.Big, tokenIdentifier *types.Address) (*types.UnsignedTransaction, error) {
	if tokenIdentifier == nil {
		tx, err := rc.client.CreateUnsignedTransaction(from, to, amount, nil)
		if err != nil {
			msg := fmt.Sprintf("Create Unsigned Transaction by from {%+v}, to {%+v}, amount {%+v} error", from, to, amount)
			return nil, types.WrapError(err, msg)
		}
		return tx, nil
	}

	cInfo, err := rc.GetContractInfo(*tokenIdentifier, true, false)
	if err != nil {
		// msg := fmt.Sprintf("get and unmarsal data from contract manager server with path {%+v}, paramas {%+v} error", contractQueryPath, params)
		msg := fmt.Sprintf("get contract info of %v error", tokenIdentifier)
		return nil, types.WrapError(err, msg)
	}

	contract, err := rc.client.GetContract([]byte(cInfo.ABI), tokenIdentifier)
	if err != nil {
		msg := fmt.Sprintf("get contract by ABI {%+v}, tokenIdentifier {%+v} error", cInfo.ABI, tokenIdentifier)
		return nil, types.WrapError(err, msg)
	}

	data, err := rc.getDataForTransToken(cInfo.GetContractTypeByABI(), contract, to, amount)
	if err != nil {
		msg := fmt.Sprintf("get data for transfer token method error, contract type {%+v} ", cInfo.GetContractTypeByABI())
		return nil, types.WrapError(err, msg)
	}

	tx, err := rc.client.CreateUnsignedTransaction(from, *tokenIdentifier, nil, data)
	if err != nil {
		msg := fmt.Sprintf("create transaction with params {from: %+v, to: %+v, data: %+v} error ", from, to, data)
		return nil, types.WrapError(err, msg)
	}
	return tx, nil
}

func (rc *RichClient) getDataForTransToken(contractType richtypes.ContractType, contract sdk.Contractor, to types.Address, amount *hexutil.Big) ([]byte, error) {
	var data []byte
	var err error

	// erc20 or fanscoin method signature are transfer(address,uint256)
	if contractType == richtypes.ERC20 || contractType == richtypes.FANSCOIN {
		data, err = contract.GetData("transfer", common.HexToAddress(string(to)), amount.ToInt())
		if err != nil {
			msg := fmt.Sprintf("get data of contract {%+v}, method {%+v}, params {to: %+v, amount: %+v} error ", contract, "transfer", to, amount)
			return nil, types.WrapError(err, msg)
		}
		return data, nil
	}

	// erc721 send by token_id
	//
	// if cInfo.ContractType == scantypes.ERC721 {
	// 	data, err = contract.GetData()
	// }

	// erc777 method signature is send(address,uint256,bytes)
	if contractType == richtypes.ERC777 {
		data, err = contract.GetData("send", common.HexToAddress(string(to)), amount.ToInt(), []byte{})
		if err != nil {
			msg := fmt.Sprintf("get data of contract {%+v}, method {%+v}, params {to: %+v, amount: %+v} error ", contract, "send", to, amount)
			return nil, types.WrapError(err, msg)
		}
		return data, nil
	}

	// if cInfo.ContractType == scantypes.DEX {
	// 	data, err = contract.GetData()
	// }

	msg := fmt.Sprintf("Do not support build data for transfer token function of contract type %+v", contractType)
	err = errors.New(msg)
	return nil, err
}

// GetContractInfo returns contract detail infomation, it will contains token info if it is token contract,
// it will contains abi if set needABI to be true.
func (rc *RichClient) GetContractInfo(contractAddress types.Address, needABI, needIcon bool) (*richtypes.Contract, error) {
	params := make(map[string]interface{})

	params["fields"] = ""
	if needIcon {
		params["fields"] = "icon"
	}
	if needABI {
		params["fields"] = params["fields"].(string) + ",abi"
	}

	var contractQueryFullPath = fmt.Sprintf("%v/%v", contractQueryBasePath, contractAddress)
	var contract richtypes.Contract
	err := rc.contractManager.Get(contractQueryFullPath, params, &contract)
	if err != nil {
		msg := fmt.Sprintf("get and unmarshal result of ContractManager server and path {%+v}, params: {%+v} error", contractQueryFullPath, params)
		return nil, types.WrapError(err, msg)
	}

	// get token info
	var tokenQueryFullPath = fmt.Sprintf("%v/%v", tokenQueryBasePath, contractAddress)
	rc.contractManager.Get(tokenQueryFullPath, params, &contract.Token)

	return &contract, nil
}

// GetAccountTokens returns coin balance and all token balances of specified address
func (rc *RichClient) GetAccountTokens(account types.Address) (*richtypes.TokenWithBlanceList, error) {
	params := make(map[string]interface{})
	params["accountAddress"] = account

	var tbs richtypes.TokenWithBlanceList
	err := rc.cfxScanBackend.Get(accountTokensPath, params, &tbs)
	if err != nil {
		msg := fmt.Sprintf("get and unmarshal result of ContractManager server and path {%+v}, params: {%+v} error", accountTokensPath, params)
		return nil, types.WrapError(err, msg)
	}
	return &tbs, nil
}

// GetTransactionsFromPool returns all pending transactions in mempool of conflux node.
//
// it only works on local conflux node currently.
func (rc *RichClient) GetTransactionsFromPool() (*[]types.Transaction, error) {
	var txs []types.Transaction

	if err := rc.client.CallRPC(&txs, "getTransactionsFromPool"); err != nil {
		msg := fmt.Sprintf("rpc getTransactionsFromPool error")
		return nil, types.WrapError(err, msg)
	}

	if txs == nil {
		return nil, nil
	}

	return &txs, nil
}

// GetTxDictByTxHash returns all cfx transfers and token transfers of transaction
func (rc *RichClient) GetTxDictByTxHash(hash types.Hash) (*richtypes.TxDict, error) {
	tx, err := rc.client.GetTransactionByHash(hash)
	if err != nil {
		msg := fmt.Sprintf("get transaction by hash %v error", hash)
		return nil, types.WrapError(err, msg)
	}

	tc, err := NewTxDictConverter(rc)
	if err != nil {
		return nil, fmt.Errorf("create TxDictConverter error")
	}

	return tc.ConvertByTransaction(tx, nil, nil)
}

// GetTxDictsByEpoch returns all cfx transfers and token transfers of the epoch
func (rc *RichClient) GetTxDictsByEpoch(epoch *types.Epoch) ([]richtypes.TxDict, error) {

	// start := time.Now()

	client := rc.GetClient()

	blockhashes, err := client.GetBlocksByEpoch(epoch)
	if err != nil {
		msg := fmt.Sprintf("get blocks by epoch %v error", epoch)
		return nil, types.WrapError(err, msg)
	}
	//fmt.Printf("get block hashes by epoch done, passed time: %v\n", time.Now().Sub(start))

	cache, errs := createBlockAndRevertrateCache(client, blockhashes)
	if errs != nil {
		return nil, joinError(errs)
	}
	// fmt.Printf("cache: %+v\n", cache)

	//fmt.Println("create block and reverrate cache done, passed time: %", time.Now().Sub(start))

	txdict, errs := rc.createTxDictsByBlockhashes(blockhashes, cache)
	if errs != nil {
		return nil, joinError(errs)
	}
	//fmt.Println("create tx dic done, passed time: %", time.Now().Sub(start))
	return txdict, nil
}

func createBlockAndRevertrateCache(client sdk.ClientOperator, blockhashes []types.Hash) (map[types.Hash]*blockAndRevertrate, []error) {
	// cache block and it's revertrate
	cache := make(map[types.Hash]*blockAndRevertrate)
	var errors []error

	// blockhashes = []types.Hash{"0x28d5a5b1b8f6c83e274b7ba1f027d16215596f27ea5effb745994401d23f8a18"}
	// concurrence get block and revertrate
	var wg sync.WaitGroup
	wg.Add(len(blockhashes) * 2)

	for _, blockhash := range blockhashes {
		cache[blockhash] = &blockAndRevertrate{}

		go func(bh types.Hash) {
			defer wg.Done()

			block, err := client.GetBlockByHash(bh)
			if err != nil {
				msg := fmt.Sprintf("get block by hash %v error", bh)
				if errors == nil {
					errors = make([]error, 0)
				}
				errors = append(errors, types.WrapError(err, msg))
			}
			cache[bh].block = block
		}(blockhash)

		// get risk rate and block time
		go func(bh types.Hash) {
			defer wg.Done()

			revertRate, err := client.GetBlockConfirmationRisk(bh)
			if err != nil {
				msg := fmt.Sprintf("get block revert rate by hash %v error", bh)
				if errors == nil {
					errors = make([]error, 0)
				}
				errors = append(errors, types.WrapError(err, msg))
			}
			cache[bh].revertRate = revertRate
		}(blockhash)
	}
	wg.Wait()

	return cache, errors
}

func (rc *RichClient) createTxDictsByBlockhashes(blockhashes []types.Hash, cache map[types.Hash]*blockAndRevertrate) ([]richtypes.TxDict, []error) {

	var errors = make([]error, 0)

	tc, err := NewTxDictConverter(rc)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	txDicts := make([]richtypes.TxDict, 0)

	txs := make([]types.Transaction, 0)
	for _, blockhash := range blockhashes {
		// fmt.Printf("cache[%v]= %+v\n", blockhash, cache[blockhash])
		txs = append(txs, cache[blockhash].block.Transactions...)
	}

	all := len(txs)
	con := constants.RPCConcurrence
	excuted := 0
	for {
		isLastLoop := (all-excuted)/con == 0
		if isLastLoop {
			con = all % con
		}

		var wg sync.WaitGroup
		wg.Add(con)

		var mutex = new(sync.Mutex)

		for i := 0; i < con; i++ {

			go func(_tx types.Transaction) {

				defer wg.Done()
				//fmt.Println("excute tx done:", excuted)

				// blockhash null means that tx is excuted by other block, so skip it
				if _tx.BlockHash == nil {
					return
				}

				cacheVal := cache[*_tx.BlockHash]

				txDict, err := tc.ConvertByTransaction(&_tx, cacheVal.revertRate, cacheVal.block.Timestamp)

				mutex.Lock()
				defer mutex.Unlock()
				if err != nil {
					errors = append(errors, err)
					return
				}
				txDicts = append(txDicts, *txDict)
			}(txs[excuted])
			excuted++
			//fmt.Println("excuting tx :", excuted)
		}

		wg.Wait()

		if isLastLoop {
			break
		}
	}

	return txDicts, nil
}

func joinError(errs []error) error {
	if errs != nil && len(errs) > 0 {
		errorStrs := make([]string, len(errs))
		for i, e := range errs {
			errorStrs[i] = e.Error()
		}
		joinedErr := strings.Join(errorStrs, "\n")
		return errors.New(joinedErr)
	}
	return nil
}

func jsonIt(input interface{}) string {
	j, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	return string(j)
}
