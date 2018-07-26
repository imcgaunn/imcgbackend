package index

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"log"

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
	CreateIndexTable(db)
	res, _ := AddIndexEntry(BlogIndexEntry{
		PostS3Loc:     "greatbucket/key/path/here",
		Title: "great title!",
		Tags: "silly, billy",
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

func createAndPopulateIndexTableWithTestData() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal("failed to open sqlite database in memory")
	}
	CreateIndexTable(db)
	for i := 0; i < 444; i++ {
		AddIndexEntry(BlogIndexEntry{PostS3Loc: fmt.Sprintf("loc%d", i),
			Title: fmt.Sprintf("Great Title%d", i),
			Tags: "all, posts, have, these",
			CreatedTime:   time.Now()}, db)
	}
	return db
}

func TestGetAllIndexEntries(t *testing.T) {
	indexDb := createAndPopulateIndexTableWithTestData()
	entries, err := GetAllIndexEntries(indexDb)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		t.Log(e)
	}
}

func TestCanAddAndRetrieve(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	if err != nil {
		t.Errorf("failed to open sqlite database in memory")
	}
	CreateIndexTable(db)
	res, _ := AddIndexEntry(BlogIndexEntry{
		ID:            111111,
		PostS3Loc:     "greatbucket/key/path/here",
		Title:         "what a nice title!",
		Tags:          "tag1, tag2",
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

func TestCanRetrieveByS3Uri(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	if err != nil {
		t.Errorf("failed to open sqlite database in memory :(")
	}
	CreateIndexTable(db)
	res, _ := AddIndexEntry(BlogIndexEntry{
		ID:            111111,
		PostS3Loc:     "greatbucket/key/path/here",
		Title:         "Doesn't Matter!",
		Tags:          "cool, school",
		CreatedTime:   time.Now()}, db)
	id, err := res.LastInsertId()
	if err != nil {
		t.Errorf("failed to insert index entry")
	}
	t.Logf("added index entry %d", id)

	retrievedEntry, err := GetIndexEntryByS3Location("greatbucket/key/path/here", db)
	if err != nil {
		t.Errorf("couldn't retrieve post that i just inserted")
	}
	t.Logf("added index entry with id: %d", retrievedEntry.ID)
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
