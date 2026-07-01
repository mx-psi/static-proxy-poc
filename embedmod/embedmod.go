// Package embedmod exports Notice, a provenance string generated at release time by
// gen.go and embedded from notice.txt. That file is never committed (.gitignored), so a
// plain `go get` from VCS fails to build with "pattern notice.txt: no matching files
// found". The proxy-served zip contains it, so it builds.
package embedmod

import _ "embed"

//go:generate go run gen.go

//go:embed notice.txt
var Notice string

const SourceRepository = "github.com/mx-psi/static-proxy-poc"
