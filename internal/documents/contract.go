package documents

type Decision string

const (
	DecisionProposeDocs        Decision = "propose_docs"
	DecisionNeedsClarification Decision = "needs_clarification"
	DecisionNoDocsImpact       Decision = "no_docs_impact"
)

type Contract struct {
	Version            string             `json:"version"`
	Provider           Provider           `json:"provider"`
	WorkItem           WorkItem           `json:"work_item"`
	Change             Change             `json:"change"`
	Decision           Decision           `json:"decision"`
	Evidence           []Evidence         `json:"evidence"`
	Claims             []Claim            `json:"claims"`
	AffectedDocuments  []DocumentProposal `json:"affected_documents"`
	Conflicts          []Conflict         `json:"conflicts"`
	MissingInformation []string           `json:"missing_information"`
}

type Provider struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

type WorkItem struct {
	Number          int      `json:"number"`
	Title           string   `json:"title"`
	URL             string   `json:"url"`
	UserGoal        string   `json:"user_goal"`
	Audience        string   `json:"audience"`
	ExpectedOutcome string   `json:"expected_outcome"`
	Limitations     []string `json:"limitations"`
}

type Change struct {
	Feature                string          `json:"feature"`
	CustomerVisible        bool            `json:"customer_visible"`
	DocumentationAudiences []string        `json:"documentation_audiences"`
	Settings               []SettingChange `json:"settings"`
}

type SettingChange struct {
	Name string `json:"name"`
	Old  string `json:"old,omitempty"`
	New  string `json:"new,omitempty"`
}

type Evidence struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Source string `json:"source"`
	Detail string `json:"detail"`
}

type Claim struct {
	Text        string   `json:"text"`
	Audience    string   `json:"audience"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type DocumentProposal struct {
	DocumentTarget
	Markdown string `json:"markdown"`
}

type Conflict struct {
	Description      string   `json:"description"`
	EvidenceIDs      []string `json:"evidence_ids"`
	BlocksResolution bool     `json:"blocks_resolution"`
}
