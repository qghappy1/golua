package golua

import (
	. "luago/compiler"
)

const MAXARG_Bx = 1<<18 - 1       // 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 131071

/*
 31       22       13       5    0
  +-------+^------+-^-----+-^-----
  |b=9bits |c=9bits |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |    bx=18bits    |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |   sbx=18bits    |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |    ax=26bits            |op=6|
  +-------+^------+-^-----+-^-----
 31      23      15       7      0
*/
type Instruction uint32

func (inst Instruction) Opcode() int {
	return int(inst & 0x3F)
}

func (inst Instruction) ABC() (a, b, c int) {
	a = int(inst >> 6 & 0xFF)
	c = int(inst >> 14 & 0x1FF)
	b = int(inst >> 23 & 0x1FF)
	return
}

func (inst Instruction) ABx() (a, bx int) {
	a = int(inst >> 6 & 0xFF)
	bx = int(inst >> 14)
	return
}

func (inst Instruction) AsBx() (a, sbx int) {
	a, bx := inst.ABx()
	return a, bx - MAXARG_sBx
}

func (inst Instruction) Ax() int {
	return int(inst >> 6)
}

func (inst Instruction) OpName() string {
	return opcodes[inst.Opcode()].name
}

func (inst Instruction) OpMode() byte {
	return opcodes[inst.Opcode()].opMode
}

func (inst Instruction) BMode() byte {
	return opcodes[inst.Opcode()].argBMode
}

func (inst Instruction) CMode() byte {
	return opcodes[inst.Opcode()].argCMode
}

func (inst Instruction) Execute(ls *LuaState) {
	action := opcodes[inst.Opcode()].action
	if action != nil {
		action(inst, ls)
	} else {
		ls.raiseError1(inst.OpName())
	}
}

type opcode struct {
	testFlag byte // operator is a test (next instruction must be a jump)
	setAFlag byte // instruction set register A
	argBMode byte // B arg mode
	argCMode byte // C arg mode
	opMode   byte // op mode
	name     string
	action   func(i Instruction, vm *LuaState)
}

var opcodes = []opcode{}

