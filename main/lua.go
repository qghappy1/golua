package main

import (
	"golua"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		ls := golua.NewLuaState()
		ls.OpenLibs()
		ls.LoadFile(os.Args[1])
		ls.PCall(0, 0, 0)
	}
}
