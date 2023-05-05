package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
	if _, err := rw.buffer.Write(p); err != nil {
		return 0, err
	}
	defer rw.buffer.Reset()

	n = len(p)
	req, err := http.NewRequest("POST", rw.endpoint, rw.buffer)
	if err != nil {
		return 0, err
	}

	resp, err := rw.client.Do(req)
	if err != nil {
		return n, err
	}
	defer func(Body io.ReadCloser) {
		tErr := Body.Close()
		if tErr != nil {
			err = tErr
		}
	}(resp.Body)

	// Check the response status.
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to write to the RESTful endpoint %s; status code: %s", rw.endpoint, resp.Status)
	}

	return
}
