package events

import "github.com/filecoin-project/mir/codegen/model/types"

// Origin represents a field marked with [(mir.origin_request) = true] or [(mir.origin_response) = true].
// The field must be of a message type with a field "Module" and a oneof "Type".
type Origin struct {
	// The field marked with [(mir.origin_request) = true] or [(mir.origin_response) = true].
	Field *types.Field

	// Same as Field.Type.(*types.Message).
	Message *types.Message

	// The field "Module" of the origin type.
	ModuleField *types.Field

	// The oneof "Type" of the origin type.
	TypeOneof *types.Oneof
}
