package decoder

import (
	"math/big"
	"reflect"
	"testing"

	richtypes "github.com/Conflux-Chain/go-conflux-sdk-for-wallet/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

func TestDecode(t *testing.T) {
	value, _ := big.NewInt(0).SetString("0x000000000000000000000000000000000000000000000000000000000000000a", 0)

	datas := []struct {
		expect interface{}
		log    types.LogEntry
	}{
		// erc20
		{
			expect: &richtypes.ERC20TokenTransferEventParams{
				TokenTransferEventParams: richtypes.TokenTransferEventParams{
					From: *types.NewAddress("0x00000000000000000000000019f4bcf113e0b896d9b34294fd3da86b4adf0302").ToCommonAddress(),
					To:   *types.NewAddress("0x000000000000000000000000160ebef20c1f739957bf9eecd040bce699cc42c6").ToCommonAddress(),
				},
				Value: value,
			},
			log: types.LogEntry{
				Topics: []types.Hash{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
					"0x00000000000000000000000019f4bcf113e0b896d9b34294fd3da86b4adf0302",
					"0x000000000000000000000000160ebef20c1f739957bf9eecd040bce699cc42c6"},
				Data: "0x000000000000000000000000000000000000000000000000000000000000000a",
			},
		},

		// erc721
		{
			expect: &richtypes.ERC721TokenTransferEventParams{
				TokenTransferEventParams: richtypes.TokenTransferEventParams{
					From: *types.NewAddress("0x00000000000000000000000019f4bcf113e0b896d9b34294fd3da86b4adf0302").ToCommonAddress(),
					To:   *types.NewAddress("0x000000000000000000000000160ebef20c1f739957bf9eecd040bce699cc42c6").ToCommonAddress(),
				},
				TokenId: value,
			},
			log: types.LogEntry{
				Topics: []types.Hash{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
					"0x00000000000000000000000019f4bcf113e0b896d9b34294fd3da86b4adf0302",
					"0x000000000000000000000000160ebef20c1f739957bf9eecd040bce699cc42c6",
					"0x000000000000000000000000000000000000000000000000000000000000000a"},
			},
		},

		// erc777
		{
			expect: &richtypes.ERC777TokenTransferEventParams{
				TokenTransferEventParams: richtypes.TokenTransferEventParams{
					From: *types.NewAddress("0x0000000000000000000000001195c6b43264113a75719202716cc763bacb7da5").ToCommonAddress(),
					To:   *types.NewAddress("0x000000000000000000000000154F4d3229416B47732D93a8c5E42e481794Aff8").ToCommonAddress(),
				},
				Amount:       value,
				Operator:     *types.NewAddress("0x0000000000000000000000001195c6b43264113a75719202716cc763bacb7da5").ToCommonAddress(),
				Data:         []byte{0x12, 0x34, 0x56},
				OperatorData: []byte{},
			},
			log: types.LogEntry{
				Topics: []types.Hash{"0x06b541ddaa720db2b10a4d0cdac39b8d360425fc073085fac19bc82614677987",
					"0x0000000000000000000000001195c6b43264113a75719202716cc763bacb7da5",
					"0x0000000000000000000000001195c6b43264113a75719202716cc763bacb7da5",
					"0x000000000000000000000000154f4d3229416b47732d93a8c5e42e481794aff8"},
				Data: "0x000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000003123456000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}

	for _, data := range datas {
		eventDecoder, err := NewContractDecoder()
		if err != nil {
			t.Error(err)
		}

		actual, err := eventDecoder.DecodeEvent(&data.log)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(actual, data.expect) {
			t.Errorf("expect: %#v, acutal: %#v", data.expect, actual)
		}
	}

}
