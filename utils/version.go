package utils

import (
	"strings"

	"github.com/blang/semver"
	"github.com/rancher/catalog-service/utils/version"
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

func VersionSatisfiesRange(v, rng string) (bool, error) {
	v = strings.TrimLeft(v, "v")
	sv, err := semver.Parse(v)
	if err != nil {
		return false, err
	}
	rangeFunc, err := semver.ParseRange(rng)
	if err != nil {
		return false, err
	}
	return rangeFunc(sv), nil
}

func VersionGreaterThan(a, b string) bool {
	return version.GreaterThan(a, b)
}
