package uuid

import (
    "crypto/rand"
    "encoding/hex"
    "io"
)

type UUID []byte

func NewUUID() UUID {
    uuid := make([]byte, 16)

    if _, err := io.ReadFull(rand.Reader, uuid); err != nil {
        panic("Failed to read random values")
    }

    var version byte = 4 << 4
    var variant byte = 2 << 4

    uuid[6] = version | (uuid[6] & 15)
    uuid[8] = variant | (uuid[8] & 15)

    return uuid
}

func (uuid UUID) Raw() []byte {
    raw := make([]byte, 16)
    copy(raw, uuid[0:16])
    return raw
}

func (uuid UUID) String() string {
    base := hex.EncodeToString(uuid.Raw())
    return base[0:8] + "-" + base[8:12] + "-" + base[12:16] + "-" + base[16:20] + "-" + base[20:32]
}
