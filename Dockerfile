FROM golang:latest 
RUN mkdir -p /go/src/imcgbackend
ADD . /go/src/imcgbackend/
WORKDIR /go/src/imcgbackend
# how to add software to golang container
RUN go get -u github.com/golang/dep/cmd/dep
ENV PATH "${PATH}:$(go env GOPATH)/bin"
RUN make && make deploy