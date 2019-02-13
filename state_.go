package golua

import (
	"fmt"
	"luago/number"
	"strings"
)

// [-n, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pop
func luaPop(ls *LuaState, n int) {
	for i := 0; i < n; i++ {
		ls.stack.pop()
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_gettop
func luaGetTop(ls *LuaState) int {
	return ls.stack.top
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_absindex
func luaAbsIndex(ls *LuaState, idx int) int {
	return ls.stack.absIndex(idx)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_checkstack
func luaCheckStack(ls *LuaState, n int) bool {
	ls.stack.check(n)
	return true
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_type
func luaType(ls *LuaState, idx int) LuaValueType {
	if ls.stack.isValid(idx) {
		val := ls.stack.get(idx)
		return val.Type()
	}
	return LUA_TNONE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isinteger
func luaIsInteger(ls *LuaState, idx int) bool {
	val := ls.stack.get(idx)
	switch val.Type() {
	case LUA_TNUMBER:
		_, ok := floatToInteger(val.(LuaNumber))
		return ok
	default:
		return false
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_copy
func luaCopy(ls *LuaState, fromIdx, toIdx int) {
	val := ls.stack.get(fromIdx)
	ls.stack.set(toIdx, val)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushvalue
func luaPushValue(ls *LuaState, idx int) {
	val := ls.stack.get(idx)
	ls.stack.push(val)
}

// [-1, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_replace
func luaReplace(ls *LuaState, idx int) {
	val := ls.stack.pop()
	ls.stack.set(idx, val)
}

// [-1, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_insert
func luaInsert(ls *LuaState, idx int) {
	luaRotate(ls, idx, 1)
}

// [-1, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_remove
func luaRemove(ls *LuaState, idx int) {
	luaRotate(ls, idx, -1)
	luaPop(ls, 1)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_rotate
func luaRotate(ls *LuaState, idx, n int) {
	t := ls.stack.top - 1           /* end of stack segment being rotated */
	p := ls.stack.absIndex(idx) - 1 /* start of segment */
	var m int                       /* end of prefix */
	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}
	ls.stack.reverse(p, m)   /* reverse the prefix with length 'n' */
	ls.stack.reverse(m+1, t) /* reverse the suffix */
	ls.stack.reverse(p, t)   /* reverse the entire segment */
}

// [-?, +?, –]
// http://www.lua.org/manual/5.3/manual.html#lua_settop
func luaSetTop(ls *LuaState, idx int) {
	newTop := ls.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow!")
	}

	n := ls.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			ls.stack.pop()
		}
	} else if n < 0 {
		for i := 0; i > n; i-- {
			ls.stack.push(nil)
		}
	}
}

// [-?, +?, –]
// http://www.lua.org/manual/5.3/manual.html#lua_xmove
func luaXMove(ls *LuaState, to *LuaState, n int) {
	vals := ls.stack.popN(n)
	to.stack.pushN(vals, n)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_rawlen
func luaRawLen(ls *LuaState, idx int) int {
	val := ls.stack.get(idx)
	return val.Len()
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnone
func luaIsNone(ls *LuaState, idx int) bool {
	return luaType(ls, idx) == LUA_TNONE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnil
func luaIsNil(ls *LuaState, idx int) bool {
	return luaType(ls, idx) == LUA_TNIL
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnoneornil
func luaIsNoneOrNil(ls *LuaState, idx int) bool {
	return luaType(ls, idx) <= LUA_TNIL
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isboolean
func luaIsBoolean(ls *LuaState, idx int) bool {
	return luaType(ls, idx) == LUA_TBOOLEAN
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_istable
func luaIsTable(ls *LuaState, idx int) bool {
	return luaType(ls, idx) == LUA_TTABLE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isfunction
func luaIsFunction(ls *LuaState, idx int) bool {
	return luaType(ls, idx) == LUA_TCLOSURE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isthread
func luaIsThread(ls *LuaState, idx int) bool {
	return luaType(ls, idx) == LUA_TSTATE
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isstring
func luaIsString(ls *LuaState, idx int) bool {
	t := luaType(ls, idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isnumber
func luaIsNumber(ls *LuaState, idx int) bool {
	_, ok := luaToNumberX(ls, idx)
	return ok
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_iscfunction
func luaIsGoFunction(ls *LuaState, idx int) bool {
	val := ls.stack.get(idx)
	if c, ok := val.(*LuaClosure); ok {
		return c.goFunc != nil
	}
	return false
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_toboolean
func luaToBoolean(ls *LuaState, idx int) bool {
	val := ls.stack.get(idx)
	return convertToBoolean(val)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tointeger
func luaToInteger(ls *LuaState, idx int) int64 {
	i, _ := luaToIntegerX(ls, idx)
	return i
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tointegerx
func luaToIntegerX(ls *LuaState, idx int) (int64, bool) {
	val := ls.stack.get(idx)
	return convertToInteger(val)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tonumber
func luaToNumber(ls *LuaState, idx int) float64 {
	n, _ := luaToNumberX(ls, idx)
	return n
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tonumberx
func luaToNumberX(ls *LuaState, idx int) (float64, bool) {
	val := ls.stack.get(idx)
	return convertToFloat(val)
}

// [-0, +0, m]
// http://www.lua.org/manual/5.3/manual.html#lua_tostring
func luaToString(ls *LuaState, idx int) string {
	s, _ := luaToStringX(ls, idx)
	return s
}

func luaToStringX(ls *LuaState, idx int) (string, bool) {
	val := ls.stack.get(idx)
	switch val.Type() {
	case LUA_TSTRING:
		return val.String(), true
	case LUA_TNUMBER:
		s := fmt.Sprintf("%v", val)
		ls.stack.set(idx, LuaString(s))
		return s, true
	default:
		return "", false
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tocfunction
func luaToGoFunction(ls *LuaState, idx int) GoFunction {
	val := ls.stack.get(idx)
	if c, ok := val.(*LuaClosure); ok {
		return c.goFunc
	}
	return nil
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_tothread
func luaToThread(ls *LuaState, idx int) *LuaState {
	val := ls.stack.get(idx)
	if val != nil {
		if ls, ok := val.(*LuaState); ok {
			return ls
		}
	}
	return nil
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_topointer
func luaToPointer(ls *LuaState, idx int) interface{} {
	// todo
	return ls.stack.get(idx)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_rawequal
func luaRawEqual(ls *LuaState, idx1, idx2 int) bool {
	if !ls.stack.isValid(idx1) || !ls.stack.isValid(idx2) {
		return false
	}

	a := ls.stack.get(idx1)
	b := ls.stack.get(idx2)
	return _eq(a, b, nil)
}

// [-0, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_compare
func luaCompare(ls *LuaState, idx1, idx2 int, op CompareOp) bool {
	if !ls.stack.isValid(idx1) || !ls.stack.isValid(idx2) {
		return false
	}

	a := ls.stack.get(idx1)
	b := ls.stack.get(idx2)
	switch op {
	case LUA_OPEQ:
		return _eq(a, b, ls)
	case LUA_OPLT:
		return _lt(a, b, ls)
	case LUA_OPLE:
		return _le(a, b, ls)
	default:
		panic("invalid compare op!")
	}
	return false
}

func _eq(a, b LuaValue, ls *LuaState) bool {
	switch a.Type() {
	case LUA_TNIL:
		return b.Type() == LUA_TNIL
	case LUA_TBOOLEAN:
		ba := bool(a.(LuaBool))
		y, ok := b.(LuaBool)
		return ok && ba == bool(y)
	case LUA_TSTRING:
		y, ok := b.(LuaString)
		return ok && a.String() == y.String()
	case LUA_TNUMBER:
		_, ok := b.(LuaNumber)
		if !ok {
			return false
		}
		switch b.Type() {
		case LUA_TNUMBER:
			fa, _ := convertToFloat(a)
			fb, _ := convertToFloat(b)
			return fa == fb
		default:
			return false
		}
	case LUA_TTABLE:
		if b.Type() == LUA_TTABLE && a != b && ls != nil {
			if result, ok := callMetamethod(ls, a, b, "__eq"); ok {
				return convertToBoolean(result)
			}
		}
		return a == b
	default:
		return a == b
	}
}

func _lt(a, b LuaValue, ls *LuaState) bool {
	switch a.Type() {
	case LUA_TSTRING:
		if b.Type() == LUA_TSTRING {
			return a.String() < b.String()
		}
		return false
	case LUA_TNUMBER:
		if b.Type() == LUA_TNUMBER {
			x, _ := convertToFloat(a)
			y, _ := convertToFloat(b)
			return x < y
		}
		return false
	}

	if result, ok := callMetamethod(ls, a, b, "__lt"); ok {
		return convertToBoolean(result)
	}
	panic("comparison error!")
	return false
}

func _le(a, b LuaValue, ls *LuaState) bool {
	switch a.Type() {
	case LUA_TSTRING:
		if b.Type() == LUA_TSTRING {
			return a.String() <= b.String()
		}
		return false
	case LUA_TNUMBER:
		if b.Type() == LUA_TNUMBER {
			x, _ := convertToFloat(a)
			y, _ := convertToFloat(b)
			return x <= y
		}
		return false
	}
	if result, ok := callMetamethod(ls, a, b, "__le"); ok {
		return convertToBoolean(result)
	}
	if _lt(a, b, ls) == true {
		return true
	}
	if _eq(a, b, ls) == true {
		return true
	}
	panic("comparison error!")
	return false
}

// [-1, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_gettable
func luaGetTable(ls *LuaState, idx int) LuaValueType {
	t := ls.stack.get(idx)
	k := ls.stack.pop()
	return luaGetTable_(ls, t, k, false)
}

// push(t[k])
func luaGetTable_(ls *LuaState, t, k LuaValue, raw bool) LuaValueType {
	if tbl, ok := t.(*LuaTable); ok {
		v := tbl.Get(k)
		if raw || v != nil || !tbl.hasMetafield("__index") {
			ls.stack.push(v)
			return v.Type()
		}
	}

	if !raw {
		if mf := GetMetafield(ls, t, "__index"); mf != nil {
			switch x := mf.(type) {
			case *LuaTable:
				return luaGetTable_(ls, x, k, false)
			case *LuaClosure:
				ls.stack.push(mf)
				ls.stack.push(t)
				ls.stack.push(k)
				ls.Call(2, 1)
				v := ls.stack.get(-1)
				return v.Type()
			}
		}
	}
	panic("index error!")
	return LUA_TNIL
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_getfield
func luaGetField(ls *LuaState, idx int, k string) LuaValueType {
	t := ls.stack.get(idx)
	return luaGetTable_(ls, t, LuaString(k), false)
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_geti
func luaGetI(ls *LuaState, idx int, i int64) LuaValueType {
	t := ls.stack.get(idx)
	return luaGetTable_(ls, t, LuaNumber(i), false)
}

// [-1, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_rawget
func luaRawGet(ls *LuaState, idx int) LuaValueType {
	t := ls.stack.get(idx)
	k := ls.stack.pop()
	return luaGetTable_(ls, t, k, true)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_rawgeti
func luaRawGetI(ls *LuaState, idx int, i int64) LuaValueType {
	t := ls.stack.get(idx)
	return luaGetTable_(ls, t, LuaNumber(i), true)
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_getglobal
func luaGetGlobal(ls *LuaState, name string) LuaValueType {
	t := ls.registry.Get(LUA_RIDX_GLOBALS)
	return luaGetTable_(ls, t, LuaString(name), false)
}

// [-0, +(0|1), –]
// http://www.lua.org/manual/5.3/manual.html#lua_getmetatable
func luaGetMetatable(ls *LuaState, idx int) bool {
	val := ls.stack.get(idx)

	if mt := GetMetatable(ls, val); mt != nil {
		ls.stack.push(mt)
		return true
	} else {
		return false
	}
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_len
func luaLen(ls *LuaState, idx int) {
	val := ls.stack.get(idx)
	if val.Type() == LUA_TSTRING {
		ls.stack.push(LuaNumber(val.Len()))
	} else if result, ok := callMetamethod(ls, val, val, "__len"); ok {
		ls.stack.push(result)
	} else if val.Type() == LUA_TTABLE {
		ls.stack.push(LuaNumber(val.Len()))
	} else {
		panic("length error!")
	}
}

// [-n, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_concat
func luaConcat(ls *LuaState, n int) {
	if n == 0 {
		ls.stack.push(LuaString(""))
	} else if n >= 2 {
		for i := 1; i < n; i++ {
			if luaIsString(ls, -1) && luaIsString(ls, -2) {
				s2 := luaToString(ls, -1)
				s1 := luaToString(ls, -2)
				ls.stack.pop()
				ls.stack.pop()
				ls.stack.push(LuaString(s1 + s2))
				continue
			}

			b := ls.stack.pop()
			a := ls.stack.pop()
			if result, ok := callMetamethod(ls, a, b, "__concat"); ok {
				ls.stack.push(result)
				continue
			}

			panic("concatenation error!")
		}
	}
	// n == 1, do nothing
}

// [-1, +(2|0), e]
// http://www.lua.org/manual/5.3/manual.html#lua_next
func luaNext(ls *LuaState, idx int) bool {
	val := ls.stack.get(idx)
	if t, ok := val.(*LuaTable); ok {
		key := ls.stack.pop()
		if nextKey, v := t.nextKey(key); nextKey != nil {
			ls.stack.push(nextKey)
			ls.stack.push(v)
			return true
		}
		return false
	}
	panic("table expected!")
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_stringtonumber
func luaStringToNumber(ls *LuaState, s string) bool {
	if n, ok := number.ParseInteger(s); ok {
		ls.Push(LuaNumber(n))
		return true
	}
	if n, ok := number.ParseFloat(s); ok {
		ls.Push(LuaNumber(n))
		return true
	}
	return false
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_pushfstring
func luaPushFString(ls *LuaState, fmtStr string, a ...interface{}) {
	str := fmt.Sprintf(fmtStr, a...)
	ls.stack.push(LuaString(str))
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushglobaltable
func luaPushGlobalTable(ls *LuaState) {
	global := ls.registry.Get(LUA_RIDX_GLOBALS)
	ls.stack.push(global)
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#lua_pushthread
func luaPushThread(ls *LuaState) bool {
	ls.stack.push(ls)
	return ls.isMainThread()
}

// [-2, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_settable
func luaSetTable(ls *LuaState, idx int) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	k := ls.stack.pop()
	luaSetTable_(ls, t, k, v, false)
}

// [-1, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_setfield
func luaSetField(ls *LuaState, idx int, k string) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	luaSetTable_(ls, t, LuaString(k), v, false)
}

// [-1, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_seti
func luaSetI(ls *LuaState, idx int, i int64) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	luaSetTable_(ls, t, LuaNumber(i), v, false)
}

// [-2, +0, m]
// http://www.lua.org/manual/5.3/manual.html#lua_rawset
func luaRawSet(ls *LuaState, idx int) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	k := ls.stack.pop()
	luaSetTable_(ls, t, k, v, true)
}

// [-1, +0, m]
// http://www.lua.org/manual/5.3/manual.html#lua_rawseti
func luaRawSetI(ls *LuaState, idx int, i int64) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	luaSetTable_(ls, t, LuaNumber(i), v, true)
}

// [-1, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_setglobal
func luaSetGlobal(ls *LuaState, name string) {
	t := ls.registry.Get(LUA_RIDX_GLOBALS)
	v := ls.stack.pop()
	luaSetTable_(ls, t, LuaString(name), v, false)
}

// [-1, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_setmetatable
func luaSetMetatable(ls *LuaState, idx int) {
	val := ls.stack.get(idx)
	mtVal := ls.stack.pop()

	if mtVal == nil {
		SetMetatable(ls, val, nil)
	} else if mt, ok := mtVal.(*LuaTable); ok {
		SetMetatable(ls, val, mt)
	} else {
		panic("set metatable error")
	}
}

// t[k]=v
func luaSetTable_(ls *LuaState, t, k, v LuaValue, raw bool) {
	if tb, ok := t.(*LuaTable); ok {
		if raw || tb.Get(k) != LuaNil || !tb.hasMetafield("__newindex") {
			tb.Set(k, v)
			return
		}
	}

	if !raw {
		if mf := GetMetafield(ls, t, "__newindex"); mf != nil {
			switch x := mf.(type) {
			case *LuaTable:
				luaSetTable_(ls, x, k, v, false)
				return
			case *LuaClosure:
				ls.stack.push(mf)
				ls.stack.push(t)
				ls.stack.push(k)
				ls.stack.push(v)
				ls.Call(3, 0)
				return
			}
		}
	}
	// todo
	panic("index error!")
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkstack
func luaCheckStack2(ls *LuaState, sz int, msg string) {
	if !luaCheckStack(ls, sz) {
		if msg != "" {
			ls.Error2("stack overflow (%s)", msg)
		} else {
			ls.Error2("stack overflow")
		}
	}
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checktype
func luaCheckType(ls *LuaState, arg int, t LuaValueType) {
	if luaType(ls, arg) != t {
		ls.tagError(arg, t)
	}
}

func luaCheckTypes(ls *LuaState, n int, typs ...LuaValueType) {
	vt := luaType(ls, n)
	for _, typ := range typs {
		if vt == typ {
			return
		}
	}
	buf := []string{}
	for _, typ := range typs {
		buf = append(buf, typ.String())
	}
	ls.ArgError(n, strings.Join(buf, " or ")+" expected, got "+ vt.String())
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_optinteger
func luaOptInteger(ls *LuaState, arg int, def int64) int64 {
	if luaIsNoneOrNil(ls, arg) {
		return def
	}
	return ls.CheckInteger(arg)
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_optnumber
func luaOptNumber(ls *LuaState, arg int, def float64) float64 {
	if luaIsNoneOrNil(ls, arg) {
		return def
	}
	return ls.CheckNumber(arg)
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_optstring
func luaOptString(ls *LuaState, arg int, def string) string {
	if luaIsNoneOrNil(ls, arg) {
		return def
	}
	return ls.CheckString(arg)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#luaL_typename
func luaTypeName2(ls *LuaState, idx int) string {
	return luaType(ls, idx).String()
}

// [-0, +0, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_len
func luaLen2(ls *LuaState, idx int) int64 {
	luaLen(ls, idx)
	i, isNum := luaToIntegerX(ls, -1)
	if !isNum {
		ls.Error2("object length is not an integer")
	}
	luaPop(ls, 1)
	return i
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_tolstring
func luaToString2(ls *LuaState, idx int) string {
	if luaCallMeta(ls, idx, "__tostring") { /* metafield? */
		if !luaIsString(ls, -1) {
			ls.Error2("'__tostring' must return a string")
		}
	} else {
		switch luaType(ls, idx) {
		case LUA_TNUMBER:
			if luaIsInteger(ls, idx) {
				ls.Push(LuaString(fmt.Sprintf("%d", luaToInteger(ls, idx)))) // todo
			} else {
				ls.Push(LuaString(fmt.Sprintf("%g", luaToNumber(ls, idx)))) // todo
			}
		case LUA_TSTRING:
			luaPushValue(ls, idx)
		case LUA_TBOOLEAN:
			if luaToBoolean(ls, idx) {
				ls.Push(LuaString("true"))
			} else {
				ls.Push(LuaString("false"))
			}
		case LUA_TNIL:
			ls.Push(LuaString("nil"))
		default:
			tt := luaGetMetafield(ls, idx, "__name") /* try name */
			var kind string
			if tt == LUA_TSTRING {
				kind = ls.CheckString(-1)
			} else {
				kind = luaTypeName2(ls, idx)
			}

			ls.Push(LuaString(fmt.Sprintf("%s: %p", kind, luaToPointer(ls, idx))))
			if tt != LUA_TNIL {
				luaRemove(ls, -2) /* remove '__name' */
			}
		}
	}
	return ls.CheckString(-1)
}

// [-(2|1), +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_arith
func luaArith(ls *LuaState, op ArithOp) {
	var a, b LuaValue // operands
	b = ls.stack.pop()
	if op != LUA_OPUNM && op != LUA_OPBNOT {
		a = ls.stack.pop()
	} else {
		a = b
	}

	operator := operators[op]
	if result := _arith(a, b, operator); result != nil {
		ls.stack.push(result)
		return
	}

	mm := operator.metamethod
	if result, ok := callMetamethod(ls, a, b, mm); ok {
		ls.stack.push(result)
		return
	}
	panic("arithmetic error")
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_getsubtable
func luaGetSubTable(ls *LuaState, idx int, fname string) bool {
	if luaGetField(ls, idx, fname) == LUA_TTABLE {
		return true /* table already there */
	}
	luaPop(ls, 1) /* remove previous result */
	idx = ls.stack.absIndex(idx)
	ls.NewTable()
	luaPushValue(ls, -1)        /* copy to be left at top */
	luaSetField(ls, idx, fname) /* assign new table to field */
	return false                /* false, because did not find table there */
}

// [-0, +(0|1), m]
// http://www.lua.org/manual/5.3/manual.html#luaL_getmetafield
func luaGetMetafield(ls *LuaState, obj int, event string) LuaValueType {
	if !luaGetMetatable(ls, obj) { /* no metatable? */
		return LUA_TNIL
	}

	ls.Push(LuaString(event))
	tt := luaRawGet(ls, -2)
	if tt == LUA_TNIL { /* is metafield nil? */
		luaPop(ls, 2) /* remove metatable and metafield */
	} else {
		luaRemove(ls, -2) /* remove only metatable */
	}
	return tt /* return metafield type */
}

// [-0, +(0|1), e]
// http://www.lua.org/manual/5.3/manual.html#luaL_callmeta
func luaCallMeta(ls *LuaState, obj int, event string) bool {
	obj = luaAbsIndex(ls, obj)
	if luaGetMetafield(ls, obj, event) == LUA_TNIL { /* no metafield? */
		return false
	}

	luaPushValue(ls, obj)
	ls.Call(1, 1)
	return true
}
