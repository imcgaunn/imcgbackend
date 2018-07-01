package index

import (
	"database/sql"
	"fmt"
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
	res, _ := AddIndexEntry(BlogIndexEntry{
		PostS3Loc:     "greatbucket/key/path/here",
		PostMetaS3Loc: "greatbucket/key/path/here",
		CreatedTime:   time.Now()}, db)
	id, err := res.LastInsertId()
	t.Log(fmt.Sprintf("inserted at id: %d", id))
	if err != nil {
		t.Errorf("failed to insert index entry")
	}
	// ACT
	RemoveIndexEntry(id, db)
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
	res, _ := AddIndexEntry(BlogIndexEntry{
		ID:            111111,
		PostS3Loc:     "greatbucket/key/path/here",
		PostMetaS3Loc: "greatbucket/key/path/here",
		CreatedTime:   time.Now()}, db)
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

func TestSplitS3Uri(t *testing.T) {
	testS3Uri := "s3://imcgaunn-blog-posts/cool/key/yo"
	bucket, key := splitS3Uri(testS3Uri)
	t.Log(fmt.Sprintf("bucket: %s", bucket))
	t.Log(fmt.Sprintf("key: %s", key))
	if bucket != "imcgaunn-blog-posts" {
		t.Errorf("failed to extract bucket correctly")
	}
	if key != "cool/key/yo" {
		t.Errorf("failed to extract key correctly")
	}
}
