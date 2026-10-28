(self.webpackChunk=self.webpackChunk||[]).push([[67],{36842:(function(d,A,t){"use strict";var k=t(55420),a=t.n(k),c=t(1974),g=t.n(c),_=t(98312),w=t.n(_),E=t(11604),m=t.n(E),i=t(96540),W=t(79273),b=t(16370),x=t(20873),C=t(22817),f=t(48212),B=t(47152),o=t(58679),e=t(62885),s=t(86441),l=t(74848),u={0:"\u5BC6\u7801\u9A8C\u8BC1",1:"\u8BC1\u4E66\u9A8C\u8BC1"},K=(0,i.forwardRef)(function(z,Q){var nn=z.authTypeData,G=z.layered,h=G===void 0?!1:G,r=z.onSuccess,tn=W.A.useApp(),M=tn.message,n=(0,i.useRef)({}),L=(0,i.useRef)(0),Z=(0,i.useRef)(!1),an=x.A.useForm(),un=m()(an,1),P=un[0],X=(0,i.useState)(),H=m()(X,2),v=H[0],en=H[1],Pn=(0,i.useState)(0),mn=m()(Pn,2),rn=mn[0],on=mn[1],Sn=(0,i.useState)(0),xn=m()(Sn,2),An=xn[0],ln=xn[1],kn=(0,i.useState)(!1),hn=m()(kn,2),Mn=hn[0],cn=hn[1],D=v!==void 0,wn=(0,i.useMemo)(function(){var j=Object.keys(nn||{}).length>0?nn:u;return Object.entries(j).map(function(U){var O=m()(U,2),J=O[0],N=O[1];return{value:Number(J),label:String(N)}})},[nn]),sn=!D||rn!==An,gn=(0,i.useCallback)(function(){L.current+=1,Z.current=!1,P.resetFields(),en(void 0),on(0),ln(0),cn(!1)},[P]),En=(0,i.useCallback)(function(j){var U;if(L.current+=1,Z.current=!1,P.resetFields(),j){var O=Number(j.auth_type||0);en(j),on(O),ln(O),P.setFieldsValue({name:j.name,ip:j.ip,port:j.port||22,user:j.user,auth_type:O,password:""})}else en(void 0),on(0),ln(0),P.setFieldsValue({name:"",ip:"",port:22,user:"",auth_type:0,password:""});(U=n.current)===null||U===void 0||U.show()},[P]),jn=(0,i.useCallback)(function(){var j;L.current+=1,Z.current=!1,cn(!1),(j=n.current)===null||j===void 0||j.hide()},[]);(0,i.useImperativeHandle)(Q,function(){return{open:En}},[En]);var fn=(0,i.useCallback)((function(){var j=w()(a()().mark(function U(O){var J,N,q,S,V,$,vn,bn,yn;return a()().wrap(function(I){for(;;)switch(I.prev=I.next){case 0:if(!Z.current){I.next=2;break}return I.abrupt("return");case 2:return Z.current=!0,cn(!0),J=++L.current,N=Number(O.auth_type||0),q=Number(O.port||22),S={name:String(O.name||"").trim(),ip:String(O.ip||"").trim(),port:D?String(q):q,user:String(O.user||"").trim()},(!D||N!==Number((v==null?void 0:v.auth_type)||0))&&(S.auth_type=D?String(N):N),V=String(O.password||""),V.trim()&&(S.password=V),D&&(S.ids=[v==null?void 0:v.id]),I.prev=12,I.next=15,(D?s.dZ:s.sA)({body:S});case 15:if($=I.sent,!(L.current!==J||!$)){I.next=18;break}return I.abrupt("return");case 18:($==null?void 0:$.code)===200?(bn=D&&v?g()(g()({},v),{},{name:S.name,ip:S.ip,port:q,user:S.user,auth_type:N,searchText:[S.name,S.ip,String(q),"".concat(S.ip,":").concat(q)].join(`
`).toLocaleLowerCase()}):void 0,yn=!!(D&&v&&(v.ip!==S.ip||v.port!==q||v.user!==S.user||v.auth_type!==N||V.trim())),(vn=n.current)===null||vn===void 0||vn.hide(),M.success(($==null?void 0:$.message)||(D?"\u4FEE\u6539\u6210\u529F":"\u6DFB\u52A0\u6210\u529F")),r==null||r({serverId:D?v==null?void 0:v.id:void 0,server:bn,reconnectRequired:yn})):M.error(($==null?void 0:$.message)||(D?"\u4FEE\u6539\u7EC8\u7AEF\u5931\u8D25":"\u6DFB\u52A0\u7EC8\u7AEF\u5931\u8D25"));case 19:return I.prev=19,L.current===J&&(Z.current=!1,cn(!1)),I.finish(19);case 22:case"end":return I.stop()}},U,null,[[12,,19,22]])}));return function(U){return j.apply(this,arguments)}})(),[D,M,r,v]);return(0,l.jsx)(e.Ht,{ref:n,title:(0,l.jsx)(e.hE,{children:D?"\u7F16\u8F91\u7EC8\u7AEF":"\u6DFB\u52A0\u7EC8\u7AEF"}),onOk:P.submit,onCancel:jn,width:440,modalProps:{confirmLoading:Mn,cancelButtonProps:{disabled:!1},zIndex:h?2e3:void 0,closable:!0,keyboard:!0,forceRender:!0,okText:D?"\u4FDD\u5B58":"\u6DFB\u52A0",cancelText:"\u53D6\u6D88",style:{top:100,paddingBottom:100},focusable:{focusTriggerAfterClose:h},mask:{closable:!0},afterClose:gn},children:(0,l.jsxs)(x.A,{form:P,layout:"vertical",variant:"filled",onFinish:fn,children:[(0,l.jsx)(x.A.Item,{name:"name",label:"\u7EC8\u7AEF\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u7EC8\u7AEF\u540D\u79F0"}],children:(0,l.jsx)(C.A,{placeholder:"\u4F8B\u5982\uFF1A\u751F\u4EA7\u7EC8\u7AEF",autoComplete:"off"})}),(0,l.jsxs)(B.A,{gutter:[16,0],children:[(0,l.jsx)(b.A,{xs:24,sm:17,children:(0,l.jsx)(x.A.Item,{name:"ip",label:"IP \u5730\u5740",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165 IP \u5730\u5740"}],children:(0,l.jsx)(C.A,{placeholder:"\u4F8B\u5982\uFF1A192.168.1.10",inputMode:"text",autoComplete:"off"})})}),(0,l.jsx)(b.A,{xs:24,sm:7,children:(0,l.jsx)(x.A.Item,{name:"port",label:"SSH \u7AEF\u53E3",rules:[{required:!0,message:"\u8BF7\u8F93\u5165\u7AEF\u53E3"}],children:(0,l.jsx)(f.A,{min:1,max:65535,precision:0,style:{width:"100%"}})})})]}),(0,l.jsxs)(B.A,{gutter:[16,0],children:[(0,l.jsx)(b.A,{xs:24,sm:12,children:(0,l.jsx)(x.A.Item,{name:"user",label:"\u767B\u5F55\u8D26\u53F7",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u767B\u5F55\u8D26\u53F7"}],children:(0,l.jsx)(C.A,{placeholder:"\u4F8B\u5982\uFF1Aroot",autoComplete:"username"})})}),(0,l.jsx)(b.A,{xs:24,sm:12,children:(0,l.jsx)(x.A.Item,{name:"auth_type",label:"\u8BA4\u8BC1\u65B9\u5F0F",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u8BA4\u8BC1\u65B9\u5F0F"}],children:(0,l.jsx)(o.A,{options:wn,onChange:function(U){on(Number(U)),P.setFieldValue("password","")}})})})]}),(0,l.jsx)(x.A.Item,{name:"password",label:rn===1?"\u79C1\u94A5":"\u767B\u5F55\u5BC6\u7801",rules:[{validator:(function(){var j=w()(a()().mark(function O(J,N){return a()().wrap(function(S){for(;;)switch(S.prev=S.next){case 0:if(!(sn&&!String(N||"").trim())){S.next=2;break}throw new Error(rn===1?"\u8BF7\u8F93\u5165\u79C1\u94A5":"\u8BF7\u8F93\u5165\u767B\u5F55\u5BC6\u7801");case 2:case"end":return S.stop()}},O)}));function U(O,J){return j.apply(this,arguments)}return U})()}],children:rn===1?(0,l.jsx)(C.A.TextArea,{autoSize:{minRows:6,maxRows:12},maxLength:128*1024,placeholder:D&&!sn?"\u7559\u7A7A\u5219\u4FDD\u6301\u4E0D\u53D8":"\u7C98\u8D34 PEM \u683C\u5F0F\u79C1\u94A5",autoComplete:"off"}):(0,l.jsx)(C.A.Password,{maxLength:4*1024,placeholder:D&&!sn?"\u7559\u7A7A\u5219\u4FDD\u6301\u4E0D\u53D8":"\u8BF7\u8F93\u5165\u767B\u5F55\u5BC6\u7801",autoComplete:"new-password"})})]})})});K.displayName="Form",A.A=(0,i.memo)(K)}),88820:(function(d,A,t){"use strict";var k=t(1974),a=t.n(k),c=t(11604),g=t.n(c),_=t(71640),w=t.n(_),E=t(96540),m=t(58487),i=t(74848),W=["server","initializing","unavailable","resetKey","onStatusChange","onEdit"],b=(0,E.forwardRef)(function(x,C){var f=x.server,B=x.initializing,o=x.unavailable,e=x.resetKey,s=x.onStatusChange,l=x.onEdit,u=w()(x,W),K=(0,E.useState)("idle"),z=g()(K,2),Q=z[0],nn=z[1],G=Q==="connected"||Q==="connecting",h=f.id===0&&o&&!G,r=!h&&f.id===0&&(!f.ip||!f.user),tn=B?"\u6B63\u5728\u52A0\u8F7D\u672C\u673A\u8FDE\u63A5":h?"\u672C\u673A\u8FDE\u63A5\u52A0\u8F7D\u5931\u8D25":r?"\u8BF7\u5148\u914D\u7F6E\u672C\u673A\u8FDE\u63A5":void 0,M=B?"\u6B63\u5728\u52A0\u8F7D\u672C\u673A\u8FDE\u63A5...":h?"\u672C\u673A\u8FDE\u63A5\u52A0\u8F7D\u5931\u8D25":f.ip&&f.user?"".concat(f.user,"@").concat(f.ip,":").concat(f.port):f.id===0?"\u5C1A\u672A\u914D\u7F6E\u672C\u673A SSH":"".concat(f.ip,":").concat(f.port);return(0,i.jsx)(m.A,a()(a()({},u),{},{ref:C,path:"/server/ssh",params:{id:f.id},binaryControl:!0,title:f.name||"\u7EC8\u7AEF",endpoint:M,loading:B,disabledReason:tn,disabledIcon:h?"WarningOutlined":"SettingOutlined",resetKey:JSON.stringify([e,f.auth_type,f.ip,f.port,f.user]),onStatusChange:function(L){nn(L),s==null||s(f.id,L)},errorAction:l?{label:"\u4FEE\u6539\u4FE1\u606F",codes:["credential_invalid","password_invalid","private_key_invalid"],onClick:function(){return l(f)}}:void 0}))});b.displayName="Terminal",A.A=(0,E.memo)(b)}),68701:(function(d,A,t){"use strict";t.d(A,{v:function(){return C}});var k=t(55420),a=t.n(k),c=t(45647),g=t.n(c),_=t(98312),w=t.n(_),E=t(11604),m=t.n(E),i=t(96540),W=t(79273),b=t(86441),x=function(o){var e=Number((o==null?void 0:o.id)||0),s=String((o==null?void 0:o.ip)||""),l=Number((o==null?void 0:o.port)||22),u=String((o==null?void 0:o.user)||""),K=String((o==null?void 0:o.name)||(e===0?"\u672C\u673A\u7EC8\u7AEF":""));return{id:e,ip:s,port:l,user:u,auth_type:Number((o==null?void 0:o.auth_type)||0),name:K,create_time:o!=null&&o.create_time?String(o.create_time):void 0,update_time:o!=null&&o.update_time?String(o.update_time):void 0,searchText:[K,s,String(l),"".concat(s,":").concat(l)].join(`
`).toLocaleLowerCase()}},C=x({id:0,ip:"127.0.0.1",port:22,user:"",auth_type:0,name:"\u672C\u673A\u7EC8\u7AEF"}),f=function(){var o=W.A.useApp(),e=o.message,s=(0,i.useRef)(0),l=(0,i.useRef)(0),u=(0,i.useState)([]),K=m()(u,2),z=K[0],Q=K[1],nn=(0,i.useState)(C),G=m()(nn,2),h=G[0],r=G[1],tn=(0,i.useState)(!1),M=m()(tn,2),n=M[0],L=M[1],Z=(0,i.useState)(!0),an=m()(Z,2),un=an[0],P=an[1],X=(0,i.useState)(!1),H=m()(X,2),v=H[0],en=H[1],Pn=(0,i.useState)(!0),mn=m()(Pn,2),rn=mn[0],on=mn[1],Sn=(0,i.useState)(!1),xn=m()(Sn,2),An=xn[0],ln=xn[1],kn=(0,i.useState)(!1),hn=m()(kn,2),Mn=hn[0],cn=hn[1],D=(0,i.useState)(!1),wn=m()(D,2),sn=wn[0],gn=wn[1],En=(0,i.useState)(""),jn=m()(En,2),fn=jn[0],j=jn[1],U=(0,i.useState)(""),O=m()(U,2),J=O[0],N=O[1],q=(0,i.useState)(""),S=m()(q,2),V=S[0],$=S[1],vn=(0,i.useState)(0),bn=m()(vn,2),yn=bn[0],Tn=bn[1],I=(0,i.useRef)(!1);(0,i.useEffect)(function(){var T=window.setTimeout(function(){$(J.trim())},300);return function(){return window.clearTimeout(T)}},[J]),(0,i.useEffect)(function(){var T=++s.current,Y=!0,F=function(){return Y&&s.current===T};return on(!0),ln(!1),I.current=!1,gn(!1),j(""),(0,b.Sq)({body:{page_size:100,order_by_field:"id",order_by_type:"desc",keyword:V}}).then(function(p){var dn,pn,y,_n;if(!(!F()||!p)){if((p==null?void 0:p.code)!==200){e.error((p==null?void 0:p.message)||"\u83B7\u53D6\u7EC8\u7AEF\u5217\u8868\u5931\u8D25");return}var Dn=Array.isArray(p==null||(dn=p.data)===null||dn===void 0?void 0:dn.list)?p==null||(pn=p.data)===null||pn===void 0?void 0:pn.list.filter(Boolean):[];Q(Dn.map(x)),gn(!!(!(p==null||(y=p.data)===null||y===void 0)&&y.has_more)),j(String((p==null||(_n=p.data)===null||_n===void 0?void 0:_n.next_cursor_id)||""))}}).catch(function(){F()&&e.error("\u83B7\u53D6\u7EC8\u7AEF\u5217\u8868\u5931\u8D25")}).finally(function(){F()&&(on(!1),cn(!0))}),function(){Y=!1,s.current===T&&(s.current+=1,I.current=!1)}},[e,V,yn]),(0,i.useEffect)(function(){var T=++l.current,Y=!0,F=function(){return Y&&l.current===T};return P(!0),en(!1),(0,b.y8)({body:{id:0}}).then(function(p){if(F()){if((p==null?void 0:p.code)===200&&p!==null&&p!==void 0&&p.data){r(x(p==null?void 0:p.data)),L(!0),en(!1);return}en(!0),p&&e.error((p==null?void 0:p.message)||"\u83B7\u53D6\u672C\u673A\u8FDE\u63A5\u5931\u8D25")}}).catch(function(){F()&&(en(!0),e.error("\u83B7\u53D6\u672C\u673A\u8FDE\u63A5\u5931\u8D25"))}).finally(function(){F()&&P(!1)}),function(){Y=!1,l.current===T&&(l.current+=1)}},[e,yn]);var Cn=(0,i.useCallback)(function(){s.current+=1,l.current+=1,Tn(function(T){return T+1})},[]),Rn=(0,i.useCallback)(function(T){if(T.id===0){r(T),L(!0),en(!1);return}Q(function(Y){return Y.map(function(F){return F.id===T.id?T:F})})},[]),Bn=(0,i.useCallback)(w()(a()().mark(function T(){var Y,F,p,dn,pn,y,_n;return a()().wrap(function(R){for(;;)switch(R.prev=R.next){case 0:if(!(I.current||rn||!sn||!fn)){R.next=2;break}return R.abrupt("return");case 2:return Y=s.current,I.current=!0,ln(!0),R.prev=5,R.next=8,(0,b.Sq)({body:{page_size:100,order_by_field:"id",order_by_type:"desc",cursor_id:fn,keyword:V}});case 8:if(y=R.sent,s.current===Y){R.next=11;break}return R.abrupt("return");case 11:if(y){R.next=13;break}return R.abrupt("return");case 13:if((y==null?void 0:y.code)===200){R.next=16;break}return e.error((y==null?void 0:y.message)||"\u52A0\u8F7D\u66F4\u591A\u7EC8\u7AEF\u5931\u8D25"),R.abrupt("return");case 16:_n=Array.isArray(y==null||(F=y.data)===null||F===void 0?void 0:F.list)?y==null||(p=y.data)===null||p===void 0?void 0:p.list.filter(Boolean).map(x):[],Q(function(In){var Un=new Set(In.map(function(On){return On.id}));return[].concat(g()(In),g()(_n.filter(function(On){return!Un.has(On.id)})))}),gn(!!(!(y==null||(dn=y.data)===null||dn===void 0)&&dn.has_more)),j(String((y==null||(pn=y.data)===null||pn===void 0?void 0:pn.next_cursor_id)||""));case 20:return R.prev=20,s.current===Y&&(I.current=!1,ln(!1)),R.finish(20);case 23:case"end":return R.stop()}},T,null,[[5,,20,23]])})),[sn,rn,e,fn,V]),Ln=(0,i.useMemo)(function(){var T=V.toLocaleLowerCase();return T&&!h.searchText.includes(T)?z:[h].concat(g()(z))},[h,V,z]);return{servers:Ln,keyword:J,loading:rn,loadingMore:An,loaded:Mn,hasMore:sn,localLoaded:n,localRefreshing:un,localError:v,localServer:h,setKeyword:N,loadMore:Bn,reload:Cn,update:Rn}};A.A=f}),40062:(function(d,A,t){"use strict";var k=t(84169),a=t.n(k),c=t(1974),g=t.n(c),_=t(12185),w=t(30734),E,m,i,W,b,x,C,f,B,o,e,s,l,u,K,z,Q,nn=(0,_.rU)(function(G){var h=G.css,r=G.token,tn=G.isDarkMode,M=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{},n=!!(M!=null&&M.isCompactMode),L=typeof(M==null?void 0:M.isDarkMode)=="boolean"?M.isDarkMode:!!tn,Z=(M==null?void 0:M.terminalBackground)||(L?"rgb(15, 15, 15)":"#000000"),an=n?238:252,un=n?270:286,P=n?42:46,X=n?46:50,H=n?42:46,v=n?46:50;return g()(g()({page:h(E||(E=a()([`
            --terminal-page-height: calc(100dvh - 92px);
            width: 100%;
            max-width: 1450px;
            min-height: var(--terminal-page-height);
            margin: 0 auto;

            @supports not (height: 100dvh) {
                --terminal-page-height: calc(100vh - 92px);
            }

            @media (max-width: 780px) {
                --terminal-page-height: calc(100dvh - 20px);
                min-height: var(--terminal-page-height);

                @supports not (height: 100dvh) {
                    --terminal-page-height: calc(100vh - 20px);
                }
            }
        `]))),workspace:h(m||(m=a()([`
            container-name: terminal-workspace;
            container-type: inline-size;
            width: 100%;
            height: var(--terminal-page-height);
            min-width: 0;
            min-height: 320px;
            overflow: hidden;
            border-radius: `,`px;
            background: `,`;

            > .ant-card-body {
                width: 100%;
                height: 100%;
                min-height: 0;
                padding: 0;
            }

            @media (max-width: 780px) {
                height: auto;
                min-height: var(--terminal-page-height);
            }

            @media (min-width: 781px) and (max-height: 600px) and (orientation: landscape) {
                height: auto;
                min-height: 300px;
            }
        `])),r.borderRadiusLG,r.colorBgContainer),terminalLayout:h(i||(i=a()([`
            width: 100%;
            height: 100%;
            min-width: 0;
            min-height: 0;
            display: grid;
            grid-template-columns: `,"px minmax(420px, 1fr) ",`px;
            grid-template-rows: minmax(0, 1fr);

            &.server-collapsed {
                grid-template-columns: `,"px minmax(420px, 1fr) ",`px;
            }

            &.scripts-collapsed {
                grid-template-columns: `,"px minmax(420px, 1fr) ",`px;
            }

            &.server-collapsed.scripts-collapsed {
                grid-template-columns: `,"px minmax(420px, 1fr) ",`px;
            }

            @container terminal-workspace (max-width: 980px) {
                grid-template-columns: `,"px minmax(300px, 1fr) ",`px;
                grid-template-rows: minmax(0, 1fr);

                &.scripts-collapsed {
                    grid-template-columns: `,"px minmax(300px, 1fr) ",`px;
                }

                &.server-collapsed {
                    grid-template-columns: `,"px minmax(300px, 1fr) ",`px;
                }

                &.server-collapsed.scripts-collapsed {
                    grid-template-columns: `,"px minmax(300px, 1fr) ",`px;
                }
            }

            @container terminal-workspace (max-width: 760px) {
                height: auto;
                min-height: var(--terminal-page-height);
                grid-template-columns: 1fr;
                grid-template-rows: auto auto minmax(0, 1fr);

                &.server-collapsed,
                &.scripts-collapsed {
                    grid-template-columns: 1fr;
                }

                &.server-collapsed.scripts-collapsed {
                    grid-template-columns: 1fr;
                }
            }

            @media (min-width: 781px) and (max-height: 600px) and (orientation: landscape) {
                height: auto;
                min-height: 300px;
                grid-template-rows: minmax(300px, 1fr);
            }
        `])),an,un,P,un,an,P,P,P,n?196:208,n?220:232,n?196:208,P,P,n?220:232,P,P),serverPane:h(W||(W=a()([`
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            border-right: 1px solid `,`;
            background: `,`;

            &.collapsed {
                overflow: hidden;
            }

            @container terminal-workspace (max-width: 760px) {
                grid-column: 1;
                grid-row: 1;
                border-right: 0;
                border-bottom: 1px solid `,`;

                /* \u6A2A\u5411\u7EC8\u7AEF\u5217\u8868\u5DF2\u7ECF\u5728\u5217\u8868\u9996\u9879\u63D0\u4F9B\u6DFB\u52A0\u5165\u53E3\uFF0C\u9690\u85CF\u5E95\u90E8\u91CD\u590D\u6309\u94AE\u3002 */
                .serverFooter {
                    display: none !important;
                }

                &.collapsed {
                    overflow: visible;
                }
            }
        `])),r.colorSplit,r.colorBgContainer,r.colorSplit),serverListArea:h(b||(b=a()([`
            position: relative;
            min-width: 0;
            min-height: 0;
            display: flex;
            flex: 1;
            flex-direction: column;

            @container terminal-workspace (max-width: 760px) {
                flex: none;
            }
        `]))),serverLoadingOverlay:h(x||(x=a()([`
            position: absolute;
            z-index: 8;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            pointer-events: none;

            .ant-spin {
                line-height: 1;
            }
        `]))),serverHeader:h(C||(C=a()([`
            height: `,`px;
            min-height: `,`px;
            padding: 0 `,`px;
            display: flex;
            flex: 0 0 `,`px;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
            box-sizing: border-box;
            border-bottom: 1px solid `,`;

            .panel-heading {
                min-width: 0;
                display: flex;
                align-items: center;
                color: `,`;
                font-size: `,`px;
                font-weight: 600;
            }

            .panel-heading > div {
                min-width: 0;
            }

            .ant-btn {
                width: `,`px;
                height: `,`px;
                padding: 0;
                flex: none;
            }

            .collapsed & {
                padding: 0;
                justify-content: center;
            }

            .collapsed & .panel-heading {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                height: `,`px;
                min-height: `,`px;
                flex-basis: `,`px;
                padding-inline: `,`px;

                .collapse-trigger {
                    display: none;
                }

                .collapsed & {
                    padding-inline: `,`px;
                    justify-content: space-between;
                }

                .collapsed & .panel-heading {
                    display: flex;
                }
            }
        `])),X,X,n?11:13,X,r.colorSplit,r.colorTextHeading,n?13:14,n?30:32,n?30:32,H,H,H,n?9:11,n?9:11),serverToolbar:h(f||(f=a()([`
            min-width: 0;
            height: `,`px;
            min-height: `,`px;
            padding: 8px;
            display: grid;
            flex: 0 0 `,`px;
            grid-template-columns: minmax(0, 1fr) `,`px;
            align-items: center;
            gap: 6px;
            box-sizing: border-box;

            .collapsed & {
                display: none;
            }

            .ant-input-affix-wrapper {
                height: `,`px;
                padding-inline: `,`px;
                border-radius: `,`px;
            }

            .ant-input-prefix {
                color: `,`;
                font-size: 12px;
            }

            .ant-input {
                font-size: `,`px;
            }

            .ant-input::placeholder {
                color: `,`;
                font-size: `,`px;
            }

            .ant-btn {
                width: `,`px;
                height: `,`px;
                padding: 0;
            }

            .ant-btn .anticon {
                font-size: `,`px;
            }

            .ant-btn:not(:disabled) .anticon {
                color: `,`;
            }

            @container terminal-workspace (max-width: 760px) {
                min-height: `,`px;
                padding: 8px;
                grid-template-columns: minmax(0, 1fr) `,`px;

                .ant-btn {
                    width: `,`px;
                    height: `,`px;
                }

                .collapsed & {
                    display: grid;
                }
            }
        `])),v,v,v,n?30:34,n?30:34,n?9:10,r.borderRadiusSM,r.colorTextTertiary,n?11:12,r.colorTextQuaternary,n?10:11,n?30:34,n?30:34,n?12:13,r.colorTextTertiary,n?46:50,n?32:34,n?32:34,n?32:34),serverList:h(B||(B=a()([`
            min-height: 0;
            padding: 0 `,`px;
            display: flex;
            flex: 1;
            flex-direction: column;
            gap: 3px;
            overflow-x: hidden;
            overflow-y: auto;
            scrollbar-width: thin;
            scrollbar-color: `,` transparent;

            .collapsed & {
                padding: `,`;
                align-items: center;
                gap: 4px;
                overflow-x: hidden;
                overflow-y: auto;
                scrollbar-width: none;
            }

            &::-webkit-scrollbar {
                width: 6px !important;
            }

            &::-webkit-scrollbar-thumb {
                border-radius: `,`px;
                background: `,`;
            }

            .collapsed &::-webkit-scrollbar {
                display: none;
            }

            .server-loading,
            .server-initial-loading,
            .server-empty {
                min-height: 110px;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .server-load-more {
                min-height: 32px;
                display: flex;
                flex: 0 0 auto;
                align-items: center;
                justify-content: center;
            }

            .server-empty {
                flex: 1;
            }

            .server-empty .ant-empty {
                margin-block: 12px;
            }

            .server-empty .ant-empty-description {
                color: `,`;
                font-size: 11px;
            }

            @container terminal-workspace (max-width: 760px) {
                min-width: 0;
                padding: 0 `,"px ",`px 0;
                flex: none;
                flex-direction: row;
                gap: 6px;
                overflow-x: auto;
                overflow-y: hidden;
                scroll-snap-type: x proximity;
                scroll-padding-left: `,`px;
                overscroll-behavior-x: contain;
                scrollbar-width: none;
                touch-action: pan-x pan-y;
                -webkit-overflow-scrolling: touch;

                .collapsed & {
                    display: flex;
                    padding: 0 `,"px ",`px 0;
                    align-items: stretch;
                    gap: 6px;
                    overflow-x: auto;
                    overflow-y: hidden;
                }

                &::-webkit-scrollbar {
                    display: none;
                }

                .server-empty,
                .server-loading,
                .server-load-more {
                    min-width: 144px;
                    min-height: `,`px;
                    flex: 0 0 144px;
                }

                .server-initial-loading {
                    width: 100%;
                    min-width: 100%;
                    min-height: `,`px;
                    flex: 0 0 100%;
                }

                .server-empty .ant-empty-image {
                    display: none;
                }
            }
        `])),n?7:9,r.colorBorderSecondary,n?"6px 5px 4px":"7px 6px 5px",r.borderRadiusSM,r.colorBorderSecondary,r.colorTextTertiary,n?9:11,n?9:11,n?12:14,n?9:11,n?9:11,n?48:52,n?48:52),serverItem:h(o||(o=a()([`
            position: relative;
            min-width: 0;
            min-height: `,`px;
            overflow: hidden;
            border: 0;
            border-radius: `,`px;
            background: transparent;
            color: `,`;
            content-visibility: auto;
            contain-intrinsic-size: auto `,`px;
            transition: background-color .16s ease;

            &:hover {
                background: `,`;
            }

            &.selected {
                background: `,`;
            }

            .connection-dot {
                position: absolute;
                z-index: 2;
                top: 50%;
                left: `,`px;
                width: 6px;
                height: 6px;
                border-radius: 50%;
                background: `,`;
                transform: translateY(-50%);
            }

            .connection-dot.connected {
                background: `,`;
                box-shadow: 0 0 0 3px `,`;
            }

            .connection-dot.connecting {
                background: `,`;
                box-shadow: 0 0 0 3px `,`;
            }

            .connection-dot.error {
                background: `,`;
                box-shadow: 0 0 0 3px `,`;
            }

            .server-icon {
                display: none;
            }

            .server-select {
                width: 100%;
                min-height: `,`px;
                padding: `,`;
                display: flex;
                align-items: center;
                border: 0;
                border-radius: inherit;
                outline: none;
                background: transparent;
                color: inherit;
                cursor: pointer;
                font: inherit;
                text-align: left;
            }

            .server-select:focus-visible {
                box-shadow: inset 0 0 0 1px `,`;
            }

            .server-select:disabled {
                cursor: default;
                opacity: .72;
            }

            .server-copy {
                min-width: 0;
                display: block;
                flex: 1;
            }

            .server-name,
            .server-endpoint {
                display: block;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .server-name {
                color: `,`;
                font-size: `,`px;
                font-weight: 500;
                line-height: `,`px;
            }

            .server-endpoint {
                color: `,`;
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
                font-size: 11px;
                line-height: `,`px;
            }

            .server-actions {
                position: absolute;
                z-index: 3;
                top: 50%;
                right: `,`px;
                display: flex;
                align-items: center;
                gap: 1px;
                opacity: 0;
                transform: translateY(-50%);
                transition: opacity .16s ease;
            }

            .server-action {
                width: `,`px;
                height: `,`px;
                padding: 0;
            }

            &:hover .server-actions,
            &.selected .server-actions,
            .server-actions:focus-within {
                opacity: 1;
            }

            .collapsed & {
                width: `,`px;
                min-height: `,`px;
                flex: 0 0 `,`px;
                overflow: visible;
                contain-intrinsic-size: auto `,`px;
            }

            .collapsed & .server-select {
                width: `,`px;
                min-height: `,`px;
                padding: 0;
                justify-content: center;
            }

            .collapsed & .server-icon {
                display: inline-flex;
                align-items: center;
                justify-content: center;
                color: `,`;
                font-size: `,`px;
            }

            .collapsed &.selected .server-icon {
                color: `,`;
            }

            .collapsed & .connection-dot {
                top: auto;
                right: 3px;
                bottom: 3px;
                left: auto;
                width: 8px;
                height: 8px;
                border: 1px solid `,`;
                box-shadow: none;
                transform: none;
            }

            .collapsed & .server-copy,
            .collapsed & .server-actions {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                min-height: `,`px;
                flex: 0 0 `,`px;
                overflow: hidden;
                background: `,`;
                contain-intrinsic-size: auto `,`px;
                scroll-snap-align: start;

                .server-select {
                    min-height: `,`px;
                    padding-right: `,`px;
                }

                .server-actions {
                    opacity: 1;
                }

                .server-action {
                    width: `,`px;
                    height: `,`px;
                }

                .collapsed & {
                    width: auto;
                    min-height: `,`px;
                    flex: 0 0 `,`px;
                }

                .collapsed & .server-select {
                    width: 100%;
                    min-height: `,`px;
                    padding: `,`;
                    justify-content: flex-start;
                }

                .collapsed & .server-icon {
                    display: none;
                }

                .collapsed & .connection-dot {
                    top: 50%;
                    right: auto;
                    bottom: auto;
                    left: `,`px;
                    width: 6px;
                    height: 6px;
                    border: 0;
                    transform: translateY(-50%);
                }

                .collapsed & .connection-dot.connected {
                    box-shadow: 0 0 0 3px `,`;
                }

                .collapsed & .connection-dot.connecting {
                    box-shadow: 0 0 0 3px `,`;
                }

                .collapsed & .connection-dot.error {
                    box-shadow: 0 0 0 3px `,`;
                }

                .collapsed & .server-copy {
                    display: block;
                }

                .collapsed & .server-actions {
                    display: flex;
                    opacity: 1;
                }
            }

            @container terminal-workspace (max-width: 430px) {
                flex-basis: `,`px;

                .collapsed & {
                    flex-basis: `,`px;
                }
            }
        `])),n?50:54,r.borderRadius,r.colorText,n?50:54,r.colorFillQuaternary,r.colorPrimaryBg,n?10:11,r.colorTextQuaternary,r.colorSuccess,r.colorSuccessBg,r.colorWarning,r.colorWarningBg,r.colorError,r.colorErrorBg,n?50:54,n?"5px 38px 5px 26px":"6px 41px 6px 27px",r.colorPrimaryBorder,r.colorTextHeading,n?12:13,n?18:19,r.colorTextTertiary,n?15:16,n?5:6,n?28:30,n?28:30,n?30:32,n?30:32,n?30:32,n?30:32,n?30:32,n?30:32,r.colorTextTertiary,n?14:15,r.colorPrimary,r.colorBgContainer,n?48:52,n?160:168,r.colorFillQuaternary,n?48:52,n?48:52,n?43:46,n?32:34,n?32:34,n?48:52,n?160:168,n?48:52,n?"5px 43px 5px 26px":"6px 46px 6px 27px",n?10:11,r.colorSuccessBg,r.colorWarningBg,r.colorErrorBg,n?152:160,n?152:160),serverFooter:h(e||(e=a()([`
            flex: none;
            padding: `,`;

            .collapsed & {
                padding: `,`;
                display: block;
            }

            .collapsed & > button {
                min-height: `,`px;
                padding: 0;
                border: 0;
            }

            .collapsed & > button .add-label {
                display: none;
            }

            .collapsed &.script-footer {
                margin-top: auto;
            }

            @container terminal-workspace (max-width: 760px) {
                display: none;
            }
        `])),n?"8px 9px 10px":"9px 10px 11px",n?"5px 5px 9px":"6px 6px 10px",n?30:32),serverAdd:h(s||(s=a()([`
            width: 100%;
            min-height: `,`px;
            padding: 0 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 7px;
            border: 1px dashed `,`;
            border-radius: `,`px;
            outline: none;
            appearance: none;
            background: transparent;
            color: `,`;
            cursor: pointer;
            font: inherit;
            font-size: `,`px;
            transition: border-color .16s ease, background-color .16s ease, color .16s ease;

            &.mobile {
                display: none;
            }

            &:hover,
            &:focus-visible {
                border-color: `,`;
                background: `,`;
                color: `,`;
            }

            &:focus-visible {
                box-shadow: 0 0 0 2px `,`;
            }

            @container terminal-workspace (max-width: 760px) {
                min-height: `,`px;

                &.mobile {
                    width: auto;
                    min-width: `,`px;
                    margin-left: `,`px;
                    padding: 0 11px;
                    display: flex;
                    flex: 0 0 auto;
                    white-space: nowrap;
                    scroll-snap-align: start;
                }
            }
        `])),n?38:42,r.colorBorder,r.borderRadius,r.colorTextTertiary,n?12:13,r.colorPrimaryBorder,r.colorFillQuaternary,r.colorPrimary,r.colorPrimaryBg,n?50:54,n?122:130,n?12:14)},(0,w.x)({css:h,token:r},{compact:n,terminalBackground:Z})),{},{scriptPane:h(l||(l=a()([`
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            border-left: 1px solid `,`;
            background: `,`;

            &.collapsed {
                align-items: stretch;
            }

            @container terminal-workspace (max-width: 760px) {
                grid-column: 1;
                grid-row: 2;
                max-height: none;
                border-top: 0;
                border-bottom: 1px solid `,`;
                border-left: 0;

                /* \u6A2A\u5411\u811A\u672C\u5217\u8868\u5DF2\u7ECF\u5728\u5217\u8868\u9996\u9879\u63D0\u4F9B\u6DFB\u52A0\u5165\u53E3\uFF0C\u9690\u85CF\u5E95\u90E8\u91CD\u590D\u6309\u94AE\u3002 */
                .script-footer {
                    display: none !important;
                }

                &.collapsed {
                    max-height: none;
                    flex-direction: column;
                    align-items: stretch;
                }
            }

            @media (max-width: 760px) and (max-height: 600px) and (orientation: landscape) {
                max-height: none;

                &.collapsed {
                    max-height: none;
                }
            }
        `])),r.colorSplit,r.colorBgContainer,r.colorSplit),scriptHeader:h(u||(u=a()([`
            height: `,`px;
            min-height: `,`px;
            padding: 0 `,`px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
            flex: 0 0 `,`px;
            box-sizing: border-box;
            border-bottom: 1px solid `,`;

            .panel-heading {
                min-width: 0;
                display: flex;
                align-items: center;
                color: `,`;
                font-size: `,`px;
                font-weight: 600;
            }

            .panel-heading > div {
                min-width: 0;
            }

            .ant-btn {
                width: `,`px;
                height: `,`px;
                padding: 0;
                flex: none;
            }

            .collapsed & {
                padding: 0;
                justify-content: center;
            }

            .collapsed & .panel-heading {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                height: `,`px;
                min-height: `,`px;
                flex-basis: `,`px;
                padding-inline: `,`px;

                .collapse-trigger {
                    display: none;
                }

                .collapsed & {
                    width: auto;
                    min-width: 0;
                    padding-inline: `,`px;
                    flex: none;
                    justify-content: space-between;
                    border-right: 0;
                    border-bottom: 1px solid `,`;
                }

                .collapsed & .panel-heading {
                    display: flex;
                }
            }
        `])),X,X,n?11:13,X,r.colorSplit,r.colorTextHeading,n?13:14,n?30:32,n?30:32,H,H,H,n?9:11,n?9:11,r.colorSplit),scriptBody:h(K||(K=a()([`
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            flex: 1;
            overflow: hidden;

            .collapsed & {
                display: none;
            }

            .script-toolbar {
                min-width: 0;
                height: `,`px;
                min-height: `,`px;
                padding: 8px;
                display: grid;
                grid-template-columns: minmax(0, 1fr) auto;
                align-items: center;
                gap: 6px;
                flex: 0 0 `,`px;
                box-sizing: border-box;
                overflow-x: auto;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
            }

            .script-toolbar::-webkit-scrollbar {
                display: none;
            }

            .script-toolbar .ant-input-affix-wrapper {
                min-width: 0;
                height: `,`px;
                padding-inline: `,`px;
                border-radius: `,`px;
            }

            .script-toolbar .ant-input-prefix {
                color: `,`;
                font-size: 12px;
            }

            .script-toolbar .ant-input {
                font-size: `,`px;
            }

            .script-toolbar .ant-input::placeholder {
                color: `,`;
                font-size: `,`px;
            }

            .script-type-trigger {
                width: auto;
                min-width: max-content;
                height: `,`px;
                padding: 0 `,`px;
                gap: 5px;
                line-height: 1;
            }

            .script-type-trigger .marker {
                color: `,`;
                font-size: `,`px;
            }

            .script-type-trigger .value {
                flex: none;
                color: `,`;
                font-size: `,`px;
                font-weight: 400;
                white-space: nowrap;
                transition: color .16s ease;
            }

            .script-type-trigger .arrow {
                color: `,`;
                font-size: `,`px;
                transition: color .16s ease, transform .16s ease;
            }

            .script-type-trigger.ant-dropdown-open .arrow {
                color: `,`;
                transform: rotate(180deg);
            }

            .script-scroll {
                min-width: 0;
                min-height: 0;
                display: flex;
                flex: 1;
                flex-direction: column;
                overflow-x: hidden;
                overflow-y: auto;
                overscroll-behavior: contain;
                scrollbar-width: thin;
                scrollbar-color: `,` transparent;
            }

            .script-scroll::-webkit-scrollbar {
                width: 6px !important;
            }

            .script-scroll::-webkit-scrollbar-thumb {
                border-radius: `,`px;
                background: `,`;
            }

            .script-list {
                min-width: 0;
                padding: 0 `,`px;
                display: flex;
                flex: none;
                flex-direction: column;
                gap: 3px;
            }

            .script-item {
                min-width: 0;
                min-height: `,`px;
                padding: `,`;
                display: grid;
                grid-template-columns: minmax(0, 1fr) `,"px ",`px;
                align-items: center;
                gap: 2px;
                position: relative;
                border-radius: `,`px;
                background: transparent;
                content-visibility: auto;
                contain-intrinsic-size: auto `,`px;
                transition: background-color .16s ease;
            }

            .script-item:hover,
            .script-item:focus-within {
                background: `,`;
            }

            .script-item.copying {
                background: `,`;
            }

            .script-copying {
                position: absolute;
                inset: 0;
                z-index: 2;
                display: flex;
                align-items: center;
                justify-content: center;
                border-radius: inherit;
                cursor: wait;
            }

            .script-copy {
                min-width: 0;
                height: 100%;
                padding: 0;
                display: grid;
                grid-template-columns: minmax(0, 1fr) auto;
                align-items: center;
                gap: 5px;
                border: 0;
                outline: none;
                background: transparent;
                cursor: pointer;
                font: inherit;
                text-align: left;
            }

            .script-copy:focus-visible {
                border-radius: `,`px;
                box-shadow: inset 0 0 0 1px `,`;
            }

            .script-copy-content {
                min-width: 0;
                display: flex;
                flex-direction: column;
                align-items: flex-start;
            }

            .script-name {
                max-width: 100%;
                overflow: hidden;
                color: `,`;
                font-size: `,`px;
                font-weight: 500;
                line-height: `,`px;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .script-type {
                max-width: 100%;
                overflow: hidden;
                color: `,`;
                font-size: 11px;
                line-height: `,`px;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .script-terminal,
            .script-action {
                width: `,`px !important;
                height: `,`px !important;
                padding: 0;
                opacity: 0;
                transition: opacity .16s ease;
            }

            .script-item:hover .script-terminal,
            .script-item:focus-within .script-terminal,
            .script-item:hover .script-action,
            .script-item:focus-within .script-action,
            .script-terminal.ant-btn-loading,
            .script-action.ant-btn-loading,
            .script-action.ant-dropdown-open {
                opacity: 1;
            }

            .script-state {
                min-height: 110px;
                padding: 16px 10px;
                display: flex;
                flex: 1;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                gap: 6px;
                color: `,`;
                font-size: 12px;
                text-align: center;
            }

            .script-state .ant-empty {
                margin-block: 0;
            }

            .script-state .ant-empty-description {
                color: `,`;
                font-size: 11px;
            }

            .script-load-more {
                min-height: 32px;
                display: flex;
                align-items: center;
                justify-content: center;
            }

            .script-load-more .ant-btn {
                width: 32px;
                height: 28px;
                padding: 0;
            }

            @media (pointer: coarse) {
                .script-item {
                    min-height: `,`px;
                    grid-template-columns: minmax(0, 1fr) 36px 36px;
                }

                .script-terminal,
                .script-action {
                    width: 36px !important;
                    height: 36px !important;
                    opacity: 1;
                }
            }

            @container terminal-workspace (max-width: 760px) {
                flex: none;

                .collapsed & {
                    display: flex;
                }

                .script-scroll {
                    min-width: 0;
                    padding: 0 `,"px ",`px 0;
                    display: flex;
                    flex: none;
                    flex-direction: row;
                    align-items: stretch;
                    gap: 6px;
                    overflow-x: auto;
                    overflow-y: hidden;
                    overscroll-behavior-x: contain;
                    scroll-padding-left: `,`px;
                    scroll-snap-type: x proximity;
                    scrollbar-width: none;
                    touch-action: pan-x pan-y;
                    -webkit-overflow-scrolling: touch;
                }

                .script-scroll::-webkit-scrollbar {
                    display: none;
                }

                .script-list {
                    width: max-content;
                    padding: 0;
                    display: flex;
                    flex: 0 0 auto;
                    flex-direction: row;
                    gap: 6px;
                }

                .script-item {
                    min-height: `,`px;
                    flex: 0 0 `,`px;
                    grid-template-columns: minmax(0, 1fr) 36px 36px;
                    overflow: hidden;
                    background: `,`;
                    contain-intrinsic-size: auto `,`px;
                    scroll-snap-align: start;
                }

                .script-terminal,
                .script-action {
                    width: 36px !important;
                    height: 36px !important;
                    opacity: 1;
                }

                .script-state {
                    min-width: 144px;
                    min-height: `,`px;
                    padding: 0 10px;
                    flex: 0 0 144px;
                    scroll-snap-align: start;
                }

                .script-state .ant-empty-image {
                    display: none;
                }

                .script-load-more {
                    min-width: 52px;
                    min-height: `,`px;
                    flex: 0 0 52px;
                    scroll-snap-align: start;
                }
            }
        `])),v,v,v,n?30:34,n?9:10,r.borderRadiusSM,r.colorTextTertiary,n?11:12,r.colorTextQuaternary,n?10:11,n?30:34,n?7:10,r.colorTextQuaternary,n?11:12,r.colorTextSecondary,n?11:12,r.colorTextQuaternary,n?9:10,r.colorTextTertiary,r.colorBorderSecondary,r.borderRadiusSM,r.colorBorderSecondary,n?7:9,n?50:54,n?"4px 2px 4px 8px":"5px 3px 5px 9px",n?28:30,n?28:30,r.borderRadius,n?50:54,r.colorFillQuaternary,r.colorFillQuaternary,r.borderRadius,r.colorPrimaryBorder,r.colorTextHeading,n?12:13,n?18:19,r.colorTextTertiary,n?15:16,n?28:30,n?28:30,r.colorTextTertiary,r.colorTextTertiary,n?48:52,n?9:11,n?9:11,n?12:14,n?48:52,n?196:208,r.colorFillQuaternary,n?48:52,n?48:52,n?48:52),scriptRail:h(z||(z=a()([`
            min-width: 0;
            min-height: 0;
            padding: `,`;
            display: none;
            flex: 1;
            flex-direction: column;
            align-items: center;
            gap: 4px;
            overflow-x: hidden;
            overflow-y: auto;
            overscroll-behavior: contain;
            scrollbar-width: none;

            .collapsed & {
                display: flex;
            }

            &::-webkit-scrollbar {
                display: none;
            }

            .rail-action {
                width: `,`px;
                height: `,`px;
                padding: 0;
                flex: none;
                font-size: `,`px;
            }

            .rail-load-more {
                margin-top: 2px;
                font-size: `,`px;
            }

            .rail-loading {
                padding-block: 7px;
                flex: none;
                color: `,`;
                font-size: `,`px;
            }

            @container terminal-workspace (max-width: 760px) {
                .collapsed & {
                    display: none;
                }
            }
        `])),n?"6px":"7px",n?30:32,n?30:32,n?14:15,n?10:11,r.colorPrimary,n?12:13),scriptDropdown:h(Q||(Q=a()([`
            && {
                max-width: calc(100vw - 24px);
                box-sizing: border-box;
                padding: 0;
                border-radius: `,`px;
            }

            && .ant-dropdown-menu {
                width: 100%;
                min-width: 0 !important;
                max-height: min(320px, calc(100dvh - 24px));
                padding: 2px !important;
                overflow-y: auto;
                overscroll-behavior: contain;
                border-radius: `,`px !important;
                background: `,`;
                box-shadow: `,`;
            }

            && .ant-dropdown-menu-item,
            && .ant-dropdown-menu-title-content {
                min-width: 0;
            }

            && .ant-dropdown-menu-item {
                min-height: `,`px !important;
                margin: 0 !important;
                padding: 0 6px !important;
                border-radius: `,`px !important;
                color: `,`;
                font-size: `,`px !important;
                font-weight: 400;
                line-height: `,`px !important;
            }

            && .ant-dropdown-menu-title-content {
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        `])),r.borderRadius,r.borderRadius,r.colorBgElevated,r.boxShadowSecondary,n?24:30,r.borderRadiusSM,r.colorTextSecondary,n?11:12,n?24:30)})});A.A=nn}),86441:(function(d,A,t){"use strict";t.d(A,{Sq:function(){return w},dZ:function(){return x},sA:function(){return W},tg:function(){return f},y8:function(){return m}});var k=t(55420),a=t.n(k),c=t(98312),g=t.n(c),_=t(78781);function w(){return E.apply(this,arguments)}function E(){return E=g()(a()().mark(function o(){var e,s=arguments;return a()().wrap(function(u){for(;;)switch(u.prev=u.next){case 0:return e=s.length>0&&s[0]!==void 0?s[0]:{},u.abrupt("return",(0,_.Jt)("/server/list",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return u.stop()}},o)})),E.apply(this,arguments)}function m(){return i.apply(this,arguments)}function i(){return i=g()(a()().mark(function o(){var e,s=arguments;return a()().wrap(function(u){for(;;)switch(u.prev=u.next){case 0:return e=s.length>0&&s[0]!==void 0?s[0]:{},u.abrupt("return",(0,_.Jt)("/server/info",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return u.stop()}},o)})),i.apply(this,arguments)}function W(){return b.apply(this,arguments)}function b(){return b=g()(a()().mark(function o(){var e,s=arguments;return a()().wrap(function(u){for(;;)switch(u.prev=u.next){case 0:return e=s.length>0&&s[0]!==void 0?s[0]:{},u.abrupt("return",(0,_.bE)("/server/create",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return u.stop()}},o)})),b.apply(this,arguments)}function x(){return C.apply(this,arguments)}function C(){return C=g()(a()().mark(function o(){var e,s=arguments;return a()().wrap(function(u){for(;;)switch(u.prev=u.next){case 0:return e=s.length>0&&s[0]!==void 0?s[0]:{},u.abrupt("return",(0,_.bE)("/server/update",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return u.stop()}},o)})),C.apply(this,arguments)}function f(){return B.apply(this,arguments)}function B(){return B=g()(a()().mark(function o(){var e,s=arguments;return a()().wrap(function(u){for(;;)switch(u.prev=u.next){case 0:return e=s.length>0&&s[0]!==void 0?s[0]:{},u.abrupt("return",(0,_.bE)("/server/delete",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return u.stop()}},o)})),B.apply(this,arguments)}}),41162:(function(d,A,t){var k=t(44814);function a(c){if(Array.isArray(c))return k(c)}d.exports=a,d.exports.__esModule=!0,d.exports.default=d.exports}),59224:(function(d,A,t){var k=t(34911);function a(c,g){var _=typeof Symbol!="undefined"&&c[Symbol.iterator]||c["@@iterator"];if(!_){if(Array.isArray(c)||(_=k(c))||g&&c&&typeof c.length=="number"){_&&(c=_);var w=0,E=function(){};return{s:E,n:function(){return w>=c.length?{done:!0}:{done:!1,value:c[w++]}},e:function(x){throw x},f:E}}throw new TypeError(`Invalid attempt to iterate non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}var m=!0,i=!1,W;return{s:function(){_=_.call(c)},n:function(){var x=_.next();return m=x.done,x},e:function(x){i=!0,W=x},f:function(){try{!m&&_.return!=null&&_.return()}finally{if(i)throw W}}}}d.exports=a,d.exports.__esModule=!0,d.exports.default=d.exports}),33032:(function(d){function A(t){if(typeof Symbol!="undefined"&&t[Symbol.iterator]!=null||t["@@iterator"]!=null)return Array.from(t)}d.exports=A,d.exports.__esModule=!0,d.exports.default=d.exports}),30922:(function(d){function A(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}d.exports=A,d.exports.__esModule=!0,d.exports.default=d.exports}),71640:(function(d,A,t){var k=t(59276);function a(c,g){if(c==null)return{};var _=k(c,g),w,E;if(Object.getOwnPropertySymbols){var m=Object.getOwnPropertySymbols(c);for(E=0;E<m.length;E++)w=m[E],!(g.indexOf(w)>=0)&&Object.prototype.propertyIsEnumerable.call(c,w)&&(_[w]=c[w])}return _}d.exports=a,d.exports.__esModule=!0,d.exports.default=d.exports}),59276:(function(d){function A(t,k){if(t==null)return{};var a={},c=Object.keys(t),g,_;for(_=0;_<c.length;_++)g=c[_],!(k.indexOf(g)>=0)&&(a[g]=t[g]);return a}d.exports=A,d.exports.__esModule=!0,d.exports.default=d.exports}),45647:(function(d,A,t){var k=t(41162),a=t(33032),c=t(34911),g=t(30922);function _(w){return k(w)||a(w)||c(w)||g()}d.exports=_,d.exports.__esModule=!0,d.exports.default=d.exports})}]);
