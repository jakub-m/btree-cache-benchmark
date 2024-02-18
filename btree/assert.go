package btree

import "fmt"

func assert(condition bool, message ...any) {
	if !condition {
		if len(message) == 0 {
			panic("assertion failed")
		} else {
			format := fmt.Sprint(message[0])
			panic(fmt.Sprintf(format, message[1:]...))
		}
	}
}
