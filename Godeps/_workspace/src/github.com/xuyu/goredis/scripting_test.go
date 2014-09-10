package goredis

import (
	"testing"
)

func TestEval(t *testing.T) {
	rp, err := r.Eval("return {KEYS[1], KEYS[2], ARGV[1], ARGV[2]}", []string{"key1", "key2"}, []string{"arg1", "arg2"})
	if err != nil {
		t.Error(err)
	} else if l, err := rp.ListValue(); err != nil {
		t.Error(err)
	} else if l[0] != "key1" || l[3] != "arg2" {
		t.Fail()
	}
	rp, err = r.Eval("return redis.call('set','foo','bar')", nil, nil)
	if err != nil {
		t.Error(err)
	} else if err := rp.OKValue(); err != nil {
		t.Error(err)
	}
	rp, err = r.Eval("return 10", nil, nil)
	if err != nil {
		t.Error(err)
	} else if n, err := rp.IntegerValue(); err != nil {
		t.Error(err)
	} else if n != 10 {
		t.Fail()
	}
	rp, err = r.Eval("return {1,2,{3,'Hello World!'}}", nil, nil)
	if err != nil {
		t.Error(err)
	} else if len(rp.Multi) != 3 {
		t.Fail()
	} else if rp.Multi[2].Multi[0].Integer != 3 {
		t.Fail()
	} else if s, err := rp.Multi[2].Multi[1].StringValue(); err != nil || s != "Hello World!" {
		t.Fail()
	}
}

func TestEvalSha(t *testing.T) {
	r.ScriptFlush()
	sha1, _ := r.ScriptLoad("return 10")
	if rp, err := r.EvalSha(sha1, nil, nil); err != nil {
		t.Error(err)
	} else if rp.Type != IntegerReply {
		t.Fail()
	} else if rp.Integer != 10 {
		t.Fail()
	}
}

func TestScriptExists(t *testing.T) {
	r.ScriptFlush()
	sha1, _ := r.ScriptLoad("return 10")
	if bs, err := r.ScriptExists(sha1, "sha1"); err != nil {
		t.Error(err)
	} else if len(bs) != 2 {
		t.Fail()
	} else if !bs[0] {
		t.Fail()
	} else if bs[1] {
		t.Fail()
	}
}

func TestScriptFlush(t *testing.T) {
	sha1, _ := r.ScriptLoad("return 10")
	r.ScriptFlush()
	if bs, err := r.ScriptExists(sha1); err != nil {
		t.Error(err)
	} else if bs[0] {
		t.Fail()
	}
}

func TestScriptKill(t *testing.T) {
	if err := r.ScriptKill(); err == nil {
		t.Error(err)
	}
}
