build:
	go get -u github.com/golang/dep/cmd/dep
	$(GOPATH)/bin/dep ensure
	env CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o bin/blog blog/main.go
	env CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o bin/blog_indexer blog_indexer/main.go
	env CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o bin/blogindex blogindex/main.go

.PHONY: deploy
deploy:
	sls deploy -s dev

.PHONY: clean
clean:
	rm -rf bin/*