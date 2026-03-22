FROM golang:1.23-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /out/rendezvous ./cmd/rendezvous && go build -o /out/punch-client ./cmd/punch-client

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/rendezvous /app/rendezvous
COPY --from=build /out/punch-client /app/punch-client
ENTRYPOINT ["/app/rendezvous"]
