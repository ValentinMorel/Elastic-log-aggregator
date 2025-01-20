.PHONY: all


build-proto:
	protoc -I=backend/proto --go_out=backend/proto --go-grpc_out=backend/proto logmsg.proto
