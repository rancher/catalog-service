package utils

import "strings"

func VersionGreaterThan(a, b string) bool {
	a = strings.TrimLeft(a, "v")
	b = strings.TrimLeft(b, "v")

	aSplit := strings.Split(a, ".")
	bSplit := strings.Split(b, ".")

	for i := 0; i < len(aSplit); i++ {
		if i == len(bSplit) {
			return true
		}
		if pieceGreaterThan(aSplit[i], bSplit[i]) {
			return true
		}
	}

	return false
}

func pieceGreaterThan(a, b string) bool {
	return strings.Compare(a, b) > 0
}
