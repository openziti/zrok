"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[4920],{722:(e,o,r)=>{r.r(o),r.d(o,{assets:()=>d,contentTitle:()=>t,default:()=>l,frontMatter:()=>i,metadata:()=>a,toc:()=>c});var n=r(5893),s=r(1151);const i={title:"Personalized Frontend",sidebar_label:"Personalized Frontend",sidebar_position:19},t=void 0,a={id:"guides/self-hosting/personalized-frontend",title:"Personalized Frontend",description:"This guide describes an approach that enables a zrok user to use a hosted, shared instance (zrok.io) and configure their own personalized frontend, which enables custom DNS and TLS for their shares.",source:"@site/../docs/guides/self-hosting/personalized-frontend.md",sourceDirName:"guides/self-hosting",slug:"/guides/self-hosting/personalized-frontend",permalink:"/docs/guides/self-hosting/personalized-frontend",draft:!1,unlisted:!1,editUrl:"https://github.com/openziti/zrok/blob/main/docs/../docs/guides/self-hosting/personalized-frontend.md",tags:[],version:"current",sidebarPosition:19,frontMatter:{title:"Personalized Frontend",sidebar_label:"Personalized Frontend",sidebar_position:19},sidebar:"tutorialSidebar",previous:{title:"Interstitial Pages",permalink:"/docs/guides/self-hosting/interstitial-page"},next:{title:"Docker",permalink:"/docs/guides/self-hosting/docker"}},d={},c=[{value:"Overview",id:"overview",level:2},{value:"Privacy",id:"privacy",level:2}];function h(e){const o={a:"a",admonition:"admonition",code:"code",em:"em",h2:"h2",img:"img",p:"p",pre:"pre",...(0,s.a)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(o.p,{children:"This guide describes an approach that enables a zrok user to use a hosted, shared instance (zrok.io) and configure their own personalized frontend, which enables custom DNS and TLS for their shares."}),"\n",(0,n.jsx)(o.p,{children:"In order to accomplish this, the user will need to provide their own minimal VPS instance, or container hosting. The size and capacity of these resources will be entirely dependent on the workload that they will be used to service. But generally, for most modest workloads, the most inexpensive VPS option will suffice."}),"\n",(0,n.jsx)(o.p,{children:"This approach gives you complete control over the way that your shares are exposed publicly. This approach works for HTTPS shares, and also for TCP and UDP ports, allowing you to put all of these things onto the public internet, while maintaining strong security for your protected resources."}),"\n",(0,n.jsxs)(o.p,{children:["This guide isn't a detailed ",(0,n.jsx)(o.em,{children:"how to"})," with specific steps to follow. This is more of a description of the overall concept. You'll want to figure out your own specific steps to implement this style of deployment in your own environment."]}),"\n",(0,n.jsx)(o.h2,{id:"overview",children:"Overview"}),"\n",(0,n.jsxs)(o.p,{children:["Let's imagine a hypothetical scenario where you've got 3 different resources shared using zrok. We'll refer to these as ",(0,n.jsx)(o.code,{children:"A"}),", ",(0,n.jsx)(o.code,{children:"B"}),", and ",(0,n.jsx)(o.code,{children:"C"}),". Both ",(0,n.jsx)(o.code,{children:"A"})," and ",(0,n.jsx)(o.code,{children:"B"})," are shares using the ",(0,n.jsx)(o.code,{children:"proxy"})," backend mode, which are used to share private HTTPS resources. Share ",(0,n.jsx)(o.code,{children:"C"})," uses the ",(0,n.jsx)(o.code,{children:"tcpTunnel"})," backend to expose a listening port from a private server (like a game server, or a message queue)."]}),"\n",(0,n.jsx)(o.p,{children:"We're using the shared zrok instance at zrok.io to provide our secure sharing infrastructure."}),"\n",(0,n.jsx)(o.p,{children:"Our deployment will end up looking like this:"}),"\n",(0,n.jsx)(o.p,{children:(0,n.jsx)(o.img,{alt:"personalized-frontend-1",src:r(4320).Z+"",width:"716",height:"357"})}),"\n",(0,n.jsxs)(o.p,{children:["We're using ",(0,n.jsx)(o.code,{children:"zrok reserve"})," to create the ",(0,n.jsx)(o.code,{children:"A"}),", ",(0,n.jsx)(o.code,{children:"B"}),", and ",(0,n.jsx)(o.code,{children:"C"})," shares as reserved shares (using the ",(0,n.jsx)(o.code,{children:"--unique-name"})," option to give them specific names). These shares could be located together in a single environment on a single host, or can be located at completely different spots on the planet on completely different hosts. You could want to use significantly more shares than 3, or less. The secure sharing fabric allows seamless secure connectivity for these shared resources. This implementation will scale up or down as needed (use multiple hosts behind a load balancer for really big workloads)."]}),"\n",(0,n.jsxs)(o.p,{children:["Because we're using ",(0,n.jsx)(o.code,{children:"private"})," zrok shares, they'll need to be accessed using a corresponding ",(0,n.jsx)(o.code,{children:"zrok access"})," private command. The ",(0,n.jsx)(o.code,{children:"zrok access private"}),' command binds a "network listener" where the share can be accessed on an address and port on the host where the command is executed. You can use ',(0,n.jsx)(o.code,{children:"zrok access private"})," to bind a network listener for a share in as many places as you want (up to the limit configuration of the service)."]}),"\n",(0,n.jsx)(o.admonition,{type:"note",children:(0,n.jsxs)(o.p,{children:["When you use ",(0,n.jsx)(o.code,{children:"zrok share public"}),", you are allowing your shared resources to be accessed using the shared, public frontend provided by the service instance (zrok.io). ",(0,n.jsx)(o.code,{children:"zrok share private"})," (or ",(0,n.jsx)(o.code,{children:"zrok reserve"}),"/",(0,n.jsx)(o.code,{children:"zrok share reserved"}),") creates the same kind of share, but does not provision the shared public frontend, and you'll need to use ",(0,n.jsx)(o.code,{children:"zrok access private"})," in order to ",(0,n.jsx)(o.em,{children:"bind"})," that share to a network address where it can be accessed."]})}),"\n",(0,n.jsxs)(o.p,{children:["Imagine that we own the domain ",(0,n.jsx)(o.code,{children:"example.com"}),". In our example, we want to expose our HTTPS shares ",(0,n.jsx)(o.code,{children:"A"})," and ",(0,n.jsx)(o.code,{children:"B"})," as ",(0,n.jsx)(o.code,{children:"a.example.com"})," and ",(0,n.jsx)(o.code,{children:"b.example.com"}),". And maybe our ",(0,n.jsx)(o.code,{children:"C"})," share represents a gaming server that we want to expose as ",(0,n.jsx)(o.code,{children:"gaming.example.com:25565"}),"."]}),"\n",(0,n.jsxs)(o.p,{children:["We can accomplish this easily with cheap VPS instance. You could also do it with containers through a container hosting service. The VPS will need an IP address exposed to the internet. You'll also need to be able to create DNS entries for the ",(0,n.jsx)(o.code,{children:"example.com"})," domain."]}),"\n",(0,n.jsxs)(o.p,{children:["To accomplish this, we're going to run 3 separate ",(0,n.jsx)(o.code,{children:"zrok access private"})," commands on our VPS (see the ",(0,n.jsx)(o.a,{href:"../../frontdoor/",children:"frontdoor guide"}),", or ",(0,n.jsx)(o.a,{href:"../../docker-share/docker_private_share_guide/#access-the-private-share",children:"zrok-private-access Docker Compose guide"})," for details on an approach for setting this up). One command each for shares ",(0,n.jsx)(o.code,{children:"A"}),", ",(0,n.jsx)(o.code,{children:"B"}),", and ",(0,n.jsx)(o.code,{children:"C"}),". The ",(0,n.jsx)(o.code,{children:"zrok access private"})," command works like this:"]}),"\n",(0,n.jsx)(o.pre,{children:(0,n.jsx)(o.code,{children:'$ zrok access private\nError: accepts 1 arg(s), received 0\nUsage:\n  zrok access private <shareToken> [flags]\n\nFlags:\n  -b, --bind string   The address to bind the private frontend (default "127.0.0.1:9191")\n      --headless      Disable TUI and run headless\n  -h, --help          help for private\n\nGlobal Flags:\n  -p, --panic     Panic instead of showing pretty errors\n  -v, --verbose   Enable verbose logging\n'})}),"\n",(0,n.jsxs)(o.p,{children:["Notice the ",(0,n.jsx)(o.code,{children:"--bind"})," flag. That flag is used to bind a network listener to a specific IP address and port on the host we're accessing the shares from. In this case, imagine our VPS node has a public IP address of ",(0,n.jsx)(o.code,{children:"1.2.3.4"})," and a loopback (",(0,n.jsx)(o.code,{children:"127.0.0.1"}),")."]}),"\n",(0,n.jsxs)(o.p,{children:["To expose our HTTPS shares, we're going to use a reverse proxy like nginx. The reverse proxy will be exposed to the internet, terminating TLS and reverse proxying ",(0,n.jsx)(o.code,{children:"a.example.com"})," and ",(0,n.jsx)(o.code,{children:"b.example.com"})," to the network listeners for shares ",(0,n.jsx)(o.code,{children:"A"})," and ",(0,n.jsx)(o.code,{children:"B"}),"."]}),"\n",(0,n.jsxs)(o.p,{children:["So, we'll configure our VPS to persistently launch a ",(0,n.jsx)(o.code,{children:"zrok access private"})," for both of these shares. We'll use the ",(0,n.jsx)(o.code,{children:"--bind"})," flag to bind ",(0,n.jsx)(o.code,{children:"A"})," to ",(0,n.jsx)(o.code,{children:"127.0.0.1:9191"})," and ",(0,n.jsx)(o.code,{children:"B"})," to ",(0,n.jsx)(o.code,{children:"127.0.0.1:9192"}),"."]}),"\n",(0,n.jsxs)(o.p,{children:["We'll then configure nginx to have a virtual host for ",(0,n.jsx)(o.code,{children:"a.example.com"}),", proxying that to ",(0,n.jsx)(o.code,{children:"127.0.0.1:9191"})," and ",(0,n.jsx)(o.code,{children:"b.example.com"}),", proxying that to ",(0,n.jsx)(o.code,{children:"127.0.0.1:9192"}),"."]}),"\n",(0,n.jsxs)(o.p,{children:["Exposing our TCP port for ",(0,n.jsx)(o.code,{children:"gaming.example.com"})," is simply a matter of running a third ",(0,n.jsx)(o.code,{children:"zrok access private"})," with a ",(0,n.jsx)(o.code,{children:"--bind"})," flag configured to point to ",(0,n.jsx)(o.code,{children:"1.2.3.4:25565"}),"."]}),"\n",(0,n.jsxs)(o.p,{children:["Once you've created the appropriate DNS entries for ",(0,n.jsx)(o.code,{children:"a.example.com"}),", ",(0,n.jsx)(o.code,{children:"b.example.com"}),", and ",(0,n.jsx)(o.code,{children:"gaming.example.com"})," and worked through the TLS configuration (letsencrypt is your friend here), you'll have a fully functional personalized frontend for your zrok shares that you control."]}),"\n",(0,n.jsx)(o.p,{children:"Your protected resources remain disconnected from the internet and are only reachable through your personalized endpoint."}),"\n",(0,n.jsx)(o.h2,{id:"privacy",children:"Privacy"}),"\n",(0,n.jsxs)(o.p,{children:["When you use a public frontend (with a simple ",(0,n.jsx)(o.code,{children:"zrok share public"}),") at a hosted zrok instance (like zrok.io), the operators of that service have some amount of visibility into what traffic you're sending to your shares. The load balancers in front of the public frontend maintain logs describing all of the URLs that were accessed, as well as other information (headers, etc.) that contain information about the resource you're sharing."]}),"\n",(0,n.jsxs)(o.p,{children:["If you create private shares using ",(0,n.jsx)(o.code,{children:"zrok share private"})," and then run your own ",(0,n.jsx)(o.code,{children:"zrok access private"})," from some other location, the operators of the zrok service instance only know that some amount of data moved between the environment running the ",(0,n.jsx)(o.code,{children:"zrok share private"})," and the ",(0,n.jsx)(o.code,{children:"zrok access private"}),". There is no other information available."]})]})}function l(e={}){const{wrapper:o}={...(0,s.a)(),...e.components};return o?(0,n.jsx)(o,{...e,children:(0,n.jsx)(h,{...e})}):h(e)}},4320:(e,o,r)=>{r.d(o,{Z:()=>n});const n=r.p+"assets/images/personalized-frontend-1-4a8782774dbbdff2247871d2064f51f9.png"},1151:(e,o,r)=>{r.d(o,{Z:()=>a,a:()=>t});var n=r(7294);const s={},i=n.createContext(s);function t(e){const o=n.useContext(i);return n.useMemo((function(){return"function"==typeof e?e(o):{...o,...e}}),[o,e])}function a(e){let o;return o=e.disableParentContext?"function"==typeof e.components?e.components(s):e.components||s:t(e.components),n.createElement(i.Provider,{value:o},e.children)}}}]);