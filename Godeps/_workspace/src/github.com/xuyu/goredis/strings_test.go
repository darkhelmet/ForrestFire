package goredis

import (
	"testing"
)

func TestAppend(t *testing.T) {
	r.Del("key")
	n, err := r.Append("key", "value")
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Fail()
	}
	n, err = r.Append("key", "value")
	if err != nil {
		t.Error(err)
	}
	if n != 10 {
		t.Fail()
	}
	r.Del("key")
	r.LPush("key", "value")
	if _, err := r.Append("key", "value"); err == nil {
		t.Error(err)
	}
}

func TestBitCount(t *testing.T) {
	r.Set("key", "foobar", 0, 0, false, false)
	n, err := r.BitCount("key", 0, -1)
	if err != nil {
		t.Error(err)
	}
	if n != 26 {
		t.Fail()
	}
	n, _ = r.BitCount("key", 0, 0)
	if n != 4 {
		t.Fail()
	}
}

func TestBitOp(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if _, err := r.BitOp("NOT", "key2", "key"); err != nil {
		t.Error(err)
	}
}

func TestDecr(t *testing.T) {
	r.Set("key", "10", 0, 0, false, false)
	if n, err := r.Decr("key"); err != nil {
		t.Error(err)
	} else if n != 9 {
		t.Fail()
	}
	r.Set("key", "value", 0, 0, false, false)
	if _, err := r.Decr("key"); err == nil {
		t.Fail()
	}
}

func TestDecrby(t *testing.T) {
	r.Set("key", "10", 0, 0, false, false)
	if n, err := r.DecrBy("key", 2); err != nil {
		t.Error(err)
	} else if n != 8 {
		t.Fail()
	}
	r.Set("key", "value", 0, 0, false, false)
	if _, err := r.DecrBy("key", 2); err == nil {
		t.Fail()
	}
}

func TestGet(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if value, err := r.Get("key"); err != nil {
		t.Error(err)
	} else if string(value) != "value" {
		t.Fail()
	}
	r.Del("key")
	if value, _ := r.Get("key"); value != nil {
		t.Fail()
	}
}

func BenchmarkGet(b *testing.B) {
	r.Set("key", "value", 0, 0, false, false)
	for i := 0; i < b.N; i++ {
		r.Get("key")
	}
}

func TestGetBit(t *testing.T) {
	r.SetBit("key", 7, 1)
	n, err := r.GetBit("key", 6)
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Fail()
	}
	n, _ = r.GetBit("key", 7)
	if n != 1 {
		t.Fail()
	}
}

func TestGetRange(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	s, err := r.GetRange("key", 0, -1)
	if err != nil {
		t.Error(err)
	}
	if s != "value" {
		t.Fail()
	}
	s, _ = r.GetRange("key", 0, 0)
	if s != "v" {
		t.Fail()
	}
}

func TestGetSet(t *testing.T) {
	r.Del("key")
	old, err := r.GetSet("key", "value")
	if err != nil {
		t.Error(err)
	}
	if old != nil {
		t.Fail()
	}
	old, _ = r.GetSet("key", "")
	if string(old) != "value" {
		t.Fail()
	}
	value, _ := r.Get("key")
	if string(value) != "" {
		t.Fail()
	}
}

func TestIncr(t *testing.T) {
	r.Set("key", "10", 0, 0, false, false)
	n, err := r.Incr("key")
	if err != nil {
		t.Error(err)
	}
	if n != 11 {
		t.Fail()
	}
}

func BenchmarkIncr(b *testing.B) {
	r.Del("key")
	for i := 0; i < b.N; i++ {
		r.Incr("key")
	}
}

func TestIncrBy(t *testing.T) {
	r.Set("key", "10", 0, 0, false, false)
	n, err := r.IncrBy("key", 2)
	if err != nil {
		t.Error(err)
	}
	if n != 12 {
		t.Fail()
	}
}

func TestIncrByFloat(t *testing.T) {
	r.Set("key", "10", 0, 0, false, false)
	f, err := r.IncrByFloat("key", 0.1)
	if err != nil {
		t.Error(err)
	}
	if f != 10.1 {
		t.Fail()
	}
}

func TestMGet(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	ret, err := r.MGet("key", "key1")
	if err != nil {
		t.Error(err)
	}
	if len(ret) != 2 {
		t.Fail()
	}
	if string(ret[0]) != "value" {
		t.Fail()
	}
	if ret[1] != nil {
		t.Fail()
	}
}

func TestMSet(t *testing.T) {
	pairs := map[string]string{
		"key":  "value",
		"key1": "value1",
	}
	if err := r.MSet(pairs); err != nil {
		t.Error(err)
	}
	value, _ := r.Get("key1")
	if string(value) != "value1" {
		t.Fail()
	}
}

func TestMSetnx(t *testing.T) {
	r.Del("key")
	r.Set("key1", "value", 0, 0, false, false)
	pairs := map[string]string{
		"key":  "value",
		"key1": "value1",
	}
	if b, err := r.MSetnx(pairs); err != nil {
		t.Error(err)
	} else if b {
		t.Fail()
	}
}

func TestPSetex(t *testing.T) {
	if err := r.PSetex("key", 100, "value"); err != nil {
		t.Error(err)
	}
	n, _ := r.PTTL("key")
	if n < 0 {
		t.Fail()
	}
	v, _ := r.Get("key")
	if string(v) != "value" {
		t.Fail()
	}
}

func TestSet(t *testing.T) {
	if err := r.Set("key", "value", 0, 0, false, false); err != nil {
		t.Error(err)
	}
}

func BenchmarkSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r.Set("key", "value", 0, 0, false, false)
	}
}

func TestSetBit(t *testing.T) {
	if _, err := r.SetBit("key", 7, 1); err != nil {
		t.Error(err)
	}
	bit, _ := r.GetBit("key", 7)
	if bit != 1 {
		t.Fail()
	}
}

func TestSetex(t *testing.T) {
	if err := r.Setex("key", 10, "value"); err != nil {
		t.Error(err)
	}
	n, _ := r.TTL("key")
	if n < 0 {
		t.Fail()
	}
}

func TestSetnx(t *testing.T) {
	r.Del("key")
	if b, err := r.Setnx("key", "value"); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
	if b, _ := r.Setnx("key", "value"); b {
		t.Fail()
	}
}

func TestSetRange(t *testing.T) {
	r.Del("key")
	if n, err := r.SetRange("key", 2, "value"); err != nil {
		t.Error(err)
	} else if n != 7 {
		t.Fail()
	}
}

func TestStrlen(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	n, err := r.StrLen("key")
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Fail()
	}
}
