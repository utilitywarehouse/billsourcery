package graph

import (
	"bytes"

	"github.com/utilitywarehouse/equilex"
)

type token struct {
	tok equilex.Token
	lit string
}

type statement struct {
	tokens []token
}

func (stmt *statement) String() string {
	var buf bytes.Buffer
	for _, t := range stmt.tokens {
		buf.WriteString(t.lit)
	}
	return buf.String()
}

func (stmt *statement) add(tok equilex.Token, lit string) {
	stmt.tokens = append(stmt.tokens, token{tok, lit})
}
