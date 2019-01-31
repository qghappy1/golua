package golua

import (

)


type luaStack struct {
	/* virtual stack */
	slots []LuaValue
	top   int
	/* call info */
	state   *LuaState
	Closure *LuaClosure
	varargs []LuaValue
	openuvs map[int]*upvalue
	pc      int
	/* linked list */
	prev *luaStack
}