package ht2p

import (
	"errors"
	"net/url"

	"github.com/D3vl0per/crypt/generic"
)

type Response struct {
	Body       []byte
	Headers    map[string][]string
	StatusCode int
}

type HttpClient interface {
	Request() (Response, error)
	MultiRequest(urls []string) (Response, []error)
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
