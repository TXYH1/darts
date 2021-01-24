package godarts

import (
	"fmt"
	"sort"
)

const RESIZE_DELTA = 64
const END_NODE_BASE = -1
const ROOT_NODE_BASE = 1
const ROOT_NODE_INDEX = 0

//Linked List Trie
type LinkedListTrieNode struct {
	Code                            rune
	Depth, Left, Right, Index, Base int // Index base 主要用于ac自动机，可基于此实现
	SubKey                          []rune
	Children                        [](*LinkedListTrieNode)
}

type LinkedListTrie struct {
	Root *LinkedListTrieNode
}

//Double Array Trie
type DoubleArrayTrie struct {
	Base  []int
	Check []int
}

type dartsKey []rune
type dartsKeySlice []dartsKey

type Darts struct {
	dat          *DoubleArrayTrie
	llt          *LinkedListTrie
	used         []bool
	nextCheckPos int
	key          dartsKeySlice
	Output       map[int]([]rune)
}

func (k dartsKeySlice) Len() int {
	return len(k)
}

func (k dartsKeySlice) Less(i, j int) bool {
	iKey, jKey := k[i], k[j]
	iLen, jLen := len(iKey), len(jKey)

	var pos int = 0
	for {
		if pos < iLen && pos < jLen {
			if iKey[pos] < jKey[pos] {
				return true
			} else if iKey[pos] > jKey[pos] {
				return false
			}
		} else {
			if iLen < jLen {
				return true
			} else {
				return false
			}
		}
		pos++
	}

	return false
}

func (k dartsKeySlice) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func (llt *LinkedListTrie) printTrie(n *LinkedListTrieNode) {
	for i := 0; i < n.Depth; i++ {
		fmt.Printf("\t")
	}
	for _, c := range n.Children {
		llt.printTrie(c)
	}
}

func (llt *LinkedListTrie) PrintTrie() {
	llt.printTrie(llt.Root)
}

func (dat *DoubleArrayTrie) PrintTrie() {
	fmt.Printf("+-----+-----+-----+\n")
	fmt.Printf("|%5s|%5s|%5s|\n", "id", "base", "check")
	for idx, _ := range dat.Base {
		if dat.Base[idx] == 0 && dat.Check[idx] == 0 {
			continue
		}
		fmt.Printf("+-----+-----+-----+\n")
		fmt.Printf("|%5d|%5d|%5d|\n", idx, dat.Base[idx], dat.Check[idx])
	}
	fmt.Printf("+-----+-----+-----+\n")
}

func (d *Darts) Build(keywords [][]rune) (*DoubleArrayTrie, *LinkedListTrie, error) {
	if len(keywords) == 0 {
		return nil, nil, fmt.Errorf("empty keywords")
	}

	d.dat = new(DoubleArrayTrie)
	d.resize(RESIZE_DELTA)

	for _, keyword := range keywords {
		var dk dartsKey = keyword
		d.key = append(d.key, dk)
	}
	sort.Sort(d.key)

	d.Output = make(map[int]([]rune), len(d.key))
	d.dat.Base[0] = ROOT_NODE_BASE
	d.nextCheckPos = 0

	d.llt = new(LinkedListTrie)
	d.llt.Root = new(LinkedListTrieNode)
	d.llt.Root.Depth = 0
	d.llt.Root.Left = 0
	d.llt.Root.Right = len(keywords)
	d.llt.Root.SubKey = nil
	d.llt.Root.Index = ROOT_NODE_INDEX

	siblings, err := d.fetch(d.llt.Root)
	if err != nil {
		return nil, nil, err
	}
	for idx, ns := range siblings {
		if ns.Code > 0 {
			siblings[idx].SubKey = append(d.llt.Root.SubKey, ns.Code-ROOT_NODE_BASE)
		}
	}

	_, err = d.insert(siblings)
	if err != nil {
		return nil, nil, err
	}

	return d.dat, d.llt, nil
}

func (d *Darts) resize(size int) {
	d.dat.Base = append(d.dat.Base, make([]int, (size-len(d.dat.Base)))...)
	d.dat.Check = append(d.dat.Check, make([]int, (size-len(d.dat.Check)))...)

	d.used = append(d.used, make([]bool, (size-len(d.used)))...)
}

