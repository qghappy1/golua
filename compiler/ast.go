package compiler



// chunk ::= block
// type Chunk *Block

// block ::= {stat} [retstat]
// retstat ::= return [explist] [‘;’]
// explist ::= exp {‘,’ exp}
type Block struct {
	LastLine int
	Stats    []Stat
	RetExps  []Exp
}

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp

prefixexp ::= var | functioncall | ‘(’ exp ‘)’

var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name

functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
*/

type Exp interface{}

type NilExp struct{ Line int }    // nil
type TrueExp struct{ Line int }   // true
type FalseExp struct{ Line int }  // false
type VarargExp struct{ Line int } // ...

// Numeral
type IntegerExp struct {
	Line int
	Val  int64
}
type FloatExp struct {
	Line int
	Val  float64
}

// LiteralString
type StringExp struct {
	Line int
	Str  string
}

// unop exp
type UnopExp struct {
	Line int // line of operator
	Op   int // operator
	Exp  Exp
}

// exp1 op exp2
type BinopExp struct {
	Line int // line of operator
	Op   int // operator
	Exp1 Exp
	Exp2 Exp
}

type ConcatExp struct {
	Line int // line of last ..
	Exps []Exp
}

// tableconstructor ::= ‘{’ [fieldlist] ‘}’
// fieldlist ::= field {fieldsep field} [fieldsep]
// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
// fieldsep ::= ‘,’ | ‘;’
type TableConstructorExp struct {
	Line     int // line of `{` ?
	LastLine int // line of `}`
	KeyExps  []Exp
	ValExps  []Exp
}

// functiondef ::= function funcbody
// funcbody ::= ‘(’ [parlist] ‘)’ block end
// parlist ::= namelist [‘,’ ‘...’] | ‘...’
// namelist ::= Name {‘,’ Name}
type FuncDefExp struct {
	Line     int
	LastLine int // line of `end`
	ParList  []string
	IsVararg bool
	Block    *Block
}

/*
prefixexp ::= Name |
              ‘(’ exp ‘)’ |
              prefixexp ‘[’ exp ‘]’ |
              prefixexp ‘.’ Name |
              prefixexp ‘:’ Name args |
              prefixexp args
*/

type NameExp struct {
	Line int
	Name string
}

type ParensExp struct {
	Exp Exp
}

type TableAccessExp struct {
	LastLine  int // line of `]` ?
	PrefixExp Exp
	KeyExp    Exp
}

type FuncCallExp struct {
	Line      int // line of `(` ?
	LastLine  int // line of ')'
	PrefixExp Exp
	NameExp   *StringExp
	Args      []Exp
}

/*
stat ::=  ‘;’ |
	 varlist ‘=’ explist |
	 functioncall |
	 label |
	 break |
	 goto Name |
	 do block end |
	 while exp do block end |
	 repeat block until exp |
	 if exp then block {elseif exp then block} [else block] end |
	 for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end |
	 for namelist in explist do block end |
	 function funcname funcbody |
	 local function Name funcbody |
	 local namelist [‘=’ explist]
*/
type Stat interface{}

type EmptyStat struct{}              // ‘;’
type BreakStat struct{ Line int }    // break
type LabelStat struct{ Name string } // ‘::’ Name ‘::’
type GotoStat struct{ Name string }  // goto Name
type DoStat struct{ Block *Block }   // do block end
type FuncCallStat = FuncCallExp      // functioncall

// if exp then block {elseif exp then block} [else block] end
type IfStat struct {
	Exps   []Exp
	Blocks []*Block
}

// while exp do block end
type WhileStat struct {
	Exp   Exp
	Block *Block
}

// repeat block until exp
type RepeatStat struct {
	Block *Block
	Exp   Exp
}

// for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
type ForNumStat struct {
	LineOfFor int
	LineOfDo  int
	VarName   string
	InitExp   Exp
	LimitExp  Exp
	StepExp   Exp
	Block     *Block
}

// for namelist in explist do block end
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
type ForInStat struct {
	LineOfDo int
	NameList []string
	ExpList  []Exp
	Block    *Block
}

// varlist ‘=’ explist
// varlist ::= var {‘,’ var}
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
type AssignStat struct {
	LastLine int
	VarList  []Exp
	ExpList  []Exp
}

// local namelist [‘=’ explist]
// namelist ::= Name {‘,’ Name}
// explist ::= exp {‘,’ exp}
type LocalVarDeclStat struct {
	LastLine int
	NameList []string
	ExpList  []Exp
}

// local function Name funcbody
type LocalFuncDefStat struct {
	Name string
	Exp  *FuncDefExp
}