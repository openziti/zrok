"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[9585],{7793:e=>{e.exports=JSON.parse('{"version":{"pluginId":"default","version":"current","label":"Next","banner":null,"badge":false,"noIndex":false,"className":"docs-version-current","isLast":true,"docsSidebars":{"tutorialSidebar":[{"type":"link","label":"Getting Started","href":"/docs/getting-started","docId":"getting-started","unlisted":false},{"type":"category","label":"Concepts","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Private Shares","href":"/docs/concepts/sharing-private","docId":"concepts/sharing-private","unlisted":false},{"type":"link","label":"Public Shares","href":"/docs/concepts/sharing-public","docId":"concepts/sharing-public","unlisted":false},{"type":"link","label":"Reserved Shares","href":"/docs/concepts/sharing-reserved","docId":"concepts/sharing-reserved","unlisted":false},{"type":"link","label":"Sharing HTTP Servers","href":"/docs/concepts/http","docId":"concepts/http","unlisted":false},{"type":"link","label":"Sharing TCP and UDP Servers","href":"/docs/concepts/tunnels","docId":"concepts/tunnels","unlisted":false},{"type":"link","label":"Sharing Websites and Files","href":"/docs/concepts/files","docId":"concepts/files","unlisted":false},{"type":"link","label":"Open Source","href":"/docs/concepts/opensource","docId":"concepts/opensource","unlisted":false},{"type":"link","label":"Hosting","href":"/docs/concepts/hosting","docId":"concepts/hosting","unlisted":false}],"href":"/docs/concepts/"},{"type":"category","label":"Guides","collapsible":true,"collapsed":true,"items":[{"type":"category","label":"Install","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Linux","href":"/docs/guides/install/linux","docId":"guides/install/linux","unlisted":false},{"type":"link","label":"macOS","href":"/docs/guides/install/macos","docId":"guides/install/macos","unlisted":false},{"type":"link","label":"Windows","href":"/docs/guides/install/windows","docId":"guides/install/windows","unlisted":false}],"href":"/docs/guides/install/"},{"type":"link","label":"frontdoor","href":"/docs/guides/frontdoor","docId":"guides/frontdoor","unlisted":false},{"type":"link","label":"Permission Modes","href":"/docs/guides/permission-modes","docId":"guides/permission-modes","unlisted":false},{"type":"category","label":"Docker Share","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Public Share","href":"/docs/guides/docker-share/docker_public_share_guide","docId":"guides/docker-share/docker_public_share_guide","unlisted":false},{"type":"link","label":"Private Share","href":"/docs/guides/docker-share/docker_private_share_guide","docId":"guides/docker-share/docker_private_share_guide","unlisted":false}],"href":"/docs/guides/docker-share/"},{"type":"category","label":"Self Hosting","collapsible":true,"collapsed":true,"items":[{"type":"category","label":"Linux","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"NGINX TLS","href":"/docs/guides/self-hosting/linux/nginx","docId":"guides/self-hosting/linux/nginx","unlisted":false}],"href":"/docs/guides/self-hosting/linux/"},{"type":"link","label":"Interstitial Pages","href":"/docs/guides/self-hosting/interstitial-page","docId":"guides/self-hosting/interstitial-page","unlisted":false},{"type":"link","label":"Personalized Frontend","href":"/docs/guides/self-hosting/personalized-frontend","docId":"guides/self-hosting/personalized-frontend","unlisted":false},{"type":"link","label":"Docker","href":"/docs/guides/self-hosting/docker","docId":"guides/self-hosting/docker","unlisted":false},{"type":"link","label":"Kubernetes","href":"/docs/guides/self-hosting/kubernetes","docId":"guides/self-hosting/kubernetes","unlisted":false},{"type":"category","label":"Metrics and Limits","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Configuring Metrics","href":"/docs/guides/self-hosting/metrics-and-limits/configuring-metrics","docId":"guides/self-hosting/metrics-and-limits/configuring-metrics","unlisted":false},{"type":"link","label":"Configuring Limits","href":"/docs/guides/self-hosting/metrics-and-limits/configuring-limits","docId":"guides/self-hosting/metrics-and-limits/configuring-limits","unlisted":false}],"href":"/docs/category/metrics-and-limits"},{"type":"category","label":"OAuth","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"OAuth Public Frontend Configuration","href":"/docs/guides/self-hosting/oauth/configuring-oauth","docId":"guides/self-hosting/oauth/configuring-oauth","unlisted":false}],"href":"/docs/category/oauth"},{"type":"link","label":"Instance Config","href":"/docs/guides/self-hosting/instance-configuration","docId":"guides/self-hosting/instance-configuration","unlisted":false}],"href":"/docs/category/self-hosting"},{"type":"category","label":"drives","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"The Drives CLI","href":"/docs/guides/drives/cli","docId":"guides/drives/cli","unlisted":false}]},{"type":"link","label":"VPN","href":"/docs/guides/vpn/","docId":"guides/vpn/vpn","unlisted":false}],"href":"/docs/category/guides"}]},"docs":{"concepts/files":{"id":"concepts/files","title":"Sharing Websites and Files","description":"With zrok it is possible to share files quickly and easily as well. To share files using zrok use","sidebar":"tutorialSidebar"},"concepts/hosting":{"id":"concepts/hosting","title":"Hosting","description":"Self-Hosted","sidebar":"tutorialSidebar"},"concepts/http":{"id":"concepts/http","title":"Sharing HTTP Servers","description":"zrok can share HTTP and HTTPS resources natively. If you have an existing web server that you want to share with other users, you can use the zrok share command using the --backend-mode proxy flag.","sidebar":"tutorialSidebar"},"concepts/index":{"id":"concepts/index","title":"Concepts","description":"zrok was designed to make sharing local resources both secure and easy. In this section of the zrok documentation, we\'ll tour through all of the most important features.","sidebar":"tutorialSidebar"},"concepts/opensource":{"id":"concepts/opensource","title":"Open Source","description":"It\'s important to the zrok project that it remain free and open source software. The code is available on GitHub","sidebar":"tutorialSidebar"},"concepts/sharing-private":{"id":"concepts/sharing-private","title":"Private Shares","description":"zrok was built to share and access digital resources. A private share allows a resource to be","sidebar":"tutorialSidebar"},"concepts/sharing-public":{"id":"concepts/sharing-public","title":"Public Shares","description":"zrok supports public sharing for web-based (HTTP and HTTPS) resources. These resources are easily shared with the general internet through public access points.","sidebar":"tutorialSidebar"},"concepts/sharing-reserved":{"id":"concepts/sharing-reserved","title":"Reserved Shares","description":"By default a public or private share is assigned a share token when you create a share using the zrok share command. The zrok share command is the bridge between your local environment and the users you are sharing with. When you terminate the zrok share, the bridge is eliminated and the share token is deleted. If you run zrok share again, you will be allocated a brand new share token.","sidebar":"tutorialSidebar"},"concepts/tunnels":{"id":"concepts/tunnels","title":"Sharing TCP and UDP Servers","description":"zrok includes support for sharing low-level TCP and UDP network resources using the tcpTunnel and udpTunnel backend modes.","sidebar":"tutorialSidebar"},"getting-started":{"id":"getting-started","title":"Getting Started with zrok","description":"Choose Your Path","sidebar":"tutorialSidebar"},"guides/docker-share/docker_private_share_guide":{"id":"guides/docker-share/docker_private_share_guide","title":"Docker Private Share","description":"Goal","sidebar":"tutorialSidebar"},"guides/docker-share/docker_public_share_guide":{"id":"guides/docker-share/docker_public_share_guide","title":"Docker Compose Public Share","description":"Goal","sidebar":"tutorialSidebar"},"guides/docker-share/index":{"id":"guides/docker-share/index","title":"Getting Started with Docker","description":"Overview","sidebar":"tutorialSidebar"},"guides/drives/cli":{"id":"guides/drives/cli","title":"The Drives CLI","description":"The zrok drives CLI tools allow for simple, ergonomic management and synchronization of local and remote files.","sidebar":"tutorialSidebar"},"guides/frontdoor":{"id":"guides/frontdoor","title":"zrok frontdoor","description":"zrok frontdoor is the heavy-duty front door to your app or site. It makes your website or app available to your online audience through the shield of zrok.io\'s hardened, managed frontends.","sidebar":"tutorialSidebar"},"guides/install/index":{"id":"guides/install/index","title":"Install","description":"<DownloadCard","sidebar":"tutorialSidebar"},"guides/install/linux":{"id":"guides/install/linux","title":"Install zrok in Linux","description":"Linux Binary","sidebar":"tutorialSidebar"},"guides/install/macos":{"id":"guides/install/macos","title":"Install zrok in macOS","description":"Darwin Binary","sidebar":"tutorialSidebar"},"guides/install/windows":{"id":"guides/install/windows","title":"Install zrok in Windows","description":"Windows Binary","sidebar":"tutorialSidebar"},"guides/permission-modes":{"id":"guides/permission-modes","title":"Permission Modes","description":"Shares created in zrok v0.4.26 and newer now include a choice of permission mode.","sidebar":"tutorialSidebar"},"guides/self-hosting/docker":{"id":"guides/self-hosting/docker","title":"Self-hosting guide for Docker","description":"","sidebar":"tutorialSidebar"},"guides/self-hosting/instance-configuration":{"id":"guides/self-hosting/instance-configuration","title":"Use Another zrok Instance","description":"This guide is relevant if you are self-hosting or using a friend\'s zrok instance instead of using zrok-as-a-service from zrok.io.","sidebar":"tutorialSidebar"},"guides/self-hosting/interstitial-page":{"id":"guides/self-hosting/interstitial-page","title":"Interstitial Pages","description":"On large zrok installations that support open registration and shared public frontends, abuse can become an issue. In order to mitigate phishing and other similar forms of abuse, zrok offers an interstitial page that announces to the visiting user that the share is hosted through zrok, and probably isn\'t their financial institution.","sidebar":"tutorialSidebar"},"guides/self-hosting/kubernetes":{"id":"guides/self-hosting/kubernetes","title":"Self-host a zrok Instance in Kubernetes","description":"The Helm chart for zrok is available from the main OpenZiti charts repo.","sidebar":"tutorialSidebar"},"guides/self-hosting/linux/index":{"id":"guides/self-hosting/linux/index","title":"Self-Hosting Guide for Linux","description":"Walkthrough Video","sidebar":"tutorialSidebar"},"guides/self-hosting/linux/nginx":{"id":"guides/self-hosting/linux/nginx","title":"NGINX Reverse Proxy for zrok","description":"Walkthrough Video","sidebar":"tutorialSidebar"},"guides/self-hosting/metrics-and-limits/configuring-limits":{"id":"guides/self-hosting/metrics-and-limits/configuring-limits","title":"Configuring Limits","description":"This guide is current as of zrok version v0.4.31.","sidebar":"tutorialSidebar"},"guides/self-hosting/metrics-and-limits/configuring-metrics":{"id":"guides/self-hosting/metrics-and-limits/configuring-metrics","title":"Configuring Metrics","description":"A fully configured, production-scale zrok service instance looks like this:","sidebar":"tutorialSidebar"},"guides/self-hosting/oauth/configuring-oauth":{"id":"guides/self-hosting/oauth/configuring-oauth","title":"OAuth Public Frontend Configuration","description":"As of v0.4.7, zrok includes OAuth integration for both Google and GitHub for zrok access public public frontends.","sidebar":"tutorialSidebar"},"guides/self-hosting/personalized-frontend":{"id":"guides/self-hosting/personalized-frontend","title":"Personalized Frontend","description":"This guide describes an approach that enables a zrok user to use a hosted, shared instance (zrok.io) and configure their own personalized frontend, which enables custom DNS and TLS for their shares.","sidebar":"tutorialSidebar"},"guides/vpn/vpn":{"id":"guides/vpn/vpn","title":"zrok VPN Guide","description":"zrok VPN backend allows for simple host-to-host VPN setup.","sidebar":"tutorialSidebar"}}}}')}}]);