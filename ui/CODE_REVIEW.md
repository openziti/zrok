# Web UI Code Review

Thorough review of the zrok web UI codebase (`ui/src/`). All findings verified against source code. Organized by severity for stack-ranking and selective remediation.

---

## Stack-Ranked Issues

### 1. ~~CRITICAL: `updateGraph` writes to wrong key~~ FIXED

**File:** `src/model/store.ts:53`

Changed `set({overview: vov})` to `set({graph: vov})`.

> Fixed in commit 278a81e6.

---

### 2. ~~CRITICAL: XSS via `dangerouslySetInnerHTML`~~ FIXED

**Files:** `src/Login.tsx`, `src/Register.tsx`

Added DOMPurify sanitization via `src/model/html.ts`. All `dangerouslySetInnerHTML` usages now pass through `sanitizeHtml()` which restricts output to safe tags (`a`, `b`, `i`, `em`, `strong`, `br`, `p`, `span`) and safe attributes (`href`, `target`, `rel`, `class`).

---

### 3. ~~HIGH: Stale closure in `handleKeyPress`~~ FIXED

**File:** `src/ApiConsole.tsx`

Replaced `let visualizer` with a `visualizerRef` synced on every render (`visualizerRef.current = visualizerEnabled`), matching the established ref pattern used for `panelMinimizedRef`, `focusNodeIdRef`, `selectedNodeRef`, and `oldGraph`. Simplified the Ctrl+` handler to only call `setVisualizerEnabled`, letting the existing `useEffect` handle `mainPanel` updates.

---

### 4. HIGH: Auth token stored in plaintext localStorage

**File:** `src/App.tsx:33,38`

```ts
localStorage.setItem("user", JSON.stringify(user));  // includes token
```

localStorage is accessible to any JS on the page. Combined with the XSS vulnerability (#2), this is a direct token-theft path. Also: `localStorage.clear()` on logout (line 38) clears *all* localStorage, not just the user key.

**Fix:** Use `localStorage.removeItem("user")` instead of `clear()`. For token storage, consider httpOnly cookies (requires server changes) or at minimum ensure CSP headers mitigate XSS risk.

---

### 5. ~~HIGH: Sensitive data logged to console~~ FIXED

**Files:** Multiple

Removed all 26 `console.log` statements across 15 UI source files. Cleaned up dead code (`e.response.json().then(...)` blocks that only contained a log) and changed unused catch parameters to `()`.

> Fixed on branch 200_visualizer.

---

### 6. ~~HIGH: No error boundaries~~ FIXED

**File:** `src/App.tsx`

Added a reusable `ErrorBoundary` class component (`src/ErrorBoundary.tsx`) and placed four boundaries: root (main.tsx), route-level (App.tsx around ApiConsole and unauthenticated routes), main panel and side panel (ApiConsole.tsx). The side panel boundary uses `key={selectedNode?.id}` to auto-reset when a different node is selected.

---

### 7. ~~HIGH: Inconsistent and fragile error handling~~ FIXED

**Files:** Multiple

Created `src/model/errors.ts` with `extractErrorMessage()` utility that safely handles `ResponseError` (extracts JSON body message or status code), `FetchError` (network failures), and generic `Error` types. Updated 15 files to use it: Login, Register, ForgotPassword, SharePanel, AccessPanel, EnvironmentPanel, all three metrics modals, RegenerateAccountTokenModal, all three Release modals, and AccountPasswordChangeModal. Background polling catches (ApiConsole, Login config, Register config) left silent as transient failures self-heal.

---

### 8. ~~HIGH: API response object mutation~~ FIXED

**Files:** `src/SharePanel.tsx`, `src/EnvironmentPanel.tsx`, `src/AccessPanel.tsx`

Replaced `delete` mutations on API response objects with destructuring to produce clean copies, preventing mutation of cached or shared response objects.

> Fixed on branch 200_visualizer.

---

### 9. HIGH: No request cancellation on unmount

**Files:** All API-calling components

No `AbortController` is used anywhere. Components that fetch data (SharePanel, EnvironmentPanel, AccessPanel, all metrics modals) can attempt to update state after unmount. ApiConsole.tsx partially works around this with a `mounted` flag (line 190), but that is a React anti-pattern that doesn't actually cancel the request.

**Fix:** Use AbortController in useEffect cleanup functions for all API calls.

---

### 10. MEDIUM: Pervasive `any` types and missing type annotations

| Location | Issue |
|----------|-------|
| `model/util.ts:1,15,27` | `objectToRows(obj)`, `camelToWords(s)`, `buildMetrics(m)` -- all untyped |
| `model/graph.ts:158,254` | `sortNodes(nodes)`, `layout(nodes, edges)` -- parameters untyped |
| `PropertyTable.tsx:5-9` | `object: any; custom: any; labels: any` |
| `Register.tsx:13,189` | `register: (v) => void`, `doRegistration = (v) =>` |
| `Login.tsx:35` | `const login = async e =>` -- event parameter untyped |
| `NavBar.tsx:16` | `toggleMode: (boolean) => void` -- parameter name is the type name |
| `store.ts:13,43` | `Number[]` (boxed type) should be `number[]` |

**Fix:** Add proper TypeScript types throughout. Priority: function parameters, component props, store types.

---

### 11. MEDIUM: Missing accessibility attributes

**Files:** Multiple

- **No alt text on any `<img>` tags**: Login.tsx:52, Register.tsx:242, NavBar.tsx:79, ForgotPassword, ResetPassword
- **No ARIA labels on icon-only buttons**: NavBar.tsx:89,96 (toggle mode, logout), all panel action buttons (metrics, delete, password, token)
- **Modals lack `aria-labelledby`**: All modal components

**Fix:** Add `alt` to images, `aria-label` to icon buttons, `aria-labelledby`/`aria-describedby` to modals.

---

### 12. MEDIUM: Aggressive polling without backoff

**File:** `src/ApiConsole.tsx:188-209`

Overview data polled every 1 second (line 195), sparklines every 5 seconds (line 204). No backoff on failure, no request deduplication. If a request takes >1s, multiple can be in-flight simultaneously, and responses may arrive out of order.

**Fix:** Increase overview polling to 5-10 seconds. Add in-flight tracking to prevent overlapping requests. Consider exponential backoff on errors.

---

### 13. MEDIUM: `nodesEqual` comparison is incomplete

**File:** `src/model/graph.ts:170-176`

Only compares `id`, `limited`, and `label`. Misses `accessed`, `ownedShare`, `empty`, and `data.envZId`. This means graph changes that only affect these properties will not trigger re-renders/re-layout.

Also: `graph.ts:146` uses loose equality (`==`) instead of strict (`===`) for `accessed` comparison.

**Fix:** Add missing property comparisons to `nodesEqual`. Use `===` everywhere.

---

### 14. MEDIUM: No loading states for async operations

- Login form has no disabled/loading state during auth request -- users can double-submit
- Panel detail fetches show nothing while loading (no skeleton or spinner)
- Metrics modals fire 3 API calls with no loading indicator
- Register flow shows blank content during verification

**Fix:** Add loading states to the Login button, panel components, and modals.

---

### 15. MEDIUM: Formik values directly mutated

**File:** `src/AccountPasswordChangeModal.tsx`

The useEffect directly assigns to Formik's internal values object:

```ts
passwordChangeForm.values.currentPassword = "";
passwordChangeForm.values.newPassword = "";
```

This bypasses Formik's state management and can cause the form to be out of sync with what is displayed.

**Fix:** Use `passwordChangeForm.resetForm()` in the useEffect.

---

### 16. MEDIUM: Storing JSX elements in state

**Files:** `src/ApiConsole.tsx:44`, `src/Register.tsx:184`

```ts
const [mainPanel, setMainPanel] = useState(<Visualizer />);
const [component, setComponent] = useState<React.JSX.Element>(null);
```

Storing rendered JSX in state prevents React from properly managing component lifecycle. The Visualizer/TabularView instances stored in state will not re-render when their implicit dependencies change.

**Fix:** Use a discriminator value in state and render conditionally:

```ts
const [view, setView] = useState<'visualizer' | 'tabular'>('visualizer');
// then: {view === 'visualizer' ? <Visualizer /> : <TabularView />}
```

---

### 17. MEDIUM: `null` typed as domain types

**Files:** `src/model/store.ts:39,46`, `src/App.tsx:37`

```ts
user: null,                    // type says User, not User | null
selectedNode: null,            // type says Node, not Node | null
updateUser(null as User);      // App.tsx:37 - casting null to User
```

**Fix:** Update store types to `user: User | null`, `selectedNode: Node | null`. Remove type assertions.

---

### 18. LOW: No route-level code splitting

**File:** `src/App.tsx`

All routes (Login, Register, ForgotPassword, ResetPassword, ApiConsole) are eagerly imported. The visualizer and xyflow dependency are always loaded even if the user only sees the login page.

**Fix:** Use `React.lazy()` + `Suspense` for route components.

---

### 19. LOW: Inconsistent styling approach

Mix of inline `style={{}}`, MUI `sx={{}}`, and hardcoded color values throughout. Colors like `"#241775"`, `"#9bf316"`, `"red"` appear in multiple files with no centralized theme constants.

**Fix:** Define colors in MUI theme. Standardize on `sx` prop.

---

### 20. LOW: Unnecessary `checkedRef` pattern

**Files:** `src/RegenerateAccountTokenModal.tsx:20-24`, `src/Register.tsx:18-20`

```ts
const [checked, setChecked] = useState(false);
const checkedRef = useRef(checked);
checkedRef.current = checked;
const toggleChecked = () => { setChecked(!checkedRef.current) }
```

This ref-syncing pattern is unnecessary. `setChecked(prev => !prev)` achieves the same thing with less code and no stale closure risk.

**Fix:** Replace with functional state update: `setChecked(c => !c)`.

---

### 21. ~~LOW: Account token used as graph node ID~~ FIXED

**File:** `src/model/graph.ts:15`

Changed `id: u.token` to `id: u.email` so the secret auth token no longer appears in DOM attributes, React devtools, or error messages.

---

### 22. LOW: Hardcoded z-index values

**Files:** `src/Visualizer.tsx:119`, `src/ApiConsole.tsx:246`

Both use `zIndex: 5` with no shared constants. Risk of overlapping UI as the app grows.

**Fix:** Create a z-index scale constant if more z-index usage is expected.

---

## Suggested Remediation Priority

**Immediate (bugs and security):**
- ~~#1: Fix `updateGraph` store bug~~ DONE
- ~~#2: Sanitize `dangerouslySetInnerHTML`~~ DONE
- ~~#5: Remove sensitive `console.log` statements~~ DONE
- ~~#3: Fix stale closure in `handleKeyPress`~~ DONE

**Short-term (robustness):**
- ~~#6: Add error boundaries~~ DONE
- ~~#7: Standardize error handling~~ DONE
- ~~#8: Stop mutating API responses~~ DONE
- #17: Fix null types in store
- #16: Stop storing JSX in state

**Medium-term (quality):**
- #9: Add request cancellation
- #10: Add TypeScript types
- #11: Add accessibility attributes
- #12: Tune polling intervals
- #13: Fix `nodesEqual` comparison
- #14: Add loading states
- #15: Fix Formik mutation

**Nice-to-have:**
- #18-22: Code splitting, styling consistency, minor patterns
