package tron

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLatestBlock(t *testing.T) {
	t.Parallel()
	t.Run("Return_No_Error_When_Valid_Response", func(t *testing.T) {
		t.Parallel()

		expectedBlockID := int64(10)

		mux := http.NewServeMux()
		mux.HandleFunc(APIGetLatestBlockEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			file, err := os.Open("testdata/block.json")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})

		client := setup(t, mux)
		block, err := client.Network.GetLatestBlock(context.Background())
		require.NoError(t, err)
		require.NotNil(t, block)
		require.Equal(t, expectedBlockID, block.BlockHeader.RawData.Number)
	})
	t.Run("Return_Error_When_Building_Request_Url", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		client := setup(t, mux)
		client.baseURL.Host = "htt://localhost:1234"

		block, err := client.Network.GetLatestBlock(context.Background())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to build request url")
		require.Nil(t, block)
	})
	t.Run("Return_Error_When_Request_Fails", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIGetLatestBlockEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			if _, err := w.Write([]byte("bad request")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		client := setup(t, mux)
		block, err := client.Network.GetLatestBlock(context.Background())
		require.Error(t, err)
		require.Contains(t, err.Error(), "HTTP request failed")
		require.Nil(t, block)
	})
	t.Run("Return_Error_When_Invalid_Response", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIGetLatestBlockEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Length", "50")
			w.WriteHeader(http.StatusOK)

			if _, err := w.Write([]byte("invalid response")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		client := setup(t, mux)
		block, err := client.Network.GetLatestBlock(context.Background())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read response body")
		require.Nil(t, block)
	})
	t.Run("Return_Error_When_Decode_Response", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIGetLatestBlockEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			file, err := os.Open("testdata/invalid_block.json")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})
		client := setup(t, mux)
		block, err := client.Network.GetLatestBlock(context.Background())
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decode response")
		require.Nil(t, block)
	})
}
func TestGetBlockByNumber(t *testing.T) {
	t.Parallel()
	t.Run("Return_No_Error_When_Valid_Response", func(t *testing.T) {
		t.Parallel()

		blockID := int64(10)
		mux := http.NewServeMux()
		mux.HandleFunc(APIGetBlockByNumEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			file, err := os.Open("testdata/block.json")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})
		client := setup(t, mux)
		block, err := client.Network.GetBlockByNumber(context.Background(), blockID)
		require.NoError(t, err)
		require.NotNil(t, block)
		require.Equal(t, blockID, block.BlockHeader.RawData.Number)
	})
	t.Run("Return_Error_When_Building_Request_Url", func(t *testing.T) {
		t.Parallel()

		blockID := int64(10)
		mux := http.NewServeMux()
		client := setup(t, mux)
		client.baseURL.Host = "htt://localhost:1234"

		block, err := client.Network.GetBlockByNumber(context.Background(), blockID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to build request URL")
		require.Nil(t, block)
	})
	t.Run("Return_Error_When_Request_Fails", func(t *testing.T) {
		t.Parallel()

		blockID := int64(10)
		mux := http.NewServeMux()
		mux.HandleFunc(APIGetBlockByNumEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			if _, err := w.Write([]byte("bad request")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})
		client := setup(t, mux)
		block, err := client.Network.GetBlockByNumber(context.Background(), blockID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "HTTP request failed")
		require.Nil(t, block)
	})
	t.Run("Return_Error_When_Invalid_Response", func(t *testing.T) {
		t.Parallel()

		blockID := int64(10)
		mux := http.NewServeMux()
		mux.HandleFunc(APIGetBlockByNumEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Length", "50")
			w.WriteHeader(http.StatusOK)

			if _, err := w.Write([]byte("invalid response")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		client := setup(t, mux)
		block, err := client.Network.GetBlockByNumber(context.Background(), blockID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read response body")
		require.Nil(t, block)
	})
	t.Run("Return_Error_When_Decode_Response", func(t *testing.T) {
		t.Parallel()

		blockID := int64(10)
		mux := http.NewServeMux()
		mux.HandleFunc(APIGetBlockByNumEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			file, err := os.Open("testdata/invalid_block.json")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})
		client := setup(t, mux)
		block, err := client.Network.GetBlockByNumber(context.Background(), blockID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decode response")
		require.Nil(t, block)
	})
}
