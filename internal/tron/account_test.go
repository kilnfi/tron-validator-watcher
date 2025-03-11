package tron

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const invalidHost = "htt://localhost:1234"

func TestAccountClient_GetAccount(t *testing.T) {
	t.Parallel()

	t.Run("WithValidParameters_Should_Return_No_Error", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc("/wallet/getaccount", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// open testdata file
			file, err := os.Open("testdata/single_account.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			// copy file content to response writer
			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})

		mux.HandleFunc(APIListWitnessesEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// open testdata file
			file, err := os.Open("testdata/list_witnesses.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			// copy file content to response writer
			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})

		client := setup(t, mux)
		account, err := client.Account.GetAccount("TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ")
		if err != nil {
			t.Fatal(err)
		}

		require.NoError(t, err)
		require.Equal(t, "TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ", account.Address)
		require.Equal(t, "Kiln_Staking", account.AccountName)
	})

	t.Run("WithValidParameters_Should_Return_ErrorJoiningPath", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		client := setup(t, mux)
		client.baseURL.Host = invalidHost
		account, err := client.Account.GetAccount("TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ")

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to build request URL")
		require.Nil(t, account)
	})

	t.Run("WithValidParameters_Should_Return_FailToGetAccount", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc("/wallet/getaccount", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			if _, err := w.Write([]byte("bad request")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		client := setup(t, mux)
		account, err := client.Account.GetAccount("TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ")

		require.Error(t, err)
		require.Contains(t, err.Error(), "request failed")
		require.Nil(t, account)
	})

	t.Run("WithValidParameters_Should_Return_FailToDecodeAccountResponse", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc("/wallet/getaccount", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// open testdata file
			file, err := os.Open("testdata/single_invalid_account.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			// copy file content to response writer
			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})

		client := setup(t, mux)
		account, err := client.Account.GetAccount("TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ")

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decode response ")
		require.Nil(t, account)
	})

	t.Run("WithValidParameters_Should_Return_FailGetWitnessInfo", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc("/wallet/getaccount", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// open testdata file
			file, err := os.Open("testdata/single_account.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			// copy file content to response writer
			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})

		mux.HandleFunc(APIListWitnessesEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			if _, err := w.Write([]byte("not found")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		client := setup(t, mux)
		account, err := client.Account.GetAccount("TQzd66b9EFVHJfZK5AmiVhBjtJvXGeSPPZ")

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to retrieve witness info")
		require.Nil(t, account)
	})
}

func TestListWitnesses(t *testing.T) {
	t.Parallel()
	t.Run("Return_No_Error_When_Valid_Response", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIListWitnessesEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			file, err := os.Open("testdata/list_witnesses.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})
		client := setup(t, mux)
		witnesses, err := client.Account.ListWitnesses()
		require.NoError(t, err)
		require.Len(t, witnesses.Witnesses, 7)
	})

	t.Run("Return_Error_When_Building_Request_URL", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()

		client := setup(t, mux)
		client.baseURL.Host = invalidHost
		witnesses, err := client.Account.ListWitnesses()

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to build request URL")
		require.Nil(t, witnesses)
	})

	t.Run("Return_Error_When_Http_Request_Fails", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIListWitnessesEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			if _, err := w.Write([]byte("bad request")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		client := setup(t, mux)
		witnesses, err := client.Account.ListWitnesses()
		require.Error(t, err)
		require.Contains(t, err.Error(), "HTTP request failed")
		require.Nil(t, witnesses)
	})
	t.Run("Return_Error_When_Decoding_Response", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIListWitnessesEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			file, err := os.Open("testdata/list_invalid_witnesses.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})
		client := setup(t, mux)
		witnesses, err := client.Account.ListWitnesses()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to decode response")
		require.Nil(t, witnesses)
	})
}

func TestGetWitnesses(t *testing.T) {
	t.Parallel()

	t.Run("Return_No_Error_When_Valid_Response", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIListWitnessesEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			file, err := os.Open("testdata/list_witnesses.json")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()

			_, err = io.Copy(w, file)
			if err != nil {
				t.Fatal(err)
			}
		})
		client := setup(t, mux)
		witness, err := client.Account.GetWitnesses("418440ffd578f7a5abf3537b5f46a6980d382db581")
		require.NoError(t, err)
		require.Equal(t, "418440ffd578f7a5abf3537b5f46a6980d382db581", witness.Address)
	})
	t.Run("Return_Error_When_List_Witnesses", func(t *testing.T) {
		t.Parallel()

		mux := http.NewServeMux()
		mux.HandleFunc(APIListWitnessesEndpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)

			if _, err := w.Write([]byte("not found")); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})

		client := setup(t, mux)
		witness, err := client.Account.GetWitnesses("418440ffd578f7a5abf3537b5f46a6980d382db581")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to retrieve witnesses")
		require.Nil(t, witness)
	})
}
