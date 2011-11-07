package ruby

import (
    "exec"
    "fmt"
)

var bundle string

func init() {
    var err error
    bundle, err = exec.LookPath("bundle")
    if err != nil {
        panic("ruby/bundle not found")
    }
}

func makeArgs(script string, requires []string) []string {
    args := []string{bundle, "exec", "ruby", "-Eutf-8:utf-8"}
    for _, require := range requires {
        args = append(args, fmt.Sprintf("-r%s", require))
    }
    args = append(args, "-e", script)
    return args
}

func Run(script string, requires []string) ([]byte, error) {
    args := makeArgs(script, requires)
    cmd := exec.Command("ruby", args...)
    return cmd.Output()
}

func RunWithInput(script, stdin string, requires []string) ([]byte, error) {
    args := makeArgs(script, requires)
    cmd := exec.Command("ruby", args...)
    in, err := cmd.StdinPipe()
    if err != nil {
        return nil, err
    }
    if _, err = in.Write([]byte(stdin)); err != nil {
        in.Close()
        return nil, err
    }
    if err = in.Close(); err != nil {
        return nil, err
    }
    return cmd.Output()
}
