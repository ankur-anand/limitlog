## Limit Log

#### Problem Statement

```text
We need to manage logs on a system with limited memory. Each log is identified by a key( a
number of up to 15 digits) and a value (a string containing up to 15 alphanumeric words of at
most 15 characters each separated by spaces). We need to query which of the logs contain a
given word.

Read from the standard input and write to the standard output. The first line of the input is the
maximum size of logs you should keep S. Each of the subsequent lines either contains an:

i) ADD key value operation denoting to add this log. If the key is already present overwrite it
with the new value.

ii) SEARCH word limit operation denoting to search the word among the logs. Print the key
of the logs separated by space with the most recently added logs being printed first. Print only
up to limit keys. If the word is present in none of the logs output NONE.
The last line contains the single word END denoting the end of the program. Print END to the
output stream.

Limits:
1≤S≤10^6
1≤Size of word≤15
1≤key≤10^15
1≤limit≤1000

Sample Input:
2
ADD 56 the first
SEARCH the 1

ADD 25 the second log
SEARCH the 2
ADD 67 the third log
SEARCH the 3
SEARCH fourth 1
END
Sample output:
56
25 56
67 25
NONE
END
```

### Inverted Index:
Fast Search Build using the Inverted Index.

1. Each time `ADD KEY Text` is given, we preprocess the text and build an index in advance.
2. The inverted Index associates every word in Text with document ID that contain the word.
3. To Search the word, we search the word we used for indexing in the inverted index and return all document ID.


## Algorithm
1. `ADD KEY Text` will find check for in Fixed Size LRU Cache for size. If Size exceeds it will delete the oldest key from the
Cache and association.
2. Put the Text in inverted index and map it with document ID.
3. Document IDs are incremental based upon counter+unixNano Time.
4. Key and Document ID is mapped for retrieval.
5. `SEARCH WORD Limit` finds the document ID from the Inverted Index and returns the document ID.
6. Valid Key's are build from the returned document ID's after sorting and returned to user.

## How to Run.
1. Run Sample Input. `cat ./testdata/sample.input| go run ./cmd/main.go`
2. Run in interactive mode. `go run ./cmd/main.go` Enter the command as mentioned.
3. Run from other input source of file which have all command. Example `cat ./testdata/sample.input| go run ./cmd/main.go`

## Search Benchmark
BenchmarkSearch_FilledCapacity_2Keys-16 - **S=2** all filled.
BenchmarkSearch_FilledCapacity_100KKeys-16 **S=100000** all filled.
```text
goos: darwin
goarch: amd64
pkg: github.com/ankur-anand/limitlog
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
BenchmarkSearch_FilledCapacity_2Keys-16       	 1178023	      1001 ns/op	     408 B/op	      19 allocs/op
BenchmarkSearch_FilledCapacity_100KKeys-16    	 1013840	      1277 ns/op	     496 B/op	      23 allocs/op
PASS
```