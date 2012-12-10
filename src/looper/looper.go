package looper

import (
    "cache"
)

const (
    TTL    = 60 * 60 // 1 hour
    Resent = "resent"
)

func MapUrl(messageId, url string) {
    cache.Set(messageId, url, TTL)
}

func MarkResent(messageId, email string) (uri string) {
    uri, _ = cache.Get(messageId)
    if uri != "" {
        cache.Set(email+uri, Resent, TTL)
    }
    return uri
}

func AlreadyResent(messageId, email string) bool {
    uri, _ := cache.Get(messageId)
    if uri == "" {
        return false
    }
    v, _ := cache.Get(email + uri)
    return v == Resent
}
