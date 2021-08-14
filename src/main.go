package main

import (
	"fmt"
	"go-inverted-index/src/indexer"
	"go-inverted-index/src/retriever"
	"log"
	"strings"
)

func main() {
	indexer.Index("/home/alok/workspace/github.com/yosemite/skilling-j")
	log.Printf("\nindexing done..you can do quick lookups now\n %v ", indexer.Looker)
	for {
		fmt.Println("Enter a key to search. 'q' to Quit ")
		var key string
		fmt.Scanln(&key)
		if strings.TrimSpace(key) == "q" {
			return
		}
		log.Printf("\nResult: %v", retriever.Retrieve(key))
	}

}
