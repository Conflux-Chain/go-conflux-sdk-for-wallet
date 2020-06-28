// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package richtypes

import (
	"math/big"

	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

// Token describes token detail messages, such as erc20, erc777, fanscoin and so on.
type Token struct {
	TokenName    string `json:"name"`
	TokenSymbol  string `json:"symbol"`
	TokenDecimal uint64 `json:"decimals"`
	// Address      *types.Address `json:"address,omitempty"`
	// TokenType    ContractType   `json:"token_type,omitempty"`
}

// TokenWithBalance describes token with balace information
type TokenWithBalance struct {
	Token `json:"token"`
	// TokenName    string `json:"tokenName"`
	// TokenSymbol  string `json:"tokenSymbol"`
	// TokenDecimal int    `json:"tokenDecimal"`
	Balance string `json:"balance"`
	Address string `json:"address"`
}

// TokenWithBlanceList describes list of token with balance
type TokenWithBlanceList struct {
	List []TokenWithBalance `json:"list"`
}

// TokenTransferEvent describes token transfer event information
type TokenTransferEvent struct {
	Token           `json:"token"`
	ContractAddress types.Address `json:"address,omitempty"`
	TransactionHash types.Hash    `json:"transactionHash"`
	From            types.Address `json:"from"`
	To              types.Address `json:"to"`
	Value           string        `json:"value"`
	Timestamp       JSONTime      `json:"timestamp"`
	BlockHash       types.Hash    `json:"blockHash"`
	RevertRate      *big.Float    `json:"revertRate"`
	// Status          uint64        `json:"status"`
}

// TokenTransferEventList describes list of token tranfer event information
type TokenTransferEventList struct {
	Total uint64               `json:"total"`
	List  []TokenTransferEvent `json:"list"`
}
