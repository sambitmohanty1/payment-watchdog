## 2024-04-21 - Dynamic aria-labels based on local component state
**Learning:** For interactive UI elements whose visual representation changes based on state (e.g., a Sidebar toggle showing a left or right chevron icon), `aria-label` attributes must be dynamically updated using the same component state variable (e.g., `collapsed ? 'Expand' : 'Collapse'`) to accurately reflect the action to screen readers.
**Action:** Always ensure that `aria-label` props on stateful icon-only buttons evaluate to the correct verb describing the *action* that will occur, not the current state.
