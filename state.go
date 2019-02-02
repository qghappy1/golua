package golua

import (
	"golua/compiler"
)

const LUA_MINSTACK = 20
const LUAI_MAXSTACK = 1000000
const LUA_REGISTRYINDEX = -LUAI_MAXSTACK - 1000
const LUA_RIDX_MAINTHREAD LuaNumber = 1
const LUA_RIDX_GLOBALS LuaNumber = 2
const LUA_MULTRET = -1

func NewLuaState() *LuaState {
	ls := &LuaState{}
	registry := newLuaTable(8, 0)
	registry.Set(LUA_RIDX_MAINTHREAD, ls)
	registry.Set(LUA_RIDX_GLOBALS, newLuaTable(0, 20))
	ls.registry = registry
	ls.pushLuaStack(newLuaStack(LUA_MINSTACK, ls))
	return ls
}

func (ls *LuaState) pushLuaStack(stack *luaStack) {
	stack.prev = ls.stack
	ls.stack = stack
}

func (ls *LuaState) popLuaStack() {
	stack := ls.stack
	ls.stack = stack.prev
	stack.prev = nil
}

func (ls *LuaState) isMainThread() bool {
	return ls.registry.Get(LUA_RIDX_MAINTHREAD) == ls
}

func (ls *LuaState) Push(value LuaValue) {
	ls.stack.push(value)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushcfunction
func (ls *LuaState) PushGoFunction(f GoFunction) {
	ls.stack.push(newGoClosure(f, 0))
}

// [-n, +1, m]
// http://www.lua.org/manual/5.3/manual.html#lua_pushcClosure
func (ls *LuaState) PushGoClosure(f GoFunction, n int) {
	closure := newGoClosure(f, n)
	for i := n; i > 0; i-- {
		val := ls.stack.pop()
		closure.upvals[i-1] = &upvalue{&val}
	}
	ls.stack.push(closure)
}

func (ls *LuaState) SetGlobal(name string, v LuaValue) {
	t := ls.registry.Get(LUA_RIDX_GLOBALS)
	luaSetTable_(ls, t, LuaString(name), v, false)
}

// [-0, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_register
func (ls *LuaState) Register(name string, f GoFunction) {
	ls.PushGoFunction(f)
	luaSetGlobal(ls, LuaString(name))
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_load
func (ls *LuaState) Load(chunk []byte, chunkName string) int {
	proto := compiler.Compile(chunk, chunkName)
	c := newLuaClosure(proto)
	ls.stack.push(c)
	if len(proto.Upvalues) > 0 {
		env := ls.registry.Get(LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{&env}
	}
	return LUA_OK
}

// [-(nargs+1), +nresults, e]
// http://www.lua.org/manual/5.3/manual.html#lua_call
func (ls *LuaState) Call(nArgs, nResults int) {
}
