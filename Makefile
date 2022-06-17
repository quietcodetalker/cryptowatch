protoc_gen:
	sh ./scripts/protoc_gen.sh

run:
	go run ./cmd/grpc

build_image:
	DOCKER_BUILDKIT=0 docker build -t gitlab-registry.ozon.dev/unknownspacewalker/cryptowatch:latest --tag cryptowatch:latest -f ./deployments/cryptowatch/Dockerfile .

.PHONY:
	protoc_gen, run, build_image