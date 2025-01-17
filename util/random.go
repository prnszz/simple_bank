package util

import (
	"math/rand"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// func init() {
// 	// Seed the random number generator with the current time
// 	rand.Seed(time.Now().UnixNano())
// }
// after go 1.20, we don't need to initialize the random number generator

func RandomInt(min, max int64) int64 {
	// Generate a random number between min and max
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	// Generate a random string of length n
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomOwner() string {
	return RandomString(5)
}

func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "CAD"}
	return currencies[rand.Intn(len(currencies))]
}
