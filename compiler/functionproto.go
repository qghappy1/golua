package compiler

const (
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x53
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = 0x5678
	LUAC_NUM         = 370.5
)

const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_NUMBER    = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type binaryChunk struct {
	header
	sizeUpvalues byte // ?
	mainFunc     *FunctionProto
}

type header struct {
	signature       [4]byte
	version         byte
	format          byte
	luacData        [6]byte
	cintSize        byte
	sizetSize       byte
	instructionSize byte
	luaIntegerSize  byte
	luaNumberSize   byte
	luacInt         int64
	luacNum         float64
}

type DbgLocVar struct {
	VarName string
	StartPC int
	EndPC   int
}

type DbgCall struct {
	Name string
	Pc   int
}

// function prototype
type FunctionProto struct {
	Source          string // debug
	LineDefined     uint32
	LastLineDefined uint32
	NumParams       byte
	IsVararg        byte
	MaxStackSize    byte
	Code            []uint32
	Constants       []interface{}
	Upvalues        []Upvalue
	Protos          []*FunctionProto

	DbgSourcePositions []uint32
	DbgLocVars         []DbgLocVar
	DbgCalls           []DbgCall
	DbgUpvalues        []string
}

func (fp *FunctionProto) LocalName(regno, pc int) (string, bool) {
	for i := 0; i < len(fp.DbgLocVars) && fp.DbgLocVars[i].StartPC < pc; i++ {
		if pc < fp.DbgLocVars[i].EndPC {
			regno--
			if regno == 0 {
				return fp.DbgLocVars[i].VarName, true
			}
		}
	}
	return "", false
}

type Upvalue struct {
	Instack byte
	Idx     byte
}

func isBinaryChunk(data []byte) bool {
	return len(data) > 4 &&
		string(data[:4]) == LUA_SIGNATURE
}

func undump(data []byte) *FunctionProto {
	reader := &reader{data}
	reader.checkHeader()
	reader.readByte() // size_upvalues
	return reader.readProto("")
}
