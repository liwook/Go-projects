package gee

import (
	"strings"
)

type node struct {
	path     string  //路由路径 例如 /aa.com/home
	part     string  //路由中由'/'分隔的部分
	children []*node //子节点
	isWild   bool    //是否是通配符节点，是为true
}

func (n *node) insert(path string, parts []string) {
	cur := n
	//将parts插入到路由树
	for _, part := range parts {
		var tmp *node
		for _, child := range cur.children {
			if child.part == part {
				tmp = child
				break
			}
		}

		if tmp == nil {
			tmp = &node{
				part:   part,
				isWild: part[0] == ':' || part[0] == '*',
			}
			cur.children = append(cur.children, tmp)
		}
		cur = tmp
	}
	cur.path = path
}

func (n *node) search(searchParts []string, height int) *node {
	if len(searchParts) == height || strings.HasPrefix(n.part, "*") {
		if n.path == "" {
			return nil
		}
		return n
	}

	part := searchParts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(searchParts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}

func (n *node) matchChildren(part string) (result []*node) {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part {
			result = append(result, child)
		} else if child.isWild {
			nodes = append(nodes, child)
		}
	}
	return append(result, nodes...)
}
