import * as path from 'path';
import type { PluginConfig } from '@docusaurus/types';

export function zrokDocsPluginConfig(rootDir: string): PluginConfig {
    const zp = path.resolve(rootDir, 'docs');
    const zsbp = path.resolve(rootDir, 'sidebars.js');
    console.log('zrokDocsPluginConfig: zp=', zp);
    console.log('zrokDocsPluginConfig: sbp=', zsbp);
    return [
        '@docusaurus/plugin-content-docs',
        {
            id: 'zrok', // do not change - affects algolia search
            path: zp,
            routeBasePath: 'zrok',
            sidebarPath: zsbp,
            lastVersion: 'current',
            includeCurrentVersion: true,
            versions: {
                current: { label: '1.1 (Current)', path: '' },
                '1.0': { label: '1.0', path: '1.0' },
                '0.4': { label: '0.4', path: '0.4' },
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
            ],
        } as any,
    ];
}