package goredis

import (
	"testing"
)

func TestHDel(t *testing.T) {
	r.Del("key")
	if n, err := r.HDel("key", "hkey"); err != nil {
		t.Error(err)
	} else if n != 0 {
		t.Fail()
	}
}

func TestHExists(t *testing.T) {
	r.Del("key")
	if b, err := r.HExists("key", "field"); err != nil {
		t.Error(err)
	} else if b {
		t.Fail()
	}
	r.HSet("key", "field", "value")
	if b, _ := r.HExists("key", "field"); !b {
		t.Fail()
	}
}

func TestHGet(t *testing.T) {
	r.Del("key")
	if data, err := r.HGet("key", "field"); err != nil {
		t.Error(err)
	} else if data != nil {
		t.Fail()
	}
	r.HSet("key", "field", "value")
	if data, _ := r.HGet("key", "field"); string(data) != "value" {
		t.Fail()
	}
}

func TestHGetAll(t *testing.T) {
	r.Del("key")
	if m, err := r.HGetAll("key"); err != nil {
		t.Error(err)
	} else if len(m) != 0 {
		t.Fail()
	}
	r.HSet("key", "field", "value")
	if m, _ := r.HGetAll("key"); m["field"] != "value" {
		t.Fail()
	}
}

func TestHIncrBy(t *testing.T) {
	r.Del("key")
	r.HSet("key", "field", "10")
	if n, err := r.HIncrBy("key", "field", 2); err != nil {
		t.Error(err)
	} else if n != 12 {
		t.Fail()
	}
}

func TestHIncrByFloat(t *testing.T) {
	r.Del("key")
	r.HSet("key", "field", "10")
	if f, err := r.HIncrByFloat("key", "field", 0.1); err != nil {
		t.Error(err)
	} else if f != 10.1 {
		t.Fail()
	}
}

func TestHKeys(t *testing.T) {
	r.Del("key")
	r.HSet("key", "field", "value")
	if keys, err := r.HKeys("key"); err != nil {
		t.Error(err)
	} else if len(keys) != 1 {
		t.Fail()
	} else if keys[0] != "field" {
		t.Fail()
	}
}

func TestHLen(t *testing.T) {
	r.Del("key")
	r.HSet("key", "field", "value")
	if n, err := r.HLen("key"); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
}

func TestHMGet(t *testing.T) {
	r.HSet("key", "field", "value")
	data, err := r.HMGet("key", "field", "nofield")
	if err != nil {
		t.Error(err)
	}
	if string(data[0]) != "value" {
		t.Fail()
	}
	if data[1] != nil {
		t.Fail()
	}
}

func TestHMSet(t *testing.T) {
	pairs := map[string]string{
		"field": "value",
		"foo":   "bar",
	}
	r.Del("key")
	if err := r.HMSet("key", pairs); err != nil {
		t.Error(err)
	}
}

func TestHSet(t *testing.T) {
	r.Del("key")
	if b, err := r.HSet("key", "field", "value"); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
}

func TestHSetnx(t *testing.T) {
	r.Del("key")
	r.HSet("key", "field", "value")
	if b, err := r.HSetnx("key", "field", "value"); err != nil {
		t.Error(err)
	} else if b {
		t.Fail()
	}
	r.Del("key")
	if b, _ := r.HSetnx("key", "field", "value"); !b {
		t.Fail()
	}
}

func TestHVals(t *testing.T) {
	r.Del("key")
	r.HSet("key", "field", "value")
	if vals, err := r.HVals("key"); err != nil {
		t.Error(err)
	} else if len(vals) != 1 {
		t.Fail()
	} else if vals[0] != "value" {
		t.Fail()
	}
}

func TestHScan(t *testing.T) {
	r.Del("key")
	r.HSet("key", "field", "value")
	if _, hash, err := r.HScan("key", 0, "", 0); err != nil {
		t.Error(err)
	} else if len(hash) == 0 {
		t.Fail()
	}
}
