FROM golang:latest AS compiling_stage
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
RUN go install .

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Ivan Ivanov<test@test.ru>"
WORKDIR /root/
COPY --from=compiling_stage /go/bin/main .
ENTRYPOINT ./main