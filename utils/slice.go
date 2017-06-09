package utils

import (
	"sort"
)

/**

 */
func InStringSlice(slice []string, needle string) bool {
	sort.Strings(slice)
	/** search returns length if not found */
	i := sort.Search(len(slice),
		func(i int) bool { return slice[i] >= needle })
	if i < len(slice) && slice[i] == needle {
		return true
	}
	return false
}
