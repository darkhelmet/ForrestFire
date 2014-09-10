package goredis

import (
	"testing"
)

func TestSAdd(t *testing.T) {
	r.Del("key")
	if n, err := r.SAdd("key", "value"); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
	if n, _ := r.SAdd("key", "value"); n != 0 {
		t.Fail()
	}
}

func TestSCard(t *testing.T) {
	r.Del("key")
	r.SAdd("key", "value")
	if n, err := r.SCard("key"); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
}

func TestSDiff(t *testing.T) {
	r.Del("key1", "key2", "key3")
	r.SAdd("key1", "a", "b", "c", "d")
	r.SAdd("key2", "c")
	r.SAdd("key3", "a", "c", "e")
	if result, err := r.SDiff("key1", "key2", "key3"); err != nil {
		t.Error(err)
	} else if len(result) != 2 {
		t.Fail()
	}
}

func TestSDiffStore(t *testing.T) {
	r.Del("key1", "key2", "key3")
	r.SAdd("key1", "a", "b", "c", "d")
	r.SAdd("key2", "c")
	r.SAdd("key3", "a", "c", "e")
	if n, err := r.SDiffStore("key", "key1", "key2", "key3"); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Fail()
	}
}

func TestSInter(t *testing.T) {
	r.Del("key1", "key2", "key3")
	r.SAdd("key1", "a", "b", "c", "d")
	r.SAdd("key2", "c")
	r.SAdd("key3", "a", "c", "e")
	if result, err := r.SInter("key1", "key2", "key3"); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Fail()
	}
}

func TestSInterStore(t *testing.T) {
	r.Del("key1", "key2", "key3")
	r.SAdd("key1", "a", "b", "c", "d")
	r.SAdd("key2", "c")
	r.SAdd("key3", "a", "c", "e")
	if n, err := r.SInterStore("key", "key1", "key2", "key3"); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
}

func TestSIsMember(t *testing.T) {
	r.Del("key")
	r.SAdd("key", "value")
	if b, err := r.SIsMember("key", "value"); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
	if b, _ := r.SIsMember("key", "member"); b {
		t.Fail()
	}
}

func TestSMembers(t *testing.T) {
	r.Del("key")
	r.SAdd("key", "value")
	if result, err := r.SMembers("key"); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Fail()
	} else if result[0] != "value" {
		t.Fail()
	}
}

func TestSMove(t *testing.T) {
	r.Del("key", "key1")
	r.SAdd("key", "value")
	if b, err := r.SMove("key", "key1", "value"); err != nil {
		t.Error(err)
	} else if !b {
		t.Fail()
	}
}

func TestSPop(t *testing.T) {
	r.Del("key")
	r.SAdd("key", "value")
	if item, err := r.SPop("key"); err != nil {
		t.Error(err)
	} else if item == nil {
		t.Fail()
	} else if string(item) != "value" {
		t.Fail()
	}
	if item, _ := r.SPop("key"); item != nil {
		t.Fail()
	}
}

func TestSRandMember(t *testing.T) {
	r.Del("key")
	r.SAdd("key", "one", "two", "three")
	if m, err := r.SRandMember("key"); err != nil {
		t.Error(err)
	} else if m == nil {
		t.Fail()
	}
	if result, err := r.SRandMemberCount("key", 2); err != nil {
		t.Error(err)
	} else if len(result) != 2 {
		t.Fail()
	}
	if result, err := r.SRandMemberCount("key", -5); err != nil {
		t.Error(err)
	} else if len(result) != 5 {
		t.Fail()
	}
}

func TestSRem(t *testing.T) {
	r.Del("key")
	r.SAdd("key", "one", "two", "three")
	if n, err := r.SRem("key", "one", "four"); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
}

func TestSUnion(t *testing.T) {
	r.Del("key1", "key2", "key3")
	r.SAdd("key1", "a", "b", "c", "d")
	r.SAdd("key2", "c")
	r.SAdd("key3", "a", "c", "e")
	if result, err := r.SUnion("key1", "key2", "key3"); err != nil {
		t.Error(err)
	} else if len(result) != 5 {
		t.Fail()
	}
}

func TestSUnionStore(t *testing.T) {
	r.Del("key1", "key2", "key3")
	r.SAdd("key1", "a", "b", "c", "d")
	r.SAdd("key2", "c")
	r.SAdd("key3", "a", "c", "e")
	if n, err := r.SUnionStore("key", "key1", "key2", "key3"); err != nil {
		t.Error(err)
	} else if n != 5 {
		t.Fail()
	}
}

func TestSScan(t *testing.T) {
	r.Del("key")
	r.SAdd("key", "one", "two", "three")
	if _, list, err := r.SScan("key", 0, "", 0); err != nil {
		t.Error(err)
	} else if len(list) == 0 {
		t.Fail()
	}
}
