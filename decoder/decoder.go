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
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ContractDecoder for decode event
type ContractDecoder struct {
	// ElemIdToToConcreteDicCache maps event ID to event contrete
	ElemIdToToConcreteDicCache map[string][]richtypes.ContractElemConcrete
}

// type FunctionDecoder struct {
// 	ElemIdToToConcreteDicCache map[string][]richtypes.ContractElemConcrete
// }

var contractElemIdToToConcreteDicCache map[string][]richtypes.ContractElemConcrete

// NewContractDecoder creates an EventDecoder instance
func NewContractDecoder() (*ContractDecoder, error) {
	dic, err := createContractElemIdToConcreteDic()
	if err != nil {
		return nil, err
	}

	return &ContractDecoder{
		ElemIdToToConcreteDicCache: dic,
	}, nil
}

// // NewFunctionDecoder creates an EventDecoder instance
// func NewFunctionDecoder() (*FunctionDecoder, error) {
// 	dic, err := createContractElemIdToConcreteDic()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &FunctionDecoder{
// 		ElemIdToToConcreteDicCache: dic,
// 	}, nil
// }

func createContractElemIdToConcreteDic() (map[string][]richtypes.ContractElemConcrete, error) {
	if contractElemIdToToConcreteDicCache != nil {
		return contractElemIdToToConcreteDicCache, nil
	}

	contractElemIdToToConcreteDicCache = make(map[string][]richtypes.ContractElemConcrete)

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

			// generate dic for every enent
			for _, event := range contract.ABI.Events {

				if event.RawName == contrete.ElemName {
					hash := event.ID.Hex()
					if contractElemIdToToConcreteDicCache[hash] == nil {
						contractElemIdToToConcreteDicCache[hash] = make([]richtypes.ContractElemConcrete, 0)
					}

					contractElemIdToToConcreteDicCache[hash] = append(contractElemIdToToConcreteDicCache[hash], contrete)
					// event name in contract is unique, so jump out of loop
					break
				}
			}

			// generate dic for every function
			for _, function := range contract.ABI.Methods {

				if function.RawName == contrete.ElemName {
					sign := hexutil.Encode(function.ID)
					if contractElemIdToToConcreteDicCache[sign] == nil {
						contractElemIdToToConcreteDicCache[sign] = make([]richtypes.ContractElemConcrete, 0)
					}

					contractElemIdToToConcreteDicCache[sign] = append(contractElemIdToToConcreteDicCache[sign], contrete)
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

	return contractElemIdToToConcreteDicCache, nil
}

// GetMatchedConcrete ...
func (cd *ContractDecoder) GetMatchedConcrete(log *types.LogEntry) (*richtypes.ContractElemConcrete, error) {
	if len(log.Topics) == 0 {
		return nil, nil
	}
	// event parameters of abi needs be "from" "to" "value"
	contretes := cd.ElemIdToToConcreteDicCache[log.Topics[0].String()]

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

// DecodeEvent finds the unique matched event concrete with the log and decodes the log into instance of event params struct
func (cd *ContractDecoder) DecodeEvent(log *types.LogEntry) (eventParmsPtr interface{}, err error) {
	concrete, err := cd.GetMatchedConcrete(log)
	if err != nil {
		return nil, err
	}
	if concrete != nil {
		return concrete.DecodeEvent(log)
	}
	return nil, nil
}

// DecodeFunction finds the unique matched event concrete with the log and decodes the log into instance of event params struct
func (cd *ContractDecoder) DecodeFunction(data []byte) (eventParmsPtr interface{}, err error) {
	concrete := cd.ElemIdToToConcreteDicCache[hexutil.Encode(data[:4])]

	if concrete != nil && len(concrete) > 0 {
		return concrete[0].DecodeFunction(data)
	}
	return nil, nil
}
