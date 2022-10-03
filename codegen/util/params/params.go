package params

import (
	"strconv"

	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen/generators/types-gen/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// FunctionParam represents a parameter of a function.
type FunctionParam struct {
	originalName string
	uniqueName   string
	type_        types.Type
}

// OriginalName return the original name of the function parameter.
func (p FunctionParam) OriginalName() string {
	return p.originalName
}

// Name returns the name of the function parameter after it has been made unique.
func (p FunctionParam) Name() string {
	return p.uniqueName
}

// Type returns the type of the function parameter.
func (p FunctionParam) Type() types.Type {
	return p.type_
}

// MirCode returns the code for the parameter that can be used in a function declaration using Mir-generated
func (p FunctionParam) MirCode() jen.Code {
	return jen.Id(p.Name()).Add(p.Type().MirType())
}

// ConstructorParam represents a parameter in a constructor of a message.
type ConstructorParam struct {
	FunctionParam
	field *types.Field
}

// Field returns the field associated with the constructor parameter.
func (p ConstructorParam) Field() *types.Field {
	return p.field
}

// FunctionParamList represents a list of parameters of a function.
type FunctionParamList struct {
	slice []FunctionParam
}

// Slice returns the underlying slice corresponding to the list.
func (l FunctionParamList) Slice() []FunctionParam {
	return l.slice
}

// FunctionParamListOf returns a FunctionParamList containing the given parameters.
func FunctionParamListOf(params ...FunctionParam) FunctionParamList {
	return FunctionParamList{params}
}

// UncheckedAppend returns a new FunctionParamList with an item appended to it.
// uniqueName must be different from all the parameters already in the list.
func (l FunctionParamList) UncheckedAppend(originalName, uniqueName string, tp types.Type) FunctionParamList {
	param := FunctionParam{originalName, uniqueName, tp}

	return FunctionParamList{append(l.Slice(), param)}
}

// Append returns a new FunctionParamList with an item appended to it.
// If a parameter with the same name is already in the list, it will change the name to make it unique.
func (l FunctionParamList) Append(name string, tp types.Type) FunctionParamList {
	return l.UncheckedAppend(name, UniqueName(name, l), tp)
}

// UncheckedAppendAll adds all parameters to the list.
// It is assumed that the names of the parameters being added are different from each other and all
// the parameters already in the list.
func (l FunctionParamList) UncheckedAppendAll(other FunctionParamList) FunctionParamList {
	res := l
	for _, param := range other.Slice() {
		res = res.UncheckedAppend(param.OriginalName(), param.Name(), param.Type())
	}

	return res
}

// AppendAll adds all parameters to the list, making sure that all parameters have unique names.
func (l FunctionParamList) AppendAll(other FunctionParamList) FunctionParamList {
	res := l
	for _, param := range other.Slice() {
		res = res.Append(param.OriginalName(), param.Type())
	}

	return res
}

// Adapt returns a new list that has no name collisions with any of the parameter in any of the provided lists.
func (l FunctionParamList) Adapt(others ...interface{ Names() []string }) FunctionParamList {
	res := FunctionParamList{}
	for _, param := range l.Slice() {
		uniqueName := UniqueName(param.OriginalName(), append(others, res)...)
		res = res.UncheckedAppend(param.OriginalName(), uniqueName, param.Type())
	}

	return res
}

// MirCode returns the slice of function parameters as a jen.Code list to be used in a function declaration.
func (l FunctionParamList) MirCode() []jen.Code {
	return sliceutil.Transform(l.Slice(), func(_ int, p FunctionParam) jen.Code { return p.MirCode() })
}

// IDs returns the slice of the names of the parameters as jen.Code.
func (l FunctionParamList) IDs() []jen.Code {
	return sliceutil.Transform(l.Slice(), func(_ int, p FunctionParam) jen.Code { return jen.Id(p.Name()) })
}

// Names returns the slice of the names of the parameters as string.
func (l FunctionParamList) Names() []string {
	return sliceutil.Transform(l.Slice(), func(_ int, p FunctionParam) string { return p.Name() })
}

// Sublist returns a FunctionParamList corresponding to a slice of the original list.
func (l FunctionParamList) Sublist(first, afterLast int) FunctionParamList {
	return FunctionParamList{slice: l.slice[first:afterLast]}
}

// ConstructorParamList represents a list of parameters of a constructor function of a message.
type ConstructorParamList struct {
	slice []ConstructorParam
}

// Slice returns the underlying slice corresponding to the list.
func (l ConstructorParamList) Slice() []ConstructorParam {
	return l.slice
}

// UncheckedAppend returns a new FunctionParamList with an item appended to it.
// uniqueName must be different from all the parameters already in the list.
func (l ConstructorParamList) UncheckedAppend(originalName, uniqueName string, field *types.Field) ConstructorParamList {
	param := ConstructorParam{
		FunctionParam: FunctionParam{originalName, uniqueName, field.Type},
		field:         field,
	}

	return ConstructorParamList{append(l.Slice(), param)}
}

// Append adds a parameter to the list, making sure that all parameters have unique names.
func (l ConstructorParamList) Append(name string, field *types.Field) ConstructorParamList {
	return l.UncheckedAppend(name, UniqueName(name, l), field)
}

// UncheckedAppendAll adds all parameters to the list.
// It is assumed that the names of the parameters being added are different from each other and all
// the parameters already in the list.
func (l ConstructorParamList) UncheckedAppendAll(other ConstructorParamList) ConstructorParamList {
	res := l
	for _, param := range other.Slice() {
		res = res.UncheckedAppend(param.OriginalName(), param.Name(), param.Field())
	}

	return res
}

// AppendAll adds all parameters to the list, making sure that all parameters have unique names.
func (l ConstructorParamList) AppendAll(other ConstructorParamList) ConstructorParamList {
	res := l
	for _, param := range other.Slice() {
		res = res.Append(param.OriginalName(), param.Field())
	}

	return res
}

// Adapt returns a new list that has no name collisions with any of the parameter in any of the provided lists.
func (l ConstructorParamList) Adapt(others ...interface{ Names() []string }) ConstructorParamList {
	res := ConstructorParamList{}
	for _, param := range l.Slice() {
		uniqueName := UniqueName(param.OriginalName(), append(others, res)...)
		res = res.UncheckedAppend(param.OriginalName(), uniqueName, param.Field())
	}

	return res
}

// MirCode returns the slice of function parameters as a jen.Code list to be used in a function declaration.
func (l ConstructorParamList) MirCode() []jen.Code {
	return sliceutil.Transform(l.Slice(), func(_ int, p ConstructorParam) jen.Code { return p.MirCode() })
}

// IDs returns the slice of the names of the parameters as jen.Code.
func (l ConstructorParamList) IDs() []jen.Code {
	return sliceutil.Transform(l.Slice(), func(_ int, p ConstructorParam) jen.Code { return jen.Id(p.Name()) })
}

// Names returns the slice of the names of the parameters as string.
func (l ConstructorParamList) Names() []string {
	return sliceutil.Transform(l.Slice(), func(_ int, p ConstructorParam) string { return p.Name() })
}

// Sublist returns a ConstructorParamList corresponding to a slice of the original list.
func (l ConstructorParamList) Sublist(first, afterLast int) ConstructorParamList {
	return ConstructorParamList{slice: l.slice[first:afterLast]}
}

// FunctionParamList transforms ConstructorParamList to FunctionParamList.
func (l ConstructorParamList) FunctionParamList() FunctionParamList {
	return FunctionParamList{
		slice: sliceutil.Transform(l.Slice(), func(_ int, p ConstructorParam) FunctionParam { return p.FunctionParam }),
	}
}

// UniqueName returns originalName modified in such a way that it is different
// from all the names in the provided parameter lists.
func UniqueName(originalName string, ls ...interface{ Names() []string }) string {
	suffixNumber := int64(-1)
	nameWithSuffix := originalName

	for _, l := range ls {
		for _, paramName := range l.Names() {
			if nameWithSuffix == paramName {
				suffixNumber += 1
				nameWithSuffix = originalName + strconv.FormatInt(suffixNumber, 10)
			}
		}
	}

	return nameWithSuffix
}
