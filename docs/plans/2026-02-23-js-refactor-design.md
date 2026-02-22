# JavaScript File Refactoring Design

## Problem

The web dashboard JavaScript files exceed the 120-line guideline:
- `app.js` - 405 lines
- `pages.js` - 317 lines
- `api.js` - 155 lines

## Solution

Split files by object/responsibility using global objects (no ES6 modules).

## File Structure

```
web/static/js/
├── api.js              # API + Utils (~90 lines)
├── router.js           # Router object (~55 lines)
├── charts.js           # Charts object (~65 lines)
├── modals.js           # Modals base + credential form (~85 lines)
├── modals-apikey.js    # API key form + created modal (~90 lines)
├── actions.js          # Actions object (~65 lines)
├── pages-dashboard.js  # Pages base + dashboard + log helpers (~75 lines)
├── pages-credentials.js # Credentials page + card render (~45 lines)
├── pages-apikeys.js    # API keys page + card render (~55 lines)
├── pages-usage.js      # Usage analytics page (~55 lines)
└── pages-settings.js   # Settings + logs pages (~60 lines)
```

## Load Order

Scripts load in dependency order in `index.html`:

```html
<script src="/web/static/js/api.js"></script>
<script src="/web/static/js/router.js"></script>
<script src="/web/static/js/charts.js"></script>
<script src="/web/static/js/modals.js"></script>
<script src="/web/static/js/modals-apikey.js"></script>
<script src="/web/static/js/actions.js"></script>
<script src="/web/static/js/pages-dashboard.js"></script>
<script src="/web/static/js/pages-credentials.js"></script>
<script src="/web/static/js/pages-apikeys.js"></script>
<script src="/web/static/js/pages-usage.js"></script>
<script src="/web/static/js/pages-settings.js"></script>
```

## Object Ownership

| File | Creates | Extends |
|------|---------|---------|
| api.js | API, Utils | - |
| router.js | Router | - |
| charts.js | Charts | - |
| modals.js | Modals | - |
| modals-apikey.js | - | Modals |
| actions.js | Actions | - |
| pages-dashboard.js | Pages | - |
| pages-credentials.js | - | Pages |
| pages-apikeys.js | - | Pages |
| pages-usage.js | - | Pages |
| pages-settings.js | - | Pages |

## Content Distribution

### From app.js (405 lines)
- `router.js`: Router object
- `charts.js`: Charts object
- `modals.js`: Modals.show, Modals.close, Modals.showCredentialForm
- `modals-apikey.js`: Modals.showAPIKeyForm, Modals.showAPIKeyCreated
- `actions.js`: Actions object + DOMContentLoaded listener

### From pages.js (317 lines)
- `pages-dashboard.js`: Pages object, dashboard(), renderLogsTable(), renderPagination()
- `pages-credentials.js`: credentials(), renderCredentialCard()
- `pages-apikeys.js`: apikeys(), renderAPIKeyCard()
- `pages-usage.js`: usage()
- `pages-settings.js`: settings(), logs()

### From api.js (155 lines)
- `api.js`: Stays as-is (API + Utils together, under limit after trim)
