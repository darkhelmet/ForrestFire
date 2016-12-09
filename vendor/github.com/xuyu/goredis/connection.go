package goredis

// Returns message.
func (r *Redis) Echo(message string) (string, error) {
	rp, err := r.ExecuteCommand("ECHO", message)
	if err != nil {
		return "", err
	}
	return rp.StringValue()
}

// Returns PONG. This command is often used to test if a connection is still alive, or to measure latency.
func (r *Redis) Ping() error {
	_, err := r.ExecuteCommand("PING")
	return err
}
