package chunker

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	DefaultChunkSize    = 1000
	DefaultChunkOverlap = 200
)

type chunkState struct {
	chunks  []string
	current strings.Builder
	length  int
}

type Chunker struct {
	chunkSize    int
	chunkOverlap int
}

func New(chunkSize, chunkOverlap int) *Chunker {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if chunkOverlap <= 0 || chunkOverlap >= chunkSize {
		chunkOverlap = DefaultChunkOverlap
	}
	return &Chunker{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}
}

func Default() *Chunker {
	return &Chunker{chunkSize: DefaultChunkSize, chunkOverlap: DefaultChunkOverlap}
}

func (c *Chunker) Chunk(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	state := &chunkState{}

	for _, paragraph := range c.splitParagraphs(text) {
		c.processParagraph(state, paragraph)
	}

	if state.length > 0 {
		state.chunks = append(state.chunks, state.current.String())
	}

	return state.chunks
}

func (c *Chunker) processParagraph(state *chunkState, paragraph string) {
	if c.appendText(state, paragraph, "\n\n") {
		return
	}

	c.flush(state)

	if c.appendText(state, paragraph, "\n\n") {
		return
	}

	for _, sentence := range c.splitSentences(paragraph) {
		c.processSentence(state, sentence)
	}
}

func (c *Chunker) processSentence(state *chunkState, sentence string) {
	if c.appendText(state, sentence, " ") {
		return
	}

	c.flush(state)

	if c.appendText(state, sentence, " ") {
		return
	}

	for _, word := range strings.Fields(sentence) {
		c.processWord(state, word)
	}
}

func (c *Chunker) processWord(state *chunkState, word string) {
	if c.appendText(state, word, " ") {
		return
	}

	c.flush(state)

	if c.appendText(state, word, " ") {
		return
	}

	c.splitLongWord(state, word)
}

func (c *Chunker) splitLongWord(state *chunkState, word string) {
	runes := []rune(word)

	for len(runes) > 0 {
		if state.length == c.chunkSize {
			c.flush(state)
		}

		available := c.chunkSize - state.length
		if available <= 0 {
			c.flush(state)
			available = c.chunkSize - state.length
		}

		if available > len(runes) {
			available = len(runes)
		}

		part := string(runes[:available])

		state.current.WriteString(part)
		state.length += available

		runes = runes[available:]

		if len(runes) > 0 {
			c.flush(state)
		}
	}
}

func (c *Chunker) appendText(state *chunkState, text, separator string) bool {
	add := utf8.RuneCountInString(text)

	if state.length > 0 {
		add += utf8.RuneCountInString(separator)
	}

	if state.length+add > c.chunkSize {
		return false
	}

	if state.length > 0 {
		state.current.WriteString(separator)
		state.length += utf8.RuneCountInString(separator)
	}

	state.current.WriteString(text)
	state.length += utf8.RuneCountInString(text)

	return true
}

func (c *Chunker) flush(state *chunkState) {
	if state.length == 0 {
		return
	}

	chunk := state.current.String()
	state.chunks = append(state.chunks, chunk)

	tail := overlap(chunk, c.chunkOverlap)

	state.current.Reset()
	state.current.WriteString(tail)
	state.length = utf8.RuneCountInString(tail)
}

func (c *Chunker) splitParagraphs(text string) []string {
	parts := strings.Split(text, "\n\n")

	result := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}

func (c *Chunker) splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)

	for i, r := range runes {
		current.WriteRune(r)
		if isSentenceEnd(r) {
			if i == len(runes)-1 {
				sentences = append(sentences, strings.TrimSpace(current.String()))
				current.Reset()
				break
			}
			next := runes[i+1]

			if unicode.IsSpace(next) {
				j := i + 1
				for j < len(runes) && unicode.IsSpace(runes[j]) {
					j++
				}
				if j >= len(runes) || unicode.IsUpper(runes[j]) || unicode.IsDigit(runes[j]) {
					s := strings.TrimSpace(current.String())
					if s != "" {
						sentences = append(sentences, s)
					}
					current.Reset()
				}
			}
		}
	}

	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())
		if s != "" {
			sentences = append(sentences, s)
		}
	}

	return sentences
}

func isSentenceEnd(r rune) bool {
	return r == '.' || r == '!' || r == '?'
}

func overlap(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[len(runes)-n:])
}
