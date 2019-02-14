package golua

import (
	"math"
	"math/rand"
	"golua/number"
)

var mathLib = map[string]GoFunction{
	"random":     mathRandom,
	"randomseed": mathRandomSeed,
	"max":        mathMax,
	"min":        mathMin,
	"exp":        mathExp,
	"log":        mathLog,
	"deg":        mathDeg,
	"rad":        mathRad,
	"sin":        mathSin,
	"cos":        mathCos,
	"tan":        mathTan,
	"asin":       mathAsin,
	"acos":       mathAcos,
	"atan":       mathAtan,
	"ceil":       mathCeil,
	"floor":      mathFloor,
	"fmod":       mathFmod,
	"modf":       mathModf,
	"abs":        mathAbs,
	"sqrt":       mathSqrt,
	"ult":        mathUlt,
	"tointeger":  mathToInt,
	"type":       mathType,
	/* placeholders */
	"pi":         nil,
	"huge":       nil,
	"maxinteger": nil,
	"mininteger": nil,
}

func OpenMathLib(ls *LuaState) int {
	ls.NewLib(mathLib)
	ls.Push(LuaNumber(math.Pi))
	luaSetField(ls, -2, "pi")
	ls.Push(LuaNumber(math.Inf(1)))
	luaSetField(ls, -2, "huge")
	ls.Push(LuaNumber(math.MaxInt64))
	luaSetField(ls, -2, "maxinteger")
	ls.Push(LuaNumber(math.MinInt64))
	luaSetField(ls, -2, "mininteger")
	return 1
}

/* pseudo-random numbers */

// math.random ([m [, n]])
// http://www.lua.org/manual/5.3/manual.html#pdf-math.random
// lua-5.3.4/src/lmathlib.c#math_random()
func mathRandom(ls *LuaState) int {
	var low, up int64
	switch luaGetTop(ls) { /* check number of arguments */
	case 0: /* no arguments */
		ls.Push(LuaNumber(rand.Float64())) /* Number between 0 and 1 */
		return 1
	case 1: /* only upper limit */
		low = 1
		up = ls.CheckInteger(1)
	case 2: /* lower and upper limits */
		low = ls.CheckInteger(1)
		up = ls.CheckInteger(2)
	default:
		return ls.Error2("wrong number of arguments")
	}

	/* random integer in the interval [low, up] */
	ls.ArgCheck(low <= up, 1, "interval is empty")
	ls.ArgCheck(low >= 0 || up <= math.MaxInt64+low, 1,
		"interval too large")
	if up-low == math.MaxInt64 {
		ls.Push(LuaNumber(low + rand.Int63()))
	} else {
		ls.Push(LuaNumber(low + rand.Int63n(up-low+1)))
	}
	return 1
}

// math.randomseed (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.randomseed
// lua-5.3.4/src/lmathlib.c#math_randomseed()
func mathRandomSeed(ls *LuaState) int {
	x := ls.CheckNumber(1)
	rand.Seed(int64(x))
	return 0
}

/* max & min */

// math.max (x, ···)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.max
// lua-5.3.4/src/lmathlib.c#math_max()
func mathMax(ls *LuaState) int {
	n := luaGetTop(ls) /* number of arguments */
	imax := 1        /* index of current maximum value */
	ls.ArgCheck(n >= 1, 1, "value expected")
	for i := 2; i <= n; i++ {
		if luaCompare(ls, imax, i, LUA_OPLT) {
			imax = i
		}
	}
	luaPushValue(ls, imax)
	return 1
}

// math.min (x, ···)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.min
// lua-5.3.4/src/lmathlib.c#math_min()
func mathMin(ls *LuaState) int {
	n := luaGetTop(ls) /* number of arguments */
	imin := 1        /* index of current minimum value */
	ls.ArgCheck(n >= 1, 1, "value expected")
	for i := 2; i <= n; i++ {
		if luaCompare(ls, i, imin, LUA_OPLT) {
			imin = i
		}
	}
	luaPushValue(ls, imin)
	return 1
}

/* exponentiation and logarithms */

// math.exp (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.exp
// lua-5.3.4/src/lmathlib.c#math_exp()
func mathExp(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(math.Exp(x)))
	return 1
}

// math.log (x [, base])
// http://www.lua.org/manual/5.3/manual.html#pdf-math.log
// lua-5.3.4/src/lmathlib.c#math_log()
func mathLog(ls *LuaState) int {
	x := ls.CheckNumber(1)
	var res float64

	if luaIsNoneOrNil(ls, 2) {
		res = math.Log(x)
	} else {
		base := luaToNumber(ls, 2)
		if base == 2 {
			res = math.Log2(x)
		} else if base == 10 {
			res = math.Log10(x)
		} else {
			res = math.Log(x) / math.Log(base)
		}
	}

	ls.Push(LuaNumber(res))
	return 1
}

/* trigonometric functions */

// math.deg (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.deg
// lua-5.3.4/src/lmathlib.c#math_deg()
func mathDeg(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(x * 180 / math.Pi))
	return 1
}

// math.rad (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.rad
// lua-5.3.4/src/lmathlib.c#math_rad()
func mathRad(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(x * math.Pi / 180))
	return 1
}

// math.sin (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.sin
// lua-5.3.4/src/lmathlib.c#math_sin()
func mathSin(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(math.Sin(x)))
	return 1
}

