// Redis Golang Client with full features
//
// Protocol Specification: http://redis.io/topics/protocol.
//
// Redis reply has five types: status, error, integer, bulk, multi bulk.
// A Status Reply is in the form of a single line string starting with "+" terminated by "\r\n".
// Error Replies are very similar to Status Replies. The only difference is that the first byte is "-".
// Integer reply is just a CRLF terminated string representing an integer, prefixed by a ":" byte.
// Bulk replies are used by the server in order to return a single binary safe string up to 512 MB in length.
// A Multi bulk reply is used to return an array of other replies.
// Every element of a Multi Bulk Reply can be of any kind, including a nested Multi Bulk Reply.
// So five reply type is defined:
//  const (
//  	ErrorReply = iota
//  	StatusReply
//  	IntegerReply
//  	BulkReply
//  	MultiReply
//  )
// And then a Reply struct which represent the redis response data is defined:
//  type Reply struct {
//  	Type    int
//  	Error   string
//  	Status  string
//  	Integer int64  // Support Redis 64bit integer
//  	Bulk    []byte // Support Redis Null Bulk Reply
//  	Multi   []*Reply
//  }
// Reply struct has many useful methods:
//  func (rp *Reply) IntegerValue() (int64, error)
//  func (rp *Reply) BoolValue() (bool, error)
//  func (rp *Reply) StatusValue() (string, error)
//  func (rp *Reply) OKValue() error
//  func (rp *Reply) BytesValue() ([]byte, error)
//  func (rp *Reply) StringValue() (string, error)
//  func (rp *Reply) MultiValue() ([]*Reply, error)
//  func (rp *Reply) HashValue() (map[string]string, error)
//  func (rp *Reply) ListValue() ([]string, error)
//  func (rp *Reply) BytesArrayValue() ([][]byte, error)
//  func (rp *Reply) BoolArrayValue() ([]bool, error)
//
// Connect redis has two function: DialTimeout and DialURL, for example:
//  client, err := Dial()
//  client, err := Dial(&DialConfig{Address: "127.0.0.1:6379"})
//  client, err := DialTimeout("tcp", "127.0.0.1:6379", 0, "", 10*time.Second, 10)
//  client, err := DialURL("tcp://auth:password@127.0.0.1:6379/0?timeout=10s&maxidle=1")
//
// Try a redis command is simple too, let's do GET/SET:
//  err := client.Set("key", "value", 0, 0, false, false)
//  value, err := client.Get("key")
//
// Or you can execute customer command with Redis.ExecuteCommand method:
//  reply, err := client.ExecuteCommand("SET", "key", "value")
//  err := reply.OKValue()
//
// Redis Pipelining is defined as:
//  type Pipelined struct {
//  	redis *Redis
//  	conn  *Connection
//  	times int
//  }
//  func (p *Pipelined) Close()
//  func (p *Pipelined) Command(args ...interface{})
//  func (p *Pipelined) Receive() (*Reply, error)
//  func (p *Pipelined) ReceiveAll() ([]*Reply, error)
//
// Transaction, Lua Eval, Publish/Subscribe, Monitor, Scan, Sort are also supported.
//
package goredis

