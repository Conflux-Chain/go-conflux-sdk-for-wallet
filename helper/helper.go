package helper

import (
	"fmt"
	"os"

	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/ethereum/go-ethereum/common"
)

// IsFileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// PanicIfErrf panic and reports error message with args
func PanicIfErrf(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Printf(msg, args...)
		fmt.Println()
		panic(err)
	}
}

// PanicIfErr panic and reports error message
func PanicIfErr(err error, msg string) {
	if err != nil {
		fmt.Printf(msg)
		fmt.Println()
		panic(err)
	}
}

func MustGetCommonAddressPtr(address *cfxaddress.Address) *common.Address {
	if address == nil {
		return nil
	}
	commonAddr := address.MustGetCommonAddress()
	return &commonAddr
}

func MustNewCfxAddressPtr(commonAddress *common.Address, networkID uint32) *types.Address {
	if commonAddress == nil {
		return nil
	}
	cfxAddr := cfxaddress.MustNewFromCommon(*commonAddress, networkID)
	return &cfxAddr
}
