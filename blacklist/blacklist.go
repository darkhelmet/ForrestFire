package blacklist

import (
    "cache"
    "crypto/sha1"
    "fmt"
    "net/url"
)

const TTL = 24 * 60 * 60 // 1 day

func hash(uri *url.URL) string {
    hash := sha1.New()
    return fmt.Sprintf("%x", hash.Sum([]byte(uri.String())))
}

func IsBlacklisted(uri *url.URL) bool {
    if _, err := cache.Get(hash(uri)); err == nil {
        return true
    }
    return false
}

func Blacklist(uri *url.URL) {
    cache.Set(hash(uri), "blacklisted", TTL)
}
