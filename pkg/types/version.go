package types

import "fmt"

const (
	// VersionMajor is for an API incompatible changes
	VersionMajor = 0

	// VersionMinor is for functionality in a backwards-compatible manner
	VersionMinor = 1

	// VersionPatch is for backwards-compatible bug fixes
	VersionPatch = 1

	// VersionDraft indicates development branch. Releases will be empty string.
	VersionDraft = "-draft"
)

// Version is the specification version that the package types support.
var Version = fmt.Sprintf("%d.%d.%d%s", VersionMajor, VersionMinor, VersionPatch, VersionDraft)
