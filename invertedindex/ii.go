package invertedindex

import (
	"strings"
	"sync"
	"unicode"
)

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		// Split on any character that is not a letter or a number.
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

func analyze(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	return tokens
}

// Index keeps a inverted index for each added document
type Index struct {
	mu sync.RWMutex
	// key is token and values are document ids.
	invertedIndex map[string]map[int64]struct{}
}

func NewInvertedIndex() *Index {
	return &Index{invertedIndex: make(map[string]map[int64]struct{}, 0)}
}

// Add the provided corpus text to the inverted index.
func (ii *Index) Add(text string, id int64) {
	ii.mu.Lock()
	defer ii.mu.Unlock()
	for _, token := range analyze(text) {
		ids := ii.invertedIndex[token]
		// add new
		if ids == nil {
			ii.invertedIndex[token] = make(map[int64]struct{}, 0)
		}

		// this given token is present inside the provided document id,
		ii.invertedIndex[token][id] = struct{}{}
	}
}

// RemoveAssociation deletes association between the provided corpus text and the provided ID.
func (ii *Index) RemoveAssociation(text string, id int64) {
	ii.mu.Lock()
	defer ii.mu.Unlock()
	for _, token := range analyze(text) {
		ids := ii.invertedIndex[token]
		// add new
		if ids == nil {
			continue
		}
		// delete the mapping
		delete(ii.invertedIndex[token], id)
	}
}

// Search query the index and returns the document id which has the given word.
func (ii *Index) Search(word string) map[int64]struct{} {
	ii.mu.RLock()
	defer ii.mu.RUnlock()
	result := make(map[int64]struct{})
	// single word would be normalized
	for _, token := range analyze(word) {
		if ids, ok := ii.invertedIndex[token]; ok {
			for k, v := range ids {
				result[k] = v
			}
		}
	}

	return result
}
