package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/api/accounts"
	"github.com/vechain/thor/api/transactions"
	"github.com/vechain/thor/api/utils"
	"github.com/vechain/thor/thor"
	"github.com/vechain/thor/tx"
)

func getTx(url string, id thor.Bytes32) (*transactions.Transaction, error) {
	resp, err := getHttpResp("GET", trim(url)+"/transactions/"+id.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tx := new(transactions.Transaction)
	if err := utils.ParseJSON(resp.Body, &tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func getReceipt(url string, id thor.Bytes32) (*transactions.Receipt, error) {
	resp, err := getHttpResp("GET", trim(url)+"/transactions/"+id.String()+"/receipt", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := new(transactions.Receipt)
	if err := utils.ParseJSON(resp.Body, &r); err != nil {
		return nil, err
	}

	return r, nil
}

type TxOption struct {
	GasPriceCoef uint8
	Expiration   uint32
	Gas          uint64
}

var (
	DefaultExpiration uint32 = 18
)

func signTx(pk *ecdsa.PrivateKey, from thor.Address, clauses []accounts.Clause, opt TxOption) ([]byte, error) {
	best := bestID()
	if best.IsZero() {
		return nil, errors.New("zero best block id")
	}

	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	f := new(tx.Features)
	f.SetDelegated(false)

	builder := new(tx.Builder).
		ChainTag(chainTag()).
		Gas(opt.Gas).
		GasPriceCoef(opt.GasPriceCoef).
		BlockRef(tx.NewBlockRefFromID(best)).
		DependsOn(nil).
		Nonce(binary.BigEndian.Uint64(nonce)).
		Features(*f)

	for _, c := range clauses {
		clause := tx.NewClause(c.To).WithValue((*big.Int)(c.Value))
		data, err := hexutil.Decode(c.Data)
		if err != nil {
			return nil, err
		}
		clause = clause.WithData(data)
		builder.Clause(clause)
	}

	if opt.Expiration > 0 {
		builder.Expiration(opt.Expiration)
	} else {
		builder.Expiration(DefaultExpiration)
	}

	t := builder.Build()
	sig, err := crypto.Sign(t.SigningHash().Bytes(), pk)
	if err != nil {
		return nil, err
	}
	t = t.WithSignature(sig)

	raw, err := rlp.EncodeToBytes(t)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func sendRawTx(url string, raw []byte) (thor.Bytes32, error) {
	body := transactions.RawTx{Raw: hexutil.Encode(raw)}

	fmt.Println(getJSONString(body))

	jsonValue, err := json.Marshal(body)
	if err != nil {
		return thor.Bytes32{}, err
	}
	resp, err := getHttpResp("POST", trim(url)+"/transactions", bytes.NewBuffer(jsonValue))
	if err != nil {
		return thor.Bytes32{}, err
	}
	defer resp.Body.Close()

	txResponse := new(struct {
		TxId thor.Bytes32 `json:"id"`
	})
	if err := utils.ParseJSON(resp.Body, txResponse); err != nil {
		return thor.Bytes32{}, err
	}

	return txResponse.TxId, nil
}

func estimateGas(url string, from thor.Address, clauses []accounts.Clause) (uint64, error) {
	var sum uint64

	var txClauses []*tx.Clause
	for _, c := range clauses {
		clause := tx.NewClause(c.To).WithValue((*big.Int)(c.Value))
		data, err := hexutil.Decode(c.Data)
		if err != nil {
			return 0, err
		}
		clause = clause.WithData(data)
		txClauses = append(txClauses, clause)
	}

	if g, err := tx.IntrinsicGas(txClauses...); err != nil {
		return 0, err
	} else {
		sum += g
	}

	res, err := dryRun(url, from, clauses)
	if err != nil {
		return 0, err
	}

	var execGas uint64
	for _, r := range res {
		execGas += r.GasUsed
	}

	if execGas > 0 {
		execGas += 15000
	}

	sum += execGas

	return sum, nil
}
