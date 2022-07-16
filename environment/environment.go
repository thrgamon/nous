package environment

type Environment int

const (
	Production Environment = iota + 1
	Development
)
