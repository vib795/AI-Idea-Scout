package llm

// ExtractedIdea represents an idea extracted by Claude
type ExtractedIdea struct {
	Title              string              `json:"title"`
	OneLiner           string              `json:"one_liner"`
	Problem            string              `json:"problem"`
	TargetUsers        string              `json:"target_users"`
	JobToBeDone        string              `json:"job_to_be_done"`
	DataRequirements   DataRequirements    `json:"data_requirements"`
	TechnicalApproach  []string            `json:"technical_approach"`
	Complexity         string              `json:"complexity"`
	Category           string              `json:"category"`
	Tags               []string            `json:"tags"`
	BuildabilityScore  int                 `json:"buildability_score"`
	DemandScore        int                 `json:"demand_score"`
	NoveltyScore       int                 `json:"novelty_score"`
	DistributionWedge  string              `json:"distribution_wedge"`
	Moat               string              `json:"moat"`
	Risks              []string            `json:"risks"`
	Evidence           []Evidence          `json:"evidence"`
	NextSteps          []string            `json:"next_steps"`
}

type DataRequirements struct {
	Inputs      []string `json:"inputs"`
	APIs        []string `json:"apis"`
	Permissions []string `json:"permissions"`
}

type Evidence struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
	URL   string `json:"url"`
}
