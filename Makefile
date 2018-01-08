all: clean build_proto

build_proto:
	protoc -I proto/ proto/crawl.proto --go_out=plugins=grpc:proto

clean:
	if [ -f ./proto/crawl.pb.go ]; \
	then \
		 rm ./proto/crawl.pb.go; \
	fi;
