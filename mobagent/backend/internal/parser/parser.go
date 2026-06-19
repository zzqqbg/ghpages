package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/ghpages/mobagent/backend/internal/models"
)

type RawLine struct {
	Text      string
	Timestamp time.Time
}

type Parser interface {
	Parse(line RawLine) (*models.AgentEvent, bool)
}

type UniversalParser struct {
	agent     string
	taskID    string
	sessionID string
	workspace string
	lastKey   string
	lastCount int
}

func NewUniversalParser(agent, taskID, sessionID, workspace string) *UniversalParser {
	return &UniversalParser{
		agent: agent, taskID: taskID, sessionID: sessionID, workspace: workspace,
	}
}

var (
	reReading  = regexp.MustCompile(`(?i)(reading|read)\s+(.+)`)
	reSearch   = regexp.MustCompile(`(?i)(search(ing)?|grep)\s+(.+)`)
	reEdit     = regexp.MustCompile(`(?i)(edit(ing)?|writ(ing|e)|updat(ing|e))\s+(.+)`)
	reTest     = regexp.MustCompile(`(?i)(test(ing)?|npm run test|vitest|pytest)`)
	rePlan     = regexp.MustCompile(`(?i)(plan(n)?ing|think(ing)?|reason(ing)?)`)
	reBuild    = regexp.MustCompile(`(?i)(build(ing)?|compil(e|ing))`)
	reInstall  = regexp.MustCompile(`(?i)(install(ing)?|pnpm|npm install|pip install)`)
	reReview   = regexp.MustCompile(`(?i)(review(ing)?|self[- ]review)`)
	reGitCommit = regexp.MustCompile(`(?i)(git commit|commit(ted)?)`)
	reGitPush  = regexp.MustCompile(`(?i)(git push|push(ed)?)`)
	reError    = regexp.MustCompile(`(?i)(error|failed|panic|fatal)`)
	reDone     = regexp.MustCompile(`(?i)(done|finished|complete(d)?|success)`)
)

func (p *UniversalParser) Parse(line RawLine) (*models.AgentEvent, bool) {
	text := strings.TrimSpace(line.Text)
	if text == "" {
		return nil, false
	}

	ev := &models.AgentEvent{
		TaskID: p.taskID, SessionID: p.sessionID, Agent: p.agent, Workspace: p.workspace,
		Timestamp: line.Timestamp, Status: models.StatusActive,
		Payload: map[string]interface{}{"raw": text},
	}

	switch {
	case reError.MatchString(text):
		ev.Type = models.EventError
		ev.Status = models.StatusFailed
		ev.Payload["message"] = text
	case reDone.MatchString(text):
		ev.Type = models.EventFinished
		ev.Status = models.StatusCompleted
		ev.Payload["message"] = text
	case rePlan.MatchString(text):
		ev.Type = models.EventPlanning
		ev.Payload["message"] = "Planning implementation..."
	case reSearch.MatchString(text):
		ev.Type = models.EventSearching
		m := reSearch.FindStringSubmatch(text)
		ev.Payload["query"] = strings.TrimSpace(m[len(m)-1])
	case reReading.MatchString(text):
		path := strings.TrimSpace(reReading.FindStringSubmatch(text)[2])
		if strings.HasSuffix(path, ".json") || strings.Contains(path, "package") {
			ev.Type = models.EventReadingProject
			ev.Payload["message"] = "Reading project..."
		} else {
			ev.Type = models.EventReadingFile
			ev.Payload["path"] = path
		}
	case reEdit.MatchString(text):
		ev.Type = models.EventEditing
		m := reEdit.FindStringSubmatch(text)
		path := strings.TrimSpace(m[len(m)-1])
		ev.Payload["path"] = path
		ev.Payload["files"] = []map[string]interface{}{{"path": path, "additions": 1, "deletions": 0}}
	case reTest.MatchString(text):
		ev.Type = models.EventRunningTests
		ev.Payload["message"] = "Running tests..."
	case reBuild.MatchString(text):
		ev.Type = models.EventBuilding
		ev.Payload["message"] = "Building..."
	case reInstall.MatchString(text):
		ev.Type = models.EventInstallingDeps
		ev.Payload["message"] = "Installing dependencies..."
	case reReview.MatchString(text):
		ev.Type = models.EventReviewing
		ev.Payload["message"] = "Code review in progress..."
	case reGitCommit.MatchString(text):
		ev.Type = models.EventGitCommit
		ev.Payload["message"] = text
	case reGitPush.MatchString(text):
		ev.Type = models.EventGitPush
		ev.Payload["message"] = text
	default:
		ev.Type = models.EventConsoleOutput
		ev.Payload["line"] = text
	}

	key := string(ev.Type) + dedupeKey(ev)
	if key == p.lastKey {
		p.lastCount++
		if p.lastCount > 2 && ev.Type != models.EventConsoleOutput {
			return nil, false
		}
	} else {
		p.lastKey = key
		p.lastCount = 1
	}
	return ev, true
}

func dedupeKey(ev *models.AgentEvent) string {
	if ev.Payload == nil {
		return ""
	}
	for _, k := range []string{"path", "query", "message", "line"} {
		if v, ok := ev.Payload[k]; ok {
			return stringOf(v)
		}
	}
	return ""
}

func stringOf(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}
