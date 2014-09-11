package goredis

import (
	"testing"
	"time"
)

func TestBgRewriteAof(t *testing.T) {
	if err := r.BgRewriteAof(); err != nil {
		t.Error(err)
	}
}

func TestBgSave(t *testing.T) {
	if err := r.BgSave(); err != nil {
		t.Error(err)
	}
}

func TestClientKill(t *testing.T) {
	if err := r.ClientKill("127.0.0.1", 80); err == nil {
		t.Fail()
	}
}

func TestClientList(t *testing.T) {
	_, err := r.ClientList()
	if err != nil {
		t.Error(err)
	}
}

func TestClientGetName(t *testing.T) {
	if _, err := r.ClientGetName(); err != nil {
		t.Error(err)
	}
}

func TestClientSetName(t *testing.T) {
	if err := r.ClientSetName("name"); err != nil {
		t.Error(err)
	}
}

func TestConfigGet(t *testing.T) {
	if result, err := r.ConfigGet("daemonize"); err != nil {
		t.Error(err)
	} else if result == nil {
		t.Fail()
	} else if len(result) != 1 {
		t.Fail()
	}
}

func TestConfigResetStat(t *testing.T) {
	if err := r.ConfigResetStat(); err != nil {
		t.Error(err)
	}
}

func TestDBSize(t *testing.T) {
	r.FlushDB()
	n, err := r.DBSize()
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Fail()
	}
}

func TestDebugObject(t *testing.T) {
	r.Del("key")
	r.LPush("key", "value")
	if _, err := r.DebugObject("key"); err != nil {
		t.Error(err)
	}
}

func TestFlushAll(t *testing.T) {
	if err := r.FlushAll(); err != nil {
		t.Error(err)
	}
}

func TestFlushDB(t *testing.T) {
	if err := r.FlushDB(); err != nil {
		t.Error(err)
	}
}

func TestInfo(t *testing.T) {
	if _, err := r.Info(""); err != nil {
		t.Error(err)
	}
	if _, err := r.Incr("CPU"); err != nil {
		t.Error(err)
	}
}

func TestLastSave(t *testing.T) {
	r.Save()
	if timestamp, err := r.LastSave(); err != nil {
		t.Error(err)
	} else if timestamp <= 0 {
		t.Fail()
	}
}

func TestMonitor(t *testing.T) {
	quit := false
	m, err := r.Monitor()
	if err != nil {
		t.Error(err)
	}
	defer m.Close()
	go func() {
		for {
			if s, err := m.Receive(); err != nil {
				if !quit {
					t.Error(err)
				}
			} else if s == "" {
				t.Fail()
			}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	r.LPush("key", "value")
	time.Sleep(100 * time.Microsecond)
}

func TestSave(t *testing.T) {
	if err := r.Save(); err != nil {
		t.Error(err)
	}
}

func TestSlowLogGet(t *testing.T) {
	r.Del("key")
	r.LPush("key", "value")
	if result, err := r.SlowLogGet(1); err != nil {
		t.Error(err)
	} else if len(result) > 1 {
		t.Fail()
	}
}

func TestSlowLogLen(t *testing.T) {
	r.Del("key")
	r.LPush("key", "value")
	if _, err := r.SlowLogLen(); err != nil {
		t.Error(err)
	}
}

func TestSlowLogReset(t *testing.T) {
	if err := r.SlowLogReset(); err != nil {
		t.Error(err)
	}
}

func TestTime(t *testing.T) {
	tt, err := r.Time()
	if err != nil {
		t.Error(err)
	}
	if len(tt) != 2 {
		t.Fail()
	}
}
