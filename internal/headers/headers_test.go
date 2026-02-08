package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:        localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	data = []byte("Content-Type: application/json\r\nHeader: value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", headers["content-type"])
	assert.Equal(t, 32, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[32:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "value", headers["header"])
	assert.Equal(t, 15, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[47:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "application/json", headers["content-type"])
	assert.Equal(t, "value", headers["header"])
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Header with multiple values
	headers = NewHeaders()
	data = []byte("Header: value-one\r\nHeader: value-two\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "value-one", headers["header"])
	assert.Equal(t, 19, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[19:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "value-one, value-two", headers["header"])
	assert.Equal(t, 19, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[38:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "value-one, value-two", headers["header"])
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character in header
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
