FROM golang:latest
ENV GO111MODULE on
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
#RUN go get github.com/pilu/fresh
#RUN fresh &
