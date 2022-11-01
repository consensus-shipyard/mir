package events

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/filecoin-project/mir/codegen"

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

	originRequest, err := p.getoriginRequest(fields)
	if err != nil {
		return nil, err
	}

	originResponse, err := p.getoriginResponse(fields)
	if err != nil {
		return nil, err
	}

	node = &EventNode{
		message:                       msg,
		oneofOption:                   optionInParentOneof,
		originRequest:                 originRequest,
		originResponse:                originResponse,
		typeOneof:                     getTypeOneof(fields),
		children:                      nil, // to be filled separately
		parent:                        parent,
		allConstructorParameters:      accumulatedConstructorParameters,
		thisNodeConstructorParameters: thisNodeConstructorParameters,
	}

	// Check if this is an event class.
	if node.typeOneof != nil {
		if !codegen.IsMirEventClass(msg.ProtoDesc()) && parent != nil {
			return nil, fmt.Errorf("message %v contains a oneof marked with option (mir.event_type) = true, "+
				"but is not marked with option (mir.event_class) = true", msg.PbReflectType())
		}

		for _, opt := range node.typeOneof.Options {
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
	}

	// If this is not an event class, it must be marked as a leaf (i.e. an event).
	if node.typeOneof == nil && !codegen.IsMirEvent(msg.ProtoDesc()) {
		return nil, fmt.Errorf("message %v should be marked with option (mir.event) = true", msg.PbReflectType())
	}

	return node, nil
}

func getTypeOneof(fields types.Fields) *types.Oneof {
	for _, field := range fields {
		if IsEventTypeOneof(field) {
			return field.Type.(*types.Oneof)
		}
	}
	return nil
}

func (p *Parser) getoriginRequest(fields types.Fields) (*Origin, error) {
	for _, field := range fields {
		if IsoriginRequest(field) {
			return p.ParseOrigin(field)
		}
	}

	return nil, nil
}

func (p *Parser) getoriginResponse(fields types.Fields) (*Origin, error) {
	for _, field := range fields {
		if IsoriginResponse(field) {
			return p.ParseOrigin(field)
		}
	}
	return nil, nil
}

func (p *Parser) ParseOrigin(field *types.Field) (*Origin, error) {
	message, ok := field.Type.(*types.Message)
	if !ok {
		return nil, fmt.Errorf("field %v.%v marked as origin request/response field must be of message type",
			field.Parent.Name(), field.Name)
	}

	fields, err := p.TypesParser().ParseFields(message)
	if err != nil {
		return nil, err
	}

	if len(fields) != 2 {
		return nil, fmt.Errorf("origin request/response message %v must have exactly 2 fields: Module and Type",
			message.Name())
	}

	return &Origin{
		Field:       field,
		Message:     message,
		ModuleField: fields.ByName("Module"),
		TypeOneof:   fields.ByName("Type").Type.(*types.Oneof),
	}, nil
}

// IsEventTypeOneof returns true iff the field is a oneof and is marked with option (mir.event_type) = true.
func IsEventTypeOneof(field *types.Field) bool {
	oneofDesc, ok := field.ProtoDesc.(protoreflect.OneofDescriptor)
	if !ok {
		return false
	}

	return proto.GetExtension(oneofDesc.Options().(*descriptorpb.OneofOptions), mir.E_EventType).(bool)
}

// IsoriginRequest returns true iff the field is marked with [(mir.origin_request) = true].
func IsoriginRequest(field *types.Field) bool {
	fieldDesc, ok := field.ProtoDesc.(protoreflect.FieldDescriptor)
	if !ok {
		return false
	}

	return proto.GetExtension(fieldDesc.Options().(*descriptorpb.FieldOptions), mir.E_OriginRequest).(bool)
}

// IsoriginResponse returns true iff the field is marked with [(mir.origin_response) = true].
func IsoriginResponse(field *types.Field) bool {
	fieldDesc, ok := field.ProtoDesc.(protoreflect.FieldDescriptor)
	if !ok {
		return false
	}

	return proto.GetExtension(fieldDesc.Options().(*descriptorpb.FieldOptions), mir.E_OriginResponse).(bool)
}

// OmitInEventConstructor returns true iff the field is marked with [(mir.omit_in_event_constructor) = true].
func OmitInEventConstructor(field *types.Field) bool {
	oneofDesc, ok := field.ProtoDesc.(protoreflect.FieldDescriptor)
	if !ok {
		return false
	}

	return proto.GetExtension(oneofDesc.Options().(*descriptorpb.FieldOptions), mir.E_OmitInEventConstructors).(bool)
}
