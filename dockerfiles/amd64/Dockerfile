FROM alpine:3.8
RUN apk --no-cache add ca-certificates
COPY fxapi /
WORKDIR /
ENTRYPOINT ["/fxapi"]