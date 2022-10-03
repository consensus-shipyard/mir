package events

import (
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/codegen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/codegen/util/params"
	"github.com/filecoin-project/mir/pkg/pb/mir"
)

type Parser struct {
	typesParser    *types.Parser
	eventNodeCache map[reflect.Type]*EventNode
}

var defaultParser = newParser(types.DefaultParser())

// DefaultParser returns a singleton Parser.
// It must not be accessed concurrently.
func DefaultParser() *Parser {
	return defaultParser
}

// newParser is not exported as DefaultParser() is supposed to be used instead.
func newParser(messageParser *types.Parser) *Parser {
	return &Parser{
		typesParser:    messageParser,
		eventNodeCache: make(map[reflect.Type]*EventNode),
	}
}

// TypesParser returns the types.Parser used to parse the types in the event hierarchy.
func (p *Parser) TypesParser() *types.Parser {
	return p.typesParser
}

// ParseEventHierarchy extracts the information about the whole event hierarchy by its root.
func (p *Parser) ParseEventHierarchy(eventRootMsg *types.Message) (root *EventNode, err error) {

	if !codegen.IsMirEventRoot(eventRootMsg.ProtoDesc()) {
		return nil, fmt.Errorf("message %v is not marked as event root", eventRootMsg.Name())
	}

	root, err = p.parseEventNodeRecursively(eventRootMsg, nil, nil, params.FunctionParamList{})
	return
}

// parseEventNodeRecursively parses a message from the event hierarchy.
// parent is the parent in the event hierarchy. Note that the parent's list of children may not be complete.
func (p *Parser) parseEventNodeRecursively(
	msg *types.Message,
	optionInParentOneof *types.OneofOption,
	parent *EventNode,
	accumulatedConstructorParameters params.FunctionParamList,
) (node *EventNode, err error) {

	// First, check the cache.
	if tp, ok := p.eventNodeCache[msg.PbReflectType()]; ok {
		return tp, nil
	}

	// Remember the result in the cache when finished
	defer func() {
		if err == nil && node != nil {
			p.eventNodeCache[msg.PbReflectType()] = node
		}
	}()

	fields, err := p.typesParser.ParseFields(msg)
	if err != nil {
		return nil, err
	}

	thisNodeConstructorParameters := params.ConstructorParamList{}
	for _, field := range fields {
		if IsEventTypeOneof(field) || OmitInEventConstructor(field) {
			continue
		}

		originalName := field.LowercaseName()
		uniqueName := params.UniqueName(originalName, accumulatedConstructorParameters, thisNodeConstructorParameters)
		thisNodeConstructorParameters = thisNodeConstructorParameters.UncheckedAppend(originalName, uniqueName, field)
	}
	accumulatedConstructorParameters =
		accumulatedConstructorParameters.UncheckedAppendAll(thisNodeConstructorParameters.FunctionParamList())

	// Check if this is an event class.
	if typeOneof, ok := getTypeOneof(fields); ok {
		if !codegen.IsMirEventClass(msg.ProtoDesc()) && parent != nil {
			return nil, fmt.Errorf("message %v contains a oneof marked with option (mir.event_type) = true, "+
				"but is not marked with option (mir.event_class) = true", msg.PbReflectType())
		}

		node := &EventNode{
			message:                       msg,
			oneofOption:                   optionInParentOneof,
			typeOneof:                     typeOneof,
			children:                      nil, // to be filled separately
			parent:                        parent,
			allConstructorParameters:      accumulatedConstructorParameters,
			thisNodeConstructorParameters: thisNodeConstructorParameters,
		}

		for _, opt := range typeOneof.Options {
			childMsg, ok := opt.Field.Type.(*types.Message)
			if !ok {
				return nil, fmt.Errorf("non-message type in the event hierarchy: %v", opt.Name())
			}

			if !childMsg.ShouldGenerateMirType() {
				// Skip children that are not marked as Mir
				continue
			}

			childNode, err := p.parseEventNodeRecursively(childMsg, opt, node, accumulatedConstructorParameters)
			if err != nil {
				return nil, err
			}

			node.children = append(node.children, childNode)
		}

		return node, nil
	}

	if !codegen.IsMirEvent(msg.ProtoDesc()) {
		return nil, fmt.Errorf("message %v should be marked with option (mir.event) = true", msg.PbReflectType())
	}

	return &EventNode{
		message:                       msg,
		oneofOption:                   optionInParentOneof,
		typeOneof:                     nil,
		children:                      nil,
		parent:                        parent,
		allConstructorParameters:      accumulatedConstructorParameters,
		thisNodeConstructorParameters: thisNodeConstructorParameters,
	}, nil
}

func getTypeOneof(fields types.Fields) (*types.Oneof, bool) {
	for _, field := range fields {
		// Recursively call the generator on all subtypes.
		if IsEventTypeOneof(field) {
			return field.Type.(*types.Oneof), true
		}
	}
	return nil, false
}

// IsEventTypeOneof returns true iff the field is marked with `option (mir.event_type) = true`.
func IsEventTypeOneof(field *types.Field) bool {
	oneofDesc, ok := field.ProtoDesc.(protoreflect.OneofDescriptor)
	if !ok {
		return false
	}

	return proto.GetExtension(oneofDesc.Options().(*descriptorpb.OneofOptions), mir.E_EventType).(bool)
}

func OmitInEventConstructor(field *types.Field) bool {
	oneofDesc, ok := field.ProtoDesc.(protoreflect.FieldDescriptor)
	if !ok {
		return false
	}

	return proto.GetExtension(oneofDesc.Options().(*descriptorpb.FieldOptions), mir.E_OmitInEventsConstructor).(bool)
}
