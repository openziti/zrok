// Footer configuration for zrok - uses plain objects for docusaurus.config.js compatibility
export const zrokFooter = {
    description:
        'zrok is an open-source, self-hostable sharing platform that simplifies shielding ' +
        'and sharing network services or files.',
    socialProps: {
        githubUrl: 'https://github.com/openziti/zrok',
        youtubeUrl: 'https://youtube.com/netfoundry/',
        linkedInUrl: 'https://www.linkedin.com/company/netfoundry/',
        twitterUrl: 'https://twitter.com/netfoundry/',
    },
    documentationLinks: [
        { href: '/docs/zrok/getting-started', label: 'Get started with zrok' },
    ],
    communityLinks: [
        { href: 'https://openziti.discourse.group/', label: 'Discourse Forum' },
    ],
    resourceLinks: [
        { href: 'https://blog.openziti.io', label: 'OpenZiti Tech Blog' },
        { href: 'https://netfoundry.io/', label: 'NetFoundry' },
    ],
};
