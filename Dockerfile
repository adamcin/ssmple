FROM golang:1.10.3-alpine as build
RUN apk add --no-cache --virtual git
RUN go get \
    github.com/aws/aws-sdk-go-v2 \
    github.com/jmespath/go-jmespath \
    github.com/go-ini/ini \
    github.com/rickar/props \
    gopkg.in/yaml.v2

RUN mkdir /app

ADD . /app/
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssmple .

FROM scratch
COPY --from=build /app/ssmple /root/ssmple
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/root/ssmple"]