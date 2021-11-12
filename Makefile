test:
	go test -v ./... -race -covermode atomic -count=1 -timeout 30s