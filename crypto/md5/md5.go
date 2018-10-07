package md5

import (
	"crypto/md5"
	"io"
	"log"
	"os"
)

func MySelf() ([]byte, error) {
	f, err := os.Open(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
