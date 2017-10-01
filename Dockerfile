FROM iron/base

EXPOSE 6969
ADD dvizz-linux-amd64 dvizz-linux-amd64
ADD static/*.css static/
ADD static/*.html static/
ADD static/js static/js
ENTRYPOINT ["./dvizz-linux-amd64"]