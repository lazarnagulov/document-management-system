package chunker

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestChunk_EmptyText(t *testing.T) {
	c := New(1000, 200)
	chunks := c.Chunk("")
	if len(chunks) != 0 {
		t.Errorf("expected no chunks for empty text, got %d", len(chunks))
	}
}

func TestChunk_WhitespaceOnly(t *testing.T) {
	c := New(1000, 200)
	chunks := c.Chunk("   \n\n   ")
	if len(chunks) != 0 {
		t.Errorf("expected no chunks for whitespace, got %d", len(chunks))
	}
}

func TestChunk_ShortText_SingleChunk(t *testing.T) {
	c := New(1000, 200)
	text := "This is a short document."
	chunks := c.Chunk(text)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != text {
		t.Errorf("expected %q, got %q", text, chunks[0])
	}
}

func TestChunk_RespectsChunkSize(t *testing.T) {
	c := New(100, 20)
	text := strings.Repeat("word ", 100)
	chunks := c.Chunk(text)

	for i, chunk := range chunks {
		if len(chunk) > 100+20 {
			t.Errorf("chunk %d exceeds size limit: len=%d", i, len(chunk))
		}
	}
}

func TestChunk_OverlapCarriedForward(t *testing.T) {
	c := New(50, 20)
	text := strings.Repeat("abcdefghij", 20)

	chunks := c.Chunk(text)

	if len(chunks) < 2 {
		t.Fatal("expected multiple chunks")
	}

	end := chunks[0][len(chunks[0])-20:]
	if !strings.HasPrefix(chunks[1], end) {
		t.Errorf("expected chunk[1] to start with last 20 chars of chunk[0]\nchunk[0] end: %q\nchunk[1] start: %q",
			end, chunks[1][:min(20, len(chunks[1]))])
	}
}

func TestChunk_PreservesParagraphBoundaries(t *testing.T) {
	c := New(1000, 200)
	text := "First paragraph about topic A.\n\nSecond paragraph about topic B.\n\nThird paragraph about topic C."

	chunks := c.Chunk(text)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if !strings.Contains(chunks[0], "First paragraph") {
		t.Error("expected first paragraph in chunk")
	}
	if !strings.Contains(chunks[0], "Third paragraph") {
		t.Error("expected third paragraph in chunk")
	}
}

func TestChunk_LargeParagraphSplitsOnSentences(t *testing.T) {
	c := New(100, 20)

	text := "This is the first sentence. This is the second sentence. This is the third sentence. This is the fourth sentence."

	chunks := c.Chunk(text)

	if len(chunks) < 2 {
		t.Fatal("expected large paragraph to be split into multiple chunks")
	}

	for i, chunk := range chunks {
		trimmed := strings.TrimSpace(chunk)
		last := rune(trimmed[len(trimmed)-1])
		if !isSentenceEnd(last) && i < len(chunks)-1 {
			t.Errorf("chunk %d does not end on sentence boundary: %q", i, trimmed)
		}
	}
}

func TestChunk_NoCoverage_Lost(t *testing.T) {
	c := New(100, 20)
	text := "Word " + strings.Repeat("x", 500)

	chunks := c.Chunk(text)

	combined := strings.Join(chunks, "")
	if len(combined) < len(text)-len(text)/10 {
		t.Errorf("too much text lost in chunking: original=%d combined=%d", len(text), len(combined))
	}
}

func TestChunk_DefaultsApplied(t *testing.T) {
	c := New(0, -1)
	chunks := c.Chunk("Some text to chunk.")

	if len(chunks) == 0 {
		t.Error("expected at least one chunk with default settings")
	}
}

func TestChunk_TextWithNoSentenceBoundaries(t *testing.T) {
	c := New(20, 5)
	text := "one two three four five six seven eight nine ten"

	chunks := c.Chunk(text)

	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}
	for i, chunk := range chunks {
		if len(chunk) > c.chunkSize+c.chunkOverlap {
			t.Errorf("chunk %d exceeds size: %d", i, len(chunk))
		}
	}
}

func TestChunk_SingleVeryLongWord(t *testing.T) {
	c := New(10, 3)
	text := strings.Repeat("a", 50)

	chunks := c.Chunk(text)

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
}

func TestChunk_MultipleParagraphsSpanChunks(t *testing.T) {
	c := New(60, 10)
	text := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.\n\nFourth paragraph."

	chunks := c.Chunk(text)

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	combined := strings.Join(chunks, " ")
	for _, para := range []string{"First", "Second", "Third", "Fourth"} {
		if !strings.Contains(combined, para) {
			t.Errorf("expected %q to appear in some chunk", para)
		}
	}
}

func TestChunk_OverlapDoesNotExceedChunkContent(t *testing.T) {
	c := New(30, 25)
	text := strings.Repeat("hello world. ", 20)

	chunks := c.Chunk(text)

	for i, chunk := range chunks {
		if len(chunk) == 0 {
			t.Errorf("chunk %d is empty", i)
		}
	}
}

func TestChunk_UTF8LongWord(t *testing.T) {
	c := New(10, 3)

	word := strings.Repeat("љ", 50)

	chunks := c.Chunk(word)

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	for i, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > c.chunkSize+c.chunkOverlap {
			t.Fatalf("chunk %d too large: %d runes",
				i,
				utf8.RuneCountInString(chunk))
		}
	}
}

func TestChunk_UTF8MixedCharacters(t *testing.T) {
	c := New(20, 5)

	text := strings.Repeat("чđšљž ", 20)

	chunks := c.Chunk(text)

	if len(chunks) < 2 {
		t.Fatal("expected multiple chunks")
	}

	combined := strings.Join(chunks, "")

	for _, r := range []rune{'ч', 'đ', 'š', 'љ', 'ž'} {
		if !strings.ContainsRune(combined, r) {
			t.Fatalf("lost rune %q", r)
		}
	}
}

func TestChunk_UTF8Overlap(t *testing.T) {
	c := New(10, 4)

	text := strings.Repeat("љ", 30)

	chunks := c.Chunk(text)

	if len(chunks) < 2 {
		t.Fatal("expected multiple chunks")
	}

	end := []rune(chunks[0])
	start := []rune(chunks[1])

	end = end[len(end)-4:]

	if string(end) != string(start[:4]) {
		t.Fatalf("expected overlap %q, got %q",
			string(end),
			string(start[:4]))
	}
}

func TestChunk_ChunkSizeMeasuredInRunes(t *testing.T) {
	c := New(10, 2)

	text := strings.Repeat("ч", 25)

	chunks := c.Chunk(text)

	for i, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > c.chunkSize+c.chunkOverlap {
			t.Fatalf(
				"chunk %d contains %d runes",
				i,
				utf8.RuneCountInString(chunk),
			)
		}
	}
}
