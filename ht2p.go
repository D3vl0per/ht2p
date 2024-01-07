package ht2p

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/D3vl0per/crypt/compression"
	"github.com/D3vl0per/crypt/generic"
	"github.com/klauspost/compress/gzhttp"
)

type NetHttp struct {
	URL                string
	URLParameters      map[string]string
	Method             string
	Body               []byte
	Headers            map[string]string
	ExpectedStatusCode int
	Client             http.Client
	Compressor         compression.Compressor
	UserAgent          string
	Ctx                context.Context
}

func (n *NetHttp) Request() (Response, error) {

	nn, err := defaultParameterSet(n)
	if err != nil {
		return Response{}, err
	}

	request, err := http.NewRequestWithContext(n.Ctx, nn.Method, nn.URL, bytes.NewReader(nn.Body))
	if err != nil {
		return Response{}, err
	}

	if len(nn.Headers) != 0 {
		for key, value := range nn.Headers {
			request.Header.Set(key, value)
		}
	}

	response, err := nn.Client.Do(request)
	if err != nil {
		return Response{}, err
	}

	responseStruct := Response{
		Headers:    response.Header,
		StatusCode: response.StatusCode,
	}

	if response.ContentLength > 0 {
		defer response.Body.Close()
	}

	if response.StatusCode != nn.ExpectedStatusCode {
		rawBody, err := io.ReadAll(response.Body)
		if err != nil {
			return Response{},
				errors.New(
					generic.StrCnct([]string{
						"failed to read response body [body reader]: ", err.Error(),
						" body error: ", err.Error()}...))
		} else {
			return Response{}, errors.New(generic.StrCnct([]string{"expected status code mismatch [http client]: ", response.Status, " body: ", string(rawBody)}...))
		}
	}

	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Response{}, err
	}

	if nn.Compressor == nil {
		responseStruct.Body = rawBody
		return responseStruct, nil
	}

	if nn.Compressor.GetName() == "gzip" {
		// Handled on transport level
		responseStruct.Body = rawBody
		return responseStruct, nil
	}

	if !strings.Contains(response.Header.Get("Content-Encoding"), nn.Compressor.GetName()) {
		return Response{}, errors.New(
			generic.StrCnct([]string{
				"requested decompressor mismatch by response content header: ", err.Error()}...))
	}

	responseStruct.Body, err = nn.Compressor.Decompress(rawBody)
	if err != nil {
		return Response{}, errors.New(generic.StrCnct(
			[]string{
				"failed to decompress response body [crypt compression]: ", err.Error()}...))
	}
	return responseStruct, nil
}

func (n *NetHttp) MultiRequest(urls []string) (Response, []error) {
	var errs []error
	for _, url := range urls {

		nn := n
		nn.URL = url

		response, err := nn.Request()
		if err != nil {
			errs = append(errs, err)
			continue
		} else {
			return response, errs
		}
	}
	return Response{}, errs
}

func defaultParameterSet(r *NetHttp) (*NetHttp, error) {
	rr := r
	parsedUrl, err := URIParser(r.URL, r.URLParameters)
	if err != nil {
		return &NetHttp{}, err
	}

	rr.URL = parsedUrl

	if r.Method == "" {
		rr.Method = http.MethodGet
	}

	if len(r.Headers) == 0 {
		rr.Headers = make(map[string]string)
	}

	if r.Compressor != nil {
		rr.Headers["Accept-Encoding"] = r.Compressor.GetName()
		switch r.Compressor.GetName() {
		case "gzip":
			rr.Compressor.SetLevel(compression.BestSpeed)
			if rr.Client.Transport != nil {
				transport, ok := rr.Client.Transport.(*http.Transport)
				if !ok {
					return &NetHttp{}, errors.New("failed to cast transport to http.Transport [http client]")
				}
				transport.DisableCompression = false
				rr.Client.Transport = gzhttp.Transport(transport, gzhttp.TransportEnableGzip(true))
			} else {
				rr.Client.Transport = gzhttp.Transport(&http.Transport{
					DisableCompression: false,
				})
			}
		case "br":
			rr.Compressor.SetLevel(compression.BrotliBestSpeed)
		}
	}

	if r.UserAgent != "" {
		rr.Headers["User-Agent"] = r.UserAgent
	}

	if r.ExpectedStatusCode == 0 {
		rr.ExpectedStatusCode = http.StatusOK
	}

	return rr, nil

}
