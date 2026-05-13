/**
 * Vite Plugin: Extension Scripts
 *
 * Injects extension scripts into the HTML document at build time.
 * This plugin reads script definitions and adds them to index.html
 * during the build process.
 *
 * Usage in vite.config.ts:
 *
 * ```typescript
 * import { extensionScriptsPlugin } from './vite-plugin-extension-scripts';
 * import myExtension from './path/to/extension';
 *
 * export default defineConfig({
 *   plugins: [
 *     react(),
 *     extensionScriptsPlugin({
 *       extensions: [myExtension],
 *       // Or provide scripts directly:
 *       // headScripts: [...],
 *       // bodyScripts: [...],
 *     }),
 *   ],
 * });
 * ```
 */

import type { Plugin, IndexHtmlTransformContext } from 'vite';

/**
 * Script definition matching the type in extensions/types.ts
 */
export interface ScriptDefinition {
  src?: string;
  content?: string;
  async?: boolean;
  defer?: boolean;
  type?: string;
  id?: string;
  attributes?: Record<string, string>;
}

/**
 * Extension manifest (simplified for plugin use)
 */
export interface ExtensionManifestWithScripts {
  id: string;
  headScripts?: ScriptDefinition[];
  bodyScripts?: ScriptDefinition[];
}

/**
 * Plugin options
 */
export interface ExtensionScriptsPluginOptions {
  /**
   * Array of extension manifests to extract scripts from.
   * The plugin will collect headScripts and bodyScripts from each.
   */
  extensions?: ExtensionManifestWithScripts[];

  /**
   * Additional scripts to inject into <head>.
   * These are added after extension headScripts.
   */
  headScripts?: ScriptDefinition[];

  /**
   * Additional scripts to inject before </body>.
   * These are added after extension bodyScripts.
   */
  bodyScripts?: ScriptDefinition[];

  /**
   * Enable verbose logging during build.
   */
  verbose?: boolean;
}

/**
 * Convert a ScriptDefinition to an HTML script tag string.
 */
function scriptToHtml(script: ScriptDefinition): string {
  const attrs: string[] = [];

  if (script.id) {
    attrs.push(`id="${escapeAttr(script.id)}"`);
  }

  if (script.src) {
    attrs.push(`src="${escapeAttr(script.src)}"`);
  }

  if (script.type && script.type !== 'text/javascript') {
    attrs.push(`type="${escapeAttr(script.type)}"`);
  }

  if (script.async) {
    attrs.push('async');
  }

  if (script.defer) {
    attrs.push('defer');
  }

  // Add any additional attributes
  if (script.attributes) {
    for (const [key, value] of Object.entries(script.attributes)) {
      attrs.push(`${escapeAttr(key)}="${escapeAttr(value)}"`);
    }
  }

  const attrString = attrs.length > 0 ? ' ' + attrs.join(' ') : '';

  if (script.content) {
    return `<script${attrString}>\n${script.content}\n</script>`;
  }

  return `<script${attrString}></script>`;
}

/**
 * Escape HTML attribute value
 */
function escapeAttr(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

/**
 * Collect scripts from extensions and additional options
 */
function collectScripts(
  options: ExtensionScriptsPluginOptions,
  location: 'head' | 'body'
): ScriptDefinition[] {
  const scripts: ScriptDefinition[] = [];

  // Collect from extensions
  if (options.extensions) {
    for (const ext of options.extensions) {
      const extScripts = location === 'head' ? ext.headScripts : ext.bodyScripts;
      if (extScripts) {
        scripts.push(...extScripts);
      }
    }
  }

  // Add additional scripts
  const additionalScripts = location === 'head' ? options.headScripts : options.bodyScripts;
  if (additionalScripts) {
    scripts.push(...additionalScripts);
  }

  return scripts;
}

/**
 * Vite plugin for injecting extension scripts into HTML.
 */
export function extensionScriptsPlugin(
  options: ExtensionScriptsPluginOptions = {}
): Plugin {
  const { verbose = false } = options;

  return {
    name: 'vite-plugin-extension-scripts',

    transformIndexHtml: {
      order: 'post',
      handler(html: string, ctx: IndexHtmlTransformContext) {
        const headScripts = collectScripts(options, 'head');
        const bodyScripts = collectScripts(options, 'body');

        if (verbose) {
          console.log(`[extension-scripts] Injecting ${headScripts.length} head scripts`);
          console.log(`[extension-scripts] Injecting ${bodyScripts.length} body scripts`);
        }

        let result = html;

        // Inject head scripts before </head>
        if (headScripts.length > 0) {
          const headHtml = headScripts
            .map(scriptToHtml)
            .map(s => '    ' + s) // Indent for readability
            .join('\n');

          const headComment = '<!-- EXTENSION_HEAD_SCRIPTS -->';
          if (result.includes(headComment)) {
            // Replace placeholder comment if present
            result = result.replace(headComment, headHtml);
          } else {
            // Otherwise inject before </head>
            result = result.replace('</head>', `${headHtml}\n  </head>`);
          }

          if (verbose) {
            console.log('[extension-scripts] Head scripts injected');
          }
        }

        // Inject body scripts before </body>
        if (bodyScripts.length > 0) {
          const bodyHtml = bodyScripts
            .map(scriptToHtml)
            .map(s => '    ' + s) // Indent for readability
            .join('\n');

          const bodyComment = '<!-- EXTENSION_BODY_SCRIPTS -->';
          if (result.includes(bodyComment)) {
            // Replace placeholder comment if present
            result = result.replace(bodyComment, bodyHtml);
          } else {
            // Otherwise inject before </body>
            result = result.replace('</body>', `${bodyHtml}\n  </body>`);
          }

          if (verbose) {
            console.log('[extension-scripts] Body scripts injected');
          }
        }

        return result;
      },
    },
  };
}

export default extensionScriptsPlugin;
