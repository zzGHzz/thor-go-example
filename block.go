package main

import (
	"fmt"
	"net/http"

	"github.com/vechain/thor/api/blocks"
	"github.com/vechain/thor/api/utils"
	"github.com/vechain/thor/thor"
)

func getBlockByNumber(url string, n uint64) (*blocks.JSONCollapsedBlock, error) {
	resp, err := getHttpResp("GET", fmt.Sprintf("%s%d", trim(url)+"/blocks/", n), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return getBlock(resp)
}

func getBlockById(url string, id thor.Bytes32) (*blocks.JSONCollapsedBlock, error) {
	resp, err := getHttpResp("GET", trim(url)+"/blocks/"+id.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return getBlock(resp)
}

func getBestBlock(url string) (*blocks.JSONCollapsedBlock, error) {
	resp, err := getHttpResp("GET", trim(url)+"/blocks/best", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return getBlock(resp)
}

func getBlock(resp *http.Response) (*blocks.JSONCollapsedBlock, error) {
	blk := new(blocks.JSONCollapsedBlock)

	if err := utils.ParseJSON(resp.Body, &blk); err != nil {
		return nil, err
	}

	return blk, nil
}
