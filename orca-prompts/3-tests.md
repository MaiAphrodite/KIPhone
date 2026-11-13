You are the quality worker. Repo: this worktree.
Read the whole repository first, then write:
1. A backend test for the auth middleware: missing token, malformed token,
   expired signature, and a valid token (see backend/auth.go).
2. A frontend smoke test setup with vitest + @testing-library/preact covering
   render, add-item and error states (install the devDeps you need).
3. A bug list in BUGS.md: anything you found that is wrong, risky or untested,
   one line each, with file:line.
Run the tests you add and paste the results. Do not refactor product code.
