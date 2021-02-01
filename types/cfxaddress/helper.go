package cfxaddress

import "github.com/Conflux-Chain/go-conflux-sdk/types"

// FormatAddressStrToHex format hex or base32 address to hex string
func FormatAddressStrToHex(address string) string {
	if address == "" || address[0:2] == "0x" {
		return address
	}
	cfxAddr := MustNewFromBase32(address)
	return "0x" + cfxAddr.GetHexAddress()
}

func FormatAddressToHex(address types.Address) types.Address {
	formated := FormatAddressStrToHex(address.String())
	return types.Address(formated)
}
