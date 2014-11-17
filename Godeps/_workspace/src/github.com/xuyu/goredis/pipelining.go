package goredis

// A Request/Response server can be implemented so that it is able to process new requests
// even if the client didn't already read the old responses.
// This way it is possible to send multiple commands to the server without waiting for the replies at all,
// and finally read the replies in a single step.
type Pipelined struct {
	redis *Redis
	conn  *Connection
	times int
}

func (r *Redis) Pipelining() (*Pipelined, error) {
	c, err := r.pool.Get()
	if err != nil {
		return nil, err
	}
	return &Pipelined{r, c, 0}, nil
}

func (p *Pipelined) Close() {
	p.redis.pool.Put(p.conn)
	p.times = 0
}

func (p *Pipelined) Command(args ...interface{}) error {
	err := p.conn.SendCommand(args...)
	if err == nil {
		p.times++
	}
	return err
}

func (p *Pipelined) Receive() (*Reply, error) {
	rp, err := p.conn.RecvReply()
	if err == nil {
		p.times--
	}
	return rp, err
}

func (p *Pipelined) ReceiveAll() ([]*Reply, error) {
	if p.times <= 0 {
		return nil, nil
	}
	rps := make([]*Reply, p.times)
	num := p.times
	for i := 0; i < num; i++ {
		rp, err := p.Receive()
		if err != nil {
			return rps, err
		}
		rps[i] = rp
	}
	return rps, nil
}
