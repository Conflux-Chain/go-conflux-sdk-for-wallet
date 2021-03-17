// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package walletsdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/pkg/errors"
)

// scanServer represents a centralized server
type scanServer struct {
	Scheme        string
	Address       string
	HTTPRequester sdk.HTTPRequester
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
		rspBytes.Body.Close()
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
		return fmt.Errorf("failed to unmarshal '%v' to richtypes.ErrorResponse, error:%v", string(body), err.Error())
	}
	// fmt.Printf("unmarshaled body: %+v\n\n", rsp)

	if rsp.Code != 0 {
		msg := fmt.Sprintf("code:%+v, message:%+v", rsp.Code, rsp.Message)
		return errors.New(msg)
	}

	// unmarshl to result
	err = json.Unmarshal(body, unmarshaledResult)
	if err != nil {
		return fmt.Errorf("failed to unmarshal '%v' to unmarshaledResult, error:%v", string(body), err.Error())
	}
	// fmt.Printf("unmarshaled result: %+v\n\n", unmarshaledResult)
	return nil
}
