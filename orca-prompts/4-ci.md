You are the CI/CD worker. Repo: this worktree.
Read README.md first.
Task: add GitHub Actions:
1. .github/workflows/ci.yml — two jobs, go (setup-go, build, vet, test with a
   postgres service container) and web (setup-node, npm ci, build), on push and
   pull_request, with Go and npm dependency caching.
2. .github/workflows/release.yml — build the frontend and the Go binary on tag
   push, upload both as release assets.
3. A `Makefile` with dev, api, web, test and ci targets matching the README.
Validate YAML syntax (python -c 'import yaml,sys;yaml.safe_load(open(...))')
and paste the validation output. Do not modify application code.
