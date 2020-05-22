package decoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strings"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/helper"
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/constants"
	types "github.com/Conflux-Chain/go-conflux-sdk/types"
)

// EventDecoder for decode event
type EventDecoder struct {
	// EventHashToConcreteDic maps event ID to event contrete
	EventHashToConcreteDic map[types.Hash][]richtypes.ContractElemConcrete
}

var eventHashToConcreteDicCache map[types.Hash][]richtypes.ContractElemConcrete

// NewEventDecoder creates an EventDecoder instance
func NewEventDecoder() (*EventDecoder, error) {
	dic, err := createEventHashToConcreteDic()
	if err != nil {
		return nil, err
	}

	return &EventDecoder{
		EventHashToConcreteDic: dic,
	}, nil
}

func createEventHashToConcreteDic() (map[types.Hash][]richtypes.ContractElemConcrete, error) {
	if eventHashToConcreteDicCache != nil {
		return eventHashToConcreteDicCache, nil
	}

	eventHashToConcreteDicCache = make(map[types.Hash][]richtypes.ContractElemConcrete)

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("could not get current file path")
	}

	var abiDir = path.Join(path.Dir(currentFile), "../resource/contract/abi")
	var typeMapDir = path.Join(path.Dir(currentFile), "../resource/contract/type_map")
	// foreach dir abi files to get contract

	abiFiles, err := ioutil.ReadDir(abiDir)
	if err != nil {
		msg := fmt.Sprintf("read dir %v error", abiDir)
		return nil, types.WrapError(err, msg)
	}

	for _, abiFile := range abiFiles {
		// get contract
		abiJSON, err := ioutil.ReadFile(path.Join(abiDir, abiFile.Name()))
		if err != nil {
			msg := fmt.Sprintf("read file %v error", path.Join(abiDir, abiFile.Name()))
			return nil, types.WrapError(err, msg)
		}

		var client *sdk.Client
		contract, err := client.GetContract(abiJSON, types.NewAddress(constants.ZeroAddress.String()))
		if err != nil {
			msg := fmt.Sprintf("unmarshal json {%+v} to ABI error", abiJSON)
			return nil, types.WrapError(err, msg)
		}

		// get event type map
		typeMapFileName := path.Join(typeMapDir, abiFile.Name())
		if !helper.IsFileExists(typeMapFileName) {
			continue
		}

		// unmarshal event type map
		typeMapJSON, err := ioutil.ReadFile(typeMapFileName)
		if err != nil {
			msg := fmt.Sprintf("read file %v error", typeMapFileName)
			return nil, types.WrapError(err, msg)
		}

		typeMap := []richtypes.ContractElemConcrete{}
		err = json.Unmarshal(typeMapJSON, &typeMap)
		if err != nil {
			msg := fmt.Sprintf("unmarshal json {%+v} to typeMap error", typeMapJSON)
			return nil, types.WrapError(err, msg)
		}

		// fmt.Printf("get typeMap: %+v\n\n", typeMap)
		for _, contrete := range typeMap {

			// contrete := EventConcrete{}
			contrete.Contract = contract
			// get contract type by abi file name
			dotIndex := strings.Index(abiFile.Name(), ".")
			contrete.ContractType = richtypes.ContractType(abiFile.Name()[0:dotIndex])

			// generate dic for every enents
			for _, event := range contract.ABI.Events {

				if event.RawName == contrete.ElemName {
					hash := types.Hash(event.ID.Hex())
					if eventHashToConcreteDicCache[hash] == nil {
						eventHashToConcreteDicCache[hash] = make([]richtypes.ContractElemConcrete, 0)
					}

					eventHashToConcreteDicCache[hash] = append(eventHashToConcreteDicCache[hash], contrete)
					// event name in contract is unique, so jump out of loop
					break
				}
			}
		}
	}

	// fmt.Println("EventHashToConcreteDic length:", len(eventHashToConcreteDicCache))
	// for k, v := range eventHashToConcreteDicCache {
	// fmt.Printf("hash:%v,concrete:%+v\n", k, v)
	// }

	return eventHashToConcreteDicCache, nil
}

// GetMatchedConcrete ...
func (ed *EventDecoder) GetMatchedConcrete(log *types.LogEntry) (*richtypes.ContractElemConcrete, error) {
	if len(log.Topics) == 0 {
		return nil, nil
	}
	// event parameters of abi needs be "from" "to" "value"
	contretes := ed.EventHashToConcreteDic[log.Topics[0]]

	// if contretes length larger than 0, decode it
	if len(contretes) > 0 {
		var erc20 *richtypes.ContractElemConcrete
		var erc721 *richtypes.ContractElemConcrete
		var others []*richtypes.ContractElemConcrete
		for i := range contretes {
			if contretes[i].ElemType != richtypes.TransferEvent {
				break
			}
			if contretes[i].ContractType == richtypes.ERC20 {
				erc20 = &contretes[i]
			} else if contretes[i].ContractType == richtypes.ERC721 {
				erc721 = &contretes[i]
			} else {
				if others == nil {
					others = make([]*richtypes.ContractElemConcrete, 0)
				}
				others = append(others, &contretes[i])
			}
		}

		// if contract type is erc20 or Erc721, judge contract type by topics length, 3 for erc20 and 4 for erc721
		topicLen := len(log.Topics)
		if erc20 != nil && topicLen == 3 {
			return erc20, nil
		}

		if erc721 != nil && topicLen == 4 {
			return erc721, nil
		}

		// if others len is 1, docede it
		if len(others) == 1 {
			return others[0], nil
		}

		return nil, fmt.Errorf("could not determine contract type by log: %+v", log)
	}
	return nil, nil
}

// Decode finds the unique matched event concrete with the log and decodes the log into instance of event params struct
func (ed *EventDecoder) Decode(log *types.LogEntry) (eventParmsPtr interface{}, err error) {
	concrete, err := ed.GetMatchedConcrete(log)
	if err != nil {
		return nil, err
	}
	if concrete != nil {
		return concrete.Decode(log)
	}
	return nil, nil
}
