package ht2p_test

import (
	"net/http"
	"testing"

	"github.com/D3vl0per/ht2p"
	r "github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestFastRequest(t *testing.T) {

	tests := []struct {
		name         string
		request      ht2p.HttpClient
		expectedBody []byte
	}{
		{
			name: "Simple GET request",
			request: &ht2p.FastHttp{
				URL:       "https://httpbin.org/json",
				UserAgent: "ht2p/0.0.1",
			},
		},
		{
			name: "Specified GET request",
			request: &ht2p.FastHttp{
				URL:       "https://httpbin.org/json",
				Method:    http.MethodGet,
				UserAgent: "ht2p/0.0.1",
			},
		},
		{
			name: "Simple GET request with custom header",
			request: &ht2p.FastHttp{
				URL: "https://httpbin.org/json",
				Headers: map[string]string{
					"Test": "Test",
				},
				UserAgent: "ht2p/0.0.1",
			},
		},
		{
			name: "Specified GET request with custom header and user agent (User-agent overwrite)",
			request: &ht2p.FastHttp{
				URL:    "https://httpbin.org/json",
				Method: http.MethodGet,
				Headers: map[string]string{
					"User-Agent": "Test",
				},
				UserAgent: "ht2p/0.0.1",
			},
		},
		{
			name: "Simple GET request with Brotil compression",
			request: &ht2p.FastHttp{
				URL:        "https://httpbin.org/brotli",
				Compressor: ht2p.Brotil,
				UserAgent:  "ht2p/0.0.1",
			},
		},
		{
			name: "Simple GET request with Gzip compression",
			request: &ht2p.FastHttp{
				URL:        "https://httpbin.org/gzip",
				Compressor: ht2p.Gzip,
				UserAgent:  "ht2p/0.0.1",
			},
		},
		{
			name: "Simple GET request with DEFLATE compression",
			request: &ht2p.FastHttp{
				URL:        "https://httpbin.org/gzip",
				Compressor: ht2p.Deflate,
				UserAgent:  "ht2p/0.0.1",
			},
		},
		{
			name: "Query parameters GET request",
			request: &ht2p.FastHttp{
				URL: "https://httpbin.org/cookies/set",
				URLParameters: map[string]string{
					"freeform": "test",
				},
				Headers: map[string]string{
					"Accept": "text/plain",
				},
				MaxRedirects: 1,
				UserAgent:    "ht2p/0.0.1",
				Client:       fasthttp.Client{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			response, err := test.request.Request()
			r.NoError(t, err)
			r.Equal(t, 200, response.StatusCode)

			t.Log("Body:", string(response.Body))
			t.Log("Status code:", response.StatusCode)

			if test.expectedBody != nil {
				r.Equal(t, test.expectedBody, response.Body)
			}
		})
	}
}
