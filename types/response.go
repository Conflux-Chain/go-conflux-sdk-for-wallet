// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package richtypes

// ErrorResponse represents response of http request from scan service
type ErrorResponse struct {
	Code    uint64      `json:"code"`    // 	"code"
	Message string      `json:"message"` // "message"
	Result  interface{} `json:"result"`  // "result"
}
