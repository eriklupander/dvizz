FROM iron/base

EXPOSE 6969
ADD dvizz-linux-amd64 dvizz-linux-amd64
ADD static/*.html static/
ENTRYPOINT ["./dvizz-linux-amd64"]