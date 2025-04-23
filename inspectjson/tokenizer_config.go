package inspectjson

import "github.com/dpb587/cursorio-go/cursorio"

type TokenizerOption interface {
	applyTokenizer(t *tokenizerOptions)
}

type TokenizerConfig struct {
	sourceOffsets       *bool
	sourceInitialOffset *cursorio.TextOffset
	emitWhitespace      *bool
	laxBehaviors        map[SyntaxBehavior]bool
	syntaxRecoveryHook  SyntaxRecoveryHookFunc
	multistream         *bool
}

func (o TokenizerConfig) applyParser(p *parser) {
	panic("should not be called directly")
}

func (o TokenizerConfig) applyTokenizer(t *tokenizerOptions) {
	if o.sourceOffsets != nil {
		t.sourceOffsets = *o.sourceOffsets
	}

	if o.sourceInitialOffset != nil {
		t.sourceInitialOffset = *o.sourceInitialOffset
	}

	if o.emitWhitespace != nil {
		t.emitWhitespace = *o.emitWhitespace
	}

	for behavior, enable := range o.laxBehaviors {
		if enable {
			t.laxBehaviors = t.laxBehaviors | behavior
		} else {
			t.laxBehaviors = t.laxBehaviors &^ behavior
		}
	}

	if o.syntaxRecoveryHook != nil {
		t.laxListener = o.syntaxRecoveryHook
	}

	if o.multistream != nil {
		t.multistream = *o.multistream
	}
}

func (o TokenizerConfig) SetEmitWhitespace(enable bool) TokenizerConfig {
	o.emitWhitespace = &enable

	return o
}

func (o TokenizerConfig) SetLax(enable bool) TokenizerConfig {
	o.laxBehaviors = map[SyntaxBehavior]bool{
		LaxIgnoreBlockComment:        enable,
		LaxIgnoreLineComment:         enable,
		LaxStringEscapeInvalidEscape: enable,
		LaxStringEscapeMissingEscape: enable,
		LaxNumberTrimLeadingZero:     enable,
		LaxLiteralCaseInsensitive:    enable,
		LaxIgnoreExtraComma:          enable,
		LaxIgnoreTrailingSemicolon:   enable,
	}

	return o
}

func (o TokenizerConfig) SetLaxBehavior(behavior SyntaxBehavior, enable bool) TokenizerConfig {
	if o.laxBehaviors == nil {
		o.laxBehaviors = map[SyntaxBehavior]bool{}
	}

	o.laxBehaviors[behavior] = enable

	return o
}

func (o TokenizerConfig) SetSyntaxRecoveryHook(hook SyntaxRecoveryHookFunc) TokenizerConfig {
	o.syntaxRecoveryHook = hook

	return o
}

func (o TokenizerConfig) SetMultistream(enable bool) TokenizerConfig {
	o.multistream = &enable

	return o
}

func (o TokenizerConfig) SetSourceOffsets(enable bool) TokenizerConfig {
	o.sourceOffsets = &enable

	return o
}

func (o TokenizerConfig) SetSourceInitialOffset(offset cursorio.TextOffset) TokenizerConfig {
	v := true

	o.sourceOffsets = &v
	o.sourceInitialOffset = &offset

	return o
}
