package goredis

import (
	"errors"
)

// Document: http://redis.io/topics/transactions
// MULTI, EXEC, DISCARD and WATCH are the foundation of transactions in Redis.
// A Redis script is transactional by definition,
// so everything you can do with a Redis transaction, you can also do with a script,
// and usually the script will be both simpler and faster.
type Transaction struct {
	redis *Redis
	conn  *Connection
}

func (r *Redis) Transaction() (*Transaction, error) {
	c, err := r.pool.Get()
	if err != nil {
		return nil, err
	}
	if err := c.SendCommand("MULTI"); err != nil {
		r.pool.Put(c)
		return nil, err
	}
	if _, err := c.RecvReply(); err != nil {
		r.pool.Put(c)
		return nil, err
	}
	return &Transaction{r, c}, nil
}

func (t *Transaction) Close() {
	t.redis.pool.Put(t.conn)
}

// Flushes all previously queued commands in a transaction and restores the connection state to normal.
// If WATCH was used, DISCARD unwatches all keys.
func (t *Transaction) Discard() error {
	if err := t.conn.SendCommand("DISCARD"); err != nil {
		return err
	}
	_, err := t.conn.RecvReply()
	return err
}

// Marks the given keys to be watched for conditional execution of a transaction.
func (t *Transaction) Watch(keys ...string) error {
	args := packArgs("WATCH", keys)
	if err := t.conn.SendCommand(args...); err != nil {
		return err
	}
	_, err := t.conn.RecvReply()
	return err
}

// Flushes all the previously watched keys for a transaction.
// If you call EXEC or DISCARD, there's no need to manually call UNWATCH.
func (t *Transaction) UnWatch() error {
	if err := t.conn.SendCommand("UNWATCH"); err != nil {
		return err
	}
	_, err := t.conn.RecvReply()
	return err
}

// Executes all previously queued commands in a transaction and restores the connection state to normal.
// When using WATCH, EXEC will execute commands only if the watched keys were not modified,
// allowing for a check-and-set mechanism.
func (t *Transaction) Exec() ([]*Reply, error) {
	if err := t.conn.SendCommand("EXEC"); err != nil {
		return nil, err
	}
	rp, err := t.conn.RecvReply()
	if err != nil {
		return nil, err
	}
	return rp.MultiValue()
}

func (t *Transaction) Command(args ...interface{}) error {
	args2 := packArgs(args...)
	if err := t.conn.SendCommand(args2...); err != nil {
		return err
	}
	rp, err := t.conn.RecvReply()
	if err != nil {
		return err
	}
	s, err := rp.StatusValue()
	if err != nil {
		return err
	}
	if s != "QUEUED" {
		return errors.New(s)
	}
	return nil
}
