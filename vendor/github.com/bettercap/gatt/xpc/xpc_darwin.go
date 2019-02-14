package xpc

/*
#include "xpc_wrapper_darwin.h"
#include "sys/utsname.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"log"
	r "reflect"
	"unsafe"
)

type XPC struct {
	conn C.xpc_connection_t
}

func (x *XPC) Send(msg interface{}, verbose bool) {
	// verbose == true converts the type from bool to C._Bool
	C.XpcSendMessage(x.conn, goToXpc(msg), true, verbose == true)
}

//
// minimal XPC support required for BLE
//

// a dictionary of things
type Dict map[string]interface{}

func (d Dict) Contains(k string) bool {
	_, ok := d[k]
	return ok
}

func (d Dict) MustGetDict(k string) Dict {
	return d[k].(Dict)
}

func (d Dict) MustGetArray(k string) Array {
	return d[k].(Array)
}

func (d Dict) MustGetBytes(k string) []byte {
	switch t := d[k].(type) {
	case string:
		return []byte(d[k].(string))
	case []byte:
		return d[k].([]byte)
	default:
		panic(fmt.Sprintf("unknown type %v", t))
	}
}

func (d Dict) MustGetHexBytes(k string) string {
	return fmt.Sprintf("%x", d[k].([]byte))
}

func (d Dict) MustGetInt(k string) int {
	return int(d[k].(int64))
}

func (d Dict) MustGetUUID(k string) []byte {
	u := d[k].(UUID)
	return u[:]
}

func (d Dict) GetString(k, defv string) string {
	if v := d[k]; v != nil {
		//log.Printf("GetString %s %#v\n", k, v)
		return v.(string)
	} else {
		//log.Printf("GetString %s default %#v\n", k, defv)
		return defv
	}
}

func (d Dict) GetBytes(k string, defv []byte) []byte {
	if v := d[k]; v != nil {
		//log.Printf("GetBytes %s %#v\n", k, v)
		return v.([]byte)
	} else {
		//log.Printf("GetBytes %s default %#v\n", k, defv)
		return defv
	}
}

func (d Dict) GetInt(k string, defv int) int {
	if v := d[k]; v != nil {
		//log.Printf("GetString %s %#v\n", k, v)
		return int(v.(int64))
	} else {
		//log.Printf("GetString %s default %#v\n", k, defv)
		return defv
	}
}

func (d Dict) GetUUID(k string) UUID {
	return GetUUID(d[k])
}

// an array of things
type Array []interface{}

func (a Array) GetUUID(k int) UUID {
	return GetUUID(a[k])
}

// a UUID
type UUID []byte

func MakeUUID(s string) UUID {
	var sl []byte
	fmt.Sscanf(s, "%32x", &sl)

	var uuid [16]byte
	copy(uuid[:], sl)
	return UUID(uuid[:])
}

func (uuid UUID) String() string {
	return fmt.Sprintf("%x", []byte(uuid))
}

func GetUUID(v interface{}) UUID {
	if v == nil {
		return UUID{}
	}

	if uuid, ok := v.(UUID); ok {
		return uuid
	}

	if bytes, ok := v.([]byte); ok {
		uuid := UUID{}

		for i, b := range bytes {
			uuid[i] = b
		}

		return uuid
	}

	if bytes, ok := v.([]uint8); ok {
		uuid := UUID{}

		for i, b := range bytes {
			uuid[i] = b
		}

		return uuid
	}

	log.Fatalf("invalid type for UUID: %#v", v)
	return UUID{}
}

var (
	CONNECTION_INVALID     = errors.New("connection invalid")
	CONNECTION_INTERRUPTED = errors.New("connection interrupted")
	CONNECTION_TERMINATED  = errors.New("connection terminated")

	TYPE_OF_UUID  = r.TypeOf(UUID{})
	TYPE_OF_BYTES = r.TypeOf([]byte{})
	handlers      = map[uintptr]XpcEventHandler{}
)

type XpcEventHandler interface {
	HandleXpcEvent(event Dict, err error)
}

func XpcConnect(service string, eh XpcEventHandler) XPC {
	ctx := uintptr(unsafe.Pointer(&eh))
	handlers[ctx] = eh
	cservice := C.CString(service)
	defer C.free(unsafe.Pointer(cservice))
	return XPC{conn: C.XpcConnect(cservice, C.uintptr_t(ctx))}
}

//export handleXpcEvent
func handleXpcEvent(event C.xpc_object_t, p C.ulong) {
	//log.Printf("handleXpcEvent %#v %#v\n", event, p)

	t := C.xpc_get_type(event)
	eh := handlers[uintptr(p)]
	if eh == nil {
		//log.Println("no handler for", p)
		return
	}

	if t == C.TYPE_ERROR {
		if event == C.ERROR_CONNECTION_INVALID {
			// The client process on the other end of the connection has either
			// crashed or cancelled the connection. After receiving this error,
			// the connection is in an invalid state, and you do not need to
			// call xpc_connection_cancel(). Just tear down any associated state
			// here.
			//log.Println("connection invalid")
			eh.HandleXpcEvent(nil, CONNECTION_INVALID)
		} else if event == C.ERROR_CONNECTION_INTERRUPTED {
			//log.Println("connection interrupted")
			eh.HandleXpcEvent(nil, CONNECTION_INTERRUPTED)
		} else if event == C.ERROR_CONNECTION_TERMINATED {
			// Handle per-connection termination cleanup.
			//log.Println("connection terminated")
			eh.HandleXpcEvent(nil, CONNECTION_TERMINATED)
		} else {
			//log.Println("got some error", event)
			eh.HandleXpcEvent(nil, fmt.Errorf("%v", event))
		}
	} else {
		eh.HandleXpcEvent(xpcToGo(event).(Dict), nil)
	}
}

// goToXpc converts a go object to an xpc object
func goToXpc(o interface{}) C.xpc_object_t {
	return valueToXpc(r.ValueOf(o))
}

// valueToXpc converts a go Value to an xpc object
//
// note that not all the types are supported, but only the subset required for Blued
func valueToXpc(val r.Value) C.xpc_object_t {
	if !val.IsValid() {
		return nil
	}

	var xv C.xpc_object_t

	switch val.Kind() {
	case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
		xv = C.xpc_int64_create(C.int64_t(val.Int()))

	case r.Uint, r.Uint8, r.Uint16, r.Uint32:
		xv = C.xpc_int64_create(C.int64_t(val.Uint()))

	case r.String:
		xv = C.xpc_string_create(C.CString(val.String()))

	case r.Map:
		xv = C.xpc_dictionary_create(nil, nil, 0)
		for _, k := range val.MapKeys() {
			v := valueToXpc(val.MapIndex(k))
			C.xpc_dictionary_set_value(xv, C.CString(k.String()), v)
			if v != nil {
				C.xpc_release(v)
			}
		}

	case r.Array, r.Slice:
		if val.Type() == TYPE_OF_UUID {
			// array of bytes
			var uuid [16]byte
			r.Copy(r.ValueOf(uuid[:]), val)
			xv = C.xpc_uuid_create(C.ptr_to_uuid(unsafe.Pointer(&uuid[0])))
		} else if val.Type() == TYPE_OF_BYTES {
			// slice of bytes
			if val.Len() > 0 {
				v := val.Interface().([]byte)
				xv = C.xpc_data_create(unsafe.Pointer(&v[0]), C.size_t(val.Len()))
			} else {
				xv = C.xpc_data_create(nil, C.size_t(0))
			}
		} else {
			xv = C.xpc_array_create(nil, 0)
			l := val.Len()

			for i := 0; i < l; i++ {
				v := valueToXpc(val.Index(i))
				C.xpc_array_append_value(xv, v)
				if v != nil {
					C.xpc_release(v)
				}
			}
		}

	case r.Interface, r.Ptr:
		xv = valueToXpc(val.Elem())

	default:
		log.Fatalf("unsupported %#v", val.String())
	}

	return xv
}

//export arraySet
func arraySet(u C.uintptr_t, i C.int, v C.xpc_object_t) {
	a := *(*Array)(unsafe.Pointer(uintptr(u)))
	a[i] = xpcToGo(v)
}

//export dictSet
func dictSet(u C.uintptr_t, k *C.char, v C.xpc_object_t) {
	d := *(*Dict)(unsafe.Pointer(uintptr(u)))
	d[C.GoString(k)] = xpcToGo(v)
}

// xpcToGo converts an xpc object to a go object
//
// note that not all the types are supported, but only the subset required for Blued
func xpcToGo(v C.xpc_object_t) interface{} {
	t := C.xpc_get_type(v)

	switch t {
	case C.TYPE_ARRAY:
		a := make(Array, C.int(C.xpc_array_get_count(v)))
		p := uintptr(unsafe.Pointer(&a))
		C.XpcArrayApply(C.uintptr_t(p), v)
		return a

	case C.TYPE_DATA:
		return C.GoBytes(C.xpc_data_get_bytes_ptr(v), C.int(C.xpc_data_get_length(v)))

	case C.TYPE_DICT:
		d := make(Dict)
		p := uintptr(unsafe.Pointer(&d))
		C.XpcDictApply(C.uintptr_t(p), v)
		return d

	case C.TYPE_INT64:
		return int64(C.xpc_int64_get_value(v))

	case C.TYPE_STRING:
		return C.GoString(C.xpc_string_get_string_ptr(v))

	case C.TYPE_UUID:
		a := [16]byte{}
		C.XpcUUIDGetBytes(unsafe.Pointer(&a), v)
		return UUID(a[:])

	default:
		log.Fatalf("unexpected type %#v, value %#v", t, v)
	}

	return nil
}

// xpc_release is needed by tests, since they can't use CGO
func xpc_release(xv C.xpc_object_t) {
	C.xpc_release(xv)
}

// this is used to check the OS version

type Utsname struct {
	Sysname  string
	Nodename string
	Release  string
	Version  string
	Machine  string
}

func Uname(utsname *Utsname) error {
	var cstruct C.struct_utsname
	if err := C.uname(&cstruct); err != 0 {
		return errors.New("utsname error")
	}

	// XXX: this may crash if any value is exactly 256 characters (no 0 terminator)
	utsname.Sysname = C.GoString(&cstruct.sysname[0])
	utsname.Nodename = C.GoString(&cstruct.nodename[0])
	utsname.Release = C.GoString(&cstruct.release[0])
	utsname.Version = C.GoString(&cstruct.version[0])
	utsname.Machine = C.GoString(&cstruct.machine[0])

	return nil
}
