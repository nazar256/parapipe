test:
	golangci-lint run ./... --fix
	go test -v  -race -timeout 5s
bench:
	go test -benchmem -bench=. -timeout 10s -benchtime 5s -memprofile=mem.out