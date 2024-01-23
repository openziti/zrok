// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Zrok',
  staticDirectories: ['static', '../docs/images', '../docker/compose'],
  tagline: 'Globally distributed reverse proxy',
  url: 'https://docs.zrok.io',
  baseUrl: '/',
  trailingSlash: true,
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/space-ziggy.png',

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
            to: '/docs/guides/self-hosting/metrics-and-limits/configuring-limits',
            from: ['/docs/guides/metrics-and-limits/configuring-limits'],
          },
          {
            to: '/docs/guides/self-hosting/metrics-and-limits/configuring-metrics',
            from: ['/docs/guides/metrics-and-limits/configuring-metrics'],
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
  ],
  
  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl:
            'https://github.com/openziti/zrok/blob/main/docs',
          path: '../docs',
          include: ['**/*.md', '**/*.mdx'],
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
          alt: 'Ziggy Goes to Space',
          src: 'img/space-ziggy.png',
          href: 'https://zrok.io',
          target: '_self',
        },
        items: [
          {
            type: 'doc',
            docId: 'getting-started',
            position: 'right',
            label: 'Docs',
          },
          {
            href: 'https://github.com/orgs/openziti/projects/16',
            label: 'Roadmap',
            position: 'right',
          },
          {
            href: 'https://github.com/openziti/zrok',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
        ],
        copyright: `Copyright © ${new Date().getFullYear()} NetFoundry Inc. Built with Docusaurus.`,
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
