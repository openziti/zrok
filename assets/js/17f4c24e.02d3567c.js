"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[8875],{4185:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>u,contentTitle:()=>d,default:()=>x,frontMatter:()=>c,metadata:()=>r,toc:()=>h});const r=JSON.parse('{"id":"guides/install/macos","title":"Install zrok in macOS","description":"Darwin Binary","source":"@site/versioned_docs/version-0.4/guides/install/macos.mdx","sourceDirName":"guides/install","slug":"/guides/install/macos","permalink":"/docs/guides/install/macos","draft":false,"unlisted":false,"editUrl":"https://github.com/openziti/zrok/blob/main/docs/versioned_docs/version-0.4/guides/install/macos.mdx","tags":[],"version":"0.4","frontMatter":{"title":"Install zrok in macOS","sidebar_label":"macOS"},"sidebar":"tutorialSidebar","previous":{"title":"Linux","permalink":"/docs/guides/install/linux"},"next":{"title":"Windows","permalink":"/docs/guides/install/windows"}}');var t=s(4848),o=s(8453),a=s(8151),i=s(595),l=s(1342);const c={title:"Install zrok in macOS",sidebar_label:"macOS"},d=void 0,u={},h=[{value:"Darwin Binary",id:"darwin-binary",level:2}];function m(e){const n={code:"code",h2:"h2",li:"li",ol:"ol",p:"p",pre:"pre",...(0,o.R)(),...e.components};return(0,t.jsxs)(t.Fragment,{children:[(0,t.jsx)(n.h2,{id:"darwin-binary",children:"Darwin Binary"}),"\n",(0,t.jsx)(a.F,{children:(0,t.jsx)("div",{className:l.A.downloadContainer,children:(0,t.jsx)(i.A,{osName:"macOS",osLogo:"/img/logo-apple.svg"})})}),"\n",(0,t.jsxs)(n.p,{children:["Download the binary distribution for your macOS architecture. For Intel Macs use the ",(0,t.jsx)(n.code,{children:"amd64"})," distribution. For Apple Silicon Macs use the ",(0,t.jsx)(n.code,{children:"arm64"})," distribution."]}),"\n",(0,t.jsxs)(n.ol,{children:["\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsx)(n.p,{children:"Unarchive the distribution in a temporary directory."}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"cd ~/Downloads && mkdir -p /tmp/zrok && tar -xf ./zrok*darwin*.tar.gz -C /tmp/zrok\n"})}),"\n"]}),"\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsxs)(n.p,{children:["Install the ",(0,t.jsx)(n.code,{children:"zrok"})," executable."]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"mkdir -p ~/bin && install /tmp/zrok/zrok ~/bin/\n"})}),"\n"]}),"\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsxs)(n.p,{children:["Add ",(0,t.jsx)(n.code,{children:"~/bin"})," to your shell's executable search path. Optionally add this to your ~/.zshenv to persist the change."]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"PATH=~/bin:$PATH\n"})}),"\n"]}),"\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsxs)(n.p,{children:["With the ",(0,t.jsx)(n.code,{children:"zrok"})," executable in your path, you can then execute the ",(0,t.jsx)(n.code,{children:"zrok"})," command from your shell:"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"zrok version\n"})}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-buttonless",metastring:'title="Output"',children:"               _    \n _____ __ ___ | | __\n|_  / '__/ _ \\| |/ /\n / /| | | (_) |   < \n/___|_|  \\___/|_|\\_\\\n\nv0.4.0 [c889005]\n"})}),"\n"]}),"\n"]})]})}function x(e={}){const{wrapper:n}={...(0,o.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(m,{...e})}):m(e)}},8151:(e,n,s)=>{s.d(n,{F:()=>i,d:()=>a});var r=s(6540),t=s(4848);const o=(0,r.createContext)([]),a=()=>(0,r.useContext)(o),i=e=>{let{children:n}=e;const[s,a]=(0,r.useState)([]);return(0,r.useEffect)((()=>{(async()=>{try{const e=await fetch("https://api.github.com/repos/openziti/zrok/releases/latest");if(!e.ok)throw new Error(`HTTP error! status: ${e.status}`);const n=(await e.json()).assets.map((e=>({name:e.name,url:e.browser_download_url,arch:e.name.replace(".tar.gz","").split("_")[3]})));console.log("Fetched assets:",n),a(n)}catch(e){console.error("Error fetching the release assets:",e)}})()}),[]),(0,t.jsx)(o.Provider,{value:s,children:n})}},595:(e,n,s)=>{s.d(n,{A:()=>l});s(6540);var r=s(8151),t=s(1342),o=s(5293),a=s(4848);const i=e=>{switch(e){case"amd64":return"x86_64";case"arm64":return"ARM64";case"armv7":return"ARM";default:return e.toUpperCase()}},l=e=>{let{osName:n,osLogo:s,infoText:l,guideLink:c}=e;const{colorMode:d}=(0,o.G)(),u=(0,r.d)();console.log("Assets in DownloadCard:",u);const h=(e=>{switch(e){case"Windows":return"windows";case"macOS":return"darwin";case"Linux":return"linux";default:return""}})(n),m=u.filter((e=>e.name.includes(h)));return console.log("Filtered assets for",n,"in DownloadCard:",m),(0,a.jsxs)("div",{className:t.A.downloadCard,children:[(0,a.jsx)("div",{className:t.A.imgContainer,children:(0,a.jsx)("img",{src:s,alt:`${n} logo`})}),(0,a.jsx)("h3",{children:n}),m.length>0&&(0,a.jsx)("ul",{children:m.map(((e,n)=>(0,a.jsx)("li",{className:t.A.downloadButtons,children:(0,a.jsx)("a",{href:e.url,className:t.A.downloadLinks,children:i(e.arch)})},n)))}),c&&(0,a.jsxs)("div",{className:t.A.cardFooter,children:[(0,a.jsx)("p",{children:l}),(0,a.jsx)("a",{href:c,children:"GUIDE"}),(0,a.jsx)("p",{})]})]})}},1342:(e,n,s)=>{s.d(n,{A:()=>r});const r={downloadContainer:"downloadContainer_nNgj",downloadCard:"downloadCard_D_EY",cardFooter:"cardFooter_Rhom",downloadButtons:"downloadButtons_NPAP",downloadLinks:"downloadLinks_thSu",imgContainer:"imgContainer_r0QA"}},8453:(e,n,s)=>{s.d(n,{R:()=>a,x:()=>i});var r=s(6540);const t={},o=r.createContext(t);function a(e){const n=r.useContext(o);return r.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function i(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:a(e.components),r.createElement(o.Provider,{value:n},e.children)}}}]);