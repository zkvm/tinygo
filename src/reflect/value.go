package reflect

import (
	"unsafe"
)

type valueFlags uint8

// Flags list some useful flags that contain some extra information not
// contained in an interface{} directly, like whether this value was exported at
// all (it is possible to read unexported fields using reflection, but it is not
// possible to modify them).
const (
	valueFlagIndirect valueFlags = 1 << iota
	valueFlagExported
)

type Value struct {
	typecode Type
	value    unsafe.Pointer
	flags    valueFlags
}

// isIndirect returns whether the value pointer in this Value is always a
// pointer to the value. If it is false, it is only a pointer to the value if
// the value is bigger than a pointer.
func (v Value) isIndirect() bool {
	return v.flags&valueFlagIndirect != 0
}

func Indirect(v Value) Value {
	if v.Kind() != Ptr {
		return v
	}
	return v.Elem()
}

func ValueOf(i interface{}) Value {
	v := (*interfaceHeader)(unsafe.Pointer(&i))
	return Value{
		typecode: v.typecode,
		value:    v.value,
		flags:    valueFlagExported,
	}
}

func (v Value) Interface() interface{} {
	i := interfaceHeader{
		typecode: v.typecode,
		value:    v.value,
	}
	if v.isIndirect() && v.Type().Size() <= unsafe.Sizeof(uintptr(0)) {
		// Value was indirect but must be put back directly in the interface
		// value.
		var value uintptr
		for j := v.Type().Size(); j != 0; j-- {
			value = (value << 8) | uintptr(*(*uint8)(unsafe.Pointer(uintptr(v.value) + j - 1)))
		}
		i.value = unsafe.Pointer(value)
	}
	return *(*interface{})(unsafe.Pointer(&i))
}

func (v Value) Type() Type {
	return v.typecode
}

func (v Value) Kind() Kind {
	return v.Type().Kind()
}

// IsNil returns whether the value is the nil value. It panics if the value Kind
// is not a channel, map, pointer, function, slice, or interface.
func (v Value) IsNil() bool {
	switch v.Kind() {
	case Chan, Map, Ptr:
		if v.isIndirect() {
			return *(*uintptr)(v.value) == 0
		}
		return v.value == nil
	case Func:
		if v.value == nil {
			return true
		}
		fn := (*funcHeader)(v.value)
		return fn.Code == nil
	case Slice:
		if v.value == nil {
			return true
		}
		slice := (*SliceHeader)(v.value)
		return slice.Data == 0
	case Interface:
		if v.value == nil {
			return true
		}
		itf := (*interfaceHeader)(v.value)
		return itf.value == nil
	default:
		panic(&ValueError{"IsNil"})
	}
}

// Pointer returns the underlying pointer of the given value for the following
// types: chan, map, pointer, unsafe.Pointer, slice, func.
func (v Value) Pointer() uintptr {
	switch v.Kind() {
	case Chan, Map, Ptr, UnsafePointer:
		if v.isIndirect() {
			return *(*uintptr)(v.value)
		}
		return uintptr(v.value)
	case Slice:
		slice := (*SliceHeader)(v.value)
		return slice.Data
	case Func:
		panic("unimplemented: (reflect.Value).Pointer()")
	default: // not implemented: Func
		panic(&ValueError{"Pointer"})
	}
}

func (v Value) IsValid() bool {
	return v.typecode != 0
}

func (v Value) CanInterface() bool {
	// No Value types of private data can be constructed at the moment.
	return true
}

func (v Value) CanAddr() bool {
	panic("unimplemented: (reflect.Value).CanAddr()")
}

func (v Value) Addr() Value {
	panic("unimplemented: (reflect.Value).Addr()")
}

func (v Value) CanSet() bool {
	return v.flags&(valueFlagExported|valueFlagIndirect) == valueFlagExported|valueFlagIndirect
}

func (v Value) Bool() bool {
	switch v.Kind() {
	case Bool:
		if v.isIndirect() {
			return *((*bool)(v.value))
		} else {
			return uintptr(v.value) != 0
		}
	default:
		panic(&ValueError{"Bool"})
	}
}

func (v Value) Int() int64 {
	switch v.Kind() {
	case Int:
		if v.isIndirect() || unsafe.Sizeof(int(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int)(v.value))
		} else {
			return int64(int(uintptr(v.value)))
		}
	case Int8:
		if v.isIndirect() {
			return int64(*(*int8)(v.value))
		} else {
			return int64(int8(uintptr(v.value)))
		}
	case Int16:
		if v.isIndirect() {
			return int64(*(*int16)(v.value))
		} else {
			return int64(int16(uintptr(v.value)))
		}
	case Int32:
		if v.isIndirect() || unsafe.Sizeof(int32(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int32)(v.value))
		} else {
			return int64(int32(uintptr(v.value)))
		}
	case Int64:
		if v.isIndirect() || unsafe.Sizeof(int64(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int64)(v.value))
		} else {
			return int64(int64(uintptr(v.value)))
		}
	default:
		panic(&ValueError{"Int"})
	}
}

