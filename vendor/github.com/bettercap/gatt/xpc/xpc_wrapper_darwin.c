#include <dispatch/dispatch.h>
#include <xpc/xpc.h>
#include <xpc/connection.h>
#include <Block.h>
#include <stdlib.h>
#include <stdio.h>

#include "_cgo_export.h"

//
// types and errors are implemented as macros
// create some real objects to make them accessible to Go
//
xpc_type_t TYPE_ERROR = XPC_TYPE_ERROR;

xpc_type_t TYPE_ARRAY = XPC_TYPE_ARRAY;
xpc_type_t TYPE_DATA = XPC_TYPE_DATA;
xpc_type_t TYPE_DICT = XPC_TYPE_DICTIONARY;
xpc_type_t TYPE_INT64 = XPC_TYPE_INT64;
xpc_type_t TYPE_STRING = XPC_TYPE_STRING;
xpc_type_t TYPE_UUID = XPC_TYPE_UUID;

xpc_object_t ERROR_CONNECTION_INVALID = (xpc_object_t) XPC_ERROR_CONNECTION_INVALID;
xpc_object_t ERROR_CONNECTION_INTERRUPTED = (xpc_object_t) XPC_ERROR_CONNECTION_INTERRUPTED;
xpc_object_t ERROR_CONNECTION_TERMINATED = (xpc_object_t) XPC_ERROR_TERMINATION_IMMINENT;

const ptr_to_uuid_t ptr_to_uuid(void *p) { return (ptr_to_uuid_t)p; }


//
// connect to XPC service
//
xpc_connection_t XpcConnect(char *service, uintptr_t ctx) {
    dispatch_queue_t queue = dispatch_queue_create(service, 0);
    xpc_connection_t conn = xpc_connection_create_mach_service(service, queue, XPC_CONNECTION_MACH_SERVICE_PRIVILEGED);

    // making a local copy, that should be made "persistent" with the following Block_copy
    // GoInterface ictx = *((GoInterface*)ctx);

    xpc_connection_set_event_handler(conn,
        Block_copy(^(xpc_object_t event) {
                handleXpcEvent(event, ctx);
        })
    );

    xpc_connection_resume(conn);
    return conn;
}

void XpcSendMessage(xpc_connection_t conn, xpc_object_t message, bool release, bool reportDelivery) {
    xpc_connection_send_message(conn,  message);
    xpc_connection_send_barrier(conn, ^{
        // Block is invoked on connection's target queue
        // when 'message' has been sent.
        if (reportDelivery) { // maybe this could be a callback
            puts("message delivered");
        }
    });
    if (release) {
        xpc_release(message);
    }
}

void XpcArrayApply(uintptr_t v, xpc_object_t arr) {
  xpc_array_apply(arr, ^bool(size_t index, xpc_object_t value) {
    arraySet(v, index, value);
    return true;
  });
}

void XpcDictApply(uintptr_t v, xpc_object_t dict) {
  xpc_dictionary_apply(dict, ^bool(const char *key, xpc_object_t value) {
    dictSet(v, (char *)key, value);
    return true;
  });
}

void XpcUUIDGetBytes(void *v, xpc_object_t uuid) {
   const uint8_t *src = xpc_uuid_get_bytes(uuid);
   uint8_t *dest = (uint8_t *)v;

   for (int i=0; i < sizeof(uuid_t); i++) {
     dest[i] = src[i];
   }
}
