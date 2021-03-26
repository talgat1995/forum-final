package database

import (
	"fmt"

	"github.com/astgot/forum/internal/model"
)

// InsertCategories ...
func (d *Database) InsertCategories() error {
	if err := d.Open(); err != nil {
		return err
	}
	categories := []string{"music", "tech", "sport"}
	for _, category := range categories {
		stmnt, err := d.db.Prepare("INSERT INTO Categories (Name) VALUES (?)")
		if err != nil {
			return err
		}
		defer stmnt.Close()
		if _, err = stmnt.Exec(category); err != nil {
			return err
		}
	}
	return nil
}

// GetAllCategories ...
func (d *Database) GetAllCategories() []*model.Category {
	var categories []*model.Category
	res, err := d.db.Query("SELECT * FROM Categories")
	if err != nil {
		fmt.Println("Category query error")
		return nil
	}
	defer res.Close()
	for res.Next() {
		category := model.NewCategory()
		if err := res.Scan(&category.ID, &category.Name); err != nil {
			fmt.Println("Category scan error")
			return nil
		}
		categories = append(categories, category)

	}
	return categories
}

// GetCategoryID ...
func (d *Database) GetCategoryID(name string) int64 {
	var id int64
	if err := d.db.QueryRow("SELECT ID FROM Categories WHERE Name = ?", name).Scan(&id); err != nil {
		fmt.Println("CategoryID retrieve error")
	}
	return id
}

// GetCategoryByID ...
func (d *Database) GetCategoryByID(id int64) (*model.Category, error) {
	category := model.NewCategory()
	if err := d.db.QueryRow("SELECT * FROM Categories WHERE ID = ?", id).Scan(&category.ID, &category.Name); err != nil {
		fmt.Println("error on func GetThreadByID()")
		return nil, err
	}
	return category, nil
}
