package zerogod

import "github.com/phin1x/go-ipp"

func ippOnUnhandledRequest(ctx *HandlerContext, ipp_req *ipp.Request, ipp_op_name string) {
	ctx.mod.Warning("unhandled request from %v: operation=%s - %++v", ctx.client.RemoteAddr(), ipp_op_name, *ipp_req)

	ippSendResponse(ctx, ipp.NewResponse(
		ipp.StatusErrorOperationNotSupported,
		ipp_req.RequestId))
}
