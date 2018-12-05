package v8

// #include "v8_c_bridge.h"
// #cgo CXXFLAGS: -I${SRCDIR} -I${SRCDIR}/include -g3 -fpic -std=c++11
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

type CallerInfo struct {
	Name     string
	Filename string
	Line     int
	Column   int
}

type FunctionTemplate struct {
	context *Context
	pointer C.FunctionTemplatePtr
}

type ObjectTemplate struct {
	context *Context
	pointer C.ObjectTemplatePtr
}

type Function func(FunctionArgs) (*Value, error)
type Getter func(GetterArgs) (*Value, error)
type Setter func(SetterArgs) error

type FunctionArgs struct {
	Context         *Context
	Caller          CallerInfo
	This            *Value
	Holder          *Value
	IsConstructCall bool
	Args            []*Value
}

func (c *FunctionArgs) Arg(n int) *Value {
	if n < len(c.Args) && n >= 0 {
		return c.Args[n]
	}
	return c.Context.Undefined()
}

type GetterArgs struct {
	Context *Context
	Caller  CallerInfo
	This    *Value
	Holder  *Value
	Key     string
}

type SetterArgs struct {
	Context *Context
	Caller  CallerInfo
	This    *Value
	Holder  *Value
	Key     string
	Value   *Value
}

type functionInfo struct {
	Function
	id ID
}

type accessorInfo struct {
	Getter
	Setter
	id ID
}

func (i *functionInfo) GetID() ID {
	return i.id
}

func (i *functionInfo) SetID(id ID) {
	i.id = id
}

func (i *accessorInfo) GetID() ID {
	return i.id
}

func (i *accessorInfo) SetID(id ID) {
	i.id = id
}

func (c *Context) NewFunctionTemplate(cb Function) *FunctionTemplate {
	iid := c.isolate.ref()
	defer c.isolate.unref()

	cid := c.ref()
	defer c.unref()

	id := c.functions.Ref(&functionInfo{cb, 0})
	pid := C.CString(fmt.Sprintf("%d:%d:%d", iid, cid, id))
	defer C.free(unsafe.Pointer(pid))

	pf := C.v8_FunctionTemplate_New(c.pointer, pid)

	f := &FunctionTemplate{c, pf}
	runtime.SetFinalizer(f, (*FunctionTemplate).release)
	return f
}

func (f *FunctionTemplate) Inherit(parent *FunctionTemplate) {
	f.context.ref()
	defer f.context.unref()

	C.v8_FunctionTemplate_Inherit(f.context.pointer, f.pointer, parent.pointer)
}

func (f *FunctionTemplate) SetName(name string) {
	pname := C.CString(name)
	defer C.free(unsafe.Pointer(pname))

	f.context.ref()
	defer f.context.unref()

	C.v8_FunctionTemplate_SetName(f.context.pointer, f.pointer, pname)
}

func (f *FunctionTemplate) SetHiddenPrototype(value bool) {
	f.context.ref()
	defer f.context.unref()

	C.v8_FunctionTemplate_SetHiddenPrototype(f.context.pointer, f.pointer, C.bool(value))
}

func (f *FunctionTemplate) GetFunction() *Value {
	f.context.ref()
	defer f.context.unref()

	pv := C.v8_FunctionTemplate_GetFunction(f.context.pointer, f.pointer)

	return f.context.newValue(pv, unionKindFunction)
}

func (f *FunctionTemplate) GetInstanceTemplate() *ObjectTemplate {
	f.context.ref()
	defer f.context.unref()

	po := C.v8_FunctionTemplate_InstanceTemplate(f.context.pointer, f.pointer)
	return &ObjectTemplate{f.context, po}
}

func (f *FunctionTemplate) GetPrototypeTemplate() *ObjectTemplate {
	f.context.ref()
	defer f.context.unref()

	pp := C.v8_FunctionTemplate_PrototypeTemplate(f.context.pointer, f.pointer)
	return &ObjectTemplate{f.context, pp}
}

func (f *FunctionTemplate) release() {
	if f.pointer != nil {
		f.context.ref()
		C.v8_FunctionTemplate_Release(f.context.pointer, f.pointer)
		f.context.unref()
	}
	f.context = nil
	f.pointer = nil
	runtime.SetFinalizer(f, nil)
}

func (o *ObjectTemplate) SetInternalFieldCount(count int) {
	o.context.ref()
	defer o.context.unref()

	C.v8_ObjectTemplate_SetInternalFieldCount(o.context.pointer, o.pointer, C.int(count))
}

func (o *ObjectTemplate) SetAccessor(name string, getter Getter, setter Setter) {
	iid := o.context.isolate.ref()
	defer o.context.isolate.unref()

	cid := o.context.ref()
	defer o.context.unref()

	id := o.context.accessors.Ref(&accessorInfo{getter, setter, 0})
	pid := C.CString(fmt.Sprintf("%d:%d:%d", iid, cid, id))
	defer C.free(unsafe.Pointer(pid))

	pname := C.CString(name)
	defer C.free(unsafe.Pointer(pname))

	C.v8_ObjectTemplate_SetAccessor(o.context.pointer, o.pointer, pname, pid, setter != nil)
}

func (o *ObjectTemplate) release() {
	if o.pointer != nil {
		o.context.ref()
		C.v8_ObjectTemplate_Release(o.context.pointer, o.pointer)
		o.context.unref()
	}
	o.context = nil
	o.pointer = nil
	runtime.SetFinalizer(o, nil)
}