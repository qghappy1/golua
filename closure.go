package golua

import (
	"golua/compiler"
)

func newLuaClosure(proto *compiler.FunctionProto) *LuaClosure {
	c := &LuaClosure{proto: proto}
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}

func newGoClosure(f GoFunction, nUpvals int) *LuaClosure {
	c := &LuaClosure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}
