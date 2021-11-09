package wordfilter

import (
	"fmt"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/rpubsub"
	"strings"
	"unicode/utf8"
)

var (
	filter         *wordFilter
	replaceRune, _ = utf8.DecodeRuneInString("*")
)

type treeNode struct {
	data       map[int32]*treeNode
	isEnd      bool // 是否是敏感词的词尾字，敏感词树的叶子节点必然是词尾字，父节点不一定是
	isGeneral  bool //是否是普通敏感词
	isFuzzy    bool //是否是模糊敏感词
	isAccurate bool //是否是精确敏感词
	parent     *treeNode
	value      int32
}

func newTreeNode() *treeNode {
	return &treeNode{
		data: map[int32]*treeNode{},
	}
}

func (tn *treeNode) getChild(c int32) *treeNode {
	return tn.data[c]
}

func (tn *treeNode) addChild(c int32) *treeNode {
	n := newTreeNode()
	tn.data[c] = n
	n.value = c
	n.parent = tn
	return n
}

type wordFilter struct {
	treeRoot *treeNode
}

func initFilter(loadWordFunc func() map[int][]string) {
	wordsList := loadWordFunc()
	_filter := &wordFilter{}
	_filter.treeRoot = newTreeNode()
	wordTypes := []int{consts.GeneralWords, consts.FuzzyWords, consts.AccurateWords}
	for _, wordType := range wordTypes {
		if words, ok := wordsList[wordType]; ok {
			if words != nil && len(words) > 0 {
				for _, word := range words {
					currentBranch := _filter.treeRoot
					for _, char := range word {
						tmp := currentBranch.getChild(char)
						if tmp != nil {
							currentBranch = tmp
						} else {
							currentBranch = currentBranch.addChild(char)
						}
					}
					currentBranch.isEnd = true
					switch wordType {
					case consts.GeneralWords:
						currentBranch.isGeneral = true
					case consts.FuzzyWords:
						currentBranch.isFuzzy = true
					case consts.AccurateWords:
						currentBranch.isAccurate = true
					}
				}
			}
		}
	}
	filter = _filter

	return
}

func RegisterWords(loadWordFunc func() map[int][]string) {
	rpubsub.SubscribeWithOption("reload_word", func(i map[string]interface{}) {
		initFilter(loadWordFunc)
	}, true)

	initFilter(loadWordFunc)
}

// 判断是否包含敏感词
func ContainsDirtyWords(word string, needReplace bool) (newWord string, hasDirty bool, dirtyWords string, wordsType int) {
	newWord = word
	var newWordRune []rune
	if needReplace {
		newWordRune = []rune(word)
	}
	if filter == nil {
		return
	}

	curTree := filter.treeRoot
	var childTree *treeNode
	headIndex := -1 // 敏感词词首索引
	chars := []rune(word)
	var i int
	n := len(chars)
	var dirtyWordLen int
	for i < n {
		//chars[i] =
		char := chars[i]
		childTree = curTree.getChild(char)
		if childTree != nil {
			dirtyWordLen++
			if headIndex == -1 {
				headIndex = i
			}

			if childTree.isEnd {
				// 检查到一个敏感词
				hasDirty = true

				if childTree.isGeneral {
					wordsType = consts.GeneralWords
				}
				if childTree.isFuzzy {
					wordsType = consts.FuzzyWords
				}
				if childTree.isAccurate {
					wordsType = consts.AccurateWords
				}

				if !needReplace {
					return
				}

				// 替换敏感词s
				var dw string
				for j := 0; j < dirtyWordLen; j++ {
					dw = fmt.Sprintf("%s%c", dw, newWordRune[headIndex+j])
					newWordRune[headIndex+j] = replaceRune
				}
				dw = strings.Replace(dw, " ", "", -1)
				if dw != "" {
					dirtyWords = fmt.Sprintf("%s %s", dirtyWords, dw)
				}

				dirtyWordLen = 0
				headIndex = -1
				curTree = filter.treeRoot
				i++
				continue

			} else {
				curTree = childTree
			}

		} else {

			if curTree != filter.treeRoot {
				//如果之前有遍历到敏感词非词尾，匹配部分未完全匹配，则设置循环索引为敏感词词首索引
				i = headIndex
				headIndex = -1
			}
			dirtyWordLen = 0
			curTree = filter.treeRoot
		}

		i++
	}

	if needReplace {
		newWord = string(newWordRune)
	}

	return
}

func getReplaceStr(length int) string {
	result := strings.Builder{}
	for i := 0; i < length; i++ {
		result.WriteString("*")
	}
	return result.String()
}

func UpdateDirtyWords(words []string, isAccurate bool, isAdd bool) {

	if words == nil && len(words) <= 0 {
		return
	}

	if isAdd {
		for _, word := range words {
			if word == "" {
				continue
			}
			currentBranch := filter.treeRoot
			for _, char := range word {
				tmp := currentBranch.getChild(char)
				if tmp != nil {
					currentBranch = tmp
				} else {
					currentBranch = currentBranch.addChild(char)
				}
			}
			currentBranch.isEnd = true
			if isAccurate {
				currentBranch.isAccurate = true
			} else {
				currentBranch.isFuzzy = true
			}
		}
	} else {
	outFor:
		for _, word := range words {
			if word == "" {
				continue
			}
			currentBranch := filter.treeRoot
			//secondFor:
			for _, char := range word {
				tmp := currentBranch.getChild(char)
				if tmp != nil {
					currentBranch = tmp
				} else {
					continue outFor
				}
			}
			if isAccurate {
				currentBranch.isAccurate = false
			} else {
				currentBranch.isFuzzy = false
			}

			if !currentBranch.isAccurate && !currentBranch.isFuzzy && !currentBranch.isGeneral {
				currentBranch.isEnd = false
			}

		}
	}

}
