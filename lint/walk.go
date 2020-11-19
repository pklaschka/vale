package lint

import (
	"bytes"

	"github.com/errata-ai/vale/v2/core"
	"golang.org/x/net/html"
)

// walker ...
type walker struct {
	lines   int
	section string
	context string

	idx int
	z   *html.Tokenizer

	// queue holds each segment of text we encounter in a block, which we then
	// use to sequentially update our context.
	queue []string

	// tagHistory holds the HTML tags we encounter in a given block -- e.g.,
	// if we see <ul>, <li>, <p>, we'd get tagHistory = [ul li p]. It's reset
	// on every non-inline end tag.
	tagHistory []string

	activeTag string
}

func newWalker(f *core.File, raw []byte, offset int) walker {
	return walker{
		lines:   len(f.Lines) + offset,
		context: f.Content,
		z:       html.NewTokenizer(bytes.NewReader(raw))}
}

func (w *walker) reset() {
	for _, s := range w.queue {
		w.context = updateCtx(w.context, s, html.TextToken)
	}
	w.queue = []string{}
	w.tagHistory = []string{}
}

func (w *walker) append(s string) {
	w.queue = append(w.queue, s)
}

func (w *walker) addTag(t string) {
	w.tagHistory = append(w.tagHistory, t)
	w.activeTag = t
}

func (w *walker) block(text, scope string) core.Block {
	return core.NewBlock(w.context, text, scope)
}

func (w *walker) replaceToks(tok html.Token) {
	if core.StringInSlice(tok.Data, []string{"img", "a", "p", "script"}) {
		for _, a := range tok.Attr {
			if a.Key == "href" || a.Key == "id" || a.Key == "src" {
				w.context = updateCtx(w.context, a.Val, html.TextToken)
			}
		}
	}
}
