package ops

// Ternary is the standard ternary operator which returns posVal
// if the boolean condition is true, otherwise negVal.
func Ternary(cond bool, posVal interface{}, negVal interface{}) interface{} {
	if cond {
		return posVal
	}
	return negVal
}
