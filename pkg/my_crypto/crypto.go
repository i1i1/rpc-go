package my_crypto

import (
	"encoding/binary"
)

func Decrypt(message []byte, key []byte) uint32 {
	return binary.LittleEndian.Uint32(message)
}
