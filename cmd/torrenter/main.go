package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudyy74/torrenter/internal/torrentfile"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <path to torrent> <output path>", os.Args[0])
		return
	}
	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrentfile.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := tf.DownloadToFile(outPath); err != nil {
		log.Fatal(err)
	}
}
