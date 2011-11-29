package template

import (
    "bytes"
    T "text/template"
    "util"
)

func RenderToString(name, data string, context interface{}) string {
    var buffer bytes.Buffer
    tmpl := T.Must(T.New(name).Parse(data))
    util.Must(tmpl.Execute(&buffer, context))
    return buffer.String()
}
