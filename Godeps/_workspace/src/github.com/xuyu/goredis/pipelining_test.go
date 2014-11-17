package goredis

import (
	"testing"
)

func TestPipelining(t *testing.T) {
	p, err := r.Pipelining()
	if err != nil {
		t.Error(err)
	}
	defer p.Close()
	n := 3
	for i := 0; i < n; i++ {
		if err := p.Command("PING"); err != nil {
			t.Error(err)
		}
	}
	rps, err := p.ReceiveAll()
	if err != nil {
		t.Error(err)
	}
	if len(rps) != n {
		t.Fail()
	}
	if s, err := rps[1].StatusValue(); err != nil {
		t.Error(err)
	} else if s != "PONG" {
		t.Fail()
	}
}
