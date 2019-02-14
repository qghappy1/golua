package golua

import (
	"fmt"
	"strings"
	"regexp"
)

var strLib = map[string]GoFunction{
	"len":      strLen,
	"rep":      strRep,
	"reverse":  strReverse,
	"lower":    strLower,
	"upper":    strUpper,
	"sub":      strSub,
	"byte":     strByte,
	"char":     strChar,
	"dump":     strDump,
	"format":   strFormat,
	"packsize": strPackSize,
	"pack":     strPack,
	"unpack":   strUnpack,
	"find":     strFind,
	"match":    strMatch,
	"gsub":     strGsub,
	"gmatch":   strGmatch,
}

func OpenStringLib(ls *LuaState) int {
	ls.NewLib(strLib)
	createMetatable(ls)
	return 1
}

func createMetatable(ls *LuaState) {
	ls.createTable(0, 1)       /* table to be metatable for strings */
	ls.Push(LuaString("dummy"))     /* dummy string */
	luaPushValue(ls, -2)           /* copy table */
	luaSetMetatable(ls, -2)        /* set table as metatable for strings */
	luaPop(ls, 1)                  /* pop dummy string */
	luaPushValue(ls, -2)           /* get string library */
	luaSetField(ls, -2, "__index") /* metatable.__index = string */
	luaPop(ls, 1)                  /* pop metatable */
}

/* Basic String Functions */

// string.len (s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.len
// lua-5.3.4/src/lstrlib.c#str_len()
func strLen(ls *LuaState) int {
	s := ls.CheckString(1)
	ls.Push(LuaNumber(len(s)))
	return 1
}

// string.rep (s, n [, sep])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.rep
// lua-5.3.4/src/lstrlib.c#str_rep()
func strRep(ls *LuaState) int {
	s := ls.CheckString(1)
	n := ls.CheckInteger(2)
	sep := luaOptString(ls, 3, "")

	if n <= 0 {
		ls.Push(LuaString(""))
	} else if n == 1 {
		ls.Push(LuaString(s))
	} else {
		a := make([]string, n)
		for i := 0; i < int(n); i++ {
			a[i] = s
		}
		ls.Push(LuaString(strings.Join(a, sep)))
	}

	return 1
}

// string.reverse (s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.reverse
// lua-5.3.4/src/lstrlib.c#str_reverse()
func strReverse(ls *LuaState) int {
	s := ls.CheckString(1)

	if strLen := len(s); strLen > 1 {
		a := make([]byte, strLen)
		for i := 0; i < strLen; i++ {
			a[i] = s[strLen-1-i]
		}
		ls.Push(LuaString(a))
	}

	return 1
}

// string.lower (s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.lower
// lua-5.3.4/src/lstrlib.c#str_lower()
func strLower(ls *LuaState) int {
	s := ls.CheckString(1)
	ls.Push(LuaString(strings.ToLower(s)))
	return 1
}

// string.upper (s)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.upper
// lua-5.3.4/src/lstrlib.c#str_upper()
func strUpper(ls *LuaState) int {
	s := ls.CheckString(1)
	ls.Push(LuaString(strings.ToUpper(s)))
	return 1
}

// string.sub (s, i [, j])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.sub
// lua-5.3.4/src/lstrlib.c#str_sub()
func strSub(ls *LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	i := posRelat(ls.CheckInteger(2), sLen)
	j := posRelat(luaOptInteger(ls, 3, -1), sLen)

	if i < 1 {
		i = 1
	}
	if j > sLen {
		j = sLen
	}

	if i <= j {
		ls.Push(LuaString(s[i-1 : j]))
	} else {
		ls.Push(LuaString(""))
	}

	return 1
}

