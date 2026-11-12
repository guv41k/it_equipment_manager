package domain

type Category struct {
	ID          int64
	Name        string
	Description *string
}

func (c Category) Validate() error {
	return validateCatalogFields(c.Name, c.Description)
}
