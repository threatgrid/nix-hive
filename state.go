package main

import (
	"bytes"
)

func appendFact(state []byte, predicate string, terms ...string) []byte {
	state = append(state, '(')
	state = append(state, quoteTerm(predicate)...)
	for _, term := range terms {
		state = append(state, '|')
		state = append(state, quoteTerm(term)...)
	}
	return append(state, ')', '\n')
}

func processState(state []byte, process func(predicate string, terms ...string) error) (int, error) {
	t := 0
	for {
		fact, n := splitFact(state)
		t += n
		if fact == nil {
			break
		}
		state = state[n:]
		predicate, terms := parseFact(fact)
		err := process(predicate, terms...)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

func splitFact(state []byte) (fact []byte, n int) {
	// each fact starts with a '(' and does not contain a '('.
	start := bytes.IndexByte(state, '(')
	if start == -1 {
		return nil, len(state) // there are no remaining facts in the state.
	}
	start += 1
	state = state[start:]

	// each fact ends with a ')' and does not contain a ')'.
	end := bytes.IndexByte(state, ')')
	if end == -1 {
		return nil, 0 // the state is incomplete.
	}
	return state[:end], start + end
}

func parseFact(fact []byte) (predicate string, terms []string) {
	if len(fact) == 0 {
		return ``, nil
	}
	terms = make([]string, 0, len(fact)/8)
	for len(fact) > 0 {
		p, n := splitTerm(fact)
		terms = append(terms, unquoteTerm(p))
		fact = fact[n:]
	}
	return terms[0], terms[1:]
}

func splitTerm(fact []byte) ([]byte, int) {
	ix := bytes.IndexByte(fact, '|')
	if ix == -1 {
		return fact, len(fact)
	}
	return fact[:ix], ix + 1
}

func unquoteTerm(term []byte) string {
	buf := make([]byte, 0, len(term))
	end := len(term) - 1
	for pos := 0; pos <= end; pos++ {
		ch := term[pos]
		switch ch {
		case '~':
			if pos < end {
				pos++
				switch term[pos] {
				case 'q':
					ch = '~'
				case 'p':
					ch = '|'
				case 's':
					ch = '('
				case 'e':
					ch = ')'
				default:
					// leave the ~ alone, but this is an error.
				}
			}
		}
		buf = append(buf, ch)
	}
	return string(buf)
}

func quoteTerm(term string) string {
	buf := make([]byte, 0, len(term)+4)
	for i := 0; i < len(term); i++ {
		ch := term[i]
		switch ch {
		case '~':
			buf = append(buf, '~', 'q')
		case '|':
			buf = append(buf, '~', 'p')
		case '(':
			buf = append(buf, '~', 's')
		case ')':
			buf = append(buf, '~', 'e')
		default:
			buf = append(buf, ch)
		}
	}
	return string(buf)
}
