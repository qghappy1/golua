package golua

// [-0, +0, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_openlibs
func (ls *LuaState) OpenLibs() {
	// libs := map[string]GoFunction{
	// 	"_G":        OpenBaseLib,
	// 	"math":      OpenMathLib,
	// 	"table":     OpenTableLib,
	// 	"string":    OpenStringLib,
	// 	"utf8":      OpenUTF8Lib,
	// 	"os":        OpenOSLib,
	// 	"package":   OpenPackageLib,
	// 	"coroutine": OpenCoroutineLib,
	// }

	// for name, fun := range libs {
	// 	ls.RequireF(name, fun, true)
	// 	luaPop(ls, 1)
	// }
}

type FuncReg map[string]GoFunction

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_newlib
func (ls *LuaState) NewLib(l FuncReg) {
	ls.NewLibTable(l)
	ls.SetFuncs(l, 0)
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_newlibtable
func (ls *LuaState) NewLibTable(l FuncReg) {
	ls.createTable(0, len(l))
}

// [-nup, +0, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_setfuncs
func (ls *LuaState) SetFuncs(l FuncReg, nup int) {
	luaCheckStack2(ls, nup, "too many upvalues")
	for name, fun := range l { /* fill the table with given functions */
		for i := 0; i < nup; i++ { /* copy upvalues to the top */
			luaPushValue(ls, -nup)
		}
		// r[-(nup+2)][name]=fun
		ls.PushGoClosure(fun, nup) /* Closure with those upvalues */
		luaSetField(ls, -(nup + 2), name)
	}
	luaPop(ls, nup) /* remove upvalues */
}
