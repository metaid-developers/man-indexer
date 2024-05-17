package main

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/pebble"
)

func TestMerge(t *testing.T) {
	db, err := pebble.Open("./pebble_data/test", &pebble.Options{})
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	key := []byte("my_key")
	value1 := []byte("value1")
	value2 := []byte("value2")
	value3 := []byte("value3")
	if err := db.Merge(key, value1, pebble.Sync); err != nil {
		fmt.Println(err)
	}
	if err := db.Merge(key, value2, pebble.Sync); err != nil {
		fmt.Println(err)
	}
	if err := db.Merge(key, value3, pebble.Sync); err != nil {
		fmt.Println(err)
	}
	mergedValue, closer, err := db.Get(key)
	if err != nil && err != pebble.ErrNotFound {
		fmt.Println(err)
		return
	}
	defer closer.Close()

	fmt.Printf("key:%s,value:%s\n", key, mergedValue)

}
func TestIter(t *testing.T) {
	db, err := pebble.Open("./pebble_data/test", &pebble.Options{})
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	for i := 1; i <= 20; i++ {
		key := []byte(fmt.Sprintf("%d", i))
		value := []byte(fmt.Sprintf("value%d", i))
		if err := db.Set(key, value, pebble.Sync); err != nil {
			return
		}
	}
	iter, _ := db.NewIter(nil)
	defer iter.Close()
	iter.Last()
	fmt.Println(string(iter.Key()))

}
