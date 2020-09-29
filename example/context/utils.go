package context

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Conflux-Chain/go-conflux-sdk/types"
)

func JsonFmt(v interface{}) string {
	j, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	var str bytes.Buffer
	_ = json.Indent(&str, j, "", "    ")
	return str.String()
}

func WaitPacked(txhash types.Hash) {
	fmt.Println("wait for transaction be packed")
	for {
		time.Sleep(time.Duration(1) * time.Second)
		txReceipt, err := client.GetTransactionReceipt(txhash)
		if err != nil {
			panic(err)
		}
		if txReceipt != nil {
			fmt.Printf("transaction is packed:%+v\n\n", JsonFmt(txReceipt))
			break
		}
		fmt.Printf(".")
	}
}
