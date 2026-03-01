package response

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/rousage/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)
const (
	stateStatusLine = iota
	stateHeaders
	stateBody
	stateDone
)

var codeToReasonPhrase = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

type Writer struct {
	io.Writer
	writeState int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w, stateStatusLine}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writeState != stateStatusLine {
		return errors.New("state is not status line")
	}
	w.writeState = stateHeaders

	reasonPhrase := codeToReasonPhrase[statusCode]

	if _, err := io.WriteString(w, fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)); err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteHeaders(hs headers.Headers) error {
	if w.writeState != stateHeaders {
		return errors.New("state is not headers")
	}
	w.writeState = stateBody

	for k, v := range hs {
		if _, err := io.WriteString(w, fmt.Sprintf("%s: %s\r\n", k, v)); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, "\r\n")

	return err
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.writeState != stateBody {
		return 0, errors.New("state is not body")
	}
	w.writeState = stateDone

	return w.Write(body)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writeState != stateBody {
		return 0, errors.New("state is not body")
	}

	n, err := io.WriteString(w, fmt.Sprintf("%x\r\n", len(p)))
	if err != nil {
		return n, err
	}

	n2, err := w.Write(p)
	if err != nil {
		return n + n2, err
	}

	n3, err := io.WriteString(w, "\r\n")
	return n + n2 + n3, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writeState != stateBody {
		return 0, errors.New("state is not body")
	}
	w.writeState = stateDone

	return w.Write([]byte("0\r\n\r\n"))
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	hs := headers.NewHeaders()
	hs[strings.ToLower("Content-Length")] = fmt.Sprintf("%d", contentLen)
	hs[strings.ToLower("Connection")] = "close"
	hs[strings.ToLower("Content-Type")] = "text/plain"

	return hs
}
