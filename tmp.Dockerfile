# STEP 1 build executable binary
FROM golang:alpine as builder
# Install SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates
# Create appuser
RUN adduser -D -g '' appuser
COPY . $GOPATH/src/mypackage/myapp/
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . /go/bin/bill18test/
#get dependancies
RUN go get -d -v
#build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/bill18test

# STEP 2 build a small image
# start from scratch
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /go/bin/bill18test/ /go/bin/bill18test/
# Copy our static executable
COPY --from=builder /go/bin/bill18test /go/bin/bill18test

USER appuser

# Всегда экспортируйте порты > 1024 если есть возможность иначе нужны дополнительные привелегии
EXPOSE 9990

ENTRYPOINT ["/go/bin/bill18test"]
#ENTRYPOINT ["/go/bin/bill18test"]

# компиляция образа
#docker build -t puzanovma/bill18test .

# запуск контейнера
#docker run --rm -it -p 8091:9990 puzanovma/bill18test 