package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/messeb/escpos-cli/internal/escpos"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

func TestNewESCPOSRenderer(t *testing.T) {
	r := NewESCPOSRenderer("/test/dir", 75)

	if r == nil {
		t.Fatal("NewESCPOSRenderer() returned nil")
	}

	if r.markdownDir != "/test/dir" {
		t.Errorf("markdownDir = %q, want %q", r.markdownDir, "/test/dir")
	}

	if r.imageThreshold != 75 {
		t.Errorf("imageThreshold = %d, want %d", r.imageThreshold, 75)
	}

	if r.listDepth != 0 {
		t.Errorf("listDepth = %d, want 0", r.listDepth)
	}

	if r.currentAlign != 0 {
		t.Errorf("currentAlign = %d, want 0", r.currentAlign)
	}

	if r.inBold {
		t.Error("inBold = true, want false")
	}
}

func TestESCPOSRenderer_Render(t *testing.T) {
	tests := []struct {
		name          string
		markdown      string
		checkContains [][]byte
	}{
		{
			name:     "simple text",
			markdown: "Hello World",
			checkContains: [][]byte{
				escpos.Init(),
				[]byte("Hello World"),
				escpos.Cut(),
			},
		},
		{
			name:     "heading",
			markdown: "# Title",
			checkContains: [][]byte{
				escpos.AlignCenter(),
				escpos.BoldOn(),
				escpos.DoubleSize(),
				[]byte("Title"),
				escpos.BoldOff(),
				escpos.AlignLeft(),
			},
		},
		{
			name:     "bold text",
			markdown: "**bold**",
			checkContains: [][]byte{
				escpos.BoldOn(),
				[]byte("bold"),
				escpos.BoldOff(),
			},
		},
		{
			name:     "italic text",
			markdown: "*italic*",
			checkContains: [][]byte{
				escpos.UnderlineOn(),
				[]byte("italic"),
				escpos.UnderlineOff(),
			},
		},
		{
			name:     "unordered list",
			markdown: "- item1\n- item2",
			checkContains: [][]byte{
				[]byte("item1"),
				[]byte("item2"),
			},
		},
		{
			name:     "ordered list",
			markdown: "1. first\n2. second",
			checkContains: [][]byte{
				[]byte("first"),
				[]byte("second"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New()
			reader := text.NewReader([]byte(tt.markdown))
			doc := md.Parser().Parse(reader)

			renderer := NewESCPOSRenderer("/tmp", 75)
			var buf bytes.Buffer
			writer := &bufWriter{buf: &buf}

			err := renderer.Render(writer, []byte(tt.markdown), doc)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			output := buf.Bytes()

			for _, check := range tt.checkContains {
				if !containsBytes(output, check) {
					t.Errorf("Render() output does not contain expected bytes: %v", check)
				}
			}
		})
	}
}

func TestESCPOSRenderer_renderHeading(t *testing.T) {
	tests := []struct {
		level         int
		wantSize      []byte
		wantAlignment []byte
	}{
		{1, escpos.DoubleSize(), escpos.AlignCenter()},
		{2, escpos.DoubleHeight(), escpos.AlignCenter()},
		{3, escpos.DoubleWidth(), escpos.AlignCenter()},
		{4, escpos.NormalSize(), escpos.AlignCenter()},
	}

	for _, tt := range tests {
		t.Run(strings.Repeat("H", tt.level), func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			heading := ast.NewHeading(tt.level)

			// Test entering
			status, err := r.renderHeading(heading, true)
			if err != nil {
				t.Errorf("renderHeading(entering) error = %v", err)
			}
			if status != ast.WalkContinue {
				t.Errorf("renderHeading(entering) status = %v, want WalkContinue", status)
			}

			output := r.buf.Bytes()
			if !containsBytes(output, tt.wantAlignment) {
				t.Error("renderHeading(entering) missing alignment command")
			}
			if !containsBytes(output, escpos.BoldOn()) {
				t.Error("renderHeading(entering) missing bold on command")
			}
			if !containsBytes(output, tt.wantSize) {
				t.Error("renderHeading(entering) missing size command")
			}

			// Test exiting
			status, err = r.renderHeading(heading, false)
			if err != nil {
				t.Errorf("renderHeading(exiting) error = %v", err)
			}
			if status != ast.WalkContinue {
				t.Errorf("renderHeading(exiting) status = %v, want WalkContinue", status)
			}

			output = r.buf.Bytes()
			if !containsBytes(output, escpos.BoldOff()) {
				t.Error("renderHeading(exiting) missing bold off command")
			}
			if !containsBytes(output, escpos.AlignLeft()) {
				t.Error("renderHeading(exiting) missing align left command")
			}
		})
	}
}

