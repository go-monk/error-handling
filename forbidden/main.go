package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
)

func main() {
	paths := []string{
		"/does/not/exist",
		"/etc/sudoers",
		"/etc/hosts",
		"/etc/master.passwd",
	}

	var forbidden []string

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				forbidden = append(forbidden, path)
				continue
			}
			log.Print(err)
		}
		f.Close()
	}

	fmt.Println(forbidden)
}
