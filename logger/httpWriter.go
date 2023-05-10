package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HttpWriter is an interface that defines an io.Writer that writes to an HTTP endpoint.
type HttpWriter interface {
	io.Writer
	IsReady() bool
}

// httpWriter is an io.Writer implementation that writes to an HTTP endpoint.
type httpWriter struct {
	endpoint string
	client   *http.Client
	buffer   *bytes.Buffer
}

// IsReady returns true if the HTTP writer is ready to write.
func (rw *httpWriter) IsReady() bool {
	return rw.buffer != nil && rw.client != nil
}

// NewHttpWriter creates a new io.Writer that writes to an HTTP stream.
func NewHttpWriter(url string) HttpWriter {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = time.Second * 90

	return &httpWriter{
		endpoint: url,
		buffer:   new(bytes.Buffer),
		client: &http.Client{
			Timeout:   time.Second * 3,
			Transport: t,
		},
	}
}

// Write satisfies the io.Writer interface and writes data to the HTTP endpoint.
func (rw *httpWriter) Write(p []byte) (n int, err error) {
	n, err = rw.buffer.Write(p)
	defer rw.buffer.Reset()

	req, _ := http.NewRequest("POST", rw.endpoint, rw.buffer)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := rw.client.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to write to the RESTful endpoint %s; status code: %s", rw.endpoint, resp.Status)
	}

	return
}
