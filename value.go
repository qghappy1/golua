package golua

import (
	"fmt"
	"golua/number"
	"golua/compiler"
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
}

// 空类型
type LuaNilType struct{}
func (nl *LuaNilType) String() string			{ return "nil" }
func (nl *LuaNilType) Type() LuaValueType		{ return LUA_TNIL }

var LuaNil = LuaValue(&LuaNilType{})

// bool类型
type LuaBool bool
func (bl LuaBool) String() string 				{ if bool(bl) { return "true" }; return "false"}
func (bl LuaBool) Type() LuaValueType			{ return LUA_TBOOLEAN }

var LTrue = LuaBool(true)
var LFalse = LuaBool(false)

// 字符串类型
type LuaString string
func (st LuaString) String() string				{ return string(st) }
func (st LuaString) Type() LuaValueType			{ return LUA_TSTRING }

// 数字类型
type LuaNumber float64
func (nm LuaNumber) String() string {
	i, ok := floatToInteger(nm)
	if ok {
		return fmt.Sprint(i)
	}
	return fmt.Sprint(float64(nm))
}
func (nm LuaNumber) Type() LuaValueType			{ return LUA_TNUMBER }

func floatToInteger(n LuaNumber)  (int64, bool) {
	return number.FloatToInteger(float64(n))
}

// 表类型
type LuaTable struct {
	metatable *LuaTable
	arr       []LuaValue
	_map      map[LuaValue]LuaValue
	keys      map[LuaValue]LuaValue // used by next()
	lastKey   LuaValue              // used by next()
	changed   bool                  // used by next()
}
func (tb *LuaTable) String() string				{ return fmt.Sprintf("table:%p", tb) }
func (tb *LuaTable) Type() LuaValueType			{ return LUA_TTABLE }


// lua栈
type LuaState struct {
	registry *LuaTable
	//stack    *luaStack
	/* coroutine */
	coStatus int
	coCaller *LuaState
	coChan   chan int
}
func (ls *LuaState) String() string				{ return fmt.Sprintf("state:%p", ls) }
func (ls *LuaState) Type() LuaValueType			{ return LUA_TSTATE }

// 用户数据
type LuaUserData struct {
	Value     	interface{}
	Env       	*LuaTable
	Metatable 	*LuaTable
}
func (ud *LuaUserData) String() string			{ return fmt.Sprintf("userdata:%p", ud) }
func (ud *LuaUserData) Type() LuaValueType		{ return LUA_TUSERDATA }


type upvalue struct {
	val *LuaValue
}
// go function
type GoFunction func(*LuaState) int
// lua闭包
type LuaClosure struct {
	proto  *compiler.FunctionProto 		// lua Closure
	goFunc GoFunction          			// go Closure
	upvals []*upvalue
}
func (ud *LuaClosure) String() string			{ return fmt.Sprintf("closure:%p", ud) }
func (ud *LuaClosure) Type() LuaValueType		{ return LUA_TCLOSURE }

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

/* metatable */

//func getMetatable(val LuaValue, ls *LuaState) *LuaTable {
//	if t, ok := val.(*LuaTable); ok {
//		return t.metatable
//	}
//	if u, ok := val.(*LuaUserData); ok {
//		return u.Metatable
//	}
//	key := fmt.Sprintf("_MT%d", typeOf(val))
//	if mt := ls.registry.get(key); mt != nil {
//		return mt.(*LuaTable)
//	}
//	return nil
//}
//
//func setMetatable(val LuaValue, mt *LuaTable, ls *LuaState) {
//	if t, ok := val.(*LuaTable); ok {
//		t.metatable = mt
//		return
//	}
//	if u, ok := val.(*LuaUserData); ok {
//		u.Metatable = mt
//		return
//	}
//	key := fmt.Sprintf("_MT%d", typeOf(val))
//	ls.registry.put(key, mt)
//}
//
//func getMetafield(val LuaValue, fieldName string, ls *LuaState) LuaValue {
//	if mt := getMetatable(val, ls); mt != nil {
//		return mt.get(fieldName)
//	}
//	return nil
//}
//
//func callMetamethod(a, b LuaValue, mmName string, ls *LuaState) (LuaValue, bool) {
//	var mm LuaValue
//	if mm = getMetafield(a, mmName, ls); mm == nil {
//		if mm = getMetafield(b, mmName, ls); mm == nil {
//			return nil, false
//		}
//	}
//
//	ls.stack.check(4)
//	ls.stack.push(mm)
//	ls.stack.push(a)
//	ls.stack.push(b)
//	ls.Call(2, 1)
//	return ls.stack.pop(), true
//}