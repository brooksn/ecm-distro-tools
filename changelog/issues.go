package changelog

import (
	"context"
	"strings"

	"github.com/google/go-github/v39/github"
)

type Issue struct {
	Title  string
	Note   string
	Number int
	URL    string
}

const (
	releaseNoteSection = "```release-note"
	emptyReleaseNote   = "```release-note\r\n\r\n```"
	noneReleaseNote    = "```release-note\r\nNONE\r\n```"
)

// stripBackportTag returns a string with a prefix backport tag removed
func stripBackportTag(s string) string {
	if strings.Contains(s, "Release") || strings.Contains(s, "release") && strings.Contains(s, "[") || strings.Contains(s, "]") {
		s = strings.Split(s, "]")[1]
	}
	s = strings.Trim(s, " ")
	return s
}

// Issues requests the GitHub Pulls resource for each commit in the repo
func (r *Repo) Issues(ctx context.Context, gh *github.Client) ([]*Issue, error) {
	var found []*Issue
	addedPRs := make(map[int]bool)
	for _, commit := range r.Commits {
		sha := commit.GetSHA()
		if sha == "" {
			continue
		}

		prs, _, err := gh.PullRequests.ListPullRequestsWithCommit(ctx, r.Organization, r.Name, sha, &github.PullRequestListOptions{})
		if err != nil {
			return nil, err
		}
		if len(prs) == 1 {
			if exists := addedPRs[prs[0].GetNumber()]; exists {
				continue
			}

			title := stripBackportTag(strings.TrimSpace(prs[0].GetTitle()))
			body := prs[0].GetBody()

			var releaseNote string
			var inNote bool
			if strings.Contains(body, releaseNoteSection) && !strings.Contains(body, emptyReleaseNote) && !strings.Contains(body, noneReleaseNote) {
				lines := strings.Split(body, "\n")
				for _, line := range lines {
					if strings.Contains(line, releaseNoteSection) {
						inNote = true
						continue
					}
					if strings.Contains(line, "```") {
						inNote = false
					}
					if inNote && line != "" {
						releaseNote += line
					}
				}
				releaseNote = strings.TrimSpace(releaseNote)
				releaseNote = strings.ReplaceAll(releaseNote, "\r", "\n")
			}

			found = append(found, &Issue{
				Title:  title,
				Note:   releaseNote,
				Number: prs[0].GetNumber(),
				URL:    prs[0].GetHTMLURL(),
			})
			addedPRs[prs[0].GetNumber()] = true
		}
	}
	return found, nil
}
