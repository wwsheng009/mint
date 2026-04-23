package e2e

import "strconv"

func parseFixtureInt(value string) (int, error) {
	return strconv.Atoi(value)
}
