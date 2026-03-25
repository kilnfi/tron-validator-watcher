package tron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type NetworkClient interface {
	GetLatestBlock(ctx context.Context) (*Block, error)
	GetBlockByNumber(ctx context.Context, number int64) (*Block, error)
}

type NetworkClientImpl struct {
	client *Client
}

type Block struct {
	BlockID      string        `json:"blockID"`
	BlockHeader  BlockHeader   `json:"block_header"`
	Transactions []Transaction `json:"transactions"`
}

// BlockHeader represents a block header.
type BlockHeader struct {
	RawData          BlockHeaderRawData `json:"raw_data"`
	WitnessSignature string             `json:"witness_signature"`
}

// BlockHeaderRawData represents the raw data of a block header.
type BlockHeaderRawData struct {
	Number         int64  `json:"number"`
	TxTrieRoot     string `json:"txTrieRoot"`
	WitnessAddress string `json:"witness_address"`
	ParentHash     string `json:"parentHash"`
	Version        int    `json:"version"`
	Timestamp      int64  `json:"timestamp"`
}

// Transaction represents a Tron transaction.
type Transaction struct {
	Ret        []TransactionResult `json:"ret"`
	Signature  []string            `json:"signature"`
	TxID       string              `json:"txID"`
	RawData    TransactionRawData  `json:"raw_data"`
	RawDataHex string              `json:"raw_data_hex"`
}

// TransactionResult represents the result of a transaction.
type TransactionResult struct {
	ContractRet string `json:"contractRet"`
}

// TransactionRawData represents the raw data of a transaction.
type TransactionRawData struct {
	Contract      []Contract `json:"contract"`
	RefBlockBytes string     `json:"ref_block_bytes"`
	RefBlockHash  string     `json:"ref_block_hash"`
	Expiration    int64      `json:"expiration"`
	FeeLimit      int64      `json:"fee_limit,omitempty"`
	Timestamp     int64      `json:"timestamp"`
}

// Contract represents a contract within a transaction.
type Contract struct {
	Parameter    ContractParameter `json:"parameter"`
	Type         string            `json:"type"`
	PermissionID int               `json:"Permission_id,omitempty"`
}

// ContractParameter represents the parameters of a contract.
type ContractParameter struct {
	Value   ContractValue `json:"value"`
	TypeURL string        `json:"type_url"`
}

// ContractValue represents the possible values of contract parameters.
type ContractValue struct {
	Data            string `json:"data,omitempty"`
	OwnerAddress    string `json:"owner_address"`
	ContractAddress string `json:"contract_address,omitempty"`
	ToAddress       string `json:"to_address,omitempty"`
	Amount          int64  `json:"amount,omitempty"`
	Balance         int64  `json:"balance,omitempty"`
	Resource        string `json:"resource,omitempty"`
	ReceiverAddress string `json:"receiver_address,omitempty"`
	Lock            bool   `json:"lock,omitempty"`
	LockPeriod      int    `json:"lock_period,omitempty"`
}

func NewNetworkClient(client *Client) *NetworkClientImpl {
	return &NetworkClientImpl{
		client: client,
	}
}

func (c *NetworkClientImpl) GetLatestBlock(ctx context.Context) (*Block, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	requestURL, err := url.JoinPath(c.client.baseURL.String(), APIGetLatestBlockEndpoint)
	if err != nil {
		return nil, fmt.Errorf("GetLatestBlock: failed to build request url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("GetLatestBlock: failed to create http request: %w", err)
	}

	block := &Block{}
	res, err := c.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetLatestBlock - HTTP request failed: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("GetLatestBlock: HTTP request failed (status: %d): %s)", res.StatusCode, string(resBody))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("GetLatestBlock: failed to read response body: %v", err)
	}
	if err := json.Unmarshal(resBody, block); err != nil {
		return nil, fmt.Errorf("GetLatestBlock: failed to decode response: %v", err)
	}

	return block, nil
}

func (c *NetworkClientImpl) GetBlockByNumber(ctx context.Context, number int64) (*Block, error) {
	block := &Block{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	requestURL, err := url.JoinPath(c.client.baseURL.String(), APIGetBlockByNumEndpoint)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber: failed to build request URL: %w", err)
	}

	payload, err := json.Marshal(
		map[string]int64{
			"num": number,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber: failed to build payload: %w", err)
	}
	bodyReader := bytes.NewReader(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber: failed to build HTTP request: %w", err)
	}
	res, err := c.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber: HTTP request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("GetBlockByNumber: HTTP request failed (status: %d): %s", res.StatusCode, string(resBody))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber: failed to read response body: %w", err)
	}
	if err := json.Unmarshal(resBody, block); err != nil {
		return nil, fmt.Errorf("GetBlockByNumber: failed to decode response: %w", err)
	}
	return block, nil
}
