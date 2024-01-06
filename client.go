package ht2p

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/D3vl0per/crypt/compression"
	"github.com/D3vl0per/crypt/generic"
	"github.com/klauspost/compress/gzhttp"
	"github.com/valyala/fasthttp"
)

type Response struct {
	Body       []byte
	Headers    http.Header
	StatusCode int
}

type HttpClient interface {
	Request() (Response, error)
	MultiRequest(urls []string) (Response, []error)
}

type Compression struct {
	Decompression bool
	AcceptHeader  bool
	Header        string //TODO: implement GetName in crypt library
	Compressor    compression.Compressor
}

type NetHttp struct {
	URL                string
	URLParameters      map[string]string
	Method             string
	Body               []byte
	Headers            map[string]string
	ExpectedStatusCode int
	Client             http.Client
	Compression        Compression
	UserAgent          string
}

func (n *NetHttp) Request() (Response, error) {

	nn, err := defaultParameterSet(n)
	if err != nil {
		return Response{}, err
	}

	request, err := http.NewRequest(nn.Method, nn.URL, bytes.NewReader(nn.Body))
	if err != nil {
		return Response{}, errors.New(generic.StrCnct([]string{"failed to create request [http client]: ", err.Error()}...))
	}

	if len(nn.Headers) != 0 {
		for key, value := range nn.Headers {
			request.Header.Set(key, value)
		}
	}

	if n.Compression.Decompression && n.Compression.Compressor.GetName() == "gzip" {
		n.Client.Transport = gzhttp.Transport(n.Client.Transport.(*http.Transport))
	}

	response, err := nn.Client.Do(request)
	if err != nil {
		return Response{}, errors.New(generic.StrCnct([]string{"failed to send request [http client]: ", err.Error()}...))
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
			return Response{}, errors.New(generic.StrCnct([]string{"failed to read response body [body reader]: ", err.Error(), " body error: ", err.Error()}...))
		} else {
			return Response{}, errors.New(generic.StrCnct([]string{"expected status code mismatch [http client]: ", response.Status, " body: ", string(rawBody)}...))
		}
	}

	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Response{}, errors.New(generic.StrCnct([]string{"failed to read response body [body reader]: ", err.Error()}...))
	}

	if nn.Compression.Decompression {
		if strings.Contains(response.Header.Get("Content-Encoding"), "gzip") || !nn.Compression.AcceptHeader {
			responseStruct.Body, err = nn.Compression.Compressor.Decompress(rawBody)
			if err != nil {
				return Response{}, errors.New(generic.StrCnct([]string{"failed to decompress response body [crypt compression]: ", err.Error()}...))
			}
			return responseStruct, nil
		}
	}

	responseStruct.Body = rawBody
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

type FastHttp struct {
	URL                string
	URLParameters      map[string]string
	Method             string
	Body               []byte
	Headers            map[string]string
	ExpectedStatusCode int
	Client             fasthttp.Client
	Compression        Compression
	UserAgent          string
}

func (f *FastHttp) Request() (Response, error) {

	ff, err := defaultFastParameterSet(f)
	if err != nil {
		return Response{}, err
	}

	request := fasthttp.AcquireRequest()
	request.SetRequestURI(ff.URL)

	if len(ff.Headers) != 0 {
		for key, value := range ff.Headers {
			request.Header.Set(key, value)
		}
	}

	request.Header.SetMethod(ff.Method)

	if f.Body != nil {
		request.SetBodyRaw(f.Body)
	}

	response := fasthttp.AcquireResponse()

	err = f.Client.Do(request, response)

	fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)
	if err != nil {
		return Response{}, errors.New(generic.StrCnct([]string{"failed to send request [fasthttp client]: ", err.Error()}...))
	}

	responseStruct := Response{
		StatusCode: response.StatusCode(),
	}

	respBody := response.Body()

	if responseStruct.StatusCode != f.ExpectedStatusCode {
		return Response{}, errors.New(generic.StrCnct([]string{"expected status code mismatch [fasthttp client]: ", string(response.Header.StatusMessage()), " body: ", string(respBody)}...))
	}

	if f.Compression.Decompression {
		if bytes.Contains(response.Header.Peek("Content-Encoding"), []byte("gzip")) || !ff.Compression.AcceptHeader {
			responseStruct.Body, err = ff.Compression.Compressor.Decompress(respBody)
			if err != nil {
				return Response{}, errors.New(generic.StrCnct([]string{"failed to decompress response body [crypt compression]: ", err.Error()}...))
			}
			return responseStruct, nil
		}
	}
	responseStruct.Body = respBody
	return responseStruct, nil
}

func (f *FastHttp) MultiRequest(urls []string) (Response, []error) {
	var errs []error
	for _, url := range urls {

		ff := f
		ff.URL = url

		response, err := ff.Request()
		if err != nil {
			errs = append(errs, err)
			continue
		} else {
			return response, errs
		}
	}
	return Response{}, errs
}

func URIParser(baseUrl string, parameters map[string]string) (string, error) {
	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return "", errors.New(generic.StrCnct([]string{"failed to parse url [url parser]: ", err.Error()}...))
	}

	if len(parameters) != 0 {
		query := parsedUrl.Query()
		for key, value := range parameters {
			query.Set(key, value)
		}
		parsedUrl.RawQuery = query.Encode()
	}
	return parsedUrl.String(), nil
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

	if r.Compression.Decompression {
		transport := rr.Client.Transport.(*http.Transport)
		transport.DisableCompression = false
		rr.Client.Transport = transport
	}

	if r.Compression.AcceptHeader {
		rr.Headers["Accept-Encoding"] = r.Compression.Compressor.GetName()
	}

	if r.UserAgent != "" {
		rr.Headers["User-Agent"] = r.UserAgent
	}

	if r.ExpectedStatusCode == 0 {
		rr.ExpectedStatusCode = http.StatusOK
	}

	return rr, nil

}

func defaultFastParameterSet(f *FastHttp) (*FastHttp, error) {
	ff := f
	parsedUrl, err := URIParser(ff.URL, ff.URLParameters)
	if err != nil {
		return &FastHttp{}, err
	}

	ff.URL = parsedUrl

	if f.Method == "" {
		ff.Method = http.MethodGet
	}

	if len(f.Headers) == 0 {
		ff.Headers = make(map[string]string)
	}

	if f.Compression.AcceptHeader {
		ff.Headers["Accept-Encoding"] = ff.Compression.Compressor.GetName()
	}

	if f.UserAgent != "" {
		ff.Headers["User-Agent"] = f.UserAgent
	}

	if f.ExpectedStatusCode == 0 {
		ff.ExpectedStatusCode = fasthttp.StatusOK
	}

	return ff, nil

}
