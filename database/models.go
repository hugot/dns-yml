package database

type Domain struct {
	ID   int
	Name string
	Type string
}

type Record struct {
	ID       int
	Type     string
	DomainID int
	Name     string
	Content  string
	TTL      int
	Priority int
}
