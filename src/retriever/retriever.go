package retriever

import (
	"go-inverted-index/src/indexer"
)

func Retrieve(key string) string {
	val, ok := indexer.Looker.Load(key)
	if !ok {
		return "Not Found"
	}
	return val.(indexer.Result).ToString()
}
