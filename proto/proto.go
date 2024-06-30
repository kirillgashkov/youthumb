//go:generate go run ../tools/protoc --go_out=. --go_opt=module=github.com/kirillgashkov/assignment-youthumb/proto --go-grpc_out=. --go-grpc_opt=module=github.com/kirillgashkov/assignment-youthumb/proto youthumb/v1/youthumb.proto

package proto
