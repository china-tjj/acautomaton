package acautomaton

import (
	"io"
	"unicode/utf8"
)

type ITrie interface {
	// PreMatchFirst 前缀匹配第一个 term
	PreMatchFirst(query string) (MatchResult, bool)
	// PreMatchAll 前缀匹配所有的 term
	PreMatchAll(query string) []MatchResult
	// Save 序列化
	Save(w io.Writer) error
	// SaveToFile 序列化并保存到文件
	SaveToFile(filename string) error
}

type IAcAutomaton interface {
	ITrie
	// MatchFirst 多模式匹配第一个 term
	MatchFirst(query string) (MatchResult, bool)
	// MatchAll 多模式匹配所有的 term
	MatchAll(query string) []MatchResult
	// MatchAllUnique 多模式匹配所有的 term，每个 term 最多匹配一次，建树时 terms 数组里的每个 term 均被视为不同的 term（因为可能会关联不同的业务信息）
	MatchAllUnique(query string) []MatchResult
}

type uints interface {
	uint8 | uint16 | uint32 | uint64
}

func mid[T uints](i, j T) T {
	return (i & j) + ((i ^ j) >> 1)
}

type iTrie[U uints] interface {
	root() U
	getChild(p U, char rune) (U, bool)
	// 不用iter.Seq2，因为这个场景下闭包开销比较大
	rangeTerms(p U, yield func(term termMetadata[U]) bool)
	getFirstTerm(p U) (term termMetadata[U], ok bool)
}

func triePreMatchFirst[U uints, T iTrie[U]](t T, query string) (result MatchResult, ok bool) {
	p := t.root()
	for byteIdx, char := range query {
		// 匹配成功，移动到子节点，失配时，break
		var found bool
		p, found = t.getChild(p, char)
		if !found {
			break
		}
		if term, ok := t.getFirstTerm(p); ok {
			endIdx := byteIdx + utf8.RuneLen(char)
			return makeMatchResult(query, term, endIdx), true
		}
	}
	return MatchResult{}, false
}

func triePreMatchAll[U uints, T iTrie[U]](t T, query string) []MatchResult {
	var results []MatchResult
	p := t.root()
	for byteIdx, char := range query {
		// 匹配成功，移动到子节点，失配时，break
		var found bool
		p, found = t.getChild(p, char)
		if !found {
			break
		}
		endIdx := byteIdx + utf8.RuneLen(char)
		t.rangeTerms(p, func(term termMetadata[U]) bool {
			results = append(results, makeMatchResult(query, term, endIdx))
			return true
		})
	}
	return results
}
