package utils

import (
	"regexp"
	"strconv"
	"strings"
)

func VersionGreaterThan(a, b string) bool {
	a = strings.TrimLeft(a, "v")
	b = strings.TrimLeft(b, "v")

	aSplit := periodDashSplit(a)
	bSplit := periodDashSplit(b)

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

func periodDashSplit(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		switch r {
		case '.', '-':
			return true
		}
		return false
	})
}

func pieceGreaterThan(a, b string) bool {
	re := regexp.MustCompile("[0-9]+")
	aMatch := re.FindString(a)
	bMatch := re.FindString(b)
	if aMatch == "" || bMatch == "" {
		return strings.Compare(a, b) > 0
	}
	aNum, _ := strconv.Atoi(aMatch)
	bNum, _ := strconv.Atoi(bMatch)
	return aNum > bNum
}
