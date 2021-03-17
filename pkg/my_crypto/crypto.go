package my_crypto

import (
	"encoding/binary"
)

func Decrypt(message []byte, key []byte) uint32 {
	// fmt.Println("crypto", message)
	return binary.LittleEndian.Uint32(message)
}
