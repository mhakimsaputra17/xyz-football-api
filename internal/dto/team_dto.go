package dto

// CreateTeamRequest represents the request payload for creating a team.
type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required" example:"Persija Jakarta"`
	LogoURL     string `json:"logo_url" binding:"omitempty,url" example:"https://example.com/persija-logo.png"`
	FoundedYear int    `json:"founded_year" binding:"omitempty,min=1800,max=2100" example:"1928"`
	Address     string `json:"address" binding:"omitempty" example:"Jakarta International Stadium"`
	City        string `json:"city" binding:"omitempty" example:"Jakarta"`
}

// UpdateTeamRequest represents the request payload for updating a team.
type UpdateTeamRequest struct {
	Name        string `json:"name" binding:"required" example:"Persija Jakarta"`
	LogoURL     string `json:"logo_url" binding:"omitempty,url" example:"https://example.com/persija-logo.png"`
	FoundedYear int    `json:"founded_year" binding:"omitempty,min=1800,max=2100" example:"1928"`
	Address     string `json:"address" binding:"omitempty" example:"Jakarta International Stadium"`
	City        string `json:"city" binding:"omitempty" example:"Jakarta"`
}

// TeamResponse represents the team data returned in API responses.
type TeamResponse struct {
	ID          string `json:"id" example:"019292f0-6b00-7a50-8d00-000000000010"`
	Name        string `json:"name" example:"Persija Jakarta"`
	LogoURL     string `json:"logo_url" example:"https://example.com/persija-logo.png"`
	FoundedYear int    `json:"founded_year" example:"1928"`
	Address     string `json:"address" example:"Jakarta International Stadium"`
	City        string `json:"city" example:"Jakarta"`
	CreatedAt   string `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2025-01-15T10:30:00Z"`
}
