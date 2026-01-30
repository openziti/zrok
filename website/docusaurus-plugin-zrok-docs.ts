import * as path from 'path';
import type { PluginConfig } from '@docusaurus/types';
import { LogLevel, remarkScopedPath } from "@netfoundry/docusaurus-theme/plugins";

export function zrokDocsPluginConfig(
    rootDir: string,
    linkMappings: { from: string; to: string }[],
    routeBasePath: string = 'docs/zrok'  // default for standalone zrok; unified-doc passes 'zrok'
): PluginConfig {
    const zp = path.resolve(rootDir, 'docs');
    const zsbp = path.resolve(rootDir, 'sidebars.ts');
    console.log('zrokDocsPluginConfig: zp=', zp);
    console.log('zrokDocsPluginConfig: sbp=', zsbp);
    console.log('zrokDocsPluginConfig: routeBasePath=', routeBasePath);
    return [
        '@docusaurus/plugin-content-docs',
        {
            id: 'zrok', // do not change - affects algolia search
            path: zp,
            routeBasePath,
            sidebarPath: zsbp,
            lastVersion: '1.1',
            includeCurrentVersion: true,
            versions: {
                current: { label: '2.x (Future)', path: '2.x', banner: 'unreleased' },
                '1.1': { label: '1.1 (Current)', path: '', banner: 'none' },
                '1.0': { label: '1.0', path: '1.0', banner: 'unmaintained' },
                '0.4': { label: '0.4', path: '0.4', banner: 'unmaintained' },
            },
            remarkPlugins: [
                function forbidSite() {
                    return (tree, file) => {
                        const src = String(file);
                        if (src.includes('@site')) {
                            throw new Error(`[FORBIDDEN] @site is not allowed in docs - use @zrokroot.\nFile: ${file.path}`);
                        }
                    };
                },
                [remarkScopedPath, { mappings: linkMappings, logLevel: LogLevel.Silent }],
            ],
        } as any,
    ];
}
