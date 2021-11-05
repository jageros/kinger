package main

import (
	"container/list"
	"kinger/gopuppy/common/glog"
	"regexp"
	"strconv"
)

const (
	LEVEL1 = iota
	LEVEL2
	LEVEL3
)

var po = map[string]int{
	"+": LEVEL1,
	"-": LEVEL1,
	"*": LEVEL2,
	"/": LEVEL2,
}

func parseExp(s string) (exps []string, err error) {
	re, err := regexp.Compile("[0-9]+|[+*/\\-\\(\\)]{1}")
	//re, err := regexp.Compile(`^(\d+(\.\d+)?|\+|\-|\*|\/|and|or|\(|\)|==|>=|<=|!=|>|<)+$`)
	if err != nil {
		return
	}
	for _, exp := range re.FindAll([]byte(s), -1) {
		exps = append(exps, string(exp))
	}
	return
}

func isPop(list *list.List, s string) (op []string, ok bool) {
	switch string(s) {
	case "(":
		ok = false
		return
	case ")":
		ok = true
		cur := list.Back()
		for {
			prev := cur.Prev()
			if curValue, ok2 := cur.Value.(string); ok2 {
				if string(curValue) == "(" {
					list.Remove(cur)
					return
				}
				op = append(op, curValue)
				list.Remove(cur)
				cur = prev
			}
		}
	default:
		for cur := list.Back(); cur != nil; {
			prev := cur.Prev()
			if curValue, ok2 := cur.Value.(string); ok2 {
				if level, ok3 := po[curValue]; ok3 && level >= po[s] {
					ok = true
					op = append(op, curValue)
					// fmt.Println(op)
					list.Remove(cur)
				} else if curValue == "(" {
					// fmt.Println(curValue, op)
					if len(op) != 0 {
						ok = true
					} else {
						ok = false
					}
					return
				}
			}
			cur = prev
		}
	}
	return
}

func isOperate(s string) bool {
	re, _ := regexp.Compile("[+*/\\-\\(\\)]{1}")
	ok := re.MatchString(s)
	// fmt.Println(ok, s)
	return ok
}

func pre2stuf(exps []string) (exps2 []string) {
	list1 := list.New()
	list2 := list.New()

	for _, exp := range exps {
		if isOperate(exp) {
			if op, ok := isPop(list1, exp); ok {
				for _, s := range op {
					list2.PushBack(s)
				}
			}
			if exp == ")" {
				continue
			}
			list1.PushBack(exp)
		} else {
			list2.PushBack(exp)
		}
	}

	for cur := list1.Back(); cur != nil; cur = cur.Prev() {
		// fmt.Print(cur.Value)
		list2.PushBack(cur.Value)
	}

	for cur := list2.Front(); cur != nil; cur = cur.Next() {
		if curValue, ok := cur.Value.(string); ok {
			exps2 = append(exps2, curValue)
		}
	}
	return
}

func caculate(exps []string) int {
	list1 := list.New()

	for _, s := range exps {
		if isOperate(s) {
			back := list1.Back()
			prev := back.Prev()
			backVal, _ := back.Value.(int)
			prevVal := 0
			if prev != nil {
				prevVal, _ = prev.Value.(int)
			}
			var res int
			switch s {
			case "+":
				res = prevVal + backVal
			case "-":
				res = prevVal - backVal
			case "*":
				res = prevVal * backVal
			case "/":
				res = prevVal / backVal
			}
			list1.Remove(back)
			if prev != nil {
				list1.Remove(prev)
			}
			list1.PushBack(res)
		} else {
			v, _ := strconv.Atoi(s)
			list1.PushBack(v)
		}
	}
	if list1.Len() != 1 {
		glog.Errorf("caculate err")
		return 0
	}
	res, _ := list1.Back().Value.(int)
	return res
}
