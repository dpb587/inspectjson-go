package inspectjson

type ParserOptionsApplier interface {
	applyParser(t *parser)
}

type ParserOptions struct {
	keepReplacedObjectMembers *bool
}

func (p ParserOptions) applyParser(t *parser) {
	if p.keepReplacedObjectMembers != nil {
		t.keepReplacedObjectMembers = *p.keepReplacedObjectMembers
	}
}

func (p ParserOptions) KeepReplacedObjectMembers(enable bool) ParserOptions {
	p.keepReplacedObjectMembers = &enable

	return p
}
