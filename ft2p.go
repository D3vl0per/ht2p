package ht2p

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/D3vl0per/crypt/compression"
	"github.com/D3vl0per/crypt/generic"
	"github.com/valyala/fasthttp"
)

type FastHttp struct {
	URL                string
	URLParameters      map[string]string
	Method             string
	Body               []byte
	Headers            map[string]string
	ExpectedStatusCode int
	Client             fasthttp.Client
	Compressor         compression.Compressor
	UserAgent          string
	MaxRedirects       int
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

	if f.MaxRedirects != 0 {
		err = fasthttp.DoRedirects(request, response, f.MaxRedirects)
	} else {
		err = f.Client.Do(request, response)
	}
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
		return Response{},
			errors.New(
				generic.StrCnct([]string{
					"expected status code mismatch [fasthttp client]: ", string(response.Header.StatusMessage()),
					" body: ", string(respBody)}...))
	}

	if f.Compressor == nil {
		responseStruct.Body = respBody
		return responseStruct, nil
	}

	if !bytes.Contains(response.Header.Peek("Content-Encoding"), []byte(f.Compressor.GetName())) {
		return Response{}, errors.New(generic.StrCnct([]string{"requested decompressor mismatch by response content header : ", string(respBody)}...))
	}

	responseStruct.Body, err = ff.Compressor.Decompress(respBody)
	if err != nil {
		return Response{}, errors.New(generic.StrCnct([]string{"failed to decompress response body [crypt compression]: ", err.Error()}...))
	}
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

	if f.Compressor != nil {
		ff.Headers["Accept-Encoding"] = f.Compressor.GetName()
		switch f.Compressor.GetName() {
		case "gzip":
			ff.Compressor.SetLevel(compression.BestSpeed)
		case "br":
			ff.Compressor.SetLevel(compression.BrotliBestSpeed)
		}
	}

	if f.UserAgent != "" {
		ff.Headers["User-Agent"] = f.UserAgent
	}

	if f.ExpectedStatusCode == 0 {
		ff.ExpectedStatusCode = fasthttp.StatusOK
	}

	return ff, nil
}
