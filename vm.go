package golua

import (
	"golua/number"
	"math"
	"reflect"
	"fmt"
)

/* arithmetic functions */
const (
	LUA_OPADD  = iota // +
	LUA_OPSUB         // -
	LUA_OPMUL         // *
	LUA_OPMOD         // %
	LUA_OPPOW         // ^
	LUA_OPDIV         // /
	LUA_OPIDIV        // //
	LUA_OPBAND        // &
	LUA_OPBOR         // |
	LUA_OPBXOR        // ~
	LUA_OPSHL         // <<
	LUA_OPSHR         // >>
	LUA_OPUNM         // -
	LUA_OPBNOT        // ~
)

/* comparison functions */
const (
	LUA_OPEQ = iota // ==
	LUA_OPLT        // <
	LUA_OPLE        // <=
)

var (
	_iadd  = func(a, b int64) int64 { return a + b }
	_fadd  = func(a, b float64) float64 { return a + b }
	_isub  = func(a, b int64) int64 { return a - b }
	_fsub  = func(a, b float64) float64 { return a - b }
	_imul  = func(a, b int64) int64 { return a * b }
	_fmul  = func(a, b float64) float64 { return a * b }
	_imod  = number.IMod
	_fmod  = number.FMod
	_pow   = math.Pow
	_div   = func(a, b float64) float64 { return a / b }
	_iidiv = number.IFloorDiv
	_fidiv = number.FFloorDiv
	_band  = func(a, b int64) int64 { return a & b }
	_bor   = func(a, b int64) int64 { return a | b }
	_bxor  = func(a, b int64) int64 { return a ^ b }
	_shl   = number.ShiftLeft
	_shr   = number.ShiftRight
	_iunm  = func(a, _ int64) int64 { return -a }
	_funm  = func(a, _ float64) float64 { return -a }
	_bnot  = func(a, _ int64) int64 { return ^a }
)

var operators = []operator{
	operator{"__add", _iadd, _fadd},
	operator{"__sub", _isub, _fsub},
	operator{"__mul", _imul, _fmul},
	operator{"__mod", _imod, _fmod},
	operator{"__pow", nil, _pow},
	operator{"__div", nil, _div},
	operator{"__idiv", _iidiv, _fidiv},
	operator{"__band", _band, nil},
	operator{"__bor", _bor, nil},
	operator{"__bxor", _bxor, nil},
	operator{"__shl", _shl, nil},
	operator{"__shr", _shr, nil},
	operator{"__unm", _iunm, _funm},
	operator{"__bnot", _bnot, nil},
}

type ArithOp = int
type CompareOp = int

type operator struct {
	metamethod  string
	integerFunc func(int64, int64) int64
	floatFunc   func(float64, float64) float64
}

func _arith(a, b LuaValue, op operator) LuaValue {
	if op.floatFunc == nil { // bitwise
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return LuaNumber(op.integerFunc(x, y))
			}
		}
	} else { // arith
		if op.integerFunc != nil { // add,sub,mul,mod,idiv,unm
			if x, ok := convertToInteger(a); ok {
				if y, ok := convertToInteger(b); ok {
					return LuaNumber(op.integerFunc(x, y))
				}
			}
		}
		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return LuaNumber(op.floatFunc(x, y))
			}
		}
	}
	return LuaNil
}

func luaUpvalueIndex(i int) int {
	return LUA_REGISTRYINDEX - i
}

func (ls *LuaState) pc() int {
	return ls.stack.pc
}

func (ls *LuaState) addPC(n int) {
	ls.stack.pc += n
}

func (ls *LuaState) fetch() uint32 {
	i := ls.stack.closure.proto.Code[ls.stack.pc]
	ls.stack.pc++
	return i
}

func (ls *LuaState) getConst(idx int) {
	c := ls.stack.closure.proto.Constants[idx]
	switch x := c.(type) {
	case nil:
		ls.stack.push(LuaNil)
	case string:
		ls.stack.push(LuaString(x))
	case int:
		ls.stack.push(LuaNumber(x))
	case int64:
		ls.stack.push(LuaNumber(x))
	case float64:
		ls.stack.push(LuaNumber(x))
	default:
		panic(fmt.Errorf("const type:%v error", reflect.TypeOf(c).Name()) )
	}
}

func (ls *LuaState) getRK(rk int) {
	if rk > 0xFF { // constant
		ls.getConst(rk & 0xFF)
	} else { // register
		luaPushValue(ls, rk+1)
	}
}

