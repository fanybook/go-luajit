package luajit

/*
#cgo pkg-config: luajit
#include <lua.h>
#include <luajit.h>
#include <lualib.h>
#include <lauxlib.h>
#include <luaconf.h>
#include <stdlib.h>
*/
import "C"

import (
	"sync"
	"unsafe"
)

const (
	LUA_OK        = C.LUA_OK
	LUA_YIELD     = C.LUA_YIELD
	LUA_ERRRUN    = C.LUA_ERRRUN
	LUA_ERRSYNTAX = C.LUA_ERRSYNTAX
	LUA_ERRMEM    = C.LUA_ERRMEM
	LUA_ERRERR    = C.LUA_ERRERR
)

const (
	LUA_TNONE = C.LUA_TNONE

	LUA_TNIL           = C.LUA_TNIL
	LUA_TBOOLEAN       = C.LUA_TBOOLEAN
	LUA_TLIGHTUSERDATA = C.LUA_TLIGHTUSERDATA
	LUA_TNUMBER        = C.LUA_TNUMBER
	LUA_TSTRING        = C.LUA_TSTRING
	LUA_TTABLE         = C.LUA_TTABLE
	LUA_TFUNCTION      = C.LUA_TFUNCTION
	LUA_TUSERDATA      = C.LUA_TUSERDATA
	LUA_TTHREAD        = C.LUA_TTHREAD
)

const (
	LUA_REGISTRYINDEX = C.LUA_REGISTRYINDEX
	LUA_ENVIRONINDEX  = C.LUA_ENVIRONINDEX
	LUA_GLOBALSINDEX  = C.LUA_GLOBALSINDEX
)

const (
	LUA_MULTRET = C.LUA_MULTRET
)

type LuaState interface {
	GoSetExData(interface{})
	GoGetExData() (interface{}, bool)
	GoDeleteExData()

	/*
		state manipulation
	*/
	Close()
	NewThread() LuaState

	//lua_resetthread
	//lua_atpanic

	/*
		basic stack manipulation
	*/
	GetTop() int
	SetTop(int)
	PushValue(int)
	Remove(int)
	//insert
	Replace(int)
	//checkstack

	//xmove

	/*
		access functions (stack -> C)
	*/
	IsNumber(int) bool
	IsString(int) bool
	//iscfunction
	Type(int) int
	TypeName(int) string

	//equal
	//rawequal
	//lessthan

	ToNumber(int) float64
	ToInteger(int) int64
	ToBoolean(int) bool
	//pushlstring
	//objlen
	//tocfunction
	ToUserData(int) unsafe.Pointer
	//tothread
	//topointer

	/*
		push functions (C -> stack)
	*/
	PushNil()
	PushNumber(float64)
	PushInteger(int64)
	PushString(string)
	PushCClosure(unsafe.Pointer, int)
	//pushboolean
	PushLightUserData(unsafe.Pointer)
	//pushthread

	/*
		get functions (Lua -> stack)
	*/
	GetTable(int)
	GetField(int, string)
	RawGet(int)
	//rawgeti
	CreateTable(int, int)
	//newuserdata
	//getmetatable
	//getfenv

	/*
		set functions (stack -> Lua)
	*/
	SetTable(int)
	SetField(int, string)
	RawSet(int)
	//rawseti
	SetMetaTable(int)
	//setfenv

	/*
		`load' and `call' functions (load and run Lua code)
	*/
	Call(int, int)
	PCall(int, int, int) int
	//cpcall
	//load
	//dump

	/*
		coroutine functions
	*/
	//yield
	//resume
	//status

	/*
		garbage-collection function and options
	*/
	//gc

	/*
		miscellaneous functions
	*/
	//error
	Next(int) bool
	//concat
	//getallocf
	//setallocf
	SetExData(unsafe.Pointer)
	GetExData() unsafe.Pointer
	//luasetexdata2
	//getexdata2

	/*
		some useful macros
	*/
	Pop(int)
	NewTable()
	Register(string, unsafe.Pointer)
	PushCFunction(unsafe.Pointer)
	//strlen
	IsFunction(int) bool
	IsTable(int) bool
	//islightuserdata
	//isnil
	//isboolean
	//isthread
	//isnone
	//isnoneornil
	//pushliteral
	GetGlobal(string)
	SetGlobal(string)
	ToString(int) string

	/*
		compatibility macros and functions
	*/
	//open()
	//getregistry
	//getgccount

	/* hack */
	//setlevel

	//lauxlib.h

	//l_openlib
	//l_register
	//l_getmetafield
	//l_callmeta
	//l_typerror
	//l_argerror
	//l_checklstring
	//l_optlstirng
	//l_optnumber
	//l_checkinteger
	//l_optinteger
	//l_checkstack
	//l_checktype
	//l_checkany
	LNewMetaTable(string)
	//l_checkudata
	//l_where
	//l_error
	//l_checkoption

	/* pre-defined references */
	//l_ref
	//l_unref

	LLoadFile(string) int
	LLoadBuffer([]byte, string) int
	LLoadString(string) int

	//l_newstate
	//l_gsub
	//l_findtable

	/* From Lua 5.2. */
	//l_fileresult
	//l_execresult
	//l_loadfilex
	//l_loadbufferx
	//l_traceback
	//l_setfuncs
	//l_pushmodule
	//l_testudata
	//l_setmetateble

	/*
		some useful macros
	*/

	//l_argcheck
	//l_checkstrig
	//l_optstring
	//l_optint
	//l_chekclong
	//l_optlong
	//l_typename
	LDoFile(string) int
	LDoString(string) int
	LGetMetaTable(string)
	//l_opt
	//l_newlibtable
	//l_newlib

	LOpenLibs()
}

