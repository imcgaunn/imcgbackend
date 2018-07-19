build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/blog blog/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/blog_indexer blog_indexer/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/blogindex blogindex/main.go

.PHONY: deploy
deploy:
	sls deploy -s dev