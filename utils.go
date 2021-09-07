package main

import (
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/vechain/thor/api/blocks"
	"github.com/vechain/thor/thor"
)

func validateHeader(resp *http.Response) bool {
	if genesis == nil {
		return true
	}
	return resp.Header["X-Genesis-Id"][0] == genesis.ID.String()
}

func chainTag() byte {
	id := genesis.ID.Bytes()
	return id[len(id)-1]
}

func isTestnet(genesis *blocks.JSONCollapsedBlock) bool {
	if genesis == nil {
		return false
	}
	return chainTag() == 0x27
}

func isMainnet(genesis *blocks.JSONCollapsedBlock) bool {
	if genesis == nil {
		return false
	}
	return chainTag() == 0x4a
}

func trim(url string) string {
	return strings.TrimSuffix(url, "/")
}

func decodeBytes32(s string) (thor.Bytes32, error) {
	return thor.ParseBytes32(s)
}

func decodeAddress(s string) (thor.Address, error) {
	return thor.ParseAddress(s)
}

func getHttpResp(method, url string, body io.Reader) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	client := http.Client{Timeout: time.Duration(10) * time.Second}

	switch method {
	case "GET":
		resp, err = client.Get(url)
	case "POST":
		resp, err = client.Post(url, "application/json", body)
	default:
		return nil, errors.New("invalid method")
	}

	if err != nil {
		return nil, err
	}

	if !validateHeader(resp) {
		resp.Body.Close()
		return nil, errors.New("genesis id mismatched")
	}

	if resp.Status != "200 OK" {
		resp.Body.Close()
		return nil, errors.New(resp.Status)
	}

	return resp, nil
}

func printBalance(n math.HexOrDecimal256, prec int) string {
	x := (*big.Int)(&n)
	return new(big.Rat).Quo(new(big.Rat).SetInt(x), new(big.Rat).SetInt(big.NewInt(1e18))).FloatString(prec)
}

func getJSONString(obj interface{}) string {
	msg, _ := json.Marshal(obj)
	return string(msg)
}