type CFunction func(unsafe.Pointer) C.int

type luaState struct {
	l *C.struct_lua_State
}

func UpValueIndex(idx int) int {
	return LUA_GLOBALSINDEX - idx
}

func NewState() LuaState {
	l := C.luaL_newstate()
	return &luaState{l}
}

func FromCLuaState(l unsafe.Pointer) LuaState {
	return &luaState{(*C.struct_lua_State)(l)}
}

var m sync.Map

func (l *luaState) GoSetExData(i interface{}) {
	m.Store(l.l, i)
}

func (l *luaState) GoGetExData() (interface{}, bool) {
	return m.Load(l.l)
}

func (l *luaState) GoDeleteExData() {
	m.Delete(l.l)
}

func (l *luaState) NewThread() LuaState {
	t := C.lua_newthread(l.l)

	return &luaState{l: t}
}

func (l *luaState) Close() {
	C.lua_close(l.l)
	l.l = nil
}

func (l *luaState) GetTop() int {
	return int(C.lua_gettop(l.l))
}

func (l *luaState) SetTop(idx int) {
	C.lua_settop(l.l, C.int(idx))
}

func (l *luaState) PushValue(idx int) {
	C.lua_pushvalue(l.l, C.int(idx))
}

func (l *luaState) Remove(idx int) {
	C.lua_remove(l.l, C.int(idx))
}

func (l *luaState) Replace(idx int) {
	C.lua_replace(l.l, C.int(idx))
}

func (l *luaState) LOpenLibs() {
	C.luaL_openlibs(l.l)
}

func (l *luaState) LNewMetaTable(tname string) {
	tnameCStr := C.CString(tname)
	defer C.free(unsafe.Pointer(tnameCStr))

	C.luaL_newmetatable(l.l, tnameCStr)
}

func (l *luaState) LLoadFile(filename string) int {
	fileNameCStr := C.CString(filename)
	defer C.free(unsafe.Pointer(fileNameCStr))

	return int(C.luaL_loadfile(l.l, fileNameCStr))
}

func (l *luaState) LLoadBuffer(buff []byte, name string) int {
	buffCChar := (*C.char)(C.CBytes(buff))
	nameCStr := C.CString(name)
	defer C.free(unsafe.Pointer(nameCStr))

	return int(C.luaL_loadbuffer(l.l, buffCChar, C.ulong(len(buff)), nameCStr))
}

func (l *luaState) LLoadString(s string) int {
	sCStr := C.CString(s)
	defer C.free(unsafe.Pointer(sCStr))

	return int(C.luaL_loadstring(l.l, sCStr))
}

func (l *luaState) Call(nargs, nresults int) {
	C.lua_call(l.l, C.int(nargs), C.int(nresults))
}

func (l *luaState) PCall(nargs, nresults, errfunc int) int {
	res := C.lua_pcall(l.l, C.int(nargs), C.int(nresults), C.int(errfunc))
	return int(res)
}

func (l *luaState) ToNumber(idx int) float64 {
	return float64(C.lua_tonumber(l.l, C.int(idx)))
}

func (l *luaState) ToBoolean(idx int) bool {
	return C.lua_toboolean(l.l, C.int(idx)) != 0
}

func (l *luaState) ToInteger(idx int) int64 {
	return int64(C.lua_tointeger(l.l, C.int(idx)))
}

func (l *luaState) ToString(idx int) string {
	var len C.size_t
	cStr := C.lua_tolstring(l.l, C.int(idx), &len)
	return C.GoStringN(cStr, C.int(len))
}

