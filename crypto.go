package main

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
)

func verify(token, timestamp, nonce, signature string) bool {
	s := []string{token, timestamp, nonce}

	sort.Strings(s)
	str := strings.Join(s, "")
	hashs := sha1.New()
	hashs.Write([]byte(str))

	sig := hex.EncodeToString(hashs.Sum(nil))
	if sig == signature {
		return true
	} else {
		return false
	}
}

func verifyOnCipherMode() bool {
	return true
}
