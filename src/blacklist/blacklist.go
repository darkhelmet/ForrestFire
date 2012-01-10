package blacklist

import (
    "cache"
    "hashie"
)

const TTL = 24 * 60 * 60 // 1 day

func IsBlacklisted(thing string) bool {
    if _, err := cache.Get(hashie.Sha1([]byte(thing))); err == nil {
        return true
    }
    return false
}

func Blacklist(thing string) {
    cache.Set(hashie.Sha1([]byte(thing)), "blacklisted", TTL)
}
