package dto

// PaginationQuery holds parsed pagination query parameters.
type PaginationQuery struct {
	Page      int    `form:"page,default=1" binding:"omitempty,min=1"`
	PerPage   int    `form:"per_page,default=10" binding:"omitempty,min=1,max=100"`
	SortBy    string `form:"sort_by,default=created_at"`
	SortOrder string `form:"sort_order,default=desc" binding:"omitempty,oneof=asc desc"`
}

// GetOffset calculates the SQL offset from page and per_page values.
func (p *PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// Sanitize applies defaults to empty or zero-value fields.
func (p *PaginationQuery) Sanitize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PerPage <= 0 {
		p.PerPage = 10
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}
