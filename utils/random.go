package utils

import (
	"math/rand"
	"time"
)

var STR_RANDOM = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var HEX_RANDOM = []byte("0123456789ABCDEF")
var NUM_RANDOM = []byte("0123456789")

func GetRandomString(l int64, base []byte) string {
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < l; i++ {
		result = append(result, base[r.Intn(len(base))])
	}
	return string(result)
}
