package main

import (
	"fmt"
	"log"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

func main() {

	// Open the default database from the system cluster
	db := fdb.MustOpenDefault()

	_, err := db.Transact(func(tr fdb.Transaction) (ret interface{}, e error) {
		tr.Set(fdb.Key("hello"), []byte("world"))
		return
	})
	if err != nil {
		log.Fatalf("Unable to set FDB database value (%v)", err)
	}

	ret, err := db.Transact(func(tr fdb.Transaction) (ret interface{}, e error) {
		ret = tr.Get(fdb.Key("hello")).MustGet()
		return
	})
	if err != nil {
		log.Fatalf("Unable to read FDB database value (%v)", err)
	}

	v := ret.([]byte)
	fmt.Printf("hello, %s\n", string(v))
}
