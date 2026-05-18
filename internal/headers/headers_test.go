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

	// Test: Valid done
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Valid two headers and done
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nFoo: bar\r\n\r\n")
	n1, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n1)
	assert.False(t, done)
	n2, done, err := headers.Parse(data[n1:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "bar", headers["foo"])
	assert.Equal(t, 10, n2)
	assert.False(t, done)
	n3, done, err := headers.Parse(data[n1+n2:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "bar", headers["foo"])
	assert.Equal(t, 2, n3)
	assert.True(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:   localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 29, n)
	assert.False(t, done)

	// Test: Valid case-insensitive name
	headers = NewHeaders()
	data = []byte("HoSt:   localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "", headers["HoSt"])
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 29, n)
	assert.False(t, done)

	// Test: Valid special symbols
	headers = NewHeaders()
	data = []byte("Ho!#$%&'*+-.^_`|~sT:   localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["ho!#$%&'*+-.^_`|~st"])
	assert.Equal(t, 44, n)
	assert.False(t, done)

	// Test: Valid no headers
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 0, len(headers))
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid no separator header
	headers = NewHeaders()
	data = []byte("Host localhost\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid symbols in header name
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
