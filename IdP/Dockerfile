FROM golang:1.13

WORKDIR /go/src/app
COPY . .

RUN go build


FROM debian
COPY --from=0 /go/src/app/idp .
COPY template template
RUN apt-get update && apt-get install -y --no-install-recommends \
          ca-certificates
CMD ["sh", "-c", "./idp -dsn $RUNTIME_DSN -admin-url $RUNTIME_HYDRA_ADMIN"]