// math.cos (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.cos
// lua-5.3.4/src/lmathlib.c#math_cos()
func mathCos(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(math.Cos(x)))
	return 1
}

// math.tan (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.tan
// lua-5.3.4/src/lmathlib.c#math_tan()
func mathTan(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(math.Tan(x)))
	return 1
}

// math.asin (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.asin
// lua-5.3.4/src/lmathlib.c#math_asin()
func mathAsin(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(math.Asin(x)))
	return 1
}

// math.acos (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.acos
// lua-5.3.4/src/lmathlib.c#math_acos()
func mathAcos(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(math.Acos(x)))
	return 1
}

// math.atan (y [, x])
// http://www.lua.org/manual/5.3/manual.html#pdf-math.atan
// lua-5.3.4/src/lmathlib.c#math_atan()
func mathAtan(ls *LuaState) int {
	y := ls.CheckNumber(1)
	x := luaOptNumber(ls, 2, 1.0)
	ls.Push(LuaNumber(math.Atan2(y, x)))
	return 1
}

/* rounding functions */

// math.ceil (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.ceil
// lua-5.3.4/src/lmathlib.c#math_ceil()
func mathCeil(ls *LuaState) int {
	if luaIsInteger(ls, 1) {
		luaSetTop(ls, 1) /* integer is its own ceil */
	} else {
		x := ls.CheckNumber(1)
		_pushNumInt(ls, math.Ceil(x))
	}
	return 1
}

// math.floor (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.floor
// lua-5.3.4/src/lmathlib.c#math_floor()
func mathFloor(ls *LuaState) int {
	if luaIsInteger(ls, 1) {
		luaSetTop(ls, 1) /* integer is its own floor */
	} else {
		x := ls.CheckNumber(1)
		_pushNumInt(ls, math.Floor(x))
	}
	return 1
}

/* others */

// math.fmod (x, y)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.fmod
// lua-5.3.4/src/lmathlib.c#math_fmod()
func mathFmod(ls *LuaState) int {
	if luaIsInteger(ls, 1) && luaIsInteger(ls, 2) {
		d := luaToInteger(ls, 2)
		if uint64(d)+1 <= 1 { /* special cases: -1 or 0 */
			ls.ArgCheck(d != 0, 2, "zero")
			ls.Push(LuaNumber(0)) /* avoid overflow with 0x80000... / -1 */
		} else {
			ls.Push(LuaNumber(luaToInteger(ls, 1) % d))
		}
	} else {
		x := ls.CheckNumber(1)
		y := ls.CheckNumber(2)
		ls.Push(LuaNumber(x - math.Trunc(x/y)*y))
	}

	return 1
}

// math.modf (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.modf
// lua-5.3.4/src/lmathlib.c#math_modf()
func mathModf(ls *LuaState) int {
	if luaIsInteger(ls, 1) {
		luaSetTop(ls, 1)     /* number is its own integer part */
		ls.Push(LuaNumber(0)) /* no fractional part */
	} else {
		x := ls.CheckNumber(1)
		i, f := math.Modf(x)
		_pushNumInt(ls, i)
		if math.IsInf(x, 0) {
			ls.Push(LuaNumber(0))
		} else {
			ls.Push(LuaNumber(f))
		}
	}

	return 2
}

// math.abs (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.abs
// lua-5.3.4/src/lmathlib.c#math_abs()
func mathAbs(ls *LuaState) int {
	if luaIsInteger(ls, 1) {
		x := luaToInteger(ls, 1)
		if x < 0 {
			ls.Push(LuaNumber(-x))
		}
	} else {
		x := ls.CheckNumber(1)
		ls.Push(LuaNumber(math.Abs(x)))
	}
	return 1
}

// math.sqrt (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.sqrt
// lua-5.3.4/src/lmathlib.c#math_sqrt()
func mathSqrt(ls *LuaState) int {
	x := ls.CheckNumber(1)
	ls.Push(LuaNumber(math.Sqrt(x)))
	return 1
}

// math.ult (m, n)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.ult
// lua-5.3.4/src/lmathlib.c#math_ult()
func mathUlt(ls *LuaState) int {
	m := ls.CheckInteger(1)
	n := ls.CheckInteger(2)
	ls.Push(LuaBool(uint64(m) < uint64(n)))
	return 1
}

// math.tointeger (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.tointeger
// lua-5.3.4/src/lmathlib.c#math_toint()
func mathToInt(ls *LuaState) int {
	if i, ok := luaToIntegerX(ls, 1); ok {
		ls.Push(LuaNumber(i))
	} else {
		ls.CheckAny(1)
		ls.Push(LuaNil) /* value is not convertible to integer */
	}
	return 1
}

// math.type (x)
// http://www.lua.org/manual/5.3/manual.html#pdf-math.type
// lua-5.3.4/src/lmathlib.c#math_type()
func mathType(ls *LuaState) int {
	if luaType(ls, 1) == LUA_TNUMBER {
		if luaIsInteger(ls, 1) {
			ls.Push(LuaString("integer"))
		} else {
			ls.Push(LuaString("float"))
		}
	} else {
		ls.CheckAny(1)
		ls.Push(LuaNil)
	}
	return 1
}

func _pushNumInt(ls *LuaState, d float64) {
	if i, ok := number.FloatToInteger(d); ok { /* does 'd' fit in an integer? */
		ls.Push(LuaNumber(i)) /* result is integer */
	} else {
		ls.Push(LuaNumber(d)) /* result is float */
	}
}
