package util

type BytePool struct {
	pool  chan []byte
	width int
}

func NewBytePool(width int, depth int) *BytePool {
	return &BytePool{
		pool:  make(chan []byte, depth),
		width: width,
	}
}

func (p *BytePool) Get() (b []byte) {
	select {
	case b = <-p.pool:
	default:
		b = make([]byte, p.width)
	}
	return b
}

func (p *BytePool) Put(b []byte) {
	select {
	case p.pool <- b:
	default:
	}
}
