package simpletransport

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestHTTP(t *testing.T) {
	transport := &SimpleTransport{}
	req, _ := http.NewRequest("GET", "http://jsonip.com", nil)
	res, _ := transport.RoundTrip(req)
	body, _ := ioutil.ReadAll(res.Body)

	res.Body.Close()

	fmt.Println(string(body))
}

func TestHTTPS(t *testing.T) {
	transport := &SimpleTransport{}
	req, _ := http.NewRequest("GET", "https://google.com", nil)
	res, _ := transport.RoundTrip(req)
	body, _ := ioutil.ReadAll(res.Body)

	res.Body.Close()

	fmt.Println(string(body))
}
