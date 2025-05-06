package models

type Client struct {
	ID              uint64
	IP              string
	Capacity        int
	RatePerInterval int
	Tokens          int
}
