package types

type Event struct {
	Fd      int
	Message string
}

type Response struct {
	Fd      int
	Message string
	Close   bool
}
