FROM scratch
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY bill18test .
CMD ["/bill18test"]

# компиляция образа
#docker build -t puzanovma/bill18test-scratch -f Dockerfile.scratch .