func TestESCPOSRenderer_renderEmphasis(t *testing.T) {
	tests := []struct {
		name      string
		level     int
		wantOn    []byte
		wantOff   []byte
		wantField *bool
	}{
		{
			name:    "strong (bold)",
			level:   2,
			wantOn:  escpos.BoldOn(),
			wantOff: escpos.BoldOff(),
		},
		{
			name:    "emphasis (underline)",
			level:   1,
			wantOn:  escpos.UnderlineOn(),
			wantOff: escpos.UnderlineOff(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			emph := ast.NewEmphasis(tt.level)

			// Test entering
			status, err := r.renderEmphasis(emph, true)
			if err != nil {
				t.Errorf("renderEmphasis(entering) error = %v", err)
			}
			if status != ast.WalkContinue {
				t.Errorf("renderEmphasis(entering) status = %v, want WalkContinue", status)
			}

			output := r.buf.Bytes()
			if !containsBytes(output, tt.wantOn) {
				t.Errorf("renderEmphasis(entering) missing command: %v", tt.wantOn)
			}

			// Test exiting
			r.buf.Reset()
			_, err = r.renderEmphasis(emph, false)
			if err != nil {
				t.Errorf("renderEmphasis(exiting) error = %v", err)
			}

			output = r.buf.Bytes()
			if !containsBytes(output, tt.wantOff) {
				t.Errorf("renderEmphasis(exiting) missing command: %v", tt.wantOff)
			}
		})
	}
}

func TestESCPOSRenderer_renderList(t *testing.T) {
	tests := []struct {
		name      string
		ordered   bool
		start     int
		wantDepth int
	}{
		{
			name:      "unordered list",
			ordered:   false,
			start:     0,
			wantDepth: 1,
		},
		{
			name:      "ordered list",
			ordered:   true,
			start:     1,
			wantDepth: 1,
		},
		{
			name:      "ordered list starting at 5",
			ordered:   true,
			start:     5,
			wantDepth: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			var list *ast.List
			if tt.ordered {
				list = ast.NewList('.')
				list.Start = tt.start
			} else {
				list = ast.NewList('-')
			}

			// Test entering
			status, _ := r.renderList(list, true)
			if status != ast.WalkContinue {
				t.Errorf("renderList(entering) status = %v, want WalkContinue", status)
			}
			if r.listDepth != tt.wantDepth {
				t.Errorf("listDepth = %d, want %d", r.listDepth, tt.wantDepth)
			}
			if len(r.listCounters) != tt.wantDepth {
				t.Errorf("len(listCounters) = %d, want %d", len(r.listCounters), tt.wantDepth)
			}
			if tt.ordered && r.listCounters[0] != tt.start {
				t.Errorf("listCounters[0] = %d, want %d", r.listCounters[0], tt.start)
			}

			// Test exiting
			status, _ = r.renderList(list, false)
			if status != ast.WalkContinue {
				t.Errorf("renderList(exiting) status = %v, want WalkContinue", status)
			}
			if r.listDepth != 0 {
				t.Errorf("listDepth after exit = %d, want 0", r.listDepth)
			}
		})
	}
}

