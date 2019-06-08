# build stage: fetch bower dependencies
FROM node AS bower

WORKDIR /dvizz

# Copy only needed frontend files instead of everything
ADD .bowerrc /dvizz
ADD bower.json /dvizz
ADD static/* /dvizz/static/

RUN npm install -g bower && bower --allow-root install


# build stage: dvizz golang binary
FROM golang:1.12.0-stretch AS build_base

ENV GO111MODULE=on \
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=amd64

WORKDIR /go/src/github.com/eriklupander/dvizz

# allows docker to cache go modules based on these layers remaining unchanged.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN	go build -a \
	-o bin/dvizz $PWD/cmd/dvizz

# final image
FROM alpine:latest

EXPOSE 6969

WORKDIR /app

# Copy frontend code
ADD static/ static/

# Copy binary from build_base image
COPY --from=build_base /go/src/github.com/eriklupander/dvizz/bin/* /app

# Copy frontend/js dependencies from bower build image
COPY --from=bower /dvizz/static/js static/js

# Support static build docker binary
RUN mkdir /lib64 \
&& ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# Not sure why I have to chmod/chown
RUN chmod +x /app/dvizz
RUN chmod 777 /app/dvizz

CMD ["./dvizz"]
