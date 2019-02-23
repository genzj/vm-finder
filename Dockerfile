FROM golang:alpine
RUN echo "https://mirrors.aliyun.com/alpine/v3.9/main/" > /etc/apk/repositories

RUN apk update && \
    apk add --no-cache --virtual .build-deps git

RUN mkdir /app && \
    mkdir -p /go/src/ && \
    mkdir -p /go/bin/ && \
    mkdir -p /go/pkg/

ENV GOPATH /go/

# Go dep!
RUN go get -u github.com/golang/dep/...

RUN apk del .build-deps

ADD . /go/src/github.com/genzj/vm-finder
WORKDIR /go/src/github.com/genzj/vm-finder

RUN dep ensure

# Build my app
RUN go build -ldflags "-s -w" -v -o /app/vm-finder .
RUN GOOS=windows GOARCH=386 go build -ldflags "-s -w"  -v -o /app/vm-finder-32bit.exe .
RUN GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -v -o /app/vm-finder-64bit.exe .

CMD ["/app/vm-finder"]
