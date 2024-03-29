package stdtypes

import (
	"fmt"
	"strconv"
)

// RawData represents raw data contained in a deserialized event.
// When an event from the stdevents package undergoes serialization and deserialization,
// application-specific data types (represented by the Serializable type) are represented as RawData.
// In this form, they are handed over to application-specific code which is responsible for its deserialization.
type RawData []byte

func (d RawData) ToBytes() ([]byte, error) {
	return d, nil
}

func (d RawData) String() string {
	if isPrintableString(d) {
		return string(d)
	}
	return fmt.Sprintf("%v", []byte(d))
}

func isPrintableString(data []byte) bool {
	for _, r := range string(data) {
		if !strconv.IsPrint(r) {
			return false
		}
	}
	return true
	// TODO: Replace the above by the commented code below once the appropriate version of Go is supported.
	//   Currently we do not support such a high version of Go, because we are using an old version of libp2p
	//   that requires an old Go version (tested with go 1.19). However, updating libp2p to a higher version
	//   breaks the libp2p transport module. That needs to be fixed first.
	//return !bytes.ContainsFunc(data, func(r rune) bool {
	//	return !strconv.IsPrint(r)
	//})
}
