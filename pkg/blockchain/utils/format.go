package utils

import (
	"fmt"
)

func FormatBlockId(blockId uint64) string {
	strId := fmt.Sprint(blockId)
	// just in case we have a very small id for some reason
	if len(strId) > 6 {
		strId = strId[:6]
	}
	return strId
}

func FormatBlockIdSlice(blockIds []uint64) string {
	str := ""
	for _, blockId := range blockIds {
		str += FormatBlockId(blockId) + " "
	}
	return str
}
