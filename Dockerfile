FROM golang:1.11.9

WORKDIR /go/src/service
COPY . .

EXPOSE 1323
ENV GO111MODULE=on
RUN go mod init main
RUN go build
CMD [ "./main" ]