package extractor

import (
    "net"
    "net/http"
    "time"
)

var closeAfter = 1 * time.Second

func dialer(timeout time.Duration, deadline time.Time) func(string, string) (net.Conn, error) {
    return func(netw, addr string) (net.Conn, error) {
        conn, err := net.DialTimeout(netw, addr, timeout)
        if err != nil {
            return nil, err
        }
        if err := conn.SetDeadline(deadline); err != nil {
            return nil, err
        }
        go func() {
            // Explicitly close the connection about 1 after the deadline
            <-time.After(deadline.Sub(time.Now()) + closeAfter)
            conn.Close()
        }()
        return conn, nil
    }
}

func newTimeoutDeadlineDialer(timeout time.Duration, deadline time.Time) *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            Dial: dialer(timeout, deadline),
        },
    }
}
