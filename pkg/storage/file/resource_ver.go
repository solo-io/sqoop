package file

import (
	"fmt"
	"strconv"
)

func newOrIncrementResourceVer(resourceVersion string) string {
	curr, err := strconv.Atoi(resourceVersion)
	if err != nil {
		curr = 1
	}
	return fmt.Sprintf("%v", curr)
}

func lessThan(rv1, rv2 string) bool {
	return newOrIncrementResourceVer(rv1) < newOrIncrementResourceVer(rv2)
}
