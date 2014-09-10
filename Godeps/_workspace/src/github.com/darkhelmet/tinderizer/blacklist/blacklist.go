package blacklist

import (
    "github.com/darkhelmet/tinderizer/cache"
    "github.com/darkhelmet/tinderizer/hashie"
)

const (
    TTL   = 24 * 60 * 60 // 1 day
    Value = "blacklisted"
)

func key(thing string) string {
    return hashie.Sha1([]byte(thing), []byte(":blacklisted"))
}

func IsBlacklisted(thing string) bool {
    value, _ := cache.Get(key(thing))
    return value == Value
}

func Blacklist(thing string) {
    cache.Set(key(thing), Value, TTL)
}
