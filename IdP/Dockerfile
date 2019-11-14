FROM golang:1.13

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["sh", "-c", "idp -dsn $RUNTIME_DSN -admin-url $RUNTIME_HYDRA_ADMIN"]
