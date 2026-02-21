package dto

// CreateTeamRequest represents the request payload for creating a team.
type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required"`
	LogoURL     string `json:"logo_url" binding:"omitempty,url"`
	FoundedYear int    `json:"founded_year" binding:"omitempty,min=1800,max=2100"`
	Address     string `json:"address" binding:"omitempty"`
	City        string `json:"city" binding:"omitempty"`
}

// UpdateTeamRequest represents the request payload for updating a team.
type UpdateTeamRequest struct {
	Name        string `json:"name" binding:"required"`
	LogoURL     string `json:"logo_url" binding:"omitempty,url"`
	FoundedYear int    `json:"founded_year" binding:"omitempty,min=1800,max=2100"`
	Address     string `json:"address" binding:"omitempty"`
	City        string `json:"city" binding:"omitempty"`
}

// TeamResponse represents the team data returned in API responses.
type TeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	FoundedYear int    `json:"founded_year"`
	Address     string `json:"address"`
	City        string `json:"city"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
