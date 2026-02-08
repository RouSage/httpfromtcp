package headers

import (
	"errors"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	headersStr := string(data)
	// if \r\n is not in the string, it needs more data
	if !strings.Contains(headersStr, crlf) {
		return 0, false, nil
	}
	// if CRLF is at the start of the data, it's the end of headers
	// return done=true and consume the CRLF
	if strings.HasPrefix(headersStr, crlf) {
		return len(crlf), true, nil
	}

	headerStr := strings.Split(headersStr, crlf)[0]
	if headerStr == "" {
		return 0, false, errors.New("empty header")
	}

	parts := strings.SplitN(headerStr, ":", 2)
	if len(parts) != 2 {
		return 0, false, errors.New("invalid header")
	}

	parsedKey, ok := parseKey(parts[0])
	if !ok {
		return 0, false, errors.New("invalid header key")
	}
	parsedValue, ok := parseValue(parts[1])
	if !ok {
		return 0, false, errors.New("invalid header value")
	}

	val, ok := h[parsedKey]
	if ok {
		h[parsedKey] = val + ", " + parsedValue
	} else {
		h[parsedKey] = parsedValue
	}

	return len(headerStr) + len(crlf), false, nil
}

func parseKey(key string) (string, bool) {
	// There cannot be an empty space between the key and the colon
	if strings.HasSuffix(key, " ") {
		return "", false
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return "", false
	}

	if !isValidToken(trimmedKey) {
		return "", false
	}

	return strings.ToLower(trimmedKey), true
}

// https://datatracker.ietf.org/doc/html/rfc9110#name-tokens
func isValidToken(s string) bool {
	for _, c := range s {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(c rune) bool {
	if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

func parseValue(value string) (string, bool) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "", false
	}

	return trimmedValue, true
}
