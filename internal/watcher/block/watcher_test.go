package blockwatcher

import (
	"context"
	"fmt"
	"testing"
	"time"

	clog "github.com/kilnfi/tron-validator-watcher/internal/logger"
	"github.com/kilnfi/tron-validator-watcher/internal/tron"
	"github.com/kilnfi/tron-validator-watcher/internal/tron/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func getWitnessID(t *testing.T, timeSlot int64) int64 {
	t.Helper()

	const n = 2
	var witnessID int64
	slot := (timeSlot - tron.GenesisBlockTime) / (tron.BlockTime * 1000)

	if (timeSlot/1000)%tron.RoundDuration == 0 {
		witnessID = ((slot - tron.MaintenanceSkipSlots) % (n * tron.SingleRepeat)) / tron.SingleRepeat
	} else {
		witnessID = (slot % (n * tron.SingleRepeat)) / tron.SingleRepeat
	}

	return witnessID + 1
}

func TestStart(t *testing.T) {
	t.Parallel()

	t.Run("Returns_No_Error_When_Watcher_Is_Running_Normally", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      time.Now().UnixMilli() + 3000,
					WitnessAddress: "4178c842ee63b253f8f0d2955bbc582c661a078c9d20",
				},
			},
		}

		proposerAddress, err := tron.ConvertAddressToBase58(startBlock.BlockHeader.RawData.WitnessAddress)
		if err != nil {
			t.Fatalf("Failed to convert address to base58: %v", err)
		}

		proposerAccount := &tron.Account{
			Address:     proposerAddress,
			AccountName: "TronWatcher",
			WitnessInfo: &tron.Witness{
				Rank:    1,
				Address: startBlock.BlockHeader.RawData.WitnessAddress,
				IsJobs:  true,
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      time.Now().UnixMilli(),
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					Rank:    12,
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil)
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil)
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil)

		mockAccount.EXPECT().
			GetAccount(proposerAddress).
			Return(proposerAccount, nil)
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))

		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)

		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)

		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "context canceled", "Error message mismatch")
	})

	t.Run("Returns_No_Error_When_We_Propose_A_Block", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      time.Now().UnixMilli(),
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      time.Now().UnixMilli() + 3000,
					WitnessAddress: "4178c842ee63b253f8f0d2955bbc582c661a078c9d20",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					Rank:    int(getWitnessID(t, startBlock.BlockHeader.RawData.Timestamp)),
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil)
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil)
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil)

		mockAccount.EXPECT().
			GetAccount(validators[0].Address).
			Return(&validators[0], nil)
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))

		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)

		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)

		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "context canceled", "Error message mismatch")
	})

	t.Run("Returns_No_Error_When_We_Miss_A_Block", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      time.Now().UnixMilli() + 3000,
					WitnessAddress: "4178c842ee63b253f8f0d2955bbc582c661a078c9d20",
				},
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      time.Now().UnixMilli(),
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		proposerAddress, err := tron.ConvertAddressToBase58(startBlock.BlockHeader.RawData.WitnessAddress)
		if err != nil {
			t.Fatalf("Failed to convert address to base58: %v", err)
		}

		proposerAccount := &tron.Account{
			Address:     proposerAddress,
			AccountName: "TronWatcher",
			WitnessInfo: &tron.Witness{
				Rank:    int(getWitnessID(t, startBlock.BlockHeader.RawData.Timestamp)),
				Address: startBlock.BlockHeader.RawData.WitnessAddress,
				IsJobs:  true,
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					// we always assign the corresponding rank for the slot to make
					// this validator the slot leader
					Rank:    int(getWitnessID(t, startBlock.BlockHeader.RawData.Timestamp)),
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil)
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil)
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil)

		mockAccount.EXPECT().
			GetAccount(proposerAccount.Address).
			Return(proposerAccount, nil)
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))

		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)

		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)

		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "context canceled", "Error message mismatch")
	})

	t.Run("Returns_Error_When_Fetching_Latest_Block", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      time.Now().UnixMilli() + 3000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					// we always assign the corresponding rank for the slot to make
					// this validator the slot leader
					Rank:    int(getWitnessID(t, startBlock.BlockHeader.RawData.Timestamp)),
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil)
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(nil, fmt.Errorf("timeout"))
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))
		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)

		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "failed to fetch latest block")
	})

	t.Run("Returns_Error_When_Fetching_Current_Block", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      time.Now().UnixMilli(),
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      time.Now().UnixMilli() + 3000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					// we always assign the corresponding rank for the slot to make
					// this validator the slot leader
					Rank:    int(getWitnessID(t, startBlock.BlockHeader.RawData.Timestamp)),
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil)
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(nil, fmt.Errorf("timeout")).Once()
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))
		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)

		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "failed to retrieve block")
	})

	t.Run("Returns_Error_When_New_Round_Detected", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      1741262400000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      1741262400000 + 3000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					// we always assign the corresponding rank for the slot to make
					// this validator the slot leader
					Rank:    int(getWitnessID(t, startBlock.BlockHeader.RawData.Timestamp)),
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(15*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()

		// need to do this hack because the start block is too old. We need to get the latest block but return the start block
		// to simulate a new round
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(startBlock, nil).Once()

		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil).Once()
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()

		// When we try to refresh validator data on a new round.
		mockAccount.EXPECT().
			GetAccount(validators[0].Address).
			Return(nil, fmt.Errorf("timeout")).
			Once()
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))
		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}
		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)

		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "failed to handle round change")
	})

	t.Run("Returns_No_Error_When_New_Round_Detected", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      1741262400000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      1741262400000 + 3000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					// we always assign the corresponding rank for the slot to make
					// this validator the slot leader
					Rank:    10,
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()

		// need to do this hack because the start block is too old. We need to get the latest block but return the start block
		// to simulate a new round
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(startBlock, nil).Once()

		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil).Once()
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()

		mockAccount.EXPECT().
			GetAccount(validators[0].Address).
			Return(&validators[0], nil).
			Once()

		nextBlock := latestBlock
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, nextBlock.BlockHeader.RawData.Number).Return(nextBlock, nil).Once()

		mockAccount.EXPECT().
			GetAccount(validators[0].Address).
			Return(&validators[0], nil).
			Once()
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil).Times(2)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))
		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)

		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "context canceled")
	})

	t.Run("Returns_Error_When_Convert_Proposer_Address", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      time.Now().UnixMilli(),
					WitnessAddress: "badaddress",
				},
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      time.Now().UnixMilli() + 3000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					// we always assign the corresponding rank for the slot to make
					// this validator the slot leader
					Rank:    10,
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil).Once()
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))
		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)
		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "failed to convert address to base58")
	})
	t.Run("Returns_Error_When_Fetching_Proposer_Account", func(t *testing.T) {
		t.Parallel()

		startBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         100,
					Timestamp:      time.Now().UnixMilli(),
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		latestBlock := &tron.Block{
			BlockHeader: tron.BlockHeader{
				RawData: tron.BlockHeaderRawData{
					Number:         101,
					Timestamp:      time.Now().UnixMilli() + 3000,
					WitnessAddress: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
				},
			},
		}

		validators := tron.AccountList{
			{
				Address:     "TFKkZiRTYxMud3BrJGQxZxWufwcRrMHKko",
				AccountName: "test",
				WitnessInfo: &tron.Witness{
					// we always assign the corresponding rank for the slot to make
					// this validator the slot leader
					Rank:    10,
					Address: "413abb4bafa37c6b9086ec458d560438c2e0fab66c",
					IsJobs:  true,
				},
			},
		}

		// context to automatically close the watcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			time.AfterFunc(12*time.Second, cancel)
		}()

		// mocks clients
		mockNetwork := mocks.NewNetworkClient(t)
		mockAccount := mocks.NewAccountClient(t)

		// mocks expectations
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()
		mockNetwork.EXPECT().GetLatestBlock(mock.Anything).Return(latestBlock, nil).Once()
		mockNetwork.EXPECT().GetBlockByNumber(mock.Anything, startBlock.BlockHeader.RawData.Number).Return(startBlock, nil).Once()

		mockAccount.EXPECT().
			GetAccount(validators[0].Address).
			Return(nil, fmt.Errorf("timeout")).
			Once()
		mockAccount.EXPECT().ListWitnesses().Return(&tron.Witnesses{Witnesses: []tron.Witness{{IsJobs: true}, {IsJobs: true}}}, nil)

		client, err := tron.NewClient(tron.WithBaseURL("http://localhost:8090"))
		if err != nil {
			t.Fatalf("Failed to create Tron client: %v", err)
		}

		client.Network = mockNetwork
		client.Account = mockAccount

		logger := logrus.New()
		logger.SetFormatter(&clog.CustomTextFormatter{})
		registry := prometheus.NewRegistry()
		metrics := NewCollection()
		metrics.MustRegister(registry)

		refreshInterval := 10

		watcher, err := NewBlockWatcher(
			WithLogger(logger),
			WithTronClient(client),
			WithStartBlock(startBlock),
			WithValidators(validators),
			WithRefreshInterval(refreshInterval),
			WithMetrics(metrics),
		)
		if err != nil {
			t.Fatalf("Failed to create BlockWatcher: %v", err)
		}

		err = watcher.Start(ctx)
		require.Error(t, err, "Expected an error but got nil")
		require.ErrorContains(t, err, "failed to get proposer account")
	})
}
