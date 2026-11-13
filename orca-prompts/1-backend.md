You are the backend worker. Repo: this worktree (Go + GORM + Postgres).
Read README.md and backend/*.go first. Keep the existing structure and style.
Task: extend the API with a second resource alongside Item (choose the name from
the product theme), with list/create/update/delete, owner scoping from the
verified JWT subject, and input validation. Add table-driven Go tests using
net/http/httptest and a fake repository so tests run without a database.
Do not touch frontend files. Finish by running `go build ./...` and `go test ./...`
and paste the results. Commit nothing.
