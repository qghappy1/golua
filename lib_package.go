package golua

import (
	"os"
	"strings"
)

/* key, in the registry, for table of loaded modules */
const LUA_LOADED_TABLE = "_LOADED"

/* key, in the registry, for table of preloaded loaders */
const LUA_PRELOAD_TABLE = "_PRELOAD"

const (
	LUA_DIRSEP    = string(os.PathSeparator)
	LUA_PATH_SEP  = ";"
	LUA_PATH_MARK = "?"
	LUA_EXEC_DIR  = "!"
	LUA_IGMARK    = "-"
)

var pkgFuncs = map[string]GoFunction{
	"searchpath": pkgSearchPath,
	/* placeholders */
	"preload":   nil,
	"cpath":     nil,
	"path":      nil,
	"searchers": nil,
	"loaded":    nil,
}

var llFuncs = map[string]GoFunction{
	"require": pkgRequire,
}

func OpenPackageLib(ls *LuaState) int {
	ls.NewLib(pkgFuncs) /* create 'package' table */
	createSearchersTable(ls)
	/* set paths */
	ls.Push(LuaString("./?.lua;./?/init.lua"))
	luaSetField(ls, -2, "path")
	/* store config information */
	ls.Push(LuaString(LUA_DIRSEP + "\n" + LUA_PATH_SEP + "\n" +
		LUA_PATH_MARK + "\n" + LUA_EXEC_DIR + "\n" + LUA_IGMARK + "\n"))
	luaSetField(ls, -2, "config")
	/* set field 'loaded' */
	luaGetSubTable(ls, LUA_REGISTRYINDEX, LUA_LOADED_TABLE)
	luaSetField(ls, -2, "loaded")
	/* set field 'preload' */
	luaGetSubTable(ls, LUA_REGISTRYINDEX, LUA_PRELOAD_TABLE)
	luaSetField(ls, -2, "preload")
	luaPushGlobalTable(ls)
	luaPushValue(ls, -2)        /* set 'package' as upvalue for next lib */
	ls.SetFuncs(llFuncs, 1) /* open lib into global table */
	luaPop(ls, 1)               /* pop global table */
	return 1                /* return 'package' table */
}

func createSearchersTable(ls *LuaState) {
	searchers := []GoFunction{
		preloadSearcher,
		luaSearcher,
	}
	/* create 'searchers' table */
	ls.createTable(len(searchers), 0)
	/* fill it with predefined searchers */
	for idx, searcher := range searchers {
		luaPushValue(ls, -2) /* set 'package' as upvalue for all searchers */
		ls.PushGoClosure(searcher, 1)
		luaRawSetI(ls, -2, int64(idx+1))
	}
	luaSetField(ls, -2, "searchers") /* put it in field 'searchers' */
}

func preloadSearcher(ls *LuaState) int {
	name := ls.CheckString(1)
	luaGetField(ls, LUA_REGISTRYINDEX, "_PRELOAD")
	if luaGetField(ls, -1, name) == LUA_TNIL { /* not found? */
		ls.Push(LuaString("\n\tno field package.preload['" + name + "']"))
	}
	return 1
}

func luaSearcher(ls *LuaState) int {
	name := ls.CheckString(1)
	luaGetField(ls, luaUpvalueIndex(1), "path")
	path, ok := luaToStringX(ls, -1)
	if !ok {
		ls.Error2("'package.path' must be a string")
	}

	filename, errMsg := _searchPath(name, path, ".", LUA_DIRSEP)
	if errMsg != "" {
		ls.Push(LuaString(errMsg))
		return 1
	}

	if ls.LoadFile(filename) == LUA_OK { /* module loaded successfully? */
		ls.Push(LuaString(filename)) /* will be 2nd argument to module */
		return 2                /* return open function and file name */
	} else {
		return ls.Error2("error loading module '%s' from file '%s':\n\t%s",
			ls.CheckString(1), filename, ls.CheckString(-1))
	}
}

// package.searchpath (name, path [, sep [, rep]])
// http://www.lua.org/manual/5.3/manual.html#pdf-package.searchpath
// loadlib.c#ll_searchpath
func pkgSearchPath(ls *LuaState) int {
	name := ls.CheckString(1)
	path := ls.CheckString(2)
	sep := luaOptString(ls, 3, ".")
	rep := luaOptString(ls, 4, LUA_DIRSEP)
	if filename, errMsg := _searchPath(name, path, sep, rep); errMsg == "" {
		ls.Push(LuaString(filename))
		return 1
	} else {
		ls.Push(LuaNil)
		ls.Push(LuaString(errMsg))
		return 2
	}
}

func _searchPath(name, path, sep, dirSep string) (filename, errMsg string) {
	if sep != "" {
		name = strings.Replace(name, sep, dirSep, -1)
	}

	for _, filename := range strings.Split(path, LUA_PATH_SEP) {
		filename = strings.Replace(filename, LUA_PATH_MARK, name, -1)
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			return filename, ""
		}
		errMsg += "\n\tno file '" + filename + "'"
	}

	return "", errMsg
}

// require (modname)
// http://www.lua.org/manual/5.3/manual.html#pdf-require
func pkgRequire(ls *LuaState) int {
	name := ls.CheckString(1)
	luaSetTop(ls, 1) /* LOADED table will be at index 2 */
	luaGetField(ls, LUA_REGISTRYINDEX, LUA_LOADED_TABLE)
	luaGetField(ls, 2, name)  /* LOADED[name] */
	if luaToBoolean(ls, -1) { /* is it there? */
		return 1 /* package is already loaded */
	}
	/* else must load package */
	luaPop(ls, 1) /* remove 'getfield' result */
	_findLoader(ls, name)
	ls.Push(LuaString(name)) /* pass name as argument to module loader */
	luaInsert(ls, -2)       /* name is 1st argument (before search data) */
	ls.Call(2, 1)       /* run loader to load module */
	if !luaIsNil(ls, -1) {  /* non-nil return? */
		luaSetField(ls, 2, name) /* LOADED[name] = returned value */
	}
	if luaGetField(ls, 2, name) == LUA_TNIL { /* module set no value? */
		ls.Push(LuaTrue) /* use true as result */
		luaPushValue(ls, -1)     /* extra copy to be returned */
		luaSetField(ls, 2, name) /* LOADED[name] = true */
	}
	return 1
}

func _findLoader(ls *LuaState, name string) {
	/* push 'package.searchers' to index 3 in the stack */
	if luaGetField(ls, luaUpvalueIndex(1), "searchers") != LUA_TTABLE {
		ls.Error2("'package.searchers' must be a table")
	}

	/* to build error message */
	errMsg := "module '" + name + "' not found:"

	/*  iterate over available searchers to find a loader */
	for i := int64(1); ; i++ {
		if luaRawGetI(ls, 3, i) == LUA_TNIL { /* no more searchers? */
			luaPop(ls, 1)         /* remove nil */
			ls.Error2(errMsg) /* create error message */
		}

		ls.Push(LuaString(name))
		ls.Call(1, 2)          /* call it */
		if luaIsFunction(ls, -2) { /* did it find a loader? */
			return /* module loader found */
		} else if luaIsString(ls, -2) { /* searcher returned error message? */
			luaPop(ls, 1)                    /* remove extra return */
			errMsg += ls.CheckString(-1) /* concatenate error message */
		} else {
			luaPop(ls, 2) /* remove both returns */
		}
	}
}
