package main

import (
	"context"
	"log"
	"os"

	"github.com/u-root/u-root/pkg/boot/syslinux"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("Expecting 1 argument.")
		return
	}
	path := args[0]

	iso, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer iso.Close()

	syslinux.ParseLocalConfig(context.Background(), path)
}
