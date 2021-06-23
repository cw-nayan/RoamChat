INPUT_DIR = proto
OUTPUT_PATH = proto
SERVICE_NAME = roamchat
INPUT_PATH = ${INPUT_DIR}/${SERVICE_NAME}.proto

go_gen:
	protoc -I=${OUTPUT_PATH} --go_out=${OUTPUT_PATH} ${INPUT_PATH}
grpc_gen:
	protoc -I${OUTPUT_PATH} --go-grpc_out=require_unimplemented_servers=false:${OUTPUT_PATH} ${INPUT_PATH}
gateway_gen:
	protoc -I${OUTPUT_PATH} --grpc-gateway_out=logtostderr=true:${OUTPUT_PATH} ${INPUT_PATH}
swagger_gen:
	protoc -I${OUTPUT_PATH} --swagger_out=logtostderr=true:${OUTPUT_PATH} ${INPUT_PATH}
inject_gen:
	protoc-go-inject-tag -XXX_skip=yaml,xml,json -input=${OUTPUT_PATH}/${SERVICE_NAME}_grpc.pb.go
validator_gen:
	protoc -I${OUTPUT_PATH} --govalidators_out=logtostderr=true:${OUTPUT_PATH} ${INPUT_PATH}

proto_gen: proto_build

proto_build: go_gen grpc_gen gateway_gen swagger_gen inject_gen validator_gen

build_run:
	cd server && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o ${SERVICE_NAME} .
	./server/${SERVICE_NAME}

build_run_mac:
	cd server && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o ${SERVICE_NAME} .
	./server/${SERVICE_NAME}

build_run_silicon:
	cd server && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o ${SERVICE_NAME} .
	./server/${SERVICE_NAME}