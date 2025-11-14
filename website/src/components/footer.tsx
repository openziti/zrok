import {defaultNetFoundryFooterProps, defaultSocialProps} from "@openclint/docusaurus-shared/ui";

export const zrokFooter = {
    ...defaultNetFoundryFooterProps(),
    description:
        'zrok is an open-source, self-hostable sharing platform that simplifies shielding ' +
        'and sharing network services or files.',
    socialProps: {
        ...defaultSocialProps,
        githubUrl: 'https://github.com/netfoundry/',
        youtubeUrl: 'https://youtube.com/netfoundry/',
        linkedInUrl: 'https://www.linkedin.com/in/netfoundry/',
        twitterUrl: 'https://twitter.com/netfoundry/',
    },
    documentationLinks: [
        <a key="new" href="/docs/zrok/getting-started">Get started with zrok</a>
    ],
    communityLinks: [
        <a key="new" href="https://openziti.discourse.group/" target="_blank" rel="noopener noreferrer">Discourse Forum</a>
    ],
    resourceLinks: [
        <a href="https://blog.openziti.io">OpenZiti Tech Blog</a>,
        <a href="https://netfoundry.io/">NetFoundry</a>,
    ],
}