// string.byte (s [, i [, j]])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.byte
// lua-5.3.4/src/lstrlib.c#str_byte()
func strByte(ls *LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	i := posRelat(luaOptInteger(ls, 2, 1), sLen)
	j := posRelat(luaOptInteger(ls, 3, int64(i)), sLen)

	if i < 1 {
		i = 1
	}
	if j > sLen {
		j = sLen
	}

	if i > j {
		return 0 /* empty interval; return no values */
	}
	//if (j - i >= INT_MAX) { /* arithmetic overflow? */
	//  return ls.Error2("string slice too long")
	//}

	n := j - i + 1
	luaCheckStack2(ls, n, "string slice too long")

	for k := 0; k < n; k++ {
		ls.Push(LuaNumber(s[i+k-1]))
	}
	return n
}

// string.char (···)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.char
// lua-5.3.4/src/lstrlib.c#str_char()
func strChar(ls *LuaState) int {
	nArgs := luaGetTop(ls)

	s := make([]byte, nArgs)
	for i := 1; i <= nArgs; i++ {
		c := ls.CheckInteger(i)
		ls.ArgCheck(int64(byte(c)) == c, i, "value out of range")
		s[i-1] = byte(c)
	}

	ls.Push(LuaString(s))
	return 1
}

// string.dump (function [, strip])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.dump
// lua-5.3.4/src/lstrlib.c#str_dump()
func strDump(ls *LuaState) int {
	// strip := luaToBoolean(ls, 2)
	// luaCheckType(ls, 1, LUA_TFUNCTION)
	// luaSetTop(ls, 1)
	// ls.PushString(string(ls.Dump(strip)))
	// return 1
	panic("todo: strDump!")
}

/* PACK/UNPACK */

// string.packsize (fmt)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.packsize
func strPackSize(ls *LuaState) int {
	fmt := ls.CheckString(1)
	if fmt == "j" {
		ls.Push(LuaNumber(8)) // todo
	} else {
		panic("todo: strPackSize!")
	}
	return 1
}

// string.pack (fmt, v1, v2, ···)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.pack
func strPack(ls *LuaState) int {
	panic("todo: strPack!")
}

// string.unpack (fmt, s [, pos])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.unpack
func strUnpack(ls *LuaState) int {
	panic("todo: strUnpack!")
}

/* STRING FORMAT */

// string.format (formatstring, ···)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.format
func strFormat(ls *LuaState) int {
	fmtStr := ls.CheckString(1)
	if len(fmtStr) <= 1 || strings.IndexByte(fmtStr, '%') < 0 {
		ls.Push(LuaString(fmtStr))
		return 1
	}

	argIdx := 1
	arr := parseFmtStr(fmtStr)
	for i, s := range arr {
		if s[0] == '%' {
			if s == "%%" {
				arr[i] = "%"
			} else {
				argIdx += 1
				arr[i] = _fmtArg(s, ls, argIdx)
			}
		}
	}

	ls.Push(LuaString(strings.Join(arr, "")))
	return 1
}

func _fmtArg(tag string, ls *LuaState, argIdx int) string {
	switch tag[len(tag)-1] { // specifier
	case 'c': // character
		return string([]byte{byte(luaToInteger(ls, argIdx))})
	case 'i':
		tag = tag[:len(tag)-1] + "d" // %i -> %d
		return fmt.Sprintf(tag, luaToInteger(ls, argIdx))
	case 'd', 'o': // integer, octal
		return fmt.Sprintf(tag, luaToInteger(ls, argIdx))
	case 'u': // unsigned integer
		tag = tag[:len(tag)-1] + "d" // %u -> %d
		return fmt.Sprintf(tag, uint(luaToInteger(ls, argIdx)))
	case 'x', 'X': // hex integer
		return fmt.Sprintf(tag, uint(luaToInteger(ls, argIdx)))
	case 'f': // float
		return fmt.Sprintf(tag, luaToNumber(ls, argIdx))
	case 's', 'q': // string
		return fmt.Sprintf(tag, luaToString2(ls, argIdx))
	default:
		panic("todo! tag=" + tag)
	}
}

/* PATTERN MATCHING */

