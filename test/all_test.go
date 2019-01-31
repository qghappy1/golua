package compiler

import (
	"fmt"
	"testing"
	"io/ioutil"
	"github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	"strings"
)

func printProto(proto *lua.FunctionProto, depth int)  {
	t := ""
	for i := 0; i<depth; i++ {
		t = t + "\t"
	}
	fmt.Printf("%vSource:%v\n", t, proto.SourceName)
	fmt.Printf("%vLineDefined:%v\n", t, proto.LineDefined)
	fmt.Printf("%vLastLineDefined:%v\n", t, proto.LastLineDefined)
	fmt.Printf("%vNumParams:%v\n", t, proto.NumParameters)
	fmt.Printf("%vIsVararg:%v\n", t, proto.IsVarArg)

	fmt.Printf("%vMaxStackSize:%v\n", t, proto.NumUsedRegisters)
	fmt.Printf("%vCode:%v\n", t, proto.Code)
	fmt.Printf("%vConstants:%v\n", t, proto.Constants)
	fmt.Printf("%vDbgSourcePositions:%v\n", t, proto.DbgSourcePositions)
	for _, v := range(proto.DbgLocals) {
		fmt.Printf("%vDbgLocVars:%+v\n", t, v)
	}
	fmt.Printf("%vDbgCalls:%v\n", t, proto.DbgCalls)
	fmt.Printf("%vDbgUpvalues:%v\n", t, proto.DbgUpvalues)

	for _, p := range(proto.FunctionPrototypes) {
		printProto(p, depth+1)
	}
}

// go test -v
func Test_All(t *testing.T){
	//str := "local a, b = 1, 2 \n a = b + 1"
	//proto := Compile(str, "str")
	//printProto(proto, 0)
	filename := "main.lua"
	if data, err := ioutil.ReadFile(filename); err == nil {
		chunk, err := parse.Parse(strings.NewReader(string(data)), "@"+filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		if proto, err := lua.Compile(chunk, "@"+filename); err == nil {
			printProto(proto, 0)
		}
	}

}