func init() {
	opcodes = []opcode{
		/*     T  A    B       C     mode         name       action */
		opcode{0, 1, OpArgR, OpArgN, IABC /* */, "MOVE    ", move},     // R(A) := R(B)
		opcode{0, 1, OpArgK, OpArgN, IABx /* */, "LOADK   ", loadK},    // R(A) := Kst(Bx)
		opcode{0, 1, OpArgN, OpArgN, IABx /* */, "LOADKX  ", loadKx},   // R(A) := Kst(extra arg)
		opcode{0, 1, OpArgU, OpArgU, IABC /* */, "LOADBOOL", loadBool}, // R(A) := (bool)B; if (C) pc++
		opcode{0, 1, OpArgU, OpArgN, IABC /* */, "LOADNIL ", loadNil},  // R(A), R(A+1), ..., R(A+B) := nil
		opcode{0, 1, OpArgU, OpArgN, IABC /* */, "GETUPVAL", getUpval}, // R(A) := UpValue[B]
		opcode{0, 1, OpArgU, OpArgK, IABC /* */, "GETTABUP", getTabUp}, // R(A) := UpValue[B][RK(C)]
		opcode{0, 1, OpArgR, OpArgK, IABC /* */, "GETTABLE", getTable}, // R(A) := R(B)[RK(C)]
		opcode{0, 0, OpArgK, OpArgK, IABC /* */, "SETTABUP", setTabUp}, // UpValue[A][RK(B)] := RK(C)
		opcode{0, 0, OpArgU, OpArgN, IABC /* */, "SETUPVAL", setUpval}, // UpValue[B] := R(A)
		opcode{0, 0, OpArgK, OpArgK, IABC /* */, "SETTABLE", setTable}, // R(A)[RK(B)] := RK(C)
		opcode{0, 1, OpArgU, OpArgU, IABC /* */, "NEWTABLE", newTable}, // R(A) := {} (size = B,C)
		opcode{0, 1, OpArgR, OpArgK, IABC /* */, "SELF    ", self},     // R(A+1) := R(B); R(A) := R(B)[RK(C)]
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "ADD     ", add},      // R(A) := RK(B) + RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "SUB     ", sub},      // R(A) := RK(B) - RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "MUL     ", mul},      // R(A) := RK(B) * RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "MOD     ", mod},      // R(A) := RK(B) % RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "POW     ", pow},      // R(A) := RK(B) ^ RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "DIV     ", div},      // R(A) := RK(B) / RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "IDIV    ", idiv},     // R(A) := RK(B) // RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "BAND    ", band},     // R(A) := RK(B) & RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "BOR     ", bor},      // R(A) := RK(B) | RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "BXOR    ", bxor},     // R(A) := RK(B) ~ RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "SHL     ", shl},      // R(A) := RK(B) << RK(C)
		opcode{0, 1, OpArgK, OpArgK, IABC /* */, "SHR     ", shr},      // R(A) := RK(B) >> RK(C)
		opcode{0, 1, OpArgR, OpArgN, IABC /* */, "UNM     ", unm},      // R(A) := -R(B)
		opcode{0, 1, OpArgR, OpArgN, IABC /* */, "BNOT    ", bnot},     // R(A) := ~R(B)
		opcode{0, 1, OpArgR, OpArgN, IABC /* */, "NOT     ", not},      // R(A) := not R(B)
		opcode{0, 1, OpArgR, OpArgN, IABC /* */, "LEN     ", length},   // R(A) := length of R(B)
		opcode{0, 1, OpArgR, OpArgR, IABC /* */, "CONCAT  ", concat},   // R(A) := R(B).. ... ..R(C)
		opcode{0, 0, OpArgR, OpArgN, IAsBx /**/, "JMP     ", jmp},      // pc+=sBx; if (A) close all upvalues >= R(A - 1)
		opcode{1, 0, OpArgK, OpArgK, IABC /* */, "EQ      ", eq},       // if ((RK(B) == RK(C)) ~= A) then pc++
		opcode{1, 0, OpArgK, OpArgK, IABC /* */, "LT      ", lt},       // if ((RK(B) <  RK(C)) ~= A) then pc++
		opcode{1, 0, OpArgK, OpArgK, IABC /* */, "LE      ", le},       // if ((RK(B) <= RK(C)) ~= A) then pc++
		opcode{1, 0, OpArgN, OpArgU, IABC /* */, "TEST    ", test},     // if not (R(A) <=> C) then pc++
		opcode{1, 1, OpArgR, OpArgU, IABC /* */, "TESTSET ", testSet},  // if (R(B) <=> C) then R(A) := R(B) else pc++
		opcode{0, 1, OpArgU, OpArgU, IABC /* */, "CALL    ", call},     // R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))
		opcode{0, 1, OpArgU, OpArgU, IABC /* */, "TAILCALL", tailCall}, // return R(A)(R(A+1), ... ,R(A+B-1))
		opcode{0, 0, OpArgU, OpArgN, IABC /* */, "RETURN  ", _return},  // return R(A), ... ,R(A+B-2)
		opcode{0, 1, OpArgR, OpArgN, IAsBx /**/, "FORLOOP ", forLoop},  // R(A)+=R(A+2); if R(A) <?= R(A+1) then { pc+=sBx; R(A+3)=R(A) }
		opcode{0, 1, OpArgR, OpArgN, IAsBx /**/, "FORPREP ", forPrep},  // R(A)-=R(A+2); pc+=sBx
		opcode{0, 0, OpArgN, OpArgU, IABC /* */, "TFORCALL", tForCall}, // R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2));
		opcode{0, 1, OpArgR, OpArgN, IAsBx /**/, "TFORLOOP", tForLoop}, // if R(A+1) ~= nil then { R(A)=R(A+1); pc += sBx }
		opcode{0, 0, OpArgU, OpArgU, IABC /* */, "SETLIST ", setList},  // R(A)[(C-1)*FPF+i] := R(A+i), 1 <= i <= B
		opcode{0, 1, OpArgU, OpArgN, IABx /* */, "CLOSURE ", closure},  // R(A) := Closure(KPROTO[Bx])
		opcode{0, 1, OpArgU, OpArgN, IABC /* */, "VARARG  ", vararg},   // R(A), R(A+1), ..., R(A+B-2) = vararg
		opcode{0, 0, OpArgU, OpArgU, IAx /*  */, "EXTRAARG", nil},      // extra (larger) argument for previous opcode
	}
}
