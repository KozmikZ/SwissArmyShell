package utils

import "fmt"

type SizeType struct {
	Type string
	size float64
}

func BytesHumanReadable(inBytes int64) string {
	newInBytes := float64(inBytes)
	sizeTypes := []SizeType{{Type: "B", size: 1}, {Type: "KB", size: 1000}, {Type: "MB", size: 1000000}, {Type: "GB", size: 1000000000}}
	for _, t := range sizeTypes {
		newInBytes = newInBytes / t.size
		if newInBytes < 1000 {
			return fmt.Sprint(newInBytes) + " " + t.Type
		}
	}
	return fmt.Sprint(newInBytes) + " B" // edge case
}
