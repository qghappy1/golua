package golua

import (
	"os"
	"time"
)

var sysLib = map[string]GoFunction{
	"clock":     osClock,
	"difftime":  osDiffTime,
	"time":      osTime,
	"date":      osDate,
	"remove":    osRemove,
	"rename":    osRename,
	"tmpname":   osTmpName,
	"getenv":    osGetEnv,
	"execute":   osExecute,
	"exit":      osExit,
	"setlocale": osSetLocale,
}

func OpenOSLib(ls *LuaState) int {
	ls.NewLib(sysLib)
	return 1
}

// os.clock ()
// http://www.lua.org/manual/5.3/manual.html#pdf-os.clock
// lua-5.3.4/src/loslib.c#os_clock()
func osClock(ls *LuaState) int {
	// c := float64(C.clock()) / float64(C.CLOCKS_PER_SEC)
	c := float64(time.Now().Unix())
	ls.Push(LuaNumber(c))
	return 1
}

// os.difftime (t2, t1)
// http://www.lua.org/manual/5.3/manual.html#pdf-os.difftime
// lua-5.3.4/src/loslib.c#os_difftime()
func osDiffTime(ls *LuaState) int {
	t2 := ls.CheckInteger(1)
	t1 := ls.CheckInteger(2)
	ls.Push(LuaNumber(t2 - t1))
	return 1
}

// os.time ([table])
// http://www.lua.org/manual/5.3/manual.html#pdf-os.time
// lua-5.3.4/src/loslib.c#os_time()
func osTime(ls *LuaState) int {
	if luaIsNoneOrNil(ls, 1) { /* called without args? */
		t := time.Now().Unix() /* get current time */
		ls.Push(LuaNumber(t))
	} else {
		luaCheckType(ls, 1, LUA_TTABLE)
		sec := _getField(ls, "sec", 0)
		min := _getField(ls, "min", 0)
		hour := _getField(ls, "hour", 12)
		day := _getField(ls, "day", -1)
		month := _getField(ls, "month", -1)
		year := _getField(ls, "year", -1)
		// todo: isdst
		t := time.Date(year, time.Month(month), day,
			hour, min, sec, 0, time.Local).Unix()
		ls.Push(LuaNumber(t))
	}
	return 1
}

// lua-5.3.4/src/loslib.c#getfield()
func _getField(ls *LuaState, key string, dft int64) int {
	t := luaGetField(ls, -1, key) /* get field and its type */
	res, isNum := luaToIntegerX(ls, -1)
	if !isNum { /* field is not an integer? */
		if t != LUA_TNIL { /* some other value? */
			return ls.Error2("field '%s' is not an integer", key)
		} else if dft < 0 { /* absent field; no default? */
			return ls.Error2("field '%s' missing in date table", key)
		}
		res = dft
	}
	luaPop(ls, 1)
	return int(res)
}

// os.date ([format [, time]])
// http://www.lua.org/manual/5.3/manual.html#pdf-os.date
// lua-5.3.4/src/loslib.c#os_date()
func osDate(ls *LuaState) int {
	format := luaOptString(ls, 1, "%c")
	var t time.Time
	if luaIsInteger(ls, 2) {
		t = time.Unix(luaToInteger(ls, 2), 0)
	} else {
		t = time.Now()
	}

	if format != "" && format[0] == '!' { /* UTC? */
		format = format[1:] /* skip '!' */
		t = t.In(time.UTC)
	}

	if format == "*t" {
		ls.createTable(0, 9) /* 9 = number of fields */
		_setField(ls, "sec", t.Second())
		_setField(ls, "min", t.Minute())
		_setField(ls, "hour", t.Hour())
		_setField(ls, "day", t.Day())
		_setField(ls, "month", int(t.Month()))
		_setField(ls, "year", t.Year())
		_setField(ls, "wday", int(t.Weekday())+1)
		_setField(ls, "yday", t.YearDay())
	} else if format == "%c" {
		ls.Push(LuaString(t.Format(time.ANSIC)))
	} else {
		ls.Push(LuaString(format)) // TODO
	}

	return 1
}

func _setField(ls *LuaState, key string, value int) {
	ls.Push(LuaNumber(value))
	luaSetField(ls, -2, key)
}

// os.remove (filename)
// http://www.lua.org/manual/5.3/manual.html#pdf-os.remove
func osRemove(ls *LuaState) int {
	filename := ls.CheckString(1)
	if err := os.Remove(filename); err != nil {
		ls.Push(LuaNil)
		ls.Push(LuaString(err.Error()))
		return 2
	} else {
		ls.Push(LuaTrue)
		return 1
	}
}

// os.rename (oldname, newname)
// http://www.lua.org/manual/5.3/manual.html#pdf-os.rename
func osRename(ls *LuaState) int {
	oldName := ls.CheckString(1)
	newName := ls.CheckString(2)
	if err := os.Rename(oldName, newName); err != nil {
		ls.Push(LuaNil)
		ls.Push(LuaString(err.Error()))
		return 2
	} else {
		ls.Push(LuaTrue)
		return 1
	}
}

// os.tmpname ()
// http://www.lua.org/manual/5.3/manual.html#pdf-os.tmpname
func osTmpName(ls *LuaState) int {
	panic("todo: osTmpName!")
}

// os.getenv (varname)
// http://www.lua.org/manual/5.3/manual.html#pdf-os.getenv
// lua-5.3.4/src/loslib.c#os_getenv()
func osGetEnv(ls *LuaState) int {
	key := ls.CheckString(1)
	if env := os.Getenv(key); env != "" {
		ls.Push(LuaString(env))
	} else {
		ls.Push(LuaNil)
	}
	return 1
}

// os.execute ([command])
// http://www.lua.org/manual/5.3/manual.html#pdf-os.execute
func osExecute(ls *LuaState) int {
	panic("todo: osExecute!")
}

// os.exit ([code [, close]])
// http://www.lua.org/manual/5.3/manual.html#pdf-os.exit
// lua-5.3.4/src/loslib.c#os_exit()
func osExit(ls *LuaState) int {
	if luaIsBoolean(ls, 1) {
		if luaToBoolean(ls, 1) {
			os.Exit(0)
		} else {
			os.Exit(1) // todo
		}
	} else {
		code := luaOptInteger(ls, 1, 1)
		os.Exit(int(code))
	}
	if luaToBoolean(ls, 2) {
		//ls.Close()
	}
	return 0
}

// os.setlocale (locale [, category])
// http://www.lua.org/manual/5.3/manual.html#pdf-os.setlocale
func osSetLocale(ls *LuaState) int {
	panic("todo: osSetLocale!")
}