// string.find (s, pattern [, init [, plain]])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.find
func strFind(ls *LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	pattern := ls.CheckString(2)
	init := posRelat(luaOptInteger(ls, 3, 1), sLen)
	if init < 1 {
		init = 1
	} else if init > sLen+1 { /* start after string's end? */
		ls.Push(LuaNil)
		return 1
	}
	plain := luaToBoolean(ls, 4)

	start, end := find(s, pattern, init, plain)

	if start < 0 {
		ls.Push(LuaNil)
		return 1
	}
	ls.Push(LuaNumber(start))
	ls.Push(LuaNumber(end))
	return 2
}

// string.match (s, pattern [, init])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.match
func strMatch(ls *LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	pattern := ls.CheckString(2)
	init := posRelat(luaOptInteger(ls, 3, 1), sLen)
	if init < 1 {
		init = 1
	} else if init > sLen+1 { /* start after string's end? */
		ls.Push(LuaNil)
		return 1
	}

	captures := match(s, pattern, init)

	if captures == nil {
		ls.Push(LuaNil)
		return 1
	} else {
		for i := 0; i < len(captures); i += 2 {
			capture := s[captures[i]:captures[i+1]]
			ls.Push(LuaString(capture))
		}
		return len(captures) / 2
	}
}

// string.gsub (s, pattern, repl [, n])
// http://www.lua.org/manual/5.3/manual.html#pdf-string.gsub
func strGsub(ls *LuaState) int {
	str := ls.CheckString(1)
	pattern := ls.CheckString(2)
	luaCheckTypes(ls, 3, LUA_TSTRING, LUA_TTABLE, LUA_TCLOSURE)
	repl := ls.CheckAny(3) // todo
	limit := int(luaOptInteger(ls, 4, -1))

	mds, err := Find(pattern, unsafeFastStringToReadOnlyBytes(str), 0, limit)
	if err != nil {
		ls.Error2(err.Error())
	}
	if len(mds) == 0 {
		luaSetTop(ls, 1)
		ls.Push(LuaNumber(0))
		return 2
	}

	switch lv := repl.(type) {
	case LuaString:
		ls.Push(LuaString(strGsubStr(ls, str, lv.String(), mds)))
	case *LuaTable:
		ls.Push(LuaString(strGsubTable(ls, str, lv, mds)))
	case *LuaClosure:
		ret := strGsubClosure(ls, str, lv, mds)
		ls.Push(LuaString(ret))
	default:
		fmt.Printf("gsub default:%v lv:%+v\n", repl.String(), lv)
		ls.Push(LuaString(""))
	}
	ls.Push(LuaNumber(len(mds)))
	return 2
}

type replaceInfo struct {
	Indicies []int
	String   string
}

func checkCaptureIndex(ls *LuaState, m *MatchData, idx int) {
	if idx <= 2 {
		return
	}
	if idx >= m.CaptureLength() {
		ls.Error2("invalid capture index")
	}
}

func capturedString(ls *LuaState, m *MatchData, str string, idx int) string {
	checkCaptureIndex(ls, m, idx)
	if idx >= m.CaptureLength() && idx == 2 {
		idx = 0
	}
	if m.IsPosCapture(idx) {
		return fmt.Sprint(m.Capture(idx))
	} else {
		return str[m.Capture(idx):m.Capture(idx+1)]
	}
}

func strGsubDoReplace(str string, info []replaceInfo) string {
	offset := 0
	buf := []byte(str)
	for _, replace := range info {
		oldlen := len(buf)
		b1 := append([]byte(""), buf[0:offset+replace.Indicies[0]]...)
		b2 := []byte("")
		index2 := offset + replace.Indicies[1]
		if index2 <= len(buf) {
			b2 = append(b2, buf[index2:len(buf)]...)
		}
		buf = append(b1, replace.String...)
		buf = append(buf, b2...)
		offset += len(buf) - oldlen
	}
	return string(buf)
}

