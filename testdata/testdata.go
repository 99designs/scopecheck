package testdata

import (
	"fmt"
	"testing"

	"github.com/pressly/chi"
)

// Fails
func _() {
	r1 := chi.NewRouter()

	r1.Group(func(r2 chi.Router) {
		r1.Use(nil)
	})
}

// Fails
func _() {
	var t1 *testing.T

	t1.Run("foo", func(t2 *testing.T) {
		t1.Fail()
	})
}

// Passes
func _() {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(nil)
	})
}

// Passes
func _() {
	var t *testing.T

	t.Run("foo", func(t *testing.T) {
		t.Fail()
	})
}

// Passes
func _() {
	s := "foo"

	func(e interface{}) {
		fmt.Println(e)
	}(s)
}

// Passes
func _() {
	s := "foo"

	func(s interface{}) {
		fmt.Println(s)
	}(s)
}

// Passes
func _() {
	var i int = 1

	func(j int) {
		fmt.Println(i, j)
	}(i)
}

type DB interface {
	Query(x, y int) bool
	InTX(func(TX))
}

type TX interface {
	DB

	Commit(s string) int
}

// Fails
func _() {
	var db DB

	db.InTX(func(tx TX) {
		db.Query(1, 2)
	})
}

// Passes
func _() {
	var db DB

	db.InTX(func(tx TX) {
		tx.Query(1, 2)
	})
}
