package pipeline

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type RepositoryContext struct {
	Root           string                `json:"root"`
	ScannedDirs    []string              `json:"scannedDirs"`
	FileCount      int                   `json:"fileCount"`
	CandidateFiles []RepositoryFileMatch `json:"candidateFiles"`
}

type RepositoryFileMatch struct {
	Path       string   `json:"path"`
	MatchCount int      `json:"matchCount"`
	Keywords   []string `json:"keywords"`
	Summary    string   `json:"summary"`
}

func BuildRepositoryContext(requirement string) RepositoryContext {
	root := findWorkspaceRoot()
	context := RepositoryContext{Root: root, ScannedDirs: []string{"apps/api-go/internal", "apps/web/src", "docs"}}
	keywords := contextKeywords(requirement)
	matches := make([]RepositoryFileMatch, 0)

	for _, dir := range context.ScannedDirs {
		base := filepath.Join(root, dir)
		_ = filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			if !isContextFile(path) {
				return nil
			}
			context.FileCount++
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			match := scoreContextFile(rel, string(content), keywords)
			if match.MatchCount > 0 {
				matches = append(matches, match)
			}
			return nil
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].MatchCount == matches[j].MatchCount {
			return matches[i].Path < matches[j].Path
		}
		return matches[i].MatchCount > matches[j].MatchCount
	})
	if len(matches) > 8 {
		matches = matches[:8]
	}
	context.CandidateFiles = matches
	return context
}

func findWorkspaceRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	current := wd
	for {
		if exists(filepath.Join(current, "apps")) && exists(filepath.Join(current, "docs")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return wd
		}
		current = parent
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isContextFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".ts", ".tsx", ".md":
		return true
	default:
		return false
	}
}

func contextKeywords(requirement string) []string {
	base := []string{"pipeline", "stage", "checkpoint", "artifact", "agent", "review", "delivery", "需求", "方案", "检查点", "评审", "交付"}
	seen := map[string]bool{}
	keywords := make([]string, 0, len(base)+8)
	for _, item := range append(base, strings.Fields(strings.ToLower(requirement))...) {
		item = strings.Trim(item, "，。,.!?:;()[]{}\"'`")
		if item == "" || len([]rune(item)) < 2 || seen[item] {
			continue
		}
		seen[item] = true
		keywords = append(keywords, item)
	}
	return keywords
}

func scoreContextFile(path string, content string, keywords []string) RepositoryFileMatch {
	lowerPath := strings.ToLower(path)
	lowerContent := strings.ToLower(content)
	matched := make([]string, 0)
	count := 0
	for _, keyword := range keywords {
		lowerKeyword := strings.ToLower(keyword)
		pathHits := strings.Count(lowerPath, lowerKeyword)
		contentHits := strings.Count(lowerContent, lowerKeyword)
		if pathHits+contentHits == 0 {
			continue
		}
		count += pathHits*3 + contentHits
		matched = append(matched, keyword)
	}
	return RepositoryFileMatch{Path: filepath.ToSlash(path), MatchCount: count, Keywords: matched, Summary: contextFileSummary(path, matched)}
}

func contextFileSummary(path string, keywords []string) string {
	if len(keywords) == 0 {
		return "相关文件"
	}
	if len(keywords) > 4 {
		keywords = keywords[:4]
	}
	return path + " 命中 " + strings.Join(keywords, ", ")
}
