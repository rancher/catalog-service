package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testAscending(t *testing.T, versions []string) {
	for i, version := range versions {
		for j := i; j < len(versions); j++ {
			if j != i {
				assert.True(t, VersionGreaterThan(versions[j], version))
			}
		}
	}
}

func TestGreaterThan(t *testing.T) {
	testAscending(t, []string{
		"v1.2.0",
		"v1.2.1",
		"v1.2.3",
		"v1.3.0",
		"v1.3.4",
		"v2.0.0",
	})

	testAscending(t, []string{
		"v0.1.0-rancher1",
		"v1.2.4-rancher6",
		"v1.2.4-rancher6.1",
		"v1.2.4-rancher7",
		"v1.2.4-rancher7.2",
		"v1.2.4-rancher7.3",
		"v1.3.0-rancher3",
		"v1.3.0-rancher4",
	})
	// TODO: this should pass
	//assert.True(t, GreaterThan("v1.2.0-rancher12", "v1.2.0-rancher3"))
}
