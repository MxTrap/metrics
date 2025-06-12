package utils

import "fmt"

func formatFlagValue(val string) string {
	if val == "" {
		return "N/A"
	}
	return val
}

func PrintBuildFlags(buildDate string, buildCommit string, buildVersion string) {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		formatFlagValue(buildVersion),
		formatFlagValue(buildDate),
		formatFlagValue(buildCommit),
	)
}
