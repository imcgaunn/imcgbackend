#!/usr/bin/env bash
docker run -it \
-v "$PWD":/go/src/imcgbackend \
-w /go/src/imcgbackend golang:latest make