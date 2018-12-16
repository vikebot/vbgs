# STEP 1 = clean compiling environment
FROM golang:alpine as builder

RUN apk update && apk add git && apk add ca-certificates

RUN adduser -D -g '' vbgs

# copy source code
COPY . /go/src/vbgs
WORKDIR /go/src/vbgs/

# set version extracted from git with makefile
ARG VERSION
ENV VERSION=$VERSION

# set go env variables
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# build
RUN go mod verify
RUN GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags="-w -s -X main.Version=$VERSION" -o /go/bin/vbgs

# STEP 2 = final production container
FROM gcr.io/distroless/base

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/vbgs /usr/local/bin/vbgs

USER vbgs

ENTRYPOINT [ "/usr/local/bin/vbgs" ]
