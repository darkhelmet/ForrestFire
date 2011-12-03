package blacklist

import (
    "cache"
    "hashie"
    "net/url"
)

const TTL = 24 * 60 * 60 // 1 day

func IsBlacklisted(uri *url.URL) bool {
    if _, err := cache.Get(hashie.Sha1([]byte(uri.String()))); err == nil {
        return true
    }
    return false
}

func Blacklist(uri *url.URL) {
    cache.Set(hashie.Sha1([]byte(uri.String())), "blacklisted", TTL)
}
