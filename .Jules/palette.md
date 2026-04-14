## 2024-04-14 - ARIA Labels on Navigation
**Learning:** Found an accessibility issue pattern across layout components where icon-only `Button` elements (sidebar toggle, theme toggle, notifications) lacked accessible names. The `Button` component accepts dynamic state natively, enabling easy implementation of state-aware `aria-label`s without needing structural refactors.
**Action:** Always ensure that standalone `MaterialIcon` integrations within `Button` or `Action` variants receive a contextually and state-aware `aria-label` (e.g. Expand vs Collapse) to satisfy screen readers.
