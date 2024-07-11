package restapi

import (
	"net/url"
	"strconv"
)

// IsUrl простая валидация url
func IsUrl(str string) bool {
    u, err := url.Parse(str)
    return err == nil && u.Scheme != "" && u.Host != ""
}

// IsInt простая валидация int
func IsInt(str string) bool {
	if _, err := strconv.Atoi(str); err != nil {
		return false
	}
	return true
}