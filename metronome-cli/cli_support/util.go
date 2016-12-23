package cli

// In - checks whether the string is in the array
func In(val string, targ []string) bool {
	for _, cur := range targ {
		if cur == val {
			return true
		}
	}
	return false
}

