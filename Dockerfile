FROM golang:alpine as builder
COPY . /app
WORKDIR /app
RUN go build -o MultiTOR

FROM alpine
RUN apk update --no-cache
RUN apk add privoxy tor curl --no-cache
RUN adduser -D app
WORKDIR /home/app
COPY --from=builder /app/MultiTOR .  
COPY --from=builder /app/privoxy/config.template ./privoxy/config.template 
RUN chown app.app -R .
USER app
EXPOSE 2525 8080 1412
CMD ["./MultiTOR","-privoxy","/usr/sbin/privoxy"]
