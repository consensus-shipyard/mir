package stdtypes

import (
	"fmt"
	"strconv"
)

type RawMessage []byte

func (m RawMessage) ToBytes() ([]byte, error) {
	return m, nil
}

func (m RawMessage) String() string {
	if isPrintableString(m) {
		return string(m)
	}
	return fmt.Sprintf("%v", []byte(m))
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
