proto:
	protoc --proto_path=./server/protos --go_out=./server/apis/pb --go_opt=paths=source_relative --go-grpc_out=./server/apis/pb --go-grpc_opt=paths=source_relative ./server/protos/service.proto ./server/protos/dto.proto

srv.up:
	go run ./server/main.go