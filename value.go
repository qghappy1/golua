package golua

import (
	"fmt"
	"golua/compiler"
	"golua/number"
)

/* basic types */
const (
	LUA_TNONE = iota - 1 // -1
	LUA_TNIL
	LUA_TBOOLEAN
	LUA_TNUMBER
	LUA_TSTRING
	LUA_TTABLE
	LUA_TCLOSURE
	LUA_TUSERDATA
	LUA_TSTATE
)

var luaValueTypeNames = [8]string{"nil", "boolean", "number", "string", "table", "closure", "userdata", "state"}

type LuaValueType int

func (vt LuaValueType) String() string {
	return luaValueTypeNames[int(vt)]
}

type LuaValue interface {
	String() string
	Type() LuaValueType
	Len() int
}

// 空类型
type LuaNilType struct{}

func (nl *LuaNilType) String() string     { return "nil" }
func (nl *LuaNilType) Type() LuaValueType { return LUA_TNIL }
func (nl *LuaNilType) Len() int           { return 0 }

var LuaNil = LuaValue(&LuaNilType{})

// bool类型
type LuaBool bool

func (bl LuaBool) String() string {
	if bool(bl) {
		return "true"
	}
	return "false"
}
func (bl LuaBool) Type() LuaValueType { return LUA_TBOOLEAN }
func (bl LuaBool) Len() int           { return 0 }

var LuaTrue = LuaBool(true)
var LuaFalse = LuaBool(false)

// 字符串类型
type LuaString string

func (st LuaString) String() string     { return string(st) }
func (st LuaString) Type() LuaValueType { return LUA_TSTRING }
func (st LuaString) Len() int           { return len(st) }

// 数字类型
type LuaNumber float64

func (nm LuaNumber) String() string {
	i, ok := floatToInteger(nm)
	if ok {
		return fmt.Sprint(i)
	}
	return fmt.Sprint(float64(nm))
}
func (nm LuaNumber) Type() LuaValueType { return LUA_TNUMBER }
func (nm LuaNumber) Len() int           { return 0 }

func floatToInteger(n LuaNumber) (int64, bool) {
	return number.FloatToInteger(float64(n))
}

// 表类型
type LuaTable struct {
	metatable *LuaTable
	arr       []LuaValue
	map_      map[LuaValue]LuaValue
	keys      map[LuaValue]LuaValue // used by next()
	changed   bool                  // used by next()
}

func (tb *LuaTable) String() string     { return fmt.Sprintf("table:%p", tb) }
func (tb *LuaTable) Type() LuaValueType { return LUA_TTABLE }
func (tb *LuaTable) Len() int           { return len(tb.arr) + len(tb.map_) }

// lua栈
type LuaState struct {
	registry *LuaTable
	stack    *luaStack
	/* coroutine */
	coStatus int
	coCaller *LuaState
	coChan   chan int
}

func (ls *LuaState) String() string     { return fmt.Sprintf("state:%p", ls) }
func (ls *LuaState) Type() LuaValueType { return LUA_TSTATE }
func (ls *LuaState) Len() int           { return 0 }

// 用户数据
type LuaUserData struct {
	Value     interface{}
	Env       *LuaTable
	Metatable *LuaTable
}

func (ud *LuaUserData) String() string     { return fmt.Sprintf("userdata:%p", ud) }
func (ud *LuaUserData) Type() LuaValueType { return LUA_TUSERDATA }
func (ud *LuaUserData) Len() int           { return 0 }

type upvalue struct {
	val *LuaValue
}

// go function
type GoFunction func(*LuaState) int

// lua闭包
type LuaClosure struct {
	proto  *compiler.FunctionProto // lua Closure
	goFunc GoFunction              // go Closure
	upvals []*upvalue
}

func (ud *LuaClosure) String() string     { return fmt.Sprintf("closure:%p", ud) }
func (ud *LuaClosure) Type() LuaValueType { return LUA_TCLOSURE }
func (ud *LuaClosure) Len() int           { return 0 }

func convertToBoolean(val LuaValue) bool {
	switch val.Type() {
	case LUA_TNIL:
		return false
	case LUA_TBOOLEAN:
		return bool(val.(LuaBool))
	default:
		return true
	}
}

// http://www.lua.org/manual/5.3/manual.html#3.4.3
func convertToFloat(val LuaValue) (float64, bool) {
	switch val.Type() {
	case LUA_TNUMBER:
		return float64(val.(LuaNumber)), true
	case LUA_TSTRING:
		return number.ParseFloat(string(val.(LuaString)))
	default:
		return 0, false
	}
}

// http://www.lua.org/manual/5.3/manual.html#3.4.3
func convertToInteger(val LuaValue) (int64, bool) {
	switch val.Type() {
	case LUA_TNUMBER:
		return floatToInteger(val.(LuaNumber))
	case LUA_TSTRING:
		return _stringToInteger(string(val.(LuaString)))
	default:
		return 0, false
	}
}

func _stringToInteger(s string) (int64, bool) {
	if i, ok := number.ParseInteger(s); ok {
		return i, true
	}
	if f, ok := number.ParseFloat(s); ok {
		return number.FloatToInteger(f)
	}
	return 0, false
}
