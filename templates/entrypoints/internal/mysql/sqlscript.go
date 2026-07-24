package mysql

import (
	"fmt"
	"strings"
)

type sqlStatementParser struct {
	delimiter    string
	statement    strings.Builder
	statements   []string
	quote        byte
	lineComment  bool
	blockComment bool
	keepComment  bool
}

func splitSQLStatements(script string) ([]string, error) {
	p := &sqlStatementParser{delimiter: ";"}
	for len(script) > 0 {
		line := script
		if index := strings.IndexByte(script, '\n'); index >= 0 {
			line = script[:index+1]
			script = script[index+1:]
		} else {
			script = ""
		}
		if p.canChangeDelimiter() {
			fields := strings.Fields(strings.TrimSpace(line))
			if len(fields) == 2 && strings.EqualFold(fields[0], "DELIMITER") {
				p.delimiter = fields[1]
				continue
			}
		}
		if err := p.write(line); err != nil {
			return nil, err
		}
	}
	if p.quote != 0 {
		return nil, fmt.Errorf("unterminated quoted string")
	}
	if p.blockComment {
		return nil, fmt.Errorf("unterminated block comment")
	}
	p.flush()
	return p.statements, nil
}

func (p *sqlStatementParser) canChangeDelimiter() bool {
	return p.quote == 0 && !p.lineComment && !p.blockComment && strings.TrimSpace(p.statement.String()) == ""
}

func (p *sqlStatementParser) write(value string) error {
	for index := 0; index < len(value); {
		character := value[index]
		if p.lineComment {
			if character == '\n' {
				p.lineComment = false
				p.statement.WriteByte('\n')
			}
			index++
			continue
		}
		if p.blockComment {
			if index+1 < len(value) && value[index:index+2] == "*/" {
				if p.keepComment {
					p.statement.WriteString("*/")
				} else {
					p.statement.WriteByte(' ')
				}
				p.blockComment = false
				p.keepComment = false
				index += 2
				continue
			}
			if p.keepComment {
				p.statement.WriteByte(character)
			}
			index++
			continue
		}
		if p.quote != 0 {
			p.statement.WriteByte(character)
			if character == '\\' && index+1 < len(value) {
				p.statement.WriteByte(value[index+1])
				index += 2
				continue
			}
			if character == p.quote {
				if index+1 < len(value) && value[index+1] == p.quote {
					p.statement.WriteByte(value[index+1])
					index += 2
					continue
				}
				p.quote = 0
			}
			index++
			continue
		}

		if strings.HasPrefix(value[index:], p.delimiter) {
			p.flush()
			index += len(p.delimiter)
			continue
		}
		if character == '#' || (character == '-' && index+2 < len(value) && value[index:index+2] == "--" && isSQLSpace(value[index+2])) {
			p.lineComment = true
			if character == '-' {
				index += 2
			} else {
				index++
			}
			continue
		}
		if index+1 < len(value) && value[index:index+2] == "/*" {
			p.blockComment = true
			p.keepComment = index+2 < len(value) && (value[index+2] == '!' || value[index+2] == '+')
			if p.keepComment {
				p.statement.WriteString("/*")
			}
			index += 2
			continue
		}
		if character == '\'' || character == '"' || character == '`' {
			p.quote = character
		}
		p.statement.WriteByte(character)
		index++
	}
	return nil
}

func (p *sqlStatementParser) flush() {
	statement := strings.TrimSpace(p.statement.String())
	p.statement.Reset()
	if statement != "" {
		p.statements = append(p.statements, statement)
	}
}

func isSQLSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n'
}
