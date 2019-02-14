package golua

import (
	"sort"
	"strings"
)

const MAX_LEN = 1000000 // TODO

const (
	LUA_MAXINTEGER = 1<<63 - 1
	LUA_MININTEGER = -1 << 63
)

/*
** Operations that an object must define to mimic a table
** (some functions only need some of them)
 */
const (
	TAB_R  = 1               /* read */
	TAB_W  = 2               /* write */
	TAB_L  = 4               /* length */
	TAB_RW = (TAB_R | TAB_W) /* read/write */
)

var tabFuncs = map[string]GoFunction{
	"move":   tabMove,
	"insert": tabInsert,
	"remove": tabRemove,
	"sort":   tabSort,
	"concat": tabConcat,
	"pack":   tabPack,
	"unpack": tabUnpack,
}

func OpenTableLib(ls *LuaState) int {
	ls.NewLib(tabFuncs)
	return 1
}

// table.move (a1, f, e, t [,a2])
// http://www.lua.org/manual/5.3/manual.html#pdf-table.move
// lua-5.3.4/src/ltablib.c#tremove()
func tabMove(ls *LuaState) int {
	f := ls.CheckInteger(2)
	e := ls.CheckInteger(3)
	t := ls.CheckInteger(4)
	tt := 1 /* destination table */
	if !luaIsNoneOrNil(ls, 5) {
		tt = 5
	}
	_checkTab(ls, 1, TAB_R)
	_checkTab(ls, tt, TAB_W)
	if e >= f { /* otherwise, nothing to move */
		var n, i int64
		ls.ArgCheck(f > 0 || e < LUA_MAXINTEGER+f, 3,
			"too many elements to move")
		n = e - f + 1 /* number of elements to move */
		ls.ArgCheck(t <= LUA_MAXINTEGER-n+1, 4,
			"destination wrap around")
		if t > e || t <= f || (tt != 1 && !luaCompare(ls, 1, tt, LUA_OPEQ)) {
			for i = 0; i < n; i++ {
				luaGetI(ls, 1, f+i)
				luaSetI(ls, tt, t+i)
			}
		} else {
			for i = n - 1; i >= 0; i-- {
				luaGetI(ls, 1, f+i)
				luaSetI(ls, tt, t+i)
			}
		}
	}
	luaPushValue(ls, tt) /* return destination table */
	return 1
}

// table.insert (list, [pos,] value)
// http://www.lua.org/manual/5.3/manual.html#pdf-table.insert
// lua-5.3.4/src/ltablib.c#tinsert()
func tabInsert(ls *LuaState) int {
	e := _auxGetN(ls, 1, TAB_RW) + 1 /* first empty element */
	var pos int64                    /* where to insert new element */
	switch luaGetTop(ls) {
	case 2: /* called with only 2 arguments */
		pos = e /* insert new element at the end */
	case 3:
		pos = ls.CheckInteger(2) /* 2nd argument is the position */
		ls.ArgCheck(1 <= pos && pos <= e, 2, "position out of bounds")
		for i := e; i > pos; i-- { /* move up elements */
			luaGetI(ls, 1, i-1)
			luaSetI(ls, 1, i) /* t[i] = t[i - 1] */
		}
	default:
		return ls.Error2("wrong number of arguments to 'insert'")
	}
	luaSetI(ls, 1, pos) /* t[pos] = v */
	return 0
}

// table.remove (list [, pos])
// http://www.lua.org/manual/5.3/manual.html#pdf-table.remove
// lua-5.3.4/src/ltablib.c#tremove()
func tabRemove(ls *LuaState) int {
	size := _auxGetN(ls, 1, TAB_RW)
	pos := luaOptInteger(ls, 2, size)
	if pos != size { /* validate 'pos' if given */
		ls.ArgCheck(1 <= pos && pos <= size+1, 1, "position out of bounds")
	}
	luaGetI(ls, 1, pos) /* result = t[pos] */
	for ; pos < size; pos++ {
		luaGetI(ls, 1, pos+1)
		luaSetI(ls, 1, pos) /* t[pos] = t[pos + 1] */
	}
	ls.Push(LuaNil)
	luaSetI(ls, 1, pos) /* t[pos] = nil */
	return 1
}

