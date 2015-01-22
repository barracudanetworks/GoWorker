package database

import "testing"

func TestOpen(t *testing.T) {
	db, err := Open("test.db")
	if err != nil {
		t.Error(err)
	}

	if db == nil {
		t.Fail()
	}
	db.Close()
}

func TestOpenMultiple(t *testing.T) {
	db1, err1 := Open("test.db")
	if err1 != nil {
		t.Error(err1)
	}
	db2, err2 := Open("test.db")

	if err2 != nil {
		t.Error(err2)
	}

	if db1 != db2 {
		t.Fail()
	}
}

func TestClose(t *testing.T) {
	db, err := Open("test.db")
	if err != nil {
		t.Error(err)
	}

	err = Close(db)
	if err != nil {
		t.Error(err)
	}
}