import (
	"bufio"
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bufferPool struct {
	bufs  []*bytes.Buffer
	mutex sync.Mutex
}

func (b *bufferPool) GetBuffer() *bytes.Buffer {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if len(b.bufs) > 0 {
		buf := b.bufs[0]
		b.bufs[0] = nil
		b.bufs = b.bufs[1:]
		return buf
	}
	return bytes.NewBuffer(nil)
}

func (b *bufferPool) PutBuffer(buf *bytes.Buffer) {
	b.mutex.Lock()
	if len(b.bufs) < math.MaxInt16 {
		buf.Reset()
		b.bufs = append(b.bufs, buf)
	}
	b.mutex.Unlock()
}

var buffers = &bufferPool{}

func packArgs(items ...interface{}) (args []interface{}) {
	for _, item := range items {
		v := reflect.ValueOf(item)
		switch v.Kind() {
		case reflect.Slice:
			if v.IsNil() {
				continue
			}
			for i := 0; i < v.Len(); i++ {
				args = append(args, v.Index(i).Interface())
			}
		case reflect.Map:
			if v.IsNil() {
				continue
			}
			for _, key := range v.MapKeys() {
				args = append(args, key.Interface(), v.MapIndex(key).Interface())
			}
		case reflect.String:
			if v.String() != "" {
				args = append(args, v.Interface())
			}
		default:
			args = append(args, v.Interface())
		}
	}
	return args
}

func packCommand(args ...interface{}) ([]byte, error) {
	buf := buffers.GetBuffer()
	defer buffers.PutBuffer(buf)
	if _, err := fmt.Fprintf(buf, "*%d\r\n", len(args)); err != nil {
		return nil, err
	}
	var s string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			s = v
		case int:
			s = strconv.Itoa(v)
		case int64:
			s = strconv.FormatInt(v, 10)
		case uint64:
			s = strconv.FormatUint(v, 10)
		case float64:
			s = strconv.FormatFloat(v, 'g', -1, 64)
		default:
			return nil, errors.New("invalid argument type when pack command")
		}
		if _, err := fmt.Fprintf(buf, "$%d\r\n%s\r\n", len(s), s); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

type Connection struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

func (c *Connection) SendCommand(args ...interface{}) error {
	request, err := packCommand(args...)
	if err != nil {
		return err
	}
	if _, err := c.Conn.Write(request); err != nil {
		return err
	}
	return nil
}

func (c *Connection) RecvReply() (*Reply, error) {
	line, err := c.Reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	line = line[:len(line)-2]
	switch line[0] {
	case '-':
		return &Reply{
			Type:  ErrorReply,
			Error: string(line[1:]),
		}, nil
	case '+':
		return &Reply{
			Type:   StatusReply,
			Status: string(line[1:]),
		}, nil
	case ':':
		i, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Reply{
			Type:    IntegerReply,
			Integer: i,
		}, nil
	case '$':
		size, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, err
		}
		bulk, err := c.ReadBulk(size)
		if err != nil {
			return nil, err
		}
		return &Reply{
			Type: BulkReply,
			Bulk: bulk,
		}, nil
	case '*':
		i, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, err
		}
		rp := &Reply{Type: MultiReply}
		if i >= 0 {
			multi := make([]*Reply, i)
			for j := 0; j < i; j++ {
				rp, err := c.RecvReply()
				if err != nil {
					return nil, err
				}
				multi[j] = rp
			}
			rp.Multi = multi
		}
		return rp, nil
	}
	return nil, errors.New("redis protocol error")
}

func (c *Connection) ReadBulk(size int) ([]byte, error) {
	// If the requested value does not exist the bulk reply will use the special value -1 as data length
	if size < 0 {
		return nil, nil
	}
	buf := make([]byte, size+2)
	if _, err := io.ReadFull(c.Reader, buf); err != nil {
		return nil, err
	}
	return buf[:size], nil
}

type ConnPool struct {
	MaxIdle int
	Dial    func() (*Connection, error)

	idle   *list.List
	active int
	closed bool
	mutex  sync.Mutex
}

func NewConnPool(maxidle int, dial func() (*Connection, error)) *ConnPool {
	return &ConnPool{
		MaxIdle: maxidle,
		Dial:    dial,
		idle:    list.New(),
	}
}

func (p *ConnPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.closed = true
	for e := p.idle.Front(); e != nil; e = e.Next() {
		e.Value.(*Connection).Conn.Close()
	}
}

func (p *ConnPool) Get() (*Connection, error) {
	p.mutex.Lock()
	p.active++
	if p.closed {
		return nil, errors.New("connection pool closed")
	}
	if p.idle.Len() > 0 {
		back := p.idle.Back()
		p.idle.Remove(back)
		p.mutex.Unlock()
		return back.Value.(*Connection), nil
	}
	p.mutex.Unlock()
	return p.Dial()
}

func (p *ConnPool) Put(c *Connection) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.active--
	if p.closed {
		c.Conn.Close()
		return
	}
	if c == nil {
		return
	}
	if p.idle.Len() >= p.MaxIdle {
		p.idle.Remove(p.idle.Front())
	}
	p.idle.PushBack(c)
}

type Redis struct {
	network  string
	address  string
	db       int
	password string
	timeout  time.Duration
	pool     *ConnPool
}

