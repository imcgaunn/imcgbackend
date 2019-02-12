package main

import (
	"database/sql"
	"fmt"
	"github.com/imcgaunn/imcgbackend/blog/index"
	"log"
	"testing"
	"time"

	"olympos.io/encoding/edn"
)

func TestIndexEntriesToEdn(t *testing.T) {
	conn := createAndPopulateIndexTableWithTestData()
	t.Log("transforming index entries to edn")
	entries, err := index.GetAllIndexEntries(conn)
	if err != nil {
		t.Log("could not retrieve index entries :(")
		t.Fatal(err)
	}
	ednEntries, err := edn.MarshalIndent(entries, "", "    ")
	if err != nil {
		t.Log("could not retrieve index entries :(")
		t.Fatal(err)
	}
	fmt.Println(string(ednEntries))
}

func createAndPopulateIndexTableWithTestData() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal("failed to open sqlite database in memory")
	}
	index.CreateIndexTable(db)
	for i := 0; i < 444; i++ {
		index.AddIndexEntry(index.BlogIndexEntry{PostS3Loc: fmt.Sprintf("loc%d", i),
			Title: fmt.Sprintf("great%d", i),
			Tags: "cool, nice",
			CreatedTime:   time.Now()}, db)
	}
	return db
}
