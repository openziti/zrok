import * as path from 'path';
import type { PluginConfig } from '@docusaurus/types';
import { LogLevel, remarkScopedPath } from "@netfoundry/docusaurus-theme/plugins";

export function zrokRedirects(routeBasePath: string = 'docs/zrok'): PluginConfig {
    const p = '/' + routeBasePath;
    return [
        '@docusaurus/plugin-client-redirects',
        {
            redirects: [
                // getting-started → get-started
                { to: `${p}/get-started/`, from: [`${p}/getting-started`] },
                // category/guides → category/how-to-guides
                { to: `${p}/category/how-to-guides`, from: [`${p}/category/guides`] },
                // guides/* → how-tos/*
                { to: `${p}/how-tos/agent/`, from: [`${p}/guides/agent/`] },
                { to: `${p}/how-tos/agent/http-healthcheck`, from: [`${p}/guides/agent/http-healthcheck`] },
                { to: `${p}/how-tos/agent/linux-service`, from: [`${p}/guides/agent/linux-service`] },
                { to: `${p}/how-tos/agent/remoting`, from: [`${p}/guides/agent/remoting`] },
                { to: `${p}/how-tos/agent/windows-service/`, from: [`${p}/guides/agent/windows-service/`] },
                { to: `${p}/how-tos/docker-share/`, from: [`${p}/guides/docker-share/`] },
                { to: `${p}/how-tos/docker-share/private-share`, from: [`${p}/guides/docker-share/docker_private_share_guide`] },
                { to: `${p}/how-tos/docker-share/public-share`, from: [`${p}/guides/docker-share/docker_public_share_guide`] },
                { to: `${p}/how-tos/drives`, from: [`${p}/guides/drives`] },
                { to: `${p}/how-tos/frontdoor`, from: [`${p}/guides/frontdoor`] },
                { to: `${p}/how-tos/install/`, from: [`${p}/guides/install/`] },
                { to: `${p}/how-tos/install/linux`, from: [`${p}/guides/install/linux`] },
                { to: `${p}/how-tos/install/macos`, from: [`${p}/guides/install/macos`] },
                { to: `${p}/how-tos/install/windows`, from: [`${p}/guides/install/windows`] },
                { to: `${p}/how-tos/permission-modes`, from: [`${p}/guides/permission-modes`] },
                { to: `${p}/how-tos/v2-migration-guide`, from: [`${p}/guides/v2-migration-guide`] },
                { to: `${p}/how-tos/vpn`, from: [`${p}/guides/vpn`] },
                // guides/self-hosting/* → self-hosting/*
                { to: `${p}/self-hosting/docker`, from: [`${p}/guides/self-hosting/docker`] },
                { to: `${p}/self-hosting/dynamic-proxy`, from: [`${p}/guides/self-hosting/dynamicProxy`] },
                { to: `${p}/self-hosting/error-pages`, from: [`${p}/guides/self-hosting/error-pages`] },
                { to: `${p}/self-hosting/instance-configuration`, from: [`${p}/guides/self-hosting/instance-configuration`] },
                { to: `${p}/self-hosting/interstitial-page`, from: [`${p}/guides/self-hosting/interstitial-page`] },
                { to: `${p}/self-hosting/kubernetes`, from: [`${p}/guides/self-hosting/kubernetes`] },
                { to: `${p}/self-hosting/linux/`, from: [`${p}/guides/self-hosting/self_hosting_guide`, `${p}/guides/self-hosting/linux`] },
                { to: `${p}/self-hosting/linux/nginx`, from: [`${p}/guides/self-hosting/nginx_tls_guide/`, `${p}/guides/self-hosting/linux/nginx`] },
                { to: `${p}/self-hosting/metrics-and-limits/configuring-limits`, from: [`${p}/guides/metrics-and-limits/configuring-limits`, `${p}/guides/self-hosting/metrics-and-limits/configuring-limits`] },
                { to: `${p}/self-hosting/metrics-and-limits/configuring-metrics`, from: [`${p}/guides/metrics-and-limits/configuring-metrics`, `${p}/guides/self-hosting/metrics-and-limits/configuring-metrics`] },
                { to: `${p}/self-hosting/oauth/configuring-oauth`, from: [`${p}/guides/self-hosting/oauth/configuring-oauth`] },
                { to: `${p}/self-hosting/oauth/integrations/github`, from: [`${p}/guides/self-hosting/oauth/integrations/github`] },
                { to: `${p}/self-hosting/oauth/integrations/google`, from: [`${p}/guides/self-hosting/oauth/integrations/google`] },
                { to: `${p}/self-hosting/oauth/integrations/oidc`, from: [`${p}/guides/self-hosting/oauth/integrations/oidc`] },
                { to: `${p}/self-hosting/organizations`, from: [`${p}/guides/self-hosting/organizations`] },
                { to: `${p}/self-hosting/personalized-frontend`, from: [`${p}/guides/self-hosting/personalized-frontend`] },
                { to: `${p}/self-hosting/self-service-invite`, from: [`${p}/guides/self-hosting/self-service-invite`] },
            ],
        },
    ];
}

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
            lastVersion: 'current',
            includeCurrentVersion: true,
            versions: {
                'current': { label: '2.0 (Current)', path: '', banner: 'none' },
                '1.1': { label: '1.1', path: '1.1', banner: 'unmaintained' },
                '1.0': { label: '1.0', path: '1.0', banner: 'unmaintained' },
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
