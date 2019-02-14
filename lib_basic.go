package golua

import (
	"fmt"
	"strconv"
	"strings"
)

var baseFuncs = map[string]GoFunction{
	"print":        basePrint,
	"assert":       baseAssert,
	"error":        baseError,
	"select":       baseSelect,
	"ipairs":       baseIPairs,
	"pairs":        basePairs,
	"next":         baseNext,
	"load":         baseLoad,
	"loadfile":     baseLoadFile,
	"dofile":       baseDoFile,
	"pcall":        basePCall,
	"xpcall":       baseXPCall,
	"getmetatable": baseGetMetatable,
	"setmetatable": baseSetMetatable,
	"rawequal":     baseRawEqual,
	"rawlen":       baseRawLen,
	"rawget":       baseRawGet,
	"rawset":       baseRawSet,
	"type":         baseType,
	"tostring":     baseToString,
	"tonumber":     baseToNumber,
	/* placeholders */
	"_G":       nil,
	"_VERSION": nil,
}

// lua-5.3.4/src/lbaselib.c#luaopen_base()
func OpenBaseLib(ls *LuaState) int {
	/* open lib into global table */
	luaPushGlobalTable(ls)
	ls.SetFuncs(baseFuncs, 0)
	/* set global _G */
	luaPushValue(ls, -1)
	luaSetField(ls, -2, "_G")
	/* set global _VERSION */
	ls.Push(LuaString("Lua 5.3")) // todo
	luaSetField(ls, -2, "_VERSION")
	return 1
}

// print (···)
// http://www.lua.org/manual/5.3/manual.html#pdf-print
// lua-5.3.4/src/lbaselib.c#luaB_print()
func basePrint(ls *LuaState) int {
	n := luaGetTop(ls) /* number of arguments */
	luaGetGlobal(ls, "tostring")
	for i := 1; i <= n; i++ {
		luaPushValue(ls, -1) /* function to be called */
		luaPushValue(ls, i)  /* value to print */
		ls.Call(1, 1)
		s, ok := luaToStringX(ls, -1) /* get result */
		if !ok {
			return ls.Error2("'tostring' must return a string to 'print'")
		}
		if i > 1 {
			fmt.Print("\t")
		}
		fmt.Print(s)
		luaPop(ls, 1) /* pop result */
	}
	fmt.Println()
	return 0
}

// assert (v [, message])
// http://www.lua.org/manual/5.3/manual.html#pdf-assert
// lua-5.3.4/src/lbaselib.c#luaB_assert()
func baseAssert(ls *LuaState) int {
	if luaToBoolean(ls, 1) { /* condition is true? */
		return luaGetTop(ls) /* return all arguments */
	} else { /* error */
		ls.CheckAny(1)                     /* there must be a condition */
		luaRemove(ls, 1)                       /* remove it */
		ls.Push(LuaString("assertion failed!")) /* default message */
		luaSetTop(ls, 1)                       /* leave only message (default if no other one) */
		return baseError(ls)               /* call 'error' */
	}
}

// error (message [, level])
// http://www.lua.org/manual/5.3/manual.html#pdf-error
// lua-5.3.4/src/lbaselib.c#luaB_error()
func baseError(ls *LuaState) int {
	level := int(luaOptInteger(ls, 2, 1))
	luaSetTop(ls, 1)
	if luaType(ls, 1) == LUA_TSTRING && level > 0 {
		// ls.Where(level) /* add extra information */
		// luaPushValue(ls, 1)
		// luaConcat(ls, 2)
	}
	return ls.Error()
}

// select (index, ···)
// http://www.lua.org/manual/5.3/manual.html#pdf-select
// lua-5.3.4/src/lbaselib.c#luaB_select()
func baseSelect(ls *LuaState) int {
	n := int64(luaGetTop(ls))
	if luaType(ls, 1) == LUA_TSTRING && ls.CheckString(1) == "#" {
		ls.Push(LuaNumber(n - 1))
		return 1
	} else {
		i := ls.CheckInteger(1)
		if i < 0 {
			i = n + i
		} else if i > n {
			i = n
		}
		ls.ArgCheck(1 <= i, 1, "index out of range")
		return int(n - i)
	}
}

// ipairs (t)
// http://www.lua.org/manual/5.3/manual.html#pdf-ipairs
// lua-5.3.4/src/lbaselib.c#luaB_ipairs()
func baseIPairs(ls *LuaState) int {
	ls.CheckAny(1)
	ls.PushGoFunction(iPairsAux) /* iteration function */
	luaPushValue(ls, 1)              /* state */
	ls.Push(LuaNumber(0))            /* initial value */
	return 3
}

