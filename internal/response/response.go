package response

import (
	"fmt"
	"io"

	"github.com/rousage/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var codeToReasonPhrase = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := codeToReasonPhrase[statusCode]

	if _, err := io.WriteString(w, fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)); err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	hs := headers.NewHeaders()
	hs["Content-Length"] = fmt.Sprintf("%d", contentLen)
	hs["Connection"] = "close"
	hs["Content-Type"] = "text/plain"

	return hs
}

func WriteHeaders(w io.Writer, hs headers.Headers) error {
	for k, v := range hs {
		if _, err := io.WriteString(w, fmt.Sprintf("%s: %s\r\n", k, v)); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, "\r\n")

	return err
}
