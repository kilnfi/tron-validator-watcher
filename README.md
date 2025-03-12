# 🚀 Tron Validator Watcher

**Tron Validator Watcher** is a powerful Prometheus Exporter designed to help you efficiently monitor your validators on the Tron blockchain.


## 🌟 Features

✅ Track and analyze validators' block production on the Tron blockchain with periodic polling.

## 📋 Prerequisites

To use **Tron Validator Watcher**, you need:

- A running **Tron RPC node**, either:
  - A local node, fully synchronized with the network.
  - A remote provider such as **QuickNode**, **TronGrid**, or other Tron RPC services.
- **Go 1.23+** installed _(for development purposes only)_.
- **Make**
- **Docker** if you want to build local images

## 📦 Installation

### Manual Installation (Development)

```sh
git clone https://github.com/yourusername/tron-validator-watcher.git
cd tron-validator-watcher
make build
```

## Releases
With each release, we provide precompiled binaries for various platforms, making it easy to install and use Tron Validator Watcher without needing to build from source.

You can find the latest releases here: [Tron Validator Watcher Releases](https://github.com/kilnfi/tron-validator-watcher/release).

## 🚀 Usage

```bash
Usage:
  tron-validator-watcher [flags]

Flags:
      --block-watcher-enabled                enable block watcher (default true)
      --block-watcher-refresh-interval int   block watcher refresh interval in seconds (default 30)
      --config-file string                   config file (default is config.yml)
  -h, --help                                 help for tron-validator-watcher
      --http-server-host string              HTTP server host
      --http-server-port int                 HTTP server port (default 8080)
      --log-level string                     log level (debug, info, warn, error, fatal, panic) (default "info")
      --rpc-endpoint string                  block watcher API URL
```

Run the tool with:

```sh
./tron-validator-watcher --config-file=config.yaml
```

### Available Flags
| Flag                                  | Type   | Description                                          | Default      |
| ------------------------------------- | ------ | ---------------------------------------------------- | ------------ |
| `--config-file`                       | string | Path to the configuration file                       | `config.yml` |
| `--block-watcher-enabled`             | bool   | Enable block watcher                                 | `true`       |
| `--block-watcher-refresh-interval`    | int    | Block watcher refresh interval in seconds            | `30`         |
| `--http-server-host`                  | string | HTTP server host                                     |              |
| `--http-server-port`                  | int    | HTTP server port                                     | `8080`       |
| `--log-level`                         | string | Log level (`debug`, `info`, `warn`, `error`)         | `info`       |
| `--rpc-endpoint`                      | string | Block watcher API URL                                |              |

## ⚙️ Configuration File

Modify the `config.example.yaml` file and rename it to `config.yaml`. Below is a breakdown of the configuration options:

| Option                           | Description                                                | Example                      |
| -------------------------------- | ---------------------------------------------------------- | ---------------------------- |
| `validators`                     | List of validators to monitor                              | See below                    |
| `validators.address`             | The blockchain address of the validator                    | `TABC123...XYZ`              |
| `validators.name`                | A human-readable name for the validator                    | `"Validator Name"`           |
| `validators.instance`            | The instance identifier for tracking                       | `"tron-validator-mainnet-0"` |
| `log-level`                      | Logging verbosity level (`debug`, `info`, `warn`, `error`) | `"debug"`                    |
| `http-server.host`               | The host IP address for the HTTP server                    | `"0.0.0.0"`                  |
| `http-server.port`               | The port for the HTTP server                               | `8080`                       |
| `block-watcher.enabled`          | Enables or disables block monitoring                       | `true`                       |
| `block-watcher.refresh-interval` | Interval (in seconds) for polling validator activity       | `10`                         |
| `rpc.endpoint`                   | The RPC node URL for blockchain data retrieval             | `"https://localhost:8090"`   |

### Example `config.yaml`

```yaml
validators:
  - address: "your_validator_address"
    name: "Validator Name"
    instance: "tron-validator-mainnet-0"
log-level: "debug"
http-server:
  host: "0.0.0.0"
  port: 8080
block-watcher:
  enabled: true
  refresh-interval: 10
rpc:
  endpoint: "https://localhost:8090"
```

## 🌡️ Metrics

This exporter provides the following set of metrics, which can be used to build a dashboard or create alerts.

| Metric Name                                                       | Description          | Type | Labels |
| ----------------------------------------------------------------- | -------------------- |---- | --- |
| `tron_validator_watcher_block_producer_info`                      | Block producer info                                         | GaugeVec    | `validator_name`, `validator_address`, `rank` |
| `tron_validator_watcher_consecutive_missed_blocks_total`          | Total number of consecutive blocks missed by the validator  | GaugeVec    | `validator_name`, `validator_address`|
| `tron_validator_watcher_latest_block_processed_by_block_watcher`  | The latest block processed by the block watcher             | Gauge       | - |
| `tron_validator_watcher_missed_blocks_total`                      | Total number of blocks missed by the validator              | CounterVec  | `validator_name`, `validator_address`|
| `tron_validator_watcher_proposed_blocks_total`                    | Total number of blocks proposed by the validator            | CounterVec  | `validator_name`, `validator_address`|
| `tron_validator_watcher_round_progress`                           | The current progress of the round                           | Gauge       | - |


## 🛠 Development

### Running Locally

```sh
git clone https://github.com/yourusername/tron-validator-watcher.git
cd tron-validator-watcher
make run
```

### Running Tests

```sh
make tests
make coverage
```

## 📜 License

This project is licensed under the [MIT License](LICENSE).

## 🎉 Contributing

Pull requests are welcome! Feel free to open an issue or suggest improvements.
