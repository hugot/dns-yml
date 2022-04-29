package util

func SliceContainsString(slice []string, thing string) bool {
	for _, i := range slice {
		if thing == i {
			return true
		}
	}

	return false
}
