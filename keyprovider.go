package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func keyProvider(n int, threshold int) (key string, err error) {
	keyBuffer := make([]byte, n)

	if threshold == 0 {
		return
	}

	for i := 0; i < threshold; i++ {
		_, err = io.ReadFull(rand.Reader, keyBuffer)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}

	return base64.StdEncoding.EncodeToString(keyBuffer), nil
}
