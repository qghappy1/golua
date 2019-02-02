package golua

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

// [-1, +0, e]
// http://www.lua.org/manual/5.3/manual.html#lua_setglobal
func luaSetGlobal(ls *LuaState, name LuaString) {
	t := ls.registry.Get(LUA_RIDX_GLOBALS)
	v := ls.stack.pop()
	luaSetTable_(ls, t, name, v, false)
}
