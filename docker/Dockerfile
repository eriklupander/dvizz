# build stage: fetch bower dependencies
FROM node AS bower

WORKDIR /dvizz

ADD . /dvizz
RUN npm install -g bower && bower --allow-root install



# build stage: dvizz golang binary
FROM golang AS golang

WORKDIR /dvizz

ADD . /dvizz
RUN go get -v -d && CGO_ENABLED=0 go build -a -o dvizz



# final image
FROM scratch

EXPOSE 6969

WORKDIR /dvizz

ADD static/ static/
COPY --from=golang /dvizz/dvizz .
COPY --from=bower /dvizz/static/js static/js

ENTRYPOINT ["./dvizz"]
CMD []
