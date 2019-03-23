FROM golang:latest
ADD . /go/src/github.com/bid_auction
WORKDIR go/src/github.com/bid_auction
EXPOSE 9000
CMD ["go", "run", "app.go"]