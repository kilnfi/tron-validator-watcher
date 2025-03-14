package tron

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFirstBlock(t *testing.T) {
	t.Parallel()

	type inputFields struct {
		blockNum       int64
		blockTimestamp int64
	}

	cases := []struct {
		name          string
		inputs        inputFields
		expectedBlock *Block
		mockResponse  *Block
		wantAPIError  bool
		wantErr       bool
		errorMessage  string
	}{
		{
			name: "Return_No_Error",
			inputs: inputFields{
				blockNum:       5000,
				blockTimestamp: 1740668400000,
			},
			mockResponse: &Block{
				BlockHeader: BlockHeader{
					RawData: BlockHeaderRawData{
						Number:    1400,
						Timestamp: 1740657600000,
					},
				},
			},
			expectedBlock: &Block{
				BlockHeader: BlockHeader{
					RawData: BlockHeaderRawData{
						Number:    1400,
						Timestamp: 1740657600000,
					},
				},
			},
			wantAPIError: false,
			wantErr:      false,
		},
		{
			name: "Return_Error_When_Block_Not_Found",
			inputs: inputFields{
				blockNum:       0,
				blockTimestamp: 0,
			},
			mockResponse:  nil,
			expectedBlock: nil,
			wantAPIError:  true,
			wantErr:       true,
			errorMessage:  "block not found",
		},
		{
			name: "Return_Error_When_No_First_Block_Found",
			inputs: inputFields{
				blockNum:       5000,
				blockTimestamp: 1740668400000,
			},
			mockResponse: &Block{
				BlockHeader: BlockHeader{
					RawData: BlockHeaderRawData{
						Number:    1300,
						Timestamp: 1740664800000,
					},
				},
			},
			expectedBlock: &Block{
				BlockHeader: BlockHeader{
					RawData: BlockHeaderRawData{
						Number:    1400,
						Timestamp: 1740657600000,
					},
				},
			},
			wantAPIError: false,
			wantErr:      true,
			errorMessage: "unable to find the first block for the current round",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()

			if c.wantErr && c.wantAPIError {
				mux.HandleFunc(APIGetBlockByNumEndpoint, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)

					if _, err := w.Write([]byte("block not found")); err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
				})
			} else {
				mux.HandleFunc(APIGetBlockByNumEndpoint, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)

					data, err := json.Marshal(c.mockResponse)
					if err != nil {
						t.Fatalf("failed to marshal block: %v", err)
					}

					if _, err := w.Write(data); err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
				})
			}

			client := setup(t, mux)
			block, err := GetFirstBlockNum(context.Background(), client, c.inputs.blockNum, c.inputs.blockTimestamp)
			if c.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errorMessage)
				require.Nil(t, block)
			} else {
				require.Equal(t, c.expectedBlock.BlockHeader.RawData.Number, block.BlockHeader.RawData.Number)
				require.Equal(t, c.expectedBlock.BlockHeader.RawData.Timestamp, block.BlockHeader.RawData.Timestamp)
			}
		})
	}
}

// TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ
func TestConvertAddressToHex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Return_No_Error_When_Valid_Address",
			input:    "TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ",
			expected: "41a4ce68cfcdd27884bde52cec653354048e0aa989",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			hex := ConvertAddressToHex(c.input)
			require.Equal(t, c.expected, hex)
		})
	}
}

func TestConvertAddressToBase58(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		input        string
		expected     string
		wantErr      bool
		errorMessage string
	}{
		{
			name:     "Return_No_Error_When_Valid_Hex",
			input:    "41a4ce68cfcdd27884bde52cec653354048e0aa989",
			expected: "TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ",
			wantErr:  false,
		},
		{
			name:         "Return_Error_When_Invalid_Hex",
			input:        "invalid",
			expected:     "",
			wantErr:      true,
			errorMessage: "unable to decode hex address",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			address, err := ConvertAddressToBase58(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMessage)
				require.Empty(t, address)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, address)
			}
		})
	}
}

func TestGetEpochID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    *Block
		expected int
	}{
		{
			name: "Return_No_Error_When_Valid_Block",
			input: &Block{
				BlockHeader: BlockHeader{
					RawData: BlockHeaderRawData{
						Timestamp: 1741932000000,
					},
				},
			},
			expected: 80645,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			epochID := GetEpochID(c.input)
			require.Equal(t, c.expected, epochID)
		})
	}
}
