package main

import (
	"crypto/rand"
	"io"
)

var (
	cryptogenSet       = "abcdefghijklmnopqrstuvwxyzABZDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	cryptogenThreshold = 10
)

// cryptogen generates a `n`-long random string that can be used in any
// security specific context. It doesn't use `PRNGs`
func cryptogen(n int) (string, error) {
	var err error
	buf := make([]byte, n)
	for i := 0; i < cryptogenThreshold; i++ {
		_, err = io.ReadFull(rand.Reader, buf)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", err
	}

	str := ""
	for _, b := range buf {
		str += string(cryptogenSet[(b>>2)%62])
	}

	return str, nil
}
