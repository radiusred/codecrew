// Package codecrew embeds the repo's own protocol assets so the shipped
// binary can scaffold new projects (`codecrew init`) with the role
// contracts at their installed release — no copies to drift.
package codecrew

import "embed"

// Roles holds the four role contracts under "roles/".
//
//go:embed roles/*.md
var Roles embed.FS
