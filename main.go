package main

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/xingheliufang/go-luajit/luajit"
)

/*
typedef struct lua_State lua_State;

extern int Index(lua_State*);
extern int NewIndex(lua_State*);
extern int Move(lua_State*);
extern int Add(lua_State*);
extern int ToString(lua_State*);
extern int NewPoint(lua_State*);

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

type Point struct {
	x int64
	y int64
}

func (p *Point) Move(x, y int64) {
	p.x += x
	p.y += y
}

func (p Point) Add(o Point) Point {
	return Point{
		x: p.x + o.x,
		y: p.y + o.y,
	}
}

func (p Point) ToString() string {
	return fmt.Sprintf("x = %d, y = %d", p.x, p.y)
}

//export NewPoint
func NewPoint(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	x := l.ToInteger(1)
	y := l.ToInteger(2)

	l.PushLightUserData(unsafe.Pointer(&Point{x, y}))
	l.LGetMetaTable("PointMetaTable")
	l.SetMetaTable(-2)

	return 1
}

//export Add
func Add(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p1 := (*Point)(l.ToUserData(1))
	p2 := (*Point)(l.ToUserData(2))

	p3 := p1.Add(*p2)

	l.PushLightUserData(unsafe.Pointer(&p3))

	return 1
}

//export ToString
func ToString(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p := (*Point)(l.ToUserData(1))

	l.PushString(p.ToString())

	return 1
}

//export Move
func Move(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p := (*Point)(l.ToUserData(1))
	x := l.ToInteger(2)
	y := l.ToInteger(3)

	p.Move(x, y)

	return 0
}

//export Index
func Index(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p := (*Point)(l.ToUserData(1))
	name := l.ToString(2)

	v := reflect.ValueOf(p)
	if field := v.Elem().FieldByName(name); field.IsValid() {
		switch field.Kind() {
		case reflect.Int64:
			l.PushInteger(field.Int())
		default:
			return 0
		}
	} else {
		l.GetGlobal("Point")
		l.PushValue(2)
		l.RawGet(-2)
	}

	return 1
}

//export NewIndex
func NewIndex(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	p := (*Point)(l.ToUserData(1))
	name := l.ToString(2)

	v := reflect.ValueOf(p)
	if field := v.Elem().FieldByName(name); field.IsValid() {
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		switch field.Kind() {
		case reflect.Int64:
			field.SetInt(l.ToInteger(3))
		default:
			return 0
		}
		return 1
	} else {
		return 0
	}
}

func MetaTable() {
	l := luajit.NewState()
	defer l.Close()

	l.LOpenLibs()

	const LUA_SCRIPT = `
		p1 = Point.new(1, 2)
		p2 = Point.new(3, 4)
		-- call __add metamethod
		p3 = p1 + p2
		-- call __index metamethod
		-- will find Point.move() function
		p3:move(1, 1)

		-- call __index metamethod
		print("x = " .. (p3.x) .. ", y = " .. (p3.y))

		-- call __newindex metamethod
		p3.x = 11
		p3.y = 12

		-- call __string metamethod
		print(p3)
	`

	// create a new table
	l.NewTable()

	// bind new() and move() function
	l.PushCFunction(unsafe.Pointer(C.NewPoint))
	l.SetField(-2, "new")
	l.PushCFunction(unsafe.Pointer(C.Move))
	l.SetField(-2, "move")

	// named the table as Point and set it global
	l.SetGlobal("Point")

	// create a metatable
	l.LNewMetaTable("PointMetaTable")

	l.PushCFunction(unsafe.Pointer(C.Add))
	l.SetField(-2, "__add")
	l.PushCFunction(unsafe.Pointer(C.ToString))
	l.SetField(-2, "__tostring")

	// bind metamethod __index()
	// just like rewrite '->' in c++
	l.PushCFunction(unsafe.Pointer(C.Index))
	l.SetField(-2, "__index")

	l.PushCFunction(unsafe.Pointer(C.NewIndex))
	l.SetField(-2, "__newindex")

	if l.LDoString(LUA_SCRIPT) != luajit.LUA_OK {
		fmt.Println(l.ToString(-1))
	}
}

//export Count
func Count(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	// get our upvalue
	c := l.ToInteger(luajit.UpValueIndex(1))
	l.PushInteger(c + 1)

	// copy our result and update our upvalue
	l.PushValue(-1)
	l.Replace(luajit.UpValueIndex(1))

	return 1
}

//export NewCounter
func NewCounter(ptr *C.struct_lua_State) C.int {
	l := luajit.FromCLuaState(unsafe.Pointer(ptr))

	// initial our upvalue
	l.PushInteger(0)

	// return the closure and the number of upvalue is 1
	l.PushCClosure(unsafe.Pointer(C.Count), 1)
	return 1
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
	//Tables()

	MetaTable()

	//Closure()
}
