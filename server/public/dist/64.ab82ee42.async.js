"use strict";(self.webpackChunk=self.webpackChunk||[]).push([[64],{67064:(function(nn,ce,c){c.d(ce,{A:function(){return an}});var W=c(1974),$=c.n(W),se=c(45647),K=c.n(se),Z=c(11604),o=c.n(Z),J=c(71640),X=c.n(J),s=c(96540),w=c(79273),Oe=c(40372),F=c(82591),H=c(22817),we=c(88826),q=c(50400),ze=c(55270),de=c(20873),Ie=c(90124),M=c(62885),pe=c(41100),d=c(11834),ie=c(49216),S=c(58679),v=c(44960),l=c(24816),m=c(91695),i=c(74848),C=function(a){var O=a.className,t=a.compact,R=a.mode,z=a.path,le=a.loadedPath,h=a.selectedFile,T=a.disks,ve=a.selectedDisk,ee=a.pathBreadcrumbs,me=a.parentPath,p=a.diskRoot,e=a.canGoUp,r=a.editingPath,y=a.pathDraft,n=a.loading,E=a.error,ne=a.items,N=a.page,_=a.pageSize,P=a.total,he=a.pathRef,ge=a.pathInputRef,Je=a.currentPathRef,G=a.tableRef,I=a.onNavigate,_e=a.onOpenPathEditor,je=a.onCancelPathEditor,He=a.onSubmitPathEditor,Ge=a.onPathDraftChange,Ne=a.onSelectFile,Le=a.onConfirm,te=a.onPageChange,Ye=a.onReload,ln=(0,s.useMemo)(function(){return[{title:"\u540D\u79F0",dataIndex:"name",render:function(x,j){var re=(0,d.e_)(z,x),oe=!j.is_dir&&(0,d.Mt)((h==null?void 0:h.path)||"")===(0,d.Mt)(re);return(0,i.jsxs)("div",{className:"file-picker-name-cell",children:[(0,i.jsx)(M.In,{type:(0,d.Ky)(x,j.is_dir),className:"file-picker-icon"}),(0,i.jsx)("span",{className:"file-picker-name",title:x,children:x}),!j.is_dir&&R==="file"?(0,i.jsx)("span",{className:"file-picker-row-action",children:oe&&(0,i.jsx)(M.In,{type:"CheckOutlined"})}):null]})}},{title:"\u5927\u5C0F",dataIndex:"size",width:t?80:96,render:function(x,j){return j.is_dir?"-":(0,d.v7)(x)}},{title:"\u4FEE\u6539\u65F6\u95F4",dataIndex:"update_time",width:t?132:154,render:d.Zc}]},[t,R,z,h==null?void 0:h.path]),Qe=(0,s.useMemo)(function(){return T.map(function(b){return{label:b,value:b}})},[T]),g=(0,s.useCallback)(function(b){if(n)return"file-picker-passive";var x=(0,d.e_)(z,b.name),j=!b.is_dir&&(0,d.Mt)((h==null?void 0:h.path)||"")===(0,d.Mt)(x);return j?"file-picker-selectable file-picker-selected":b.is_dir?"file-picker-directory":R==="file"?"file-picker-selectable":"file-picker-passive"},[n,R,z,h==null?void 0:h.path]),be=(0,s.useCallback)(function(b){if(n)return{"aria-disabled":!0};var x=(0,d.e_)(z,b.name);if(b.is_dir)return{role:"button",tabIndex:0,"aria-label":"\u6253\u5F00\u76EE\u5F55 ".concat(b.name),title:"\u8FDB\u5165 ".concat(b.name),onClick:function(){return I(x)},onKeyDown:function(Y){(Y.key==="Enter"||Y.key===" ")&&(Y.preventDefault(),I(x))}};if(R!=="file")return{title:b.name};var j=(0,d.Mt)((h==null?void 0:h.path)||"")===(0,d.Mt)(x),re={name:b.name,path:x};return{role:"button",tabIndex:0,"aria-label":"\u9009\u62E9\u6587\u4EF6 ".concat(b.name),"aria-pressed":j,title:"\u9009\u62E9 ".concat(b.name),onClick:function(){return Ne(re)},onDoubleClick:function(){return Le(re)},onKeyDown:function(Y){(Y.key==="Enter"||Y.key===" ")&&(Y.preventDefault(),Ne(re))}}},[n,R,Le,I,Ne,z,h==null?void 0:h.path]),on=E?(0,i.jsxs)("div",{className:"file-picker-error",children:[(0,i.jsx)(ie.A,{image:ie.A.PRESENTED_IMAGE_SIMPLE,description:E}),(0,i.jsx)(q.A,{size:"small",onClick:Ye,children:"\u91CD\u8BD5"})]}):(0,i.jsx)(ie.A,{image:ie.A.PRESENTED_IMAGE_SIMPLE,description:"\u6682\u65E0\u6570\u636E"}),Ue=R==="directory"?z:h==null?void 0:h.path,ye=R==="directory"?"\u5F53\u524D\u76EE\u5F55":"\u9009\u62E9\u7ED3\u679C",xe=Ue||(R==="directory"?"\u6B63\u5728\u8BFB\u53D6\u78C1\u76D8...":"\u672A\u9009\u62E9"),un=R==="directory"?"FolderOpenOutlined":h?(0,d.Ky)(h.name):"FileOutlined",Ve=P>0?(N-1)*_+1:0,Ce=Math.min(N*_,P),De=n&&!le?"\u52A0\u8F7D\u4E2D":P>_?"".concat(Ve,"-").concat(Ce," / ").concat(P," \u9879"):"".concat(P," \u9879");return(0,i.jsxs)("div",{className:O,children:[(0,i.jsxs)("div",{className:"file-picker-toolbar".concat(T.length>1?" file-picker-toolbar-with-disks":""),children:[(0,i.jsx)(we.A,{title:"\u4E0A\u4E00\u7EA7",children:(0,i.jsx)(q.A,{type:"text",className:"file-picker-up",disabled:!e||n,"aria-label":"\u8FD4\u56DE\u4E0A\u4E00\u7EA7",icon:(0,i.jsx)(M.In,{type:"ArrowLeftOutlined"}),onClick:function(){return I((0,d.ru)(me,p)?me:p)}})}),r?(0,i.jsx)(H.A,{ref:ge,className:"file-picker-path-input",value:y,"aria-label":"\u7F16\u8F91\u5F53\u524D\u8DEF\u5F84",onChange:function(x){return Ge(x.target.value)},onBlur:function(){return je()},onKeyDown:function(x){x.nativeEvent.isComposing||x.nativeEvent.keyCode===229||(x.key==="Enter"?(x.preventDefault(),He()):x.key==="Escape"&&(x.preventDefault(),x.stopPropagation(),je(!0)))}}):(0,i.jsx)("ol",{ref:he,className:"file-picker-path","aria-label":"\u5F53\u524D\u8DEF\u5F84\uFF1A".concat(z||"\u6B63\u5728\u8BFB\u53D6"),children:ee.length>0?ee.map(function(b,x){var j=x===ee.length-1;return(0,i.jsxs)("li",{children:[x>0&&(0,i.jsx)(M.In,{type:"RightOutlined",className:"file-picker-path-separator","aria-hidden":"true"}),j?(0,i.jsx)("button",{ref:Je,className:"file-picker-path-current",type:"button",title:"\u70B9\u51FB\u7F16\u8F91\u8DEF\u5F84",disabled:n,"aria-current":"page","aria-label":"\u7F16\u8F91\u8DEF\u5F84 ".concat(b.path),onClick:_e,children:(0,i.jsx)("span",{children:b.label})}):(0,i.jsx)("button",{className:"file-picker-path-button",type:"button",title:b.path,disabled:n,"aria-label":"\u524D\u5F80 ".concat(b.path),onClick:function(){return I(b.path)},children:(0,i.jsx)("span",{children:b.label})})]},b.path)}):(0,i.jsx)("li",{children:(0,i.jsx)("span",{className:"file-picker-path-current",children:"\u6B63\u5728\u8BFB\u53D6\u78C1\u76D8..."})})}),T.length>1&&(0,i.jsx)(S.A,{className:"file-picker-disks",value:ve,"aria-label":"\u5207\u6362\u78C1\u76D8",placeholder:"\u5207\u6362\u78C1\u76D8",disabled:n,options:Qe,onChange:I})]}),(0,i.jsx)(v.A,{ref:G,className:"file-picker-table",rowKey:"name",size:"small",rowHoverable:!1,tableLayout:"auto",loading:n,columns:ln,dataSource:ne,locale:{emptyText:on},scroll:{x:"max-content",y:t?"clamp(180px, 34dvh, 238px)":"clamp(200px, 38dvh, 278px)"},pagination:!1,rowClassName:g,onRow:be}),(0,i.jsxs)("div",{className:"file-picker-status",children:[(0,i.jsxs)("div",{className:"file-picker-status-main",children:[(0,i.jsx)(M.In,{type:un,className:"file-picker-status-icon"}),(0,i.jsx)("span",{className:"file-picker-status-label",children:ye}),(0,i.jsx)(l.A.Text,{className:"file-picker-status-value",ellipsis:{tooltip:Ue},children:xe})]}),(0,i.jsxs)("div",{className:"file-picker-status-meta",children:[(0,i.jsx)("span",{className:"file-picker-count",children:De}),P>_&&(0,i.jsx)(m.A,{className:"file-picker-pagination",size:"small",simple:{readOnly:!0},current:N,pageSize:_,total:P,disabled:n,showSizeChanger:!1,showTitle:!1,onChange:te})]})]})]})},B=(0,s.memo)(C),fe=c(84169),A=c.n(fe),ae=c(12185),Se,Ke,Pe,tn=(0,ae.rU)(function(D,a){var O=D.css,t=D.token;return{modal:O(Se||(Se=A()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
        }

        .ant-modal-content {
            max-height: calc(100dvh - 24px);
            overflow: hidden;
            display: flex;
            flex-direction: column;
        }

        .ant-modal-body {
            min-height: 0;
            min-width: 0;
            overflow-y: auto;
        }

    `]))),field:O(Ke||(Ke=A()([`
        width: 100%;

        .ant-input {
            min-width: 0;
        }

        .ant-btn-color-default:not(:disabled) .ant-btn-icon {
            color: `,`;
        }
    `])),t.colorTextSecondary),browser:O(Pe||(Pe=A()([`
        min-width: 0;
        overflow: hidden;
        border: 1px solid `,`;
        border-radius: `,`px;
        background: `,`;

        .file-picker-toolbar {
            display: grid;
            min-width: 0;
            grid-template-columns: auto minmax(0, 1fr) auto;
            align-items: center;
            gap: `,`px;
            padding: `,`px;
            border-bottom: 1px solid `,`;
            background: `,`;
        }

        .file-picker-toolbar:not(.file-picker-toolbar-with-disks) {
            grid-template-columns: auto minmax(0, 1fr);
        }

        .file-picker-up {
            width: `,`px;
            min-width: `,`px;
            height: `,`px;
            padding: 0;
            border: 1px solid transparent;
            border-radius: `,`px;
            background: `,`;
            color: `,`;
        }

        .file-picker-up:not(:disabled):hover {
            border-color: `,`;
            background: `,`;
            color: `,`;
        }

        .file-picker-path {
            box-sizing: border-box;
            height: `,`px;
            min-width: 0;
            margin: 0;
            padding: 0 `,`px;
            display: flex;
            align-items: center;
            overflow-x: auto;
            overflow-y: hidden;
            border: 1px solid `,`;
            border-radius: `,`px;
            background: `,`;
            list-style: none;
            scrollbar-width: none;

            &::-webkit-scrollbar {
                display: none;
            }
        }

        .file-picker-path li {
            min-width: 0;
            display: inline-flex;
            flex: none;
            align-items: center;
        }

        .file-picker-path-separator {
            margin-inline: `,`px;
            color: `,`;
            font-size: `,`px;
        }

        .file-picker-path-button,
        .file-picker-path-current {
            max-width: `,`px;
            min-width: 0;
            padding: 0 `,`px;
            display: inline-flex;
            align-items: center;
            border: 0;
            border-radius: `,`px;
            outline: none;
            appearance: none;
            background: transparent;
            color: `,`;
            font: inherit;
            font-size: `,`px;
            line-height: `,`px;
            white-space: nowrap;
            cursor: pointer;
            transition: background-color `,", color ",`;
        }

        .file-picker-path-button > span,
        .file-picker-path-current > span {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .file-picker-path-button:hover {
            background: `,`;
            color: `,`;
        }

        .file-picker-path-button:disabled {
            color: `,`;
            cursor: not-allowed;
        }

        .file-picker-path-button:focus-visible {
            box-shadow: 0 0 0 2px `,`;
        }

        .file-picker-path-current {
            color: `,`;
            cursor: text;
            font-weight: 500;
        }

        .file-picker-path-current:not(:disabled):hover {
            background: `,`;
            color: `,`;
        }

        .file-picker-path-current:focus-visible {
            color: `,`;
            box-shadow: 0 0 0 2px `,`;
        }

        .file-picker-path-current:disabled {
            color: `,`;
            cursor: default;
        }

        .file-picker-path-input.ant-input {
            width: min(100%, `,`px);
            min-width: 40px;
            max-width: 100%;
            height: `,`px;
            padding: 0 `,`px;
            justify-self: start;
            border-color: `,`;
            border-radius: `,`px;
            background: `,`;
            color: `,`;
            box-shadow: 0 0 0 2px `,`;
            font-size: `,`px;
        }

        .file-picker-disks {
            width: `,`px;
        }

        .file-picker-table .ant-table,
        .file-picker-table .ant-table-container {
            background: transparent;
        }

        .file-picker-table .ant-table-thead > tr > th {
            padding: `,` !important;
            border-bottom: 1px solid `,`;
            background: `,`;
            color: `,`;
            font-size: `,`px;
            font-weight: 500;
        }

        .file-picker-table .ant-table-tbody > tr > td {
            padding: `,` !important;
            border-bottom: 1px solid `,`;
            background: `,`;
            font-size: `,`px;
            transition: background-color `,`;
        }

        .file-picker-table .ant-table-tbody > tr.ant-table-measure-row > td {
            padding-block: 0 !important;
            border-block: 0 !important;
        }

        .file-picker-table .ant-table-tbody > tr:last-child > td {
            border-bottom: 0;
        }

        .file-picker-table .ant-table-body {
            min-height: `,`;
            scrollbar-width: thin;
        }

        .file-picker-table .ant-table-placeholder td {
            height: `,`;
            border-bottom: 0;
        }

        .file-picker-name-cell {
            display: flex;
            width: max-content;
            align-items: center;
            gap: `,`px;
        }

        .file-picker-icon {
            flex: none;
            color: `,`;
            font-size: `,`px;
        }

        .file-picker-directory .file-picker-icon {
            color: `,`;
        }

        .file-picker-name {
            flex: none;
            color: `,`;
            white-space: nowrap;
        }

        .file-picker-row-action {
            width: `,`px;
            flex: none;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            color: `,`;
            font-size: `,`px;
        }

        .file-picker-directory,
        .file-picker-selectable {
            cursor: pointer;
        }

        .file-picker-directory:hover > td,
        .file-picker-selectable:hover > td {
            background: `,` !important;
        }

        .file-picker-directory:focus-visible > td,
        .file-picker-selectable:focus-visible > td {
            background: `,` !important;
        }

        .file-picker-directory:focus-visible > td:first-child,
        .file-picker-selectable:focus-visible > td:first-child {
            box-shadow: inset 3px 0 `,`;
        }

        .file-picker-selected > td {
            background: `,` !important;
        }

        .file-picker-selected:hover > td,
        .file-picker-selected:focus-visible > td {
            background: `,` !important;
        }

        .file-picker-selected .file-picker-row-action {
            color: `,`;
        }

        .file-picker-error {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
        }

        .file-picker-status {
            box-sizing: border-box;
            min-height: `,`px;
            padding: `,`;
            display: flex;
            min-width: 0;
            align-items: center;
            justify-content: space-between;
            gap: `,`px;
            border-top: 1px solid `,`;
            background: `,`;
        }

        .file-picker-status-main {
            min-width: 0;
            display: flex;
            flex: 1;
            align-items: center;
            gap: `,`px;
            overflow: hidden;
        }

        .file-picker-status-icon {
            flex: none;
            color: `,`;
        }

        .file-picker-status-label {
            flex: none;
            color: `,`;
            font-size: `,`px;
        }

        .file-picker-status-value {
            min-width: 0;
            flex: 1;
            display: block;
            overflow: hidden;
            color: `,`;
            font-size: `,`px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .file-picker-status-meta {
            min-width: 0;
            display: flex;
            flex: none;
            align-items: center;
            flex-wrap: nowrap;
            gap: `,`px;
            white-space: nowrap;
        }

        .file-picker-count {
            min-width: 0;
            overflow: hidden;
            color: `,`;
            font-size: `,`px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .file-picker-pagination {
            flex: none;
            margin: 0;
            color: `,`;
            font-size: `,`px;
        }

        .file-picker-pagination .ant-pagination-prev,
        .file-picker-pagination .ant-pagination-next {
            min-width: `,`px;
            height: `,`px;
            line-height: `,`px;
        }

        .file-picker-pagination .ant-pagination-item-link {
            border-radius: `,`px;
            color: `,`;
        }

        .file-picker-pagination .ant-pagination-prev:not(.ant-pagination-disabled):hover .ant-pagination-item-link,
        .file-picker-pagination .ant-pagination-next:not(.ant-pagination-disabled):hover .ant-pagination-item-link {
            background: `,`;
        }

        .file-picker-pagination .ant-pagination-simple-pager {
            height: `,`px;
            margin-inline: `,`px;
            color: `,`;
            line-height: `,`px;
        }

        @media (max-width: 520px) {
            .file-picker-disks {
                width: 104px;
            }

            .file-picker-status {
                min-height: 0;
                padding: 6px 9px 7px;
                align-items: stretch;
                flex-direction: column;
                gap: 4px;
            }

            .file-picker-status-main,
            .file-picker-status-meta {
                width: 100%;
                min-width: 0;
            }

            .file-picker-status-meta {
                justify-content: space-between;
            }
        }

        @media (pointer: coarse) {
            .file-picker-up {
                width: 40px;
                min-width: 40px;
                height: 40px;
            }

            .file-picker-path {
                height: 40px;
            }

            .file-picker-path-button,
            .file-picker-path-current {
                display: inline-flex;
                align-items: center;
            }

            .file-picker-path-button,
            .file-picker-path-current {
                min-height: 40px;
            }

            .file-picker-path-input.ant-input {
                height: 40px;
            }

            .file-picker-disks,
            .file-picker-disks .ant-select-selector {
                height: 40px !important;
            }

            .file-picker-disks .ant-select-selector {
                align-items: center;
            }

            .file-picker-table .ant-table-tbody > tr > td {
                padding-block: 10px !important;
            }
        }
    `])),t.colorBorderSecondary,t.borderRadius,t.colorBgContainer,a.compact?6:8,a.compact?7:9,t.colorSplit,t.colorFillQuaternary,a.compact?28:30,a.compact?28:30,a.compact?28:30,t.borderRadiusSM,t.colorBgContainer,t.colorTextSecondary,t.colorBorderSecondary,t.colorFillTertiary,t.colorText,a.compact?30:32,a.compact?6:8,t.colorBorderSecondary,t.borderRadiusSM,t.colorBgContainer,a.compact?2:3,t.colorTextQuaternary,a.compact?8:9,a.compact?150:190,a.compact?2:3,t.borderRadiusSM,t.colorTextSecondary,a.compact?11:12,a.compact?18:20,t.motionDurationFast,t.motionDurationFast,t.colorFillTertiary,t.colorText,t.colorTextDisabled,t.colorPrimaryBg,t.colorText,t.colorFillTertiary,t.colorPrimary,t.colorPrimary,t.colorPrimaryBg,t.colorText,a.compact?420:480,a.compact?30:32,a.compact?6:8,t.colorPrimaryBorder,t.borderRadiusSM,t.colorBgContainer,t.colorText,t.colorPrimaryBg,a.compact?11:12,a.compact?116:132,a.compact?"7px 10px":"9px 12px",t.colorSplit,t.colorBgContainer,t.colorTextTertiary,t.fontSizeSM,a.compact?"7px 10px":"9px 12px",t.colorSplit,t.colorBgContainer,t.fontSizeSM,t.motionDurationFast,a.compact?"clamp(180px, 34dvh, 238px)":"clamp(200px, 38dvh, 278px)",a.compact?"clamp(180px, 34dvh, 238px)":"clamp(200px, 38dvh, 278px)",a.compact?7:9,t.colorTextSecondary,a.compact?14:16,t.colorPrimary,t.colorText,a.compact?12:14,t.colorTextQuaternary,a.compact?11:12,t.colorFillQuaternary,t.colorFillQuaternary,t.colorPrimary,t.colorPrimaryBg,t.colorPrimaryBgHover,t.colorPrimary,a.compact?36:40,a.compact?"5px 9px":"6px 11px",a.compact?8:10,t.colorSplit,t.colorFillQuaternary,a.compact?6:8,t.colorTextTertiary,t.colorTextTertiary,t.fontSizeSM,t.colorTextSecondary,t.fontSizeSM,a.compact?7:9,t.colorTextTertiary,t.fontSizeSM,t.colorTextSecondary,t.fontSizeSM,a.compact?24:26,a.compact?24:26,a.compact?24:26,t.borderRadiusSM,t.colorTextSecondary,t.colorFillTertiary,a.compact?24:26,a.compact?5:7,t.colorTextTertiary,a.compact?24:26)}}),Be=tn,rn=["buttonProps","disabled","initialPath","mode","onChange","onSelect","placeholder","size","value"],Ee=(0,s.forwardRef)(function(D,a){var O,t=D.buttonProps,R=D.disabled,z=D.initialPath,le=D.mode,h=le===void 0?"directory":le,T=D.onChange,ve=D.onSelect,ee=D.placeholder,me=D.size,p=D.value,e=p===void 0?"":p,r=X()(D,rn),y=(0,M.DP)(),n=w.A.useApp(),E=n.message,ne=!!(!(y==null||(O=y.isCompactTheme)===null||O===void 0)&&O.call(y)),N=Oe.Ay.useBreakpoint(),_=Be({compact:ne}),P=_.styles,he=(0,Ie.A)("input",r.variant,r.bordered),ge=o()(he,1),Je=ge[0],G=(0,s.useRef)(0),I=(0,s.useRef)(0),_e=(0,s.useRef)(null),je=(0,s.useRef)(null),He=(0,s.useRef)(null),Ge=(0,s.useRef)(null),Ne=(0,s.useState)(!1),Le=o()(Ne,2),te=Le[0],Ye=Le[1],ln=(0,s.useState)(""),Qe=o()(ln,2),g=Qe[0],be=Qe[1],on=(0,s.useState)(""),Ue=o()(on,2),ye=Ue[0],xe=Ue[1],un=(0,s.useState)(),Ve=o()(un,2),Ce=Ve[0],De=Ve[1],b=(0,s.useState)([]),x=o()(b,2),j=x[0],re=x[1],oe=(0,s.useState)([]),Y=o()(oe,2),Pn=Y[0],$e=Y[1],En=(0,s.useState)(1),mn=o()(En,2),Ze=mn[0],Xe=mn[1],_n=(0,s.useState)(0),hn=o()(_n,2),jn=hn[0],qe=hn[1],Cn=(0,s.useState)(!1),gn=o()(Cn,2),Re=gn[0],Te=gn[1],Dn=(0,s.useState)(""),bn=o()(Dn,2),yn=bn[0],We=bn[1],Rn=(0,s.useState)(0),xn=o()(Rn,2),Tn=xn[0],Mn=xn[1],An=(0,s.useState)(!1),kn=o()(An,2),ke=kn[0],Fe=kn[1],On=(0,s.useState)(""),Fn=o()(On,2),cn=Fn[0],ue=Fn[1];(0,s.useEffect)(function(){return function(){G.current+=1,I.current+=1}},[]);var sn=h==="directory"?"\u9009\u62E9\u76EE\u5F55":"\u9009\u62E9\u6587\u4EF6",Q=(0,s.useMemo)(function(){var f;if(!g)return j[0]||"/";var u=j.filter(function(k){return(0,d.ru)(g,k)}).sort(function(k,V){return(0,d.Mt)(V).length-(0,d.Mt)(k).length});return u[0]||((f=(0,d.dY)(g)[0])===null||f===void 0?void 0:f.path)||"/"},[j,g]),zn=j.find(function(f){return(0,d.Mt)(f)===(0,d.Mt)(Q)}),wn=(0,s.useMemo)(function(){if(!g)return[];var f=(0,d.dY)(g),u=f.findIndex(function(k){return(0,d.Mt)(k.path)===(0,d.Mt)(Q)});return u<=0?f:[{label:Q,path:Q}].concat(K()(f.slice(u+1)))},[Q,g]),In=g?(0,d.W2)(g):Q,Bn=!!g&&(0,d.Mt)(g)!==(0,d.Mt)(Q),dn=(0,s.useCallback)(function(f){var u=(0,d.j7)(f);G.current+=1,I.current+=1,Te(!0),Fe(!1),ue(u),be(u),xe(""),De(void 0),Xe(1),We("")},[]),Nn=(0,s.useCallback)(function(){var f=String(e||"").trim(),u=(0,d.F6)(f)?(0,d.j7)(f):"",k=String(z||"").trim(),V=(0,d.F6)(k)?(0,d.j7)(k):"",L=V||(h==="file"&&u?(0,d.W2)(u):u),Me=u?(0,d.dY)(u):[],Ae=Me.length>1?Me[Me.length-1].label:"",Jn=++I.current;G.current+=1,be(L),xe(""),De(void 0),re([]),$e([]),qe(0),Xe(1),We(""),Te(!0),Fe(!1),ue(L),Ye(!0),h==="file"&&u&&Ae&&(0,pe.R8)({body:{path:u}}).then(function(U){var fn;if(I.current===Jn&&(U==null?void 0:U.code)===200&&U!==null&&U!==void 0&&U.data&&!(U!=null&&(fn=U.data)!==null&&fn!==void 0&&fn.is_dir)){var vn;De({name:(U==null||(vn=U.data)===null||vn===void 0?void 0:vn.name)||Ae,path:u})}}).catch(function(){})},[z,h,e]),pn=(0,s.useCallback)(function(){G.current+=1,I.current+=1,Te(!1),Fe(!1),ue(g),Ye(!1)},[g]),Sn=(0,s.useCallback)(function(){var f=arguments.length>0&&arguments[0]!==void 0?arguments[0]:Ce,u=h==="directory"?ye:f==null?void 0:f.path;if(u){var k=(0,d.j7)(u);if(ve){if(ve(k)===!1)return}else T==null||T(k);pn()}},[pn,ye,h,T,ve,Ce]),Ln=(0,s.useCallback)(function(f){I.current+=1,De(f)},[]),Un=(0,s.useCallback)(function(f){Te(!0),Xe(f),window.requestAnimationFrame(function(){var u;(u=Ge.current)===null||u===void 0||u.scrollTo({top:0})})},[]),$n=(0,s.useCallback)(function(){Re||!g||(ue(g),Fe(!0))},[Re,g]),en=(0,s.useCallback)(function(){var f=arguments.length>0&&arguments[0]!==void 0?arguments[0]:!1;ue(g),Fe(!1),f&&window.requestAnimationFrame(function(){var u;return(u=He.current)===null||u===void 0?void 0:u.focus()})},[g]),Wn=(0,s.useCallback)(function(){var f=cn.trim();if(!f){en(!0);return}if(!(0,d.F6)(f)){E.warning("\u8BF7\u8F93\u5165\u7EDD\u5BF9\u8DEF\u5F84");return}var u=(0,d.j7)(f),k=Array.from(new Set([].concat(K()(j),[Q])));if(!k.some(function(V){return(0,d.ru)(u,V)})){E.warning("\u8DEF\u5F84\u4E0D\u5728\u53EF\u7528\u78C1\u76D8\u4E2D");return}if((0,d.Mt)(u)===(0,d.Mt)(g)){en(!0);return}dn(u)},[en,Q,j,E,dn,g,cn]),Kn=(0,s.useCallback)(function(){Mn(function(f){return f+1})},[]);return(0,s.useEffect)(function(){if(te){var f=!0;return(0,pe.J0)().then(function(u){var k;if(f){if((u==null?void 0:u.code)!==200||!Array.isArray(u==null?void 0:u.data)){be(function(L){return L||"/"});return}var V=Array.from(new Set(u==null||(k=u.data)===null||k===void 0?void 0:k.filter(function(L){return typeof L=="string"&&L}).map(d.j7)));re(V),be(function(L){return L||V[0]||"/"})}}).catch(function(){f&&be(function(u){return u||"/"})}),function(){f=!1}}},[te]),(0,s.useLayoutEffect)(function(){te&&!ke&&_e.current&&(_e.current.scrollLeft=_e.current.scrollWidth)},[ke,te,wn]),(0,s.useEffect)(function(){Fe(!1),ue(g)},[g]),(0,s.useLayoutEffect)(function(){var f,u;ke&&((f=je.current)===null||f===void 0||f.focus(),(u=je.current)===null||u===void 0||u.select())},[ke]),(0,s.useEffect)(function(){Re&&ke&&(Fe(!1),ue(g))},[ke,Re,g]),(0,s.useEffect)(function(){if(!(!te||!g)){var f=++G.current;Te(!0),We(""),(0,pe.Jk)({body:{path:g,page:Ze,page_size:50,order_by_field:"name",order_by_type:"asc"}}).then(function(u){if(G.current===f){if((u==null?void 0:u.code)!==200){$e([]),qe(0),xe(""),We((u==null?void 0:u.message)||"\u83B7\u53D6\u6587\u4EF6\u5217\u8868\u5931\u8D25");return}var k=u==null?void 0:u.data,V=Array.isArray(k==null?void 0:k.list)?k==null?void 0:k.list.filter(function(Ae){return Ae==null?void 0:Ae.name}):[],L=Math.max(0,Number(k==null?void 0:k.total)||0),Me=Math.max(1,Math.ceil(L/50));if(qe(L),Ze>Me){$e([]),Xe(Me);return}$e(V),xe(g)}}).catch(function(){G.current===f&&($e([]),qe(0),xe(""),We("\u83B7\u53D6\u6587\u4EF6\u5217\u8868\u5931\u8D25"))}).finally(function(){G.current===f&&Te(!1)})}},[h,te,Ze,g,Tn]),(0,i.jsxs)(i.Fragment,{children:[(0,i.jsxs)(F.A.Compact,{className:P.field,block:!0,children:[(0,i.jsx)(H.A,$()($()({ref:a},r),{},{size:me,value:e,disabled:R,placeholder:ee||(h==="directory"?"\u8BF7\u8F93\u5165\u6216\u9009\u62E9\u76EE\u5F55":"\u8BF7\u8F93\u5165\u6216\u9009\u62E9\u6587\u4EF6"),onChange:function(u){return T==null?void 0:T(u.target.value)}})),(0,i.jsx)(we.A,{title:sn,children:(0,i.jsx)(q.A,$()($()({color:t!=null&&t.type||t!=null&&t.danger?void 0:"default",variant:t!=null&&t.type||t!=null&&t.danger?void 0:Je==="filled"?"filled":Je==="outlined"?"outlined":"text"},t),{},{size:me,disabled:R,"aria-label":sn,icon:(0,i.jsx)(M.In,{type:h==="directory"?"FolderOpenOutlined":"FileSearchOutlined"}),onClick:Nn}))})]}),(0,i.jsx)(ze.Ay,{componentSize:"middle",variant:"outlined",children:(0,i.jsx)(M.Ht,{title:(0,i.jsx)(M.hE,{children:sn}),width:"640px",okText:"\u9009\u62E9",onOk:function(){return Sn()},onCancel:pn,modalProps:{open:te,rootClassName:P.modal,style:{top:N.md?100:24,paddingBottom:N.md?100:24},cancelText:"\u53D6\u6D88",okButtonProps:{disabled:Re||!!yn||(h==="directory"?!ye||(0,d.Mt)(ye)!==(0,d.Mt)(g):!Ce)},focusable:{focusTriggerAfterClose:!1},mask:{closable:!0},styles:{body:{paddingTop:ne?10:15}}},children:(0,i.jsx)(de.A,{component:!1,size:ne?"small":"middle",variant:"outlined",children:(0,i.jsx)(B,{className:P.browser,compact:ne,mode:h,path:g,loadedPath:ye,selectedFile:Ce,disks:j,selectedDisk:zn,pathBreadcrumbs:wn,parentPath:In,diskRoot:Q,canGoUp:Bn,editingPath:ke,pathDraft:cn,loading:Re,error:yn,items:Pn,page:Ze,pageSize:50,total:jn,pathRef:_e,pathInputRef:je,currentPathRef:He,tableRef:Ge,onNavigate:dn,onOpenPathEditor:$n,onCancelPathEditor:en,onSubmitPathEditor:Wn,onPathDraftChange:ue,onSelectFile:Ln,onConfirm:Sn,onPageChange:Un,onReload:Kn})})})})]})});Ee.displayName="FilePicker";var an=(0,s.memo)(Ee)}),11834:(function(nn,ce,c){c.d(ce,{B$:function(){return ze},F6:function(){return M},Ky:function(){return d},Mt:function(){return F},W2:function(){return J},Zc:function(){return de},dY:function(){return X},e_:function(){return we},j7:function(){return o},qs:function(){return ie},ru:function(){return H},v7:function(){return q}});var W=c(11604),$=c.n(W),se=c(74353),K=c.n(se),Z=function(v){var l=String(v||""),m=l.match(/^([a-zA-Z]):(?:[\\/]+|$)/),i=m?l.replace(/\\/g,"/"):l,C=m?"".concat(m[1].toUpperCase(),":"):"",B=C?i.match(/^([a-zA-Z]):(?:\/+|$)/):null,fe=C?i.slice((B==null?void 0:B[0].length)||0):i.replace(/^\/+/,""),A=[];return fe.split(/\/+/).forEach(function(ae){if(!(!ae||ae===".")){if(ae===".."){A.pop();return}A.push(ae)}}),{drive:C,parts:A}},o=function(v){var l=Z(v),m=l.drive,i=l.parts,C=m?"".concat(m,"/"):"/";return i.length>0?"".concat(C).concat(i.join("/")):C},J=function(v){var l=Z(v),m=l.drive,i=l.parts;return i.pop(),o("".concat(m?"".concat(m,"/"):"/").concat(i.join("/")))},X=function(v){var l=Z(v),m=l.drive,i=l.parts,C=m?"".concat(m,"/"):"/",B=[{label:m||"/",path:C}];return i.forEach(function(fe,A){B.push({label:fe,path:o("".concat(C).concat(i.slice(0,A+1).join("/")))})}),B},s={archive:new Set(["7z","bz","bz2","gz","rar","tar","tgz","xz","zip"]),image:new Set(["avif","bmp","gif","ico","jpeg","jpg","png","svg","webp"]),video:new Set(["avi","flv","m4v","mkv","mov","mp4","mpeg","mpg","webm","wmv"]),audio:new Set(["aac","flac","m4a","mp3","ogg","wav","wma"]),pdf:new Set(["pdf"]),document:new Set(["doc","docx","odt","rtf"]),spreadsheet:new Set(["csv","ods","xls","xlsx"]),presentation:new Set(["odp","ppt","pptx"]),code:new Set(["c","cc","conf","cpp","css","env","go","h","hpp","html","ini","java","js","json","jsx","less","lua","php","properties","py","rb","rs","scss","sh","sql","toml","ts","tsx","vue","xml","yaml","yml"]),text:new Set(["log","md","text","txt"])},w=new Map;Object.entries(s).forEach(function(S){var v=$()(S,2),l=v[0],m=v[1];m.forEach(function(i){return w.set(i,l)})});var Oe={folder:"FolderOutlined",archive:"FileZipOutlined",image:"FileImageOutlined",video:"VideoCameraOutlined",audio:"AudioOutlined",pdf:"FilePdfOutlined",document:"FileWordOutlined",spreadsheet:"FileExcelOutlined",presentation:"FilePptOutlined",code:"CodeOutlined",text:"FileTextOutlined",file:"FileOutlined"},F=function(v){var l=o(v),m=l==="/"||/^[a-z]:\/$/i.test(l)?l:l.replace(/\/$/,"");return/^[a-z]:(?:\/|$)/i.test(m)?m.toLowerCase():m},H=function(v,l){var m=F(v),i=F(l);return m===i||m.startsWith(i.endsWith("/")?i:"".concat(i,"/"))},we=function(v,l){var m=o(v),i=String(l!=null?l:"");return"".concat(m).concat(m.endsWith("/")?"":"/").concat(i)},q=function(v){var l=Number(v||0);if(!Number.isFinite(l)||l<=0)return"0 B";for(var m=["B","KB","MB","GB","TB","PB"],i=l,C=0;i>=1024&&C<m.length-1;)i/=1024,C+=1;return"".concat(i.toLocaleString("zh-CN",{minimumFractionDigits:0,maximumFractionDigits:C===0?0:2})," ").concat(m[C])},ze=function(v){var l=Number(v);return!Number.isFinite(l)||l<0?"-":(Math.trunc(l)&511).toString(8).padStart(4,"0")},de=function(v){if(v==null||v==="")return"-";var l=Number(v),m=Number.isFinite(l)?l>1e12?K()(l):K().unix(l):K()(String(v));return m.isValid()?m.format("YYYY-MM-DD HH:mm:ss"):"-"},Ie=function(v){var l=String(v||"").toLocaleLowerCase(),m=l.lastIndexOf(".");return m>0&&m<l.length-1?l.slice(m+1):""},M=function(v){var l=String(v||"");return l.startsWith("/")||/^[a-zA-Z]:[\\/]/.test(l)},pe=function(v){var l=arguments.length>1&&arguments[1]!==void 0?arguments[1]:!1;return l?"folder":w.get(Ie(v))||"file"},d=function(v){var l=arguments.length>1&&arguments[1]!==void 0?arguments[1]:!1;return Oe[pe(v,l)]},ie=function(v){var l=String(v||"").toLowerCase();return l.endsWith(".zip")||l.endsWith(".tar")||l.endsWith(".gz")||l.endsWith(".gzip")||l.endsWith(".tgz")||l.endsWith(".tar.gz")}}),41100:(function(nn,ce,c){c.d(ce,{AD:function(){return tn},F9:function(){return ve},Hn:function(){return me},Io:function(){return v},J$:function(){return t},J0:function(){return we},Jk:function(){return ze},R8:function(){return Ie},UF:function(){return pe},Ww:function(){return an},Xi:function(){return Ke},_o:function(){return a},bp:function(){return rn},eR:function(){return C},iK:function(){return h},m3:function(){return fe},mV:function(){return ae},tO:function(){return z},x3:function(){return ie}});var W,$,se,K,Z=c(55420),o=c.n(Z),J=c(11604),X=c.n(J),s=c(98312),w=c.n(s),Oe=c(5309),F=c(78781),H=function(e,r){var y=new Error(r);return y.name=e,y};function we(){return q.apply(this,arguments)}function q(){return q=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.Jt)("/file/disk",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),q.apply(this,arguments)}function ze(){return de.apply(this,arguments)}function de(){return de=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.Jt)("/file/list",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),de.apply(this,arguments)}function Ie(){return M.apply(this,arguments)}function M(){return M=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.Jt)("/file/info",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),M.apply(this,arguments)}function pe(){return d.apply(this,arguments)}function d(){return d=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.Jt)("/file/sign",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),d.apply(this,arguments)}function ie(){return S.apply(this,arguments)}function S(){return S=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/create",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),S.apply(this,arguments)}function v(){return l.apply(this,arguments)}function l(){return l=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/update",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),l.apply(this,arguments)}function m(){return i.apply(this,arguments)}function i(){return i=$(W().mark(function p(){var e,r=arguments;return W().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",K("/file/upload",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally,void 0,{headers:se(),timeout:91e4}));case 2:case"end":return n.stop()}},p)})),i.apply(this,arguments)}function C(){return B.apply(this,arguments)}function B(){return B=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.Jt)("/file/count",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),B.apply(this,arguments)}function fe(){return A.apply(this,arguments)}function A(){return A=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/copy",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),A.apply(this,arguments)}function ae(){return Se.apply(this,arguments)}function Se(){return Se=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/move",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),Se.apply(this,arguments)}function Ke(){return Pe.apply(this,arguments)}function Pe(){return Pe=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/compress",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),Pe.apply(this,arguments)}function tn(){return Be.apply(this,arguments)}function Be(){return Be=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/extract",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),Be.apply(this,arguments)}function rn(){return Ee.apply(this,arguments)}function Ee(){return Ee=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/download",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),Ee.apply(this,arguments)}function an(){return D.apply(this,arguments)}function D(){return D=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/file/delete",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),D.apply(this,arguments)}function a(){return O.apply(this,arguments)}function O(){return O=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.Jt)("/recycle/list",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),O.apply(this,arguments)}function t(){return R.apply(this,arguments)}function R(){return R=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.Jt)("/recycle/count",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),R.apply(this,arguments)}function z(){return le.apply(this,arguments)}function le(){return le=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/recycle/restore",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),le.apply(this,arguments)}function h(){return T.apply(this,arguments)}function T(){return T=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/recycle/delete",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),T.apply(this,arguments)}function ve(){return ee.apply(this,arguments)}function ee(){return ee=w()(o()().mark(function p(){var e,r=arguments;return o()().wrap(function(n){for(;;)switch(n.prev=n.next){case 0:return e=r.length>0&&r[0]!==void 0?r[0]:{},n.abrupt("return",(0,F.bE)("/recycle/clear",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return n.stop()}},p)})),ee.apply(this,arguments)}function me(p,e,r){return new Promise(function(y,n){if(e.aborted){n(new DOMException("\u4E0A\u4F20\u5DF2\u53D6\u6D88","AbortError"));return}var E=new XMLHttpRequest,ne=new URL(p);Object.entries(r.query||{}).forEach(function(_){var P=X()(_,2),he=P[0],ge=P[1];return ne.searchParams.set(he,ge)});var N=function(){return E.abort()};E.open("POST",ne.toString()),E.timeout=900*1e3,Object.entries((0,Oe.dJ)()).forEach(function(_){var P=X()(_,2),he=P[0],ge=P[1];return E.setRequestHeader(he,ge)}),E.upload.onprogress=function(_){var P;return(P=r.onProgress)===null||P===void 0?void 0:P.call(r,_.loaded,_.total)},E.onload=function(){if(e.removeEventListener("abort",N),E.status<200||E.status>=300){n(H("FileUploadTransportError","\u4E0A\u4F20\u5931\u8D25\uFF08HTTP ".concat(E.status,"\uFF09")));return}try{var _=JSON.parse(E.responseText);if(!_||typeof(_==null?void 0:_.code)!="number"){n(H("FileUploadRequestError","\u670D\u52A1\u5668\u8FD4\u56DE\u7684\u6570\u636E\u683C\u5F0F\u4E0D\u6B63\u786E"));return}y(_)}catch(P){n(H("FileUploadRequestError","\u670D\u52A1\u5668\u8FD4\u56DE\u7684\u6570\u636E\u683C\u5F0F\u4E0D\u6B63\u786E"))}},E.onerror=function(){e.removeEventListener("abort",N),n(H("FileUploadTransportError","\u4E0A\u4F20\u6587\u4EF6\u5931\u8D25\uFF0C\u8BF7\u68C0\u67E5\u7F51\u7EDC\u8FDE\u63A5"))},E.ontimeout=function(){e.removeEventListener("abort",N),n(H("FileUploadTransportError","\u4E0A\u4F20\u6587\u4EF6\u8D85\u65F6"))},E.onabort=function(){e.removeEventListener("abort",N),n(new DOMException("\u4E0A\u4F20\u5DF2\u53D6\u6D88","AbortError"))},e.addEventListener("abort",N,{once:!0}),E.send(r.body)})}}),5309:(function(nn,ce,c){c.d(ce,{H0:function(){return Z},aY:function(){return $},dJ:function(){return o},h8:function(){return se}});var W="token",$="web";function se(J){localStorage.setItem(W,J)}function K(){return localStorage.getItem(W)||""}function Z(){localStorage.removeItem(W)}function o(){var J={device:$},X=K();return X&&(J.token=X),J}})}]);
