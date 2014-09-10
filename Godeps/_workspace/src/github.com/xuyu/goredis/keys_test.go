package goredis

import (
	"testing"
	"time"
)

func TestDel(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if n, err := r.Del("key"); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
}

func TestDump(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	data, err := r.Dump("key")
	if err != nil {
		t.Error(err)
	}
	if data == nil || len(data) == 0 {
		t.Fail()
	}
}

func TestExists(t *testing.T) {
	r.Del("key")
	b, err := r.Exists("key")
	if err != nil {
		t.Error(err)
	}
	if b {
		t.Fail()
	}
}

func TestExpire(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if b, err := r.Expire("key", 10); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
	if n, err := r.TTL("key"); err != nil {
		t.Error(err)
	} else if n != 10 {
		t.Fail()
	}
}

func TestExpireAt(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if b, err := r.ExpireAt("key", time.Now().Add(10*time.Second).Unix()); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
	if n, err := r.TTL("key"); err != nil {
		t.Error(err)
	} else if n < 0 {
		t.Fail()
	}
}

func TestKeys(t *testing.T) {
	r.FlushDB()
	keys, err := r.Keys("*")
	if err != nil {
		t.Error(err)
	}
	if len(keys) != 0 {
		t.Fail()
	}
	r.Set("key", "value", 0, 0, false, false)
	keys, err = r.Keys("*")
	if err != nil {
		t.Error(err)
	}
	if len(keys) != 1 || keys[0] != "key" {
		t.Fail()
	}
}

func TestMove(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if _, err := r.Move("key", db+1); err != nil {
		t.Error(err)
	}
}

func TestObject(t *testing.T) {
	r.Del("key")
	r.LPush("key", "hello world")
	if rp, err := r.Object("refcount", "key"); err != nil {
		t.Error(err)
	} else if rp.Type != IntegerReply {
		t.Fail()
	}
	if rp, err := r.Object("encoding", "key"); err != nil {
		t.Error(err)
	} else if rp.Type != BulkReply {
		t.Fail()
	}
	if rp, err := r.Object("idletime", "key"); err != nil {
		t.Error(err)
	} else if rp.Type != IntegerReply {
		t.Fail()
	}
}

func TestPersist(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	r.Expire("key", 500)
	if n, _ := r.TTL("key"); n < 0 {
		t.Fail()
	}
	if b, err := r.Persist("key"); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
	if n, _ := r.TTL("key"); n > 0 {
		t.Fail()
	}
}

func TestPExpire(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if b, err := r.PExpire("key", 100); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
}

func TestPExpireAt(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if b, err := r.PExpireAt("key", time.Now().Add(500*time.Second).Unix()*1000); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
}

func TestPTTL(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	r.PExpire("key", 1000)
	if n, err := r.PTTL("key"); err != nil {
		t.Error(err)
	} else if n < 0 {
		t.Fail()
	}
}

func TestRandomKey(t *testing.T) {
	r.FlushDB()
	key, err := r.RandomKey()
	if err != nil {
		t.Error(err)
	}
	if key != nil {
		t.Fail()
	}
	r.Set("key", "value", 0, 0, false, false)
	key, _ = r.RandomKey()
	if string(key) != "key" {
		t.Fail()
	}
}

func TestRename(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	if err := r.Rename("key", "newkey"); err != nil {
		t.Error(err)
	}
	b, _ := r.Exists("key")
	if b {
		t.Fail()
	}
	v, _ := r.Get("newkey")
	if string(v) != "value" {
		t.Fail()
	}
}

func TestRenamenx(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	r.Set("newkey", "value", 0, 0, false, false)
	if b, err := r.Renamenx("key", "newkey"); err != nil {
		t.Error(err)
	} else if b {
		t.Fail()
	}
}

func TestRestore(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	data, _ := r.Dump("key")
	r.Del("key")
	if err := r.Restore("key", 0, string(data)); err != nil {
		t.Error(err)
	}
}

func TestTTL(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	r.Expire("key", 100)
	n, err := r.TTL("key")
	if err != nil {
		t.Error(err)
	}
	if n < 0 {
		t.Fail()
	}
	r.Persist("key")
	n, _ = r.TTL("key")
	if n > 0 {
		t.Fail()
	}
}

func TestType(t *testing.T) {
	r.Set("key", "value", 0, 0, false, false)
	ty, err := r.Type("key")
	if err != nil {
		t.Error(err)
	}
	if ty != "string" {
		t.Fail()
	}
}

func TestScan(t *testing.T) {
	r.FlushDB()
	cursor, list, err := r.Scan(0, "", 0)
	if err != nil {
		t.Error(err)
	} else if len(list) != 0 {
		t.Fail()
	} else if cursor != 0 {
		t.Fail()
	}
}
