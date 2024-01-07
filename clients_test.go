package ht2p_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/D3vl0per/ht2p"
	r "github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func BenchmarkRequest(b *testing.B) {
	url := "https://1.1.1.1"

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
		Ctx:    context.Background(),
	}

	b.Run("ht2p abstraction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			response, err := ht2pc.Request()
			r.NoError(b, err)
			r.Equal(b, 200, response.StatusCode)
		}
	})

	readTimeout, err := time.ParseDuration("5000ms")
	r.NoError(b, err)
	writeTimeout, err := time.ParseDuration("5000ms")
	r.NoError(b, err)
	maxIdleConnDuration, err := time.ParseDuration("1h")
	r.NoError(b, err)

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
