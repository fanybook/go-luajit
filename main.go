package main

/*
#cgo pkg-config: luajit
#include <luajit.h>
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luaconf.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const luaScript = `
local ffi = require "ffi"
ffi.cdef[[
	extern void print();
]]

print("hello from lua")
ffi.C.print()
`

//export print
func print() {
	fmt.Println("hello from go")
}

func main() {
	L := C.luaL_newstate()
	defer C.lua_close(L)

	C.luaL_openlibs(L)

	luaScriptCStr := C.CString(luaScript)
	defer C.free(unsafe.Pointer(luaScriptCStr))

	luaScriptName := C.CString("print-test")
	defer C.free(unsafe.Pointer(luaScriptName))

	C.luaL_loadbuffer(L, luaScriptCStr, C.ulong(len(luaScript)), luaScriptName)

	result := C.lua_pcall(L, 0, -1, 0)
	if result != 0 {
		panic("lua error")
	}
}