func (l *luaState) ToUserData(idx int) unsafe.Pointer {
	return C.lua_touserdata(l.l, C.int(idx))
}

func (l *luaState) GetTable(idx int) {
	C.lua_gettable(l.l, C.int(idx))
}

func (l *luaState) GetGlobal(k string) {
	l.GetField(LUA_GLOBALSINDEX, k)
}

func (l *luaState) SetGlobal(k string) {
	l.SetField(LUA_GLOBALSINDEX, k)
}

func (l *luaState) GetField(idx int, k string) {
	kCStr := C.CString(k)
	defer C.free(unsafe.Pointer(kCStr))

	C.lua_getfield(l.l, C.int(idx), kCStr)
}

func (l *luaState) RawGet(idx int) {
	C.lua_rawget(l.l, C.int(idx))
}

func (l *luaState) CreateTable(narr int, nrec int) {
	C.lua_createtable(l.l, C.int(narr), C.int(nrec))
}

func (l *luaState) SetTable(idx int) {
	C.lua_settable(l.l, C.int(idx))
}

func (l *luaState) SetField(idx int, k string) {
	kCStr := C.CString(k)
	defer C.free(unsafe.Pointer(kCStr))

	C.lua_setfield(l.l, C.int(idx), kCStr)
}

func (l *luaState) RawSet(idx int) {
	C.lua_rawset(l.l, C.int(idx))
}

func (l *luaState) SetMetaTable(idx int) {
	C.lua_setmetatable(l.l, C.int(idx))
}

func (l *luaState) LDoFile(filename string) int {
	if ret := l.LLoadFile(filename); ret != LUA_OK {
		return ret
	} else {
		if r := l.PCall(0, LUA_MULTRET, 0); r != LUA_OK {
			return r
		} else {
			return LUA_OK
		}
	}
}

func (l *luaState) LDoString(s string) int {
	if ret := l.LLoadString(s); ret != LUA_OK {
		return ret
	} else {
		if r := l.PCall(0, LUA_MULTRET, 0); r != LUA_OK {
			return r
		} else {
			return LUA_OK
		}
	}
}

func (l *luaState) LGetMetaTable(s string) {
	l.GetField(LUA_REGISTRYINDEX, s)
}

func (l *luaState) IsNumber(idx int) bool {
	return l.Type(idx) == LUA_TNUMBER
}

func (l *luaState) IsString(idx int) bool {
	return l.Type(idx) == LUA_TSTRING
}

func (l *luaState) Type(idx int) int {
	return int(C.lua_type(l.l, C.int(idx)))
}

func (l *luaState) TypeName(tp int) string {
	return C.GoString(C.lua_typename(l.l, C.int(tp)))
}

func (l *luaState) IsFunction(idx int) bool {
	return l.Type(idx) == LUA_TFUNCTION
}

func (l *luaState) IsTable(idx int) bool {
	return l.Type(idx) == LUA_TTABLE
}

func (l *luaState) PushNil() {
	C.lua_pushnil(l.l)
}

func (l *luaState) PushNumber(n float64) {
	C.lua_pushnumber(l.l, C.double(n))
}

func (l *luaState) PushInteger(n int64) {
	C.lua_pushinteger(l.l, C.long(n))
}

func (l *luaState) PushString(s string) {
	cStr := C.CString(s)
	defer C.free(unsafe.Pointer(cStr))

	C.lua_pushstring(l.l, cStr)
}

func (l *luaState) PushLightUserData(p unsafe.Pointer) {
	C.lua_pushlightuserdata(l.l, p)
}

func (l *luaState) Next(idx int) bool {
	return C.lua_next(l.l, C.int(idx)) != 0
}

func (l *luaState) SetExData(exdata unsafe.Pointer) {
	C.lua_setexdata(l.l, exdata)
}

func (l *luaState) GetExData() unsafe.Pointer {
	return C.lua_getexdata(l.l)
}

func (l *luaState) Pop(idx int) {
	l.SetTop(-idx - 1)
}

func (l *luaState) NewTable() {
	l.CreateTable(0, 0)
}

func (l *luaState) Register(n string, fn unsafe.Pointer) {
	l.PushCFunction(fn)
	l.SetGlobal(n)
}

func (l *luaState) PushCFunction(fn unsafe.Pointer) {
	l.PushCClosure(fn, 0)
}

func (l *luaState) PushCClosure(fn unsafe.Pointer, n int) {
	C.lua_pushcclosure(l.l, (C.lua_CFunction)(fn), C.int(n))
}
