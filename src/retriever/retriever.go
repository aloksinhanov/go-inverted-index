package retriever

import (
	"go-inverted-index/src/indexer"
)

//Retrieve: Retrieves the list of documents in which the given keyword is found
//Returns a "Not Found" if not found.
func Retrieve(key string) string {
	val, ok := indexer.Looker.Load(key)
	if !ok {
		return "Not Found"
	}
	return val.(indexer.Result).ToString()
}
