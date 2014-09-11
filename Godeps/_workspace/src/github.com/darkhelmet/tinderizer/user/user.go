package user

import "github.com/darkhelmet/tinderizer/cache"

const TTL = 2 * 60 // 2 minutes

func Notify(key string, message string) {
    cache.Set(key, message, TTL)
}
