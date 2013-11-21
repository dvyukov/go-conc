func foo(v int) int {
	if v == 42 {
		select{}
	} else {
		return v
	}
}