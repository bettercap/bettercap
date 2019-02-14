#ifndef _XPC_WRAPPER_H_
#define _XPC_WRAPPER_H_

#include <stdlib.h>
#include <stdio.h>
#include <xpc/xpc.h>

extern xpc_type_t TYPE_ERROR;

extern xpc_type_t TYPE_ARRAY;
extern xpc_type_t TYPE_DATA;
extern xpc_type_t TYPE_DICT;
extern xpc_type_t TYPE_INT64;
extern xpc_type_t TYPE_STRING;
extern xpc_type_t TYPE_UUID;

extern xpc_object_t ERROR_CONNECTION_INVALID;
extern xpc_object_t ERROR_CONNECTION_INTERRUPTED;
extern xpc_object_t ERROR_CONNECTION_TERMINATED;

extern xpc_connection_t XpcConnect(char *, uintptr_t);
extern void XpcSendMessage(xpc_connection_t, xpc_object_t, bool, bool);
extern void XpcArrayApply(uintptr_t, xpc_object_t);
extern void XpcDictApply(uintptr_t, xpc_object_t);
extern void XpcUUIDGetBytes(void *, xpc_object_t);

// the input type for xpc_uuid_create should be uuid_t but CGO instists on unsigned char *
// typedef uuid_t * ptr_to_uuid_t;
typedef unsigned char * ptr_to_uuid_t;
extern const ptr_to_uuid_t ptr_to_uuid(void *p);

#endif
