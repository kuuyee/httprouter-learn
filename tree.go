package httprouter

import (
	"fmt"
	_ "strings"
)

// 比较两数值，返回小的
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// countParams 统计参数个数
func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] != ':' && path[i] != '*' {
			continue
		}
		n++
	}
	return uint8(n)
}

type nodeType uint8

const (
	static   nodeType = iota //默认
	root                     //根
	param                    //带请求参数
	catchAll                 //全部
)

type node struct {
	path      string
	wildChild bool
	nType     nodeType
	maxParams uint8
	indices   string
	children  []*node
	handle    Handle
	priority  uint32
}

func (n *node) perttyNode() {
	fmt.Printf("%11s = %-s\n ", "path", n.path)
	fmt.Printf("%10s = %-t\n ", "wildChild", n.wildChild)
	fmt.Printf("%10s = %-d\n ", "nType", n.nType)
	fmt.Printf("%10s = %-d\n ", "maxParams", n.maxParams)
	fmt.Printf("%10s = %-s\n ", "indices", n.indices)
	fmt.Printf("%10s = %-d\n ", "children", len(n.children))
	if len(n.children) > 0 {
		n.perttyNode2()
	}
	fmt.Printf("%10s = %-v\n ", "handle", n.handle)
	fmt.Printf("%10s = %-d\n ", "priority", n.priority)
	fmt.Printf("\n\n")
}

func (n *node) perttyNode2() {
	for _, v := range n.children {
		fmt.Printf("%21s = %-s\n ", "path", v.path)
		fmt.Printf("%20s = %-t\n ", "wildChild", v.wildChild)
		fmt.Printf("%20s = %-d\n ", "nType", v.nType)
		fmt.Printf("%20s = %-d\n ", "maxParams", v.maxParams)
		fmt.Printf("%20s = %-s\n ", "indices", v.indices)
		fmt.Printf("%20s = %-d\n ", "children", len(v.children))
		fmt.Printf("%20s = %-v\n ", "handle", v.handle)
		fmt.Printf("%20s = %-d\n ", "priority", n.priority)
		fmt.Printf("\n")
	}
}

// increments priority of the given child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {
	n.children[pos].priority++
	prio := n.children[pos].priority

	// adjust position(move to front)
	newPos := pos
	for newPos > 0 && n.children[newPos-1].priority < prio {
		// swap node positions
		tmpN := n.children[newPos-1]
		n.children[newPos-1] = n.children[newPos]
		n.children[newPos] = tmpN

		newPos--
	}

	// build new index char string
	if newPos != pos {
		n.indices = n.indices[:newPos] + // unchanged prefix, might be empty
			n.indices[pos:pos+1] + // the index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // rest without char at 'pos'
	}

	return newPos
}

