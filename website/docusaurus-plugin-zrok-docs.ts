import * as path from 'path';
import type { PluginConfig } from '@docusaurus/types';

export function zrokDocsPluginConfig(rootDir: string): PluginConfig {
    const sbp = path.resolve(rootDir, 'sidebars.js');
    console.log('zrokDocsPluginConfig: sbp=', sbp);
    return [
        '@docusaurus/plugin-content-docs',
        {
            id: 'zrok',
            path: path.resolve(rootDir, 'docs'),
            routeBasePath: 'docs/zrok',
            sidebarPath: sbp,
            lastVersion: 'current',
            includeCurrentVersion: true,
            versions: {
                current: { label: 'Current', path: '' },
                '1.0': { label: '1.0', path: '1.0' },
                '0.4': { label: '0.4', path: '0.4' },
            },
        } as any,
    ];
}