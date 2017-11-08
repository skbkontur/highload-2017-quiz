package hlconf2017

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

type (
	prefixListNodes []*prefixListNode

	prefixList struct {
		lst prefixListNodes

		bytesBuf       bytes.Buffer
		byteSliceSlice [][]byte
		stringsPool    sync.Pool
	}

	prefixListNode struct {
		anyPrefix, anySuffix bool
		prefix, infix        []byte

		patternName string

		children prefixListNodes
	}
)

type (
	FastPatternMatcher struct {
		tree *prefixList

		AllowedPatterns []Pattern
	}
)

func MakePrefixList() (pl *prefixList) {
	return &prefixList{
		stringsPool: sync.Pool{
			New: func() interface{} {
				return make([]string, 0, 30)
			},
		},
	}
}

func (pl *prefixList) addPart(part []byte) (nodes prefixListNodes) {
	var node prefixListNode

	if l := len(part); part[l-1] == '*' {
		node.anySuffix = true
		part = part[:l-1]
	}

	asterixPos := bytes.IndexByte(part, '*')
	braceFrom := bytes.IndexByte(part, '{')

	if asterixPos == -1 && braceFrom == -1 {
		node.prefix = part

		nodes = append(nodes, &node)
		return
	}

	if asterixPos > 0 {
		node.anySuffix = true
		node.prefix = part[:asterixPos]
		part = part[asterixPos+1:]
	} else if asterixPos == 0 {
		node.anyPrefix = true
		part = part[1:]
	}

	if braceFrom == -1 {
		node.infix = part

		nodes = append(nodes, &node)
	} else {
		braceFrom := bytes.IndexByte(part, '{')
		braceTo := bytes.IndexByte(part, '}')

		var prefix, suffix []byte
		if braceFrom > 0 {
			prefix = part[0:braceFrom]
		}
		if braceTo < len(part) {
			suffix = part[braceTo+1:]
		}

		prev := braceFrom + 1
		for i := prev; i <= braceTo; i++ {
			ch := part[i]
			if ch == ',' || ch == '}' {
				nodeCopy := node
				bb := &pl.bytesBuf
				bb.Reset()
				bb.Write(prefix)
				bb.Write(part[prev:i])
				bb.Write(suffix)
				nodeCopy.infix = append([]byte{}, bb.Bytes()...)
				nodes = append(nodes, &nodeCopy)

				if ch == ',' {
					prev = i + 1
				} else {
					break
				}
			}
		}
	}

	return
}

func (pl *prefixList) splitNameByParts(name []byte, reuse [][]byte) (parts [][]byte) {
	parts = reuse[0:0]

	prev := 0
	for i, ch := range name {
		if ch == '.' {
			part := name[prev:i]
			prev = i + 1

			parts = append(parts, part)
		}
	}
	if prev < len(name) {
		part := name[prev:]
		parts = append(parts, part)
	}
	return
}

func (t prefixListNodes) ToString() string {
	var bb bytes.Buffer

	bb.WriteString(`[`)

	for _, node := range t {
		fmt.Fprintf(&bb, `{anyPref:%v, anySuf:%v, pref:%s, inf:%s, name:%s, child:[%s]}, `,
			node.anyPrefix, node.anySuffix, node.prefix, node.infix, node.patternName, node.children.ToString(),
		)
	}

	bb.WriteString(`]`)

	return string(bb.Bytes())
}

func (pl *prefixList) Add(pat string) {
	pattern := []byte(pat)

	parts := pl.splitNameByParts(pattern, nil)

	var prevNodes prefixListNodes

	for i := len(parts) - 1; i >= 0; i-- {
		curNodes := pl.addPart(parts[i])

		for j := range curNodes {
			curNodes[j].children = append(curNodes[j].children, prevNodes...)
		}

		if len(prevNodes) == 0 {
			for j := range curNodes {
				curNodes[j].patternName = pat
			}
		}

		prevNodes = curNodes
	}

	pl.lst = append(pl.lst, prevNodes...)
}

func (pl *prefixList) checkPartWithNode(part []byte, node *prefixListNode) bool {
	// Simple [{anyPref:false, anySuf:false, pref:Simple, inf:
	if !node.anyPrefix && !node.anySuffix && len(node.infix) == 0 && bytes.Equal(node.prefix, part) {
		return true
	}

	// one [{anyPref:false, anySuf:false, pref:, inf:one
	if !node.anyPrefix && !node.anySuffix && len(node.prefix) == 0 && bytes.Equal(node.infix, part) {
		return true
	}

	// prefixonesuffix [{anyPref:false, anySuf:true, pref:, inf:
	if !node.anyPrefix && node.anySuffix && len(node.prefix) == 0 && len(node.infix) == 0 {
		return true
	}

	// prefixonesuffix [{anyPref:true, anySuf:true, pref:, inf:onesuf
	if node.anyPrefix && node.anySuffix && len(node.prefix) == 0 && bytes.Index(part, node.infix) != -1 {
		return true
	}

	// anything [{anyPref:false, anySuf:true, pref:any, inf:
	if !node.anyPrefix && node.anySuffix && len(node.infix) == 0 && bytes.HasPrefix(part, node.prefix) {
		return true
	}

	return false
}

func (pl *prefixList) find(parts [][]byte, node *prefixListNode, patternNames *[]string) {
	if len(parts) == 0 {
		return
	}

	part := parts[0]

	succ := pl.checkPartWithNode(part, node)
	if !succ {
		return
	}

	if node.children == nil {
		*patternNames = append(*patternNames, node.patternName)
		return
	}

	for _, child := range node.children {
		pl.find(parts[1:], child, patternNames)
	}
}

func (pl *prefixList) Find(name string) (patternNames []string) {
	// можно просто []byte(name)
	nameBytes := *(*[]byte)(unsafe.Pointer(
		&reflect.SliceHeader{
			Data: (*reflect.StringHeader)(unsafe.Pointer(&name)).Data,
			Len:  len(name),
			Cap:  len(name),
		}),
	)

	pl.byteSliceSlice = pl.splitNameByParts(nameBytes, pl.byteSliceSlice)

	namesIface := pl.stringsPool.Get()

	names := namesIface.([]string)[0:0]

	for _, node := range pl.lst {
		pl.find(pl.byteSliceSlice, node, &names)
	}

	if len(names) > 0 {
		patternNames = append(patternNames, names...)
	}

	pl.stringsPool.Put(namesIface)

	return
}

// InitPatterns accepts allowed patterns in Graphite format, e.g.
//   metric.name.single
//   metric.name.*
//   metric.name.wild*card
//   metric.name.{one,two}.maybe.longer
func (p *FastPatternMatcher) InitPatterns(allowedPatterns []string) {
	p.AllowedPatterns = make([]Pattern, len(allowedPatterns))
	p.tree = MakePrefixList()

	for _, pat := range allowedPatterns {
		p.tree.Add(pat)
	}
}

// DetectMatchingPatterns returns a list of allowed patterns that match given metric
func (p *FastPatternMatcher) DetectMatchingPatterns(metricName string) (matchingPatterns []string) {
	return p.tree.Find(metricName)
}
