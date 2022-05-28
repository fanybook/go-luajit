package main

import (
	"fmt"
	"unsafe"

	"github.com/xingheliufang/go-luajit/luajit"
)

/*
typedef struct lua_State lua_State;

extern int Move(lua_State*);
extern int Add(lua_State*);
extern int ToString(lua_State*);
extern int NewPointer(lua_State*);

extern int Count(lua_State*);
extern int NewCounter(lua_State*);
*/
import "C"

func Tables() {
	l := luajit.NewState()
	defer l.Close()

	const LUA_SCRIPT = `
		x = { dave = "busy", ian = "idle" }
	`

	l.LDoString(LUA_SCRIPT)

	l.GetGlobal("x")
	if l.IsTable(-1) {
		fmt.Println("is table!")
	}

	l.PushString("dave")
	l.GetTable(-2)
	fmt.Println(l.ToString(-1))

	l.GetGlobal("x")
	l.GetField(-1, "ian")
	fmt.Println(l.ToString(-1))

	l.GetGlobal("x")
	l.PushString("sleeping")
	l.SetField(-2, "john")

	l.GetGlobal("x")
	l.GetField(-1, "john")
	fmt.Println(l.ToString(-1))

}

type Pointer struct {
	x int64
	y int64
}

func (p *Pointer) Move(x, y int64) {
	p.x += x
	p.y += y
}

func (p Pointer) Add(o Pointer) Pointer {
	return Pointer{
		x: p.x + o.x,
		y: p.y + o.y,
	}
}

func (p Pointer) ToString() string {
	return fmt.Sprintf("x = %d, y = %d", p.x, p.y)
}

//export Move
func Move(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p := (*Pointer)(l.ToUserData(1))
	x := l.ToInteger(2)
	y := l.ToInteger(3)

	p.Move(x, y)

	return 0
}

//export NewPointer
func NewPointer(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	x := l.ToInteger(1)
	y := l.ToInteger(2)

	l.PushLightUserData(unsafe.Pointer(&Pointer{x, y}))
	l.LGetMetaTable("PointerMetaTable")
	l.SetMetaTable(-2)

	return 1
}

//export Add
func Add(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p1 := (*Pointer)(l.ToUserData(1))
	p2 := (*Pointer)(l.ToUserData(2))

	p3 := p1.Add(*p2)

	l.PushLightUserData(unsafe.Pointer(&p3))

	return 1
}

//export ToString
func ToString(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p := (*Pointer)(l.ToUserData(1))

	l.PushString(p.ToString())

	return 1
}

//export Count
func Count(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	c := l.ToInteger(luajit.UpValueIndex(1))
	l.PushInteger(c + 1)
	l.PushValue(-1)

	l.Replace(luajit.UpValueIndex(1))

	return 1
}

//export NewCounter
func NewCounter(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	l.PushInteger(0)

	l.PushCClosure(unsafe.Pointer(C.Count), 1)
	return 1
}

func MetaTable() {
	l := luajit.NewState()
	defer l.Close()

	l.LOpenLibs()

	const LUA_SCRIPT = `
		x = Pointer.new(1, 2)
		y = Pointer.new(3, 4)
		z = x + y
		z:move(1, 1)
		print(z)
	`

	l.NewTable()
	l.PushCFunction(unsafe.Pointer(C.NewPointer))
	l.SetField(-2, "new")
	l.SetGlobal("Pointer")

	l.LNewMetaTable("PointerMetaTable")

	l.PushCFunction(unsafe.Pointer(C.Add))
	l.SetField(-2, "__add")
	l.PushCFunction(unsafe.Pointer(C.ToString))
	l.SetField(-2, "__tostring")

	l.PushValue(-1)
	l.SetField(-2, "__index")

	l.PushCFunction(unsafe.Pointer(C.Move))
	l.SetField(-2, "move")

	if l.LDoString(LUA_SCRIPT) != luajit.LUA_OK {
		fmt.Println(l.ToString(-1))
	}
}

func Closure() {
	l := luajit.NewState()
	defer l.Close()

	l.LOpenLibs()

	const LUA_SCRIPT = `
	c1 = NewCounter()
	c2 = NewCounter()

	print("c1: " .. c1())
	print("c1: " .. c1())
	print("c1: " .. c1())
	print("c1: " .. c1())

	print("c2: " .. c2())
`

	l.Register("NewCounter", C.NewCounter)

	if l.LDoString(LUA_SCRIPT) != luajit.LUA_OK {
		fmt.Println(l.ToString(-1))
	}
}

func main() {
	Tables()

	MetaTable()

	Closure()
}
