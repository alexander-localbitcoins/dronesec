FROM golang:alpine

WORKDIR /go/src/dronesec
COPY . .

RUN go mod init && go mod tidy
RUN go install

WORKDIR /
ENTRYPOINT ["dronesec"]
