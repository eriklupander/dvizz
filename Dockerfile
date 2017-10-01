FROM iron/base

EXPOSE 6969

ADD dist/dvizz dvizz
ADD static/*.css static/
ADD static/*.html static/
ADD static/js static/js

ENTRYPOINT ["./dvizz"]
