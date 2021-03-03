// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package richtypes

import (
	"math/big"

	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
)

// Token describes token detail messages, such as erc20, erc777, fanscoin and so on.
type Token struct {
	TokenName    string `json:"name"`
	TokenSymbol  string `json:"symbol"`
	TokenDecimal uint64 `json:"decimals"`
	TokenIcon    string `json:"icon,omitempty"`
	// Granularity   uint64       `json:"granularity,omitempty"`
	// IsERC721      bool         `json:"isERC721"`
	// TotalSupply   *hexutil.Big `json:"totalSupply,omitempty"`
	// HolderCount   uint64       `json:"holderCount,omitempty"`
	// TransferCount uint64       `json:"transferCount,omitempty"`
	// SentCount     uint64       `json:"sentCount,omitempty"`
	// Icon          string       `json:"icon,omitempty"`
}

// TokenWithBalance describes token with balace information
type TokenWithBalance struct {
	Token
	Balance string `json:"balance"`
	Address string `json:"address"`
}

// TokenWithBlanceList describes list of token with balance
type TokenWithBlanceList struct {
	List []TokenWithBalance `json:"list"`
}

// TokenTransferEvent describes token transfer event information
type TokenTransferEvent struct {
	Token               `json:"token"`
	ContractAddress     types.Address `json:"address,omitempty"`
	TransactionHash     types.Hash    `json:"transactionHash"`
	TransactionLogIndex uint          `json:"transactionLogIndex"`
	From                types.Address `json:"from"`
	To                  types.Address `json:"to"`
	Value               string        `json:"value"`
	Timestamp           JSONTime      `json:"timestamp"`
	BlockHash           types.Hash    `json:"blockHash"`
	RevertRate          *big.Float    `json:"revertRate"`
	// Status          uint64        `json:"status"`
}

// TokenTransferEventList describes list of token tranfer event information
type TokenTransferEventList struct {
	Total uint64               `json:"total"`
	List  []TokenTransferEvent `json:"list"`
}

func (tbl *TokenWithBlanceList) FormatAddress() {
	for i := range tbl.List {
		tbl.List[i].Address = cfxaddress.FormatAddressStrToHex(tbl.List[i].Address)
	}
}
