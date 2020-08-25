package decoder

import (
	"fmt"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/resource/contract/abi"
	"github.com/Conflux-Chain/go-conflux-sdk-for-wallet/resource/contract/elem"
	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/constants"
	types "github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ContractDecoder for decode event
type ContractDecoder struct {
	// ElemIdToConcreteDicCache maps event ID to event contrete
	ElemIdToConcreteDicCache map[string][]richtypes.ContractElemConcrete
}

var contractElemIdToConcreteDicCache map[string][]richtypes.ContractElemConcrete

// NewContractDecoder creates an EventDecoder instance
func NewContractDecoder() (*ContractDecoder, error) {
	dic, err := createContractElemIdToConcreteDic()
	if err != nil {
		return nil, err
	}

	return &ContractDecoder{
		ElemIdToConcreteDicCache: dic,
	}, nil
}

// createContractElemIdToConcreteDic creat mappings for contract element id, containts event id or function signature, to element concrete information.
func createContractElemIdToConcreteDic() (map[string][]richtypes.ContractElemConcrete, error) {
	if contractElemIdToConcreteDicCache != nil {
		return contractElemIdToConcreteDicCache, nil
	}

	contractElemIdToConcreteDicCache = make(map[string][]richtypes.ContractElemConcrete)

	for contractType, abiJSON := range abi.ABIJsonDic {
		// get contract

		var client *sdk.Client
		contract, err := client.GetContract([]byte(abiJSON), types.NewAddress(constants.ZeroAddress.String()))
		if err != nil {
			msg := fmt.Sprintf("unmarshal json {%+v} to ABI error", abiJSON)
			return nil, types.WrapError(err, msg)
		}

		elemConcretes := []richtypes.ContractElemConcrete{}
		for _, value := range elem.GetContractElems(contractType) {
			elemConcretes = append(elemConcretes, richtypes.ContractElemConcrete{ContractElem: value})
		}

		// fmt.Printf("get typeMap: %+v\n\n", typeMap)
		for _, contrete := range elemConcretes {

			contrete.Contract = contract
			// get contract type by abi file name

			contrete.ContractType = contractType

			// generate dic for every enent
			for _, event := range contract.ABI.Events {

				if event.RawName == contrete.ElemName {
					hash := event.ID.Hex()
					if contractElemIdToConcreteDicCache[hash] == nil {
						contractElemIdToConcreteDicCache[hash] = make([]richtypes.ContractElemConcrete, 0)
					}

					contractElemIdToConcreteDicCache[hash] = append(contractElemIdToConcreteDicCache[hash], contrete)
					// event name in contract is unique, so jump out of loop
					break
				}
			}

			// generate dic for every function
			for _, function := range contract.ABI.Methods {

				if function.RawName == contrete.ElemName {
					sign := hexutil.Encode(function.ID)
					if contractElemIdToConcreteDicCache[sign] == nil {
						contractElemIdToConcreteDicCache[sign] = make([]richtypes.ContractElemConcrete, 0)
					}

					contractElemIdToConcreteDicCache[sign] = append(contractElemIdToConcreteDicCache[sign], contrete)
					// event name in contract is unique, so jump out of loop
					break
				}
			}

		}
	}

	return contractElemIdToConcreteDicCache, nil
}

// GetMatchedConcrete ...
func (cd *ContractDecoder) GetMatchedConcrete(log *types.LogEntry) (*richtypes.ContractElemConcrete, error) {
	if len(log.Topics) == 0 {
		return nil, nil
	}
	// event parameters of abi needs be "from" "to" "value"
	contretes := cd.ElemIdToConcreteDicCache[log.Topics[0].String()]

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
	concrete := cd.ElemIdToConcreteDicCache[hexutil.Encode(data[:4])]

	if concrete != nil && len(concrete) > 0 {
		return concrete[0].DecodeFunction(data)
	}
	return nil, nil
}
