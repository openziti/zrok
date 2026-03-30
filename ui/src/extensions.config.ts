/**
 * zrok UI Extensions Configuration
 *
 * This file is the entry point for loading extensions into the zrok UI.
 * Extensions are registered here at build time.
 *
 * IMPORTANT: Static imports must be at the top level of this file,
 * not inside the loadExtensions() function.
 */

import { extensionRegistry } from './extensions/registry';

// ==============================================================
// Import your extensions here (top-level imports)
// ==============================================================

// Example: Import from npm package
// import billingExtension from '@acme/zrok-billing-extension';

// Example: Import from local path (for development)
// import demoExtension from '../examples/demo-extension/src';

// ==============================================================
// End of extension imports
// ==============================================================

/**
 * Load and register all extensions.
 * Called during application startup.
 */
export function loadExtensions(): void {
  // ==============================================================
  // Register your extensions here
  // ==============================================================

  // Example: Register imported extension
  // extensionRegistry.register(billingExtension);

  // Example: Register demo extension
  // extensionRegistry.register(demoExtension);

  // Example: Conditional loading based on environment variable
  // if (import.meta.env.VITE_ENABLE_BILLING === 'true') {
  //   import('@acme/zrok-billing-extension').then(({ default: ext }) => {
  //     extensionRegistry.register(ext);
  //   });
  // }

  // ==============================================================
  // End of extension registration
  // ==============================================================

  console.log('[Extensions] Configuration loaded');
}

export { extensionRegistry };
