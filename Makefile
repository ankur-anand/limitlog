test:
	go test -v ./... -race -covermode atomic -count=1 -timeout 30s

## This will update the golden file
testu:
	go test -v . -race -covermode atomic -count=1 -timeout 60s -update

bench:
	go test -run=xxx -bench=. -cpuprofile profile_cpu.out

sample:
	time cat ./testdata/sample.input| go run ./cmd/main.go

large:
	time cat ./testdata/key100000.golden| go run ./cmd/main.go