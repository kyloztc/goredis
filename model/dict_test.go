package model

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"testing"
)

type testDictType struct {
}

func (t testDictType) HashFunction(key interface{}) int64 {
	_key := key.(string)
	bytes := []byte(_key)
	return int64(crc32.ChecksumIEEE(bytes))
}

func TestDictAdd(t *testing.T) {
	d := DictCreate(testDictType{})
	d.DictAdd("key1", "hello1")
	d.DictAdd("key2", "hello2")
	d.DictAdd("key3", "hello3")
	d.DictAdd("key4", "hello4")
	d.DictAdd("key5", "hello5")
	byteData, _ := json.Marshal(d)
	fmt.Printf("dict: %v\n", string(byteData))
	val := d.DictFetchValue("key3")
	fmt.Printf("%v\n", val)
}

func TestRehash(t *testing.T) {
	d := DictCreate(testDictType{})
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("hello%d", i)
		d.DictAdd(key, value)
		if d.dictIsRehashing() {
			fmt.Printf("%s %s\n", key, value)
			fmt.Printf("ht0 used: %v|ht1 used: %v\n", d.Ht[0].Used, d.Ht[1].Used)
		}
		byteData, _ := json.Marshal(d)
		fmt.Printf("dict: %v\n", string(byteData))
	}

}
