package utils

import (
	"regexp"
	"strconv"
	"strings"
)

func VersionBetween(a, b, c string) bool {
	if a == "" && c == "" {
		return true
	} else if a == "" {
		return !VersionGreaterThan(b, c)
	} else if b == "" {
		return true
	} else if c == "" {
		return !VersionGreaterThan(a, b)
	}
	return !VersionGreaterThan(a, b) && !VersionGreaterThan(b, c)
}

func VersionGreaterThan(a, b string) bool {
	re := regexp.MustCompile("[0-9]+")

	a = strings.TrimLeft(a, "v")
	b = strings.TrimLeft(b, "v")

	aSplit := periodDashSplit(a)
	bSplit := periodDashSplit(b)

	for i := 0; i < len(aSplit); i++ {
		if i == len(bSplit) {
			return true
		}
		aMatch := re.FindString(aSplit[i])
		bMatch := re.FindString(bSplit[i])
		if aMatch == "" || bMatch == "" {
			if strings.Compare(aSplit[i], bSplit[i]) > 0 {
				return true
			}
			if strings.Compare(bSplit[i], aSplit[i]) > 0 {
				return false
			}
		}
		aNum, _ := strconv.Atoi(aMatch)
		bNum, _ := strconv.Atoi(bMatch)
		if aNum > bNum {
			return true
		}
		if bNum > aNum {
			return false
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
