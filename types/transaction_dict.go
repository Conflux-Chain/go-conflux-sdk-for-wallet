// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package richtypes

import (
	"math/big"

	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

// TxDictBase is another representation of unsigned transaction which is designed for bitpie wallet
type TxDictBase struct {
	Inputs  []TxUnit `json:"inputs"`
	Outputs []TxUnit `json:"outputs"`
}

// TxDict is another representation of confirmed transaction which is designed for bitpie wallet
type TxDict struct {
	TxDictBase
	TxHash types.Hash `json:"tx_hash"`
	TxAt   uint64     `json:"'tx_at'"`
	// Extra      interface{} `json:"extra"`
	RevertRate *big.Float  `json:"confirmed_at,omitempty"`
	BlockHash  *types.Hash `json:"block_no,omitempty"`
}

// TxUnit represents a transaction unit
type TxUnit struct {
	Value           *big.Int       `json:"value"`
	Address         *types.Address `json:"address"`
	Sn              uint64         `json:"sn"`
	TokenCode       string         `json:"token_code"`
	TokenIdentifier *types.Address `json:"token_identifier"`
	TokenDecimal    uint64         `json:"token_decimal"`
}
