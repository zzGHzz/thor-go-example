package main

import (
	"bytes"
	"encoding/json"

	"github.com/vechain/thor/api/accounts"
	"github.com/vechain/thor/api/utils"
	"github.com/vechain/thor/thor"
)

func getAccount(url string, addr thor.Address) (*accounts.Account, error) {
	resp, err := getHttpResp("GET", trim(url)+"/accounts/"+addr.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	acc := new(accounts.Account)
	if err := utils.ParseJSON(resp.Body, &acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func dryRun(url string, from thor.Address, clauses []accounts.Clause) ([]*accounts.CallResult, error) {
	body := new(struct {
		Caller  *thor.Address    `json:"caller"`
		Clauses accounts.Clauses `json:"clauses"`
	})
	body.Clauses = append(body.Clauses, clauses...)
	body.Caller = &from

	jsonValue, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := getHttpResp("POST", trim(url)+"/accounts/*", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	batch := new(accounts.BatchCallResults)
	if err := utils.ParseJSON(resp.Body, batch); err != nil {
		return nil, err
	}

	return *batch, nil
}
