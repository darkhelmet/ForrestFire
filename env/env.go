package env

import (
    "fmt"
    "os"
)

func Get(key string) string {
    val := os.Getenv(key)
    if val == "" {
        fmt.Println("Missing", key)
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