func iPairsAux(ls *LuaState) int {
	i := ls.CheckInteger(2) + 1
	ls.Push(LuaNumber(i))
	if luaGetI(ls, 1, i) == LUA_TNIL {
		return 1
	} else {
		return 2
	}
}

// pairs (t)
// http://www.lua.org/manual/5.3/manual.html#pdf-pairs
// lua-5.3.4/src/lbaselib.c#luaB_pairs()
func basePairs(ls *LuaState) int {
	ls.CheckAny(1)
	if luaGetMetafield(ls, 1, "__pairs") == LUA_TNIL { /* no metamethod? */
		ls.PushGoFunction(baseNext) /* will return generator, */
		luaPushValue(ls, 1)             /* state, */
		ls.Push(LuaNil)
	} else {
		luaPushValue(ls, 1) /* argument 'self' to metamethod */
		ls.Call(1, 3)   /* get 3 values from metamethod */
	}
	return 3
}

// next (table [, index])
// http://www.lua.org/manual/5.3/manual.html#pdf-next
// lua-5.3.4/src/lbaselib.c#luaB_next()
func baseNext(ls *LuaState) int {
	luaCheckType(ls, 1, LUA_TTABLE)
	luaSetTop(ls, 2) /* create a 2nd argument if there isn't one */
	if luaNext(ls, 1) {
		return 2
	} else {
		ls.Push(LuaNil)
		return 1
	}
}

// load (chunk [, chunkname [, mode [, env]]])
// http://www.lua.org/manual/5.3/manual.html#pdf-load
// lua-5.3.4/src/lbaselib.c#luaB_load()
func baseLoad(ls *LuaState) int {
	var status int
	chunk, isStr := luaToStringX(ls, 1)
	//mode := luaOptString(ls, 3, "bt")
	env := 0 /* 'env' index or 0 if no 'env' */
	if !luaIsNone(ls, 4) {
		env = 4
	}
	if isStr { /* loading a string? */
		chunkname := luaOptString(ls, 2, chunk)
		status = ls.Load([]byte(chunk), chunkname)
	} else { /* loading from a reader function */
		panic("loading from a reader function") // todo
	}
	return loadAux(ls, status, env)
}

// lua-5.3.4/src/lbaselib.c#load_aux()
func loadAux(ls *LuaState, status, envIdx int) int {
	if status == LUA_OK {
		if envIdx != 0 { /* 'env' parameter? */
			panic("todo!")
		}
		return 1
	} else { /* error (message is on top of the stack) */
		ls.Push(LuaNil)
		luaInsert(ls, -2) /* put before error message */
		return 2      /* return nil plus error message */
	}
}

// loadfile ([filename [, mode [, env]]])
// http://www.lua.org/manual/5.3/manual.html#pdf-loadfile
// lua-5.3.4/src/lbaselib.c#luaB_loadfile()
func baseLoadFile(ls *LuaState) int {
	fname := luaOptString(ls, 1, "")
	//mode := luaOptString(ls, 1, "bt")
	env := 0 /* 'env' index or 0 if no 'env' */
	if !luaIsNone(ls, 3) {
		env = 3
	}
	status := ls.LoadFileX(fname)
	return loadAux(ls, status, env)
}

// dofile ([filename])
// http://www.lua.org/manual/5.3/manual.html#pdf-dofile
// lua-5.3.4/src/lbaselib.c#luaB_dofile()
func baseDoFile(ls *LuaState) int {
	fname := luaOptString(ls, 1, "bt")
	luaSetTop(ls, 1)
	if ls.LoadFile(fname) != LUA_OK {
		return ls.Error()
	}
	ls.Call(0, LUA_MULTRET)
	return luaGetTop(ls) - 1
}

// pcall (f [, arg1, ···])
// http://www.lua.org/manual/5.3/manual.html#pdf-pcall
func basePCall(ls *LuaState) int {
	nArgs := luaGetTop(ls) - 1
	err := ls.PCall(nArgs, -1, 0)
	ls.Push(LuaBool(err == nil))
	luaInsert(ls, 1)
	return luaGetTop(ls)
}

// xpcall (f, msgh [, arg1, ···])
// http://www.lua.org/manual/5.3/manual.html#pdf-xpcall
func baseXPCall(ls *LuaState) int {
	panic("todo!")
}

