package finder

type Dict interface {
	Len() int
	Ideograph(i int) (string, error)
	IdeographsSimplified(i int) (string, error)
}
