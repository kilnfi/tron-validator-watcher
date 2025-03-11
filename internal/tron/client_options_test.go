package tron

import (
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestWithBaseURL(t *testing.T) {
	t.Parallel()

	type inputFields struct {
		baseURL string
	}

	tests := []struct {
		name         string
		inputs       inputFields
		wantErr      bool
		errorMessage string
	}{
		{
			name: "Return_No_Error",
			inputs: inputFields{
				baseURL: "https://localhost:8080",
			},
			wantErr: false,
		},
		{
			name: "Return_Error_When_BaseURL_Is_Not_Valid",
			inputs: inputFields{
				baseURL: "https://localhost\x7F:8080",
			},
			wantErr:      true,
			errorMessage: "failed to parse base URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Client{}
			err := WithBaseURL(tt.inputs.baseURL)(c)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errorMessage)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	type inputFields struct {
		client *http.Client
	}

	tests := []struct {
		name         string
		inputs       inputFields
		wantErr      bool
		errorMessage string
	}{
		{
			name: "Return_No_Error",
			inputs: inputFields{
				client: http.DefaultClient,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Client{}
			err := WithHTTPClient(tt.inputs.client)(c)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errorMessage)
				return
			}
			require.NoError(t, err)
		})
	}
}
func TestWithLogger(t *testing.T) {
	t.Parallel()

	type inputFields struct {
		logger *logrus.Logger
	}

	tests := []struct {
		name         string
		inputs       inputFields
		wantErr      bool
		errorMessage string
	}{
		{
			name: "Return_No_Error",
			inputs: inputFields{
				logger: logrus.New(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Client{}
			err := WithLogger(tt.inputs.logger)(c)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errorMessage)
				return
			}
			require.NoError(t, err)
		})
	}
}
