package ids

import (
	"github.com/jaevor/go-nanoid"
)

var Alphabet = [32]byte{
	'a', 'b', 'c', 'd',
	'e', 'f', 'g', 'h',
	'j', 'k', 'm', 'n',
	'p', 'q', 'r', 's',
	't', 'u', 'v', 'w',
	'x', 'y', 'z',
	'2', '3', '4', '5',
	'6', '7', '8', '9',
	'-',
}

func Generator(length int) (func() string, error) {
	g, err := nanoid.CustomASCII(string(Alphabet[:]), length)
	return g, err
}
