"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[58],{2732:(e,s,n)=>{n.r(s),n.d(s,{assets:()=>c,contentTitle:()=>a,default:()=>l,frontMatter:()=>i,metadata:()=>t,toc:()=>d});var r=n(5893),o=n(1151);const i={sidebar_position:22,sidebar_label:"Permission Modes"},a="Permission Modes",t={id:"guides/permission-modes",title:"Permission Modes",description:"Shares created in zrok v0.4.26 and newer now include a choice of permission mode.",source:"@site/../docs/guides/permission-modes.md",sourceDirName:"guides",slug:"/guides/permission-modes",permalink:"/docs/guides/permission-modes",draft:!1,unlisted:!1,editUrl:"https://github.com/openziti/zrok/blob/main/docs/../docs/guides/permission-modes.md",tags:[],version:"current",sidebarPosition:22,frontMatter:{sidebar_position:22,sidebar_label:"Permission Modes"},sidebar:"tutorialSidebar",previous:{title:"frontdoor",permalink:"/docs/guides/frontdoor"},next:{title:"Docker Share",permalink:"/docs/category/docker-share"}},c={},d=[{value:"Creating a Share with Closed Permission Mode",id:"creating-a-share-with-closed-permission-mode",level:2},{value:"Adding and Removing Access Grants for Existing Shares",id:"adding-and-removing-access-grants-for-existing-shares",level:2},{value:"Limitations",id:"limitations",level:2}];function h(e){const s={code:"code",em:"em",h1:"h1",h2:"h2",p:"p",pre:"pre",...(0,o.a)(),...e.components};return(0,r.jsxs)(r.Fragment,{children:[(0,r.jsx)(s.h1,{id:"permission-modes",children:"Permission Modes"}),"\n",(0,r.jsxs)(s.p,{children:["Shares created in zrok ",(0,r.jsx)(s.code,{children:"v0.4.26"})," and newer now include a choice of ",(0,r.jsx)(s.em,{children:"permission mode"}),"."]}),"\n",(0,r.jsxs)(s.p,{children:["Shares created with zrok ",(0,r.jsx)(s.code,{children:"v0.4.25"})," and older were created using what is now called the ",(0,r.jsx)(s.em,{children:"open permission mode"}),". Whether ",(0,r.jsx)(s.em,{children:"public"})," or ",(0,r.jsx)(s.em,{children:"private"}),", these shares can be accessed by any user of the zrok service instance, as long as they know the ",(0,r.jsx)(s.em,{children:"share token"})," of the share. Effectively shares with the ",(0,r.jsx)(s.em,{children:"open permission mode"})," are accessible by any user of the zrok service instance."]}),"\n",(0,r.jsxs)(s.p,{children:["zrok now supports a ",(0,r.jsx)(s.em,{children:"closed permission mode"}),", which allows for more fine-grained control over which zrok users are allowed to privately access your shares using ",(0,r.jsx)(s.code,{children:"zrok access private"}),"."]}),"\n",(0,r.jsxs)(s.p,{children:["zrok defaults to continuing to create shares with the ",(0,r.jsx)(s.em,{children:"open permission mode"}),". This will likely change in a future release. We're leaving the default behavior in place to allow users a period of time to get comfortable with the new permission modes."]}),"\n",(0,r.jsx)(s.h2,{id:"creating-a-share-with-closed-permission-mode",children:"Creating a Share with Closed Permission Mode"}),"\n",(0,r.jsxs)(s.p,{children:["Adding the ",(0,r.jsx)(s.code,{children:"--closed"})," flag to the ",(0,r.jsx)(s.code,{children:"zrok share"})," or ",(0,r.jsx)(s.code,{children:"zrok reserve"})," commands will create shares using the ",(0,r.jsx)(s.em,{children:"closed permission mode"}),":"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"$ zrok share private --headless --closed -b web .\n[   0.066]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:\nzrok access private 0vzwzodf0c7g\n"})}),"\n",(0,r.jsxs)(s.p,{children:["By default any environment owned by the account that created the share is ",(0,r.jsx)(s.em,{children:"allowed"})," to access the new share. But a user trying to access the share from an environment owned by a different account will enounter the following error message:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"$ zrok access private 0vzwzodf0c7g\n[ERROR]: unable to access ([POST /access][401] accessUnauthorized)\n"})}),"\n",(0,r.jsxs)(s.p,{children:["The ",(0,r.jsx)(s.code,{children:"zrok share"})," and ",(0,r.jsx)(s.code,{children:"zrok reserve"})," commands now include an ",(0,r.jsx)(s.code,{children:"--access-grant"})," flag, which allows you to specify additional zrok accounts that are allowed to access your shares:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"$ zrok share private --headless --closed --access-grant anotheruser@test.com -b web .\n[   0.062]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:\nzrok access private y6h4at5xvn6o\n"})}),"\n",(0,r.jsxs)(s.p,{children:["And now ",(0,r.jsx)(s.code,{children:"anotheruser@test.com"})," will be allowed to access the share:"]}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"$ zrok access private --headless y6h4at5xvn6o\n[   0.049]    INFO main.(*accessPrivateCommand).run: allocated frontend 'VyvrJihAOEHD'\n[   0.051]    INFO main.(*accessPrivateCommand).run: access the zrok share at the following endpoint: http://127.0.0.1:9191\n"})}),"\n",(0,r.jsx)(s.h2,{id:"adding-and-removing-access-grants-for-existing-shares",children:"Adding and Removing Access Grants for Existing Shares"}),"\n",(0,r.jsxs)(s.p,{children:["If you've created a share (either reserved or ephemeral) and you forgot to include an access grant, or want to remove an access grant that was mistakenly added, you can use the ",(0,r.jsx)(s.code,{children:"zrok modify share"})," command to make the adjustments:"]}),"\n",(0,r.jsx)(s.p,{children:"Create a share:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"$ zrok share private --headless --closed -b web .\n[   0.064]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:\nzrok access private s4czjylwk7wa\n"})}),"\n",(0,r.jsx)(s.p,{children:"In another shell in the same environment you can execute:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"$ zrok modify share s4czjylwk7wa --add-access-grant anotheruser@test.com\nupdated\n"})}),"\n",(0,r.jsx)(s.p,{children:"And to remove the grant:"}),"\n",(0,r.jsx)(s.pre,{children:(0,r.jsx)(s.code,{children:"$ zrok modify share s4czjylwk7wa --remove-access-grant anotheruser@test.com\nupdated\n"})}),"\n",(0,r.jsx)(s.h2,{id:"limitations",children:"Limitations"}),"\n",(0,r.jsxs)(s.p,{children:["As of ",(0,r.jsx)(s.code,{children:"v0.4.26"})," there is currently no way to ",(0,r.jsx)(s.em,{children:"list"})," the current access grants. This will be addressed shortly in a subsequent update."]})]})}function l(e={}){const{wrapper:s}={...(0,o.a)(),...e.components};return s?(0,r.jsx)(s,{...e,children:(0,r.jsx)(h,{...e})}):h(e)}},1151:(e,s,n)=>{n.d(s,{Z:()=>t,a:()=>a});var r=n(7294);const o={},i=r.createContext(o);function a(e){const s=r.useContext(i);return r.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function t(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:a(e.components),r.createElement(i.Provider,{value:s},e.children)}}}]);