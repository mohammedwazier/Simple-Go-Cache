FROM golang

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go get -t
RUN go build -o main .

EXPOSE 3000

CMD ["/app/main"]