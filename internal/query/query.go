package query

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/tokopedia/graphql-go/errors"
	"github.com/tokopedia/graphql-go/internal/common"
	"github.com/tokopedia/graphql-go/types"
)

const (
	Query        types.OperationType = "QUERY"
	Mutation     types.OperationType = "MUTATION"
	Subscription types.OperationType = "SUBSCRIPTION"
)

func Parse(queryString string) (*types.ExecutableDefinition, *errors.QueryError) {
	l := common.NewLexer(queryString, false)

	var execDef *types.ExecutableDefinition
	err := l.CatchSyntaxError(func() { execDef = parseExecutableDefinition(l) })
	if err != nil {
		return nil, err
	}

	return execDef, nil
}

func parseExecutableDefinition(l *common.Lexer) *types.ExecutableDefinition {
	ed := &types.ExecutableDefinition{}
	l.ConsumeWhitespace()
	for l.Peek() != scanner.EOF {
		if l.Peek() == '{' {
			op := &types.OperationDefinition{Type: Query, Loc: l.Location()}
			op.Selections = parseSelectionSet(l)
			ed.Operations = append(ed.Operations, op)
			continue
		}

		loc := l.Location()
		switch x := l.ConsumeIdent(); x {
		case "query":
			op := parseOperation(l, Query)
			op.Loc = loc
			ed.Operations = append(ed.Operations, op)

		case "mutation":
			ed.Operations = append(ed.Operations, parseOperation(l, Mutation))

		case "subscription":
			ed.Operations = append(ed.Operations, parseOperation(l, Subscription))

		case "fragment":
			frag := parseFragment(l)
			frag.Loc = loc
			ed.Fragments = append(ed.Fragments, frag)

		default:
			l.SyntaxError(fmt.Sprintf(`unexpected %q, expecting "fragment"`, x))
		}
	}
	return ed
}

func parseOperation(l *common.Lexer, opType types.OperationType) *types.OperationDefinition {
	op := &types.OperationDefinition{Type: opType}
	op.Name.Loc = l.Location()
	if l.Peek() == scanner.Ident {
		op.Name = l.ConsumeIdentWithLoc()
	}
	op.Directives = common.ParseDirectives(l)
	if l.Peek() == '(' {
		l.ConsumeToken('(')
		for l.Peek() != ')' {
			loc := l.Location()
			l.ConsumeToken('$')
			iv := common.ParseInputValue(l)
			iv.Loc = loc
			op.Vars = append(op.Vars, iv)
		}
		l.ConsumeToken(')')
	}
	op.Selections = parseSelectionSet(l)
	return op
}

func parseFragment(l *common.Lexer) *types.FragmentDefinition {
	f := &types.FragmentDefinition{}
	f.Name = l.ConsumeIdentWithLoc()
	l.ConsumeKeyword("on")
	f.On = types.TypeName{Ident: l.ConsumeIdentWithLoc()}
	f.Directives = common.ParseDirectives(l)
	f.Selections = parseSelectionSet(l)
	return f
}

func parseSelectionSet(l *common.Lexer) []types.Selection {
	var sels []types.Selection
	originalFields := make(map[string]*types.Field, 0)
	l.ConsumeToken('{')
	for l.Peek() != '}' {
		f := parseSelection(l)
		switch sel := f.(type) {
		case *types.Field:
			if !strings.HasSuffix(sel.Name.Name, types.DUPLICATION_SUFFIX) {
				sels = append(sels, f)
				originalFields[sel.Name.Name] = sel
			} else {
				origFieldName := strings.ReplaceAll(sel.Name.Name, types.DUPLICATION_SUFFIX, "")
				// we instantly check map, so original field MUST BE BEFORE duplicate field in request
				origField, ok := originalFields[origFieldName]
				if ok {
					origField.NeedStrCounterpart = true
					origField.CounterpartAlias = sel.Alias.Name
				} else { // no original field, we append to schema so it will respons with invalid schema
					sels = append(sels, f)
				}
			}
		default:
			sels = append(sels, f)
		}
	}
	l.ConsumeToken('}')
	return sels
}

func parseSelection(l *common.Lexer) types.Selection {
	if l.Peek() == '.' {
		return parseSpread(l)
	}
	return parseFieldDef(l)
}

func parseFieldDef(l *common.Lexer) *types.Field {
	f := &types.Field{}
	f.Alias = l.ConsumeIdentWithLoc()
	f.Name = f.Alias
	if l.Peek() == ':' {
		l.ConsumeToken(':')
		f.Name = l.ConsumeIdentWithLoc()
	}
	if l.Peek() == '(' {
		f.Arguments = common.ParseArgumentList(l)
	}
	f.Directives = common.ParseDirectives(l)
	if l.Peek() == '{' {
		f.SelectionSetLoc = l.Location()
		f.SelectionSet = parseSelectionSet(l)
	}
	return f
}

func parseSpread(l *common.Lexer) types.Selection {
	loc := l.Location()
	l.ConsumeToken('.')
	l.ConsumeToken('.')
	l.ConsumeToken('.')

	f := &types.InlineFragment{Loc: loc}
	if l.Peek() == scanner.Ident {
		ident := l.ConsumeIdentWithLoc()
		if ident.Name != "on" {
			fs := &types.FragmentSpread{
				Name: ident,
				Loc:  loc,
			}
			fs.Directives = common.ParseDirectives(l)
			return fs
		}
		f.On = types.TypeName{Ident: l.ConsumeIdentWithLoc()}
	}
	f.Directives = common.ParseDirectives(l)
	f.Selections = parseSelectionSet(l)
	return f
}
