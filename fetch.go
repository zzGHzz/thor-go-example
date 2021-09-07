package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/vechain/thor/api/blocks"
	"github.com/vechain/thor/thor"
)

var (
	genesis *blocks.JSONCollapsedBlock
	best    atomic.Value
)

func fetchLoop(ctx context.Context, url string) {
	fmt.Printf("Fetch Loop: opened\n\n")
	
	ticker := time.NewTicker(time.Duration(8) * time.Second)

	var err error
	genesis, err = getBlockByNumber(url, 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fetch Loop: fetched genesis with id = %#x\n", genesis.ID.Bytes())
	fmt.Printf("Fetch Loop: chainTag = %#x\n", chainTag())
	fmt.Printf("Fetch Loop: if connected to testnet: %t\n", isTestnet(genesis))
	fmt.Printf("Fetch Loop: if connected to mainnet: %t\n\n", isMainnet(genesis))

	fetch := func() {
		if blk, err := getBestBlock(url); err != nil {
			panic(err)
		} else {
			id := bestID()
			if id.IsZero() || (blk.IsTrunk && number(id) < blk.Number) {
				best.Store(blk.ID)
				fmt.Printf("Fetch Loop: updated best (%d)\n\n", blk.Number)
			}
		}
	}

	fetch()

	for {
		select {
		case <-ticker.C:
			fetch()
		case <-ctx.Done():
			fmt.Printf("Fetch Loop: closed\n\n")
			return
		}
	}
}

func number(id thor.Bytes32) uint32 {
	return binary.BigEndian.Uint32(id[:])
}

func bestID() thor.Bytes32 {
	if id := best.Load(); id != nil {
		return id.(thor.Bytes32)
	}
	return thor.Bytes32{}
}
