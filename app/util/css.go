package util

import (
	"strconv"
	"strings"
)

func PxToInt(v string) (int, error) {
	numS := strings.Split(v, "px")[0]
	return strconv.Atoi(numS)
}
