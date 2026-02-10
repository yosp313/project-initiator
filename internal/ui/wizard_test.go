package ui

import (
	"slices"
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

func TestFrameworkDescription(t *testing.T) {
	tests := []struct {
		name      string
		language  string
		framework string
		want      string
	}{
		{"vanilla", "Go", "Vanilla", "minimal starter"},
		{"vanilla lowercase", "Go", "vanilla", "minimal starter"},
		{"cobra", "Go", "Cobra", "CLI app structure"},
		{"express", "JavaScript", "Express", "Node.js web server"},
		{"hono", "JavaScript", "Hono", "lightweight web framework"},
		{"nestjs", "TypeScript", "NestJS", "typed Node framework"},
		{"bun", "TypeScript", "Bun", "Bun runtime server"},
		{"fastapi", "Python", "FastAPI", "Python API server"},
		{"laravel", "PHP", "Laravel", "PHP web framework"},
		{"unknown framework uses language name", "Rust", "Actix", "Rust template"},
		{"unknown framework different language", "Elixir", "Phoenix", "Elixir template"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frameworkDescription(tt.language, tt.framework)
			if got != tt.want {
				t.Errorf("frameworkDescription(%q, %q) = %q, want %q", tt.language, tt.framework, got, tt.want)
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "no duplicates",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "exact duplicates",
			input: []string{"a", "b", "a", "c", "b"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "case insensitive dedup keeps first occurrence",
			input: []string{"Go", "go", "GO"},
			want:  []string{"Go"},
		},
		{
			name:  "empty list",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "nil list",
			input: nil,
			want:  []string{},
		},
		{
			name:  "whitespace only entries are removed",
			input: []string{"a", "  ", "b", "\t", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "entries with leading trailing whitespace dedup on trimmed key",
			input: []string{"  Go ", "Go"},
			want:  []string{"  Go "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniqueStrings(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("uniqueStrings(%v) returned %d elements, want %d: got %v", tt.input, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("uniqueStrings(%v)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSortStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "already sorted",
			input: []string{"alpha", "beta", "gamma"},
			want:  []string{"alpha", "beta", "gamma"},
		},
		{
			name:  "reverse order",
			input: []string{"gamma", "beta", "alpha"},
			want:  []string{"alpha", "beta", "gamma"},
		},
		{
			name:  "case insensitive sort",
			input: []string{"Banana", "apple", "Cherry"},
			want:  []string{"apple", "Banana", "Cherry"},
		},
		{
			name:  "single element",
			input: []string{"only"},
			want:  []string{"only"},
		},
		{
			name:  "empty",
			input: []string{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// copy to avoid mutating test data
			values := make([]string, len(tt.input))
			copy(values, tt.input)

			sortStrings(values)

			if len(values) != len(tt.want) {
				t.Fatalf("sortStrings(%v) produced %d elements, want %d", tt.input, len(values), len(tt.want))
			}
			for i := range values {
				if values[i] != tt.want[i] {
					t.Errorf("sortStrings(%v)[%d] = %q, want %q", tt.input, i, values[i], tt.want[i])
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		target string
		want   bool
	}{
		{"found exact", []string{"Go", "Python", "Rust"}, "Python", true},
		{"not found", []string{"Go", "Python", "Rust"}, "Java", false},
		{"case insensitive match", []string{"Go", "Python"}, "go", true},
		{"case insensitive match reverse", []string{"go"}, "Go", true},
		{"empty list", []string{}, "Go", false},
		{"nil list", nil, "Go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.values, tt.target)
			if got != tt.want {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.values, tt.target, got, tt.want)
			}
		})
	}
}

func TestSelectedLibraries(t *testing.T) {
	tests := []struct {
		name     string
		selected map[string]bool
		want     []string
	}{
		{
			name:     "all true",
			selected: map[string]bool{"zap": true, "cobra": true, "viper": true},
			want:     []string{"cobra", "viper", "zap"}, // sorted
		},
		{
			name:     "all false",
			selected: map[string]bool{"zap": false, "cobra": false},
			want:     []string{},
		},
		{
			name:     "mixed",
			selected: map[string]bool{"zap": true, "cobra": false, "viper": true},
			want:     []string{"viper", "zap"},
		},
		{
			name:     "empty map",
			selected: map[string]bool{},
			want:     []string{},
		},
		{
			name:     "nil map",
			selected: nil,
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectedLibraries(tt.selected)
			if len(got) != len(tt.want) {
				t.Fatalf("selectedLibraries(%v) returned %d elements, want %d: got %v", tt.selected, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("selectedLibraries(%v)[%d] = %q, want %q", tt.selected, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSelectedLibraries_Sorted(t *testing.T) {
	selected := map[string]bool{
		"zap":   true,
		"cobra": true,
		"viper": true,
		"air":   true,
	}
	got := selectedLibraries(selected)
	if !slices.IsSorted(got) {
		t.Errorf("selectedLibraries result is not sorted: %v", got)
	}
}

func TestStageTitle(t *testing.T) {
	tests := []struct {
		stage stage
		want  string
	}{
		{stageLanguage, "Choose a language"},
		{stageFramework, "Choose a framework"},
		{stageLibraries, "Choose libraries"},
		{stageName, "Name your project"},
		{stageConfirm, "Confirm your selections"},
		{stageDone, ""},
		{stage(99), ""}, // unknown stage
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := stageTitle(tt.stage)
			if got != tt.want {
				t.Errorf("stageTitle(%d) = %q, want %q", tt.stage, got, tt.want)
			}
		})
	}
}

func TestStageSubtitle(t *testing.T) {
	tests := []struct {
		stage stage
		want  string
	}{
		{stageLanguage, "Pick the main language for the starter"},
		{stageFramework, "Select the starter template"},
		{stageLibraries, "Select optional packages (space to toggle)"},
		{stageName, "This will create the folder name"},
		{stageConfirm, "Review before creating the project"},
		{stageDone, ""},
		{stage(99), ""}, // unknown stage
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := stageSubtitle(tt.stage)
			if got != tt.want {
				t.Errorf("stageSubtitle(%d) = %q, want %q", tt.stage, got, tt.want)
			}
		})
	}
}

func TestStageProgress(t *testing.T) {
	tests := []struct {
		name    string
		stage   stage
		hasLibs bool
		want    float64
	}{
		{"language", stageLanguage, false, 0.0},
		{"framework no libs", stageFramework, false, 1.0 / 3.0},
		{"framework with libs", stageFramework, true, 1.0 / 4.0},
		{"libraries", stageLibraries, true, 2.0 / 4.0},
		{"name no libs", stageName, false, 2.0 / 3.0},
		{"name with libs", stageName, true, 3.0 / 4.0},
		{"confirm", stageConfirm, false, 1.0},
		{"done", stageDone, false, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{stage: tt.stage}
			if tt.hasLibs {
				m.libraries = newCleanList([]list.Item{listItem{label: "test", description: "d"}}, listDelegate{}, 0, 0)
			} else {
				m.libraries = newCleanList([]list.Item{}, listDelegate{}, 0, 0)
			}
			got := m.stageProgress()
			if got != tt.want {
				t.Errorf("stageProgress() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestTriggerTransition(t *testing.T) {
	tests := []struct {
		name    string
		panelW  int
		forward bool
		wantPos bool // true = positive offset, false = negative
	}{
		{"forward default", 0, true, true},
		{"backward default", 0, false, false},
		{"forward with panel", 100, true, true},
		{"backward with panel", 100, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{panelW: tt.panelW}
			m.triggerTransition(tt.forward)
			if !m.transActive {
				t.Error("transActive should be true after triggerTransition")
			}
			if m.transVel != 0 {
				t.Errorf("transVel should be 0, got %f", m.transVel)
			}
			if tt.wantPos && m.transOffset <= 0 {
				t.Errorf("expected positive offset, got %f", m.transOffset)
			}
			if !tt.wantPos && m.transOffset >= 0 {
				t.Errorf("expected negative offset, got %f", m.transOffset)
			}
		})
	}
}

func TestAbsF(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"positive", 3.5, 3.5},
		{"negative", -3.5, 3.5},
		{"zero", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := absF(tt.input)
			if got != tt.want {
				t.Errorf("absF(%f) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ASCII art & animation tests
// ---------------------------------------------------------------------------

func TestAsciiArt_AllNonEmptyLinesEqualWidth(t *testing.T) {
	art := asciiArt()
	if len(art) != 9 {
		t.Fatalf("asciiArt() returned %d lines, want 9", len(art))
	}

	// Collect widths of non-empty lines
	var width int
	for i, line := range art {
		if line == "" {
			continue // blank separator is allowed
		}
		n := runeLen(line)
		if width == 0 {
			width = n
		}
		if n != width {
			t.Errorf("asciiArt() line %d has %d runes, want %d", i, n, width)
		}
	}
}

func TestAsciiArt_BlankSeparatorAtLine4(t *testing.T) {
	art := asciiArt()
	if art[4] != "" {
		t.Errorf("asciiArt() line 4 should be blank separator, got %q", art[4])
	}
}

func TestAsciiArt_NonEmptyLinesContainBlockChars(t *testing.T) {
	art := asciiArt()
	for i, line := range art {
		if line == "" {
			continue
		}
		hasBlock := false
		for _, r := range line {
			if r == '▄' || r == '█' || r == '▀' {
				hasBlock = true
				break
			}
		}
		if !hasBlock {
			t.Errorf("asciiArt() line %d contains no block characters: %q", i, line)
		}
	}
}

func TestArtWidth(t *testing.T) {
	w := artWidth()
	if w != 46 {
		t.Errorf("artWidth() = %d, want 46", w)
	}
}

func TestRuneLen(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"ascii", "hello", 5},
		{"unicode blocks", "▄███▄", 5},
		{"mixed", "▄█ ab", 5},
		{"spaces", "   ", 3},
		{"emoji", "🎉🎊", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runeLen(tt.input)
			if got != tt.want {
				t.Errorf("runeLen(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestRevealTotalTicks(t *testing.T) {
	w := artWidth()
	expected := w / revealColumns
	if w%revealColumns != 0 {
		expected++
	}
	got := revealTotalTicks()
	if got != expected {
		t.Errorf("revealTotalTicks() = %d, want %d (artWidth=%d, revealColumns=%d)", got, expected, w, revealColumns)
	}

	// Also verify the value makes sense: ~16 ticks at 150ms ≈ 2.4s reveal
	if got < 10 || got > 30 {
		t.Errorf("revealTotalTicks() = %d, expected reasonable range 10-30", got)
	}
}

func TestRenderAnimatedBorder_EmptyForSmallWidth(t *testing.T) {
	cache := buildAnimCache(defaultStyles())
	result := renderAnimatedBorder(1, 0, cache)
	if result != "" {
		t.Errorf("renderAnimatedBorder(1, 0) should be empty, got %q", result)
	}
	result = renderAnimatedBorder(0, 0, cache)
	if result != "" {
		t.Errorf("renderAnimatedBorder(0, 0) should be empty, got %q", result)
	}
}

func TestRenderAnimatedBorder_NonEmptyForValidWidth(t *testing.T) {
	cache := buildAnimCache(defaultStyles())
	for _, width := range []int{2, 10, 46, 80} {
		result := renderAnimatedBorder(width, 0, cache)
		if result == "" {
			t.Errorf("renderAnimatedBorder(%d, 0) should not be empty", width)
		}
	}
}

func TestRenderAnimatedBorder_VariousFramesDoNotPanic(t *testing.T) {
	cache := buildAnimCache(defaultStyles())
	width := 40
	// Ensure the function handles a full cycle of spark positions without panicking
	// and always returns a non-empty string.
	cycle := width + 5
	for frame := 0; frame < cycle; frame++ {
		result := renderAnimatedBorder(width, frame, cache)
		if result == "" {
			t.Errorf("renderAnimatedBorder(%d, %d) should not be empty", width, frame)
		}
	}
}

func TestRenderAnimatedTitle_NonEmpty(t *testing.T) {
	s := defaultStyles()
	m := model{
		styles:     s,
		animCache:  buildAnimCache(s),
		titleFrame: 0,
	}
	result := m.renderAnimatedTitle(60)
	if result == "" {
		t.Error("renderAnimatedTitle(60) should not be empty")
	}
}

func TestRenderAnimatedTitle_SmallWidth(t *testing.T) {
	s := defaultStyles()
	m := model{
		styles:     s,
		animCache:  buildAnimCache(s),
		titleFrame: 10,
	}
	// Width smaller than art (46 chars) — should still render without panic.
	for _, w := range []int{1, 3, 20, 30, 45} {
		result := m.renderAnimatedTitle(w)
		if result == "" {
			t.Errorf("renderAnimatedTitle(%d) should not be empty", w)
		}
	}
}

func TestRenderAnimatedTitle_ContainsBorderChars(t *testing.T) {
	s := defaultStyles()
	m := model{
		styles:     s,
		animCache:  buildAnimCache(s),
		titleFrame: 20, // fully revealed
	}
	result := m.renderAnimatedTitle(60)
	if !containsRune(result, '═') && !containsRune(result, '╾') && !containsRune(result, '╼') {
		t.Error("renderAnimatedTitle should contain border characters")
	}
}

func TestRenderAnimatedTitle_FullRevealContainsArtChars(t *testing.T) {
	s := defaultStyles()
	m := model{
		styles:     s,
		animCache:  buildAnimCache(s),
		titleFrame: 100, // well past full reveal
	}
	result := m.renderAnimatedTitle(60)
	// After full reveal, the art block characters should be present
	if !containsRune(result, '█') {
		t.Error("renderAnimatedTitle at full reveal should contain block characters from ASCII art")
	}
}

func containsRune(s string, target rune) bool {
	for _, r := range s {
		if r == target {
			return true
		}
	}
	return false
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		value int
		min   int
		max   int
		want  int
	}{
		{"value in range", 5, 0, 10, 5},
		{"below min", -3, 0, 10, 0},
		{"above max", 15, 0, 10, 10},
		{"at min boundary", 0, 0, 10, 0},
		{"at max boundary", 10, 0, 10, 10},
		{"min equals max value matches", 5, 5, 5, 5},
		{"min equals max value below", 3, 5, 5, 5},
		{"min equals max value above", 7, 5, 5, 5},
		{"negative range in range", -5, -10, -1, -5},
		{"negative range below", -15, -10, -1, -10},
		{"negative range above", 0, -10, -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%d, %d, %d) = %d, want %d", tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}
