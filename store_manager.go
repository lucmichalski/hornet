package main

import (
	"regexp"
)

const (
	IDX_MEM = iota
	IDX_SSD
	IDX_HDD
)

type StoreManager struct {
	stores [3]*Store
}

func NewStoreManager() (sm *StoreManager) {
	sm = new(StoreManager)

	sm.stores[IDX_MEM] = NewStore(
		"mem",
		GConfig["cache.mem.meta"].(string),
		GConfig["cache.mem.path"].(string),
		GConfig["cache.mem.cap"].(int),
		GConfig["cache.mem.blocksize"].(int),
	)

	sm.stores[IDX_SSD] = NewStore(
		"ssd",
		GConfig["cache.ssd.meta"].(string),
		GConfig["cache.ssd.path"].(string),
		GConfig["cache.ssd.cap"].(int),
		GConfig["cache.ssd.blocksize"].(int),
	)

	sm.stores[IDX_HDD] = NewStore(
		"hdd",
		GConfig["cache.hdd.meta"].(string),
		GConfig["cache.hdd.path"].(string),
		GConfig["cache.hdd.cap"].(int),
		GConfig["cache.hdd.blocksize"].(int),
	)

	return sm
}

func (sm *StoreManager) Close() {
	for _, s := range sm.stores {
		if s != nil {
			s.Close()
		}
	}
}

func (sm *StoreManager) Add(item *Item) []byte {
	for i := IDX_HDD; i > 0; i-- {
		if sm.stores[i] != nil {
			return sm.stores[i].Add(item)
		}
	}
	return nil
}

func (sm *StoreManager) Get(id Key) (*Item, []byte) {
	for _, s := range sm.stores {
		if item, data := s.Get(id); item != nil {
			return item, data
		}
	}
	return nil, nil
}

func (sm *StoreManager) Del(id Key) {
	for _, s := range sm.stores {
		s.Delete(id)
	}
}

func (sm *StoreManager) DelByGroup(g HKey) {
	sm.delBatch(func(item *Item) bool {
		return item.Info.Grp == g
	})
}

func (sm *StoreManager) DelByRawKey(reg *regexp.Regexp) {
	sm.delBatch(func(item *Item) bool {
		return reg.Match(item.Info.RawKey[:item.Info.RawKeyLen])
	})
}

func (sm *StoreManager) delBatch(match func(*Item) bool) {
	for _, s := range sm.stores {
		s.DeleteBatch(match)
	}
}

//TODO delete by tag
//TODO delete by mask