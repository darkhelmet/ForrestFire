package goredis

import (
	"testing"
)

func TestTransaction(t *testing.T) {
	transaction, err := r.Transaction()
	if err != nil {
		t.Error(err)
	}
	defer transaction.Close()
	if err := transaction.Command("DEL", "key"); err != nil {
		t.Error(err)
	}
	if err := transaction.Command("SET", "key", 1); err != nil {
		t.Error(err)
	}
	if err := transaction.Command("INCR", "key"); err != nil {
		t.Error(err)
	}
	if err := transaction.Command("GET", "key"); err != nil {
		t.Error(err)
	}
	result, err := transaction.Exec()
	if err != nil {
		t.Error(err)
	}
	if len(result) != 4 {
		t.Fail()
	}
	if s, err := result[3].StringValue(); err != nil || s != "2" {
		t.Fail()
	}
}

func TestWatch(t *testing.T) {
	transaction, err := r.Transaction()
	if err != nil {
		t.Error(err)
	}
	defer transaction.Close()
	if err := transaction.Watch("key"); err != nil {
		t.Error(err)
	}
}

func TestUnWatch(t *testing.T) {
	transaction, err := r.Transaction()
	if err != nil {
		t.Error(err)
	}
	defer transaction.Close()
	transaction.Watch("key")
	if err := transaction.UnWatch(); err != nil {
		t.Error(err)
	}
}

func TestDiscard(t *testing.T) {
	transaction, err := r.Transaction()
	if err != nil {
		t.Error(err)
	}
	defer transaction.Close()
	transaction.Command("SET", "KEY", 1)
	if transaction.Discard(); err != nil {
		t.Error(err)
	}
}
