package codec

type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq           uint64 // sequence number chosen by client
	Error         string
}

type Codec interface {
	ReadHeader(*Header) error
	ReadBody(any) error
	WriteResponse(*Header, any) error
	Close() error
}

type CodeType string

const (
	GobType  CodeType = "application/gob"
	JsonType CodeType = "application/json" // not implemented
)
