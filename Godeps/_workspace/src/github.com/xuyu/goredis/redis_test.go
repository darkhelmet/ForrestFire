package goredis

import (
	"fmt"
	"testing"
	"time"
)

var (
	network  = "tcp"
	address  = "192.168.84.250:6379"
	db       = 1
	password = ""
	timeout  = 5 * time.Second
	maxidle  = 1
	r        *Redis

	format = "tcp://auth:%s@%s/%d?timeout=%s&maxidle=%d"
)

func init() {
	client, err := DialTimeout(network, address, db, password, timeout, maxidle)
	if err != nil {
		panic(err)
	}
	r = client
}

func TestDial(t *testing.T) {
	redis, err := Dial(&DialConfig{network, address, db, password, timeout, maxidle})
	if err != nil {
		t.Error(err)
	} else if err := redis.Ping(); err != nil {
		t.Error(err)
	}
	redis.pool.Close()
}

func TestDialTimeout(t *testing.T) {
	redis, err := DialTimeout(network, address, db, password, timeout, maxidle)
	if err != nil {
		t.Error(err)
	} else if err := redis.Ping(); err != nil {
		t.Error(err)
	}
	redis.pool.Close()
}

func TestDiaURL(t *testing.T) {
	redis, err := DialURL(fmt.Sprintf(format, password, address, db, timeout.String(), maxidle))
	if err != nil {
		t.Fatal(err)
	} else if err := redis.Ping(); err != nil {
		t.Error(err)
	}
	redis.pool.Close()
}
