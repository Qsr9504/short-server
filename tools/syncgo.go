package tools

import "fmt"

// GoSafe runs the given fn using another goroutine, recovers if fn panics.
func GoSafe(fn func()) {
	go RunSafe(fn)
}

func Recover() {
	if msg := recover(); msg != nil {
		fmt.Println(msg)
	}
}

// RunSafe runs the given fn, recovers if fn panics.
func RunSafe(fn func()) {
	defer Recover()

	fn()
}
