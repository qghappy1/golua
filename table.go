package golua

import (
	"fmt"
	"golua/number"
	"math"
)

func newLuaTable(nArr, nRec int) *LuaTable {
	t := &LuaTable{}
	if nArr > 0 {
		t.arr = make([]LuaValue, 0, nArr)
	}
	if nRec > 0 {
		t.map_ = make(map[LuaValue]LuaValue, nRec)
	}
	return t
}

func (tb *LuaTable) hasMetafield(fieldName string) bool {
	return tb.metatable != nil && tb.metatable.Get(LuaString(fieldName)) != LuaNil
}

func (tb *LuaTable) Get(key LuaValue) LuaValue {
	if key == nil || key == LuaNil {
		return LuaNil
	}
	if key.Type() == LUA_TNUMBER {
		if idx, ok := floatToInteger(key.(LuaNumber)); ok {
			if idx >= 1 && idx <= int64(len(tb.arr)) {
				return tb.arr[idx-1]
			}
		}
	}
	if v, ok := tb.map_[key]; ok {
		return v
	}
	return LuaNil
}

func (tb *LuaTable) Set(key, val LuaValue) {
	if key == nil || key == LuaNil {
		return
	}
	if val == nil {
		val = LuaNil
	}
	if key.Type() == LUA_TNUMBER {
		f, _ := key.(LuaNumber)
		if math.IsNaN(float64(f)) {
			return
		}
		tb.changed = true
		if idx, ok := number.FloatToInteger(float64(f)); ok && idx > 0 {
			arrLen := int64(len(tb.arr))
			if idx <= arrLen {
				tb.arr[idx-1] = val
				if idx == arrLen && val == LuaNil {
					tb.shrinkArray()
				}
				return
			}
			if idx == arrLen+1 {
				delete(tb.map_, key)
				if val != LuaNil {
					tb.arr = append(tb.arr, val)
					tb.expandArray()
				}
				return
			}
		}
		if val != LuaNil {
			if tb.map_ == nil {
				tb.map_ = make(map[LuaValue]LuaValue, 8)
			}
			tb.map_[key] = val
		} else {
			delete(tb.map_, key)
		}
		return
	}
	tb.changed = true
	if val != LuaNil {
		if tb.map_ == nil {
			tb.map_ = make(map[LuaValue]LuaValue, 8)
		}
		tb.map_[key] = val
	} else {
		delete(tb.map_, key)
	}
}

func (tb *LuaTable) MaxN() int {
	if len(tb.arr) == 0 {
		return 0
	}
	for i := len(tb.arr) - 1; i >= 0; i-- {
		if tb.arr[i] != LuaNil {
			return i + 1
		}
	}
	return 0
}

func (tb *LuaTable) ForEach(cb func(LuaValue, LuaValue)) {
	for i, v := range tb.arr {
		if v != LuaNil {
			cb(LuaNumber(i+1), v)
		}
	}
	for key, val := range tb.map_ {
		cb(key, val)
	}
}

func (tb *LuaTable) shrinkArray() {
	for i := len(tb.arr) - 1; i >= 0; i-- {
		if tb.arr[i] == nil {
			tb.arr = tb.arr[0:i]
		}
	}
}

func (tb *LuaTable) expandArray() {
	for idx := int64(len(tb.arr)) + 1; true; idx++ {
		key := LuaNumber(idx)
		if val, found := tb.map_[key]; found {
			delete(tb.map_, key)
			tb.arr = append(tb.arr, val)
		} else {
			break
		}
	}
}

func (tb *LuaTable) nextKey(key LuaValue) (LuaValue, LuaValue) {
	if tb.keys == nil || (key == LuaNil && tb.changed) {
		tb.initKeys()
		tb.changed = false
	}
	if key == LuaNil && len(tb.arr) > 0 {
		key = LuaNumber(0)
	}
	if key.Type() == LUA_TNUMBER {
		if idx, ok := floatToInteger(key.(LuaNumber)); ok && idx > 0 {
			if idx < int64(len(tb.arr)) {
				val := tb.arr[idx]
				return LuaNumber(idx + 1), val
			}
			if idx == int64(len(tb.arr)) {
				key = LuaNil
			}
		}
	}
	if key == LuaNil {
		ok := true
		if key, ok = tb.keys[key]; !ok {
			return LuaNil, LuaNil
		}
	}
	if val, ok := tb.map_[key]; ok {
		if nextKey, ok := tb.keys[key]; ok {
			return nextKey, val
		} else {
			return LuaNil, val
		}
	}
	return LuaNil, LuaNil
}

func (tb *LuaTable) initKeys() {
	tb.keys = make(map[LuaValue]LuaValue)
	var key LuaValue = LuaNil
	for k, v := range tb.map_ {
		if v != LuaNil {
			tb.keys[key] = k
			key = k
		}
	}
}

/* metatable */
func GetMetatable(ls *LuaState, val LuaValue) *LuaTable {
	if t, ok := val.(*LuaTable); ok {
		return t.metatable
	}
	if u, ok := val.(*LuaUserData); ok {
		return u.Metatable
	}
	if val == nil {
		val = LuaNil
	}
	key := LuaString(fmt.Sprintf("_MT%d", val.Type()))
	v := ls.registry.Get(key)
	if mt, ok := v.(*LuaTable); ok {
		return mt
	}
	return nil
}

func SetMetatable(ls *LuaState, val LuaValue, mt *LuaTable) {
	if t, ok := val.(*LuaTable); ok {
		t.metatable = mt
		return
	}
	if u, ok := val.(*LuaUserData); ok {
		u.Metatable = mt
		return
	}
	if val == nil {
		val = LuaNil
	}
	key := LuaString(fmt.Sprintf("_MT%d", val.Type()))
	ls.registry.Set(key, mt)
}

func GetMetafield(ls *LuaState, val LuaValue, fieldName string) LuaValue {
	if mt := GetMetatable(ls, val); mt != nil {
		return mt.Get(LuaString(fieldName))
	}
	return LuaNil
}

func callMetamethod(ls *LuaState, a, b LuaValue, mmName string) (LuaValue, bool) {
	var mm LuaValue
	if mm = GetMetafield(ls, a, mmName); mm == LuaNil {
		if mm = GetMetafield(ls, b, mmName); mm == LuaNil {
			return LuaNil, false
		}
	}

	ls.stack.check(4)
	ls.stack.push(mm)
	ls.stack.push(a)
	ls.stack.push(b)
	ls.Call(2, 1)
	return ls.stack.pop(), true
}

// return t[key]
func GetValueField(ls *LuaState, t, key LuaValue) LuaValue {
	for i := 0; i < 100; i++ {
		tb, istable := t.(*LuaTable)
		if istable {
			if ret := tb.Get(key); ret != LuaNil {
				return ret
			}
		}
		metaindex := GetMetafield(ls, t, "__index")
		switch metaindex.Type() {
		case LUA_TNIL:
			return LuaNil
		case LUA_TCLOSURE:
			ls.stack.push(metaindex)
			ls.stack.push(t)
			ls.stack.push(key)
			ls.Call(2, 1)
			return ls.stack.pop()
		default:
			t = metaindex
		}
	}
	return LuaNil
}
