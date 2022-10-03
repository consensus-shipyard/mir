package messages

import (
	"path"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/codegen/util/params"
)

func PackagePath(sourcePackagePath string) string {
	return sourcePackagePath + "/msgs"
}

func PackageName(sourcePackagePath string) string {
	return sourcePackagePath[strings.LastIndex(sourcePackagePath, "/")+1:] + "msgs"
}

func OutputDir(sourceDir string) string {
	return path.Join(sourceDir, "msgs")
}

// NetMessageNode represents a node in the tree corresponding to the hierarchy of net messages.
type NetMessageNode struct {
	// The protobuf message for this net message.
	message *types.Message
	// The option in the parent's Type oneof.
	oneofOption *types.OneofOption
	// The Type oneof field of the message (if present).
	typeOneof *types.Oneof
	// The children messages in the hierarchy.
	// NB: It may happen that a message class has no children.
	children []*NetMessageNode
	// The parent node in the hierarchy.
	parent *NetMessageNode
	// The accumulated parameters for the constructor function.
	allConstructorParameters params.FunctionParamList
	// The parameters of the constructor function corresponding to the fields of this node in the hierarchy.
	thisNodeConstructorParameters params.ConstructorParamList
}

// IsRoot returns true if this is the root of the message hierarchy.
func (msg *NetMessageNode) IsRoot() bool {
	return msg.parent == nil && msg.IsMsgClass()
}

// IsMsgClass returns true iff the message has a oneof field marked with [(mir.message_type) = true].
func (msg *NetMessageNode) IsMsgClass() bool {
	return msg.typeOneof != nil
}

// IsNetMessage returns true if this is not a msg class (see IsMsgClass).
func (msg *NetMessageNode) IsNetMessage() bool {
	return !msg.IsMsgClass()
}

// Name returns the name of the message.
// Same as ev.Message().Name().
func (msg *NetMessageNode) Name() string {
	return msg.Message().Name()
}

// Message returns the protobuf message for this net message.
func (msg *NetMessageNode) Message() *types.Message {
	return msg.message
}

// OneofOption returns the option in the parent's Type oneof.
// If nil, IsRoot() must be true.
func (msg *NetMessageNode) OneofOption() *types.OneofOption {
	return msg.oneofOption
}

// TypeOneof returns the Type oneof field of the message (if present).
func (msg *NetMessageNode) TypeOneof() *types.Oneof {
	return msg.typeOneof
}

// Children returns the children messages in the hierarchy.
// NB: It may happen that a message class has no children.
func (msg *NetMessageNode) Children() []*NetMessageNode {
	return msg.children
}

// Parent returns the parent node in the hierarchy.
func (msg *NetMessageNode) Parent() *NetMessageNode {
	return msg.parent
}

// AllConstructorParameters returns the accumulated parameters for the constructor function.
// The parameters include all the fields of all the ancestors in the hierarchy except those marked with
// [(mir.omit_in_constructor) = true] and the Type oneofs.
func (ev *NetMessageNode) AllConstructorParameters() params.FunctionParamList {
	return ev.allConstructorParameters
}

// ThisNodeConstructorParameters returns a suffix of AllConstructorParameters() that corresponds to the fields
// only of this in the hierarchy, without the parameters accumulated from the ancestors.
func (ev *NetMessageNode) ThisNodeConstructorParameters() params.ConstructorParamList {
	return ev.thisNodeConstructorParameters
}

func (msg *NetMessageNode) Constructor() *jen.Statement {
	return jen.Qual(PackagePath(msg.Message().PbPkgPath()), msg.Name())
}
