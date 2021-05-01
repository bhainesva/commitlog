protoc --proto_path=. --go_out=api --go_opt=paths=source_relative --js_out=import_style=commonjs,binary:frontend/gen api.proto
protoc --plugin=./frontend/node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=frontend/src/gen --ts_proto_opt=esModuleInterop=true api.proto
