package logger_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/twistingmercury/observability/logger"
)

func TestNewHttpWriter(t *testing.T) {
	testEndpoint := "http://localhost:8080"
	writer := logger.NewHttpWriter(testEndpoint)
	assert.True(t, writer.IsReady())
}

func TestHttpWriter_Write(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(r.Body)

		assert.Equal(t, "Hello, World!", string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	writer := logger.NewHttpWriter(server.URL)
	assert.True(t, writer.IsReady())

	n, err := writer.Write([]byte("Hello, World!"))

	assert.NoError(t, err)
	assert.Equal(t, 13, n) // Length of the written data
}

func TestHttpWriter_Write_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	writer := logger.NewHttpWriter(server.URL)

	_, err := writer.Write([]byte("Hello, World!"))

	assert.Error(t, err)
}
