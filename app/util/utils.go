package util

import (
	"strconv"
	"strings"
)

// StrToIntSlice если ошибка, возращает ее и частично заполенный срез
func StrToIntSlice(str, delim string) ([]int, error) {
	var err error
	nums := make([]int, 0, len(str))

	for _, el := range strings.Split(str, delim) {
		num, locErr := strconv.Atoi(el)
		if locErr != nil {
			err = locErr
			continue
		}

		nums = append(nums, num)
	}

	return nums, err
}
