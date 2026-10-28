!(function(){var st=Object.defineProperty,lt=Object.defineProperties;var dt=Object.getOwnPropertyDescriptors;var me=Object.getOwnPropertySymbols;var Ue=Object.prototype.hasOwnProperty,Ke=Object.prototype.propertyIsEnumerable;var $e=(c,g,e)=>g in c?st(c,g,{enumerable:!0,configurable:!0,writable:!0,value:e}):c[g]=e,On=(c,g)=>{for(var e in g||(g={}))Ue.call(g,e)&&$e(c,e,g[e]);if(me)for(var e of me(g))Ke.call(g,e)&&$e(c,e,g[e]);return c},Sn=(c,g)=>lt(c,dt(g));var De=(c,g)=>{var e={};for(var m in c)Ue.call(c,m)&&g.indexOf(m)<0&&(e[m]=c[m]);if(c!=null&&me)for(var m of me(c))g.indexOf(m)<0&&Ke.call(c,m)&&(e[m]=c[m]);return e};(self.webpackChunk=self.webpackChunk||[]).push([[93],{27115:(function(c,g,e){"use strict";var m=e(1974),W=e.n(m),T=e(96540),D=e(79688),B=e(74848),_=(0,T.forwardRef)(function(X,Q){return(0,B.jsx)(D.A,W()(W()({},X),{},{ref:Q}))});_.displayName="StableDropdown",g.A=_}),61225:(function(c,g,e){"use strict";e.r(g),e.d(g,{default:function(){return Le}});var m=e(55420),W=e.n(m),T=e(45647),D=e.n(T),B=e(1974),_=e.n(B),X=e(98312),Q=e.n(X),an=e(11604),K=e.n(an),cn=e(84169),u=e.n(cn),o=e(96540),z=e(47152),A=e(16370),$=e(12185),k=e(62885),ln=e(72523),Z=12,d=function(a){var t=Number(a||0);return Number.isFinite(t)?t:0},sn=function(a){return Math.min(100,Math.max(0,d(a)))},un=function(a){return"".concat(Number(sn(a).toFixed(1)),"%")},xn=new Intl.NumberFormat("zh-CN",{maximumFractionDigits:1}),mn=function(a){return xn.format(d(a))},fn=function(a){var t=Math.max(0,Math.floor(d(a))),r=Math.floor(t/86400),s=Math.floor(t%86400/3600),L=Math.floor(t%3600/60);return r>0?"".concat(r," \u5929 ").concat(s," \u5C0F\u65F6"):s>0?"".concat(s," \u5C0F\u65F6 ").concat(L," \u5206\u949F"):"".concat(L," \u5206\u949F")},gn=e(79273),rn=e(7221),wn=e(88826),yn=e(24816),_n=e(50400),ve=e(74353),ne=e.n(ve),he=e(27115),ee=e(78781),J=e(2904),n=e(74848),te,re,oe,ie,Hn,Xn,Fn,ae,se,le,de,ce={windows:{icon:"WindowsOutlined",name:"Windows"},linux:{icon:"LinuxOutlined",name:"Linux"},darwin:{icon:"AppleOutlined",name:"macOS"},ios:{icon:"AppleOutlined",name:"iOS"},android:{icon:"AndroidOutlined",name:"Android"},freebsd:{icon:"CodeOutlined",name:"FreeBSD"},openbsd:{icon:"CodeOutlined",name:"OpenBSD"},netbsd:{icon:"CodeOutlined",name:"NetBSD"},dragonfly:{icon:"CodeOutlined",name:"DragonFly BSD"},solaris:{icon:"GlobalOutlined",name:"Solaris"},illumos:{icon:"GlobalOutlined",name:"illumos"},aix:{icon:"GlobalOutlined",name:"AIX"}},xe=(0,$.rU)(function(i,a){var t=i.css,r=i.token,s=i.isDarkMode,L=a.compact;return{hostCard:t(te||(te=u()([`
        overflow: hidden;
        contain: paint;
        border-radius: `,`px;
        background: `,`;

        .ant-card-body {
            padding: 0;
        }
    `])),r.borderRadiusLG,r.colorBgContainer),host:t(re||(re=u()([`
        position: relative;
        min-height: 108px;
        padding: 14px 46px 14px 16px;
        overflow: hidden;
        isolation: isolate;
        background:
            radial-gradient(ellipse at 2% -35%, color-mix(in srgb, `," ",`%, transparent) 0%, transparent 42%),
            radial-gradient(ellipse at 50% 135%, color-mix(in srgb, `," ",`%, transparent) 0%, transparent 38%),
            radial-gradient(ellipse at 98% -35%, color-mix(in srgb, `," ",`%, transparent) 0%, transparent 40%),
            `,`;

        &::before,
        &::after {
            content: "";
            position: absolute;
            z-index: 0;
            width: 270px;
            height: 120px;
            border-radius: 999px;
            pointer-events: none;
            filter: blur(34px) saturate(1.12);
            transform: translate3d(0, 0, 0);
            transition: opacity .28s ease, transform .36s cubic-bezier(.2, .8, .2, 1);
            will-change: opacity, transform;
        }

        &::before {
            top: -78px;
            left: -58px;
            background: color-mix(in srgb, `," ",`%, transparent);
            opacity: `,`;
        }

        &::after {
            right: -48px;
            bottom: -82px;
            background: color-mix(in srgb, `," ",`%, transparent);
            opacity: `,`;
        }

        > .ant-row {
            position: relative;
            z-index: 2;
            min-width: 0;
        }

        @media (hover: hover) {
            &:hover::before {
                opacity: `,`;
                transform: translate3d(10px, 5px, 0) scale(1.06);
            }

            &:hover::after {
                opacity: `,`;
                transform: translate3d(-8px, -4px, 0) scale(1.08);
            }
        }

        @media (prefers-reduced-motion: reduce) {
            &::before,
            &::after {
                transition: none;
            }
        }

        @media (max-width: 991px) {
            min-height: auto;
            padding: 14px 44px 14px 16px;
        }

        @media (max-width: 575px) {
            padding: 12px 14px;
        }
    `])),r.colorPrimary,s?18:12,r.colorInfo,s?13:9,r.colorSuccess,s?10:7,r.colorBgContainer,r.colorPrimary,s?50:38,s?.34:.3,r.colorSuccess,s?42:32,s?.26:.22,s?.42:.36,s?.34:.28),hostIdentity:t(oe||(oe=u()([`
        position: relative;
        min-width: 0;
        min-height: 78px;
        align-items: center;

        > .ant-col {
            min-width: 0;
        }

        @media (max-width: 991px) {
            min-height: auto;
        }

        @media (max-width: 575px) {
            min-height: 52px;
            padding-right: 30px;
        }
    `]))),hostWatermark:t(ie||(ie=u()([`
        position: absolute;
        z-index: 1;
        top: 50%;
        right: 6%;
        width: 230px;
        height: 230px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: `,`;
        font-size: 208px;
        line-height: 1;
        opacity: `,`;
        pointer-events: none;
        filter: blur(`,`px);
        transform: translateY(-48%) rotate(-18deg);

        > span {
            font-size: inherit;
            line-height: inherit;
        }

        @media (max-width: 991px) {
            right: -10px;
            width: 204px;
            height: 204px;
            font-size: 184px;
            opacity: `,`;
        }

        @media (max-width: 575px) {
            top: 42%;
            right: -46px;
            width: 148px;
            height: 148px;
            font-size: 132px;
            opacity: `,`;
        }
    `])),r.colorPrimary,s?.045:.04,s?11:9,s?.04:.035,s?.028:.022),systemMark:t(Hn||(Hn=u()([`
        position: relative;
        width: 72px;
        height: 72px;
        flex: none;
        color: `,`;

        .system-visual {
            width: 100%;
            height: 100%;
            display: block;
            overflow: visible;
        }

        .system-core {
            fill: color-mix(in srgb, `," 78%, ",`);
            stroke: color-mix(in srgb, `," 34%, ",`);
            stroke-width: 1.2;
        }

        .system-ring,
        .system-arc {
            fill: none;
            stroke: currentColor;
            vector-effect: non-scaling-stroke;
        }

        .system-ring {
            stroke-width: 1;
            stroke-dasharray: 3 7;
            opacity: `,`;
        }

        .system-arc {
            stroke-width: 1.8;
            stroke-linecap: round;
            stroke-dasharray: 42 156;
            opacity: `,`;
        }

        .system-orbit {
            transform-box: view-box;
            transform-origin: 48px 48px;
            animation: systemOrbit 14s linear infinite;
        }

        .system-node {
            fill: `,`;
            stroke: currentColor;
            stroke-width: 1.5;
            vector-effect: non-scaling-stroke;
        }

        .system-node-active {
            fill: `,`;
            stroke: `,`;
            stroke-width: 2.5;
        }

        .system-icon {
            position: absolute;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 26px;
            line-height: 1;
        }

        @keyframes systemOrbit {
            to {
                transform: rotate(360deg);
            }
        }

        @media (prefers-reduced-motion: reduce) {
            .system-orbit {
                animation: none;
            }
        }

        @media (max-width: 575px) {
            width: 52px;
            height: 52px;

            .system-icon { font-size: 20px; }
        }
    `])),r.colorPrimary,r.colorPrimaryBg,r.colorBgContainer,r.colorPrimary,r.colorBorderSecondary,s?.34:.26,s?.7:.6,r.colorBgContainer,r.colorPrimary,r.colorBgContainer),hostInfo:t(Xn||(Xn=u()([`
        min-width: 0;
        flex: 1;
    `]))),systemLabel:t(Fn||(Fn=u()([`
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 8px;
        color: `,`;
        font-size: 10px;
        font-weight: 600;
        line-height: 16px;

        &::before {
            content: "";
            width: 18px;
            height: 2px;
            flex: none;
            border-radius: 1px;
            background: currentColor;
            opacity: .72;
        }

        @media (max-width: 575px) {
            gap: 6px;
            font-size: 9px;
            line-height: 14px;

            &::before {
                width: 12px;
            }
        }
    `])),r.colorPrimary),hostName:t(ae||(ae=u()([`
        margin: 3px 0 0;
        color: `,`;
        font-size: 21px;
        font-weight: 700;
        line-height: 28px;
        letter-spacing: 0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;

        @media (max-width: 575px) {
            margin-top: 2px;
            font-size: 17px;
            line-height: 23px;
        }
    `])),r.colorTextHeading),hostSummary:t(se||(se=u()([`
        min-width: 0;
        height: 100%;
        border-left: 1px solid color-mix(in srgb, `,` 72%, transparent);

        > .ant-row {
            height: 100%;
        }

        > .ant-row > .ant-col {
            min-width: 0;
            display: flex;
            align-items: center;
        }

        > .ant-row > .ant-col + .ant-col {
            border-left: 1px solid color-mix(in srgb, `,` 72%, transparent);
        }

        .host-summary-item {
            width: 100%;
            min-width: 0;
            padding: 3px 14px;
        }

        .summary-label,
        .summary-value,
        .summary-meta {
            display: block;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .summary-label {
            display: flex;
            align-items: center;
            gap: 6px;
            color: `,`;
            font-size: 10px;
            line-height: 16px;
        }

        .summary-label > span {
            flex: none;
            color: `,`;
            font-size: 12px;
        }

        .summary-value {
            margin-top: 3px;
            color: `,`;
            font-size: 15px;
            font-weight: 700;
            line-height: 22px;
        }

        .summary-meta {
            margin-top: 1px;
            color: `,`;
            font-size: 10px;
            line-height: 15px;
        }

        @media (max-width: 991px) {
            height: auto;
            border-top: 1px solid color-mix(in srgb, `,` 72%, transparent);
            border-left: 0;

            .host-summary-item {
                padding: 3px 12px;
            }
        }

        @media (max-width: 575px) {
            > .ant-row > .ant-col + .ant-col {
                border-top: 0;
                border-left: 1px solid color-mix(in srgb, `,` 72%, transparent);
            }

            .host-summary-item {
                padding: 8px 6px 7px;
                text-align: center;
            }

            .summary-label {
                justify-content: center;
                gap: 4px;
                font-size: 9px;
                line-height: 14px;
            }

            .summary-label > span {
                font-size: 11px;
            }

            .summary-value {
                margin-top: 2px;
                font-size: 12px;
                line-height: 18px;
            }

            .summary-meta {
                display: none;
            }
        }
    `])),r.colorBorderSecondary,r.colorBorderSecondary,r.colorTextTertiary,r.colorPrimary,r.colorTextHeading,r.colorTextQuaternary,r.colorBorderSecondary,r.colorBorderSecondary),restartButton:t(le||(le=u()([`
        position: absolute;
        z-index: 3;
        top: 10px;
        right: 10px;

        &.ant-btn {
            width: 30px;
            height: 30px;
            padding: 0;
        }
    `]))),restartMenu:t(de||(de=u()([`
        && {
            min-width: `,`px;
            padding: 0;
            border-radius: `,`px;
        }

        && .ant-dropdown-menu {
            padding: 3px !important;
            border-radius: `,`px !important;
            background: `,`;
            box-shadow: `,`;
        }

        && .ant-dropdown-menu-item {
            min-height: `,`px;
            padding: 0 `,`px !important;
            border-radius: `,`px !important;
            color: `,`;
            font-size: `,`px;
        }

        && .ant-dropdown-menu-item-danger {
            color: `,`;
        }
    `])),L?112:124,r.borderRadius,r.borderRadius,r.colorBgElevated,r.boxShadowSecondary,L?28:30,L?7:8,r.borderRadiusSM,r.colorTextSecondary,L?11:12,r.colorError)}}),fe=o.memo(function(i){var a,t=i.info,r=i.statis,s=(0,k.DP)(),L=xe({compact:!!(!(s==null||(a=s.isCompactTheme)===null||a===void 0)&&a.call(s))}),R=L.styles,H=gn.A.useApp(),I=H.message,w=H.modal,y=(0,o.useState)(!1),j=K()(y,2),O=j[0],b=j[1],V=(0,o.useRef)(null),G=(0,o.useRef)(null),E=(0,o.useRef)(!0),U=(0,o.useState)(!1),F=K()(U,2),h=F[0],C=F[1],p=t,l=r==null?void 0:r.cpu,x=r==null?void 0:r.memory,N=ce[String((p==null?void 0:p.os)||"").toLowerCase()]||{icon:"DesktopOutlined",name:(p==null?void 0:p.os)||"\u672A\u77E5\u7CFB\u7EDF"},Y=[p==null?void 0:p.platform,p==null?void 0:p.platform_version].filter(Boolean).join(" ")||N.name;(0,o.useEffect)(function(){return E.current=!0,function(){var S,dn;E.current=!1,(S=G.current)===null||S===void 0||S.abort(),(dn=V.current)===null||dn===void 0||dn.destroy()}},[]);var kn=function(dn){if(!(V.current||G.current)){var An=dn==="panel"?"\u91CD\u542F\u9762\u677F":"\u91CD\u542F\u670D\u52A1\u5668",v=dn==="panel"?"panel_start_time":"boot_time",M=w.confirm({title:An,content:dn==="panel"?"\u91CD\u542F\u671F\u95F4\u9762\u677F\u548C\u6258\u7BA1\u670D\u52A1\u5C06\u6682\u65F6\u4E2D\u65AD\uFF0C\u6062\u590D\u540E\u81EA\u52A8\u5237\u65B0\u5F53\u524D\u9875\u9762\u3002\u662F\u5426\u7EE7\u7EED\uFF1F":"\u670D\u52A1\u5668\u4E0A\u7684\u6240\u6709\u670D\u52A1\u5C06\u6682\u65F6\u4E2D\u65AD\uFF0C\u6062\u590D\u540E\u81EA\u52A8\u5237\u65B0\u5F53\u524D\u9875\u9762\u3002\u662F\u5426\u7EE7\u7EED\uFF1F",okText:"\u91CD\u542F",cancelText:"\u53D6\u6D88",okButtonProps:{danger:!0},afterClose:function(){V.current===M&&(V.current=null)},onOk:(function(){var tn=Q()(W()().mark(function vn(){var pn,Nn,Pn,en,Bn,Rn,qn,jn,Cn,$n,bn,Un,Kn,zn;return W()().wrap(function(f){for(;;)switch(f.prev=f.next){case 0:if(!G.current){f.next=2;break}return f.abrupt("return");case 2:return pn=new AbortController,G.current=pn,b(!0),f.prev=5,f.next=8,(0,ln.yV)();case 8:if(en=f.sent,!(!E.current||pn.signal.aborted)){f.next=11;break}return f.abrupt("return");case 11:if(!((en==null?void 0:en.code)!==200||!(en!=null&&en.data))){f.next=14;break}throw I.error((en==null?void 0:en.message)||"\u83B7\u53D6\u7CFB\u7EDF\u4FE1\u606F\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5"),new Error("\u83B7\u53D6\u7CFB\u7EDF\u4FE1\u606F\u5931\u8D25");case 14:return Bn=en==null||(Pn=en.data)===null||Pn===void 0?void 0:Pn[v],Rn=18e4,qn=Date.now()+Rn,Nn=(0,ee.Qc)(pn.signal,Rn),f.next=20,(0,ln.h8)({body:{action:dn}});case 20:if(jn=f.sent,!(!E.current||pn.signal.aborted)){f.next=23;break}return f.abrupt("return");case 23:if((jn==null?void 0:jn.code)===200){f.next=26;break}throw I.error((jn==null?void 0:jn.message)||"\u63D0\u4EA4\u91CD\u542F\u6307\u4EE4\u5931\u8D25"),new Error("\u63D0\u4EA4\u91CD\u542F\u6307\u4EE4\u5931\u8D25");case 26:M.update({title:"\u6B63\u5728".concat(An),content:"\u6B63\u5728\u7B49\u5F85\u670D\u52A1\u6062\u590D\uFF0C\u8FDE\u63A5\u6210\u529F\u540E\u5C06\u81EA\u52A8\u5237\u65B0\u3002",closable:!1,mask:{closable:!1},keyboard:!1,cancelButtonProps:{style:{display:"none"}}}),Cn=!1;case 28:if(!(E.current&&!pn.signal.aborted&&Date.now()<qn)){f.next=45;break}return f.next=31,(0,ee.Ay)("/system/info",{timeout:3e3,signal:pn.signal,errorHandler:function(){}}).catch(function(){});case 31:if(bn=f.sent,!(!E.current||pn.signal.aborted)){f.next=34;break}return f.abrupt("return");case 34:if(Un=bn==null||($n=bn.data)===null||$n===void 0?void 0:$n[v],Kn=Bn&&Un?String(Bn)!==String(Un):Cn,!((bn==null?void 0:bn.code)===200&&Kn)){f.next=40;break}return I.success("".concat(An,"\u5B8C\u6210")),window.location.reload(),f.abrupt("return");case 40:return bn||(Cn=!0),f.next=43,new Promise(function(P){return window.setTimeout(P,2e3)});case 43:f.next=28;break;case 45:E.current&&!pn.signal.aborted&&I.warning("\u6682\u672A\u8FDE\u63A5\u5230\u670D\u52A1\uFF0C\u8BF7\u7A0D\u540E\u5237\u65B0\u9875\u9762");case 46:return f.prev=46,(zn=Nn)===null||zn===void 0||zn(),G.current=null,E.current&&b(!1),f.finish(46);case 51:case"end":return f.stop()}},vn,null,[[5,,46,51]])}));function q(){return tn.apply(this,arguments)}return q})()});V.current=M,Promise.resolve(M)}};return(0,n.jsx)(rn.A,{className:R.hostCard,variant:"borderless",children:(0,n.jsxs)("section",{className:R.host,children:[(0,n.jsx)("span",{className:R.hostWatermark,children:(0,n.jsx)(k.In,{type:N.icon})}),(0,n.jsxs)(z.A,{gutter:[Z,Z],align:"middle",children:[(0,n.jsx)(A.A,{xs:24,lg:10,children:(0,n.jsxs)(z.A,{className:R.hostIdentity,gutter:Z,wrap:!1,children:[(0,n.jsx)(A.A,{flex:"none",children:(0,n.jsxs)("div",{className:R.systemMark,children:[(0,n.jsxs)("svg",{className:"system-visual",viewBox:"0 0 96 96","aria-hidden":"true",children:[(0,n.jsx)("circle",{className:"system-core",cx:"48",cy:"48",r:"25"}),(0,n.jsx)("circle",{className:"system-ring",cx:"48",cy:"48",r:"39"}),(0,n.jsxs)("g",{className:"system-orbit",children:[(0,n.jsx)("circle",{className:"system-arc",cx:"48",cy:"48",r:"33"}),(0,n.jsx)("circle",{className:"system-node system-node-active",cx:"48",cy:"9",r:"3"}),(0,n.jsx)("circle",{className:"system-node",cx:"87",cy:"48",r:"2.5"}),(0,n.jsx)("circle",{className:"system-node",cx:"48",cy:"87",r:"2.5"})]})]}),(0,n.jsx)("span",{className:"system-icon",children:(0,n.jsx)(k.In,{type:N.icon})})]})}),(0,n.jsx)(A.A,{flex:"auto",children:(0,n.jsxs)("div",{className:R.hostInfo,children:[(0,n.jsx)("div",{className:R.systemLabel,children:N.name}),(0,n.jsx)("h1",{className:R.hostName,children:(p==null?void 0:p.hostname)||"\u672A\u77E5\u4E3B\u673A"})]})})]})}),(0,n.jsx)(A.A,{xs:24,lg:14,children:(0,n.jsx)("div",{className:R.hostSummary,children:(0,n.jsxs)(z.A,{gutter:[0,Z],align:"middle",children:[(0,n.jsx)(A.A,{span:8,children:(0,n.jsxs)("div",{className:"host-summary-item",children:[(0,n.jsxs)("span",{className:"summary-label",children:[(0,n.jsx)(k.In,{type:"ClockCircleOutlined"}),"\u8FDE\u7EED\u8FD0\u884C"]}),(0,n.jsx)("strong",{className:"summary-value",children:fn(p==null?void 0:p.uptime)}),(0,n.jsxs)("small",{className:"summary-meta",children:["\u542F\u52A8\u4E8E ",p!=null&&p.boot_time?ne().unix(d(p.boot_time)).format("YYYY-MM-DD HH:mm"):"--"]})]})}),(0,n.jsx)(A.A,{span:8,children:(0,n.jsxs)("div",{className:"host-summary-item",children:[(0,n.jsxs)("span",{className:"summary-label",children:[(0,n.jsx)(k.In,{type:"DesktopOutlined"}),"\u64CD\u4F5C\u7CFB\u7EDF"]}),(0,n.jsx)(wn.A,{title:h?Y:void 0,mouseLeaveDelay:.3,styles:{root:{pointerEvents:"auto"},container:{cursor:"text",userSelect:"text"}},children:(0,n.jsx)(yn.A.Text,{className:"summary-value",ellipsis:{onEllipsis:C},children:Y})}),(0,n.jsxs)("small",{className:"summary-meta",children:["\u5185\u6838 ",(p==null?void 0:p.kernel_version)||"--"]})]})}),(0,n.jsx)(A.A,{span:8,children:(0,n.jsxs)("div",{className:"host-summary-item",children:[(0,n.jsxs)("span",{className:"summary-label",children:[(0,n.jsx)(k.In,{type:"ClusterOutlined"}),"\u4E3B\u673A\u67B6\u6784"]}),(0,n.jsx)("strong",{className:"summary-value",children:(p==null?void 0:p.kernel_arch)||"--"}),(0,n.jsxs)("small",{className:"summary-meta",children:[d(l==null?void 0:l.processor)," \u4E2A\u903B\u8F91\u5904\u7406\u5668 \xB7 ",(0,J.X)(x==null?void 0:x.total)," \u5185\u5B58"]})]})})]})})})]}),(0,n.jsx)(he.A,{trigger:["click"],placement:"bottomRight",arrow:{pointAtCenter:!0},classNames:{root:R.restartMenu},disabled:O,menu:{items:[{key:"panel",label:"\u91CD\u542F\u9762\u677F",onClick:function(){return kn("panel")}},{type:"divider"},{key:"server",label:"\u91CD\u542F\u670D\u52A1\u5668",danger:!0,onClick:function(){return kn("server")}}]},children:(0,n.jsx)(_n.A,{className:R.restartButton,color:"default",variant:"text",size:"small","aria-label":"\u91CD\u542F",title:"\u91CD\u542F",loading:O,icon:(0,n.jsx)(k.In,{type:"PoweroffOutlined"})})})]})})}),ge=fe,En=e(59522),Ln=e(58679),ye=e(66585),Wn=["B/s","KB/s","MB/s","GB/s","TB/s","PB/s"],Mn=function(a){var t=Number(a);return Number.isFinite(t)?Math.max(0,t):0},be=function(a){for(var t=Mn(a),r=0;t>=1024&&r<Wn.length-1;)t/=1024,r+=1;var s=t>=100||r===0?0:t>=10?1:2;return"".concat(t.toLocaleString("zh-CN",{maximumFractionDigits:s})," ").concat(Wn[r])},Vn=function(a){var t=Number(a),r=a instanceof Date?a:new Date(Number.isFinite(t)?t:a);if(Number.isNaN(r.getTime()))return"--:--:--";var s=function(R){return String(R).padStart(2,"0")};return"".concat(s(r.getHours()),":").concat(s(r.getMinutes()),":").concat(s(r.getSeconds()))},Gn=o.memo(function(i){var a=i.className,t=i.data,r=i.first,s=i.second,L=i.isDarkMode,R=En.A.useToken(),H=R.token,I=(0,o.useRef)(null),w=(0,o.useRef)(null),y=(r==null?void 0:r.label)||"\u7CFB\u5217\u4E00",j=(s==null?void 0:s.label)||"\u7CFB\u5217\u4E8C",O=(r==null?void 0:r.color)||"#1677ff",b=(s==null?void 0:s.color)||"#13c2c2";(0,o.useEffect)(function(){var G=I.current;if(!(!G||typeof ResizeObserver=="undefined")){var E=Math.round(G.getBoundingClientRect().width),U=0,F=new ResizeObserver(function(h){var C=K()(h,1),p=C[0],l=Math.round(p.contentRect.width);l<=0||l===E||(E=l,window.clearTimeout(U),U=window.setTimeout(function(){var x,N=(x=w.current)===null||x===void 0?void 0:x.chart;N!=null&&N.forceFit&&N.forceFit()},80))});return F.observe(G),function(){window.clearTimeout(U),F.disconnect()}}},[]);var V=(0,o.useMemo)(function(){for(var G=L?"rgba(255,255,255,0.08)":"rgba(5,5,5,0.06)",E=Array.isArray(t)?D()(t).sort(function(l,x){return Mn(l==null?void 0:l.time)-Mn(x==null?void 0:x.time)}):[],U=E.flatMap(function(l){var x=Mn(l==null?void 0:l.time);return[{time:x,value:Mn(l==null?void 0:l.first),series:y},{time:x,value:Mn(l==null?void 0:l.second),series:j}]}),F=U.reduce(function(l,x){return Math.max(l,x.value)},0),h=0;h<Wn.length-1&&F>=Math.pow(1024,h+1);)h+=1;var C=Math.pow(1024,h),p=function(x){var N=Mn(x)/C,Y=N>=100||h===0?0:N>=10?1:2;return N.toLocaleString("zh-CN",{maximumFractionDigits:Y})};return{unit:Wn[h],config:{data:U,xField:"time",yField:"value",colorField:"series",seriesField:"series",keyField:"series",shape:"smooth",paddingTop:8,insetBottom:3,scale:{x:{type:"time",nice:!1},color:{domain:[y,j],range:[O,b]},y:{domainMin:0,domainMax:F>0?void 0:1,nice:!0}},axis:{x:{title:!1,line:!1,tick:!1,tickCount:5,labelFill:H.colorTextSecondary,labelFontSize:10,labelAutoHide:!0,labelAutoRotate:!1,labelFormatter:Vn,grid:!1},y:{title:!1,line:!1,tick:!1,tickCount:4,labelFill:H.colorTextSecondary,labelFontSize:10,labelFormatter:p,grid:!0,gridLineDash:[4,4],gridLineWidth:1,gridStroke:G,gridStrokeOpacity:1}},legend:!1,tooltip:{title:function(x){return Vn(x==null?void 0:x.time)},items:[function(l){return{name:(l==null?void 0:l.series)||"--",color:(l==null?void 0:l.series)===y?O:b,value:be(l==null?void 0:l.value)}}]},interaction:{tooltip:{shared:!0,series:!0,crosshairs:!0}},theme:{type:L?"classicDark":"classic"},autoFit:!0,style:{lineWidth:2,lineCap:"round",lineJoin:"round"},animate:{enter:{type:"fadeIn",duration:280},update:{type:null}}}}},[t,O,y,L,b,j,H.colorTextSecondary]);return(0,n.jsxs)("div",{ref:I,className:a,children:[(0,n.jsxs)("div",{className:"telemetry-toolbar",children:[(0,n.jsxs)("div",{className:"telemetry-legends",children:[(0,n.jsxs)("span",{className:"telemetry-legend",style:{"--legend-color":O},children:[(0,n.jsx)("i",{}),y]}),(0,n.jsxs)("span",{className:"telemetry-legend",style:{"--legend-color":b},children:[(0,n.jsx)("i",{}),j]})]}),(0,n.jsx)("span",{className:"telemetry-unit",children:V.unit})]}),(0,n.jsx)(ye.A,_()(_()({ref:w},V.config),{},{className:"telemetry-plot"}))]})});Gn.displayName="TelemetryChart";var ue=Gn,Qn,Zn,Jn,Yn,Ee,Me,He=(0,$.rU)(function(i){var a=i.css,t=i.token,r=i.isDarkMode;return{monitorCard:a(Qn||(Qn=u()([`
        height: 100%;
        overflow: hidden;
        border-radius: `,`px;

        .ant-card-head {
            min-height: 48px;
            margin-bottom: 0;
            border-bottom: 0;
            box-shadow: inset 0 -1px 0 `,`;
        }

        .ant-card-head-title,
        .ant-card-extra {
            min-width: 0;
        }

        .ant-card-extra {
            flex: none;
            margin-inline-start: 12px;
        }

        .ant-card-body {
            padding: 0;
            overflow: hidden;
        }
    `])),t.borderRadiusLG,r?"rgba(255,255,255,0.06)":"rgba(5,5,5,0.06)"),deviceSelect:a(Zn||(Zn=u()([`
        && {
            width: 86px;
            height: 28px;
            color: `,`;
            font-size: 12px;
        }

        && .ant-select-content,
        && .ant-select-content-value {
            min-width: 0;
            color: `,` !important;
            font-size: 12px !important;
            font-weight: 500 !important;
            line-height: 28px !important;
        }

        && .ant-select-suffix {
            color: `,`;
            font-size: 10px;
        }
    `])),t.colorTextSecondary,t.colorTextSecondary,t.colorTextTertiary),deviceDropdown:a(Jn||(Jn=u()([`
        && {
            padding: 3px;
            border-radius: `,`px;
        }

        && .ant-select-item {
            min-height: 30px;
            padding: 5px 9px;
            border-radius: `,`px;
            color: `,`;
            font-size: 12px;
            line-height: 20px;
        }

        && .ant-select-item-option-content {
            overflow: hidden;
            font-size: 12px;
            font-weight: 400;
            text-overflow: ellipsis;
        }

        && .ant-select-item-option-active:not(.ant-select-item-option-disabled) {
            background: `,`;
        }

        && .ant-select-item-option-selected:not(.ant-select-item-option-disabled) {
            background: `,`;
            color: `,`;
            font-weight: 500;
        }
    `])),t.borderRadius,t.borderRadiusSM,t.colorTextSecondary,t.colorFillQuaternary,t.colorFillTertiary,t.colorText),monitorBody:a(Yn||(Yn=u()([`
        width: 100%;
        min-width: 0;
        display: grid;
        grid-template-columns: minmax(0, 1fr);
        align-items: stretch;
    `]))),throughputStats:a(Ee||(Ee=u()([`
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        align-content: center;
        padding: 10px 14px 0;

        .throughput-stat {
            position: relative;
            min-width: 0;
            min-height: 66px;
            padding: 8px 14px;
            display: flex;
            flex-direction: column;
            justify-content: center;
            box-sizing: border-box;
        }

        .throughput-stat + .throughput-stat::before {
            content: "";
            position: absolute;
            inset-inline-start: 0;
            top: 16px;
            bottom: 16px;
            width: 1px;
            background: `,`;
        }

        .stat-label {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 7px;
            overflow: hidden;
            color: `,`;
            font-size: 11px;
            font-weight: 500;
            line-height: 16px;
            white-space: nowrap;
            text-overflow: ellipsis;
            letter-spacing: 0;
        }

        .stat-label::before {
            content: "";
            width: 3px;
            height: 12px;
            flex: none;
            border-radius: 2px;
            background: var(--stat-color);
        }

        .stat-value {
            min-width: 0;
            margin-top: 5px;
            display: flex;
            align-items: baseline;
            gap: 3px;
            overflow: hidden;
            white-space: nowrap;
        }

        .stat-value strong {
            min-width: 0;
            overflow: hidden;
            color: `,`;
            font-size: 17px;
            font-weight: 600;
            font-variant-numeric: tabular-nums;
            line-height: 24px;
            text-overflow: ellipsis;
            letter-spacing: 0;
        }

        .stat-value small {
            flex: none;
            color: `,`;
            font-size: 10px;
            font-weight: 500;
            line-height: 16px;
        }

        @media (max-width: 575px) {
            grid-template-columns: repeat(2, minmax(0, 1fr));

            .throughput-stat:nth-child(odd)::before {
                display: none;
            }

            .throughput-stat:nth-child(n + 3)::after {
                content: "";
                position: absolute;
                top: 0;
                inset-inline: 14px;
                height: 1px;
                display: block;
                background: `,`;
            }
        }
    `])),r?"rgba(255,255,255,0.08)":"rgba(5,5,5,0.08)",t.colorTextTertiary,t.colorTextHeading,t.colorTextTertiary,r?"rgba(255,255,255,0.08)":"rgba(5,5,5,0.08)"),throughputChart:a(Me||(Me=u()([`
        width: 100%;
        min-width: 0;
        height: 314px;
        margin: 8px 0 12px;
        padding: 0 14px;
        display: flex;
        flex-direction: column;
        overflow: hidden;

        .telemetry-toolbar {
            height: 26px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            flex: none;
            padding: 0 2px;
        }

        .telemetry-legends {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 14px;
        }

        .telemetry-legend {
            min-width: 0;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            color: `,`;
            font-size: 10px;
            font-weight: 500;
            line-height: 16px;
            white-space: nowrap;
        }

        .telemetry-legend i {
            width: 7px;
            height: 7px;
            flex: none;
            border-radius: 50%;
            background: var(--legend-color);
        }

        .telemetry-unit {
            flex: none;
            color: `,`;
            font-size: 10px;
            font-weight: 500;
            line-height: 16px;
        }

        .telemetry-plot {
            width: 100%;
            min-height: 0;
            flex: 1;
        }

        @media (max-width: 576px) {
            height: 276px;
            margin-bottom: 10px;
            padding: 0 10px;
        }
    `])),t.colorTextTertiary,t.colorTextQuaternary)}}),Ae=o.memo(function(i){var a=i.title,t=i.device,r=i.options,s=i.onDeviceChange,L=i.metrics,R=i.data,H=i.first,I=i.second,w=i.isDarkMode,y=He(),j=y.styles;return(0,n.jsx)(rn.A,{className:j.monitorCard,variant:"borderless",title:(0,n.jsx)(k.hE,{children:a}),extra:(0,n.jsx)(Ln.A,{className:j.deviceSelect,size:"small",value:t,options:r,variant:"filled",classNames:{popup:{root:j.deviceDropdown}},onChange:s}),children:(0,n.jsxs)("div",{className:j.monitorBody,children:[(0,n.jsx)("div",{className:j.throughputStats,children:L.map(function(O){return(0,n.jsxs)("div",{className:"throughput-stat",style:{"--stat-color":O.color},children:[(0,n.jsx)("span",{className:"stat-label",children:O.label}),(0,n.jsxs)("span",{className:"stat-value",children:[(0,n.jsx)("strong",{children:O.value}),O.suffix&&(0,n.jsx)("small",{children:O.suffix})]})]},O.label)})}),(0,n.jsx)(ue,{className:j.throughputChart,data:R,first:H,second:I,isDarkMode:w})]})})});Ae.displayName="MonitorCard";var Pe=Ae,Te,hn="__all__",we="#13c2c2",Oe=new Intl.Collator("en",{numeric:!0}),Xe=(0,$.rU)(function(i){var a=i.css;return{monitorColumn:a(Te||(Te=u()([`
        min-width: 0;
        display: flex;

        > .ant-card {
            width: 100%;
        }
    `])))}}),Se=o.memo(function(i){var a,t,r=i.io,s=i.throughputHistory,L=Xe(),R=L.styles,H=En.A.useToken(),I=H.token,w=(0,k.DP)(),y=(0,o.useState)(hn),j=K()(y,2),O=j[0],b=j[1],V=(0,o.useState)(hn),G=K()(V,2),E=G[0],U=G[1],F=(0,o.useMemo)(function(){return Object.entries((r==null?void 0:r.network)||{}).map(function(v){var M=K()(v,2),tn=M[0],q=M[1];return _()(_()({},q),{},{key:tn,name:tn})}).sort(function(v,M){return Oe.compare(v.key,M.key)||v.key.localeCompare(M.key,"en")})},[r==null?void 0:r.network]),h=(0,o.useMemo)(function(){return Object.entries((r==null?void 0:r.disk)||{}).map(function(v){var M=K()(v,2),tn=M[0],q=M[1];return _()(_()({},q),{},{key:tn,name:(q==null?void 0:q.name)||tn})}).sort(function(v,M){return Oe.compare(v.key,M.key)||v.key.localeCompare(M.key,"en")})},[r==null?void 0:r.disk]),C=(0,o.useMemo)(function(){return[{label:"\u5168\u90E8\u63A5\u53E3",value:hn}].concat(D()(F.map(function(v){return{label:v.name,value:v.key}})))},[F]),p=(0,o.useMemo)(function(){return[{label:"\u5168\u90E8\u78C1\u76D8",value:hn}].concat(D()(h.map(function(v){return{label:v.name,value:v.key}})))},[h]);(0,o.useEffect)(function(){O!==hn&&!F.some(function(v){return v.key===O})&&b(hn)},[O,F]),(0,o.useEffect)(function(){E!==hn&&!h.some(function(v){return v.key===E})&&U(hn)},[E,h]);var l=(0,o.useMemo)(function(){return O===hn?F:F.filter(function(v){return v.key===O})},[O,F]),x=(0,o.useMemo)(function(){return E===hn?h:h.filter(function(v){return v.key===E})},[E,h]),N=(0,o.useMemo)(function(){return l.reduce(function(v,M){return{receiveRate:v.receiveRate+d(M.recv_rate),sendRate:v.sendRate+d(M.send_rate),receiveTotal:v.receiveTotal+d(M.bytes_recv),sendTotal:v.sendTotal+d(M.bytes_sent)}},{receiveRate:0,sendRate:0,receiveTotal:0,sendTotal:0})},[l]),Y=(0,o.useMemo)(function(){return x.reduce(function(v,M){return{readRate:v.readRate+d(M.read_bytes_rate),writeRate:v.writeRate+d(M.write_bytes_rate),readCount:v.readCount+d(M.read_count_rate),writeCount:v.writeCount+d(M.write_count_rate)}},{readRate:0,writeRate:0,readCount:0,writeCount:0})},[x]),kn=(0,o.useMemo)(function(){return[{label:"\u4E0B\u8F7D\u901F\u7387",value:(0,J.X)(N.receiveRate),suffix:"/s",color:I.colorSuccess},{label:"\u4E0A\u4F20\u901F\u7387",value:(0,J.X)(N.sendRate),suffix:"/s",color:I.colorPrimary},{label:"\u7D2F\u8BA1\u63A5\u6536",value:(0,J.X)(N.receiveTotal),suffix:"",color:I.colorSuccess},{label:"\u7D2F\u8BA1\u53D1\u9001",value:(0,J.X)(N.sendTotal),suffix:"",color:I.colorPrimary}]},[N,I.colorPrimary,I.colorSuccess]),S=(0,o.useMemo)(function(){return[{label:"\u8BFB\u53D6\u901F\u7387",value:(0,J.X)(Y.readRate),suffix:"/s",color:I.colorInfo},{label:"\u5199\u5165\u901F\u7387",value:(0,J.X)(Y.writeRate),suffix:"/s",color:we},{label:"\u8BFB\u53D6 IOPS",value:mn(Y.readCount),suffix:"\u6B21/s",color:I.colorInfo},{label:"\u5199\u5165 IOPS",value:mn(Y.writeCount),suffix:"\u6B21/s",color:we}]},[Y,I.colorInfo]),dn=(0,o.useMemo)(function(){return s.map(function(v){var M=v.network||{},tn=O===hn?Object.values(M):M[O]?[M[O]]:[];return{time:v.time,first:tn.reduce(function(q,vn){return q+d(vn.receiveRate)},0),second:tn.reduce(function(q,vn){return q+d(vn.sendRate)},0)}})},[O,s]),An=(0,o.useMemo)(function(){return s.map(function(v){var M=v.disk||{},tn=E===hn?Object.values(M):M[E]?[M[E]]:[];return{time:v.time,first:tn.reduce(function(q,vn){return q+d(vn.readRate)},0),second:tn.reduce(function(q,vn){return q+d(vn.writeRate)},0)}})},[E,s]);return(0,n.jsxs)(z.A,{gutter:[Z,Z],align:"stretch",children:[(0,n.jsx)(A.A,{className:R.monitorColumn,xs:24,xl:12,children:(0,n.jsx)(Pe,{title:"\u7F51\u7EDC\u541E\u5410",device:O,options:C,onDeviceChange:b,metrics:kn,data:dn,first:{label:"\u4E0B\u8F7D",color:I.colorSuccess},second:{label:"\u4E0A\u4F20",color:I.colorPrimary},isDarkMode:w==null||(a=w.isDarkMode)===null||a===void 0?void 0:a.call(w)})}),(0,n.jsx)(A.A,{className:R.monitorColumn,xs:24,xl:12,children:(0,n.jsx)(Pe,{title:"\u78C1\u76D8 IO",device:E,options:p,onDeviceChange:U,metrics:S,data:An,first:{label:"\u8BFB\u53D6",color:I.colorInfo},second:{label:"\u5199\u5165",color:we},isDarkMode:w==null||(t=w.isDarkMode)===null||t===void 0?void 0:t.call(w)})})]})});Se.displayName="MonitorOverview";var Ve=Se,Ge=e(28073),_e=e(49216),Ne,Be,Qe=(0,$.rU)(function(i){var a=i.css,t=i.token,r=i.isDarkMode;return{diskPopover:a(Ne||(Ne=u()([`
        width: min(350px, calc(100vw - 28px));

        .disk-overview {
            --disk-color: `,`;
            padding: 10px 12px;
            border-radius: `,`px;
            background: `,`;
        }

        .disk-overview-title,
        .disk-overview-head,
        .disk-overview-meta,
        .disk-item-head,
        .disk-item-meta {
            min-width: 0;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .disk-overview-title {
            margin-bottom: 7px;
            color: `,`;
            font-size: 10px;
            line-height: 16px;
        }

        .disk-overview-title strong {
            color: `,`;
            font-size: 11px;
            font-weight: 600;
        }

        .disk-overview-head {
            gap: 16px;
        }

        .disk-overview-capacity,
        .disk-overview-value {
            min-width: 0;
        }

        .disk-overview-capacity span,
        .disk-overview-value span {
            display: block;
            color: `,`;
            font-size: 10px;
            line-height: 15px;
        }

        .disk-overview-capacity strong,
        .disk-overview-value strong {
            display: block;
            margin-top: 2px;
            font-size: 18px;
            font-weight: 600;
            line-height: 25px;
            white-space: nowrap;
        }

        .disk-overview-capacity strong {
            color: `,`;
        }

        .disk-overview-value {
            text-align: right;
        }

        .disk-overview-value strong {
            color: var(--disk-color);
        }

        .disk-overview-track,
        .disk-track {
            overflow: hidden;
            border-radius: 999px;
            background: color-mix(in srgb, var(--disk-color) 10%, `,`);
        }

        .disk-overview-track {
            height: 5px;
            margin-top: 9px;
        }

        .disk-overview-track i,
        .disk-track i {
            display: block;
            width: var(--disk-percent);
            height: 100%;
            border-radius: inherit;
            background: var(--disk-color);
            transition: width .3s ease;
        }

        .disk-overview-meta {
            justify-content: flex-start;
            gap: 16px;
            margin-top: 7px;
            color: `,`;
            font-size: 10px;
            line-height: 15px;
        }

        .disk-overview-meta b {
            margin-left: 4px;
            color: `,`;
            font-weight: 500;
        }

        .disk-list {
            max-height: 230px;
            display: flex;
            flex-direction: column;
            gap: 7px;
            margin-top: 9px;
            padding: 1px 1px 2px;
            overflow-y: auto;
        }

        .disk-item {
            --disk-color: `,`;
            padding: 9px 10px;
            border: 1px solid transparent;
            border-radius: `,`px;
            background: `,`;
            transition: border-color .2s ease, background-color .2s ease;
        }

        @media (hover: hover) {
            .disk-item:hover {
                border-color: color-mix(in srgb, var(--disk-color) 28%, `,`);
                background: color-mix(in srgb, var(--disk-color) `,"%, ",`);
            }
        }

        .disk-item-head {
            gap: 14px;
        }

        .disk-item-path {
            min-width: 0;
            flex: 1;
            display: block;
            color: `,`;
            font-size: 12px;
            font-weight: 600;
            line-height: 19px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .disk-item-percent {
            flex: none;
            color: var(--disk-color);
            font-size: 12px;
            font-weight: 600;
            line-height: 19px;
        }

        .disk-item-meta {
            gap: 10px;
            margin-top: 3px;
            color: `,`;
            font-size: 9px;
            line-height: 14px;
        }

        .disk-item-meta span {
            min-width: 0;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .disk-item-meta span:last-child {
            flex: none;
            color: `,`;
        }

        .disk-item-meta b {
            color: `,`;
            font-weight: 500;
        }

        .disk-item-capacity {
            min-width: 0;
            margin-top: 3px;
            display: grid;
            grid-template-columns: repeat(2, minmax(0, 1fr));
            gap: 10px;
            color: `,`;
            font-size: 9px;
            line-height: 14px;
        }

        .disk-item-capacity span:last-child {
            text-align: right;
        }

        .disk-item-capacity b {
            margin-left: 3px;
            color: `,`;
            font-weight: 500;
        }

        .disk-track {
            height: 3px;
            margin-top: 7px;
        }

        @media (max-width: 575px) {
            width: min(320px, calc(100vw - 24px));

            .disk-item-head {
                gap: 10px;
            }
        }
    `])),t.colorPrimary,t.borderRadiusSM,t.colorFillQuaternary,t.colorTextTertiary,t.colorTextSecondary,t.colorTextTertiary,t.colorTextHeading,t.colorFillSecondary,t.colorTextTertiary,t.colorTextSecondary,t.colorPrimary,t.borderRadiusSM,t.colorFillQuaternary,t.colorBorderSecondary,r?6:4,t.colorFillQuaternary,t.colorText,t.colorTextQuaternary,t.colorTextTertiary,t.colorTextSecondary,t.colorTextQuaternary,t.colorTextSecondary),empty:a(Be||(Be=u()([`
        padding: 38px 0 30px;

        .ant-empty-description {
            color: `,`;
            font-size: 12px;
        }
    `])),t.colorTextTertiary)}}),Ze=o.memo(function(i){var a=i.diskFree,t=i.diskPercent,r=i.diskSummary,s=i.summaryDisks,L=i.sizeBase,R=L===void 0?1024:L,H=Qe(),I=H.styles,w=En.A.useToken(),y=w.token,j=sn(t),O=j>=90?y.colorError:j>=75?y.colorWarning:y.colorPrimary;return(0,n.jsxs)("div",{className:I.diskPopover,children:[(0,n.jsxs)("div",{className:"disk-overview",style:{"--disk-color":O,"--disk-percent":"".concat(j,"%")},children:[(0,n.jsxs)("div",{className:"disk-overview-title",children:[(0,n.jsx)("strong",{children:"\u5BB9\u91CF\u6982\u89C8"}),(0,n.jsxs)("span",{children:[s.length," \u4E2A\u6302\u8F7D\u70B9"]})]}),(0,n.jsxs)("div",{className:"disk-overview-head",children:[(0,n.jsxs)("div",{className:"disk-overview-capacity",children:[(0,n.jsx)("span",{children:"\u603B\u5BB9\u91CF"}),(0,n.jsx)("strong",{children:(0,J.X)(r.total,R)})]}),(0,n.jsxs)("div",{className:"disk-overview-value",children:[(0,n.jsx)("span",{children:"\u6574\u4F53\u4F7F\u7528\u7387"}),(0,n.jsx)("strong",{children:un(j)})]})]}),(0,n.jsx)("div",{className:"disk-overview-track",children:(0,n.jsx)("i",{})}),(0,n.jsxs)("div",{className:"disk-overview-meta",children:[(0,n.jsxs)("span",{children:["\u5DF2\u4F7F\u7528",(0,n.jsx)("b",{children:(0,J.X)(r.used,R)})]}),(0,n.jsxs)("span",{children:["\u53EF\u7528\u7A7A\u95F4",(0,n.jsx)("b",{children:(0,J.X)(a,R)})]})]})]}),s.length>0?(0,n.jsx)("div",{className:"disk-list",children:s.map(function(b,V){var G,E,U,F,h=sn(b==null||(G=b.size)===null||G===void 0?void 0:G.percentage),C=d(b==null||(E=b.size)===null||E===void 0?void 0:E.total),p=d(b==null||(U=b.size)===null||U===void 0?void 0:U.used),l=(b==null||(F=b.size)===null||F===void 0?void 0:F.un_used)===void 0?Math.max(0,C-p):d(b.size.un_used),x=h>=90?y.colorError:h>=75?y.colorWarning:y.colorPrimary;return(0,n.jsxs)("div",{className:"disk-item",style:{"--disk-color":x,"--disk-percent":"".concat(h,"%")},children:[(0,n.jsxs)("div",{className:"disk-item-head",children:[(0,n.jsx)(wn.A,{title:b.path,placement:"topLeft",children:(0,n.jsx)("strong",{className:"disk-item-path",children:b.path||"\u672A\u77E5\u6302\u8F7D\u70B9"})}),(0,n.jsx)("span",{className:"disk-item-percent",children:un(h)})]}),(0,n.jsxs)("div",{className:"disk-item-meta",children:[(0,n.jsxs)("span",{children:[b.filesystem||"--"," \xB7 ",b.type||"\u672A\u77E5\u7C7B\u578B"]}),(0,n.jsxs)("span",{children:["\u603B\u8BA1 ",(0,n.jsx)("b",{children:(0,J.X)(C,R)})]})]}),(0,n.jsxs)("div",{className:"disk-item-capacity",children:[(0,n.jsxs)("span",{children:["\u5DF2\u7528",(0,n.jsx)("b",{children:(0,J.X)(p,R)})]}),(0,n.jsxs)("span",{children:["\u53EF\u7528",(0,n.jsx)("b",{children:(0,J.X)(l,R)})]})]}),(0,n.jsx)("div",{className:"disk-track",children:(0,n.jsx)("i",{})})]},"".concat(b.path,"-").concat(b.filesystem,"-").concat(V))})}):(0,n.jsx)(_e.A,{className:I.empty,image:_e.A.PRESENTED_IMAGE_SIMPLE,description:"\u6682\u65E0\u6570\u636E"})]})}),Je=Ze,Re,Ye=(0,$.rU)(function(i){var a=i.css,t=i.token,r=i.isDarkMode;return{resource:a(Re||(Re=u()([`
        --resource-color: `,`;
        position: relative;
        min-width: 0;
        min-height: 138px;
        padding: 15px 16px;
        background:
            radial-gradient(circle at 108% 118%, color-mix(in srgb, var(--resource-color) `,`%, transparent) 0%, transparent 54%),
            radial-gradient(circle at 18% 45%, color-mix(in srgb, var(--resource-color) `,`%, transparent) 0%, transparent 58%),
            `,`;
        overflow: hidden;
        isolation: isolate;

        > .ant-row {
            position: relative;
            z-index: 1;
            min-height: 108px;
        }

        > .ant-row > .ant-col {
            min-width: 0;
        }

        &::before {
            content: "";
            position: absolute;
            z-index: 0;
            right: -62px;
            bottom: -82px;
            width: 176px;
            height: 176px;
            border-radius: 50%;
            background: color-mix(in srgb, var(--resource-color) `,`%, transparent);
            opacity: `,`;
            pointer-events: none;
            filter: blur(30px) saturate(1.12);
            transform: translate3d(0, 0, 0) scale(1);
            transition: transform .34s cubic-bezier(.2, .8, .2, 1);
            will-change: transform;
        }

        &::after {
            content: "";
            position: absolute;
            z-index: 0;
            top: 12px;
            left: -38px;
            width: 118px;
            height: 118px;
            border-radius: 50%;
            background: color-mix(in srgb, var(--resource-color) `,`%, transparent);
            opacity: `,`;
            pointer-events: none;
            filter: blur(25px) saturate(1.08);
            transform: translate3d(0, 0, 0) scale(.94);
            transition: transform .34s cubic-bezier(.2, .8, .2, 1);
            will-change: transform;
        }

        .resource-watermark {
            position: absolute;
            right: -25px;
            bottom: -29px;
            color: var(--resource-color);
            font-size: 138px;
            line-height: 1;
            opacity: `,`;
            pointer-events: none;
            filter: blur(`,`px);
            transform: translate3d(0, 0, 0) rotate(-24deg) scale(.98);
            transition: transform .34s cubic-bezier(.2, .8, .2, 1);
            will-change: transform;
            backface-visibility: hidden;
        }

        .resource-ring {
            position: relative;
            z-index: 1;
            width: 100px;
            height: 100px;
            flex: none;
            contain: layout paint;
            isolation: isolate;
            transform: translate3d(0, 0, 0);
            backface-visibility: hidden;
        }

        .resource-ring svg {
            width: 100%;
            height: 100%;
            display: block;
        }

        .ring-track,
        .ring-progress {
            fill: none;
            stroke-width: 8;
        }

        .ring-track {
            stroke: color-mix(in srgb, var(--resource-color) 8%, `,`);
        }

        .ring-progress {
            stroke: var(--resource-color);
            stroke-linecap: round;
            transition: stroke-dashoffset .25s ease;
        }

        .ring-content {
            position: absolute;
            inset: 0;
            padding: 0 11px;
            display: flex;
            align-items: center;
            justify-content: center;
            box-sizing: border-box;
            color: `,`;
            font-size: 15px;
            font-weight: 700;
            line-height: 22px;
            letter-spacing: 0;
            text-align: center;
            white-space: nowrap;
        }

        .ring-content.long {
            font-size: 13px;
            line-height: 19px;
        }

        .resource-copy {
            position: relative;
            z-index: 1;
            min-width: 0;
            flex: 1;
        }

        .resource-title {
            min-width: 0;
            display: flex;
            align-items: center;
            flex-wrap: nowrap;
            gap: 7px;
            color: `,`;
            font-size: 13px;
            line-height: 20px;
        }

        .resource-title strong {
            min-width: 0;
            font-weight: 600;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .resource-icon {
            flex: none;
            color: var(--resource-color);
            font-size: 16px;
        }

        .resource-detail {
            margin-top: 5px;
            color: `,`;
            font-size: 12px;
            line-height: 18px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .resource-meta {
            margin-top: 6px;
            color: `,`;
            font-size: 11px;
            line-height: 17px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        @media (hover: hover) {
            &:hover::before {
                transform: translate3d(-8px, -6px, 0) scale(1.1);
            }

            &:hover::after {
                transform: translate3d(7px, 3px, 0) scale(1.08);
            }

            &:hover .resource-watermark {
                transform: translate3d(-3px, -2px, 0) rotate(-21deg) scale(1.02);
            }
        }

        @media (prefers-reduced-motion: reduce) {
            &::before,
            &::after,
            .resource-watermark,
            .ring-progress {
                transition: none;
            }
        }

        @media (max-width: 576px) {
            min-height: 124px;

            > .ant-row {
                min-height: 94px;
            }

            .resource-ring {
                width: 92px;
                height: 92px;
            }
        }
    `])),t.colorPrimary,r?18:13,r?7:5,t.colorBgContainer,r?48:38,r?.28:.22,r?38:30,r?.08:.055,r?.045:.075,r?8:6,t.colorFillSecondary,t.colorTextHeading,t.colorTextHeading,t.colorTextTertiary,t.colorTextQuaternary)}}),qe=o.memo(function(i){var a=i.color,t=i.icon,r=i.title,s=i.value,L=i.detail,R=i.meta,H=i.percent,I=i.detailTooltip,w=I===void 0?!0:I,y=Ye(),j=y.styles;return(0,n.jsxs)("div",{className:j.resource,style:{"--resource-color":a},children:[(0,n.jsx)("span",{className:"resource-watermark",children:(0,n.jsx)(k.In,{type:t})}),(0,n.jsxs)(z.A,{gutter:Z,align:"middle",wrap:!1,children:[(0,n.jsx)(A.A,{flex:"none",children:(0,n.jsxs)("div",{className:"resource-ring",children:[(0,n.jsxs)("svg",{viewBox:"0 0 72 72","aria-hidden":"true",children:[(0,n.jsx)("circle",{className:"ring-track",cx:"36",cy:"36",r:"30"}),(0,n.jsx)("circle",{className:"ring-progress",cx:"36",cy:"36",r:"30",pathLength:"100",strokeDasharray:"100",strokeDashoffset:100-sn(H),transform:"rotate(-90 36 36)"})]}),(0,n.jsx)("span",{className:"ring-content ".concat(s.length>5?"long":""),children:s})]})}),(0,n.jsx)(A.A,{flex:"auto",children:(0,n.jsxs)("div",{className:"resource-copy",children:[(0,n.jsxs)("div",{className:"resource-title",children:[(0,n.jsx)("span",{className:"resource-icon",children:(0,n.jsx)(k.In,{type:t})}),(0,n.jsx)("strong",{children:r})]}),w?(0,n.jsx)(wn.A,{title:L,children:(0,n.jsx)("div",{className:"resource-detail",children:L})}):(0,n.jsx)("div",{className:"resource-detail",children:L}),(0,n.jsx)("div",{className:"resource-meta",children:R})]})})]})]})}),pe=qe,ze,nt=(0,$.rU)(function(i){var a=i.css,t=i.token;return{metricCard:a(ze||(ze=u()([`
        height: 100%;
        overflow: hidden;
        contain: paint;
        border-radius: `,`px;
        background: `,`;

        .ant-card-body {
            height: 100%;
            padding: 0;
        }
    `])),t.borderRadiusLG,t.colorBgContainer)}}),et=o.memo(function(i){var a=i.statis,t=i.diskSizeBase,r=t===void 0?1024:t,s=nt(),L=s.styles,R=En.A.useToken(),H=R.token,I=(a==null?void 0:a.disks)||[],w=a==null?void 0:a.cpu,y=a==null?void 0:a.memory,j=a==null?void 0:a.load,O=(0,o.useMemo)(function(){return Array.from(new Map(I.map(function(h){var C,p="".concat(h.filesystem||h.path,"-").concat(d(h==null||(C=h.size)===null||C===void 0?void 0:C.total));return[p,h]})).values())},[I]),b=(0,o.useMemo)(function(){return O.reduce(function(h,C){var p,l,x,N=d(C==null||(p=C.size)===null||p===void 0?void 0:p.total),Y=d(C==null||(l=C.size)===null||l===void 0?void 0:l.used);return h.total+=N,h.used+=Y,h.free+=(C==null||(x=C.size)===null||x===void 0?void 0:x.un_used)===void 0?Math.max(0,N-Y):d(C.size.un_used),h},{total:0,used:0,free:0})},[O]),V=b.total>0?b.used/b.total*100:0,G=b.free,E=d(j==null?void 0:j.load1)/Math.max(1,d(w==null?void 0:w.processor))*100,U=function(C,p){var l=d(C);return l>=95?H.colorError:l>=80?H.colorWarning:p},F="\u8D1F\u8F7D\u8F83\u4F4E\uFF0C\u7CFB\u7EDF\u8D44\u6E90\u5145\u8DB3";return E>=100?F="\u8D1F\u8F7D\u8FC7\u9AD8\uFF0C\u5EFA\u8BAE\u68C0\u67E5\u8FD0\u884C\u8FDB\u7A0B":E>=75?F="\u8D1F\u8F7D\u504F\u9AD8\uFF0C\u5EFA\u8BAE\u6301\u7EED\u5173\u6CE8":E>=40&&(F="\u8D1F\u8F7D\u6B63\u5E38\uFF0C\u7CFB\u7EDF\u8FD0\u884C\u5E73\u7A33"),(0,n.jsx)(A.A,{span:24,children:(0,n.jsxs)(z.A,{gutter:[Z,Z],children:[(0,n.jsx)(A.A,{xs:24,md:12,xl:6,children:(0,n.jsx)(rn.A,{className:L.metricCard,variant:"borderless",children:(0,n.jsx)(pe,{color:U(E,H.colorPrimary),icon:"LineChartOutlined",title:"\u7CFB\u7EDF\u8D1F\u8F7D",value:un(E),detail:"1m ".concat(d(j==null?void 0:j.load1).toFixed(2)," \xB7 5m ").concat(d(j==null?void 0:j.load5).toFixed(2)," \xB7 15m ").concat(d(j==null?void 0:j.load15).toFixed(2)),meta:F,percent:sn(E)})})}),(0,n.jsx)(A.A,{xs:24,md:12,xl:6,children:(0,n.jsx)(rn.A,{className:L.metricCard,variant:"borderless",children:(0,n.jsx)(pe,{color:U(w==null?void 0:w.usage,"#13c2c2"),icon:"DashboardOutlined",title:"CPU \u4F7F\u7528\u7387",value:un(w==null?void 0:w.usage),detail:(w==null?void 0:w.model)||"\u672A\u77E5\u5904\u7406\u5668",meta:"".concat(d(w==null?void 0:w.processor)," \u4E2A\u903B\u8F91\u5904\u7406\u5668"),percent:sn(w==null?void 0:w.usage)})})}),(0,n.jsx)(A.A,{xs:24,md:12,xl:6,children:(0,n.jsx)(rn.A,{className:L.metricCard,variant:"borderless",children:(0,n.jsx)(pe,{color:U(y==null?void 0:y.used_percent,"#52c41a"),icon:"DatabaseOutlined",title:"\u5185\u5B58\u4F7F\u7528\u7387",value:un(y==null?void 0:y.used_percent),detail:"".concat((0,J.X)(y==null?void 0:y.used)," / ").concat((0,J.X)(y==null?void 0:y.total)),meta:"".concat((0,J.X)(y==null?void 0:y.free)," \u7A7A\u95F2"),percent:sn(y==null?void 0:y.used_percent)})})}),(0,n.jsx)(A.A,{xs:24,md:12,xl:6,children:(0,n.jsx)(Ge.A,{placement:"bottom",trigger:"hover",title:(0,n.jsx)(k.hE,{size:"small",children:"\u78C1\u76D8\u7A7A\u95F4"}),content:(0,n.jsx)(Je,{diskFree:G,diskPercent:V,diskSummary:b,summaryDisks:O,sizeBase:r}),children:(0,n.jsx)(rn.A,{className:L.metricCard,variant:"borderless",children:(0,n.jsx)(pe,{color:U(V,"#5b8ff9"),icon:"PieChartOutlined",title:"\u78C1\u76D8\u4F7F\u7528\u7387",value:un(V),detail:"".concat((0,J.X)(b.used,r)," / ").concat((0,J.X)(b.total,r)),meta:"".concat(O.length," \u4E2A\u6302\u8F7D\u70B9"),percent:sn(V),detailTooltip:!1})})})})]})})}),tt=et,Ie,Fe=120,ke=30,rt=(0,$.rU)(function(i){var a=i.css;return{page:a(Ie||(Ie=u()([`
        width: 100%;
        max-width: 1450px;
        margin: 0 auto;
        font-variant-numeric: tabular-nums;

        > .ant-col {
            min-width: 0;
        }
    `])))}}),Le=(function(){var i=rt(),a=i.styles,t=(0,o.useRef)(!0),r=(0,o.useRef)(!1),s=(0,o.useRef)({statis:"",info:"",io:"",sampleId:0,sampleTime:0}),L=(0,o.useState)(function(){var F=Date.now();return{statis:void 0,info:void 0,io:void 0,throughputHistory:Array.from({length:ke},function(h,C){return{time:F-(ke-C)*1e3,network:{},disk:{},placeholder:!0}})}}),R=K()(L,2),H=R[0],I=R[1],w=(0,o.useState)(!0),y=K()(w,2),j=y[0],O=y[1],b=H.statis,V=H.info,G=H.io,E=H.throughputHistory,U=(0,o.useCallback)(Q()(W()().mark(function F(){var h,C,p,l,x,N,Y,kn,S,dn,An,v,M,tn,q,vn,pn,Nn,Pn,en,Bn,Rn,qn,jn,Cn,$n,bn,Un,Kn,zn=arguments;return W()().wrap(function(f){for(;;)switch(f.prev=f.next){case 0:if(h=zn.length>0&&zn[0]!==void 0?zn[0]:!1,!r.current){f.next=3;break}return f.abrupt("return");case 3:return r.current=!0,h&&O(!0),f.prev=5,f.next=8,Promise.all([(0,ln.iE)(),(0,ln.yV)(),(0,ln.si)({body:{after:s.current.sampleId}})]);case 8:if(C=f.sent,p=K()(C,3),l=p[0],x=p[1],N=p[2],t.current){f.next=15;break}return f.abrupt("return");case 15:Y=(l==null?void 0:l.code)===200&&l!==null&&l!==void 0&&l.data?l==null?void 0:l.data:void 0,kn=(x==null?void 0:x.code)===200&&x!==null&&x!==void 0&&x.data?x==null?void 0:x.data:void 0,S=(N==null?void 0:N.code)===200&&N!==null&&N!==void 0&&N.data?N==null?void 0:N.data:void 0,dn=S?_()(_()({},S),{},{history:void 0}):void 0,An=JSON.stringify(Y||null),v=JSON.stringify(kn||null),M=JSON.stringify(dn||null),tn=An!==s.current.statis,q=v!==s.current.info,vn=M!==s.current.io,pn=d(S==null?void 0:S.sample_id),Nn=pn>0&&pn<s.current.sampleId,Pn=Nn?0:s.current.sampleId,en=Nn?0:s.current.sampleTime,Bn=Array.isArray(S==null?void 0:S.history)?S==null?void 0:S.history:[],Rn=d(S==null?void 0:S.sample_time),qn=Bn.length===0&&Rn>0&&(Pn===0||Rn>en)?[S]:[],jn=[].concat(D()(Bn),qn).filter(function(P){var nn=d(P==null?void 0:P.sample_id);return nn>0?nn>Pn||nn===Pn&&d(P==null?void 0:P.sample_time)>en:d(P==null?void 0:P.sample_time)>en}).sort(function(P,nn){var Dn=d(P==null?void 0:P.sample_id)-d(nn==null?void 0:nn.sample_id);return Dn||d(P==null?void 0:P.sample_time)-d(nn==null?void 0:nn.sample_time)}),Cn=jn.map(function(P){return{sampleId:d(P==null?void 0:P.sample_id),time:d(P==null?void 0:P.sample_time),network:Object.fromEntries(Object.entries((P==null?void 0:P.network)||{}).map(function(nn){var Dn=K()(nn,2),In=Dn[0],on=Dn[1];return[In,{receiveRate:d(on==null?void 0:on.recv_rate),sendRate:d(on==null?void 0:on.send_rate)}]})),disk:Object.fromEntries(Object.entries((P==null?void 0:P.disk)||{}).map(function(nn){var Dn=K()(nn,2),In=Dn[0],on=Dn[1];return[In,{readRate:d(on==null?void 0:on.read_bytes_rate),writeRate:d(on==null?void 0:on.write_bytes_rate)}]}))}}),$n=Math.max.apply(Math,[Pn,pn].concat(D()(Cn.map(function(P){return P.sampleId})))),bn=Math.max.apply(Math,[en,d(S==null?void 0:S.sample_time)].concat(D()(Cn.map(function(P){return P.time})))),Un=Cn.length>0,(tn||q||vn||Un)&&(s.current={statis:An,info:v,io:M,sampleId:$n,sampleTime:bn},Kn=function(){return I(function(nn){var Dn=nn.throughputHistory;if(Cn.length>0){var In,on=new Map;Nn||nn.throughputHistory.forEach(function(Tn){Tn.placeholder||on.set(Tn.time,Tn)}),Cn.forEach(function(Tn){return on.set(Tn.time,Tn)});var je=Array.from(on.values()).sort(function(Tn,Ce){return Tn.time-Ce.time}).slice(-Fe),We=Math.max(0,ke-je.length),it=((In=je[0])===null||In===void 0?void 0:In.time)||Date.now(),at=Array.from({length:We},function(Tn,Ce){return{time:it-(We-Ce)*1e3,network:{},disk:{},placeholder:!0}});Dn=[].concat(at,D()(je)).slice(-Fe)}return{statis:tn&&Y?Y:nn.statis,info:q&&kn?kn:nn.info,io:vn&&dn?dn:nn.io,throughputHistory:Dn}})},h?Kn():o.startTransition(Kn)),f.next=42;break;case 40:f.prev=40,f.t0=f.catch(5);case 42:return f.prev=42,r.current=!1,t.current&&h&&O(!1),f.finish(42);case 46:case"end":return f.stop()}},F,null,[[5,40,42,46]])})),[]);return(0,o.useEffect)(function(){t.current=!0;var F=!1,h=0,C=Date.now()+1e3,p=function x(){var N=Math.max(0,C-Date.now());h=window.setTimeout(Q()(W()().mark(function Y(){return W()().wrap(function(S){for(;;)switch(S.prev=S.next){case 0:if(!F){S.next=2;break}return S.abrupt("return");case 2:if(document.visibilityState!=="visible"){S.next=5;break}return S.next=5,U();case 5:C+=1e3,C<Date.now()-1e3&&(C=Date.now()),x();case 8:case"end":return S.stop()}},Y)})),N)};U(!0).finally(function(){F||(C=Date.now()+1e3,p())});var l=function(){document.visibilityState==="visible"&&(C=Date.now()+1e3,U())};return document.addEventListener("visibilitychange",l),function(){F=!0,t.current=!1,window.clearTimeout(h),document.removeEventListener("visibilitychange",l)}},[U]),(0,n.jsx)(k.nB,{loading:j,children:(0,n.jsxs)(z.A,{className:a.page,gutter:[0,Z],children:[(0,n.jsx)(A.A,{span:24,children:(0,n.jsx)(ge,{statis:b,info:V})}),(0,n.jsx)(tt,{statis:b,diskSizeBase:(V==null?void 0:V.os)==="darwin"?1e3:1024}),(0,n.jsx)(A.A,{span:24,children:(0,n.jsx)(Ve,{io:G,throughputHistory:E})})]})})});e.dn(Le)}),2904:(function(c,g,e){"use strict";e.d(g,{P:function(){return m},X:function(){return W}});function m(){var T=arguments.length>0&&arguments[0]!==void 0?arguments[0]:16,D="abcdefghijklmnopqrstuvwxyz";D+=D.toUpperCase(),D+="0123456789";for(var B="",_=0;_<T;_++){var X=Math.floor(Math.random()*D.length);B+=D[X]}return B}var W=function(D){var B=arguments.length>1&&arguments[1]!==void 0?arguments[1]:1024,_=Number(D||0);if(_<=0)return"0 B";for(var X=["B","KB","MB","GB","TB"],Q=_,an=0;Q>=B&&an<X.length-1;)Q=Q/B,an+=1;return"".concat(Q.toFixed(an===0?0:2)," ").concat(X[an])}}),27755:(function(c,g,e){"use strict";var m=e(65437),W=e(71685);const T=D=>(0,m.AZ)(D)?(0,W.Tn)(D)?D():D:null;e.d(g,["b",0,T])}),35381:(function(c,g,e){"use strict";var m=e(96540),W=e(67735),T=e(65437),D=e(34164),B=e(27755),_=e(30366),X=e(62279),Q=e(92563);const an=u=>{const{title:o,content:z,prefixCls:A,classNames:$,styles:k}=u;return!(0,T.AZ)(o)&&!(0,T.AZ)(z)?null:m.createElement(m.Fragment,null,(0,T.AZ)(o)&&m.createElement("div",{className:(0,D.$)(`${A}-title`,$==null?void 0:$.title),style:k==null?void 0:k.title},o),(0,T.AZ)(z)&&m.createElement("div",{className:(0,D.$)(`${A}-content`,$==null?void 0:$.content),style:k==null?void 0:k.content},z))},K=u=>{const{hashId:o,prefixCls:z,className:A,style:$,placement:k="top",title:ln,content:Z,children:d,classNames:sn,styles:un}=u,xn=(0,B.b)(ln),mn=(0,B.b)(Z),fn=Sn(On({},u),{placement:k}),[gn,rn]=(0,_.AV)([sn],[un],{props:fn}),wn=(0,D.$)(o,z,`${z}-pure`,`${z}-placement-${k}`,A);return m.createElement("div",{className:wn,style:$},m.createElement("div",{className:`${z}-arrow`}),m.createElement(W.z,Sn(On({},u),{className:o,prefixCls:z,classNames:gn,styles:rn}),d||m.createElement(an,{prefixCls:z,title:xn,content:mn,classNames:gn,styles:rn})))},cn=u=>{const d=u,{prefixCls:o,className:z}=d,A=De(d,["prefixCls","className"]),{getPrefixCls:$}=m.useContext(X.QO),k=$("popover",o),[ln,Z]=(0,Q.A)(k);return m.createElement(K,Sn(On({},A),{prefixCls:k,hashId:ln,className:(0,D.$)(z,Z)}))};g.Ay=cn,e.d(g,["hJ",0,an,"xn",0,K])}),28073:(function(c,g,e){"use strict";var m=e(96540),W=e(65437),T=e(34164),D=e(27755),B=e(30366),_=e(23723),X=e(62279),Q=e(88826),an=e(48875),K=e(35381),cn=e(92563);const o=m.forwardRef((z,A)=>{var Qn,Zn,Jn;const ue=z,{prefixCls:$,title:k,content:ln,overlayClassName:Z,placement:d="top",trigger:sn,children:un,mouseEnterDelay:xn,mouseLeaveDelay:mn,onOpenChange:fn,overlayStyle:gn={},styles:rn,classNames:wn,motion:yn,arrow:_n}=ue,ve=De(ue,["prefixCls","title","content","overlayClassName","placement","trigger","children","mouseEnterDelay","mouseLeaveDelay","onOpenChange","overlayStyle","styles","classNames","motion","arrow"]),{getPrefixCls:ne,className:he,style:ee,classNames:J,styles:n,arrow:te,trigger:re,mouseEnterDelay:oe,mouseLeaveDelay:ie}=(0,X.TP)("popover"),Hn=(Qn=xn!=null?xn:oe)!=null?Qn:.1,Xn=(Zn=mn!=null?mn:ie)!=null?Zn:.1,Fn=ne("popover",$),[ae,se]=(0,cn.A)(Fn),le=ne(),de=(0,an.A)(_n,te),ce=sn||re||"hover",xe=Sn(On({},z),{placement:d,trigger:ce,mouseEnterDelay:Hn,mouseLeaveDelay:Xn,overlayStyle:gn,styles:rn,classNames:wn}),fe=(0,B.XT)(ee),ge=(0,B.XT)(gn),[En,Ln]=(0,B.AV)([J,wn],[n,fe,rn,ge],{props:xe}),ye=(0,T.$)(Z,ae,se,he,En.root),[Wn,Mn]=(0,W.PW)((Jn=z.defaultOpen)!=null?Jn:!1,z.open),be=Yn=>{Mn(Yn),fn==null||fn(Yn)},Vn=(0,D.b)(k),Gn=(0,D.b)(ln);return m.createElement(Q.A,Sn(On({unique:!1,arrow:de,placement:d,trigger:ce,mouseEnterDelay:Hn,mouseLeaveDelay:Xn},ve),{prefixCls:Fn,classNames:{root:ye,container:En.container,arrow:En.arrow},styles:{root:Ln.root,container:Ln.container,arrow:Ln.arrow},ref:A,open:Wn,onOpenChange:be,overlay:(0,W.AZ)(Vn)||(0,W.AZ)(Gn)?m.createElement(K.hJ,{prefixCls:Fn,title:Vn,content:Gn,classNames:En,styles:Ln}):null,motion:{motionName:(0,_.b)(le,"zoom-big",typeof(yn==null?void 0:yn.motionName)=="string"?yn==null?void 0:yn.motionName:void 0)},"data-popover-inject":!0}),un)});o._InternalPanelDoNotUseOrYouWillBeFired=K.Ay,g.A=o}),92563:(function(c,g,e){"use strict";var m=e(25905),W=e(99077),T=e(95201),D=e(20791),B=e(13950),_=e(11214),X=e(37358);const Q="50%",an=u=>{const{componentCls:o,popoverColor:z,titleMinWidth:A,fontWeightStrong:$,innerPadding:k,dropShadowPopover:ln,colorTextHeading:Z,borderRadiusLG:d,zIndexPopup:sn,titleMarginBottom:un,colorBgElevated:xn,popoverBg:mn,titleBorderBottom:fn,innerContentPadding:gn,titlePadding:rn,antCls:wn}=u,[yn,_n]=(0,X.PN)(wn,"tooltip");return[{[o]:Sn(On({},(0,m.dF)(u)),{position:"absolute",top:0,left:{_skip_check_:!0,value:0},zIndex:sn,fontWeight:400,whiteSpace:"normal",textAlign:"start",cursor:"auto",userSelect:"text",filter:ln,[yn("valid-offset-x")]:_n("arrow-offset-x","var(--arrow-x)"),transformOrigin:[_n("valid-offset-x",Q),`var(--arrow-y, ${Q})`].join(" "),[yn("arrow-background-color")]:xn,width:"max-content",maxWidth:"100vw","&-rtl":{direction:"rtl"},"&-hidden":{display:"none"},[`${o}-content`]:{position:"relative"},[`${o}-container`]:{backgroundColor:mn,backgroundClip:"padding-box",borderRadius:d,padding:k},[`${o}-title`]:{minWidth:A,marginBottom:un,color:Z,fontWeight:$,borderBottom:fn,padding:rn},[`${o}-content`]:{color:z,padding:gn}})},(0,T.A)(u,_n("arrow-background-color"),{arrowShadow:!1}),{[`${o}-pure`]:{position:"relative",maxWidth:"none",margin:u.sizePopupArrow,display:"inline-block"}}]},K=u=>{const{componentCls:o,antCls:z}=u,[A]=(0,X.PN)(z,"tooltip");return{[o]:B.s.map($=>{const k=u[`${$}6`];return{[`&${o}-${$}`]:{[A("arrow-background-color")]:k,[`${o}-container`]:{backgroundColor:k},[`${o}-arrow`]:{background:"transparent"}}}})}},cn=u=>{const{lineWidth:o,controlHeight:z,fontHeight:A,padding:$,wireframe:k,zIndexPopupBase:ln,borderRadiusLG:Z,marginXS:d,lineType:sn,colorSplit:un,paddingSM:xn}=u,mn=z-A,fn=mn/2,gn=mn/2-o,rn=$;return Sn(On(On({titleMinWidth:177,zIndexPopup:ln+30},(0,D.n)(u)),(0,T.K)({contentRadius:Z,limitVerticalRadius:!0})),{innerPadding:k?0:12,titleMarginBottom:k?0:d,titlePadding:k?`${fn}px ${rn}px ${gn}px`:0,titleBorderBottom:k?`${o}px ${sn} ${un}`:"none",innerContentPadding:k?`${xn}px ${rn}px`:0})};g.A=(0,X.OF)("Popover",u=>{const{colorBgElevated:o,colorText:z}=u,A=(0,_.oX)(u,{popoverBg:o,popoverColor:z});return[an(A),K(A),(0,W.aB)(A,"zoom-big")]},cn,{resetStyle:!1,deprecatedTokens:[["width","titleMinWidth"],["minWidth","titleMinWidth"]]})}),41162:(function(c,g,e){var m=e(44814);function W(T){if(Array.isArray(T))return m(T)}c.exports=W,c.exports.__esModule=!0,c.exports.default=c.exports}),59224:(function(c,g,e){var m=e(34911);function W(T,D){var B=typeof Symbol!="undefined"&&T[Symbol.iterator]||T["@@iterator"];if(!B){if(Array.isArray(T)||(B=m(T))||D&&T&&typeof T.length=="number"){B&&(T=B);var _=0,X=function(){};return{s:X,n:function(){return _>=T.length?{done:!0}:{done:!1,value:T[_++]}},e:function(u){throw u},f:X}}throw new TypeError(`Invalid attempt to iterate non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}var Q=!0,an=!1,K;return{s:function(){B=B.call(T)},n:function(){var u=B.next();return Q=u.done,u},e:function(u){an=!0,K=u},f:function(){try{!Q&&B.return!=null&&B.return()}finally{if(an)throw K}}}}c.exports=W,c.exports.__esModule=!0,c.exports.default=c.exports}),33032:(function(c){function g(e){if(typeof Symbol!="undefined"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}c.exports=g,c.exports.__esModule=!0,c.exports.default=c.exports}),30922:(function(c){function g(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}c.exports=g,c.exports.__esModule=!0,c.exports.default=c.exports}),84169:(function(c){function g(e,m){return m||(m=e.slice(0)),Object.freeze(Object.defineProperties(e,{raw:{value:Object.freeze(m)}}))}c.exports=g,c.exports.__esModule=!0,c.exports.default=c.exports}),45647:(function(c,g,e){var m=e(41162),W=e(33032),T=e(34911),D=e(30922);function B(_){return m(_)||W(_)||T(_)||D()}c.exports=B,c.exports.__esModule=!0,c.exports.default=c.exports}),39874:(function(c,g,e){"use strict";e.d(g,{A:function(){return D}});var m=e(53954),W=e(52176),T=e(56822);function D(B,_,X){return _=(0,m.A)(_),(0,T.A)(B,(0,W.A)()?Reflect.construct(_,X||[],(0,m.A)(B).constructor):_.apply(B,X))}})}]);
}());