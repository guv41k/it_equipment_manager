package catalog_transport_http

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"gitlab.com/guv41k/it_equipment_manager/internal/core/domain"
)

type CatalogRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (request CatalogRequest) Validate() error {
	if strings.TrimSpace(request.Name) == "" || utf8.RuneCountInString(request.Name) > 100 {
		return fmt.Errorf("name must contain 1 to 100 characters")
	}
	if request.Description != nil && utf8.RuneCountInString(*request.Description) > 500 {
		return fmt.Errorf("description must have at most 500 characters")
	}
	return nil
}

type CatalogResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func categoryDTOFromDomain(category domain.Category) CatalogResponse {
	return CatalogResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	}
}

func locationDTOFromDomain(location domain.Location) CatalogResponse {
	return CatalogResponse{
		ID:          location.ID,
		Name:        location.Name,
		Description: location.Description,
	}
}
