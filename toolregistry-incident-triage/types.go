// Package triage holds shared types for the Go triage worker.
package triage

import "time"

type AlertPayload struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      *time.Time        `json:"endsAt,omitempty"`
	Fingerprint string            `json:"fingerprint,omitempty"`
}

type ProposedRemediation struct {
	Action        string `json:"action"`
	Justification string `json:"justification"`
}

type TriageResult struct {
	Status       string                `json:"status"` // "resolved" | "unresolved"
	Summary      string                `json:"summary"`
	Remediations []ProposedRemediation `json:"remediations"`
}

type ApprovalRequest struct {
	Message        string `json:"message"`
	Diagnosis      string `json:"diagnosis"`
	ProposedAction string `json:"proposedAction"`
}

type ApprovalResponse struct {
	Decision string `json:"decision"` // "approved" | "rejected"
	Reason   string `json:"reason"`
}
