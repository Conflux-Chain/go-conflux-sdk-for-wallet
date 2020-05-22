package constants

import (
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

// func init() {
// 	// create a contract by abi which contains all events we needs
// 	var abiJSON []byte

// 	var client *sdk.Client
// 	contract, err := client.GetContract(abiJSON, nil)
// 	if err != nil {
// 		msg := fmt.Sprintf("unmarshal json {%+v} to ABI error", abiJSON)
// 		panic(msg)
// 	}
// 	// generate dic for every enents
// 	for _, event := range contract.ABI.Events {
// 		hash := types.Hash(event.ID().Hex())
// 		EventHashToConcreteDic[hash] = struct {
// 			Contract  *sdk.Contract
// 			EventName string
// 		}{
// 			Contract:  contract,
// 			EventName: event.RawName,
// 		}
// 	}
// }

const (
	// RPCConcurrence represents rpc request concurrence with conflux full node
	RPCConcurrence = 10
)

var (
	// EventHashToConcreteDic ...
	EventHashToConcreteDic map[types.Hash]struct {
		Contract  *sdk.Contract
		EventName string
	} = make(map[types.Hash]struct {
		Contract  *sdk.Contract
		EventName string
	})
)

func decodeEvent(types.LogEntry) {

}
