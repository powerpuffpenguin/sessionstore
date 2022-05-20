package sessionstore

type Coder interface {
	Unmarshal(b []byte) (session interface{}, e error)
	Marshal(session interface{}) (b []byte, e error)
}