func TestESCPOSRenderer_renderListItem(t *testing.T) {
	tests := []struct {
		name         string
		depth        int
		ordered      bool
		counter      int
		wantMarker   string
	}{
		{
			name:       "unordered depth 1",
			depth:      1,
			ordered:    false,
			wantMarker: "- ",
		},
		{
			name:       "unordered depth 2",
			depth:      2,
			ordered:    false,
			wantMarker: "*",
		},
		{
			name:       "unordered depth 3",
			depth:      3,
			ordered:    false,
			wantMarker: "+",
		},
		{
			name:       "ordered item 1",
			depth:      1,
			ordered:    true,
			counter:    1,
			wantMarker: "1. ",
		},
		{
			name:       "ordered item 5",
			depth:      1,
			ordered:    true,
			counter:    5,
			wantMarker: "5. ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			r.listDepth = tt.depth
			if tt.ordered {
				r.listCounters = make([]int, tt.depth)
				r.listCounters[tt.depth-1] = tt.counter
			} else {
				r.listCounters = make([]int, tt.depth)
			}

			listItem := ast.NewListItem(0)
			var list *ast.List
			if tt.ordered {
				list = ast.NewList('.')
			} else {
				list = ast.NewList('-')
			}
			list.AppendChild(list, listItem)

			// Test entering
			status, _ := r.renderListItem(listItem, true)
			if status != ast.WalkContinue {
				t.Errorf("renderListItem(entering) status = %v, want WalkContinue", status)
			}

			output := r.buf.String()
			if !strings.Contains(output, tt.wantMarker) {
				t.Errorf("renderListItem() output = %q, want to contain %q", output, tt.wantMarker)
			}

			// Test exiting
			r.buf.Reset()
			status, _ = r.renderListItem(listItem, false)
			if status != ast.WalkContinue {
				t.Errorf("renderListItem(exiting) status = %v, want WalkContinue", status)
			}
		})
	}
}

func TestESCPOSRenderer_renderTaskCheckBox(t *testing.T) {
	tests := []struct {
		name    string
		checked bool
		want    string
	}{
		{
			name:    "checked",
			checked: true,
			want:    "[x] ",
		},
		{
			name:    "unchecked",
			checked: false,
			want:    "[ ] ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			checkbox := &extast.TaskCheckBox{
				IsChecked: tt.checked,
			}

			status, err := r.renderTaskCheckBox(checkbox, true)
			if err != nil {
				t.Errorf("renderTaskCheckBox() error = %v", err)
			}
			if status != ast.WalkContinue {
				t.Errorf("renderTaskCheckBox() status = %v, want WalkContinue", status)
			}

			output := r.buf.String()
			if output != tt.want {
				t.Errorf("renderTaskCheckBox() output = %q, want %q", output, tt.want)
			}
		})
	}
}

func TestESCPOSRenderer_renderThematicBreak(t *testing.T) {
	r := NewESCPOSRenderer("/tmp", 75)

	status, err := r.renderThematicBreak(true)
	if err != nil {
		t.Errorf("renderThematicBreak() error = %v", err)
	}
	if status != ast.WalkSkipChildren {
		t.Errorf("renderThematicBreak() status = %v, want WalkSkipChildren", status)
	}

	output := r.buf.Bytes()
	if len(output) == 0 {
		t.Error("renderThematicBreak() produced no output")
	}

	// Should contain solid line command
	if !containsBytes(output, []byte{0x1D, 0x76, 0x30, 0x00}) {
		t.Error("renderThematicBreak() missing solid line bitmap command")
	}
}

func TestESCPOSRenderer_renderText(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "simple text",
			source: "Hello World",
			want:   "Hello World",
		},
		{
			name:   "text with leading space at column 0",
			source: "  test",
			want:   "test", // Leading space should be stripped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			r.currentColumn = 0

			textNode := ast.NewText()
			textNode.Segment = text.NewSegment(0, len(tt.source))

			status, err := r.renderText(textNode, []byte(tt.source))
			if err != nil {
				t.Errorf("renderText() error = %v", err)
			}
			if status != ast.WalkSkipChildren {
				t.Errorf("renderText() status = %v, want WalkSkipChildren", status)
			}

			output := r.buf.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("renderText() output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestESCPOSRenderer_writeWrappedText(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		startColumn   int
		wantContains  string
		wantLineBreak bool
	}{
		{
			name:         "short text",
			text:         "Hello",
			startColumn:  0,
			wantContains: "Hello",
			wantLineBreak: false,
		},
		{
			name:         "text exceeding width",
			text:         "This is a very long line that definitely exceeds the maximum width of 46 characters",
			startColumn:  0,
			wantContains: "This",
			wantLineBreak: true,
		},
		{
			name:         "text at end of line",
			text:         "test",
			startColumn:  43,
			wantContains: "test",
			wantLineBreak: true, // Should wrap because 43 + 1 (space) + 4 (test) > 46
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			r.currentColumn = tt.startColumn

			r.writeWrappedText(tt.text)

			output := r.buf.String()
			if !strings.Contains(output, tt.wantContains) {
				t.Errorf("writeWrappedText() output = %q, want to contain %q", output, tt.wantContains)
			}

			if tt.wantLineBreak {
				if !strings.Contains(output, "\r\n") && !strings.Contains(output, "\n") {
					t.Error("writeWrappedText() should contain line break")
				}
			}
		})
	}
}

