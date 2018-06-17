package blog

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestCanAddAndRemoveIndexEntry(t *testing.T) {

	// ARRANGE
	// create table, and add entry
	db, err := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	if err != nil {
		t.Errorf("failed to open sqlite database in memory")
	}
	CreateIndexTable("blogposts", db)
	res, err := AddIndexEntry(BlogIndexEntry{
		ID:               111111,
		post_s3_loc:      "greatbucket/key/path/here",
		post_meta_s3_loc: "greatbucket/key/path/here",
		created_time:     time.Now()}, db)
	id, err := res.LastInsertId()
	if err != nil {
		t.Errorf("failed to insert index entry")
	}
	// ACT
	err = RemoveIndexEntry(id, db)
	_, err = GetIndexEntry(id, db)
	if err == nil {
		t.Errorf("was able to retrieve removed post")
	}
}

func TestCanAddAndRetrieve(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	if err != nil {
		t.Errorf("failed to open sqlite database in memory")
	}
	CreateIndexTable("blogposts", db)
	res, err := AddIndexEntry(BlogIndexEntry{
		ID:               111111,
		post_s3_loc:      "greatbucket/key/path/here",
		post_meta_s3_loc: "greatbucket/key/path/here",
		created_time:     time.Now()}, db)
	id, err := res.LastInsertId()
	if err != nil {
		t.Errorf("failed to insert index entry")
	}
	// ACT / ASSERT
	retrievedEntry, err := GetIndexEntry(id, db)
	if err != nil {
		t.Errorf("failed to retrieve inserted entry")
	}
	t.Log(retrievedEntry)
}
