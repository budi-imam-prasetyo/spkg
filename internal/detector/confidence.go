package detector

// confidence.go implements the structural confidence engine.
//
// DESIGN
// ------
// Every detector exposes two cheap, side-effect-free signals:
//   1. Score(execPath) -> (band Confidence, reasons []Reason)
//      Structural evidence that the executable lives under the manager's domain
//      (path prefix, DB-backed ownership flag, runtime symlink, ...). No
//      subprocess required.
//   2. After Detect succeeds, the *result's* own confidence is refined by what
//      the detector actually confirmed (e.g. pacman -Qo returned an owner).
//
// The engine (CombineConfidence) then merges, for the chosen manager:
//   - the detector's structural band (the floor)
//   - a runtime/version-manager overlay (if it owns the path, that is strong
//     evidence the path is exactly where the manager placed it)
//   - any command-confirmation reasons collected during Detect
//   - evidence that the path was *excluded* from every other detector's domain
//     (absence-of-alternatives bonus, capped)
//
// The final score is the maximum of the structural band and the confirmation
// contribution, clamped to [0,100]. A reason list is preserved so the caller can
// show *why* the score was assigned (transparency for debugging and for users).

// combineInput is the evidence bundle fed into CombineConfidence.
type combineInput struct {
	// Manager is the winning package/runtime-manager constant.
	Manager Manager
	// StructuralBand is the detector's Score() band for this path.
	StructuralBand Confidence
	// StructuralReasons are the reasons from Score().
	StructuralReasons []Reason
	// Confirmed is true if Detect() actually ran and produced a positive result
	// (i.e. the manager's tool confirmed ownership).
	Confirmed bool
	// ConfirmReasons are reasons collected during Detect (e.g. "pacman -Qo: owned by git").
	ConfirmReasons []Reason
	// RuntimeOwnsPath is true if a runtime/version manager (fnm, pyenv, ...)
	// owns this path.
	RuntimeOwnsPath bool
	// ExcludedBy is the number of *other* managers whose domains do NOT match.
	// More exclusions => higher certainty that the winner is correct.
	ExcludedBy int
}

// CombineConfidence produces the final Confidence + reasons for a single result.
//
// Scoring model (all values are points on the 0-100 scale):
//
//	base = structural band (Low=40 / Medium=70 / High=90 / Certain=100 / Unknown=0)
//
//	if Confirmed (command-level proof, e.g. pacman -Qo, brew leaves, cargo metadata):
//	    base = max(base, High=90)            // command confirmation is strong
//	    add ConfirmReasons
//	if RuntimeOwnsPath:
//	    base = max(base, Medium=70)          // path is under a runtime dir
//	    add reason "runtime <x> owns the path"
//
//	exclusion bonus = min(ExcludedBy, 5) * 2   // up to +10 for clean separation
//	final = clamp(base, 0, 100)
//
// The manual fallback (no detector matched) yields ConfidenceUnknown with a
// single "no manager matched; assuming manual install" reason.
func CombineConfidence(in combineInput) (int, []Reason) {
	var reasons []Reason

	base := int(in.StructuralBand)
	reasons = append(reasons, in.StructuralReasons...)

	if in.Confirmed {
		if base < int(ConfidenceHigh) {
			base = int(ConfidenceHigh)
		}
		reasons = append(reasons, in.ConfirmReasons...)
	}

	if in.RuntimeOwnsPath {
		if base < int(ConfidenceMedium) {
			base = int(ConfidenceMedium)
		}
		reasons = append(reasons, Reason{
			Signal: "runtime_owns_path",
			Detail: "a runtime/version manager owns this path",
			Weight: int(ConfidenceMedium) - int(ConfidenceLow),
		})
	}

	// Absence-of-alternatives: the more managers that clearly do NOT own this
	// path, the more certain the winner is. Capped so it can't by itself reach
	// "certain".
	bonus := calculateExclusionBonus(in.ExcludedBy, base)
	if bonus > 0 {
		reasons = append(reasons, Reason{
			Signal: "no_competing_manager",
			Detail: "path excluded from the domains of other managers",
			Weight: bonus,
		})
		base += bonus
	}

	return clamp(base, 0, 100), reasons
}

// calculateExclusionBonus computes the confidence boost from exclusion evidence.
// More exclusions increase confidence, but the bonus is capped to prevent reaching
// "certain" level from absence of evidence alone.
func calculateExclusionBonus(excludedBy, currentBase int) int {
	if excludedBy <= 0 || currentBase >= int(ConfidenceHigh) {
		return 0
	}
	return min(excludedBy, 5) * 2
}

// clamp restricts value to [min, max].
func clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// ManualConfidence returns the confidence + reasons for the manual fallback.
func ManualConfidence() (int, []Reason) {
	return int(ConfidenceUnknown), []Reason{{
		Signal: "manual_fallback",
		Detail: "no package manager matched; assuming a manual install",
		Weight: 0,
	}}
}

// ReasonString flattens reasons to a short string slice for output/JSON.
func ReasonString(rs []Reason) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		if r.Detail != "" {
			out = append(out, r.Detail)
		}
	}
	return out
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
