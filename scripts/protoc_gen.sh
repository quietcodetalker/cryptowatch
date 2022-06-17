protoc \
  --go_out=. --go_opt=paths=import \
  --go-grpc_out=. --go-grpc_opt=paths=import \
  api/proto/v1/users.proto \
  api/proto/v1/portfolios.proto \
  api/proto/v1/triggers.proto

protoc -I api/proto/v1 --grpc-gateway_out pkg/api/cryptowatchv1 \
          --grpc-gateway_opt logtostderr=true \
          --grpc-gateway_opt paths=source_relative \
          --grpc-gateway_opt generate_unbound_methods=true \
            api/proto/v1/users.proto \
            api/proto/v1/portfolios.proto \
            api/proto/v1/triggers.proto
