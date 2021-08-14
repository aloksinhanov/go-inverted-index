package indexer

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Result []string

func (r Result) ToString() string {
	var b strings.Builder

	for _, v := range r {
		b.Write([]byte(v))
		b.WriteRune('\n')
	}
	return b.String()
}

var (
	//This is where we store the index
	//ideally this can go into a distributed database with fast writes
	//cassandra/redis?
	Looker sync.Map

	//A job queue to publish the indexing jobs to
	jobs = make(chan string, 1000000)

	//To notify the producer whenever a worker is done with a job
	finisher = make(chan int, 100)

	//For consensus between the producer and the workers if the indexing is done
	totalFiles, doneFiles int
)

//Bunch of words which are not really interesting/useful from inverted indexing perspective.
//Not a very comprehensive list. More can be added to it.
//This can also be read from a json file, so that we can add to this list dynamically.
var stopWords = map[string]int{
	"I":   1,
	"the": 1,
	"we":  1,
	"is":  1,
	"and": 1,
}

//Delimiting by space does not strip the punctations like comma or full-stop.
//Someone trying to find a keyword will not specify that with a punctation.
//Hence, stripping a trailing comma or full-stop.
//Can be extended to other characters.
func stripPunctuation(token string) string {
	return strings.TrimRight(strings.TrimRight(token, "."), ",")
}

//A work function that will be executed by the spawned worker routines.
//The file parsing happens in this function.
func work(ctx context.Context) {

	for {
		//Get a job which is nothing but a file path to be scanned
		path := <-jobs

		//Getting an empty line for some weird reason
		//will fix that later
		//for now just cheking and skipping
		if len(strings.TrimSpace(path)) == 0 {
			return
		}

		f, err := os.Open(path)
		if err != nil {
			log.Fatalf("Failed to open: %v | error: %v", path, err)
		}

		//Scan only if the file descriptor is not a directory
		log.Printf("\nScanning: %v", path)

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			tokens := strings.Split(line, " ")
			for _, token := range tokens {

				//An email may have a text like "jeremy.blachman@enron.com, abcdefg@gmail.com".
				//Splitting by space will give us two tokens - "jeremy.blachman@enron.com," and "abcdefg@gmail.com"
				//Notice the comma at the end of the firs token.
				//We need to strip punctuations from the end, example
				token = stripPunctuation(token)

				//Check if this token is a stop word
				if stopWords[token] == 1 {
					continue
				}

				//All the routines will try to access this shared map concurrently.
				//Hence, using a synchronized map.
				val, ok := Looker.Load(token)
				if ok {
					//log.Printf("\nAdding to Looker | key: %v | val: %v", token, path)
					Looker.Store(token, append(val.(Result), path))
					continue
				}

				paths := make(Result, 0)
				//log.Printf("\nAdding to Looker | key: %v | val: %v", token, path)
				Looker.Store(token, append(paths, path))
			}
		}
		f.Close()
		log.Println("Finisher")
		finisher <- 1
	}
}

//startWorkers : Starts a bunch of worker routines. This is a very basic worker pool.
//we can make it a more generic worker pool which can do anything by passing work as an
//arg to it.
func startWorkers(count int, ctx context.Context) {
	for i := 0; i < count; i++ {
		go work(ctx)
	}
}

//enqueue : This function is used by the job producer to enqueue jobs into the queue.
//Only the enqueuing of jobs happen in a single routine.
//The indexing can happen in parallel by workers.
func enqueue(path string, info fs.FileInfo, err error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open: %v | error: %v", path, err)
	}

	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to state: %v | error: %v", path, err)
	}

	//Adds to queue for scanning only if path is not a directory
	if !fi.IsDir() {
		totalFiles++
		jobs <- path
	}

	return nil
}

//Index : will start the indexing process
func Index(root string) {
	ctx := context.Background()
	//ctx, cancel := context.WithCancel(ctx)

	startWorkers(100, ctx)

	err := filepath.Walk(root, enqueue)
	if err != nil {
		log.Fatal(err)
	}

	for {
		<-finisher
		doneFiles++
		if doneFiles == totalFiles {
			close(jobs)
			//cancel()
			return
		}
	}
}
