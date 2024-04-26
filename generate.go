package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/HuXin0817/dots-and-boxes/pkg/models/file"
)

//go:generate go run generate.go

var GenList = []string{
	".mp3",
	".png",
}

func walkFunc(mp3Files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		for _, p := range GenList {
			if !info.IsDir() && filepath.Ext(path) == p {
				*mp3Files = append(*mp3Files, path)
			}
		}
		return nil
	}
}

func main() {
	var mp3Files []string
	root := "."

	if err := filepath.Walk(root, walkFunc(&mp3Files)); err != nil {
		fmt.Printf("error walking the path %q: %v\n", root, err)
	}

	for _, f := range mp3Files {
		if err := file.GenByteArray( f); err != nil {
			log.Println(err)
		}
	}
}