func (ls *LuaState) registerCount() int {
	return int(ls.stack.closure.proto.MaxStackSize)
}

func (ls *LuaState) loadVararg(n int) {
	if n < 0 {
		n = len(ls.stack.varargs)
	}

	ls.stack.check(n)
	ls.stack.pushN(ls.stack.varargs, n)
}

func (ls *LuaState) loadProto(idx int) {
	stack := ls.stack
	subProto := stack.closure.proto.Protos[idx]
	closure := newLuaClosure(subProto)
	stack.push(closure)

	for i, uvInfo := range subProto.Upvalues {
		uvIdx := int(uvInfo.Idx)
		if uvInfo.Instack == 1 {
			if stack.openuvs == nil {
				stack.openuvs = map[int]*upvalue{}
			}

			if openuv, found := stack.openuvs[uvIdx]; found {
				closure.upvals[i] = openuv
			} else {
				closure.upvals[i] = &upvalue{&stack.slots[uvIdx]}
				stack.openuvs[uvIdx] = closure.upvals[i]
			}
		} else {
			closure.upvals[i] = stack.closure.upvals[uvIdx]
		}
	}
}

// 退出当前域时关闭外部变量表
func (ls *LuaState) closeUpvalues(a int) {
	for i, openuv := range ls.stack.openuvs {
		if i >= a-1 {
			val := *openuv.val
			openuv.val = &val
			delete(ls.stack.openuvs, i)
		}
	}
}

// R(A+1) := R(B); R(A) := R(B)[RK(C)]
func self(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	luaCopy(ls, b, a+1)
	ls.getRK(c)
	luaGetTable(ls, b)
	luaReplace(ls, a)
}

// R(A) := Closure(KPROTO[Bx])
func closure(i Instruction, ls *LuaState) {
	a, bx := i.ABx()
	a += 1

	ls.loadProto(bx)
	luaReplace(ls, a)
}

// R(A), R(A+1), ..., R(A+B-2) = vararg
func vararg(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1

	if b != 1 { // b==0 or b>1
		ls.loadVararg(b - 1)
		_popResults(a, b, ls)
	}
}

// R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2));
func tForCall(i Instruction, ls *LuaState) {
	a, _, c := i.ABC()
	a += 1

	_pushFuncAndArgs(a, 3, ls)
	ls.Call(2, c)
	_popResults(a+3, c+1, ls)
}

// return R(A)(R(A+1), ... ,R(A+B-1))
func tailCall(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1

	// todo: optimize tail call!
	c := 0
	nArgs := _pushFuncAndArgs(a, b, ls)
	ls.Call(nArgs, c-1)
	_popResults(a, c, ls)
}

// R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
func call(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1

	// println(":::"+ ls.StackToString())
	nArgs := _pushFuncAndArgs(a, b, ls)
	ls.Call(nArgs, c-1)
	_popResults(a, c, ls)
}

func _pushFuncAndArgs(a, b int, ls *LuaState) (nArgs int) {
	if b >= 1 {
		luaCheckStack(ls, b)
		for i := a; i < a+b; i++ {
			luaPushValue(ls, i)
		}
		return b - 1
	} else {
		_fixStack(a, ls)
		return luaGetTop(ls) - ls.registerCount() - 1
	}
}

func _fixStack(a int, ls *LuaState) {
	x := int(luaToInteger(ls, -1))
	luaPop(ls, 1)

	luaCheckStack(ls, x-a)
	for i := a; i < x; i++ {
		luaPushValue(ls, i)
	}
	luaRotate(ls, ls.registerCount()+1, x-a)
}

func _popResults(a, c int, ls *LuaState) {
	if c == 1 {
		// no results
	} else if c > 1 {
		for i := a + c - 2; i >= a; i-- {
			luaReplace(ls, i)
		}
	} else {
		// leave results on stack
		luaCheckStack(ls, 1)
		ls.Push(LuaNumber(a))
	}
}

// return R(A), ... ,R(A+B-2)
func _return(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1

	if b == 1 {
		// no return values
	} else if b > 1 {
		// b-1 return values
		luaCheckStack(ls, b-1)
		for i := a; i <= a+b-2; i++ {
			luaPushValue(ls, i)
		}
	} else {
		_fixStack(a, ls)
	}
}

