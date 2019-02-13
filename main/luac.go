package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"golua/compiler"
	"golua"
)

func main() {
	if len(os.Args)>1 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		proto := compiler.Compile(data, "@")
		list(proto)
		return
	}
	fmt.Println("输入参数错误")
}

func list(f *compiler.FunctionProto) {
	printHeader(f)
	printCode(f)
	printDetail(f)
	for _, p := range f.Protos {
		list(p)
	}
}

func printHeader(f *compiler.FunctionProto) {
	funcType := "main"
	if f.LineDefined > 0 {
		funcType = "function"
	}

	varargFlag := ""
	if f.IsVararg > 0 {
		varargFlag = "+"
	}

	fmt.Printf("\n%s <%s:%d,%d> (%d instructions)\n",
		funcType, f.Source, f.LineDefined, f.LastLineDefined, len(f.Code))

	fmt.Printf("%d%s params, %d slots, %d upvalues, ",
		f.NumParams, varargFlag, f.MaxStackSize, len(f.Upvalues))

	fmt.Printf("%d locals, %d constants, %d functions\n",
		len(f.DbgLocVars), len(f.Constants), len(f.Protos))
}

func printCode(f *compiler.FunctionProto) {
	for pc, c := range f.Code {
		line := "-"
		if len(f.DbgSourcePositions) > 0 {
			line = fmt.Sprintf("%d", f.DbgSourcePositions[pc])
		}

		i := golua.Instruction(c)
		fmt.Printf("\t%d\t[%s]\t%s \t", pc+1, line, i.OpName())
		printOperands(i)
		fmt.Printf("\n")
	}
}

func printOperands(i golua.Instruction) {
	switch i.OpMode() {
	case compiler.IABC:
		a, b, c := i.ABC()

		fmt.Printf("%d", a)
		if i.BMode() != compiler.OpArgN {
			if b > 0xFF {
				fmt.Printf(" %d", -1-b&0xFF)
			} else {
				fmt.Printf(" %d", b)
			}
		}
		if i.CMode() != compiler.OpArgN {
			if c > 0xFF {
				fmt.Printf(" %d", -1-c&0xFF)
			} else {
				fmt.Printf(" %d", c)
			}
		}
	case compiler.IABx:
		a, bx := i.ABx()

		fmt.Printf("%d", a)
		if i.BMode() == compiler.OpArgK {
			fmt.Printf(" %d", -1-bx)
		} else if i.BMode() == compiler.OpArgU {
			fmt.Printf(" %d", bx)
		}
	case compiler.IAsBx:
		a, sbx := i.AsBx()
		fmt.Printf("%d %d", a, sbx)
	case compiler.IAx:
		ax := i.Ax()
		fmt.Printf("%d", -1-ax)
	}
}

func printDetail(f *compiler.FunctionProto) {
	fmt.Printf("constants (%d):\n", len(f.Constants))
	for i, k := range f.Constants {
		fmt.Printf("\t%d\t%s\n", i+1, constantToString(k))
	}

	fmt.Printf("locals (%d):\n", len(f.DbgLocVars))
	for i, locVar := range f.DbgLocVars {
		fmt.Printf("\t%d\t%s\t%d\t%d\n",
			i, locVar.VarName, locVar.StartPC+1, locVar.EndPC+1)
	}

	fmt.Printf("upvalues (%d):\n", len(f.Upvalues))
	for i, upval := range f.Upvalues {
		fmt.Printf("\t%d\t%s\t%d\t%d\n",
			i, upvalName(f, i), upval.Instack, upval.Idx)
	}
}

func constantToString(k interface{}) string {
	switch k.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("%t", k)
	case float64:
		return fmt.Sprintf("%g", k)
	case int64:
		return fmt.Sprintf("%d", k)
	case string:
		return fmt.Sprintf("%q", k)
	default:
		return "?"
	}
}

func upvalName(f *compiler.FunctionProto, idx int) string {
	if len(f.DbgUpvalues) > 0 {
		return f.DbgUpvalues[idx]
	}
	return "-"
}