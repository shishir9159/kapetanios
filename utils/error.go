package utils

func EmptyOrError(err error) string {

	if err == nil {
		return ""
	}

	return err.Error()
}
