package web

import (
    "log"
    "net"
    "net/http"
    "reflect"
    "regexp"
    "runtime"
    "time"
    "strconv"
    "path"
    "os"
    "net/http/pprof"
)

type Server struct {
    Config *ServerConfig
    routes []route
    Logger *log.Logger
    //save the listener so it can be closed
    l      net.Listener
    closed bool
}

type route struct {
    r       string
    cr      *regexp.Regexp
    method  string
    handler reflect.Value
}

func (s *Server) addRoute(r string, method string, handler interface{}) {
    cr, err := regexp.Compile(r)
    if err != nil {
        s.Logger.Printf("Error in route regex %q\n", r)
        return
    }

    if fv, ok := handler.(reflect.Value); ok {
        s.routes = append(s.routes, route{r, cr, method, fv})
    } else {
        fv := reflect.ValueOf(handler)
        s.routes = append(s.routes, route{r, cr, method, fv})
    }
}

func (s *Server) ServeHTTP(c http.ResponseWriter, req *http.Request) {
    conn := httpConn{c}
    wreq := newRequest(req, c)
    s.routeHandler(wreq, &conn)
}

//Calls a function with recover block
func (s *Server) safelyCall(function reflect.Value, args []reflect.Value) (resp []reflect.Value, e interface{}) {
    defer func() {
        if err := recover(); err != nil {
            if !s.Config.RecoverPanic {
                // go back to panic
                panic(err)
            } else {
                e = err
                resp = nil
                s.Logger.Println("Handler crashed with error", err)
                for i := 1; ; i += 1 {
                    _, file, line, ok := runtime.Caller(i)
                    if !ok {
                        break
                    }
                    s.Logger.Println(file, line)
                }
            }
        }
    }()
    return function.Call(args), nil
}



//should the context be passed to the handler?
func requiresContext(handlerType reflect.Type) bool {
    //if the method doesn't take arguments, no
    if handlerType.NumIn() == 0 {
        return false
    }

    //if the first argument is not a pointer, no
    a0 := handlerType.In(0)
    if a0.Kind() != reflect.Ptr {
        return false
    }
    //if the first argument is a context, yes
    if a0.Elem() == contextType {
        return true
    }

    return false
}

func (s *Server) routeHandler(req *Request, c conn) {
    requestPath := req.URL.Path

    //log the request
    if len(req.URL.RawQuery) == 0 {
        s.Logger.Println(req.Method + " " + requestPath)
    } else {
        s.Logger.Println(req.Method + " " + requestPath + "?" + req.URL.RawQuery)
    }

    //parse the form data (if it exists)
    perr := req.parseParams()
    if perr != nil {
        s.Logger.Printf("Failed to parse form data %q\n", perr.Error())
    }

    ctx := Context{req, s, c, false}

    //set some default headers
    ctx.SetHeader("Content-Type", "text/html; charset=utf-8", true)
    ctx.SetHeader("Server", "web.go", true)

    tm := time.Now().UTC()
    ctx.SetHeader("Date", webTime(tm), true)

    //try to serve a static file
    staticDir := s.Config.StaticDir
    if staticDir == "" {
        staticDir = defaultStaticDir()
    }
    staticFile := path.Join(staticDir, requestPath)
    if fileExists(staticFile) && (req.Method == "GET" || req.Method == "HEAD") {
        serveFile(&ctx, staticFile)
        return
    }

    for i := 0; i < len(s.routes); i++ {
        route := s.routes[i]
        cr := route.cr
        //if the methods don't match, skip this handler (except HEAD can be used in place of GET)
        if req.Method != route.method && !(req.Method == "HEAD" && route.method == "GET") {
            continue
        }

        if !cr.MatchString(requestPath) {
            continue
        }
        match := cr.FindStringSubmatch(requestPath)

        if len(match[0]) != len(requestPath) {
            continue
        }

        var args []reflect.Value
        handlerType := route.handler.Type()
        if requiresContext(handlerType) {
            args = append(args, reflect.ValueOf(&ctx))
        }
        for _, arg := range match[1:] {
            args = append(args, reflect.ValueOf(arg))
        }

        ret, err := s.safelyCall(route.handler, args)
        if err != nil {
            //fmt.Printf("%v\n", err)
            //there was an error or panic while calling the handler
            ctx.Abort(500, "Server Error")
        }

        if len(ret) == 0 {
            return
        }

        sval := ret[0]

        if sval.Kind() == reflect.String &&
            !ctx.responseStarted {
            content := []byte(sval.String())
            ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
            ctx.StartResponse(200)
            ctx.Write(content)
        }

        return
    }

    //try to serve index.html || index.htm
    if indexPath := path.Join(path.Join(staticDir, requestPath), "index.html"); fileExists(indexPath) {
        serveFile(&ctx, indexPath)
        return
    }

    if indexPath := path.Join(path.Join(staticDir, requestPath), "index.htm"); fileExists(indexPath) {
        serveFile(&ctx, indexPath)
        return
    }

    ctx.Abort(404, "Page not found")
}

func (s *Server) initServer() {
    if s.Config == nil {
        s.Config = &ServerConfig{}
    }

    if s.Logger == nil {
        s.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
    }
}

//Runs the web application and serves http requests
func (s *Server) Run(addr string) {
    s.initServer()

    mux := http.NewServeMux()

    mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
    mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
    mux.Handle("/debug/pprof/heap", http.HandlerFunc(pprof.Heap))
    mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
    mux.Handle("/", s)

    s.Logger.Printf("web.go serving %s\n", addr)

    l, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatal("ListenAndServe:", err)
    }
    s.l = l
    err = http.Serve(s.l, mux)
    s.l.Close()
}

func (s *Server) SetLogger(logger *log.Logger) {
    s.Logger = logger
}

//Adds a handler for the 'GET' http method.
func (s *Server) Get(route string, handler interface{}) {
    s.addRoute(route, "GET", handler)
}

//Adds a handler for the 'POST' http method.
func (s *Server) Post(route string, handler interface{}) {
    s.addRoute(route, "POST", handler)
}

//Adds a handler for the 'PUT' http method.
func (s *Server) Put(route string, handler interface{}) {
    s.addRoute(route, "PUT", handler)
}

//Adds a handler for the 'DELETE' http method.
func (s *Server) Delete(route string, handler interface{}) {
    s.addRoute(route, "DELETE", handler)
}

//Runs the web application and serves scgi requests for this Server object.
func (s *Server) RunFcgi(addr string) {
    s.initServer()
    s.Logger.Printf("web.go serving fcgi %s\n", addr)
    s.listenAndServeFcgi(addr)
}

func (s *Server) RunScgi(addr string) {
    s.initServer()
    s.Logger.Printf("web.go serving scgi %s\n", addr)
    s.listenAndServeScgi(addr)
}

//Stops the web server
func (s *Server) Close() {
    s.l.Close()
    s.closed = true
}

func fileExists(dir string) bool {
    info, err := os.Stat(dir)
    if err != nil {
        return false
    } else if !!info.IsDir() {
        return false
    }

    return true
}
