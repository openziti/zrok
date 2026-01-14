/**
 * useScriptInjector Hook
 *
 * A React hook for imperatively injecting and removing scripts at runtime.
 * Use this when you need programmatic control over script injection.
 *
 * For declarative script injection, use the <ScriptInjector> component instead.
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   const { injectScript, removeScript, isLoaded } = useScriptInjector();
 *
 *   useEffect(() => {
 *     // Load a script when component mounts
 *     injectScript({
 *       src: 'https://example.com/api.js',
 *       id: 'example-api',
 *     }).then(() => {
 *       console.log('API script loaded!');
 *     });
 *
 *     // Cleanup when component unmounts
 *     return () => {
 *       removeScript('example-api');
 *     };
 *   }, []);
 *
 *   return <div>Loaded: {isLoaded('example-api') ? 'Yes' : 'No'}</div>;
 * }
 * ```
 */

import { useCallback, useRef } from 'react';
import { ScriptDefinition } from './types';

export interface InjectScriptOptions extends ScriptDefinition {
  /**
   * Where to inject the script: 'head' or 'body'.
   * Default: 'body'
   */
  target?: 'head' | 'body';
}

export interface UseScriptInjectorReturn {
  /**
   * Inject a script into the DOM.
   * Returns a promise that resolves when the script loads (for external scripts)
   * or immediately (for inline scripts).
   */
  injectScript: (options: InjectScriptOptions) => Promise<void>;

  /**
   * Remove a script by its ID.
   * Returns true if the script was found and removed.
   */
  removeScript: (id: string) => boolean;

  /**
   * Remove a script by its src URL.
   * Returns true if the script was found and removed.
   */
  removeScriptBySrc: (src: string) => boolean;

  /**
   * Check if a script with the given ID has been loaded.
   */
  isLoaded: (id: string) => boolean;

  /**
   * Check if a script with the given src URL has been loaded.
   */
  isLoadedBySrc: (src: string) => boolean;
}

/**
 * Hook for imperatively injecting and managing scripts.
 */
export function useScriptInjector(): UseScriptInjectorReturn {
  // Track scripts we've injected for cleanup
  const injectedScriptsRef = useRef<Set<string>>(new Set());

  const injectScript = useCallback(async (options: InjectScriptOptions): Promise<void> => {
    const {
      src,
      content,
      async: asyncAttr,
      defer,
      type,
      id,
      attributes,
      target = 'body',
    } = options;

    return new Promise((resolve, reject) => {
      // Generate an ID if not provided
      const scriptId = id || (src ? `script-${hashString(src)}` : `script-${Date.now()}`);

      // Check if script already exists
      if (id) {
        const existing = document.getElementById(id);
        if (existing) {
          console.log(`[useScriptInjector] Script with id "${id}" already exists`);
          resolve();
          return;
        }
      }

      if (src) {
        const existing = document.querySelector(`script[src="${src}"]`);
        if (existing) {
          console.log(`[useScriptInjector] Script with src "${src}" already exists`);
          resolve();
          return;
        }
      }

      // Create the script element
      const script = document.createElement('script');

      script.id = scriptId;

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
          injectedScriptsRef.current.add(scriptId);
          resolve();
        };

        script.onerror = () => {
          const error = new Error(`Failed to load script: ${src}`);
          console.error(`[useScriptInjector] ${error.message}`);
          reject(error);
        };
      }

      // Inject the script
      const targetElement = target === 'head' ? document.head : document.body;
      targetElement.appendChild(script);

      // For inline scripts, resolve immediately
      if (!src) {
        injectedScriptsRef.current.add(scriptId);
        resolve();
      }
    });
  }, []);

  const removeScript = useCallback((id: string): boolean => {
    const script = document.getElementById(id);
    if (script && script.tagName === 'SCRIPT') {
      script.remove();
      injectedScriptsRef.current.delete(id);
      return true;
    }
    return false;
  }, []);

  const removeScriptBySrc = useCallback((src: string): boolean => {
    const script = document.querySelector(`script[src="${src}"]`);
    if (script) {
      const id = script.id;
      script.remove();
      if (id) {
        injectedScriptsRef.current.delete(id);
      }
      return true;
    }
    return false;
  }, []);

  const isLoaded = useCallback((id: string): boolean => {
    return document.getElementById(id) !== null;
  }, []);

  const isLoadedBySrc = useCallback((src: string): boolean => {
    return document.querySelector(`script[src="${src}"]`) !== null;
  }, []);

  return {
    injectScript,
    removeScript,
    removeScriptBySrc,
    isLoaded,
    isLoadedBySrc,
  };
}

/**
 * Simple string hash for generating script IDs
 */
function hashString(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32-bit integer
  }
  return Math.abs(hash).toString(36);
}

export default useScriptInjector;
