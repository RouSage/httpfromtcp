package request

import (
	"errors"
	"io"
	"strings"
)

const crlf = "\r\n"
const bufferSize = 8
const (
	stateInitialized = iota
	stateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}
type Request struct {
	RequestLine RequestLine
	state       int
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{
		state: stateInitialized,
	}

	for request.state != stateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = stateDone
				break
			}

			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if numBytesParsed > 0 {
			// Remove the parsed data from the buffer
			copy(buf, buf[numBytesParsed:readToIndex])
			readToIndex -= numBytesParsed
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == stateDone {
		return 0, errors.New("error: trying to read data in a done state")
	}

	if r.state == stateInitialized {
		requestLine, bytesParsed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		// zero bytes parsed and no error = needs more data
		if bytesParsed == 0 {
			return 0, nil
		}

		r.RequestLine = requestLine
		r.state = stateDone

		return bytesParsed, nil
	}

	return 0, errors.New("error: unknown state")
}

func parseRequestLine(request []byte) (RequestLine, int, error) {
	requestStr := string(request)
	// if \r\n is not in the string, it needs more data
	if !strings.Contains(requestStr, crlf) {
		return RequestLine{}, 0, nil
	}

	requestLineStr := strings.Split(requestStr, crlf)[0]
	if requestLineStr == "" {
		return RequestLine{}, 0, errors.New("empty request line")
	}

	parts := strings.Split(requestLineStr, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, errors.New("invalid request line")
	}

	var (
		method        = parts[0]
		requestTarget = parts[1]
		httpVersion   = parts[2]
	)
	if !isValidMethod(method) {
		return RequestLine{}, 0, errors.New("invalid method")
	}
	version, ok := parseHttpVersion(httpVersion)
	if !ok {
		return RequestLine{}, 0, errors.New("invalid http version")
	}
	if requestTarget == "" {
		return RequestLine{}, 0, errors.New("invalid request target")
	}

	// Return the number of bytes consumed: length of request line + \r\n
	bytesConsumed := len(requestLineStr) + len(crlf)

	return RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, bytesConsumed, nil
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
