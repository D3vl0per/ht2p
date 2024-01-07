package ht2p

import (
	"errors"
	"net/http"

	"github.com/D3vl0per/crypt/generic"
	"github.com/valyala/fasthttp"
)

type compressors int

const (
	All compressors = iota
	Deflate
	Gzip
	Brotil
)

type FastHttp struct {
	URL                string
	URLParameters      map[string]string
	Method             string
	Body               []byte
	Headers            map[string]string
	ExpectedStatusCode int
	Client             fasthttp.Client
	Compressor         compressors
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
		Headers:    fastHeaderToMap(&response.Header),
	}

	responseStruct.Body = response.Body()

	if responseStruct.StatusCode != f.ExpectedStatusCode {
		return responseStruct,
			errors.New(
				generic.StrCnct([]string{
					"expected status code mismatch [fasthttp client]: ", string(response.Header.StatusMessage()),
					" body: ", string(responseStruct.Body)}...))
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

	if f.UserAgent != "" {
		ff.Headers["User-Agent"] = f.UserAgent
	}

	if f.ExpectedStatusCode == 0 {
		ff.ExpectedStatusCode = fasthttp.StatusOK
	}

	if f.Compressor == All {
		ff.Headers["Accept-Encoding"] = "gzip, deflate, br"
		return ff, nil
	}

	if f.Compressor == Brotil {
		ff.Headers["Accept-Encoding"] = "br"
		return ff, nil
	}

	if f.Compressor == Gzip {
		ff.Headers["Accept-Encoding"] = "gzip"
		return ff, nil
	}

	ff.Headers["Accept-Encoding"] = "deflate"

	return ff, nil
}

func fastHeaderToMap(header *fasthttp.ResponseHeader) map[string][]string {
	headers := make(map[string][]string)
	header.VisitAll(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], string(value))
	})
	return headers
}
