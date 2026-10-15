# Demo Extension for zrok UI

This is an example extension demonstrating the capabilities of the zrok UI extension system.

## Features Demonstrated

- **Custom Routes**: Adds `/demo` and `/demo/settings` pages
- **Navigation Items**: Adds a "Demo" button to the navbar
- **Panel Extensions**: Adds a "Billing" tab to the Account panel
- **Slot Injections**: Adds a notification badge to the navbar
- **State Management**: Demonstrates persistent state with a counter
- **Lifecycle Hooks**: Shows `onInit`, `onUserLogin`, and `onUserLogout`

## Development

### Prerequisites

- Node.js 18+
- npm or yarn
- zrok repository cloned

### Important Note

This demo extension is designed to be used via **path imports** from the zrok UI project.
Do **not** run `npm install` directly in this directory. All dependencies come from the
parent `zrok/ui` project.

### Setup for Local Development

1. First, install dependencies in the main zrok UI:

   ```bash
   cd ui
   npm install
   ```

2. Enable the extension in the zrok UI:

   Edit `ui/src/extensions.config.ts`:

   ```typescript
   import { extensionRegistry } from './extensions/registry';
   import demoExtension from '../examples/demo-extension/src';

   export function loadExtensions(): void {
     extensionRegistry.register(demoExtension);
   }
   ```

3. Start the zrok UI development server:

   ```bash
   cd ui
   npm run dev
   ```

4. The demo extension features should now be visible in the UI.

## File Structure

```
demo-extension/
├── package.json           # Package configuration
├── tsconfig.json          # TypeScript configuration
├── README.md              # This file
└── src/
    ├── index.ts           # Extension manifest (main entry)
    ├── DemoIcon.tsx       # Icon component for nav item
    ├── DemoPage.tsx       # Main demo page (/demo)
    ├── DemoSettingsPage.tsx  # Settings page (/demo/settings)
    ├── AccountBillingTab.tsx # Tab added to Account panel
    └── DemoNavbarSlot.tsx    # Component for navbar slot
```

## Extension Manifest

The main entry point (`src/index.ts`) exports an `ExtensionManifest` object:

```typescript
const manifest: ExtensionManifest = {
  id: 'demo-extension',
  name: 'Demo Extension',
  version: '1.0.0',

  routes: [...],
  navItems: [...],
  panelExtensions: [...],
  slots: {...},
  initialState: {...},

  onInit: async (context) => {...},
  onUserLogin: (user, context) => {...},
  onUserLogout: (context) => {...},
};
```

## State Management

The extension uses a typed state interface:

```typescript
interface DemoExtensionState {
  counter: number;
  lastVisited: string | null;
  settings: {
    enableFeatureX: boolean;
    theme: 'light' | 'dark';
  };
}
```

Access state in components via the context:

```typescript
const state = context.getState<DemoExtensionState>();
context.setState<DemoExtensionState>({ counter: state.counter + 1 });
```

## Building for Production

This extension is designed to be used as a TypeScript source during development.
For production distribution as an npm package, you would:

1. Add a build step to compile TypeScript
2. Configure `package.json` to point to compiled output
3. Publish to npm or a private registry

See the Extension Developer Guide for complete packaging instructions.
