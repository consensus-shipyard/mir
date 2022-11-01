package events

import (
	"path"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/codegen/util/params"
)

func PackagePath(sourcePackagePath string) string {
	return sourcePackagePath + "/events"
}

func PackageName(sourcePackagePath string) string {
	return sourcePackagePath[strings.LastIndex(sourcePackagePath, "/")+1:] + "events"
}

func OutputDir(sourceDir string) string {
	return path.Join(sourceDir, "events")
}

// EventNode represents a node in the tree corresponding to the hierarchy of events.
type EventNode struct {
	// The message for this event.
	message *types.Message
	// The option in the parent's Type oneof.
	oneofOption *types.OneofOption
	// The Type oneof field of the message (if present).
	typeOneof *types.Oneof
	// A field marked with [(mir.origin_request) = true] (if any).
	originRequest *Origin
	// A field marked with [(mir.origin_response) = true] (if any).
	originResponse *Origin
	// The children events in the hierarchy.
	// NB: It may happen that an event class has no children.
	children []*EventNode
	// The parent node in the hierarchy.
	parent *EventNode
	// The accumulated parameters for the constructor function.
	allConstructorParameters params.FunctionParamList
	// The parameters of the constructor function corresponding to the fields of this node in the hierarchy.
	thisNodeConstructorParameters params.ConstructorParamList
}

// IsRoot returns true if this is the root of the event hierarchy.
func (node *EventNode) IsRoot() bool {
	return node.parent == nil && node.IsEventClass()
}

// IsEventClass returns true iff the message has a oneof field marked with [(mir.event_type) = true].
func (node *EventNode) IsEventClass() bool {
	return node.typeOneof != nil
}

// IsEvent returns true if this is not an event class (see IsEventClass).
func (node *EventNode) IsEvent() bool {
	return !node.IsEventClass()
}

// Name returns the name of the event.
// Same as ev.Message().Name().
func (node *EventNode) Name() string {
	return node.Message().Name()
}

// Message returns the message for this event.
func (node *EventNode) Message() *types.Message {
	return node.message
}

// OneofOption returns the option in the parent's Type oneof.
// If nil, IsRoot() must be true.
func (node *EventNode) OneofOption() *types.OneofOption {
	return node.oneofOption
}

// TypeOneof returns the Type oneof field of the message (if present).
func (node *EventNode) TypeOneof() *types.Oneof {
	return node.typeOneof
}

// OriginRequest returns a field marked with [(mir.origin_request) = true] (if any).
func (node *EventNode) OriginRequest() *Origin {
	return node.originRequest
}

// OriginResponse returns a field marked with [(mir.origin_response) = true] (if any).
func (node *EventNode) OriginResponse() *Origin {
	return node.originResponse
}

// Children returns the children events in the hierarchy.
// NB: It may happen that an event class has no children.
func (node *EventNode) Children() []*EventNode {
	return node.children
}

// Parent returns the parent node in the hierarchy.
func (node *EventNode) Parent() *EventNode {
	return node.parent
}

// AllConstructorParameters returns the accumulated parameters for the constructor function.
// The parameters include all the fields of all the ancestors in the hierarchy except those marked with
// [(mir.omit_in_event_constructors) = true] and the Type oneofs.
func (node *EventNode) AllConstructorParameters() params.FunctionParamList {
	return node.allConstructorParameters
}

// ThisNodeConstructorParameters returns a suffix of AllConstructorParameters() that corresponds to the fields
// only of this in the hierarchy, without the parameters accumulated from the ancestors.
func (node *EventNode) ThisNodeConstructorParameters() params.ConstructorParamList {
	return node.thisNodeConstructorParameters
}

// Constructor returns the code corresponding to the constructor function for this event.
func (node *EventNode) Constructor() *jen.Statement {
	return jen.Qual(PackagePath(node.Message().PbPkgPath()), node.Name())
}
