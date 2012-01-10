package env

import (
    "fmt"
    "os"
)

func Get(key string) string {
    val := os.Getenv(key)
    if val == "" {
        panic(fmt.Sprintf("Missing environment variable %s", key))
    }
    return val
}

func GetDefault(key, def string) string {
    val := os.Getenv(key)
    if val == "" {
        return def
    }
    return val
}
