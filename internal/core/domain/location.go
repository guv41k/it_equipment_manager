package domain

type Location struct {
	ID          int64
	Name        string
	Description *string
}

func (l Location) Validate() error {
	return validateCatalogFields(l.Name, l.Description)
}
