package session_modules

type JSHeader struct {
	Name  string
	Value string
}

type JSRequest struct {
	Method   string
	Version  string
	Path     string
	Hostname string
	Headers  []JSHeader
	Body     string
}

func (j *JSRequest) ReadBody() string {
	return "TODO: read body"
}
