package web

import (
	"bytes"
	"log"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

type conn interface {
	StartResponse(status int)
	SetHeader(hdr string, val string, unique bool)
	Write(data []byte) (n int, err error)
	Close()
}

// small optimization: cache the context type instead of repeteadly calling reflect.Typeof
var contextType reflect.Type

var exeFile string

// default
func defaultStaticDir() string {
	root, _ := path.Split(exeFile)
	return path.Join(root, "static")
}

func init() {
	contextType = reflect.TypeOf(Context{})
	//find the location of the exe file
	arg0 := path.Clean(os.Args[0])
	wd, _ := os.Getwd()
	if strings.HasPrefix(arg0, "/") {
		exeFile = arg0
	} else {
		//TODO for robustness, search each directory in $PATH
		exeFile = path.Join(wd, arg0)
	}
}

var Config = &ServerConfig{
	RecoverPanic: true,
}

var mainServer = Server{
	Config: Config,
	Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
}

//Runs the web application and serves http requests
func Run(addr string) {
	mainServer.Run(addr)
}

//Stops the web server
func Close() {
	mainServer.Close()
}

//Runs the web application and serves scgi requests
func RunScgi(addr string) {
	mainServer.RunScgi(addr)
}

//Runs the web application by serving fastcgi requests
func RunFcgi(addr string) {
	mainServer.RunFcgi(addr)
}

//Adds a handler for the 'GET' http method.
func Get(route string, handler interface{}) {
	mainServer.Get(route, handler)
}

//Adds a handler for the 'POST' http method.
func Post(route string, handler interface{}) {
	mainServer.addRoute(route, "POST", handler)
}

//Adds a handler for the 'PUT' http method.
func Put(route string, handler interface{}) {
	mainServer.addRoute(route, "PUT", handler)
}

//Adds a handler for the 'DELETE' http method.
func Delete(route string, handler interface{}) {
	mainServer.addRoute(route, "DELETE", handler)
}

func SetLogger(logger *log.Logger) {
	mainServer.Logger = logger
}

type ServerConfig struct {
	StaticDir    string
	Addr         string
	Port         int
	CookieSecret string
	RecoverPanic bool
}

func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

func Urlencode(data map[string]string) string {
	var buf bytes.Buffer
	for k, v := range data {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
		buf.WriteByte('&')
	}
	s := buf.String()
	return s[0 : len(s)-1]
}
