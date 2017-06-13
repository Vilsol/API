package data

type ServerQuery interface {
	QueryServer() (ServerQueryData, error)
}

type ServerQueryData struct {
}

type ServerData struct {
	address string
	port    int
}
