// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package richtypes

import (
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/helper"
	"github.com/Conflux-Chain/go-conflux-sdk/constants"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

// Transaction represents transction information response from scan service
type Transaction struct {
	Hash             types.Hash     `json:"hash"`
	BlockHash        types.Hash     `json:"blockHash,omitempty"`
	TransactionIndex uint64         `json:"transactionIndex,omitempty"`
	From             types.Address  `json:"from"`
	To               *types.Address `json:"to,omitempty"`
	Value            string         `json:"value"`
	GasPrice         string         `json:"gasPrice"`
	Gas              string         `json:"gas"`
	ContractCreated  types.Address  `json:"contractCreated,omitempty"`
	Status           uint64         `json:"status,omitempty"`
	Timestamp        JSONTime       `json:"timestamp"`
	EpochHeight      uint64         `json:"epochHeight"`
	EpochNumber      uint64         `json:"epochNumber"`
	SyncTimestamp    uint64         `json:"syncTimestamp"`
	Risk             float64        `json:"risk"`
	GasFee           string         `json:"gasFee"`
	GasUsed          string         `json:"gasUsed"`
	// Data             string        `json:"data,omitempty"`
	// Nonce            *big.Int      `json:"nonce"`
}

// TransactionList represents a list of transaction
type TransactionList struct {
	Total uint64 `json:"total"`
	// ListLimit uint64        `json:"listLimit"`
	List []Transaction `json:"list"`
}

// ToTokenTransferEvent converts Transaction to TokenTransferEvent
func (tx *Transaction) ToTokenTransferEvent() *TokenTransferEvent {
	var tte TokenTransferEvent
	tte.TransactionHash = tx.Hash
	// tte.Status = tx.Status
	tte.From = tx.From.MustGetCommonAddress()
	tte.To = helper.MustGetCommonAddressPtr(tx.To)
	tte.Value = tx.Value
	tte.Timestamp = tx.Timestamp
	tte.BlockHash = tx.BlockHash

	tte.TokenName = constants.CFXName
	tte.TokenSymbol = constants.CFXSymbol
	tte.TokenDecimal = constants.CFXDecimal
	// tte.TokenType = UNKNOWN

	return &tte
}

// ToTokenTransferEventList converts TransactionList to TokenTransferEventList
func (txs *TransactionList) ToTokenTransferEventList() *TokenTransferEventList {
	var tteList TokenTransferEventList

	tteList.Total = txs.Total
	// tteList.ListLimit = txs.ListLimit
	listLen := len(txs.List)
	tteList.List = make([]TokenTransferEvent, listLen)

	for i, v := range txs.List {
		tteList.List[i] = *v.ToTokenTransferEvent()
	}
	return &tteList
}