// R(A)-=R(A+2); pc+=sBx
func forPrep(i Instruction, ls *LuaState) {
	a, sBx := i.AsBx()
	a += 1

	if luaType(ls, a) == LUA_TSTRING {
		ls.Push(LuaNumber(luaToNumber(ls, a)))
		luaReplace(ls, a)
	}
	if luaType(ls, a+1) == LUA_TSTRING {
		ls.Push(LuaNumber(luaToNumber(ls, a+1)))
		luaReplace(ls, a+1)
	}
	if luaType(ls, a+2) == LUA_TSTRING {
		ls.Push(LuaNumber(luaToNumber(ls, a+2)))
		luaReplace(ls, a+2)
	}

	luaPushValue(ls, a)
	luaPushValue(ls, a+2)
	luaArith(ls, LUA_OPSUB)
	luaReplace(ls, a)
	ls.addPC(sBx)
}

// R(A)+=R(A+2);
// if R(A) <?= R(A+1) then {
//   pc+=sBx; R(A+3)=R(A)
// }
func forLoop(i Instruction, ls *LuaState) {
	a, sBx := i.AsBx()
	a += 1

	// R(A)+=R(A+2);
	luaPushValue(ls, a+2)
	luaPushValue(ls, a)
	luaArith(ls, LUA_OPADD)
	luaReplace(ls, a)

	isPositiveStep := luaToNumber(ls, a+2) >= 0
	if isPositiveStep && luaCompare(ls, a, a+1, LUA_OPLE) ||
		!isPositiveStep && luaCompare(ls, a+1, a, LUA_OPLE) {

		// pc+=sBx; R(A+3)=R(A)
		ls.addPC(sBx)
		luaCopy(ls, a, a+3)
	}
}

// if R(A+1) ~= nil then {
//   R(A)=R(A+1); pc += sBx
// }
func tForLoop(i Instruction, ls *LuaState) {
	a, sBx := i.AsBx()
	a += 1

	if !luaIsNil(ls, a+1) {
		luaCopy(ls, a+1, a)
		ls.addPC(sBx)
	}
}

// R(A), R(A+1), ..., R(A+B) := nil
func loadNil(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1

	ls.Push(LuaNil)
	for i := a; i <= a+b; i++ {
		luaCopy(ls, -1, i)
	}
	luaPop(ls, 1)
}

// R(A) := (bool)B; if (C) pc++
func loadBool(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1

	ls.Push(LuaBool(b != 0))
	luaReplace(ls, a)

	if c != 0 {
		ls.addPC(1)
	}
}

// R(A) := Kst(Bx)
func loadK(i Instruction, ls *LuaState) {
	a, bx := i.ABx()
	a += 1

	ls.getConst(bx)
	luaReplace(ls, a)
}

// R(A) := Kst(extra arg)
func loadKx(i Instruction, ls *LuaState) {
	a, _ := i.ABx()
	a += 1
	ax := Instruction(ls.fetch()).Ax()

	//luaCheckStack(ls, 1)
	ls.getConst(ax)
	luaReplace(ls, a)
}

// R(A) := R(B)
func move(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	luaCopy(ls, b, a)
}

// pc+=sBx; if (A) close all upvalues >= R(A - 1)
func jmp(i Instruction, ls *LuaState) {
	a, sBx := i.AsBx()

	ls.addPC(sBx)
	if a != 0 {
		ls.closeUpvalues(a)
	}
}

/* arith */

func add(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPADD) }  // +
func sub(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPSUB) }  // -
func mul(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPMUL) }  // *
func mod(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPMOD) }  // %
func pow(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPPOW) }  // ^
func div(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPDIV) }  // /
func idiv(i Instruction, ls *LuaState) { _binaryArith(i, ls, LUA_OPIDIV) } // //
func band(i Instruction, ls *LuaState) { _binaryArith(i, ls, LUA_OPBAND) } // &
func bor(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPBOR) }  // |
func bxor(i Instruction, ls *LuaState) { _binaryArith(i, ls, LUA_OPBXOR) } // ~
func shl(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPSHL) }  // <<
func shr(i Instruction, ls *LuaState)  { _binaryArith(i, ls, LUA_OPSHR) }  // >>
func unm(i Instruction, ls *LuaState)  { _unaryArith(i, ls, LUA_OPUNM) }   // -
func bnot(i Instruction, ls *LuaState) { _unaryArith(i, ls, LUA_OPBNOT) }  // ~

