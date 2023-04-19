.PHONY: test dep tidy
test:
	@go test ./...

dep:
	@rm -f go.mod go.sum
	@go mod init github.com/whitekid/grpcx

	@$(MAKE) tidy

tidy:
	@go mod tidy -v

proto/service.pb.go: proto/service.proto
	protoc -I=./proto \
      --go_out=./proto \
      --go_opt=paths=source_relative \
      --go-grpc_out=./proto \
      --go-grpc_opt=paths=source_relative \
	  proto/service.proto

	cd ./proto && mockery --name SampleServiceClient
