package model

// PostMap ...
type PostMap struct {
	Category string `db:"category"`
	PostID   int64  `db:"postID"`
}

// NewPostMap ..
func NewPostMap() *PostMap {
	return &PostMap{}
}
