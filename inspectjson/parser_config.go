package inspectjson

type ParserOption interface {
	applyParser(t *parser)
}

type ParserConfig struct {
	keepReplacedObjectMembers *bool
}

func (p ParserConfig) applyParser(t *parser) {
	if p.keepReplacedObjectMembers != nil {
		t.keepReplacedObjectMembers = *p.keepReplacedObjectMembers
	}
}

func (p ParserConfig) SetKeepReplacedObjectMembers(enable bool) ParserConfig {
	p.keepReplacedObjectMembers = &enable

	return p
}
