package main

import (
	"unsafe"

	llrb "github.com/petar/GoLLRB/llrb"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/proto/pb"
)

func nextLargerKey(key string) string {
	return key + "\x00" // the next string that is larger than key, but smaller than any other keys > key
}

type filterTree struct {
	btree *llrb.LLRB
}

func newFilterTree() *filterTree {
	return &filterTree{
		btree: llrb.New(),
	}
}

type filterTreeItem struct {
	cp  *clientProxy
	val string
}

func (it *filterTreeItem) Less(_other llrb.Item) bool {
	other := _other.(*filterTreeItem)
	return it.val < other.val || (it.val == other.val && uintptr(unsafe.Pointer(it.cp)) < uintptr(unsafe.Pointer(other.cp)))
}

func (ft *filterTree) Insert(cp *clientProxy, val string) {
	ft.btree.ReplaceOrInsert(&filterTreeItem{
		cp:  cp,
		val: val,
	})
}

func (ft *filterTree) Remove(cp *clientProxy, val string) {
	ft.btree.Delete(&filterTreeItem{
		cp:  cp,
		val: val,
	})
}

func (ft *filterTree) Visit(op pb.BroadcastClientFilter_OpType, val string, f func(cp *clientProxy)) {
	if op == pb.BroadcastClientFilter_EQ {
		// visit key == val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			item := _item.(*filterTreeItem)
			if item.val > val {
				return false
			}

			f(item.cp)
			return true
		})
	} else if op == pb.BroadcastClientFilter_NE {
		// visit key != val
		// visit key < val first
		ft.btree.AscendLessThan(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
		// then visit key > val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, nextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == pb.BroadcastClientFilter_GT {
		// visit key > val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, nextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == pb.BroadcastClientFilter_GTE {
		// visit key >= val
		ft.btree.AscendGreaterOrEqual(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == pb.BroadcastClientFilter_LT {
		// visit key < val
		ft.btree.AscendLessThan(&filterTreeItem{nil, val}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else if op == pb.BroadcastClientFilter_LTE {
		// visit key <= val
		ft.btree.AscendLessThan(&filterTreeItem{nil, nextLargerKey(val)}, func(_item llrb.Item) bool {
			f(_item.(*filterTreeItem).cp)
			return true
		})
	} else {
		glog.Panicf("unknown filter clients op: %s", op)
	}
}
