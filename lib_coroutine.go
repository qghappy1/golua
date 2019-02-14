package golua

/* thread status */
const (
	LUA_OK = iota
	LUA_YIELD
	LUA_ERRRUN
	LUA_ERRSYNTAX
	LUA_ERRMEM
	LUA_ERRGCMM
	LUA_ERRERR
	LUA_ERRFILE
)

var coFuncs = map[string]GoFunction{
	"create":      coCreate,
	"resume":      coResume,
	"yield":       coYield,
	"status":      coStatus,
	"isyieldable": coYieldable,
	"running":     coRunning,
	"wrap":        coWrap,
}

func OpenCoroutineLib(ls *LuaState) int {
	ls.NewLib(coFuncs)
	return 1
}

// [-?, +?, –]
// http://www.lua.org/manual/5.3/manual.html#lua_resume
func luaResume(lsTo *LuaState, lsFrom *LuaState, nArgs int) int {
	if lsFrom.coChan == nil {
		lsFrom.coChan = make(chan int)
	}

	if lsTo.coChan == nil {
		// start coroutine
		lsTo.coChan = make(chan int)
		lsTo.coCaller = lsFrom
		go func() {
			if lsTo.PCall(nArgs, -1, 0) != nil {
				lsTo.coStatus = LUA_OK
			}else {
				lsTo.coStatus = LUA_ERRERR
			}

			lsFrom.coChan <- 1
		}()
	} else {
		// resume coroutine
		if lsTo.coStatus != LUA_YIELD { // todo
			lsTo.stack.push(LuaString("cannot resume non-suspended coroutine"))
			return LUA_ERRRUN
		}
		lsTo.coStatus = LUA_OK
		lsTo.coChan <- 1
	}

	<-lsFrom.coChan // wait coroutine to finish or yield
	return lsTo.coStatus
}

// [-?, +?, e]
// http://www.lua.org/manual/5.3/manual.html#lua_yield
func luaYield(ls *LuaState, nResults int) int {
	if ls.coCaller == nil { // todo
		panic("attempt to yield from outside a coroutine")
	}
	ls.coStatus = LUA_YIELD
	ls.coCaller.coChan <- 1
	<-ls.coChan
	return luaGetTop(ls)
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isyieldable
func luaIsYieldable(ls *LuaState) bool {
	if ls.isMainThread() {
		return false
	}
	return ls.coStatus != LUA_YIELD // todo
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_status
// lua-5.3.4/src/lapi.c#lua_status()
func luaStatus(ls *LuaState) int {
	return ls.coStatus
}

// coroutine.create (f)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.create
// lua-5.3.4/src/lcorolib.c#luaB_cocreate()
func coCreate(ls *LuaState) int {
	luaCheckType(ls, 1, LUA_TCLOSURE)
	ls2 := luaNewThread(ls)
	luaPushValue(ls, 1)  /* move function to top */
	luaXMove(ls, ls2, 1) /* move function from ls to ls2 */
	return 1
}

// coroutine.resume (co [, val1, ···])
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.resume
// lua-5.3.4/src/lcorolib.c#luaB_coresume()
func coResume(ls *LuaState) int {
	co := luaToThread(ls, 1)
	ls.ArgCheck(co != nil, 1, "thread expected")

	if r := _auxResume(ls, co, luaGetTop(ls)-1); r < 0 {
		ls.Push(LuaFalse)
		luaInsert(ls, -2)
		return 2 /* return false + error message */
	} else {
		ls.Push(LuaTrue)
		luaInsert(ls, -(r + 1))
		return r + 1 /* return true + 'resume' returns */
	}
}

func _auxResume(ls, co *LuaState, narg int) int {
	if !luaCheckStack(ls, narg) {
		ls.Push(LuaString("too many arguments to resume"))
		return -1 /* error flag */
	}
	if luaStatus(co) == LUA_OK && luaGetTop(co) == 0 {
		ls.Push(LuaString("cannot resume dead coroutine"))
		return -1 /* error flag */
	}
	luaXMove(ls, co, narg)
	status := luaResume(co, ls, narg)
	if status == LUA_OK || status == LUA_YIELD {
		nres := luaGetTop(co)
		if !luaCheckStack(ls, nres + 1) {
			luaPop(co, nres) /* remove results anyway */
			ls.Push(LuaString("too many results to resume"))
			return -1 /* error flag */
		}
		luaXMove(co, ls, nres) /* move yielded values */
		return nres
	} else {
		luaXMove(co, ls, 1) /* move error message */
		return -1       /* error flag */
	}
}

// coroutine.yield (···)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.yield
// lua-5.3.4/src/lcorolib.c#luaB_yield()
func coYield(ls *LuaState) int {
	return luaYield(ls, luaGetTop(ls))
}

// coroutine.status (co)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.status
// lua-5.3.4/src/lcorolib.c#luaB_costatus()
func coStatus(ls *LuaState) int {
	co := luaToThread(ls, 1)
	ls.ArgCheck(co != nil, 1, "thread expected")
	if ls == co {
		ls.Push(LuaString("running"))
	} else {
		switch luaStatus(co) {
		case LUA_YIELD:
			ls.Push(LuaString("suspended"))
		case LUA_OK:
			if co.stack.prev!=nil { /* does it have frames? */
				ls.Push(LuaString("normal")) /* it is running */
			} else if luaGetTop(co) == 0 {
				ls.Push(LuaString("dead"))
			} else {
				ls.Push(LuaString("suspended"))
			}
		default: /* some error occurred */
			ls.Push(LuaString("dead"))
		}
	}

	return 1
}

// coroutine.isyieldable ()
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.isyieldable
func coYieldable(ls *LuaState) int {
	ls.Push(LuaBool(luaIsYieldable(ls)))
	return 1
}

// coroutine.running ()
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.running
func coRunning(ls *LuaState) int {
	isMain := luaPushThread(ls)
	ls.Push(LuaBool(isMain))
	return 2
}

// coroutine.wrap (f)
// http://www.lua.org/manual/5.3/manual.html#pdf-coroutine.wrap
func coWrap(ls *LuaState) int {
	panic("todo: coWrap!")
}