func strGsubStr(ls *LuaState, str string, repl string, matches []*MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		start, end := match.Capture(0), match.Capture(1)
		sc := newFlagScanner('%', "", "", repl)
		for c, eos := sc.Next(); !eos; c, eos = sc.Next() {
			if !sc.ChangeFlag {
				if sc.HasFlag {
					if c >= '0' && c <= '9' {
						sc.AppendString(capturedString(ls, match, str, 2*(int(c)-48)))
					} else {
						sc.AppendChar('%')
						sc.AppendChar(c)
					}
					sc.HasFlag = false
				} else {
					sc.AppendChar(c)
				}
			}
		}
		infoList = append(infoList, replaceInfo{[]int{start, end}, sc.String()})
	}

	return strGsubDoReplace(str, infoList)
}

func strGsubTable(ls *LuaState, str string, repl *LuaTable, matches []*MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		idx := 0
		if match.CaptureLength() > 2 { // has captures
			idx = 2
		}
		var value LuaValue
		if match.IsPosCapture(idx) {
			value = GetValueField(ls, repl, LuaString(match.Capture(idx)))
		} else {
			value = GetValueField(ls, repl, LuaString(str[match.Capture(idx):match.Capture(idx+1)]))
		}
		switch sv := value.(type) {
		case LuaString:
			infoList = append(infoList, replaceInfo{[]int{match.Capture(0), match.Capture(1)}, string(sv)})
		}
	}
	return strGsubDoReplace(str, infoList)
}

func strGsubClosure(ls *LuaState, str string, repl *LuaClosure, matches []*MatchData) string {
	infoList := make([]replaceInfo, 0, len(matches))
	for _, match := range matches {
		start, end := match.Capture(0), match.Capture(1)
		ls.Push(repl)
		nargs := 0
		if match.CaptureLength() > 2 { // has captures
			for i := 2; i < match.CaptureLength(); i += 2 {
				if match.IsPosCapture(i) {
					ls.Push(LuaString(match.Capture(i)))
				} else {
					ls.Push(LuaString(capturedString(ls, match, str, i)))
				}
				nargs++
			}
		} else {
			ls.Push(LuaString(capturedString(ls, match, str, 0)))
			nargs++
		}
		ls.Call(nargs, 1)
		ret := ls.stack.pop()
		switch sret := ret.(type) {
		case LuaString:
			infoList = append(infoList, replaceInfo{[]int{start, end}, string(sret)})
		}
	}
	return strGsubDoReplace(str, infoList)
}

type strMatchData struct {
	str     string
	pos     int
	matches []*MatchData
}

// string.gmatch (s, pattern)
// http://www.lua.org/manual/5.3/manual.html#pdf-string.gmatch
func strGmatch(ls *LuaState) int {
	s := ls.CheckString(1)
	pattern := ls.CheckString(2)

	gmatchAux := func(ls *LuaState) int {
		captures := match(s, pattern, 1)
		if captures != nil {
			for i := 0; i < len(captures); i += 2 {
				capture := s[captures[i]:captures[i+1]]
				ls.Push(LuaString(capture))
			}
			s = s[captures[len(captures)-1]:]
			return len(captures) / 2
		} else {
			return 0
		}
	}

	ls.PushGoFunction(gmatchAux)
	return 1
}

/* helper */

/* translate a relative string position: negative means back from end */
func posRelat(pos int64, _len int) int {
	_pos := int(pos)
	if _pos >= 0 {
		return _pos
	} else if -_pos > _len {
		return 0
	} else {
		return _len + _pos + 1
	}
}

// tag = %[flags][width][.precision]specifier
var tagPattern = regexp.MustCompile(`%[ #+-0]?[0-9]*(\.[0-9]+)?[cdeEfgGioqsuxX%]`)

func parseFmtStr(fmt string) []string {
	if fmt == "" || strings.IndexByte(fmt, '%') < 0 {
		return []string{fmt}
	}

	parsed := make([]string, 0, len(fmt)/2)
	for {
		if fmt == "" {
			break
		}

		loc := tagPattern.FindStringIndex(fmt)
		if loc == nil {
			parsed = append(parsed, fmt)
			break
		}

		head := fmt[:loc[0]]
		tag := fmt[loc[0]:loc[1]]
		tail := fmt[loc[1]:]

		if head != "" {
			parsed = append(parsed, head)
		}
		parsed = append(parsed, tag)
		fmt = tail
	}
	return parsed
}

