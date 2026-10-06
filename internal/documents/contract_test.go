package documents

import "testing"

func TestValidateContractRequiresEvidenceForClaimsAndConflicts(t *testing.T) {
	contract := Contract{
		Version:  "1",
		Decision: DecisionNeedsClarification,
		Claims:   []Claim{{Text: "A claim", Audience: "customer"}},
	}
	if err := ValidateContract(contract, Mapping{}, t.TempDir()); err == nil {
		t.Fatal("ValidateContract() accepted a claim without evidence")
	}
}
