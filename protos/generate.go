/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protos

//go:generate go build -o ../protoc-plugin/protoc-gen-mir ../protoc-plugin
//go:generate -command protoc-events protoc --proto_path=. --go_out=../pkg/pb/ --go_opt=paths=source_relative --plugin=../protoc-plugin/protoc-gen-mir --mir_out=../pkg/pb --mir_opt=paths=source_relative
//go:generate protoc-events mir/plugin.proto
//go:generate protoc-events commonpb/commonpb.proto
//go:generate protoc-events messagepb/messagepb.proto
//go:generate protoc-events requestpb/requestpb.proto
//go:generate protoc-events eventpb/eventpb.proto
//go:generate protoc-events recordingpb/recordingpb.proto
//go:generate protoc-events isspb/isspb.proto
//go:generate protoc-events bcbpb/bcbpb.proto
//go:generate protoc-events isspbftpb/isspbftpb.proto

//go:generate protoc --proto_path=. --go_out=:../pkg/ --go_opt=paths=source_relative simplewal/simplewal.proto
//go:generate protoc --proto_path=. --go_out=:../samples/ --go_opt=paths=source_relative chat-demo/chatdemo.proto
//go:generate protoc --go_out=../pkg/ --go_opt=paths=source_relative --go-grpc_out=../pkg/ --go-grpc_opt=paths=source_relative requestreceiver/requestreceiver.proto
//go:generate protoc --go_out=../pkg/ --go_opt=paths=source_relative --go-grpc_out=../pkg/ --go-grpc_opt=paths=source_relative grpctransport/grpctransport.proto
//xgo:generate protoc --proto_path=. --go_out=plugins=grpc:../pkg/ --go_opt=paths=source_relative grpctransport/grpctransport.proto
//xgo:generate protoc --proto_path=. --go_out=plugins=grpc:../pkg/ --go_opt=paths=source_relative requestreceiver/requestreceiver.proto
