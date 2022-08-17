package events

import (
	"path"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen/generators/types-gen/params"
	"github.com/filecoin-project/mir/codegen/generators/types-gen/types"
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
func (ev *EventNode) IsRoot() bool {
	return ev.parent == nil && ev.IsEventClass()
}

// IsEventClass returns true iff the message has a oneof field marked with [(mir.event_type) = true].
func (ev *EventNode) IsEventClass() bool {
	return ev.typeOneof != nil
}

// IsEvent returns true if this is not an event class (see IsEventClass).
func (ev *EventNode) IsEvent() bool {
	return !ev.IsEventClass()
}

// Name returns the name of the event.
// Same as ev.Message().Name().
func (ev *EventNode) Name() string {
	return ev.Message().Name()
}

// Message returns the message for this event.
func (ev *EventNode) Message() *types.Message {
	return ev.message
}

// OneofOption returns the option in the parent's Type oneof.
// If nil, IsRoot() must be true.
func (ev *EventNode) OneofOption() *types.OneofOption {
	return ev.oneofOption
}

// TypeOneof returns the Type oneof field of the message (if present).
func (ev *EventNode) TypeOneof() *types.Oneof {
	return ev.typeOneof
}

// Children returns the children events in the hierarchy.
// NB: It may happen that an event class has no children.
func (ev *EventNode) Children() []*EventNode {
	return ev.children
}

// Parent returns the parent node in the hierarchy.
func (ev *EventNode) Parent() *EventNode {
	return ev.parent
}

// AllConstructorParameters returns the accumulated parameters for the constructor function.
// The parameters include all the fields of all the ancestors in the hierarchy except those marked with
// [(mir.omit_in_constructor) = true] and the Type oneofs.
func (ev *EventNode) AllConstructorParameters() params.FunctionParamList {
	return ev.allConstructorParameters
}

// ThisNodeConstructorParameters returns a suffix of AllConstructorParameters() that corresponds to the fields
// only of this in the hierarchy, without the fields accumulated from the ancestors.
func (ev *EventNode) ThisNodeConstructorParameters() params.ConstructorParamList {
	return ev.thisNodeConstructorParameters
}

func (ev *EventNode) Constructor() *jen.Statement {
	return jen.Qual(PackagePath(ev.Message().PbPkgPath()), ev.Name())
}
