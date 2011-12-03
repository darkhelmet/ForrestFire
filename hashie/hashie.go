package hashie

import (
    "crypto/sha1"
    "fmt"
)

func Sha1(args ...[]byte) string {
    hash := sha1.New()
    for _, arg := range args {
        hash.Write(arg)
    }
    return fmt.Sprintf("%x", hash.Sum(nil))
}
