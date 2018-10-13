package data

// FlushPolicy is the type of flush policy to use.
type FlushPolicy int

const (
	// FlushOnEdit saves the object to disk after every modification.
	FlushOnEdit FlushPolicy = iota
	// FlushExplicit saves the object to disk only if the Flush method of
	// the object is explicitly called.
	FlushExplicit
	// FlushNone never saves the object to disk.
	FlushNone
)
