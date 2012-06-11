package blacklist

import (
    "cache"
    "hashie"
)

const TTL = 24 * 60 * 60 // 1 day

func key(thing string) string {
    return hashie.Sha1([]byte(thing), []byte(":blacklisted"))
}

func IsBlacklisted(thing string) bool {
    _, err := cache.Get(key(thing))
    return err == nil
}

func Blacklist(thing string) {
    cache.Set(key(thing), "blacklisted", TTL)
}
