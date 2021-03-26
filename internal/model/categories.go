package model

// Category ...
type Category struct {
	ID   int64    `db:"ID"`
	Name []string `db:"Name"`
}

// NewCategory ....
func NewCategory() *Category {
	return &Category{}
}
