package web

import (
    "bytes"
    "encoding/base64"
    "io/ioutil"
    "strconv"
    "strings"
    "mime"
    "time"
    "fmt"
    "crypto/hmac"
    "crypto/sha1"
)

type Context struct {
    *Request
    *Server
    conn
    responseStarted bool
}

func (ctx *Context) StartResponse(status int) {
    ctx.conn.StartResponse(status)
    ctx.responseStarted = true
}

func (ctx *Context) Write(data []byte) (n int, err error) {
    if !ctx.responseStarted {
        ctx.StartResponse(200)
    }

    //if it's a HEAD request, we just write blank data
    if ctx.Request.Method == "HEAD" {
        data = []byte{}
    }

    return ctx.conn.Write(data)
}
func (ctx *Context) WriteString(content string) {
    ctx.Write([]byte(content))
}

func (ctx *Context) Abort(status int, body string) {
    ctx.StartResponse(status)
    ctx.WriteString(body)
}

func (ctx *Context) Redirect(status int, url_ string) {
    ctx.SetHeader("Location", url_, true)
    ctx.StartResponse(status)
    ctx.WriteString("Redirecting to: " + url_)
}

func (ctx *Context) NotModified() {
    ctx.StartResponse(304)
}

func (ctx *Context) NotFound(message string) {
    ctx.StartResponse(404)
    ctx.WriteString(message)
}

//Sets the content type by extension, as defined in the mime package.
//For example, ctx.ContentType("json") sets the content-type to "application/json"
func (ctx *Context) ContentType(ext string) {
    if !strings.HasPrefix(ext, ".") {
        ext = "." + ext
    }
    ctype := mime.TypeByExtension(ext)
    if ctype != "" {
        ctx.SetHeader("Content-Type", ctype, true)
    }
}

//Sets a cookie -- duration is the amount of time in seconds. 0 = forever
func (ctx *Context) SetCookie(name string, value string, age int64) {
    var utctime time.Time
    if age == 0 {
        // 2^31 - 1 seconds (roughly 2038)
        utctime = time.Unix(2147483647, 0).UTC()
    } else {
        utctime = time.Unix(int64(time.Now().UTC().Second())+age, 0).UTC()
    }
    cookie := fmt.Sprintf("%s=%s; expires=%s", name, value, webTime(utctime))
    ctx.SetHeader("Set-Cookie", cookie, false)
}

func getCookieSig(key string, val []byte, timestamp string) string {
    hm := hmac.New(sha1.New, []byte(key))

    hm.Write(val)
    hm.Write([]byte(timestamp))

    hex := fmt.Sprintf("%02x", hm.Sum(nil))
    return hex
}

func (ctx *Context) SetSecureCookie(name string, val string, age int64) {
    //base64 encode the val
    if len(ctx.Server.Config.CookieSecret) == 0 {
        ctx.Logger.Println("Secret Key for secure cookies has not been set. Please assign a cookie secret to web.Config.CookieSecret.")
        return
    }
    var buf bytes.Buffer
    encoder := base64.NewEncoder(base64.StdEncoding, &buf)
    encoder.Write([]byte(val))
    encoder.Close()
    vs := buf.String()
    vb := buf.Bytes()
    timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
    sig := getCookieSig(ctx.Server.Config.CookieSecret, vb, timestamp)
    cookie := strings.Join([]string{vs, timestamp, sig}, "|")
    ctx.SetCookie(name, cookie, age)
}

func (ctx *Context) GetSecureCookie(name string) (string, bool) {
    for _, cookie := range ctx.Request.Cookie {
        if cookie.Name != name {
            continue
        }

        parts := strings.SplitN(cookie.Value, "|", 3)

        val := parts[0]
        timestamp := parts[1]
        sig := parts[2]

        if getCookieSig(ctx.Server.Config.CookieSecret, []byte(val), timestamp) != sig {
            return "", false
        }

        ts, _ := strconv.ParseInt(timestamp, 10, 64)

        if time.Now().Sub(time.Unix(0, ts)) > time.Duration(31*86400) {
            return "", false
        }

        buf := bytes.NewBufferString(val)
        encoder := base64.NewDecoder(base64.StdEncoding, buf)

        res, _ := ioutil.ReadAll(encoder)
        return string(res), true
    }
    return "", false
}
