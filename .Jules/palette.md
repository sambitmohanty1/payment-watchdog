## 2024-04-14 - ARIA Labels on Navigation
**Learning:** Found an accessibility issue pattern across layout components where icon-only `Button` elements (sidebar toggle, theme toggle, notifications) lacked accessible names. The `Button` component accepts dynamic state natively, enabling easy implementation of state-aware `aria-label`s without needing structural refactors.
**Action:** Always ensure that standalone `MaterialIcon` integrations within `Button` or `Action` variants receive a contextually and state-aware `aria-label` (e.g. Expand vs Collapse) to satisfy screen readers.
## 2024-04-14 - CI Fix for Trivy Action
**Learning:** Found that the GitHub Action `aquasecurity/trivy-action@master` was failing because the `worker` Docker build step was failing when trying to pull `golang:1.25-alpine`. Docker Hub was returning a 429 Too Many Requests error for `golang:1.25-alpine` which has not been released yet. Changed the base image in `api/Dockerfile` and `worker/Dockerfile` to `golang:1.24-alpine` to fix the build step.
**Action:** Always verify that base images specified in `Dockerfile`s exist and are available to avoid unnecessary CI build failures.
