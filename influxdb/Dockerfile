FROM golang:1.18

WORKDIR /chaincode
RUN ls -la
COPY  ../influxdb ./
COPY ../influxdb/ ./influxdb/
RUN echo PWD
run ls -al
RUN ls -la . &&  GO111MODULE=on GOOS=linux CGO_ENABLED=0 go build -o svHny.bin svHny.go
