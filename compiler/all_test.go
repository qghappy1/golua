package compiler

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func printProto(proto *FunctionProto, depth int) {
	t := ""
	for i := 0; i < depth; i++ {
		t = t + "\t"
	}
	fmt.Printf("%vSource:%v\n", t, proto.Source)
	fmt.Printf("%vLineDefined:%v\n", t, proto.LineDefined)
	fmt.Printf("%vLastLineDefined:%v\n", t, proto.LastLineDefined)
	fmt.Printf("%vNumParams:%v\n", t, proto.NumParams)
	fmt.Printf("%vIsVararg:%v\n", t, proto.IsVararg)

	fmt.Printf("%vMaxStackSize:%v\n", t, proto.MaxStackSize)
	fmt.Printf("%vCode:%v\n", t, proto.Code)
	fmt.Printf("%vConstants:%v\n", t, proto.Constants)
	fmt.Printf("%vUpvalues:%v\n", t, proto.Upvalues)
	fmt.Printf("%vDbgSourcePositions:%v\n", t, proto.DbgSourcePositions)
	fmt.Printf("%vDbgLocVars:%+v\n", t, proto.DbgLocVars)

	fmt.Printf("%vDbgCalls:%v\n", t, proto.DbgCalls)
	fmt.Printf("%vDbgUpvalues:%v\n", t, proto.DbgUpvalues)

	for _, p := range proto.Protos {
		printProto(p, depth+1)
	}
}

// go test -v
func Test_All(t *testing.T) {
	//str := "local a, b = 1, 2 \n a = b + 1"
	//proto := Compile(str, "str")
	//printProto(proto, 0)
	filename := "main.lua"
	if data, err := ioutil.ReadFile(filename); err == nil {
		proto := Compile(data, "@"+filename)
		printProto(proto, 0)
	}
}
