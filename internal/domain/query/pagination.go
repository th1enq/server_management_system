package query

type Pagination struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}
