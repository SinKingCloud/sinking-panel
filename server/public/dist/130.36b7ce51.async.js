"use strict";(self.webpackChunk=self.webpackChunk||[]).push([[130],{17344:(function(U,v,n){n.r(v);var P=n(11604),i=n.n(P),b=n(96540),T=n(62885),D=n(58487),E=n(52661),d=n(74848),l=function(_){var m,u,p,r=_.id,x=_.open,h=_.onClose,t=(0,T.DP)(),M=!!(!(t==null||(m=t.isCompactTheme)===null||m===void 0)&&m.call(t)),g=!!(!(t==null||(u=t.isDarkMode)===null||u===void 0)&&u.call(t)||!(t==null||(p=t.isDarkTheme)===null||p===void 0)&&p.call(t)),a=(0,b.useState)(null),o=i()(a,2),e=o[0],s=o[1];return(0,d.jsx)(E.A,{title:"\u5BB9\u5668\u7EC8\u7AEF",open:x&&!!r,onClose:h,zIndex:1050,destroyOnHidden:!0,closable:!e,toolbarRef:s,children:x&&r&&(0,d.jsx)(D.A,{path:"/container/terminal",params:{id:r},title:"\u5BB9\u5668\u7EC8\u7AEF",compact:M,showHeader:!1,toolbarContainer:e,showStatus:!1,showConnect:!1,showClear:!0,showReconnect:!0,showFullscreen:!0,showClose:!0,terminalBackground:E.n,terminalAccent:g?"#c3ccd6":void 0,acknowledge:!0,autoConnect:!0,onClose:h},r)})};v.default=(0,b.memo)(l)}),52661:(function(U,v,n){n.d(v,{n:function(){return h}});var P=n(84169),i=n.n(P),b=n(96540),T=n(40372),D=n(45478),E=n(12185),d=n(62885),l=n(74848),j,_,m,u,p,r,x,h="rgb(15, 15, 15)",t=(0,E.rU)(function(g,a){var o=g.css,e=g.token,s=a.compact,B=s?"9px 14px":"10px 16px",C=s?"min(560px, calc(100dvh - 140px))":"min(640px, calc(100dvh - 170px))",O=s?"calc(100dvh - 88px)":"calc(100dvh - 100px)";return{modal:o(j||(j=i()([`
            .ant-modal {
                max-width: calc(100vw - 24px);
            }

            .ant-modal-container {
                padding: 0;
                overflow: hidden;
                background: `,`;
            }

            .ant-modal-header {
                margin: 0;
                padding: `,`;
                border-bottom: 1px solid `,`;
                background: `,`;
            }

            .ant-modal-title {
                min-width: 0;
            }
        `])),e.colorBgContainer,B,e.colorSplit,e.colorBgContainer),titleBar:o(_||(_=i()([`
            width: 100%;
            min-width: 0;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
        `]))),title:o(m||(m=i()([`
            flex: 1;
            min-width: 0;
            overflow: hidden;

            > * {
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        `]))),toolbar:o(u||(u=i()([`
            display: flex;
            flex: none;
            align-items: center;
            gap: 2px;

            .status-text {
                margin-right: 5px;
                color: `,`;
                font-size: 11px;
                font-weight: normal;
                white-space: nowrap;
            }

            .status-text.connected {
                color: `,`;
            }

            .status-text.connecting {
                color: `,`;
            }

            .status-text.error {
                color: `,`;
            }

            .ant-btn {
                width: `,`px;
                height: `,`px;
                padding: 0;
            }

            @media (max-width: 479px) {
                .status-text {
                    display: none;
                }
            }
        `])),e.colorTextTertiary,e.colorSuccessText,e.colorWarningText,e.colorErrorText,s?30:32,s?30:32),closeButton:o(p||(p=i()([`
            width: 30px;
            height: 30px;
            flex: none;
            padding: 0;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border: 0;
            border-radius: `,`px;
            background: transparent;
            color: `,`;
            cursor: pointer;

            &:hover,
            &:focus-visible {
                background: `,`;
                color: `,`;
            }

            &:focus-visible {
                outline: 2px solid `,`;
                outline-offset: 1px;
            }
        `])),e.borderRadiusSM,e.colorTextTertiary,e.colorFillTertiary,e.colorText,e.colorPrimaryBg),host:o(r||(r=i()([`
            height: `,`;
            padding: 0;
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            overflow: hidden;
            background: `,`;
            box-sizing: border-box;

            > * {
                min-height: 0;
                flex: 1;
                overflow: hidden;
                border-radius: 0;
            }

            @media (max-width: 767px) {
                height: `,`;
                padding: 0;
            }
        `])),C,h,O),loading:o(x||(x=i()([`
            width: 100%;
            height: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            background: `,`;
        `])),h)}}),M=function(a){var o,e=a.title,s=a.open,B=a.zIndex,C=a.onClose,O=a.destroyOnHidden,k=O===void 0?!1:O,y=a.loading,L=y===void 0?!1:y,w=a.closable,W=w===void 0?!0:w,A=a.toolbarRef,K=a.children,f=(0,d.DP)(),I=!!(!(f==null||(o=f.isCompactTheme)===null||o===void 0)&&o.call(f)),R=T.Ay.useBreakpoint(),H=t({compact:I}),c=H.styles;return(0,l.jsx)(d.Ht,{title:(0,l.jsxs)("div",{className:c.titleBar,children:[(0,l.jsx)("div",{className:c.title,children:(0,l.jsx)(d.hE,{size:"small",children:e})}),A&&(0,l.jsx)("div",{ref:A,className:c.toolbar}),W&&(0,l.jsx)("button",{className:c.closeButton,type:"button","aria-label":"\u5173\u95ED".concat(e),onClick:C,children:(0,l.jsx)(d.In,{type:"CloseOutlined"})})]}),width:I?"min(960px, calc(100vw - 24px))":"min(1040px, calc(100vw - 24px))",onCancel:C,modalProps:{open:s,zIndex:B,rootClassName:c.modal,destroyOnHidden:k,style:{top:R.md?100:24,paddingBottom:R.md?100:24},closable:!1,footer:null,mask:{closable:!0},styles:{container:{padding:0},body:{padding:0}}},children:(0,l.jsx)("div",{className:c.host,children:L?(0,l.jsx)("div",{className:c.loading,children:(0,l.jsx)(D.A,{size:"small"})}):K})})};v.A=(0,b.memo)(M)})}]);
