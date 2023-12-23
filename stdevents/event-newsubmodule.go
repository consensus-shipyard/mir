package stdevents

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type serializableNewSubmodule struct {
	mirEvent
	SubmoduleID    stdtypes.ModuleID
	Params         stdtypes.RawData
	RetentionIndex stdtypes.RetentionIndex
}

func (sns *serializableNewSubmodule) NewSubmodule() *NewSubmodule {
	return &NewSubmodule{
		mirEvent:       sns.mirEvent,
		SubmoduleID:    sns.SubmoduleID,
		Params:         sns.Params,
		RetentionIndex: sns.RetentionIndex,
	}
}

type NewSubmodule struct {
	mirEvent
	SubmoduleID    stdtypes.ModuleID
	Params         stdtypes.Serializable
	RetentionIndex stdtypes.RetentionIndex
}

func (e *NewSubmodule) serializable() (*serializableNewSubmodule, error) {
	data, err := e.Params.ToBytes()
	if err != nil {
		return nil, err
	}

	return &serializableNewSubmodule{
		mirEvent:       e.mirEvent,
		SubmoduleID:    e.SubmoduleID,
		Params:         data,
		RetentionIndex: e.RetentionIndex,
	}, nil

}

func NewNewSubmodule(
	dest stdtypes.ModuleID,
	submoduleID stdtypes.ModuleID,
	params stdtypes.Serializable,
	retIdx stdtypes.RetentionIndex,
) *NewSubmodule {
	return &NewSubmodule{
		mirEvent:       mirEvent{DestModule: dest},
		SubmoduleID:    submoduleID,
		Params:         params,
		RetentionIndex: retIdx,
	}
}

func NewNewSubmoduleWithSrc(
	src stdtypes.ModuleID,
	dest stdtypes.ModuleID,
	submoduleID stdtypes.ModuleID,
	params stdtypes.Serializable,
	retIdx stdtypes.RetentionIndex,
) *NewSubmodule {
	e := NewNewSubmodule(dest, submoduleID, params, retIdx)
	e.SrcModule = src
	return e
}

func (e *NewSubmodule) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.SrcModule = newSrc
	return &newE
}

func (e *NewSubmodule) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newE := *e
	e.DestModule = newDest
	return &newE
}

func (e *NewSubmodule) ToBytes() ([]byte, error) {
	serializable, err := e.serializable()
	if err != nil {
		return nil, err
	}

	return serialize(serializable)
}

func (e *NewSubmodule) ToString() string {
	data, err := e.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", e)
	}

	return string(data)
}
