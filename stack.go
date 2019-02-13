package golua

type luaStack struct {
	/* virtual stack */
	slots []LuaValue
	top   int
	/* call info */
	state   *LuaState
	closure *LuaClosure
	varargs []LuaValue
	openuvs map[int]*upvalue
	pc      int
	/* linked list */
	prev *luaStack
}

func newLuaStack(size int, state *LuaState) *luaStack {
	return &luaStack{
		slots: make([]LuaValue, size),
		top:   0,
		state: state,
	}
}

func (self *luaStack) check(n int) {
	free := len(self.slots) - self.top
	for i := free; i < n; i++ {
		self.slots = append(self.slots, LuaNil)
	}
}

func (self *luaStack) push(val LuaValue) {
	if self.top == len(self.slots) {
		panic("stack overflow!")
	}
	if val == nil {
		val = LuaNil
	}
	self.slots[self.top] = val
	self.top++
}

func (self *luaStack) pop() LuaValue {
	if self.top < 1 {
		panic("stack underflow")
	}
	self.top--
	val := self.slots[self.top]
	self.slots[self.top] = LuaNil
	return val
}

func (self *luaStack) pushN(vals []LuaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}

	for i := 0; i < n; i++ {
		if i < nVals && vals[i] != nil {
			self.push(vals[i])
		} else {
			self.push(LuaNil)
		}
	}
}

func (self *luaStack) popN(n int) []LuaValue {
	vals := make([]LuaValue, n)
	for i := n - 1; i >= 0; i-- {
		vals[i] = self.pop()
	}
	return vals
}

func (self *luaStack) absIndex(idx int) int {
	if idx >= 0 || idx <= LUA_REGISTRYINDEX {
		return idx
	}
	return idx + self.top + 1
}

func (self *luaStack) isValid(idx int) bool {
	if idx < LUA_REGISTRYINDEX { /* upvalues */
		uvIdx := LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		return c != nil && uvIdx < len(c.upvals)
	}
	if idx == LUA_REGISTRYINDEX {
		return true
	}
	absIdx := self.absIndex(idx)
	return absIdx > 0 && absIdx <= self.top
}

func (self *luaStack) get(idx int) LuaValue {
	if idx < LUA_REGISTRYINDEX { /* upvalues */
		uvIdx := LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		if c == nil || uvIdx >= len(c.upvals) {
			return LuaNil
		}
		return *(c.upvals[uvIdx].val)
	}

	if idx == LUA_REGISTRYINDEX {
		return self.state.registry
	}

	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		return self.slots[absIdx-1]
	}
	return LuaNil
}

func (self *luaStack) set(idx int, val LuaValue) {
	if val == nil {
		val = LuaNil
	}
	if idx < LUA_REGISTRYINDEX { /* upvalues */
		uvIdx := LUA_REGISTRYINDEX - idx - 1
		c := self.closure
		if c != nil && uvIdx < len(c.upvals) {
			*(c.upvals[uvIdx].val) = val
		}
		return
	}

	if idx == LUA_REGISTRYINDEX {
		self.state.registry = val.(*LuaTable)
		return
	}

	absIdx := self.absIndex(idx)
	if absIdx > 0 && absIdx <= self.top {
		self.slots[absIdx-1] = val
		return
	}
	panic("invalid index!idx:%v val:%v")
}

func (self *luaStack) reverse(from, to int) {
	slots := self.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}
