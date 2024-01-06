package ht2p_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/D3vl0per/ht2p"
	r "github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestRequest(t *testing.T) {

	tests := []struct {
		name         string
		request      ht2p.HttpClient
		expectedBody []byte
	}{
		{
			name: "Simple GET request",
			request: &ht2p.NetHttp{
				URL: "https://httpbin.org/get",
			},
		},
		{
			name: "Specified GET request",
			request: &ht2p.NetHttp{
				URL:    "https://httpbin.org/get",
				Method: http.MethodGet,
			},
		},
		{
			name: "Simple GET request with custom header",
			request: &ht2p.NetHttp{
				URL: "https://httpbin.org/get",
				Headers: map[string]string{
					"Test": "Test",
				},
			},
		},
		{
			name: "Specified GET request with custom header and user agent (User-agent overwrite)",
			request: &ht2p.NetHttp{
				URL:    "https://httpbin.org/get",
				Method: http.MethodGet,
				Headers: map[string]string{
					"User-Agent": "Test",
				},
				UserAgent: "NotTest",
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

func BenchmarkRequest(b *testing.B) {
	url := "https://malware-filter.gitlab.io/malware-filter/phishing-filter.txt"

	client := http.Client{}

	b.Run("prue go http client", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			response, err := client.Get(url)
			r.NoError(b, err)
			r.Equal(b, 200, response.StatusCode)
		}
	})

	ht2pc := &ht2p.NetHttp{
		URL:    url,
		Client: http.Client{},
	}

	b.Run("ht2p abstraction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			response, err := ht2pc.Request()
			r.NoError(b, err)
			r.Equal(b, 200, response.StatusCode)
		}
	})

	readTimeout, _ := time.ParseDuration("5000ms")
	writeTimeout, _ := time.ParseDuration("5000ms")
	maxIdleConnDuration, _ := time.ParseDuration("1h")

	ft2pc := ht2p.FastHttp{
		URL: url,
		Client: fasthttp.Client{
			ReadTimeout:                   readTimeout,
			WriteTimeout:                  writeTimeout,
			MaxIdleConnDuration:           maxIdleConnDuration,
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			// increase DNS cache time to an hour instead of default minute
			Dial: (&fasthttp.TCPDialer{
				Concurrency:      4096,
				DNSCacheDuration: time.Hour,
			}).Dial,
		},
	}

	b.Run("ft2p abstraction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			response, err := ft2pc.Request()
			r.NoError(b, err)
			r.Equal(b, 200, response.StatusCode)
		}
	})
}
