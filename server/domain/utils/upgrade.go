package utils

import "strings"

// Upgradeable If the versionlist contains a version number higher than version, then return true
func Upgradeable(version string, versionlist []string) bool {
	upgradeablelist := Upgradeablelist(version, versionlist)
	return len(upgradeablelist) > 0
}

// Upgradeablelist If the versionlist contains a version number higher than version, then return the upgradeable list
func Upgradeablelist(version string, versionlist []string) []string {
	upgradeablelist := make([]string, 0)

	for _, item := range versionlist {
		if compareVersion(item, version) > 0 && typeVersion(item, version) {
			upgradeablelist = append(upgradeablelist, item)
		}
	}
	return upgradeablelist
}

// compareVersion compare version1 and version2
// If version1 is larger, return 1
// If version2 is larger, return -1
// otherwise return 0
func compareVersion(version1, version2 string) int {
	n, m := len(version1), len(version2)
	i, j := 0, 0
	for i < n || j < m {
		x := 0
		for ; i < n && version1[i] != '.'; i++ {
			x = x*10 + int(version1[i]-'0')
		}
		i++
		y := 0
		for ; j < m && version2[j] != '.'; j++ {
			y = y*10 + int(version2[j]-'0')
		}
		j++
		if x > y {
			return 1
		}
		if x < y {
			return -1
		}
	}
	return 0
}

// typeVersion Determine whether version1 and version2 are of the same type.
// Examples:
// - v1.10.0 & v1.9.1 is true
// - v1.10.0-fedlcm-v0.3.0 & v1.9.1-fedlcm-v0.2.0 is true
// - v1.10.0-fedlcm-v0.3.0 & v1.10.0 is false
func typeVersion(version1, version2 string) bool {
	return len(strings.Split(version1, "-fedlcm-")) == len(strings.Split(version2, "-fedlcm-"))
}
