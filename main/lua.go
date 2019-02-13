package main

import (
	"golua"
	"os"
	"fmt"
)

func main() {
	if len(os.Args) > 1 {
		ls := golua.NewLuaState()
		ls.OpenLibs()
		ls.LoadFile(os.Args[1])
		ls.Call(0, 0)
		//if err := ls.PCall(0, 0, 0); err != nil {
		//	fmt.Println(err)
		//}
		fmt.Println("")
	}
}
