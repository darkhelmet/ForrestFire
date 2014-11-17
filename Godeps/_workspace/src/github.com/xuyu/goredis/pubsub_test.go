package goredis

import (
	"testing"
	"time"
)

func TestPublish(t *testing.T) {
	if _, err := r.Publish("channel", "message"); err != nil {
		t.Error(err)
	}
}

func TestSubscribe(t *testing.T) {
	quit := make(chan bool)
	sub, err := r.PubSub()
	if err != nil {
		t.Error(err)
	}
	defer sub.Close()
	go func() {
		if err := sub.Subscribe("channel"); err != nil {
			t.Error(err)
			quit <- true
			return
		}
		for {
			list, err := sub.Receive()
			if err != nil {
				t.Error(err)
				quit <- true
				break
			}
			if list[0] == "message" {
				if list[1] != "channel" || list[2] != "message" {
					t.Fail()
				}
				quit <- true
				break
			}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	r.Publish("channel", "message")
	time.Sleep(100 * time.Millisecond)
	<-quit
}

func TestPSubscribe(t *testing.T) {
	quit := make(chan bool)
	psub, err := r.PubSub()
	if err != nil {
		t.Error(err)
	}
	defer psub.Close()
	go func() {
		if err := psub.PSubscribe("news.*"); err != nil {
			t.Error(err)
			quit <- true
			return
		}
		for {
			list, err := psub.Receive()
			if err != nil {
				t.Error(err)
				quit <- true
				break
			}
			if list[0] == "pmessage" {
				if list[1] != "news.*" || list[2] != "news.china" || list[3] != "message" {
					t.Fail()
				}
				quit <- true
				break
			}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	r.Publish("news.china", "message")
	time.Sleep(100 * time.Millisecond)
	<-quit
}

func TestUnSubscribe(t *testing.T) {
	quit := false
	ch := make(chan bool)
	sub, err := r.PubSub()
	if err != nil {
		t.Error(err)
	}
	defer sub.Close()
	go func() {
		for {
			if _, err := sub.Receive(); err != nil {
				if !quit {
					t.Error(err)
				}
			}
			ch <- true
		}
	}()
	time.Sleep(100 * time.Millisecond)
	sub.Subscribe("channel")
	time.Sleep(100 * time.Millisecond)
	<-ch
	if len(sub.Channels) != 1 {
		t.Fail()
	}
	if err := sub.UnSubscribe("channel"); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	<-ch
	time.Sleep(100 * time.Millisecond)
	if len(sub.Channels) != 0 {
		t.Fail()
	}
	quit = true
}

func TestPUnSubscribe(t *testing.T) {
	quit := false
	ch := make(chan bool)
	sub, err := r.PubSub()
	if err != nil {
		t.Error(err)
	}
	defer sub.Close()
	go func() {
		for {
			if _, err := sub.Receive(); err != nil {
				if !quit {
					t.Error(err)
				}
			}
			ch <- true
		}
	}()
	time.Sleep(100 * time.Millisecond)
	sub.PSubscribe("news.*")
	time.Sleep(100 * time.Millisecond)
	<-ch
	if len(sub.Patterns) != 1 {
		t.Fail()
	}
	if err := sub.PUnSubscribe("news.*"); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	<-ch
	time.Sleep(100 * time.Millisecond)
	if len(sub.Patterns) != 0 {
		t.Fail()
	}
	quit = true
}