// addRoute 根据给定的处理路径添加一个node。
// 注意，这不是线程安全的！
func (n *node) addRoute(path string, handle Handle) {
	var tempnode *node
	fullPath := path
	n.priority++ //默认为1起步

	//统计参数个数
	numParams := countParams(path)
	fmt.Printf("path=%s | numParams : %d\n", path, numParams)

	// non-empty tree
	if len(n.path) > 0 || len(n.children) > 0 {
		fmt.Printf("n.path: %s | n.children: %d | n.maxParams: %d\n", n.path, len(n.children), n.maxParams)
	walk:
		for {
			// 更新当前node的maxParams
			if numParams > n.maxParams {
				n.maxParams = numParams
			}
			fmt.Printf("更新当前的maxParams为：%d\n", n.maxParams)

			// 计算公共前缀(Radix tree前缀数算法)
			// 这也意味着，公共前缀不含'：'或'*'，因为现有的key不能包含那些字符。
			// n.path=/cmd/veta
			// path= /cmd/vetb/:sub
			i := 0
			max := min(len(path), len(n.path)) //这里max的值是/cmd/veta的长度9
			for i < max && path[i] == n.path[i] {
				i++ //计算出前缀索引值 注释中的例子计算出的前缀是 /cmd/vet
			}
			fmt.Printf("i=%d | max=%d\n", i, max)

			// 边界切分
			if i < len(n.path) { //如果条件成立，表示存在前缀，这里是//cmd/vet
				// 创建子节点
				child := node{
					path:      n.path[i:],
					wildChild: n.wildChild,
					indices:   n.indices,
					children:  n.children,
					handle:    n.handle,
					priority:  n.priority - 1,
				}

				// 更新maxParams(max of all children),如果某个子节点maxParams更大，那么更新
				for i := range child.children {
					if child.children[i].maxParams > child.maxParams {
						child.maxParams = child.children[i].maxParams
					}
				}
				//在n.children下加入新切分的child
				n.children = []*node{&child}
				// []byte for proper unicode char conversion, see #65
				n.indices = string([]byte{n.path[i]})
				n.path = path[:i] //设置为新计算的前缀值/s
				n.handle = nil
				n.wildChild = false

				child.perttyNode()
				n.perttyNode()
			}
			fmt.Println("[Kuuyee]===.3")
			// 弄个新节点，是这个节点的子节点
			if i < len(path) {
				path = path[i:]
				fmt.Println("[Kuuyee]===.4 path=%s", path)
				if n.wildChild {
					n = n.children[0]
					n.priority++

					// 更新子节点的maxParams
					if numParams > n.maxParams {
						n.maxParams = numParams
					}
					numParams--

					// Check if the wildcard matches
					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						// check for longer wildcard, e.g. :name and :names
						if len(n.path) >= len(path) || path[len(n.path)] == '/' {
							continue walk
						}
					}
					fmt.Println("path segment '" + path +
						"' conflicts with existing wildcard '" + n.path +
						"' in path '" + fullPath + "'")
					panic("path segment '" + path +
						"' conflicts with existing wildcard '" + n.path +
						"' in path '" + fullPath + "'")
				}

				c := path[0]
				fmt.Printf("[Kuuyee]===.5 path=%s c=%d\n", path, c)

				// slash after param
				if n.nType == param && c == '/' && len(n.children) == 1 {
					n = n.children[0]
					n.priority++
					fmt.Println("[Kuuyee]===.6")
					continue walk
				}
				fmt.Println("[Kuuyee]===.7")

				// Check if a child with the next path byte exists
				for i := 0; i < len(n.indices); i++ {
					if c == n.indices[i] {
						i = n.incrementChildPrio(i)
						n = n.children[i]
						fmt.Println("[Kuuyee]===.8")
						continue walk
					}
				}

				fmt.Println("[Kuuyee]===.9")
				// 否则插入
				if c != ':' && c != '*' {
					// []byte for proper unicode char conversion, see #65
					n.indices += string([]byte{c})
					child := &node{
						maxParams: numParams,
					}
					n.children = append(n.children, child)
					n.incrementChildPrio(len(n.indices) - 1)
					tempnode = n

					n = child
					fmt.Println("[Kuuyee]===.10")
				}
				fmt.Println("[Kuuyee]===.11")
				n.insertChild(numParams, path, fullPath, handle)
				n.perttyNode()
				fmt.Println("[Kuuyee]===.1")

				tempnode.perttyNode()
				return
			} else if i == len(path) { // Make node a (in-path) leaf
				if n.handle != nil {
					fmt.Println("a handle is already registered for path '" + fullPath + "'")
					panic("a handle is already registered for path '" + fullPath + "'")
				}
				n.handle = handle
			}
			fmt.Println("[Kuuyee]===.2")
			n.perttyNode()
			return
		}
	} else { // tree为空
		fmt.Printf("tree为空 numParams: %d | path: %s | fullPath: %s | handle: %v\n", numParams, path, fullPath, handle)
		n.insertChild(numParams, path, fullPath, handle)
		n.nType = root
		n.perttyNode()
	}

}
func (n *node) insertChild(numParams uint8, path, fullPath string, handle Handle) {
	// already handled bytes of the path
	// 表示一个偏移量，处理完参数的剩余部分
	var offset int

	// numParams>0 表示有参数
	// 查找前缀直到遇到第一个通配符(以':'或 '*'开头)
	fmt.Println("[Kuuyee] insertChild===.0")
	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}
		fmt.Println("[Kuuyee] insertChild===.1")
		// 找到通配符结尾(例如'/'或路径尾部)
		end := i + 1
		for end < max && path[end] != '/' {
			switch path[end] {
			// 通配符名字不能包含':'和'*'
			case ':', '*':
				fmt.Println("每段路径只能允许有一个通配符，：'" +
					path[i:] + "' in path '" + fullPath + "'")
				panic("每段路径只能允许有一个通配符，：'" +
					path[i:] + "' in path '" + fullPath + "'")
			default:
				end++
			}
		}
		fmt.Println("[Kuuyee] insertChild===.2")
		// 如果我们在这里插入通配符，那么node的子将不能被查找到
		if len(n.children) > 0 {
			fmt.Println("通配符路由 '" + path[i:end] +
				"' 和存在的children冲突 '" + fullPath + "'")
			panic("通配符路由 '" + path[i:end] +
				"' 和存在的children冲突 '" + fullPath + "'")
		}
		fmt.Println("[Kuuyee] insertChild===.3")
		// 检查如果通配符有名字
		if end-i < 2 {
			fmt.Println("wildcards must be named with a non-empty name in path '" + fullPath + "'")
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if c == ':' { // 这是参数
			// 在通配符开始分割路径
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType:     param,
				maxParams: numParams,
			}
			n.children = []*node{child}
			n.wildChild = true
			n = child
			n.priority++
			numParams--

			// 如果路径不是以通配符结尾，那么就是以'/'开头的子路径
			if end < max {
				n.path = path[offset:end]
				offset = end

				child := &node{
					maxParams: numParams,
					priority:  1,
				}
				n.children = []*node{child}
				n = child
			}
		} else { // catchAll
			if end != max || numParams > 1 {
				fmt.Println("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				fmt.Println("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
				panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				fmt.Println("no / before catch-all in path '" + fullPath + "'")
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			n.path = path[offset:i]

			// first node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
				maxParams: 1,
			}
			n.children = []*node{child}
			n.indices = string(path[i])
			n = child
			n.priority++

			// second node: node holding the variable
			child = &node{
				path:      path[i:],
				nType:     catchAll,
				maxParams: 1,
				handle:    handle,
				priority:  1,
			}
			n.children = []*node{child}
			return
		}
	}

	//
	n.path = path[offset:]
	n.handle = handle
}