func (v Value) Uint() uint64 {
	switch v.Kind() {
	case Uintptr:
		if v.isIndirect() {
			return uint64(*(*uintptr)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint8:
		if v.isIndirect() {
			return uint64(*(*uint8)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint16:
		if v.isIndirect() {
			return uint64(*(*uint16)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint:
		if v.isIndirect() || unsafe.Sizeof(uint(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint32:
		if v.isIndirect() || unsafe.Sizeof(uint32(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint32)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint64:
		if v.isIndirect() || unsafe.Sizeof(uint64(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint64)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	default:
		panic(&ValueError{"Uint"})
	}
}

func (v Value) Float() float64 {
	switch v.Kind() {
	case Float32:
		if v.isIndirect() || unsafe.Sizeof(float32(0)) > unsafe.Sizeof(uintptr(0)) {
			// The float is stored as an external value on systems with 16-bit
			// pointers.
			return float64(*(*float32)(v.value))
		} else {
			// The float is directly stored in the interface value on systems
			// with 32-bit and 64-bit pointers.
			return float64(*(*float32)(unsafe.Pointer(&v.value)))
		}
	case Float64:
		if v.isIndirect() || unsafe.Sizeof(float64(0)) > unsafe.Sizeof(uintptr(0)) {
			// For systems with 16-bit and 32-bit pointers.
			return *(*float64)(v.value)
		} else {
			// The float is directly stored in the interface value on systems
			// with 64-bit pointers.
			return *(*float64)(unsafe.Pointer(&v.value))
		}
	default:
		panic(&ValueError{"Float"})
	}
}

func (v Value) Complex() complex128 {
	switch v.Kind() {
	case Complex64:
		if v.isIndirect() || unsafe.Sizeof(complex64(0)) > unsafe.Sizeof(uintptr(0)) {
			// The complex number is stored as an external value on systems with
			// 16-bit and 32-bit pointers.
			return complex128(*(*complex64)(v.value))
		} else {
			// The complex number is directly stored in the interface value on
			// systems with 64-bit pointers.
			return complex128(*(*complex64)(unsafe.Pointer(&v.value)))
		}
	case Complex128:
		// This is a 128-bit value, which is always stored as an external value.
		// It may be stored in the pointer directly on very uncommon
		// architectures with 128-bit pointers, however.
		return *(*complex128)(v.value)
	default:
		panic(&ValueError{"Complex"})
	}
}

func (v Value) String() string {
	switch v.Kind() {
	case String:
		// A string value is always bigger than a pointer as it is made of a
		// pointer and a length.
		return *(*string)(v.value)
	default:
		// Special case because of the special treatment of .String() in Go.
		return "<T>"
	}
}

func (v Value) Bytes() []byte {
	panic("unimplemented: (reflect.Value).Bytes()")
}

func (v Value) Slice(i, j int) Value {
	panic("unimplemented: (reflect.Value).Slice()")
}

// Len returns the length of this value for slices, strings, arrays, channels,
// and maps. For oter types, it panics.
func (v Value) Len() int {
	t := v.Type()
	switch t.Kind() {
	case Slice:
		return int((*SliceHeader)(v.value).Len)
	case String:
		return int((*StringHeader)(v.value).Len)
	case Array:
		return v.Type().Len()
	default: // Chan, Map
		panic("unimplemented: (reflect.Value).Len()")
	}
}

func (v Value) Cap() int {
	t := v.Type()
	switch t.Kind() {
	case Slice:
		return int((*SliceHeader)(v.value).Cap)
	default: // Array, Chan
		panic("unimplemented: (reflect.Value).Cap()")
	}
}

// NumField returns the number of fields of this struct. It panics for other
// value types.
func (v Value) NumField() int {
	return v.Type().NumField()
}

func (v Value) Elem() Value {
	switch v.Kind() {
	case Ptr:
		ptr := v.value
		if v.isIndirect() {
			ptr = *(*unsafe.Pointer)(ptr)
		}
		if ptr == nil {
			return Value{}
		}
		return Value{
			typecode: v.Type().Elem(),
			value:    ptr,
			flags:    v.flags | valueFlagIndirect,
		}
	default: // not implemented: Interface
		panic(&ValueError{"Elem"})
	}
}

// Field returns the value of the i'th field of this struct.
func (v Value) Field(i int) Value {
	structField := v.Type().Field(i)
	flags := v.flags
	if structField.PkgPath != "" {
		// The fact that PkgPath is present means that this field is not
		// exported.
		flags &^= valueFlagExported
	}

	size := v.Type().Size()
	fieldSize := structField.Type.Size()
	if v.isIndirect() || fieldSize > unsafe.Sizeof(uintptr(0)) {
		// v.value was already a pointer to the value and it should stay that
		// way.
		return Value{
			flags:    flags,
			typecode: structField.Type,
			value:    unsafe.Pointer(uintptr(v.value) + structField.Offset),
		}
	}

	// The fieldSize is smaller than uintptr, which means that the value will
	// have to be stored directly in the interface value.

	if fieldSize == 0 {
		// The struct field is zero sized.
		// This is a rare situation, but because it's undefined behavior
		// to shift the size of the value (zeroing the value), handle this
		// situation explicitly.
		return Value{
			flags:    flags,
			typecode: structField.Type,
			value:    unsafe.Pointer(uintptr(0)),
		}
	}

	if size > unsafe.Sizeof(uintptr(0)) {
		// The value was not stored in the interface before but will be
		// afterwards, so load the value (from the correct offset) and return
		// it.
		ptr := unsafe.Pointer(uintptr(v.value) + structField.Offset)
		value := unsafe.Pointer(loadValue(ptr, fieldSize))
		return Value{
			flags:    0,
			typecode: structField.Type,
			value:    value,
		}
	}

	// The value was already stored directly in the interface and it still
	// is. Cut out the part of the value that we need.
	value := maskAndShift(uintptr(v.value), structField.Offset, fieldSize)
	return Value{
		flags:    flags,
		typecode: structField.Type,
		value:    unsafe.Pointer(value),
	}
}

func (v Value) Index(i int) Value {
	switch v.Kind() {
	case Slice:
		// Extract an element from the slice.
		slice := *(*SliceHeader)(v.value)
		if uint(i) >= uint(slice.Len) {
			panic("reflect: slice index out of range")
		}
		elem := Value{
			typecode: v.Type().Elem(),
			flags:    v.flags | valueFlagIndirect,
		}
		addr := uintptr(slice.Data) + elem.Type().Size()*uintptr(i) // pointer to new value
		elem.value = unsafe.Pointer(addr)
		return elem
	case String:
		// Extract a character from a string.
		// A string is never stored directly in the interface, but always as a
		// pointer to the string value.
		s := *(*StringHeader)(v.value)
		if uint(i) >= uint(s.Len) {
			panic("reflect: string index out of range")
		}
		return Value{
			typecode: Uint8.basicType(),
			value:    unsafe.Pointer(uintptr(*(*uint8)(unsafe.Pointer(s.Data + uintptr(i))))),
		}
	case Array:
		// Extract an element from the array.
		elemType := v.Type().Elem()
		elemSize := elemType.Size()
		size := v.Type().Size()
		if size == 0 {
			// The element size is 0 and/or the length of the array is 0.
			return Value{
				typecode: v.Type().Elem(),
				flags:    v.flags,
			}
		}
		if elemSize > unsafe.Sizeof(uintptr(0)) {
			// The resulting value doesn't fit in a pointer so must be
			// indirect. Also, because size != 0 this implies that the array
			// length must be != 0, and thus that the total size is at least
			// elemSize.
			addr := uintptr(v.value) + elemSize*uintptr(i) // pointer to new value
			return Value{
				typecode: v.Type().Elem(),
				flags:    v.flags,
				value:    unsafe.Pointer(addr),
			}
		}

		if size > unsafe.Sizeof(uintptr(0)) {
			// The element fits in a pointer, but the array does not.
			// Load the value from the pointer.
			addr := uintptr(v.value) + elemSize*uintptr(i) // pointer to new value
			return Value{
				typecode: v.Type().Elem(),
				flags:    v.flags,
				value:    unsafe.Pointer(loadValue(unsafe.Pointer(addr), elemSize)),
			}
		}

		// The value fits in a pointer, so extract it with some shifting and
		// masking.
		offset := elemSize * uintptr(i)
		value := maskAndShift(uintptr(v.value), offset, elemSize)
		return Value{
			typecode: v.Type().Elem(),
			flags:    v.flags,
			value:    unsafe.Pointer(value),
		}
	default:
		panic(&ValueError{"Index"})
	}
}

// loadValue loads a value that may or may not be word-aligned. The number of
// bytes given in size are loaded. The biggest possible size it can load is that
// of an uintptr.
func loadValue(ptr unsafe.Pointer, size uintptr) uintptr {
	loadedValue := uintptr(0)
	shift := uintptr(0)
	for i := uintptr(0); i < size; i++ {
		loadedValue |= uintptr(*(*byte)(ptr)) << shift
		shift += 8
		ptr = unsafe.Pointer(uintptr(ptr) + 1)
	}
	return loadedValue
}

// maskAndShift cuts out a part of a uintptr. Note that the offset may not be 0.
func maskAndShift(value, offset, size uintptr) uintptr {
	mask := ^uintptr(0) >> ((unsafe.Sizeof(uintptr(0)) - size) * 8)
	return (uintptr(value) >> (offset * 8)) & mask
}

func (v Value) MapKeys() []Value {
	panic("unimplemented: (reflect.Value).MapKeys()")
}

func (v Value) MapIndex(key Value) Value {
	panic("unimplemented: (reflect.Value).MapIndex()")
}

func (v Value) MapRange() *MapIter {
	panic("unimplemented: (reflect.Value).MapRange()")
}

type MapIter struct {
}

func (it *MapIter) Key() Value {
	panic("unimplemented: (*reflect.MapIter).Key()")
}

func (it *MapIter) Value() Value {
	panic("unimplemented: (*reflect.MapIter).Value()")
}

func (it *MapIter) Next() bool {
	panic("unimplemented: (*reflect.MapIter).Next()")
}

func (v Value) Set(x Value) {
	v.checkAddressable()
	if !v.Type().AssignableTo(x.Type()) {
		panic("reflect: cannot set")
	}
	size := v.Type().Size()
	xptr := x.value
	if size <= unsafe.Sizeof(uintptr(0)) && !x.isIndirect() {
		value := x.value
		xptr = unsafe.Pointer(&value)
	}
	memcpy(v.value, xptr, size)
}

func (v Value) SetBool(x bool) {
	v.checkAddressable()
	switch v.Kind() {
	case Bool:
		*(*bool)(v.value) = x
	default:
		panic(&ValueError{"SetBool"})
	}
}

func (v Value) SetInt(x int64) {
	v.checkAddressable()
	switch v.Kind() {
	case Int:
		*(*int)(v.value) = int(x)
	case Int8:
		*(*int8)(v.value) = int8(x)
	case Int16:
		*(*int16)(v.value) = int16(x)
	case Int32:
		*(*int32)(v.value) = int32(x)
	case Int64:
		*(*int64)(v.value) = x
	default:
		panic(&ValueError{"SetInt"})
	}
}

func (v Value) SetUint(x uint64) {
	v.checkAddressable()
	switch v.Kind() {
	case Uint:
		*(*uint)(v.value) = uint(x)
	case Uint8:
		*(*uint8)(v.value) = uint8(x)
	case Uint16:
		*(*uint16)(v.value) = uint16(x)
	case Uint32:
		*(*uint32)(v.value) = uint32(x)
	case Uint64:
		*(*uint64)(v.value) = x
	case Uintptr:
		*(*uintptr)(v.value) = uintptr(x)
	default:
		panic(&ValueError{"SetUint"})
	}
}

func (v Value) SetFloat(x float64) {
	v.checkAddressable()
	switch v.Kind() {
	case Float32:
		*(*float32)(v.value) = float32(x)
	case Float64:
		*(*float64)(v.value) = x
	default:
		panic(&ValueError{"SetFloat"})
	}
}

func (v Value) SetComplex(x complex128) {
	v.checkAddressable()
	switch v.Kind() {
	case Complex64:
		*(*complex64)(v.value) = complex64(x)
	case Complex128:
		*(*complex128)(v.value) = x
	default:
		panic(&ValueError{"SetComplex"})
	}
}

func (v Value) SetString(x string) {
	v.checkAddressable()
	switch v.Kind() {
	case String:
		*(*string)(v.value) = x
	default:
		panic(&ValueError{"SetString"})
	}
}

func (v Value) checkAddressable() {
	if !v.isIndirect() {
		panic("reflect: value is not addressable")
	}
}

//go:linkname alloc runtime.alloc
func alloc(size uintptr) unsafe.Pointer

func MakeSlice(typ Type, len, cap int) Value {
	panic("unimplemented: reflect.MakeSlice()")
}

func Zero(typ Type) Value {
	panic("unimplemented: reflect.Zero()")
}

func New(typ Type) Value {
	data := alloc(typ.Size())
	val := Value{PtrTo(typ), data, 0}
	return val
}

type funcHeader struct {
	Context unsafe.Pointer
	Code    unsafe.Pointer
}

// This is the same thing as an interface{}.
type interfaceHeader struct {
	typecode Type
	value    unsafe.Pointer
}

type SliceHeader struct {
	Data uintptr
	Len  uintptr
	Cap  uintptr
}

type StringHeader struct {
	Data uintptr
	Len  uintptr
}

type ValueError struct {
	Method string
}

func (e *ValueError) Error() string {
	return "reflect: call of reflect.Value." + e.Method + " on invalid type"
}

//go:linkname memcpy runtime.memcpy
func memcpy(dst, src unsafe.Pointer, size uintptr)
