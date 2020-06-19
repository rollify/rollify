package model

// PaginationCursors has the information regarding the listing cursors of a
// paginated list of models through the application.
type PaginationCursors struct {
	FirstCursor string
	LastCursor  string
	HasNext     bool
	HasPrevious bool
}

// PaginationOrder is the order used to list.
type PaginationOrder int

const (
	// PaginationOrderDefault is the default order used to list pages.
	PaginationOrderDefault PaginationOrder = iota
	// PaginationOrderAsc is the ascendant order used to list pages.
	PaginationOrderAsc
	// PaginationOrderDesc is the descendant order used to list pages.
	PaginationOrderDesc
)

// PaginationOpts are the options used to paginate models.
type PaginationOpts struct {
	Cursor string
	Size   uint
	Order  PaginationOrder
}
