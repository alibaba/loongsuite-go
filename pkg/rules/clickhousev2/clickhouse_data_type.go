package clickhousev2

type clickhouseRequest struct {
	Statement string
	DbName    string
	User      string
	Addr      string
	Op        string
	BatchSize int
	Params    []any
}
