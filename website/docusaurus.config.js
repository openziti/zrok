// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion
const path = require('path');

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');
const zrok = '/docs/zrok'

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'zrok',
  staticDirectories: ['static', 'docs/images', '../docker/compose', '../etc/caddy'],
  tagline: 'Globally distributed reverse proxy',
  url: 'https://docs.zrok.io',
  baseUrl: '/',
  trailingSlash: true,
  onBrokenLinks: 'throw',
  favicon: 'img/zrok-favicon.png',

  markdown: {
      hooks: {
          onBrokenMarkdownLinks: "throw",
      },
      mermaid: true
  },
  themes: ['@docusaurus/theme-mermaid'],

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'NetFoundry', // Usually your GitHub org/user name.
  projectName: 'zrok', // Usually your repo name.

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },
  plugins: [
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
    function myPlugin(context, options) {
      return {
        name: 'custom-webpack-plugin',
        configureWebpack(config, isServer, utils) {
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
    function aliasZrokRoot() {
      return {
        name: 'alias-zrok-root',
          configureWebpack() {
          return {
            resolve: {
              alias: {
                '@zrokroot': path.resolve(__dirname)    // same behavior as @zrokroot
              }
            }
          };
        }
      };
    }
  ],
  
  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: `${zrok}`,
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl:
            'https://github.com/openziti/zrok/blob/main/docs',
          path: 'docs',
          include: ['**/*.md', '**/*.mdx'],

          // Uncomment these lines when we're ready to show the 1.0 docs by default
          lastVersion: 'current',
          versions: {
            current: {
              label: '1.1',
            },
          },
            remarkPlugins: [
                function forbidSite() {
                    return (tree, file) => {
                        const src = String(file);
                        if (src.includes('@site')) {
                            throw new Error(`[FORBIDDEN] @zrokroot is not allowed in docs - use @zrokroot.\nFile: ${file.path}`);
                        }
                    };
                }
            ],
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
        pages: {
          path: './src/pages'
        },
        googleTagManager: {
          containerId: 'GTM-MDFLZPK8',
        },
        sitemap: {

        }
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
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
        links: [
        ],
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
        // The application ID provided by Algolia
        appId: 'CO73R59OLO',
  
        // Public API key: it is safe to commit it
        apiKey: '489572e91d0a750d34c127c2071ef962',
  
        indexName: 'zrok',
  
        // Optional: see doc section below
        contextualSearch: true,
  
        // Optional: Specify domains where the navigation should occur through window.location instead on history.push. Useful when our Algolia config crawls multiple documentation sites and we want to navigate with window.location.href to them.
        // externalUrlRegex: 'external\\.example\\.com|thirdparty\\.example\\.com',
  
        // Optional: Algolia search parameters
        searchParameters: {},
  
        // Optional: path for search page that enabled by default (`false` to disable it)
        searchPagePath: 'search',
  
        //... other Algolia params
      },
    }),
};

module.exports = config;
