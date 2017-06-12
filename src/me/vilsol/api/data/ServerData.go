package data

type ServerQuery interface {
	QueryServer() ServerQueryData
}

type ServerQueryData struct {
}

type ServerData struct {
	address string
	port    int
}
