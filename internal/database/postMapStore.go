package database

import (
	"fmt"

	"github.com/astgot/forum/internal/model"
)

// InsertPostMapInfo ...
func (d *Database) InsertPostMapInfo(postID int64, category string) error {
	stmnt, err := d.db.Prepare("INSERT INTO PostMapping (postID, category) VALUES (?, ?)")
	defer stmnt.Close()
	if err != nil {
		fmt.Println("postmap insert error")
		return err
	}
	stmnt.Exec(postID, category)
	return nil
}

// GetCategoryOfPost ...
func (d *Database) GetCategoryOfPost(postID int64) ([]string, error) {
	categoriesNames := []string{}
	res, err := d.db.Query("SELECT category FROM PostMapping WHERE postID = ?", postID)
	if err != nil { // if err == sql.ErrNoRows ---> if no category in the post
		return nil, err
	}
	defer res.Close()
	/* Here we retrieve all threads relating with one single post*/
	for res.Next() {
		postMap := model.NewPostMap()
		if err := res.Scan(&postMap.Category); err != nil {
			fmt.Println("error func\"GetThreadOfPost()\"")
			return nil, err
		}
		categoriesNames = append(categoriesNames, postMap.Category)
	}
	return categoriesNames, nil
}

// GetPostIDsByCategoryName ...
func (d *Database) GetPostIDsByCategoryName(category string) ([]int64, error) {
	postIDs := []int64{}
	if err := d.Open(); err != nil {
		return nil, err
	}
	query, err := d.db.Query("SELECT postID FROM PostMapping WHERE category=? ORDER BY postID DESC", category)
	if err != nil {
		fmt.Println("GetPostIDsByTID", err.Error())
		return nil, err
	}
	defer query.Close()
	for query.Next() {
		var id int64
		if err := query.Scan(&id); err != nil {
			fmt.Println("GetPostIDsByTID", err.Error())
			return nil, err
		}
		postIDs = append(postIDs, id)
	}
	return postIDs, nil
}
