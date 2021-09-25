FROM golang:alpine as builder
COPY . /app
WORKDIR /app
RUN go build -o MultiTOR

FROM alpine
RUN apk update --no-cache
RUN adduser -D app
WORKDIR /home/app
COPY --from=builder /app/MultiTOR .  
RUN chown app.app -R .
USER app
EXPOSE 2525 1412
CMD ["./MultiTOR"]
