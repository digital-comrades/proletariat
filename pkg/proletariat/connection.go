package proletariat

type Connection interface {
	Emit(value interface{}) error
}
