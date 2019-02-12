FROM golang:latest 
RUN mkdir -p /go/src/imcgbackend
ADD . /go/src/github.com/imcgaunn/imcgbackend/
WORKDIR /go/src/imcgbackend
# Q: how to add software to golang container?
# A: it's based on debian or alpine depending
# on the selected flavor, defaulting to debian.
RUN go get -u github.com/golang/dep/cmd/dep
ENV PATH "${PATH}:$(go env GOPATH)/bin"
RUN make