// table.concat (list [, sep [, i [, j]]])
// http://www.lua.org/manual/5.3/manual.html#pdf-table.concat
// lua-5.3.4/src/ltablib.c#tconcat()
func tabConcat(ls *LuaState) int {
	tabLen := _auxGetN(ls, 1, TAB_R)
	sep := luaOptString(ls, 2, "")
	i := luaOptInteger(ls, 3, 1)
	j := luaOptInteger(ls, 4, tabLen)

	if i > j {
		ls.Push(LuaString(""))
		return 1
	}

	buf := make([]string, j-i+1)
	for k := i; k > 0 && k <= j; k++ {
		luaGetI(ls, 1, k)
		if !luaIsString(ls, -1) {
			ls.Error2("invalid value (%s) at index %d in table for 'concat'",
				luaTypeName2(ls, -1), i)
		}
		buf[k-i] = luaToString(ls, -1)
		luaPop(ls, 1)
	}
	ls.Push(LuaString(strings.Join(buf, sep)))

	return 1
}

func _auxGetN(ls *LuaState, n, w int) int64 {
	_checkTab(ls, n, w|TAB_L)
	return luaLen2(ls, n)
}

/*
** Check that 'arg' either is a table or can behave like one (that is,
** has a metatable with the required metamethods)
 */
func _checkTab(ls *LuaState, arg, what int) {
	if luaType(ls, arg) != LUA_TTABLE { /* is it not a table? */
		n := 1                     /* number of elements to pop */
		if luaGetMetatable(ls, arg) && /* must have metatable */
			(what&TAB_R != 0 || _checkField(ls, "__index", &n)) &&
			(what&TAB_W != 0 || _checkField(ls, "__newindex", &n)) &&
			(what&TAB_L != 0 || _checkField(ls, "__len", &n)) {
			luaPop(ls, n) /* pop metatable and tested metamethods */
		} else {
			luaCheckType(ls, arg, LUA_TTABLE) /* force an error */
		}
	}
}

func _checkField(ls *LuaState, key string, n *int) bool {
	ls.Push(LuaString(key))
	*n++
	return luaRawGet(ls, -*n) != LUA_TNIL
}

/* Pack/unpack */

// table.pack (···)
// http://www.lua.org/manual/5.3/manual.html#pdf-table.pack
// lua-5.3.4/src/ltablib.c#pack()
func tabPack(ls *LuaState) int {
	n := int64(luaGetTop(ls))   /* number of elements to pack */
	ls.createTable(int(n), 1) /* create result table */
	luaInsert(ls, 1)              /* put it at index 1 */
	for i := n; i >= 1; i-- { /* assign elements */
		luaSetI(ls, 1, i)
	}
	ls.Push(LuaNumber(n))
	luaSetField(ls, 1, "n") /* t.n = number of elements */
	return 1            /* return table */
}

// table.unpack (list [, i [, j]])
// http://www.lua.org/manual/5.3/manual.html#pdf-table.unpack
// lua-5.3.4/src/ltablib.c#unpack()
func tabUnpack(ls *LuaState) int {
	i := luaOptInteger(ls, 2, 1)
	e := luaOptInteger(ls, 3, luaLen2(ls, 1))
	if i > e { /* empty range */
		return 0
	}

	n := int(e - i + 1)
	if n <= 0 || n >= MAX_LEN || !luaCheckStack(ls, n) {
		return ls.Error2("too many results to unpack")
	}

	for ; i < e; i++ { /* push arg[i..e - 1] (to avoid overflows) */
		luaGetI(ls, 1, i)
	}
	luaGetI(ls, 1, e) /* push last element */
	return n
}

/* sort */

// table.sort (list [, comp])
// http://www.lua.org/manual/5.3/manual.html#pdf-table.sort
func tabSort(ls *LuaState) int {
	w := wrapper{ls}
	ls.ArgCheck(w.Len() < MAX_LEN, 1, "array too big")
	sort.Sort(w)
	return 0
}

type wrapper struct {
	ls *LuaState
}

func (self wrapper) Len() int {
	return int(luaLen2(self.ls, 1))
}

func (self wrapper) Less(i, j int) bool {
	ls := self.ls
	if luaIsFunction(ls, 2) { // cmp is given
		luaPushValue(ls, 2)
		luaGetI(ls, 1, int64(i+1))
		luaGetI(ls, 1, int64(j+1))
		ls.Call(2, 1)
		b := luaToBoolean(ls, -1)
		luaPop(ls, 1)
		return b
	} else { // cmp is missing
		luaGetI(ls, 1, int64(i+1))
		luaGetI(ls, 1, int64(j+1))
		b := luaCompare(ls, -2, -1, LUA_OPLT)
		luaPop(ls, 2)
		return b
	}
}

func (self wrapper) Swap(i, j int) {
	ls := self.ls
	luaGetI(ls, 1, int64(i+1))
	luaGetI(ls, 1, int64(j+1))
	luaSetI(ls, 1, int64(i+1))
	luaSetI(ls, 1, int64(j+1))
}
