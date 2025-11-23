# Proto generation for Go services
# Generates code from all .proto files under shared/proto/** into shared/generated/**

PROTO_DIR := shared/proto
PROTO_SRC := $(shell find $(PROTO_DIR) -name '*.proto')
GO_OUT := .

.PHONY: generate-proto
generate-proto:
	protoc -I=$(PROTO_DIR) -I=. \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(PROTO_SRC)
