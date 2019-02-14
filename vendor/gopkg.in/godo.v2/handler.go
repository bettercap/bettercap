package godo

// Handler is the interface which all task handlers eventually implement.
type Handler interface {
	Handle(*Context)
}

// // HandlerFunc is Handler adapter.
// type handlerFunc func() error

// // Handle implements Handler.
// func (f handlerFunc) Handle(*Context) error {
// 	return f()
// }

// // VoidHandlerFunc is a Handler adapter.
// type voidHandlerFunc func()

// // Handle implements Handler.
// func (v voidHandlerFunc) Handle(*Context) error {
// 	v()
// 	return nil
// }

// // ContextHandlerFunc is a Handler adapter.
// type contextHandlerFunc func(*Context) error

// // Handle implements Handler.
// func (c contextHandlerFunc) Handle(ctx *Context) error {
// 	return c(ctx)
// }

// HandlerFunc is a Handler adapter.
type HandlerFunc func(*Context)

// Handle implements Handler.
func (f HandlerFunc) Handle(ctx *Context) {
	f(ctx)
}
