package messages

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/filecoin-project/mir/codegen"

	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/codegen/util/params"
	"github.com/filecoin-project/mir/pkg/pb/net"
)

type Parser struct {
	typesParser     *types.Parser
	netMessageCache map[reflect.Type]*NetMessageNode
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
		typesParser:     messageParser,
		netMessageCache: make(map[reflect.Type]*NetMessageNode),
	}
}

// TypesParser returns the types.Parser used to parse the types in the event hierarchy.
func (p *Parser) TypesParser() *types.Parser {
	return p.typesParser
}

// ParseNetMessageHierarchy extracts the information about the whole net message hierarchy by its root.
func (p *Parser) ParseNetMessageHierarchy(netMessageRootMsg *types.Message) (root *NetMessageNode, err error) {

	if !codegen.IsNetMessageRoot(netMessageRootMsg.ProtoDesc()) {
		return nil, fmt.Errorf("message %v is not marked as net message root", netMessageRootMsg.Name())
	}

	root, err = p.parseNetMessageNodeRecursively(netMessageRootMsg, nil, nil, params.FunctionParamList{})
	return
}

// parseNetMessageNodeRecursively parses a message from the net message hierarchy.
// parent is the parent in the hierarchy. Note that the parent's list of children may not be complete.
func (p *Parser) parseNetMessageNodeRecursively(
	msg *types.Message,
	optionInParentOneof *types.OneofOption,
	parent *NetMessageNode,
	accumulatedConstructorParameters params.FunctionParamList,
) (node *NetMessageNode, err error) {

	// First, check the cache.
	if tp, ok := p.netMessageCache[msg.PbReflectType()]; ok {
		return tp, nil
	}

	// Remember the result in the cache when finished
	defer func() {
		if err == nil && node != nil {
			p.netMessageCache[msg.PbReflectType()] = node
		}
	}()

	fields, err := p.typesParser.ParseFields(msg)
	if err != nil {
		return nil, err
	}

	thisNodeConstructorParameters := params.ConstructorParamList{}
	for _, field := range fields {
		if IsMessageTypeOneof(field) {
			continue
		}

		originalName := field.LowercaseName()
		uniqueName := params.UniqueName(originalName, accumulatedConstructorParameters, thisNodeConstructorParameters)
		thisNodeConstructorParameters = thisNodeConstructorParameters.UncheckedAppend(originalName, uniqueName, field)
	}
	accumulatedConstructorParameters =
		accumulatedConstructorParameters.UncheckedAppendAll(thisNodeConstructorParameters.FunctionParamList())

	node = &NetMessageNode{
		message:                       msg,
		oneofOption:                   optionInParentOneof,
		typeOneof:                     getTypeOneof(fields),
		children:                      nil, // to be filled separately
		parent:                        parent,
		allConstructorParameters:      accumulatedConstructorParameters,
		thisNodeConstructorParameters: thisNodeConstructorParameters,
	}

	// Check if this is a net message class.
	if node.typeOneof != nil {
		if !codegen.IsNetMessageClass(msg.ProtoDesc()) && parent != nil {
			return nil, fmt.Errorf("message %v contains a oneof marked with option (net.message_type) = true, "+
				"but is not marked with option (net.message_class) = true", msg.PbReflectType())
		}

		for _, opt := range node.typeOneof.Options {
			childMsg, ok := opt.Field.Type.(*types.Message)
			if !ok {
				return nil, fmt.Errorf("non-message type in the net message hierarchy: %v", opt.Name())
			}

			if !childMsg.ShouldGenerateMirType() {
				// Skip children that are not marked as Mir
				continue
			}

			childNode, err := p.parseNetMessageNodeRecursively(childMsg, opt, node, accumulatedConstructorParameters)
			if err != nil {
				return nil, err
			}

			node.children = append(node.children, childNode)
		}

		return node, nil
	}

	// If this is not a net message class, it must be marked as a leaf (i.e. a net message).
	if node.typeOneof == nil && !codegen.IsNetMessage(msg.ProtoDesc()) {
		return nil, fmt.Errorf("message %v should be marked with option (net.message) = true", msg.PbReflectType())
	}

	return node, nil
}

func getTypeOneof(fields types.Fields) *types.Oneof {
	for _, field := range fields {
		// Recursively call the generator on all subtypes.
		if IsMessageTypeOneof(field) {
			return field.Type.(*types.Oneof)
		}
	}
	return nil
}

// IsMessageTypeOneof returns true iff the field is marked with `option (net.message_type) = true`.
func IsMessageTypeOneof(field *types.Field) bool {
	oneofDesc, ok := field.ProtoDesc.(protoreflect.OneofDescriptor)
	if !ok {
		return false
	}

	return proto.GetExtension(oneofDesc.Options().(*descriptorpb.OneofOptions), net.E_MessageType).(bool)
}
