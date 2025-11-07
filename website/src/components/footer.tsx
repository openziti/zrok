import {defaultNetFoundryFooterProps, defaultSocialProps} from "@openclint/docusaurus-shared/ui";

export const zlanFooter = {
    ...defaultNetFoundryFooterProps(),
    description:
        'zLAN is a zero trust & micro-segmentation solution that makes it easy to create secure, software-defined networks.',
    socialProps: {
        ...defaultSocialProps,
        githubUrl: 'https://github.com/netfoundry/',
        youtubeUrl: 'https://youtube.com/netfoundry/',
        linkedInUrl: 'https://www.linkedin.com/in/netfoundry/',
        twitterUrl: 'https://twitter.com/netfoundry/',
    },
    documentationLinks: [
        <a key="new" href="/docs/zrok/guides/getting_started">Getting Started zLAN</a>
    ],
    communityLinks: [
        <a key="new" href="https://openziti.discourse.group/" target="_blank" rel="noopener noreferrer">Discourse Forum</a>
    ],
    resourceLinks: [
        <a href="https://blog.openziti.io">OpenZiti Tech Blog</a>,
        <a href="https://netfoundry.io/">NetFoundry</a>,
    ],
}