func (r *Redis) ExecuteCommand(args ...interface{}) (*Reply, error) {
	c, err := r.pool.Get()
	defer r.pool.Put(c)
	if err != nil {
		return nil, err
	}
	if err := c.SendCommand(args...); err != nil {
		if err != io.EOF {
			return nil, err
		}
		c, err = r.pool.Get()
		if err != nil {
			return nil, err
		}
		if err = c.SendCommand(args...); err != nil {
			return nil, err
		}
	}
	rp, err := c.RecvReply()
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
		c, err = r.pool.Get()
		if err != nil {
			return nil, err
		}
		if err = c.SendCommand(args...); err != nil {
			return nil, err
		}
		return c.RecvReply()
	}
	return rp, err
}

func (r *Redis) NewConnection() (*Connection, error) {
	conn, err := net.DialTimeout(r.network, r.address, r.timeout)
	if err != nil {
		return nil, err
	}
	c := &Connection{conn, bufio.NewReader(conn)}
	if r.password != "" {
		if err := c.SendCommand("AUTH", r.password); err != nil {
			return nil, err
		}
		rp, err := c.RecvReply()
		if err != nil {
			return nil, err
		}
		if rp.Type == ErrorReply {
			return nil, errors.New(rp.Error)
		}
	}
	if r.db > 0 {
		if err := c.SendCommand("SELECT", r.db); err != nil {
			return nil, err
		}
		rp, err := c.RecvReply()
		if err != nil {
			return nil, err
		}
		if rp.Type == ErrorReply {
			return nil, errors.New(rp.Error)
		}
	}
	return c, nil
}

func (r *Redis) ClosePool() {
	r.pool.Close()
}

const (
	DefaultNetwork = "tcp"
	DefaultAddress = ":6379"
	DefaultTimeout = 15 * time.Second
	DefaultMaxIdle = 1
)

type DialConfig struct {
	Network  string
	Address  string
	Database int
	Password string
	Timeout  time.Duration
	MaxIdle  int
}

func Dial(cfg *DialConfig) (*Redis, error) {
	if cfg == nil {
		cfg = &DialConfig{}
	}
	if cfg.Network == "" {
		cfg.Network = DefaultNetwork
	}
	if cfg.Address == "" {
		cfg.Address = DefaultAddress
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}
	if cfg.MaxIdle == 0 {
		cfg.MaxIdle = DefaultMaxIdle
	}
	return DialTimeout(cfg.Network, cfg.Address, cfg.Database, cfg.Password, cfg.Timeout, cfg.MaxIdle)
}

func DialTimeout(network, address string, db int, password string, timeout time.Duration, maxidle int) (*Redis, error) {
	r := &Redis{
		network:  network,
		address:  address,
		db:       db,
		password: password,
		timeout:  timeout,
	}
	r.pool = NewConnPool(maxidle, r.NewConnection)
	c, err := r.NewConnection()
	if err != nil {
		return nil, err
	}
	r.pool.Put(c)
	return r, nil
}

func DialURL(rawurl string) (*Redis, error) {
	ul, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	password := ""
	if ul.User != nil {
		if pw, set := ul.User.Password(); set {
			password = pw
		}
	}
	db, err := strconv.Atoi(strings.Trim(ul.Path, "/"))
	if err != nil {
		return nil, err
	}
	timeout, err := time.ParseDuration(ul.Query().Get("timeout"))
	if err != nil {
		return nil, err
	}
	maxidle, err := strconv.Atoi(ul.Query().Get("maxidle"))
	if err != nil {
		return nil, err
	}
	return DialTimeout(ul.Scheme, ul.Host, db, password, timeout, maxidle)
}

// Reply Type: Status, Integer, Bulk, Multi Bulk
// Error Reply Type return error directly
const (
	ErrorReply = iota
	StatusReply
	IntegerReply
	BulkReply
	MultiReply
)

// Represent Redis Reply
type Reply struct {
	Type    int
	Error   string
	Status  string
	Integer int64  // Support Redis 64bit integer
	Bulk    []byte // Support Redis Null Bulk Reply
	Multi   []*Reply
}

func (rp *Reply) IntegerValue() (int64, error) {
	if rp.Type == ErrorReply {
		return 0, errors.New(rp.Error)
	}
	if rp.Type != IntegerReply {
		return 0, errors.New("invalid reply type, not integer")
	}
	return rp.Integer, nil
}

