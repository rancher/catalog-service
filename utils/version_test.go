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
				assert.False(t, VersionGreaterThan(version, versions[j]))
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
		"v0.1.0-rancher0",
		"v0.1.0-rancher1",
		"v0.1.0-rancher1.1",
		"v1.2.4-rancher6",
		"v1.2.4-rancher6.1",
		"v1.2.4-rancher7",
		"v1.2.4-rancher7.2",
		"v1.2.4-rancher7.3",
		"v1.2.4-rancher9.0",
		"v1.2.4-rancher10.10",
		"v1.2.4-rancher12.0",
		"v1.2.4-rancher12.5",
		"v1.2.4-rancher14",
		"v1.2.4-rancher15.10",
		"v1.3.0-rancher3",
		"v1.3.0-rancher4",
	})

	testAscending(t, []string{
		"0.0.1",
		"v0.45.0",
	})

	// TODO: handle logic like this?
	/*testAscending(t, []string{
		"v0.1.0-alpha1",
		"v0.1.0-beta1",
		"v0.1.0-rc1",
		"v0.1.0",
	})*/
}

func TestVersionBetween(t *testing.T) {
	assert.True(t, VersionBetween("1", "2", "3"))
	assert.True(t, VersionBetween("1", "2", ""))
	assert.True(t, VersionBetween("", "2", "3"))

	assert.True(t, VersionBetween("2", "2", "2"))
	assert.True(t, VersionBetween("2", "2", ""))
	assert.True(t, VersionBetween("", "2", "2"))

}