func TestESCPOSRenderer_resetColumn(t *testing.T) {
	r := NewESCPOSRenderer("/tmp", 75)
	r.currentColumn = 25

	r.resetColumn()

	if r.currentColumn != 0 {
		t.Errorf("resetColumn() currentColumn = %d, want 0", r.currentColumn)
	}
}

func TestESCPOSRenderer_renderHTMLBlock(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		wantLineBreak bool
	}{
		{
			name:         "br tag",
			html:         "<br>",
			wantLineBreak: true,
		},
		{
			name:         "BR tag uppercase",
			html:         "<BR>",
			wantLineBreak: true,
		},
		{
			name:         "br self-closing",
			html:         "<br/>",
			wantLineBreak: true,
		},
		{
			name:         "other html",
			html:         "<div>test</div>",
			wantLineBreak: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			r.currentColumn = 10

			htmlBlock := ast.NewHTMLBlock(ast.HTMLBlockType1)
			htmlBlock.Lines().Append(text.NewSegment(0, len(tt.html)))

			status, err := r.renderHTMLBlock(htmlBlock, []byte(tt.html), true)
			if err != nil {
				t.Errorf("renderHTMLBlock() error = %v", err)
			}
			if status != ast.WalkSkipChildren {
				t.Errorf("renderHTMLBlock() status = %v, want WalkSkipChildren", status)
			}

			if tt.wantLineBreak {
				if r.currentColumn != 0 {
					t.Error("renderHTMLBlock() should reset column after <br>")
				}
				if !r.afterLineBreak {
					t.Error("renderHTMLBlock() should set afterLineBreak flag")
				}
			}
		})
	}
}

func TestESCPOSRenderer_renderRawHTML(t *testing.T) {
	tests := []struct {
		name         string
		html         string
		wantLineBreak bool
	}{
		{
			name:         "inline br",
			html:         "<br>",
			wantLineBreak: true,
		},
		{
			name:         "inline br with slash",
			html:         "<br />",
			wantLineBreak: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			r.currentColumn = 10

			rawHTML := ast.NewRawHTML()
			rawHTML.Segments.Append(text.NewSegment(0, len(tt.html)))

			status, err := r.renderRawHTML(rawHTML, []byte(tt.html), true)
			if err != nil {
				t.Errorf("renderRawHTML() error = %v", err)
			}
			if status != ast.WalkContinue {
				t.Errorf("renderRawHTML() status = %v, want WalkContinue", status)
			}

			if tt.wantLineBreak && r.currentColumn != 0 {
				t.Error("renderRawHTML() should reset column after <br>")
			}
		})
	}
}

func TestESCPOSRenderer_Integration(t *testing.T) {
	// Integration test with full markdown parsing
	markdownSource := `# Test Document

This is **bold** and *italic* text.

## List

- Item 1
- Item 2

## Ordered

1. First
2. Second

---

Done!`

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)

	reader := text.NewReader([]byte(markdownSource))
	doc := md.Parser().Parse(reader)

	renderer := NewESCPOSRenderer("/tmp", 75)
	var buf bytes.Buffer
	writer := &bufWriter{buf: &buf}

	err := renderer.Render(writer, []byte(markdownSource), doc)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.Bytes()

	// Check for key elements
	expectedElements := [][]byte{
		escpos.Init(),
		[]byte("Test Document"),
		escpos.BoldOn(),
		[]byte("bold"),
		escpos.UnderlineOn(),
		[]byte("italic"),
		[]byte("Item 1"),
		[]byte("First"),
		escpos.Cut(),
	}

	for _, elem := range expectedElements {
		if !containsBytes(output, elem) {
			t.Errorf("Integration test output missing element: %v", elem)
		}
	}
}

func TestESCPOSRenderer_AddOptions(t *testing.T) {
	r := NewESCPOSRenderer("/tmp", 75)

	// AddOptions should not panic and should be a no-op
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AddOptions() panicked: %v", r)
		}
	}()

	r.AddOptions()
}