// Integer replies are also extensively used in order to return true or false.
// For instance commands like EXISTS or SISMEMBER will return 1 for true and 0 for false.
func (rp *Reply) BoolValue() (bool, error) {
	if rp.Type == ErrorReply {
		return false, errors.New(rp.Error)
	}
	if rp.Type != IntegerReply {
		return false, errors.New("invalid reply type, not integer")
	}
	return rp.Integer != 0, nil
}

func (rp *Reply) StatusValue() (string, error) {
	if rp.Type == ErrorReply {
		return "", errors.New(rp.Error)
	}
	if rp.Type != StatusReply {
		return "", errors.New("invalid reply type, not status")
	}
	return rp.Status, nil
}

func (rp *Reply) OKValue() error {
	if rp.Type == ErrorReply {
		return errors.New(rp.Error)
	}
	if rp.Type != StatusReply {
		return errors.New("invalid reply type, not status")
	}
	if rp.Status == "OK" {
		return nil
	}
	return errors.New(rp.Status)
}

func (rp *Reply) BytesValue() ([]byte, error) {
	if rp.Type == ErrorReply {
		return nil, errors.New(rp.Error)
	}
	if rp.Type != BulkReply {
		return nil, errors.New("invalid reply type, not bulk")
	}
	return rp.Bulk, nil
}

func (rp *Reply) StringValue() (string, error) {
	if rp.Type == ErrorReply {
		return "", errors.New(rp.Error)
	}
	if rp.Type != BulkReply {
		return "", errors.New("invalid reply type, not bulk")
	}
	if rp.Bulk == nil {
		return "", nil
	}
	return string(rp.Bulk), nil
}

func (rp *Reply) MultiValue() ([]*Reply, error) {
	if rp.Type == ErrorReply {
		return nil, errors.New(rp.Error)
	}
	if rp.Type != MultiReply {
		return nil, errors.New("invalid reply type, not multi bulk")
	}
	return rp.Multi, nil
}

func (rp *Reply) HashValue() (map[string]string, error) {
	if rp.Type == ErrorReply {
		return nil, errors.New(rp.Error)
	}
	if rp.Type != MultiReply {
		return nil, errors.New("invalid reply type, not multi bulk")
	}
	result := make(map[string]string)
	if rp.Multi != nil {
		length := len(rp.Multi)
		for i := 0; i < length/2; i++ {
			key, err := rp.Multi[i*2].StringValue()
			if err != nil {
				return nil, err
			}
			value, err := rp.Multi[i*2+1].StringValue()
			if err != nil {
				return nil, err
			}
			result[key] = value
		}
	}
	return result, nil
}

func (rp *Reply) ListValue() ([]string, error) {
	if rp.Type == ErrorReply {
		return nil, errors.New(rp.Error)
	}
	if rp.Type != MultiReply {
		return nil, errors.New("invalid reply type, not multi bulk")
	}
	var result []string
	if rp.Multi != nil {
		for _, subrp := range rp.Multi {
			item, err := subrp.StringValue()
			if err != nil {
				return nil, err
			}
			result = append(result, item)
		}
	}
	return result, nil
}

func (rp *Reply) BytesArrayValue() ([][]byte, error) {
	if rp.Type == ErrorReply {
		return nil, errors.New(rp.Error)
	}
	if rp.Type != MultiReply {
		return nil, errors.New("invalid reply type, not multi bulk")
	}
	var result [][]byte
	if rp.Multi != nil {
		for _, subrp := range rp.Multi {
			b, err := subrp.BytesValue()
			if err != nil {
				return nil, err
			}
			result = append(result, b)
		}
	}
	return result, nil
}

func (rp *Reply) BoolArrayValue() ([]bool, error) {
	if rp.Type == ErrorReply {
		return nil, errors.New(rp.Error)
	}
	if rp.Type != MultiReply {
		return nil, errors.New("invalid reply type, not multi bulk")
	}
	var result []bool
	if rp.Multi != nil {
		for _, subrp := range rp.Multi {
			b, err := subrp.BoolValue()
			if err != nil {
				return nil, err
			}
			result = append(result, b)
		}
	}
	return result, nil
}
