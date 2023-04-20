package model

import "math"

type DictStatus int

const (
	DictOk  = DictStatus(0)
	DictErr = DictStatus(1)
)

var (
	DictHtInitialSize    = int64(4)
	DictCanResize        = 1
	DictForceResizeRatio = 5
)

type DictType interface {
	HashFunction(key interface{}) int64
}

type dictEntry struct {
	Key  interface{}
	Val  interface{}
	Next *dictEntry
}

type dictHt struct {
	Table    []*dictEntry
	Size     int64
	SizeMask int64
	Used     int64
}

type Dict struct {
	DictType  DictType
	Ht        []*dictHt
	RehashIdx int8
	iterators int64
}

func DictCreate(dictType DictType) *Dict {
	DictCanResize = 1
	return &Dict{
		DictType:  dictType,
		Ht:        []*dictHt{new(dictHt), new(dictHt)},
		RehashIdx: -1,
	}
}

func (d *Dict) DictAdd(key interface{}, val interface{}) DictStatus {
	entry := d.dictAddRaw(key, nil)
	if entry == nil {
		return DictErr
	}
	entry.dictSetVal(val)
	return DictOk
}

func (d *Dict) dictAddRaw(key interface{}, existing *dictEntry) *dictEntry {
	var index int64
	var entry *dictEntry
	var ht *dictHt
	if d.dictIsRehashing() {
		d.dictRehashStep()
	}
	if index = d.dictKeyIndex(key, d.DictType.HashFunction(key), existing); index == -1 {
		return nil
	}
	ht = d.Ht[0]
	if d.dictIsRehashing() {
		ht = d.Ht[1]
	}
	entry = new(dictEntry)
	entry.Next = ht.Table[index]
	ht.Table[index] = entry
	ht.Used++

	entry.dictSetKey(key)
	return entry
}

func (e *dictEntry) dictSetKey(key interface{}) {
	e.Key = key
}

func (e *dictEntry) dictSetVal(val interface{}) {
	e.Val = val
}

func (d *Dict) dictIsRehashing() bool {
	return d.RehashIdx != -1
}

func (d *Dict) dictRehashStep() {
	if d.iterators == 0 {
		d.dictRehash(1)
	}
}

func (d *Dict) dictRehash(n int) int {
	// 防止等待太久
	emptyVisits := n * 10
	if !d.dictIsRehashing() {
		return 0
	}
	for n != 0 && d.Ht[0].Used != 0 {
		n--
		de := new(dictEntry)
		nextDe := new(dictEntry)
		for d.Ht[0].Table[d.RehashIdx] == nil {
			d.RehashIdx++
			emptyVisits--
			if emptyVisits == 0 {
				return 1
			}
		}
		de = d.Ht[0].Table[d.RehashIdx]
		for de != nil {
			h := d.DictType.HashFunction(de.Key) & d.Ht[1].SizeMask
			nextDe = de.Next
			de.Next = d.Ht[1].Table[h]
			d.Ht[1].Table[h] = de
			d.Ht[0].Used--
			d.Ht[1].Used++
			de = nextDe
		}
		d.Ht[0].Table[d.RehashIdx] = nil
		d.RehashIdx++
	}

	if d.Ht[0].Used == 0 {
		d.Ht[0] = d.Ht[1]
		d.Ht[1] = new(dictHt)
		d.RehashIdx = -1
		return 0
	}
	return 1
}

func (d *dictHt) dictReset() {
	d.Table = nil
	d.Size = 0
	d.SizeMask = 0
	d.Used = 0
}

func (d *Dict) dictKeyIndex(key interface{}, hash int64, existing *dictEntry) int64 {
	var idx, table int64
	he := new(dictEntry)
	if d.dictExpandIfNeeded() == DictErr {
		return -1
	}
	for table = 0; table <= 1; table++ {
		idx = hash & d.Ht[table].SizeMask
		he = d.Ht[table].Table[idx]
		for he != nil {
			if key == he.Key {
				existing = he
				return -1
			}
			he = he.Next
		}
		if !d.dictIsRehashing() {
			break
		}
	}
	return idx
}

func (d *Dict) dictExpandIfNeeded() DictStatus {
	// 若正在扩容
	if d.dictIsRehashing() {
		return DictOk
	}
	if d.Ht[0].Size == 0 {
		return d.dictExpand(DictHtInitialSize)
	}

	if d.Ht[0].Used >= d.Ht[0].Size && (DictCanResize == 1 || int(d.Ht[0].Used/d.Ht[0].Size) > DictForceResizeRatio) {
		return d.dictExpand(d.Ht[0].Used * 2)
	}
	return DictOk
}

func (d *Dict) dictExpand(size int64) DictStatus {
	if d.dictIsRehashing() || d.Ht[0].Used > size {
		return DictErr
	}
	n := new(dictHt)
	realSize := dictNextPower(size)

	// 扩展到原来的原来的大小没有意义
	if realSize == d.Ht[0].Size {
		return DictErr
	}

	n.Size = realSize
	n.SizeMask = realSize - 1
	n.Table = make([]*dictEntry, realSize)
	n.Used = 0

	// 如果是初始化
	if d.Ht[0].Table == nil {
		d.Ht[0] = n
		return DictOk
	}

	// 否则则准备rehash
	d.Ht[1] = n
	d.RehashIdx = 0
	return DictOk
}

// 哈希表大小为2的幂次方
func dictNextPower(size int64) int64 {
	i := DictHtInitialSize
	if size >= math.MaxInt64 {
		return math.MaxInt64
	}
	for {
		if i >= size {
			return i
		}
		i *= 2
	}
}

func (d *Dict) DictFetchValue(key interface{}) interface{} {
	he := d.dictFind(key)
	if he == nil {
		return nil
	}
	return he.dictGetValue()
}

func (e *dictEntry) dictGetValue() interface{} {
	return e.Val
}

func (d *Dict) dictFind(key interface{}) *dictEntry {
	he := new(dictEntry)
	var h, idx, table int64
	if d.dictSize() == 0 {
		return nil
	}
	if d.dictIsRehashing() {
		d.dictRehashStep()
	}

	h = d.DictType.HashFunction(key)
	for table = 0; table <= 1; table++ {
		idx = h & d.Ht[table].SizeMask
		he = d.Ht[table].Table[idx]
		for he != nil {
			if key == he.Key {
				return he
			}
		}
		if !d.dictIsRehashing() {
			return nil
		}
	}
	return nil
}

func (d *Dict) dictSize() int64 {
	return d.Ht[0].Used + d.Ht[1].Used
}
