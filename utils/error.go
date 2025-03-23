package utils

import "fmt"

func EmptyOrError(err error) string {

	if err == nil {
		return ""
	}

	return fmt.Sprintf("%s", err.Error())
}
