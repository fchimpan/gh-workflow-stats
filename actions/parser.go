package actions

import "github.com/google/go-github/v57/github"

func ParseWorkflowRuns(wr []*github.WorkflowRun) (resp *WorkflowRuns) {

	if len(wr) == 0 {
		return &WorkflowRuns{}
	}
	resp.ID = *wr[0].ID
	resp.Name = *wr[0].Name
	resp.WorkflowRunsCount = len(wr)

	at := 0
	// wrc := make(map[string]RunsConclusion)
	for _, r := range wr {
		if r == nil {
			continue
		}
		at += r.GetRunAttempt()
		c := r.GetConclusion()
		if c == "" {
			continue
		} else {
		}

	}
	resp.TotalWorkflowRunsCount = at
	return resp
}
