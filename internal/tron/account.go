package tron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type AccountClient interface {
	GetAccount(address string) (*Account, error)
	ListWitnesses() (*Witnesses, error)
	GetWitnesses(address string) (*Witness, error)
	GetBrokerage(address string) (int, error)
}

type AccountClientImpl struct {
	client *Client
}

func NewAccountClient(client *Client) *AccountClientImpl {
	return &AccountClientImpl{client: client}
}

func (c *AccountClientImpl) GetAccount(address string) (*Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestURL, err := url.JoinPath(c.client.baseURL.String(), APIGetAccountInfoEndpoint)
	if err != nil {
		return nil, fmt.Errorf("GetAccount: failed to build request URL (address: %s, error: %w)", address, err)
	}

	payload, err := json.Marshal(map[string]interface{}{"address": address, "visible": true})
	if err != nil {
		return nil, fmt.Errorf("GetAccount: failed to marshal request payload (address: %s, error: %w)", address, err)
	}
	bodyReader := bytes.NewReader(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("GetAccount: failed to create request (address: %s, error: %w)", address, err)
	}

	res, err := c.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetAccount: HTTP request failed (address: %s, error: %w)", address, err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("GetAccount: request failed (address: %s, status: %d, response: %s)", address, res.StatusCode, string(resBody))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("GetAccount: failed to read response body (address: %s, error: %w)", address, err)
	}

	account := &Account{}
	if err := json.Unmarshal(resBody, account); err != nil {
		return nil, fmt.Errorf("GetAccount: failed to decode response (address: %s, error: %w)", address, err)
	}

	witnessInfo, err := c.GetWitnesses(ConvertAddressToHex(address))
	if err != nil {
		return nil, fmt.Errorf("GetAccount: failed to retrieve witness info (address: %s, error: %w)", address, err)
	}
	account.WitnessInfo = witnessInfo

	return account, nil
}

func (c *AccountClientImpl) ListWitnesses() (*Witnesses, error) {
	witnesses := &Witnesses{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestURL, err := url.JoinPath(c.client.baseURL.String(), APIListWitnessesEndpoint)
	if err != nil {
		return nil, fmt.Errorf("ListWitnesses: failed to build request URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ListWitnesses: failed to build the request: %w", err)
	}

	res, err := c.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ListWitnesses: HTTP request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("ListWitnesses: HTTP request failed (status: %d): %s", res.StatusCode, string(resBody))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("ListWitnesses: unable to read response: %v", err)
	}
	if err := json.Unmarshal(resBody, witnesses); err != nil {
		return nil, fmt.Errorf("ListWitnesses: unable to decode response: %v", err)
	}

	return witnesses, nil
}
func (c *AccountClientImpl) GetWitnesses(address string) (*Witness, error) {
	witness := &Witness{}
	witnesses, err := c.ListWitnesses()
	if err != nil {
		return nil, fmt.Errorf("GetWitnesses: failed to retrieve witnesses: error: %w", err)
	}

	sort.SliceStable(witnesses.Witnesses, func(i, j int) bool {
		voteCountI, okI := witnesses.Witnesses[i].VoteCount, true
		voteCountJ, okJ := witnesses.Witnesses[j].VoteCount, true

		if !okI {
			voteCountI = 0
		}
		if !okJ {
			voteCountJ = 0
		}

		return voteCountI > voteCountJ
	})

	for index, w := range witnesses.Witnesses {
		if w.Address == address {
			witness = &w
			witness.Rank = index + 1
			break
		}
	}
	return witness, nil
}

// GetBrokerage returns the validator's brokerage (commission) rate, in percent,
// as reported by the node's /wallet/getBrokerage endpoint.
func (c *AccountClientImpl) GetBrokerage(address string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestURL, err := url.JoinPath(c.client.baseURL.String(), APIGetBrokerageEndpoint)
	if err != nil {
		return 0, fmt.Errorf("GetBrokerage: failed to build request URL (address: %s, error: %w)", address, err)
	}

	payload, err := json.Marshal(map[string]interface{}{"address": address, "visible": true})
	if err != nil {
		return 0, fmt.Errorf("GetBrokerage: failed to marshal request payload (address: %s, error: %w)", address, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("GetBrokerage: failed to create request (address: %s, error: %w)", address, err)
	}

	res, err := c.client.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("GetBrokerage: HTTP request failed (address: %s, error: %w)", address, err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(res.Body)
		return 0, fmt.Errorf("GetBrokerage: request failed (address: %s, status: %d, response: %s)", address, res.StatusCode, string(resBody))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("GetBrokerage: failed to read response body (address: %s, error: %w)", address, err)
	}

	var out struct {
		Brokerage int `json:"brokerage"`
	}
	if err := json.Unmarshal(resBody, &out); err != nil {
		return 0, fmt.Errorf("GetBrokerage: failed to decode response (address: %s, error: %w)", address, err)
	}

	return out.Brokerage, nil
}
