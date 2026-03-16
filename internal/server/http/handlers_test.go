package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultHandler(t *testing.T) {
	t.Parallel()

	t.Run("Returns_No_Error_When_Endpoint_Is_Healthy", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		server, err := New(
			nil,
		)

		require.NoError(t, err)
		server.router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusMovedPermanently, w.Code)
	})

	t.Run("Returns_Error_When_Endpoint_Not_Exists", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fake", nil)
		w := httptest.NewRecorder()

		registry := prometheus.NewRegistry()
		server, err := New(
			registry,
		)

		require.NoError(t, err)
		server.router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestLiveProbe(t *testing.T) {
	t.Parallel()

	t.Run("Returns_No_Errors_When_Endpoint_Is_Healthy", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/livez", nil)
		w := httptest.NewRecorder()

		registry := prometheus.NewRegistry()
		server, err := New(
			registry,
		)
		require.NoError(t, err)
		server.router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestReadyProbe(t *testing.T) {
	t.Parallel()

	t.Run("Returns_No_Errors_When_Endpoint_Is_Healthy", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/readyz", nil)
		w := httptest.NewRecorder()

		registry := prometheus.NewRegistry()
		server, err := New(
			registry,
		)
		require.NoError(t, err)
		server.router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestMetricsHandler(t *testing.T) {
	t.Parallel()

	t.Run("Returns_No_Errors_When_Endpoint_Is_Healthy", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		registry := prometheus.NewRegistry()
		server, err := New(
			registry,
		)

		require.NoError(t, err)
		server.router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
