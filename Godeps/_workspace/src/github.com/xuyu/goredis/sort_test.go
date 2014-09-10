package goredis

import (
	"testing"
)

func TestSortList(t *testing.T) {
	r.Del("key")
	r.LPush("key", "one", "two", "three")
	rp, err := r.Sort("key").Limit(0, 2).DESC().Alpha(true).Run()
	if err != nil {
		t.Error(err)
	}
	if result, err := rp.ListValue(); err != nil {
		t.Error(err)
	} else if len(result) != 2 {
		t.Fail()
	} else if result[0] != "two" {
		t.Fail()
	}
}
