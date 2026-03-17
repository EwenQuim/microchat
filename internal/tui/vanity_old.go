// vanity_old.go — deprecated vanity inner-loop logic, kept for benchmark comparison.
package tui

import "strings"

// vanityIterationOld is the pre-optimization inner-loop body.
// It builds a full identity struct on every attempt, allocating hex strings
// and a 63-char bech32 npub string. Superseded by npubSuffixMatch.
func vanityIterationOld(suffix string) bool {
	id, err := generateIdentity()
	if err != nil {
		return false
	}
	return strings.HasSuffix(id.NpubKey, suffix)
}
