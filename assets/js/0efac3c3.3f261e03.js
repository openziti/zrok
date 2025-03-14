"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[8173],{9411:(e,n,s)=>{s.d(n,{Ay:()=>o,RM:()=>i});var r=s(4848),t=s(8453);const i=[];function a(e){const n={code:"code",pre:"pre",...(0,t.R)(),...e.components};return(0,r.jsx)(n.pre,{children:(0,r.jsx)(n.code,{className:"language-text",children:"brew install zrok\n"})})}function o(e={}){const{wrapper:n}={...(0,t.R)(),...e.components};return n?(0,r.jsx)(n,{...e,children:(0,r.jsx)(a,{...e})}):a(e)}},898:(e,n,s)=>{s.r(n),s.d(n,{assets:()=>m,contentTitle:()=>p,default:()=>b,frontMatter:()=>h,metadata:()=>r,toc:()=>x});const r=JSON.parse('{"id":"guides/install/linux","title":"Install zrok in Linux","description":"Install zrok from the Repository","source":"@site/versioned_docs/version-0.4/guides/install/linux.mdx","sourceDirName":"guides/install","slug":"/guides/install/linux","permalink":"/docs/0.4/guides/install/linux","draft":false,"unlisted":false,"editUrl":"https://github.com/openziti/zrok/blob/main/docs/versioned_docs/version-0.4/guides/install/linux.mdx","tags":[],"version":"0.4","frontMatter":{"title":"Install zrok in Linux","sidebar_label":"Linux"},"sidebar":"tutorialSidebar","previous":{"title":"Install","permalink":"/docs/0.4/guides/install/"},"next":{"title":"macOS","permalink":"/docs/0.4/guides/install/macos"}}');var t=s(4848),i=s(8453),a=s(8151),o=s(595),l=s(1342),c=s(6559),d=s(3902),u=s(9411);const h={title:"Install zrok in Linux",sidebar_label:"Linux"},p=void 0,m={},x=[{value:"Install <code>zrok</code> from the Repository",id:"install-zrok-from-the-repository",level:2},{value:"Homebrew",id:"homebrew",level:2},...u.RM,{value:"Linux Binary",id:"linux-binary",level:2}];function g(e){const n={a:"a",admonition:"admonition",code:"code",h2:"h2",li:"li",ol:"ol",p:"p",pre:"pre",...(0,i.R)(),...e.components},{Details:s}=n;return s||function(e,n){throw new Error("Expected "+(n?"component":"object")+" `"+e+"` to be defined: you likely forgot to import, pass, or provide it.")}("Details",!0),(0,t.jsxs)(t.Fragment,{children:[(0,t.jsxs)(n.h2,{id:"install-zrok-from-the-repository",children:["Install ",(0,t.jsx)(n.code,{children:"zrok"})," from the Repository"]}),"\n",(0,t.jsx)(n.p,{children:"This will configure the system to receive DEB or RPM package updates."}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"curl -sSf https://get.openziti.io/install.bash | sudo bash -s zrok\n"})}),"\n",(0,t.jsx)(n.admonition,{type:"info",children:(0,t.jsxs)(n.p,{children:["Check out ",(0,t.jsx)(n.a,{href:"/docs/0.4/guides/frontdoor?os=Linux",children:"zrok frontdoor"})," for running ",(0,t.jsx)(n.code,{children:"zrok"})," as an always-on service."]})}),"\n",(0,t.jsxs)(s,{children:[(0,t.jsx)("summary",{children:"Ansible Playbook"}),(0,t.jsxs)(d.A,{title:"Set up package repository and install zrok",children:[c.A,"\n- name: Install zrok package\ngather_facts: false\nhosts: all \nbecome: true\ntasks:\n- name: Install zrok\n  ansible.builtin.package:\n    name: zrok\n    state: present\n"]})]}),"\n",(0,t.jsx)(n.h2,{id:"homebrew",children:"Homebrew"}),"\n",(0,t.jsx)(u.Ay,{}),"\n",(0,t.jsx)(n.h2,{id:"linux-binary",children:"Linux Binary"}),"\n",(0,t.jsx)(a.F,{children:(0,t.jsx)("div",{className:l.A.downloadContainer,children:(0,t.jsx)(o.A,{osName:"Linux",osLogo:"/img/logo-linux.svg"})})}),"\n",(0,t.jsxs)(n.p,{children:["Download the binary distribution for your Linux distribution's architecture or run the install script below to pick the correct CPU architecture automatically. For Intel and AMD 64-bit machines use the ",(0,t.jsx)(n.code,{children:"amd64"})," distribution. For Raspberry Pi use the ",(0,t.jsx)(n.code,{children:"arm64"})," distribution."]}),"\n",(0,t.jsxs)(s,{children:[(0,t.jsxs)("summary",{children:["Manually install in ",(0,t.jsx)(n.code,{children:"~/bin/zrok"})]}),(0,t.jsxs)(n.ol,{children:["\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsx)(n.p,{children:"Unarchive the distribution in a temporary directory."}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"mkdir /tmp/zrok && tar -xf ./zrok*linux*.tar.gz -C /tmp/zrok\n"})}),"\n"]}),"\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsxs)(n.p,{children:["Install the ",(0,t.jsx)(n.code,{children:"zrok"})," executable."]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"mkdir -p ~/bin && install /tmp/zrok/zrok ~/bin/\n"})}),"\n"]}),"\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsxs)(n.p,{children:["Add ",(0,t.jsx)(n.code,{children:"~/bin"})," to your shell's executable search path. Optionally add this to your ~/.zshenv to persist the change."]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"PATH=~/bin:$PATH\n"})}),"\n"]}),"\n",(0,t.jsxs)(n.li,{children:["\n",(0,t.jsxs)(n.p,{children:["With the ",(0,t.jsx)(n.code,{children:"zrok"})," executable in your path, you can then execute the ",(0,t.jsx)(n.code,{children:"zrok"})," command from your shell:"]}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"zrok version\n"})}),"\n",(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-buttonless",metastring:'title="Output"',children:"               _    \n _____ __ ___ | | __\n|_  / '__/ _ \\| |/ /\n / /| | | (_) |   < \n/___|_|  \\___/|_|\\_\\\n\nv0.4.0 [c889005]\n"})}),"\n"]}),"\n"]})]}),"\n",(0,t.jsxs)(s,{children:[(0,t.jsxs)("summary",{children:["Script to install binary in ",(0,t.jsx)(n.code,{children:"/usr/local/bin/zrok"})]}),(0,t.jsx)(n.p,{children:"This script auto-selects the correct architecture and may be helpful for Raspberry Pi users."}),(0,t.jsx)(n.pre,{children:(0,t.jsx)(n.code,{className:"language-text",children:"cd $(mktemp -d);\n\nZROK_VERSION=$(\n  curl -sSf https://api.github.com/repos/openziti/zrok/releases/latest \\\n  | jq -r '.tag_name'\n);\n\ncase $(uname -m) in\n  x86_64)         GOXARCH=amd64\n  ;;\n  aarch64|arm64)  GOXARCH=arm64\n  ;;\n  arm*)           GOXARCH=armv7\n  ;;\n  *)              echo \"ERROR: unknown arch '$(uname -m)'\" >&2\n                  exit 1\n  ;;\nesac;\n\ncurl -sSfL \\\n  \"https://github.com/openziti/zrok/releases/download/${ZROK_VERSION}/zrok_${ZROK_VERSION#v}_linux_${GOXARCH}.tar.gz\" \\\n  | tar -xz -f -;\n\nsudo install -o root -g root ./zrok /usr/local/bin/;\n\nzrok version;\n"})})]})]})}function b(e={}){const{wrapper:n}={...(0,i.R)(),...e.components};return n?(0,t.jsx)(n,{...e,children:(0,t.jsx)(g,{...e})}):g(e)}},8151:(e,n,s)=>{s.d(n,{F:()=>o,d:()=>a});var r=s(6540),t=s(4848);const i=(0,r.createContext)([]),a=()=>(0,r.useContext)(i),o=e=>{let{children:n}=e;const[s,a]=(0,r.useState)([]);return(0,r.useEffect)((()=>{(async()=>{try{const e=await fetch("https://api.github.com/repos/openziti/zrok/releases/latest");if(!e.ok)throw new Error(`HTTP error! status: ${e.status}`);const n=(await e.json()).assets.map((e=>({name:e.name,url:e.browser_download_url,arch:e.name.replace(".tar.gz","").split("_")[3]})));console.log("Fetched assets:",n),a(n)}catch(e){console.error("Error fetching the release assets:",e)}})()}),[]),(0,t.jsx)(i.Provider,{value:s,children:n})}},3902:(e,n,s)=>{s.d(n,{A:()=>a});s(6540);var r=s(382),t=s(1432),i=s(4848);const a=e=>{let{title:n,children:s}=e;const a=s.map((e=>"string"==typeof e?e.trim():r.Ay.dump(e).trim())).join("\n\n");return(0,i.jsx)("div",{children:(0,i.jsx)(t.A,{language:"yaml",title:n,children:a})})}},595:(e,n,s)=>{s.d(n,{A:()=>l});s(6540);var r=s(8151),t=s(1342),i=s(5293),a=s(4848);const o=e=>{switch(e){case"amd64":return"x86_64";case"arm64":return"ARM64";case"armv7":return"ARM";default:return e.toUpperCase()}},l=e=>{let{osName:n,osLogo:s,infoText:l,guideLink:c}=e;const{colorMode:d}=(0,i.G)(),u=(0,r.d)();console.log("Assets in DownloadCard:",u);const h=(e=>{switch(e){case"Windows":return"windows";case"macOS":return"darwin";case"Linux":return"linux";default:return""}})(n),p=u.filter((e=>e.name.includes(h)));return console.log("Filtered assets for",n,"in DownloadCard:",p),(0,a.jsxs)("div",{className:t.A.downloadCard,children:[(0,a.jsx)("div",{className:t.A.imgContainer,children:(0,a.jsx)("img",{src:s,alt:`${n} logo`})}),(0,a.jsx)("h3",{children:n}),p.length>0&&(0,a.jsx)("ul",{children:p.map(((e,n)=>(0,a.jsx)("li",{className:t.A.downloadButtons,children:(0,a.jsx)("a",{href:e.url,className:t.A.downloadLinks,children:o(e.arch)})},n)))}),c&&(0,a.jsxs)("div",{className:t.A.cardFooter,children:[(0,a.jsx)("p",{children:l}),(0,a.jsx)("a",{href:c,children:"GUIDE"}),(0,a.jsx)("p",{})]})]})}},1342:(e,n,s)=>{s.d(n,{A:()=>r});const r={downloadContainer:"downloadContainer_nNgj",downloadCard:"downloadCard_D_EY",cardFooter:"cardFooter_Rhom",downloadButtons:"downloadButtons_NPAP",downloadLinks:"downloadLinks_thSu",imgContainer:"imgContainer_r0QA"}},6559:(e,n,s)=>{s.d(n,{A:()=>r});const r=[{name:"Set up zrok Package Repo",gather_facts:!0,hosts:"all",become:!0,tasks:[{name:"Set up apt repo",when:'ansible_os_family == "Debian"',block:[{name:"Install playbook dependencies","ansible.builtin.package":{name:["gnupg"],state:"present"}},{name:"Fetch armored pubkey","ansible.builtin.uri":{url:"https://get.openziti.io/tun/package-repos.gpg",return_content:"yes"},register:"armored_pubkey"},{name:"Dearmor pubkey","ansible.builtin.shell":'gpg --dearmor --output /usr/share/keyrings/openziti.gpg <<< "{{ armored_pubkey.content }}"\n',args:{creates:"/usr/share/keyrings/openziti.gpg",executable:"/bin/bash"}},{name:"Set pubkey filemode","ansible.builtin.file":{path:"/usr/share/keyrings/openziti.gpg",mode:"a+rX"}},{name:"Install OpenZiti repo deb source","ansible.builtin.copy":{dest:"/etc/apt/sources.list.d/openziti-release.list",content:"deb [signed-by=/usr/share/keyrings/openziti.gpg] https://packages.openziti.org/zitipax-openziti-deb-stable debian main\n"}},{name:"Refresh Repo Sources","ansible.builtin.apt":{update_cache:"yes",cache_valid_time:3600}}]},{name:"Set up yum repo",when:'ansible_os_family == "RedHat"',block:[{name:"Install OpenZiti repo rpm source","ansible.builtin.yum_repository":{name:"OpenZitiRelease",description:"OpenZiti Release",baseurl:"https://packages.openziti.org/zitipax-openziti-rpm-stable/redhat/$basearch",enabled:"yes",gpgkey:"https://packages.openziti.org/zitipax-openziti-rpm-stable/redhat/$basearch/repodata/repomd.xml.key",repo_gpgcheck:"yes",gpgcheck:"no"}}]}]}]}}]);