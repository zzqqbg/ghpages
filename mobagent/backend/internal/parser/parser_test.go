package parser_test

import (
	"testing"
	"time"

	"github.com/ghpages/mobagent/backend/internal/models"
	"github.com/ghpages/mobagent/backend/internal/parser"
)

func TestParserDedupeReading(t *testing.T) {
	t.Parallel()
	p := parser.NewUniversalParser("cursor", "t1", "s1", "/ws")
	line := parser.RawLine{Text: "Reading package.json", Timestamp: time.Now()}
	var got int
	for i := 0; i < 5; i++ {
		if ev, ok := p.Parse(line); ok {
			got++
			if ev.Type != models.EventReadingProject {
				t.Fatalf("type=%s", ev.Type)
			}
		}
	}
	if got > 3 {
		t.Fatalf("too many events %d", got)
	}
}

func TestParserError(t *testing.T) {
	t.Parallel()
	p := parser.NewUniversalParser("cursor", "t1", "s1", "/ws")
	ev, ok := p.Parse(parser.RawLine{Text: "fatal error: boom", Timestamp: time.Now()})
	if !ok || ev.Type != models.EventError {
		t.Fatalf("expected error event")
	}
}
