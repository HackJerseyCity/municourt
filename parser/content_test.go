package parser

import (
	"testing"
)

func TestExtractTextItems_TJKerning(t *testing.T) {
	// Simulates a TJ array with kerning-based concatenation.
	// (8)0(8) should concatenate to "88", and -4704.6 should separate.
	stream := []byte(`BT
[(8)0(8)-4704.6(2)0(3)]TJ
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	// Filter out empty line-break markers.
	var nonEmpty []string
	for _, s := range items {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	if len(nonEmpty) != 2 {
		t.Fatalf("expected 2 items, got %d: %v", len(nonEmpty), nonEmpty)
	}
	if nonEmpty[0] != "88" {
		t.Errorf("expected first item '88', got %q", nonEmpty[0])
	}
	if nonEmpty[1] != "23" {
		t.Errorf("expected second item '23', got %q", nonEmpty[1])
	}
}

func TestExtractTextItems_Tj(t *testing.T) {
	stream := []byte(`BT
(Hello World)Tj
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	var nonEmpty []string
	for _, s := range items {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	if len(nonEmpty) != 1 {
		t.Fatalf("expected 1 item, got %d: %v", len(nonEmpty), nonEmpty)
	}
	if nonEmpty[0] != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", nonEmpty[0])
	}
}

func TestExtractTextItems_TDLineBreaks(t *testing.T) {
	stream := []byte(`BT
(Line1)Tj
0 -12 TD
(Line2)Tj
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	// Should have: "", "Line1", "", "Line2" (with line-break markers).
	var foundLine1, foundLine2 bool
	var breakBetween bool
	for i, s := range items {
		if s == "Line1" {
			foundLine1 = true
		}
		if s == "Line2" {
			foundLine2 = true
		}
		// Check there's a break between Line1 and Line2.
		if foundLine1 && !foundLine2 && s == "" && i > 0 {
			breakBetween = true
		}
	}

	if !foundLine1 || !foundLine2 {
		t.Errorf("expected both Line1 and Line2, got items: %v", items)
	}
	if !breakBetween {
		t.Errorf("expected line break between Line1 and Line2, got items: %v", items)
	}
}

func TestExtractTextItems_SmallKerningConcatenates(t *testing.T) {
	// Small kerning values (abs <= 500) should concatenate strings.
	stream := []byte(`BT
[(H)-50(e)-30(l)(l)(o)]TJ
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	var nonEmpty []string
	for _, s := range items {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	if len(nonEmpty) != 1 {
		t.Fatalf("expected 1 item, got %d: %v", len(nonEmpty), nonEmpty)
	}
	if nonEmpty[0] != "Hello" {
		t.Errorf("expected 'Hello', got %q", nonEmpty[0])
	}
}

func TestExtractTextItems_MixedTjAndTJ(t *testing.T) {
	stream := []byte(`BT
(MUNICIPAL COURT STATISTICS)Tj
2.1882 -1.4941 TD
(JULY 2023 - JUNE 2024)Tj
3.0706 -1.4941 TD
(ATLANTIC)Tj
-.0118 -1.4941 TD
(ABSECON)Tj
0 8.52 -8.52 0 101.52 285.96 Tm
[(D.P. &)-3012.9(Other)-2811.9(Criminal)]TJ
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	var nonEmpty []string
	for _, s := range items {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	expected := []string{
		"MUNICIPAL COURT STATISTICS",
		"JULY 2023 - JUNE 2024",
		"ATLANTIC",
		"ABSECON",
		"D.P. &", "Other", "Criminal",
	}

	if len(nonEmpty) != len(expected) {
		t.Fatalf("expected %d items, got %d: %v", len(expected), len(nonEmpty), nonEmpty)
	}
	for i, exp := range expected {
		if nonEmpty[i] != exp {
			t.Errorf("item %d: expected %q, got %q", i, exp, nonEmpty[i])
		}
	}
}

func TestExtractTextItems_TmSameLineShouldNotBreak(t *testing.T) {
	// Simulates a PDF where a municipality name is split across two Tm
	// operations on the same visual line (same y in page space for
	// non-rotated text). The second Tm repositions horizontally but stays
	// on the same line. This should NOT insert a line break.
	stream := []byte(`BT
1 0 0 1 72 700 Tm
(MUNICIPAL COURT STATISTICS)Tj
1 0 0 1 72 685 Tm
(JULY 2023 - JUNE 2024)Tj
1 0 0 1 72 670 Tm
(HUDSON)Tj
1 0 0 1 72 655 Tm
(Union Cit)Tj
1 0 0 1 140 655 Tm
(y)Tj
1 0 0 1 72 640 Tm
(Next Line)Tj
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	var nonEmpty []string
	for _, s := range items {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	expected := []string{
		"MUNICIPAL COURT STATISTICS",
		"JULY 2023 - JUNE 2024",
		"HUDSON",
		"Union Cit", "y",
		"Next Line",
	}

	if len(nonEmpty) != len(expected) {
		t.Fatalf("expected %d items, got %d: %v", len(expected), len(nonEmpty), nonEmpty)
	}
	for i, exp := range expected {
		if nonEmpty[i] != exp {
			t.Errorf("item %d: expected %q, got %q", i, exp, nonEmpty[i])
		}
	}

	// Critically: "Union Cit" and "y" must be on the SAME line (no line
	// break between them), so groupIntoLines would put them together.
	lines := groupIntoLines(items)
	var muniLine []string
	for _, line := range lines {
		if len(line) > 0 && line[0] == "Union Cit" {
			muniLine = line
			break
		}
	}
	if len(muniLine) != 2 || muniLine[0] != "Union Cit" || muniLine[1] != "y" {
		t.Errorf("expected municipality line [\"Union Cit\" \"y\"], got %v", muniLine)
		t.Logf("all lines: %v", lines)
	}
}

func TestExtractTextItems_TmSameLineRotated(t *testing.T) {
	// Same test but with rotated text matrix (as seen in real municipal court PDFs).
	// Rotated 90°: a=0 b=10.2 c=-10.2 d=0. Same line = same e (x in page space).
	stream := []byte(`BT
0 10.2 -10.2 0 34.68 317.52 Tm
(MUNICIPAL COURT STATISTICS)Tj
0 10.2 -10.2 0 34.68 302.58 Tm
(JULY 2023 - JUNE 2024)Tj
0 10.2 -10.2 0 49.92 317.52 Tm
(HUDSON)Tj
0 10.2 -10.2 0 65.16 317.52 Tm
(Union Cit)Tj
0 10.2 -10.2 0 65.16 400.00 Tm
(y)Tj
0 10.2 -10.2 0 80.40 317.52 Tm
(Next Line)Tj
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	// "Union Cit" and "y" should be on the same line (same e=65.16).
	lines := groupIntoLines(items)
	var muniLine []string
	for _, line := range lines {
		if len(line) > 0 && line[0] == "Union Cit" {
			muniLine = line
			break
		}
	}
	if len(muniLine) != 2 || muniLine[0] != "Union Cit" || muniLine[1] != "y" {
		t.Errorf("expected municipality line [\"Union Cit\" \"y\"], got %v", muniLine)
		t.Logf("all lines: %v", lines)
	}
}

func TestExtractTextItems_ClippedTextAcrossBTET(t *testing.T) {
	// Real-world bug: "ATLANTIC" rendered as "ATLANTI" in one BT..ET block,
	// then "C" in a separate clipped BT..ET block. The clipped block has a
	// slightly different page position but is visually on the same line.
	// This reproduces the actual content stream from municipal-courts-2022-06.pdf page 5.
	stream := []byte(`BT
0 10.2 -10.2 0 32.64 317.52 Tm
(MUNICIPAL COURT STATISTICS)Tj
2.1882 -1.2941 TD
(JULY 2021 - JUNE 2022)Tj
3.0706 -1.2941 TD
(ATLANTI)Tj
ET
q
1 i
49.8 371.16 11.52 50.4 re
W n
BT
0 10.2 -10.2 0 59.04 414.7839 Tm
(C)Tj
ET
Q
BT
0 10.2 -10.2 0 72.24 364.8 Tm
(BRIGANTINE)Tj
ET`)

	items := ExtractTextItems(PageData{Content: stream})
	lines := groupIntoLines(items)

	// Find the line containing "ATLANTI"
	var countyLine []string
	var countyIdx int
	for i, line := range lines {
		for _, item := range line {
			if item == "ATLANTI" {
				countyLine = line
				countyIdx = i
				break
			}
		}
		if countyLine != nil {
			break
		}
	}

	// "ATLANTI" and "C" must be on the same line
	if len(countyLine) != 2 || countyLine[0] != "ATLANTI" || countyLine[1] != "C" {
		t.Errorf("expected county line [\"ATLANTI\" \"C\"], got %v", countyLine)
		t.Logf("all lines: %v", lines)
	}

	// "BRIGANTINE" should be the next line
	if countyIdx+1 < len(lines) {
		nextLine := lines[countyIdx+1]
		if len(nextLine) != 1 || nextLine[0] != "BRIGANTINE" {
			t.Errorf("expected next line [\"BRIGANTINE\"], got %v", nextLine)
		}
	} else {
		t.Errorf("no line after county line")
	}
}

func TestTokenizeEscapedParens(t *testing.T) {
	stream := []byte(`BT
(\(moving\))Tj
ET`)

	items := ExtractTextItems(PageData{Content: stream})

	var nonEmpty []string
	for _, s := range items {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	if len(nonEmpty) != 1 {
		t.Fatalf("expected 1 item, got %d: %v", len(nonEmpty), nonEmpty)
	}
	if nonEmpty[0] != "(moving)" {
		t.Errorf("expected '(moving)', got %q", nonEmpty[0])
	}
}
