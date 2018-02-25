package main

import (
	"os"
	"path/filepath"
)

func fscacheRefresh() {
	err := filepath.Walk(config.Mnt, func(path string, info os.FileInfo, err error) error {
		if err != nil {

		} else {
			log.Error(err.Error())
		}
		return nil
	})
	if err != nil {
		log.Error(err.Error())
	}
}
