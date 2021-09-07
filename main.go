package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vechain/thor/abi"
	"github.com/vechain/thor/api/accounts"
	"github.com/vechain/thor/thor"
)

func testBlockAPI(url string) {
	/////////////////////////////////////////////
	// get block by number
	/////////////////////////////////////////////
	n := uint64(10000)
	if blk, err := getBlockByNumber(url, n); err != nil {
		panic(err)
	} else {
		fmt.Printf("Main Process: get block by number: \n\t%s\n\n", getJSONString(blk))
	}

	/////////////////////////////////////////////
	// get block by id
	/////////////////////////////////////////////
	blkId, _ := decodeBytes32("0x00998489d233e148106239247926290dfe5b29295830d56eba7a8572457463a9")
	if blk, err := getBlockById(url, blkId); err != nil {
		panic(err)
	} else {
		fmt.Printf("Main Process: get block by id: \n\t%s\n\n", getJSONString(blk))
	}
}

func testTxAPI(url string) {
	/////////////////////////////////////////////
	// get tx & receipt
	/////////////////////////////////////////////
	txId, _ := decodeBytes32("0xe2c7f2a0193d8faaf5db81731b225b5520af7332e48915ea2612a37bba072def")
	fmt.Println("testTxAPI: get tx & receipt")
	if tx, err := getTx(url, txId); err != nil {
		panic(err)
	} else {
		fmt.Printf("\tTransaction: %s\n", getJSONString(tx))
	}
	if r, err := getReceipt(url, txId); err != nil {
		panic(err)
	} else {
		fmt.Printf("\tReceipt: %s\n\n", getJSONString(r))
	}

	/////////////////////////////////////////////
	// Send 1 VET
	/////////////////////////////////////////////
	pk, _ := crypto.HexToECDSA("3e61996a0a49b26a5608a55a3e0669aff271959d2e43658766e3514a07a5ccf3")
	from, _ := thor.ParseAddress("0xCDFDFD58e986130445B560276f52CE7985809238")
	to, _ := thor.ParseAddress("0x17faD1428DC464187C33C5EfD281aE7E58937Fd8")

	val := math.HexOrDecimal256(*big.NewInt(1e18))
	clauses := []accounts.Clause{
		{
			To:    &to,
			Value: &val,
			Data:  "0x",
		},
	}
	estimatedGas, err := estimateGas(url, from, clauses)
	if err != nil {
		panic(err)
	}

	raw, err := signTx(pk, from, clauses, TxOption{Gas: estimatedGas})
	if err != nil {
		panic(err)
	}

	if txId, err := sendRawTx(url, raw); err != nil {
		panic(err)
	} else {
		fmt.Printf("testTxAPI: send 1 VET: \n\ttxid = %s\n\n", txId.String())
	}

	/////////////////////////////////////////////
	// Call contract method to transfer 1 VTHO
	/////////////////////////////////////////////
	contractAddr := thor.BytesToAddress([]byte("Energy"))
	abiStr := `[{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"supply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_from","type":"address"},{"indexed":true,"name":"_to","type":"address"},{"indexed":false,"name":"_value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_owner","type":"address"},{"indexed":true,"name":"_spender","type":"address"},{"indexed":false,"name":"_value","type":"uint256"}],"name":"Approval","type":"event"}]`
	ABI, err := abi.New([]byte(abiStr))
	if err != nil {
		panic(err)
	}
	method, ok := ABI.MethodByName("transfer")
	if !ok {
		panic("ABI method not found")
	}
	data, err := method.EncodeInput(to, big.NewInt(1e18))
	if err != nil {
		panic(err)
	}

	clauses = []accounts.Clause{
		{
			To:    &contractAddr,
			Value: &math.HexOrDecimal256{},
			Data:  hexutil.Encode(data),
		},
	}

	// Dry run to check whether the tx would be reverted or not
	res, err := dryRun(url, from, clauses)
	if err != nil {
		panic(err)
	}
	if res[0].Reverted {
		panic("tx to be reverted")
	}

	estimatedGas, err = estimateGas(url, from, clauses)
	if err != nil {
		panic(err)
	}

	raw, err = signTx(pk, from, clauses, TxOption{Gas: estimatedGas})
	if err != nil {
		panic(err)
	}

	if txId, err := sendRawTx(url, raw); err != nil {
		panic(err)
	} else {
		fmt.Printf(`testTxAPI: send 1 VTHO:
	txid = %s
	from = %s
	to = %s`+"\n\n", txId.String(), from.String(), to.String())
	}
}

func testAccAPI(url string) {
	/////////////////////////////////////////////
	// get account balance
	/////////////////////////////////////////////
	addr, _ := decodeAddress("0xCDFDFD58e986130445B560276f52CE7985809238")
	if acc, err := getAccount(url, addr); err != nil {
		panic(err)
	} else {
		fmt.Printf("testAccAPI: get account balance: \n\t%s\n", getJSONString(acc))
		fmt.Printf("\tVET balance: %s\n", printBalance(acc.Balance, 2))
		fmt.Printf("\tVTHO balance: %s\n", printBalance(acc.Energy, 2))
		fmt.Printf("\tIf has code: %t\n\n", acc.HasCode)
	}
}

func main() {
	// public node for vechain testnet
	url := "https://sync-testnet.veblocks.net"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// go routine that fetches the best block id
	go fetchLoop(ctx, url)

	ticker := time.NewTicker(time.Second * time.Duration(1))
	for {
		<-ticker.C
		if !bestID().IsZero() {
			break
		}
	}

	testBlockAPI(url)
	testAccAPI(url)
	testTxAPI(url)
}
