package mc

import (
	"bytes"
	"github.com/bmizerany/assert"
	"testing"
	"net"
	"runtime"
)

const mcAddr = "localhost:11211"

func TestMCSimple(t *testing.T) {
	nc, err := net.Dial("tcp", mcAddr)
	assert.Equalf(t, nil, err, "%v", err)

	cn := &Conn{rwc: nc, buf: new(bytes.Buffer)}

	if runtime.GOOS != "darwin" {
		println("Not on Darwin, testing auth")
		err = cn.Auth("mcgo", "foo")
		assert.Equalf(t, nil, err, "%v", err)
	}

	err = cn.Del("foo")
	if err != ErrNotFound {
		assert.Equalf(t, nil, err, "%v", err)
	}

	_, _, _, err = cn.Get("foo")
	assert.Equalf(t, ErrNotFound, err, "%v", err)

	err = cn.Set("foo", "bar", 0, 0, 0)
	assert.Equalf(t, nil, err, "%v", err)

	// unconditional SET
	err = cn.Set("foo", "bar", 0, 0, 0)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Set("foo", "bar", 1, 0, 0)
	assert.Equalf(t, ErrKeyExists, err, "%v", err)

	v, _, _, err := cn.Get("foo")
	assert.Equalf(t, nil, err, "%v", err)
	assert.Equal(t, "bar", v)

	err = cn.Del("n")
	if err != ErrNotFound {
		assert.Equalf(t, nil, err, "%v", err)
	}

	n, cas, err := cn.Incr("n", 1, 0, 0)
	assert.Equalf(t, nil, err, "%v", err)
	assert.NotEqual(t, 0, cas)
	assert.Equal(t, 1, n)

	n, cas, err = cn.Incr("n", 1, 0, 0)
	assert.Equalf(t, nil, err, "%v", err)
	assert.NotEqual(t, 0, cas)
	assert.Equal(t, 2, n)

	n, cas, err = cn.Decr("n", 1, 0, 0)
	assert.Equalf(t, nil, err, "%v", err)
	assert.NotEqual(t, 0, cas)
	assert.Equal(t, 1, n)
}
