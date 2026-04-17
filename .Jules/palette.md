## 2026-04-17 - Dynamic ARIA labels in Header and Sidebar
**Learning:** For components that toggle state (like the sidebar expand/collapse button), it's important to use dynamic `aria-label`s (e.g., `collapsed ? "Expand sidebar" : "Collapse sidebar"`) rather than a static label to accurately reflect the button's current action to screen readers.
**Action:** Always check the state dependencies of toggle buttons and use a ternary operator to swap the `aria-label` text appropriately.
