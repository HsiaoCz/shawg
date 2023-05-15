run:
	@go run main.go

test:
	@go test -v ./...

.PHONY:run test