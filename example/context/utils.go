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
		tx, err := client.GetTransactionByHash(txhash)
		if err != nil {
			panic(err)
		}
		if tx.Status != nil {
			fmt.Printf("transaction is packed:%+v\n\n", JsonFmt(tx))
			break
		}
		fmt.Printf(".")
	}
}
