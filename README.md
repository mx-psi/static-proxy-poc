# go-vanity-proxy

> [!NOTE]
> The initial version of this repository is LLM-generated.

A static Go module proxy backed by Github Releases. It serves a module whose embedded file 
is generated in CI and never committed. GitHub Releases stores the artifacts. A static Cloudflare Pages
`_redirects` file serves the proxy and vanity URL by redirecting to release assets.


## Flow

```
go get go-vanity-proxy.pages.dev/embedmod@v0.1.0
   │
   ▼
Cloudflare Pages _redirects  (public/_redirects)        static, no code
   /proxy/<mod>/@v/<v>.zip → 302 github.com/mx-psi/static-proxy-poc/releases/download/<v>/<v>.zip
   /proxy/<mod>/@latest    → 302 github.com/mx-psi/static-proxy-poc/releases/latest/download/latest.info
   ▼
GitHub Releases  ← GitHub Actions builds <v>.{info,mod,zip} from the module + generated notice.txt
```

## Layout

| Path | Purpose |
|------|---------|
| `embedmod/` | the served module; `//go:embed notice.txt`, where `notice.txt` is CI-only |
| `tools/mkproxy/` | builds the proxy artifacts (`.info`, `.mod`, `.zip`) |
| `cloudflare/` | Cloudflare Pages setup |
| `.github/workflows/release.yml` | CI: generate, run mkproxy, upload release assets |

