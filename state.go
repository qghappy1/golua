package golua

import (
	"fmt"
	"golua/compiler"
	"io/ioutil"
)

const LUA_MINSTACK = 20
const LUAI_MAXSTACK = 1000000
const LUA_REGISTRYINDEX = -LUAI_MAXSTACK - 1000
const LUA_RIDX_MAINTHREAD LuaNumber = 1
const LUA_RIDX_GLOBALS LuaNumber = 2
const LUA_MULTRET = -1

/* Debug {{{ */

type luaDebug struct {
	stack           *luaStack
	Name            string
	What            string
	Source          string
	CurrentLine     int
	NUpvalues       int
	LineDefined     int
	LastLineDefined int
}

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

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#lua_newtable
func (ls *LuaState) NewTable() *LuaTable {
	return ls.createTable(0, 0)
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#lua_createtable
func (ls *LuaState) createTable(nArr, nRec int) *LuaTable {
	t := newLuaTable(nArr, nRec)
	ls.stack.push(t)
	return t
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkany
func (ls *LuaState) CheckAny(idx int) LuaValue {
	return ls.stack.get(idx)
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkinteger
func (ls *LuaState) CheckInteger(idx int) int64 {
	i, ok := luaToIntegerX(ls, idx)
	if !ok {
		ls.raiseError1("stack idx:%v value:%v not int64", idx, ls.stack.get(idx))
	}
	return i
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checknumber
func (ls *LuaState) CheckNumber(idx int) float64 {
	f, ok := luaToNumberX(ls, idx)
	if !ok {
		ls.raiseError1("stack idx:%v value:%v not float64", idx, ls.stack.get(idx))
	}
	return f
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkstring
// http://www.lua.org/manual/5.3/manual.html#luaL_checklstring
func (ls *LuaState) CheckString(idx int) string {
	s, ok := luaToStringX(ls, idx)
	if !ok {
		ls.raiseError1("stack idx:%v value:%v not string", idx, ls.stack.get(idx))
	}
	return s
}

func (ls *LuaState) CheckTable(idx int) *LuaTable {
	v := ls.stack.get(idx)
	if tb, ok := v.(*LuaTable); ok {
		return tb
	}
	ls.raiseError1("stack idx:%v value:%v not table", idx, ls.stack.get(idx))
	return nil
}

func (ls *LuaState) CheckClosure(idx int) *LuaClosure {
	v := ls.stack.get(idx)
	if c, ok := v.(*LuaClosure); ok {
		return c
	}
	ls.raiseError1("stack idx:%v value:%v not closure", idx, ls.stack.get(idx))
	return nil
}

func (ls *LuaState) CheckUserData(idx int) *LuaUserData {
	v := ls.stack.get(idx)
	if ud, ok := v.(*LuaUserData); ok {
		return ud
	}
	ls.raiseError1("stack idx:%v value:%v not userdata", idx, ls.stack.get(idx))
	return nil
}

func (ls *LuaState) SetGlobal(name string, v LuaValue) {
	t := ls.registry.Get(LUA_RIDX_GLOBALS)
	luaSetTable_(ls, t, LuaString(name), v, false)
}

// [-0, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_register
func (ls *LuaState) Register(name string, f GoFunction) {
	ls.PushGoFunction(f)
	luaSetGlobal(ls, name)
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

// [-0, +?, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_dofile
func (ls *LuaState) DoFile(filename string) bool {
	return ls.LoadFile(filename) == LUA_OK &&
		ls.PCall(0, LUA_MULTRET, 0) == LUA_OK
}

// [-0, +?, –]
// http://www.lua.org/manual/5.3/manual.html#luaL_dostring
func (ls *LuaState) DoString(str string) bool {
	return ls.LoadString(str) == LUA_OK &&
		ls.PCall(0, LUA_MULTRET, 0) == LUA_OK
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_loadfile
func (ls *LuaState) LoadFile(filename string) int {
	return ls.LoadFileX(filename)
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_loadfilex
func (ls *LuaState) LoadFileX(filename string) int {
	if data, err := ioutil.ReadFile(filename); err == nil {
		return ls.Load(data, "@"+filename)
	}
	return LUA_ERRFILE
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#luaL_loadstring
func (ls *LuaState) LoadString(s string) int {
	return ls.Load([]byte(s), s)
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_requiref
func (ls *LuaState) RequireF(modname string, openf GoFunction, glb bool) {
	luaGetSubTable(ls, LUA_REGISTRYINDEX, "_LOADED")
	luaGetField(ls, -1, modname) /* LOADED[modname] */
	if !luaToBoolean(ls, -1) {   /* package not already loaded? */
		luaPop(ls, 1) /* remove field */
		ls.PushGoFunction(openf)
		ls.Push(LuaString(modname))  /* argument to open function */
		ls.Call(1, 1)                /* call 'openf' to open module */
		luaPushValue(ls, -1)         /* make copy of module (call result) */
		luaSetField(ls, -3, modname) /* _LOADED[modname] = module */
	}
	luaRemove(ls, -2) /* remove _LOADED table */
	if glb {
		luaPushValue(ls, -1)      /* copy of module */
		luaSetGlobal(ls, modname) /* _G[modname] = module */
	}
}

// [-(nargs+1), +nresults, e]
// http://www.lua.org/manual/5.3/manual.html#lua_call
func (ls *LuaState) Call(nArgs, nResults int) {
	// todo
}

// Calls a function in protected mode.
// http://www.lua.org/manual/5.3/manual.html#lua_pcall
func (ls *LuaState) PCall(nArgs, nResults, msgh int) (status int) {
	// todo
	status = LUA_OK
	return
}

// debug error
func (ls *LuaState) raiseError1(format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	message = fmt.Sprintf("%v %v", ls.where(0, true), message)
	// ls.stack.push(LuaString(message))
	panic(message)
}

func (ls *LuaState) raiseError(level int, format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	if level > 0 {
		message = fmt.Sprintf("%v %v", ls.where(level-1, true), message)
	}
	// ls.stack.push(LuaString(message))
	panic(message)
}

func (ls *LuaState) getDebug(level int) *luaDebug {
	stack := ls.stack
	for ; level > 0 && stack != nil; stack = stack.prev {
		level--
		// todo tail call
	}
	if level == 0 && stack != nil {
		return &luaDebug{stack: stack}
	} else if level < 0 {
		return &luaDebug{stack: ls.stack}
	}
	return nil
}

func (ls *LuaState) where(level int, skipg bool) string {
	dbg := ls.getDebug(level)
	if dbg == nil {
		return ""
	}
	stack := dbg.stack
	sourcename := "[G]"
	var proto *compiler.FunctionProto = nil
	if stack.closure != nil {
		proto = stack.closure.proto
	}
	if proto != nil {
		sourcename = proto.Source
	} else if skipg {
		return ls.where(level+1, skipg)
	}
	line := ""
	if proto != nil {
		line = fmt.Sprintf("%v:", proto.DbgSourcePositions[dbg.stack.pc-1])
	}
	return fmt.Sprintf("%v:%v", sourcename, line)
}

// 退出当前域时关闭外部变量表
func (ls *LuaState) closeUpvalues(a int) {
	for i, openuv := range ls.stack.openuvs {
		if i >= a-1 {
			val := *openuv.val
			openuv.val = &val
			delete(ls.stack.openuvs, i)
		}
	}
}
