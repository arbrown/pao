FROM golang:latest

# These are first so that they can be cached when code changes
RUN go get github.com/apexskier/httpauth
RUN go get github.com/gorilla/websocket
RUN go get github.com/lib/pq
RUN go get github.com/mattn/go-sqlite3

ADD ./client                   /go/src/github.com/arbrown/pao/client/
ADD ./db                       /go/src/github.com/arbrown/pao/db/
ADD ./game                     /go/src/github.com/arbrown/pao/game/
ADD ./settings                 /go/src/github.com/arbrown/pao/settings/
ADD ./pao.go                   /go/src/github.com/arbrown/pao/
ADD ./docker/paoSettings.json  /go/src/github.com/arbrown/pao/conf/

WORKDIR /go/src/github.com/arbrown/pao/

# Run a separate build step (easy to see missing dependencies, code problems)
RUN go build -o pao /go/src/github.com/arbrown/pao/pao.go

EXPOSE 2015

CMD "./pao"

