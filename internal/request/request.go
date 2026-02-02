package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, err := parseRequestLine(string(request))
	if err != nil {
		return nil, err
	}
	return &Request{
		RequestLine: requestLine,
	}, nil
}

func parseRequestLine(request string) (RequestLine, error) {
	requestLineStr := strings.Split(request, "\r\n")[0]
	if requestLineStr == "" {
		return RequestLine{}, errors.New("empty request line")
	}

	parts := strings.Split(requestLineStr, " ")
	if len(parts) != 3 {
		return RequestLine{}, errors.New("invalid request line")
	}

	var (
		method        = parts[0]
		requestTarget = parts[1]
		httpVersion   = parts[2]
	)
	if !isValidMethod(method) {
		return RequestLine{}, errors.New("invalid method")
	}
	version, ok := parseHttpVersion(httpVersion)
	if !ok {
		return RequestLine{}, errors.New("invalid http version")
	}
	if requestTarget == "" {
		return RequestLine{}, errors.New("invalid request target")
	}

	return RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}

func isValidMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH":
		return true
	default:
		return false
	}
}

func parseHttpVersion(httpVersion string) (string, bool) {
	if httpVersion != "HTTP/1.1" {
		return "", false
	}

	return "1.1", true
}
