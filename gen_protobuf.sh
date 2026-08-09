#! /usr/bin/env bash
#
protoc \
  -I=. \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --go_opt=default_api_level=API_OPAQUE \
  api/shortner/shortner.proto \
&& mv api/shortner/shortner.pb.go api/shortner/shortner.pb.gen.go \
&& mv api/shortner/shortner_grpc.pb.go api/shortner/shortner_grpc.gen.go

