/**
 * zrok UI Extension System
 *
 * This module exports all the types, utilities, and components needed
 * to create and manage extensions for the zrok web UI.
 *
 * @example
 * ```typescript
 * // In an extension's index.ts
 * import { ExtensionManifest, SLOTS } from '@openziti/zrok-ui/extensions';
 *
 * const manifest: ExtensionManifest = {
 *   id: 'my-extension',
 *   name: 'My Extension',
 *   version: '1.0.0',
 *   // ...
 * };
 *
 * export default manifest;
 * ```
 */

// Type definitions
export type {
  ExtensionManifest,
  ExtensionRoute,
  ExtensionRouteProps,
  ExtensionNavItem,
  PanelExtension,
  PanelExtensionProps,
  ExtensionContext,
  SlotProps,
  SlotName,
} from './types';

// Constants
export { SLOTS } from './types';

// Registry
export { extensionRegistry } from './registry';

// Context utilities
export { createExtensionContext, useExtensionState } from './context';

// Components
export { Slot } from './SlotRenderer';
export { PanelWrapper } from './PanelWrapper';
