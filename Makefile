build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/blog blog/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/blog_indexer blog_indexer/main.go
