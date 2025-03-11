package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	type inputFields struct {
		config *Config
	}
	tests := []struct {
		name         string
		inputs       inputFields
		wantErr      bool
		errorMessage string
	}{
		{
			name: "Return_Error_When_No_Validators",
			inputs: inputFields{
				config: &Config{},
			},
			wantErr:      true,
			errorMessage: "no validators provided",
		},
		{
			name: "Return_Error_When_Validator_Address_Is_Empty",
			inputs: inputFields{
				config: &Config{
					Validators: []Validator{
						{
							Address: "",
						},
					},
				},
			},
			wantErr:      true,
			errorMessage: "validator address is required",
		},
		{
			name: "Return_Error_When_Validator_Name_Is_Empty",
			inputs: inputFields{
				config: &Config{
					Validators: []Validator{
						{
							Address: "address",
							Name:    "",
						},
					},
				},
			},
			wantErr:      true,
			errorMessage: "validator name is required",
		},
		{
			name: "Return_Error_When_Validator_Instance_Is_Empty",
			inputs: inputFields{
				config: &Config{
					Validators: []Validator{
						{
							Address:  "address",
							Name:     "name",
							Instance: "",
						},
					},
				},
			},
			wantErr:      true,
			errorMessage: "validator instance name is required",
		},
		{
			name: "Return_Error_When_Block_Watcher_Refresh_Interval_Is_Less_Than_Zero",
			inputs: inputFields{
				config: &Config{
					Validators: []Validator{
						{
							Address:  "address",
							Name:     "name",
							Instance: "instance",
						},
					},
					BlockWatcher: BlockWatcherConfig{
						RefreshInterval: -1,
					},
				},
			},
			wantErr:      true,
			errorMessage: "block watcher refresh interval must be greater than 0",
		},
		{
			name: "Return_Error_When_RPC_Endpoint_Is_Empty",
			inputs: inputFields{
				config: &Config{
					Validators: []Validator{
						{
							Address:  "address",
							Name:     "name",
							Instance: "instance",
						},
					},
					BlockWatcher: BlockWatcherConfig{
						RefreshInterval: 1,
					},
					RPC: RPCConfig{
						Endpoint: "",
					},
				},
			},
			wantErr:      true,
			errorMessage: "rpc endpoint is required",
		},
		{
			name: "Return_No_Error",
			inputs: inputFields{
				config: &Config{
					Validators: []Validator{
						{
							Address:  "address",
							Name:     "name",
							Instance: "instance",
						},
					},
					BlockWatcher: BlockWatcherConfig{
						RefreshInterval: 1,
					},
					RPC: RPCConfig{
						Endpoint: "endpoint",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inputs.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMessage)
				return
			}
			require.NoError(t, err)
		})
	}
}
