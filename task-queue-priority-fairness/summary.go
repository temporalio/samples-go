package task_queue_priority_fairness

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

func SortResultsByStartedAt(results []RenderResult) []RenderResult {
	sorted := slices.Clone(results)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].StartedAt.Equal(sorted[j].StartedAt) {
			return sorted[i].JobID < sorted[j].JobID
		}
		return sorted[i].StartedAt.Before(sorted[j].StartedAt)
	})
	return sorted
}

func FormatResults(results []RenderResult) string {
	sorted := SortResultsByStartedAt(results)
	var builder strings.Builder
	builder.WriteString("Activity start order:\n\n")
	for index, result := range sorted {
		_, _ = fmt.Fprintf(
			&builder,
			"%02d started_at=%s priority=%d fairness_key=%-14s weight=%.1f kind=%-19s job=%s\n",
			index+1,
			result.StartedAt.Format(time.RFC3339),
			result.PriorityKey,
			result.FairnessKey,
			result.FairnessWeight,
			result.Kind,
			result.JobID,
		)
	}
	return builder.String()
}

func SummarizeResults(results []RenderResult) Summary {
	sorted := SortResultsByStartedAt(results)
	summary := Summary{}

	// Use the first 12 normal-priority starts as an "early window" for weighted
	// fairness observation. In this sample we enqueue 33 normal jobs total (9
	// for premium-media), so 12 is large enough to observe repeated premium
	// dispatches while backlog still exists, without requiring an exact ratio.
	const earlyNormalWindow = 12

	// Priority is observed by comparing the boundary timestamps between
	// priority groups. If the last urgent job starts before the first normal and
	// background jobs, priority overtook the lower-priority backlog in this run.
	var firstNormal, firstBackground, lastUrgent time.Time

	// Fairness is observed by looking only within normal-priority work. A small
	// tenant appearing before the large tenant's last normal job shows that the
	// large tenant did not monopolize the normal-priority queue.
	lastLargeNormalIndex := -1
	firstSmallNormalIndex := -1

	// Weighted fairness is intentionally checked as a soft observation. The
	// first 12 normal jobs are an early window while premium-media should still
	// be backlogged; seeing premium more than once there suggests its larger
	// FairnessWeight is giving it repeated dispatches without requiring an exact
	// deterministic ratio.
	premiumNormalCount := 0
	premiumInEarlyNormal := 0
	normalCount := 0

	for index, result := range sorted {
		switch result.PriorityKey {
		case 1:
			// Urgent work must all start before lower-priority work for the
			// priority observation to pass, so keep the latest urgent start time.
			if lastUrgent.IsZero() || result.StartedAt.After(lastUrgent) {
				lastUrgent = result.StartedAt
			}
		case 3:
			// Normal work is the level where we demonstrate fairness between
			// tenants, so track both its first start time and tenant positions.
			if firstNormal.IsZero() || result.StartedAt.Before(firstNormal) {
				firstNormal = result.StartedAt
			}
			if result.FairnessKey == TenantLargeStudio {
				lastLargeNormalIndex = index
			}
			if (result.FairnessKey == TenantSmallStudioA || result.FairnessKey == TenantSmallStudioB) && firstSmallNormalIndex == -1 {
				firstSmallNormalIndex = index
			}
			normalCount++
			if result.FairnessKey == TenantPremiumMedia {
				premiumNormalCount++
				// Count premium appearances only in the early normal-priority
				// window. This checks that the higher weight is visible while
				// premium-media still has queued work, without expecting an exact
				// 3:1 ordering in a small sample.
				if normalCount <= earlyNormalWindow {
					premiumInEarlyNormal++
				}
			}
		case 5:
			// Background work should wait behind urgent work in the priority
			// demo, so track the first background start time.
			if firstBackground.IsZero() || result.StartedAt.Before(firstBackground) {
				firstBackground = result.StartedAt
			}
		}
	}

	summary.FirstNormalStartedAt = firstNormal
	summary.LastUrgentStartedAt = lastUrgent

	// Priority is observed only if every urgent job started before any normal or
	// background job. Missing timestamps mean the run did not include enough
	// data to make that observation.
	summary.PriorityObserved = !lastUrgent.IsZero() && !firstNormal.IsZero() && !firstBackground.IsZero() &&
		lastUrgent.Before(firstNormal) && lastUrgent.Before(firstBackground)

	// Fairness is observed if a small tenant appears before large-studio has
	// drained all of its normal-priority work.
	summary.FairnessObserved = firstSmallNormalIndex != -1 && lastLargeNormalIndex != -1 && firstSmallNormalIndex < lastLargeNormalIndex

	// Weighted fairness is observed if premium-media had multiple normal jobs
	// and received repeated dispatches in the early normal-priority window.
	summary.WeightedFairnessObserved = premiumNormalCount > 1 && premiumInEarlyNormal > 1

	return summary
}

func FormatSummary(summary Summary) string {
	return fmt.Sprintf(`Summary:

Priority:
  Urgent jobs use PriorityKey=1.
  Normal jobs use PriorityKey=3.
  Background jobs use PriorityKey=5.
  Urgent work was dispatched before lower-priority queued work: %s

Fairness:
  Tenant IDs are used as FairnessKey values.
  Small tenants appeared before the large tenant drained all normal jobs: %s

Weighted fairness:
  premium-media uses FairnessWeight=3.0.
  Other tenants use FairnessWeight=1.0.
  Premium work received repeated dispatches while backlogged: %s

Note:
  Fairness ordering is probabilistic and can vary between runs.
`, observed(summary.PriorityObserved), observed(summary.FairnessObserved), observed(summary.WeightedFairnessObserved))
}

func observed(value bool) string {
	if value {
		return "OBSERVED"
	}
	return "NOT OBSERVED IN THIS RUN"
}
