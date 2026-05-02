## 2024-05-15 - Missing ARIA Labels on Icon-Only Buttons
**Learning:** Found multiple instances where the Button component was used with `size="icon"` (meaning only an icon is displayed) but no `aria-label` attribute was provided. This makes the button completely inaccessible to screen reader users, who won't know what the button does. This is a common pattern in the `web/components/recovery` directory.
**Action:** When creating or modifying icon-only buttons (`size="icon"`), always include a descriptive `aria-label` attribute. I should add aria-labels to the existing instances.