func (d *Darts) fetch(parent *LinkedListTrieNode) (siblings [](*LinkedListTrieNode), err error) {
	siblings = make([](*LinkedListTrieNode), 0, 2)

	var prev rune = 0

	for i := parent.Left; i < parent.Right; i++ {

		if len(d.key[i]) < parent.Depth {
			continue
		}

		tmp := d.key[i]

		var cur rune = 0
		if len(d.key[i]) != parent.Depth {
			cur = tmp[parent.Depth] + 1
		}

		if prev > cur {
			return nil, fmt.Errorf("fetch error")
		}

		if cur != prev || len(siblings) == 0 {
			var subKey []rune
			if cur != 0 {
				subKey = append(parent.SubKey, cur-ROOT_NODE_BASE)
			} else {
				subKey = parent.SubKey
			}

			tmpNode := new(LinkedListTrieNode)
			tmpNode.Depth = parent.Depth + 1
			tmpNode.Code = cur
			tmpNode.Left = i
			tmpNode.SubKey = make([]rune, len(subKey)) // 到当前节点的所有前缀
			copy(tmpNode.SubKey, subKey)
			// 这里根本没必要用siblings和Children两个，这两个是同一个东西，这里操作两次是没必要的
			if len(siblings) != 0 {
				siblings[len(siblings)-1].Right = i
			}
			siblings = append(siblings, tmpNode)
			if len(parent.Children) != 0 {
				parent.Children[len(parent.Children)-1].Right = i
			}
			parent.Children = append(parent.Children, tmpNode)
		}

		prev = cur
	}

	if len(siblings) != 0 {
		siblings[len(siblings)-1].Right = parent.Right
	}
	if len(parent.Children) != 0 {
		parent.Children[len(siblings)-1].Right = parent.Right
	}

	//return siblings, nil
	return parent.Children, nil
}

// 把 begin + code看作一个字符处于base数组的索引
func (d *Darts) insert(siblings [](*LinkedListTrieNode)) (int, error) {
	var begin int = 0
	var pos int = max(int(siblings[0].Code)+1, d.nextCheckPos) - 1
	var nonZeroNum int = 0
	var first bool = false

	if len(d.dat.Base) <= pos {
		d.resize(pos + 1)
	}

	for {
	next:
		pos++

		if len(d.dat.Base) <= pos {
			d.resize(pos + 1)
		}

		if d.dat.Check[pos] > 0 {
			nonZeroNum++
			continue
		} else if !first {
			d.nextCheckPos = pos
			first = true
		}
		// 这个begin的取值还是很难理解，可能是为了防止频繁冲突？增大随机性？
		begin = pos - int(siblings[0].Code)
		if len(d.dat.Base) <= (begin + int(siblings[len(siblings)-1].Code)) {
			d.resize(begin + int(siblings[len(siblings)-1].Code) + RESIZE_DELTA)
		}

		if d.used[begin] {
			continue
		}

		for i := 1; i < len(siblings); i++ {
			if 0 != d.dat.Check[begin+int(siblings[i].Code)] {
				goto next
			}
		}
		break

	}

	if float32(nonZeroNum)/float32(pos-d.nextCheckPos+1) >= 0.95 {
		d.nextCheckPos = pos
	}
	d.used[begin] = true

	for i := 0; i < len(siblings); i++ {
		d.dat.Check[begin+int(siblings[i].Code)] = begin
	}

	for i := 0; i < len(siblings); i++ {
		newSiblings, err := d.fetch(siblings[i])
		if err != nil {
			return -1, err
		}

		if len(newSiblings) == 0 { // 字符到了结尾，还是能fetch到一个空字符的，即code为0，空字符fetch时才没东西，此时才算是结束
			d.dat.Base[begin+int(siblings[i].Code)] = -siblings[i].Left - 1
			d.Output[begin+int(siblings[i].Code)] = siblings[i].SubKey // 到结尾了，可以保存一个完整的单词(词语)
			siblings[i].Base = END_NODE_BASE
			siblings[i].Index = begin + int(siblings[i].Code)
		} else {
			h, err := d.insert(newSiblings)

			if err != nil {
				return -1, err
			}
			d.dat.Base[begin+int(siblings[i].Code)] = h // 设置当前状态的base值为 使得所有子节点d.dat.Check[begin+int(siblings[i].Code)]==0成立的begin值
			siblings[i].Index = begin + int(siblings[i].Code)
			siblings[i].Base = h
		}
	}

	return begin, nil
}

func (dat *DoubleArrayTrie) ExactMatchSearch(content []rune, nodePos int) bool {
	b := dat.Base[nodePos]
	var p int

	for _, r := range content {
		p = b + int(r) + 1
		if b == dat.Check[p] {
			b = dat.Base[p]
		} else {
			return false
		}
	}

	p = b
	n := dat.Base[p]
	if b == dat.Check[p] && n < 0 {
		return true
	}

	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
