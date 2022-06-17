FROM golang as build

COPY . /src

WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o cryptowatch ./cmd/grpc


FROM alpine

COPY --from=build /src/cryptowatch .

RUN apk --no-cache add ca-certificates

EXPOSE 8081
EXPOSE 50051

CMD ["/cryptowatch"]