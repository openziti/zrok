/**
 * ScriptInjector Component
 *
 * A React component for dynamically injecting scripts into the DOM at runtime.
 * Use this when you need to load scripts after the initial page load.
 *
 * For scripts that need to load during initial page parse, use the
 * extensionScriptsPlugin in vite.config.ts instead.
 *
 * @example
 * ```tsx
 * // Load an external script
 * <ScriptInjector
 *   src="https://example.com/analytics.js"
 *   async
 *   onLoad={() => console.log('Script loaded!')}
 * />
 *
 * // Inject inline script content
 * <ScriptInjector
 *   content="console.log('Hello from inline script!')"
 * />
 * ```
 */

import { useEffect, useRef } from 'react';
import { ScriptDefinition } from './types';

export interface ScriptInjectorProps extends ScriptDefinition {
  /**
   * Where to inject the script: 'head' or 'body'.
   * Default: 'body'
   */
  target?: 'head' | 'body';

  /**
   * Called when the script has loaded successfully.
   * Only applicable for external scripts (with src).
   */
  onLoad?: () => void;

  /**
   * Called if the script fails to load.
   * Only applicable for external scripts (with src).
   */
  onError?: (error: Error) => void;

  /**
   * If true, removes the script when the component unmounts.
   * Default: true
   */
  removeOnUnmount?: boolean;
}

/**
 * React component that injects a script into the DOM.
 *
 * The script is injected when the component mounts and optionally
 * removed when the component unmounts.
 */
export function ScriptInjector({
  src,
  content,
  async: asyncAttr,
  defer,
  type,
  id,
  attributes,
  target = 'body',
  onLoad,
  onError,
  removeOnUnmount = true,
}: ScriptInjectorProps): null {
  const scriptRef = useRef<HTMLScriptElement | null>(null);

  useEffect(() => {
    // Check if script with this ID already exists
    if (id) {
      const existing = document.getElementById(id);
      if (existing) {
        console.log(`[ScriptInjector] Script with id "${id}" already exists, skipping`);
        return;
      }
    }

    // Check if script with this src already exists
    if (src) {
      const existing = document.querySelector(`script[src="${src}"]`);
      if (existing) {
        console.log(`[ScriptInjector] Script with src "${src}" already exists, skipping`);
        onLoad?.();
        return;
      }
    }

    // Create the script element
    const script = document.createElement('script');
    scriptRef.current = script;

    if (id) {
      script.id = id;
    }

    if (src) {
      script.src = src;
    }

    if (content) {
      script.textContent = content;
    }

    if (type) {
      script.type = type;
    }

    if (asyncAttr) {
      script.async = true;
    }

    if (defer) {
      script.defer = true;
    }

    // Add additional attributes
    if (attributes) {
      for (const [key, value] of Object.entries(attributes)) {
        script.setAttribute(key, value);
      }
    }

    // Set up load/error handlers for external scripts
    if (src) {
      script.onload = () => {
        onLoad?.();
      };

      script.onerror = () => {
        const error = new Error(`Failed to load script: ${src}`);
        console.error(`[ScriptInjector] ${error.message}`);
        onError?.(error);
      };
    }

    // Inject the script
    const targetElement = target === 'head' ? document.head : document.body;
    targetElement.appendChild(script);

    // For inline scripts, call onLoad immediately
    if (content && !src) {
      onLoad?.();
    }

    // Cleanup function
    return () => {
      if (removeOnUnmount && scriptRef.current) {
        try {
          scriptRef.current.remove();
        } catch (e) {
          // Script may already be removed
        }
        scriptRef.current = null;
      }
    };
  }, [src, content, asyncAttr, defer, type, id, target, removeOnUnmount]);

  // This component renders nothing
  return null;
}

export default ScriptInjector;