// getmetatable (object)
// http://www.lua.org/manual/5.3/manual.html#pdf-getmetatable
// lua-5.3.4/src/lbaselib.c#luaB_getmetatable()
func baseGetMetatable(ls *LuaState) int {
	ls.CheckAny(1)
	if !luaGetMetatable(ls, 1) {
		ls.Push(LuaNil)
		return 1 /* no metatable */
	}
	luaGetMetafield(ls, 1, "__metatable")
	return 1 /* returns either __metatable field (if present) or metatable */

}

// setmetatable (table, metatable)
// http://www.lua.org/manual/5.3/manual.html#pdf-setmetatable
// lua-5.3.4/src/lbaselib.c#luaB_setmetatable()
func baseSetMetatable(ls *LuaState) int {
	t := luaType(ls, 2)
	luaCheckType(ls, 1, LUA_TTABLE)
	ls.ArgCheck(t == LUA_TNIL || t == LUA_TTABLE, 2,
		"nil or table expected")
	if luaGetMetafield(ls, 1, "__metatable") != LUA_TNIL {
		return ls.Error2("cannot change a protected metatable")
	}
	luaSetTop(ls, 2)
	luaSetMetatable(ls, 1)
	return 1
}

// rawequal (v1, v2)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawequal
// lua-5.3.4/src/lbaselib.c#luaB_rawequal()
func baseRawEqual(ls *LuaState) int {
	ls.CheckAny(1)
	ls.CheckAny(2)
	ls.Push(LuaBool(luaRawEqual(ls, 1, 2)))
	return 1
}

// rawlen (v)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawlen
// lua-5.3.4/src/lbaselib.c#luaB_rawlen()
func baseRawLen(ls *LuaState) int {
	t := luaType(ls, 1)
	ls.ArgCheck(t == LUA_TTABLE || t == LUA_TSTRING, 1,
		"table or string expected")
	ls.Push(LuaNumber(luaRawLen(ls, 1)))
	return 1
}

// rawget (table, index)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawget
// lua-5.3.4/src/lbaselib.c#luaB_rawget()
func baseRawGet(ls *LuaState) int {
	luaCheckType(ls, 1, LUA_TTABLE)
	ls.CheckAny(2)
	luaSetTop(ls, 2)
	luaRawGet(ls, 1)
	return 1
}

// rawset (table, index, value)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawset
// lua-5.3.4/src/lbaselib.c#luaB_rawset()
func baseRawSet(ls *LuaState) int {
	luaCheckType(ls, 1, LUA_TTABLE)
	ls.CheckAny(2)
	ls.CheckAny(3)
	luaSetTop(ls, 3)
	luaRawSet(ls, 1)
	return 1
}

// type (v)
// http://www.lua.org/manual/5.3/manual.html#pdf-type
// lua-5.3.4/src/lbaselib.c#luaB_type()
func baseType(ls *LuaState) int {
	t := luaType(ls, 1)
	ls.ArgCheck(t != LUA_TNONE, 1, "value expected")
	ls.Push(LuaString(t.String()))
	return 1
}

// tostring (v)
// http://www.lua.org/manual/5.3/manual.html#pdf-tostring
// lua-5.3.4/src/lbaselib.c#luaB_tostring()
func baseToString(ls *LuaState) int {
	ls.CheckAny(1)
	luaToString2(ls, 1)
	return 1
}

// tonumber (e [, base])
// http://www.lua.org/manual/5.3/manual.html#pdf-tonumber
// lua-5.3.4/src/lbaselib.c#luaB_tonumber()
func baseToNumber(ls *LuaState) int {
	if luaIsNoneOrNil(ls, 2) { /* standard conversion? */
		ls.CheckAny(1)
		if luaType(ls, 1) == LUA_TNUMBER { /* already a number? */
			luaSetTop(ls, 1) /* yes; return it */
			return 1
		} else {
			if s, ok := luaToStringX(ls, 1); ok {
				if luaStringToNumber(ls, s) {
					return 1 /* successful conversion to number */
				} /* else not a number */
			}
		}
	} else {
		luaCheckType(ls, 1, LUA_TSTRING) /* no numbers as strings */
		s := strings.TrimSpace(luaToString(ls, 1))
		base := int(ls.CheckInteger(2))
		ls.ArgCheck(2 <= base && base <= 36, 2, "base out of range")
		if n, err := strconv.ParseInt(s, base, 64); err == nil {
			ls.Push(LuaNumber(n))
			return 1
		} /* else not a number */
	} /* else not a number */
	ls.Push(LuaNil) /* not a number */
	return 1
}
