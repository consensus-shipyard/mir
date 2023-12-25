/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protos

//go:generate -command protoc-basic protoc --proto_path=. --go_out=../pkg/pb/ --go_opt=paths=source_relative

// Generate the code for codegen extensions.
//go:generate protoc-basic mir/codegen_extensions.proto
//go:generate protoc-basic net/codegen_extensions.proto

// Build the protoc plugin.
//go:generate go build -o ../codegen/protoc-plugin/protoc-gen-mir ../codegen/protoc-plugin

//go:generate -command protoc-events protoc --proto_path=. --go_out=../pkg/pb/ --go_opt=paths=source_relative --plugin=../codegen/protoc-plugin/protoc-gen-mir --mir_out=../pkg/pb --mir_opt=paths=source_relative

// Generate the protoc-generated code for events and messages.
//go:generate protoc-events messagepb/messagepb.proto
//go:generate protoc-events trantorpb/trantorpb.proto
//go:generate protoc-events eventpb/eventpb.proto
//go:generate protoc-events hasherpb/hasherpb.proto
//go:generate protoc-events recordingpb/recordingpb.proto
//go:generate protoc-events isspb/isspb.proto
//go:generate protoc-events bcbpb/bcbpb.proto
//go:generate protoc-events pbftpb/pbftpb.proto
//go:generate protoc-events contextstorepb/contextstorepb.proto
//go:generate protoc-events dslpb/dslpb.proto
//go:generate protoc-events mempoolpb/mempoolpb.proto
//go:generate protoc-events availabilitypb/availabilitypb.proto
//go:generate protoc-events availabilitypb/mscpb/mscpb.proto
//go:generate protoc-events factorypb/factorypb.proto
//go:generate protoc-events availabilitypb/batchdbpb/batchdbpb.proto
//go:generate protoc-events batchfetcherpb/batchfetcherpb.proto
//go:generate protoc-events threshcryptopb/threshcryptopb.proto
//go:generate protoc-events pingpongpb/pingpongpb.proto
//go:generate protoc-events checkpointpb/checkpointpb.proto
//go:generate protoc-events checkpointpb/chkpvalidatorpb/chkpvalidatorpb.proto
//go:generate protoc-events ordererpb/ordererpb.proto
//go:generate protoc-events ordererpb/pprepvalidatorpb/pprepvalidatorpb.proto
//go:generate protoc-events cryptopb/cryptopb.proto
//go:generate protoc-events apppb/apppb.proto
//go:generate protoc-events transportpb/transportpb.proto
//go:generate protoc-events accountabilitypb/accountabilitypb.proto
//go:generate protoc-events testerpb/testerpb.proto

// Build the custom code generators.
//go:generate go build -o ../codegen/generators/mir-std-gen/mir-std-gen.bin ../codegen/generators/mir-std-gen
//go:generate -command std-gen ../codegen/generators/mir-std-gen/mir-std-gen.bin

// Generate the Mir-generated code for events and messages.
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/eventpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/messagepb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/bcbpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/dslpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/checkpointpb/chkpvalidatorpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/threshcryptopb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/trantorpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/isspb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/ordererpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/ordererpb/pprepvalidatorpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/pbftpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/factorypb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/hasherpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/cryptopb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/apppb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/transportpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/accountabilitypb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/testerpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/pingpongpb"

// Generate other things.
//go:generate protoc --go_out=../pkg/ --go_opt=paths=source_relative --go-grpc_out=../pkg/ --go-grpc_opt=paths=source_relative transactionreceiver/transactionreceiver.proto
//go:generate protoc --go_out=../pkg/ --go_opt=paths=source_relative --go-grpc_out=../pkg/ --go-grpc_opt=paths=source_relative net/grpc/grpctransport.proto
//xgo:generate protoc --proto_path=. --go_out=plugins=grpc:../pkg/ --go_opt=paths=source_relative grpctransport/grpctransport.proto
//xgo:generate protoc --proto_path=. --go_out=plugins=grpc:../pkg/ --go_opt=paths=source_relative transactionreceiver/transactionreceiver.proto
