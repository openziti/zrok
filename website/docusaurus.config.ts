import type { Config } from '@docusaurus/types';
import type { ThemeConfig } from '@docusaurus/preset-classic';
import path from 'path';
import { zrokFooter } from './src/components/footer';
import {
    LogLevel,
    remarkCodeSections,
    remarkReplaceMetaUrl,
    remarkScopedPath,
    remarkYouTube
} from "@netfoundry/docusaurus-theme/plugins";
import {zrokDocsPluginConfig} from "./docusaurus-plugin-zrok-docs.ts";

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

const ZROK_ROOT = path.resolve(__dirname);
const resolvePath = (p: string) => path.resolve(ZROK_ROOT, p);

// absolute paths
const ZROK_CUSTOM_CSS = resolvePath('src/css/custom.css');
const ZROK_STATIC = resolvePath('static');
const ZROK_DOCS_IMAGES = resolvePath('docs/images');
const ZROK_DOCKER_COMPOSE = resolvePath('../docker/compose');
const ZROK_ETC = resolvePath('../etc');

const docsBase = '/';
const REMARK_MAPPINGS = [
    { from: '@zrokdocs', to: `${docsBase}zrok`},
];

// logs
console.log('ZROK_ROOT:', ZROK_ROOT);
console.log('ZROK_CUSTOM_CSS:', ZROK_CUSTOM_CSS);
console.log('ZROK_STATIC:', ZROK_STATIC);
console.log('ZROK_DOCS_IMAGES:', ZROK_DOCS_IMAGES);
console.log('ZROK_DOCKER_COMPOSE:', ZROK_DOCKER_COMPOSE);
console.log('ZROK_ETC_CADDY:', ZROK_ETC);

const zrok = '/zrok';

const config: Config = {
    title: 'zrok',
    staticDirectories: [ZROK_STATIC, ZROK_DOCS_IMAGES, ZROK_DOCKER_COMPOSE, ZROK_ETC],
    tagline: 'Globally distributed reverse proxy',
    url: 'https://docs.zrok.io',
    baseUrl: '/',
    trailingSlash: true,
    onBrokenLinks: 'throw',
    favicon: 'img/zrok-favicon.png',

    markdown: {
        hooks: {
            onBrokenMarkdownLinks: "throw"
        },
        mermaid: true,
    },

    themes: [
        '@docusaurus/theme-mermaid',
        [
            '@docusaurus/theme-classic',
            {
                customCss: ZROK_CUSTOM_CSS,
            }
        ],
        '@netfoundry/docusaurus-theme',
    ],

    // GitHub pages deployment config.
    organizationName: 'NetFoundry',
    projectName: 'zrok',

    i18n: {
        defaultLocale: 'en',
        locales: ['en'],
    },

    plugins: [
        function aliasZrokRoot() {
            return {
                name: 'alias-zrok-root',
                configureWebpack() {
                    return {
                        resolve: {
                            alias: { '@zrokroot': ZROK_ROOT }
                        },
                    };
                }
            };
        },
        [
            '@docusaurus/plugin-client-redirects',
            {
                redirects: [
                    {
                        to: `${zrok}/guides/self-hosting/linux`,
                        from: [`${zrok}/guides/self-hosting/self_hosting_guide`],
                    },
                    {
                        to: `${zrok}/guides/self-hosting/linux/nginx`,
                        from: [`${zrok}/guides/self-hosting/nginx_tls_guide/`]
                    },
                    {
                        to: `${zrok}/guides/self-hosting/metrics-and-limits/configuring-limits`,
                        from: [`${zrok}/guides/metrics-and-limits/configuring-limits`],
                    },
                    {
                        to: `${zrok}/guides/self-hosting/metrics-and-limits/configuring-metrics`,
                        from: [`${zrok}/guides/metrics-and-limits/configuring-metrics`],
                    }
                ]
            }
        ],
        function yamlLoaderPlugin() {
            return {
                name: 'custom-webpack-plugin',
                configureWebpack() {
                    return {
                        module: {
                            rules: [
                                {
                                    test: /\.yaml$/,
                                    use: 'yaml-loader',
                                },
                            ],
                        },
                    };
                },
            };
        },
        zrokDocsPluginConfig(ZROK_ROOT, [{ from: '@zrokdocs', to: `${docsBase}zrok` }]),
        [
            '@docusaurus/plugin-content-pages',
            {
                path: './src/pages'
            }
        ],
        [
            '@docusaurus/plugin-google-gtag',
            {
                trackingID: 'GTM-MDFLZPK8',
                anonymizeIP: true
            }
        ],
    ],

    themeConfig: {
        netfoundry: {
            showStarBanner: true,
            starBanner: {
                repoUrl: 'https://github.com/openziti/zrok',
                label: 'Star zrok on GitHub',
            },
            footer: zrokFooter,
        },
        navbar: {
            title: 'zrok',
            logo: {
                alt: 'zrok Logo',
                src: 'img/zrok-1.0.0-rocket-green.svg',
                href: 'https://zrok.io',
                target: '_self',
            },
            items: [
                {
                    type: 'docsVersionDropdown',
                    docsPluginId: 'zrok',
                },
                {
                    href: 'https://zrok.io/pricing/',
                    position: 'right',
                    label: 'pricing',
                },
                {
                    href: 'https://myzrok.io/',
                    position: 'right',
                    label: 'account',
                },
                {
                    href: 'https://github.com/orgs/openziti/projects/16',
                    label: 'roadmap',
                    position: 'right',
                },
                {
                    href: 'https://github.com/openziti/zrok',
                    position: 'right',
                    className: 'header-github-link',
                    title: 'GitHub'
                },
                {
                    href: 'https://openziti.discourse.group/',
                    position: 'right',
                    className: 'header-discourse-link',
                    title: 'Discourse'
                }
            ],
        },
        footer: {
            style: 'dark',
            links: [],
            copyright: `Copyright © ${new Date().getFullYear()} <a href="https://netfoundry.io">NetFoundry Inc.</a>`,
        },
        prism: {
            theme: lightCodeTheme,
            darkTheme: darkCodeTheme,
        },
        colorMode: {
            defaultMode: 'dark',
            disableSwitch: false,
            respectPrefersColorScheme: false,
        },
        docs: {
            sidebar: {
                autoCollapseCategories: true,
            }
        },
        algolia: {
            appId: 'CO73R59OLO',
            apiKey: '489572e91d0a750d34c127c2071ef962',
            indexName: 'zrok',
            contextualSearch: true,
            searchParameters: {},
            searchPagePath: 'search',
        },
    } satisfies ThemeConfig,
};

export default config;