// R(A) := RK(B) op RK(C)
func _binaryArith(i Instruction, ls *LuaState, op ArithOp) {
	a, b, c := i.ABC()
	a += 1

	ls.getRK(b)
	ls.getRK(c)
	luaArith(ls, op)
	luaReplace(ls, a)
}

// R(A) := op R(B)
func _unaryArith(i Instruction, ls *LuaState, op ArithOp) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	luaPushValue(ls, b)
	luaArith(ls, op)
	luaReplace(ls, a)
}

/* compare */

func eq(i Instruction, ls *LuaState) { _compare(i, ls, LUA_OPEQ) } // ==
func lt(i Instruction, ls *LuaState) { _compare(i, ls, LUA_OPLT) } // <
func le(i Instruction, ls *LuaState) { _compare(i, ls, LUA_OPLE) } // <=

// if ((RK(B) op RK(C)) ~= A) then pc++
func _compare(i Instruction, ls *LuaState, op CompareOp) {
	a, b, c := i.ABC()

	ls.getRK(b)
	ls.getRK(c)
	if luaCompare(ls, -2, -1, op) != (a != 0) {
		ls.addPC(1)
	}
	luaPop(ls, 2)
}

/* logical */

// R(A) := not R(B)
func not(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	ls.Push(LuaBool(!luaToBoolean(ls, b)))
	luaReplace(ls, a)
}

// if not (R(A) <=> C) then pc++
func test(i Instruction, ls *LuaState) {
	a, _, c := i.ABC()
	a += 1

	if luaToBoolean(ls, a) != (c != 0) {
		ls.addPC(1)
	}
}

// if (R(B) <=> C) then R(A) := R(B) else pc++
func testSet(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	if luaToBoolean(ls, b) == (c != 0) {
		luaCopy(ls, b, a)
	} else {
		ls.addPC(1)
	}
}

/* len & concat */

// R(A) := length of R(B)
func length(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	luaLen(ls, b)
	luaReplace(ls, a)
}

// R(A) := R(B).. ... ..R(C)
func concat(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1

	n := c - b + 1
	luaCheckStack(ls, n)
	for i := b; i <= c; i++ {
		luaPushValue(ls, i)
	}
	luaConcat(ls, n)
	luaReplace(ls, a)
}

/* number of list items to accumulate before a SETLIST instruction */
const LFIELDS_PER_FLUSH = 50

// R(A) := {} (size = B,C)
func newTable(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1

	ls.createTable(number.Fb2int(b), number.Fb2int(c))
	luaReplace(ls, a)
}

// R(A) := R(B)[RK(C)]
func getTable(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	ls.getRK(c)
	luaGetTable(ls, b)
	luaReplace(ls, a)
}

// R(A)[RK(B)] := RK(C)
func setTable(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1

	ls.getRK(b)
	ls.getRK(c)
	luaSetTable(ls, a)
}

// R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B
func setList(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1

	if c > 0 {
		c = c - 1
	} else {
		c = Instruction(ls.fetch()).Ax()
	}

	bIsZero := b == 0
	if bIsZero {
		b = int(luaToInteger(ls, -1)) - a - 1
		luaPop(ls, 1)
	}

	luaCheckStack(ls, 1)
	idx := int64(c * LFIELDS_PER_FLUSH)
	for j := 1; j <= b; j++ {
		idx++
		luaPushValue(ls, a+j)
		luaSetI(ls, a, idx)
	}

	if bIsZero {
		for j := ls.registerCount() + 1; j <= luaGetTop(ls); j++ {
			idx++
			luaPushValue(ls, j)
			luaSetI(ls, a, idx)
		}

		// clear stack
		luaSetTop(ls, ls.registerCount())
	}
}

// R(A) := UpValue[B]
func getUpval(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	luaCopy(ls, luaUpvalueIndex(b), a)
}

// UpValue[B] := R(A)
func setUpval(i Instruction, ls *LuaState) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	luaCopy(ls, a, luaUpvalueIndex(b))
}

// R(A) := UpValue[B][RK(C)]
func getTabUp(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	ls.getRK(c)
	luaGetTable(ls, luaUpvalueIndex(b))
	luaReplace(ls, a)
}

// UpValue[A][RK(B)] := RK(C)
func setTabUp(i Instruction, ls *LuaState) {
	a, b, c := i.ABC()
	a += 1

	ls.getRK(b)
	ls.getRK(c)
	luaSetTable(ls, luaUpvalueIndex(a))
}
