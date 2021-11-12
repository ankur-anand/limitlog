package limitlog

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/ankur-anand/limitlog/invertedindex"
)

var startTime = time.Now()

func nowNano() int64 { return time.Since(startTime).Nanoseconds() }

// InMemLogDB is an inMemory logs search system
// limited by capacity.
type InMemLogDB struct {
	vv  int64
	ii  *invertedindex.Index
	mu  sync.RWMutex
	lru *LRUCache
	// these are indexes
	keyDocIDIndex map[int]int64
	docIDKeyIndex map[int64]int
}

func NewInMemLogDB(cap int) *InMemLogDB {
	lru := NewLRU(cap)
	return &InMemLogDB{
		vv:            0,
		ii:            invertedindex.NewInvertedIndex(),
		lru:           lru,
		keyDocIDIndex: make(map[int]int64),
		docIDKeyIndex: make(map[int64]int),
	}
}

// Add the Key and Value to the DB
func (db *InMemLogDB) Add(key int, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// check if the key is already present
	old, evicted := db.lru.Add(key)

	docID := db.keyDocIDIndex[key]
	if evicted {
		docID = db.keyDocIDIndex[old]
	}

	// delete old association
	delete(db.docIDKeyIndex, docID)
	delete(db.keyDocIDIndex, old)

	// update new
	vv := atomic.AddInt64(&db.vv, 1)
	vv = nowNano() + vv

	db.docIDKeyIndex[vv] = key
	db.keyDocIDIndex[key] = vv
	db.ii.Add(value, vv)
}

func (db *InMemLogDB) isValidDocID(docID int64) bool {
	_, ok := db.docIDKeyIndex[docID]
	return ok
}

func (db *InMemLogDB) getAssociatedKey(docID int64) int {
	return db.docIDKeyIndex[docID]
}

// Search the given word and limit the number of slots.
func (db *InMemLogDB) Search(word string, limit int) []int {
	searchRes := db.ii.Search(word)
	validDocIDs := make([]int64, 0)

	db.mu.RLock()
	defer db.mu.RUnlock()
	for docID := range searchRes {
		ok := db.isValidDocID(docID)
		if !ok {
			// this should be deleted from the inverted index.
			// for delete we are optimistic.
			db.ii.RemoveAssociation(word, docID)
			continue
		}
		validDocIDs = append(validDocIDs, docID)
	}

	// we sort in reverse order
	sort.SliceStable(validDocIDs, func(i, j int) bool {
		return validDocIDs[i] > validDocIDs[j]
	})
	if len(validDocIDs) > limit {
		validDocIDs = validDocIDs[:limit]
	}

	keyRes := make([]int, 0)
	for _, v := range validDocIDs {
		keyRes = append(keyRes, db.getAssociatedKey(v))
	}
	return keyRes
}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		// Split on any character that is not a letter or a number.
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), ",", delim, -1), "[]") + "\n"
}

// ReaderInWriterOutLog wraps the InMemLogDB that performs the readFrom the provided reader in given format
// and write the result back until the END is encountered
type ReaderInWriterOutLog struct {
	ml *InMemLogDB
}

func (rwl *ReaderInWriterOutLog) GetInMemLogDBInstance() *InMemLogDB {
	return rwl.ml
}

func (rwl *ReaderInWriterOutLog) ReadFrom(r io.Reader, out io.Writer) (n int64, err error) {
	scanner := bufio.NewScanner(r)
	var size int
	var first bool
	for scanner.Scan() {
		line := scanner.Text()
		n = int64(len(scanner.Bytes()))
		tokens := tokenize(strings.TrimSpace(line))
		if len(tokens) == 0 {
			continue
		}

		// get the size
		if !first {
			if len(tokens) != 1 {
				return 0, fmt.Errorf("invalid input, expected length")
			}
			size, err = strconv.Atoi(tokens[0])
			if err != nil {
				return n, err
			}

			first = true
			rwl.ml = NewInMemLogDB(size)
		}

		switch tokens[0] {
		// END TOKEN
		case "END":
			_, err = out.Write([]byte("END"))
			return
		case "ADD":
			key, err1 := strconv.Atoi(tokens[1])
			if err1 != nil {
				err = err1
				return n, err
			}
			value := tokens[2:]
			rwl.ml.Add(key, strings.Join(value, " "))
		case "SEARCH":
			word := tokens[1]
			limit, err1 := strconv.Atoi(tokens[2])
			if err1 != nil {
				err = err1
				return n, err
			}
			res := rwl.ml.Search(word, limit)
			if len(res) == 0 {
				b := []byte("NONE\n")
				_, err = out.Write(b)
				if err != nil {
					return
				}
				continue
			}
			b := []byte(arrayToString(res, " "))
			_, err = out.Write(b)
			if err != nil {
				return
			}
		}

	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}
