package env

import (
    "fmt"
    "os"
    "strconv"
)

// Get a string from the environment,
// panicking if it's not there
func String(key string) string {
    val := os.Getenv(key)
    if val == "" {
        panic(fmt.Errorf("env: Environment variable %s doesn't exist", key))
    }
    return val
}

// Get a string from the environment,
// returning the default if it's not present
func StringDefault(key, def string) (val string) {
    defer func() {
        if recover() != nil {
            val = def
        }
    }()
    val = String(key)
    return
}

// Get a string from the environment,
// returning the result of the default function if it's not present
func StringDefaultF(key string, def func() string) (val string) {
    defer func() {
        if recover() != nil {
            val = def()
        }
    }()
    val = String(key)
    return
}

// Get an int from the environment,
// panicking if it's not present or doesn't parse properly
func Int(key string) int {
    i, err := strconv.ParseInt(String(key), 10, 32)
    if err != nil {
        panic(fmt.Errorf("env: failed parsing int: %s", err))
    }
    return int(i)
}

// Get an int from the environment,
// returning the default if it's not present or doesn't parse properly
func IntDefault(key string, def int) (val int) {
    defer func() {
        if recover() != nil {
            val = def
        }
    }()
    val = Int(key)
    return
}

// Get an int from the environment,
// returning the result of the default function if it's not present or doesn't parse properly
func IntDefaultF(key string, def func() int) (val int) {
    defer func() {
        if recover() != nil {
            val = def()
        }
    }()
    val = Int(key)
    return
}

// Get a float from the environment,
// panicking if it's not present or doesn't parse properly
func Float(key string) float64 {
    val, err := strconv.ParseFloat(String(key), 64)
    if err != nil {
        panic(fmt.Errorf("env: failed parsing float: %s", err))
    }
    return val
}

// Get a float from the environment,
// returning the default if it's not present or doesn't parse properly
func FloatDefault(key string, def float64) (val float64) {
    defer func() {
        if recover() != nil {
            val = def
        }
    }()
    val = Float(key)
    return
}

// Get a float from the environment,
// returning the result of the default function if it's not present or doesn't parse properly
func FloatDefaultF(key string, def func() float64) (val float64) {
    defer func() {
        if recover() != nil {
            val = def()
        }
    }()
    val = Float(key)
    return
}