func find(s, pattern string, init int, plain bool) (start, end int) {
	tail := s
	if init > 1 {
		tail = s[init-1:]
	}

	if plain {
		start = strings.Index(tail, pattern)
		end = start + len(pattern) - 1
	} else {
		re, err := _compile(pattern)
		if err != "" {
			panic(err) // todo
		} else {
			loc := re.FindStringIndex(tail)
			if loc == nil {
				start, end = -1, -1
			} else {
				start, end = loc[0], loc[1]-1
			}
		}
	}
	if start >= 0 {
		start += len(s) - len(tail) + 1
		end += len(s) - len(tail) + 1
	}

	return
}

func match(s, pattern string, init int) []int {
	tail := s
	if init > 1 {
		tail = s[init-1:]
	}

	re, err := _compile(pattern)
	if err != "" {
		panic(err) // todo
	} else {
		found := re.FindStringSubmatchIndex(tail)
		if len(found) > 2 {
			return found[2:]
		} else {
			return found
		}
	}
}

//func strGsubStr(L *LState, str string, repl string, matches []*pm.MatchData) string {
//	infoList := make([]replaceInfo, 0, len(matches))
//	for _, match := range matches {
//		start, end := match.Capture(0), match.Capture(1)
//		sc := newFlagScanner('%', "", "", repl)
//		for c, eos := sc.Next(); !eos; c, eos = sc.Next() {
//			if !sc.ChangeFlag {
//				if sc.HasFlag {
//					if c >= '0' && c <= '9' {
//						sc.AppendString(capturedString(L, match, str, 2*(int(c)-48)))
//					} else {
//						sc.AppendChar('%')
//						sc.AppendChar(c)
//					}
//					sc.HasFlag = false
//				} else {
//					sc.AppendChar(c)
//				}
//			}
//		}
//		infoList = append(infoList, replaceInfo{[]int{start, end}, sc.String()})
//	}
//
//	return strGsubDoReplace(str, infoList)
//}
//
//func strGsubTable(L *LState, str string, repl *LTable, matches []*pm.MatchData) string {
//	infoList := make([]replaceInfo, 0, len(matches))
//	for _, match := range matches {
//		idx := 0
//		if match.CaptureLength() > 2 { // has captures
//			idx = 2
//		}
//		var value LValue
//		if match.IsPosCapture(idx) {
//			value = L.GetTable(repl, LNumber(match.Capture(idx)))
//		} else {
//			value = L.GetField(repl, str[match.Capture(idx):match.Capture(idx+1)])
//		}
//		if !LVIsFalse(value) {
//			infoList = append(infoList, replaceInfo{[]int{match.Capture(0), match.Capture(1)}, LVAsString(value)})
//		}
//	}
//	return strGsubDoReplace(str, infoList)
//}
//
//func strGsubFunc(L *LuaState, str string, repl *LuaFunction, matches []*pm.MatchData) string {
//	infoList := make([]replaceInfo, 0, len(matches))
//	for _, match := range matches {
//		start, end := match.Capture(0), match.Capture(1)
//		L.Push(repl)
//		nargs := 0
//		if match.CaptureLength() > 2 { // has captures
//			for i := 2; i < match.CaptureLength(); i += 2 {
//				if match.IsPosCapture(i) {
//					L.Push(LNumber(match.Capture(i)))
//				} else {
//					L.Push(LString(capturedString(L, match, str, i)))
//				}
//				nargs++
//			}
//		} else {
//			L.Push(LString(capturedString(L, match, str, 0)))
//			nargs++
//		}
//		L.Call(nargs, 1)
//		ret := L.reg.Pop()
//		if !LVIsFalse(ret) {
//			infoList = append(infoList, replaceInfo{[]int{start, end}, LVAsString(ret)})
//		}
//	}
//	return strGsubDoReplace(str, infoList)
//}

func _compile(pattern string) (*regexp.Regexp, string) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err.Error() // todo
	} else {
		return re, ""
	}
}