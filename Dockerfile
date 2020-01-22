FROM golang:alpine AS builder
#RUN apk --no-cache add gcc g++ make git
RUN apk --no-cache add git tzdata
RUN adduser -D -g appuser appuser
WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bill18test ./main.go

FROM alpine:latest
LABEL MAINTAINER="Mikhail Puzanov <mpuzanov@mail.ru>"
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app .

RUN chown -R appuser .
USER appuser
EXPOSE 8091
#ENTRYPOINT ["/go/bin/bill18test","-conf=config.yaml"]
#CMD ["./bill18test","-conf=config.yaml"]
ENTRYPOINT ["./bill18test"]
CMD ["-conf=config.yml"]
