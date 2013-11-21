package main

func main() {
	_ = foo()
}

// STARTMAIN OMIT
func foo() *int {
	type T struct {
		x [1 << 24]byte
		i int
	}
	var x *T
	return &x.i // HL
}

// STOPMAIN OMIT
