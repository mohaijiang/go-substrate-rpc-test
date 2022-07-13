package test

import (
	"fmt"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApi(t *testing.T) {
	api, err := gsrpc.NewSubstrateAPI("ws://183.66.65.207:49944")
	if err != nil {
		panic(err)
	}

	chain, err := api.RPC.System.Chain()
	if err != nil {
		panic(err)
	}
	nodeName, err := api.RPC.System.Name()
	if err != nil {
		panic(err)
	}
	nodeVersion, err := api.RPC.System.Version()
	if err != nil {
		panic(err)
	}

	fmt.Printf("You are connected to chain %v using %v v%v\n", chain, nodeName, nodeVersion)
}

func TestDecodeEvent(t *testing.T) {
	// 区块链浏览器：  https://polkadot-ui.test.hamsternet.io/
	// go-rpc 库地址： https://pkg.go.dev/github.com/centrifuge/go-substrate-rpc-client/v4
	// 类型映射关系   rust u32 => go types.U32
	// 				AccountId => go types.AccountId

	api, err := gsrpc.NewSubstrateAPI("wss://ws.test.hamsternet.io")
	meta, err := api.RPC.State.GetMetadataLatest()
	assert.NoError(t, err)
	// 836379 成功区块号， 836391 失败区块号
	bh, err := api.RPC.Chain.GetBlockHash(836391)
	assert.NoError(t, err)
	key, err := types.CreateStorageKey(meta, "System", "Events", nil)
	assert.NoError(t, err)
	raw, err := api.RPC.State.GetStorageRaw(key, bh)
	assert.NoError(t, err)
	// Decode the event records
	events := MyEventRecords{}

	//err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)

	err = DecodeEventRecordsWithIgnoreError(types.EventRecordsRaw(*raw), meta, &events)

	fmt.Println(events)
}
