package user

import (
    "cache"
)

const TTL = 5 * 60 * 1e9 // 5 minutes

func Notify(key string, message string) {
    cache.Set(key, message, TTL)
}
