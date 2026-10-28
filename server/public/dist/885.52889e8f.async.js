"use strict";(self.webpackChunk=self.webpackChunk||[]).push([[885],{96885:(function(_e,ke,p){p.d(ke,{A:function(){return tn}});var we=p(55420),c=p.n(we),Ce=p(98312),j=p.n(Ce),z=p(11604),$=p.n(z),ie=p(84169),P=p.n(ie),r=p(96540),le=p(88826),X=p(50400),Fe=p(45478),ee=p(49216),Se=p(12185),V=p(62885),Te=p(45647),W=p.n(Te),Ee=p(1974),D=p.n(Ee),je=p(79273),ue=p(34118),Ae=p(14499),Y=300,S=1e4,e=S*5,o=1e3,T={lineHeight:20,overscan:24,viewportHeight:520},a=function(E){var h=Number(E);return Number.isFinite(h)&&h>=0?h:0},en=function(){var E=arguments.length>0&&arguments[0]!==void 0?arguments[0]:Ae.on,h=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{},N=arguments.length>2?arguments[2]:void 0,Re=je.A.useApp(),se=Re.modal,Le=(0,r.useCallback)(function(){return(N==null?void 0:N.current)||document.body},[N]),fe=ue.Ay.useMessage({getContainer:Le}),pe=$()(fe,2),B=pe[0],H=pe[1],K=(0,r.useRef)(null),A=(0,r.useRef)(void 0),U=(0,r.useRef)(null),O=(0,r.useRef)({}),ne=(0,r.useRef)(h),M=(0,r.useRef)(h);ne.current=h;var R=(0,r.useRef)({requestId:0,requesting:!1,historyLoading:!1,cursor:0,startCursor:0,lineCount:0,stickToBottom:!0,historyAnchor:null,scrollFrame:0,historyFrame:0}),me=(0,r.useRef)(j()(c()().mark(function n(){return c()().wrap(function(t){for(;;)switch(t.prev=t.next){case 0:return t.abrupt("return",void 0);case 1:case"end":return t.stop()}},n)}))),ge=(0,r.useState)(),re=$()(ge,2),J=re[0],f=re[1],xe=(0,r.useState)(""),he=$()(xe,2),oe=he[0],ce=he[1],k=(0,r.useState)([]),y=$()(k,2),x=y[0],q=y[1],m=(0,r.useRef)(!1),an=(0,r.useState)(!1),We=$()(an,2),ln=We[0],de=We[1],un=(0,r.useState)(!1),Je=$()(un,2),Ie=Je[0],be=Je[1],sn=(0,r.useState)(0),Ze=$()(sn,2),Qe=Ze[0],ye=Ze[1],on=(0,r.useState)(T.viewportHeight),Xe=$()(on,2),cn=Xe[0],dn=Xe[1],ve=x.length>S,vn=Math.max(0,Math.floor(Qe/T.lineHeight)-T.overscan),fn=Math.min(x.length,Math.ceil((Qe+cn)/T.lineHeight)+T.overscan),pn=x.length*T.lineHeight,te=(0,r.useCallback)(function(){var n=K.current;n!=null&&n.clientHeight&&(dn(n.clientHeight),R.current.stickToBottom&&(n.scrollTop=n.scrollHeight),ye(n.scrollTop))},[]),mn=(0,r.useCallback)(function(n){var u;if((u=A.current)===null||u===void 0||u.call(A),A.current=void 0,K.current=n,!!n)if(te(),typeof ResizeObserver=="undefined")window.addEventListener("resize",te),A.current=function(){return window.removeEventListener("resize",te)};else{var t=new ResizeObserver(te);t.observe(n),A.current=function(){return t.disconnect()}}},[te]),Z=(0,r.useCallback)(function(){var n;(n=U.current)===null||n===void 0||n.destroy(),U.current=null},[]),Q=(0,r.useCallback)(function(){var n=R.current;n.cursor=0,n.startCursor=0,n.lineCount=0,n.stickToBottom=!0,n.historyAnchor=null,n.scrollFrame&&(cancelAnimationFrame(n.scrollFrame),n.scrollFrame=0),n.historyFrame&&(cancelAnimationFrame(n.historyFrame),n.historyFrame=0),ye(0)},[]),ae=(0,r.useCallback)((function(){var n=j()(c()().mark(function u(t){var g,i,w,l,d,L,s,I=arguments;return c()().wrap(function(F){for(;;)switch(F.prev=F.next){case 0:if(i=I.length>1&&I[1]!==void 0?I[1]:!1,w=t==null?void 0:t.id,l=R.current,!(w==null||String((g=O.current)===null||g===void 0?void 0:g.id)!==String(w)||i&&l.requesting)){F.next=5;break}return F.abrupt("return",null);case 5:return d=++l.requestId,l.requesting=!0,i||de(!0),F.prev=8,F.next=11,E({body:D()(D()({},M.current),t)});case 11:if(s=F.sent,!(l.requestId!==d||String((L=O.current)===null||L===void 0?void 0:L.id)!==String(w))){F.next=14;break}return F.abrupt("return",null);case 14:if((s==null?void 0:s.code)===200){F.next=17;break}return i||B.error((s==null?void 0:s.message)||"\u83B7\u53D6\u65E5\u5FD7\u5931\u8D25"),F.abrupt("return",null);case 17:return F.abrupt("return",(s==null?void 0:s.data)||{});case 20:return F.prev=20,F.t0=F.catch(8),l.requestId===d&&!i&&B.error("\u83B7\u53D6\u65E5\u5FD7\u5931\u8D25"),F.abrupt("return",null);case 24:return F.prev=24,l.requestId===d&&(l.requesting=!1,de(!1)),F.finish(24);case 27:case"end":return F.stop()}},u,null,[[8,20,24,27]])}));return function(u){return n.apply(this,arguments)}})(),[B,E]),_=(0,r.useCallback)((function(){var n=j()(c()().mark(function u(t){var g,i,w,l,d,L,s=arguments;return c()().wrap(function(C){for(;;)switch(C.prev=C.next){case 0:return g=s.length>1&&s[1]!==void 0?s[1]:!1,C.next=3,ae(D()({id:t,page_size:Y},g?{after:0}:{}),g);case 3:if(i=C.sent,i){C.next=6;break}return C.abrupt("return");case 6:w=Array.isArray(i==null?void 0:i.lines)?i==null?void 0:i.lines:[],Q(),l=R.current,l.cursor=a(i==null?void 0:i.cursor),l.startCursor=a(i==null?void 0:i.start_cursor),l.lineCount=w.length,d=!!(i!=null&&i.has_previous),m.current=d,L=function(){ce(String((i==null?void 0:i.file_name)||"")),q(w)},g?(0,r.startTransition)(L):L();case 16:case"end":return C.stop()}},u)}));return function(u){return n.apply(this,arguments)}})(),[ae,Q]),He=(0,r.useCallback)((function(){var n=j()(c()().mark(function u(t){var g,i,w,l,d;return c()().wrap(function(s){for(;;)switch(s.prev=s.next){case 0:return g=R.current,s.next=3,ae({id:t,after:g.cursor,page_size:Y},!0);case 3:if(i=s.sent,i){s.next=6;break}return s.abrupt("return");case 6:if(w=g.cursor,l=a(i==null?void 0:i.cursor),d=Array.isArray(i==null?void 0:i.lines)?i==null?void 0:i.lines:[],!(l<w)){s.next=13;break}return s.next=12,_(t,!0);case 12:return s.abrupt("return");case 13:if(d.length!==0){s.next=16;break}return g.cursor=l,s.abrupt("return");case 16:if(!(g.lineCount+d.length>e)){s.next=20;break}return s.next=19,_(t,!0);case 19:return s.abrupt("return");case 20:g.cursor=l,g.lineCount+=d.length,(0,r.startTransition)(function(){return q(function(I){return[].concat(W()(I),W()(d))})});case 23:case"end":return s.stop()}},u)}));return function(u){return n.apply(this,arguments)}})(),[_,ae]),Oe=(0,r.useCallback)(j()(c()().mark(function n(){var u,t,g,i,w,l,d,L,s,I,C;return c()().wrap(function(v){for(;;)switch(v.prev=v.next){case 0:if(t=R.current,g=(u=O.current)===null||u===void 0?void 0:u.id,!(!g||!m.current||t.historyLoading||t.requesting)){v.next=4;break}return v.abrupt("return");case 4:return i=K.current,i&&(t.historyAnchor={scrollTop:i.scrollTop,scrollHeight:i.scrollHeight}),t.historyLoading=!0,w=!1,v.prev=8,v.next=11,ae({id:g,before:t.startCursor,page_size:Y},!0);case 11:if(l=v.sent,l){v.next=15;break}return t.historyAnchor=null,v.abrupt("return");case 15:if(d=Array.isArray(l==null?void 0:l.lines)?l==null?void 0:l.lines:[],L=t.startCursor,s=a(l==null?void 0:l.start_cursor),I=!(l!=null&&l.end)&&!!(l!=null&&l.has_previous)&&s<L,t.startCursor=s,d.length!==0){v.next=25;break}return t.historyAnchor=null,m.current=I,w=I,v.abrupt("return");case 25:if(!(t.lineCount+d.length>e)){v.next=29;break}return t.historyAnchor=null,m.current=!1,v.abrupt("return");case 29:t.lineCount+=d.length,q(function(Me){return[].concat(W()(d),W()(Me))}),m.current=I;case 32:return v.prev=32,t.historyLoading=!1,w&&!t.historyFrame&&(C=function Me(){t.historyFrame=0;var Ye=K.current;if(!(!Ye||Ye.scrollTop>64||!m.current)){if(t.requesting||t.historyLoading){t.historyFrame=requestAnimationFrame(Me);return}me.current()}},t.historyFrame=requestAnimationFrame(C)),v.finish(32);case 36:case"end":return v.stop()}},n,null,[[8,,32,36]])})),[ae]);me.current=Oe;var gn=(0,r.useCallback)(function(){var n=K.current;if(n){var u=R.current;u.stickToBottom=n.scrollHeight-n.scrollTop-n.clientHeight<=24,ve&&!u.scrollFrame&&(u.scrollFrame=requestAnimationFrame(function(){u.scrollFrame=0,ye(n.scrollTop)})),n.scrollTop<=64&&Oe()}},[Oe,ve]);(0,r.useEffect)(function(){if(J!=null){var n=!1,u=0,t=function i(){u=window.setTimeout(j()(c()().mark(function w(){return c()().wrap(function(d){for(;;)switch(d.prev=d.next){case 0:if(!n){d.next=2;break}return d.abrupt("return");case 2:if(document.visibilityState!=="visible"){d.next=5;break}return d.next=5,He(J);case 5:n||i();case 6:case"end":return d.stop()}},w)})),o)},g=function(){document.visibilityState==="visible"&&He(J)};return t(),document.addEventListener("visibilitychange",g),function(){n=!0,window.clearTimeout(u),document.removeEventListener("visibilitychange",g)}}},[He,J]),(0,r.useLayoutEffect)(function(){var n=K.current,u=R.current;if(n){if(u.historyAnchor){var t=u.historyAnchor;n.scrollTop=t.scrollTop+n.scrollHeight-t.scrollHeight,u.historyAnchor=null}else x.length>0&&u.stickToBottom&&(n.scrollTop=n.scrollHeight);ve&&ye(n.scrollTop)}},[x,ve]),(0,r.useEffect)(function(){return function(){Z();var n=R.current;n.requestId+=1,O.current={},n.scrollFrame&&cancelAnimationFrame(n.scrollFrame),n.historyFrame&&cancelAnimationFrame(n.historyFrame)}},[Z]);var hn=(0,r.useCallback)(function(n,u){if((n==null?void 0:n.id)===void 0||(n==null?void 0:n.id)===null)return!1;Z();var t=R.current;return t.requestId+=1,t.requesting=!1,t.historyLoading=!1,Q(),O.current=n,M.current=D()(D()({},ne.current),u||{}),m.current=!1,f(n.id),ce(""),q([]),de(!1),be(!1),_(n.id),!0},[Z,_,Q]),bn=(0,r.useCallback)(function(){Z();var n=R.current;n.requestId+=1,n.requesting=!1,n.historyLoading=!1,Q(),O.current={},M.current=ne.current,m.current=!1,f(void 0),ce(""),q([]),de(!1),be(!1)},[Z,Q]),yn=(0,r.useCallback)(function(){Z(),R.current.requestId+=1,O.current={},f(void 0)},[Z]),kn=(0,r.useCallback)(function(){var n,u=(n=O.current)===null||n===void 0?void 0:n.id;u!=null&&_(u)},[_]),wn=(0,r.useCallback)(function(){var n=O.current;if(!((n==null?void 0:n.id)===void 0||(n==null?void 0:n.id)===null||Ie||U.current)){var u=se.confirm({getContainer:function(){var g=N==null?void 0:N.current;return g&&document.fullscreenElement===g?g:document.body},title:"\u6E05\u7406\u65E5\u5FD7",content:"\u786E\u5B9A\u6E05\u7406\u201C".concat(n.name||"-","\u201D\u7684\u5168\u90E8\u65E5\u5FD7\u5417\uFF1F\u6E05\u7406\u540E\u65E0\u6CD5\u6062\u590D\u3002"),okText:"\u6E05\u7406",cancelText:"\u53D6\u6D88",okButtonProps:{danger:!0},mask:{closable:!0},afterClose:function(){U.current===u&&(U.current=null)},onOk:(function(){var t=j()(c()().mark(function i(){var w,l,d,L,s,I,C;return c()().wrap(function(v){for(;;)switch(v.prev=v.next){case 0:if(l=n.id,String((w=O.current)===null||w===void 0?void 0:w.id)===String(l)){v.next=3;break}return v.abrupt("return");case 3:return d=R.current,L=++d.requestId,d.requesting=!0,de(!1),be(!0),s="task-log-clear-".concat(String(l)),B.loading({key:s,content:"\u6B63\u5728\u6E05\u7406\u65E5\u5FD7...",duration:0}),v.prev=10,v.next=13,E({body:D()(D()({},M.current),{},{id:l,action:"clear"})});case 13:if(C=v.sent,!(d.requestId!==L||String((I=O.current)===null||I===void 0?void 0:I.id)!==String(l))){v.next=16;break}return v.abrupt("return");case 16:if((C==null?void 0:C.code)===200){v.next=19;break}return B.error((C==null?void 0:C.message)||"\u6E05\u7406\u65E5\u5FD7\u5931\u8D25"),v.abrupt("return");case 19:Q(),q([]),m.current=!1,B.success((C==null?void 0:C.message)||"\u6E05\u7406\u6210\u529F"),v.next=28;break;case 25:v.prev=25,v.t0=v.catch(10),d.requestId===L&&B.error("\u6E05\u7406\u65E5\u5FD7\u5931\u8D25");case 28:return v.prev=28,B.destroy(s),d.requestId===L&&(d.requesting=!1,be(!1)),v.finish(28);case 32:case"end":return v.stop()}},i,null,[[10,25,28,32]])}));function g(){return t.apply(this,arguments)}return g})()});U.current=u}},[Ie,B,se,N,E,Q]);return{setConsole:mn,message:B,messageHolder:H,measureViewport:te,fileName:oe,logs:x,loading:ln,clearing:Ie,virtual:ve,virtualStart:vn,virtualEnd:fn,virtualHeight:pn,open:hn,reset:bn,pause:yn,refresh:kn,clear:wn,handleScroll:gn}},nn=en,b=p(74848),$e,Pe,Be,ze,De,Ne,Ue,qe,Ve,Ge,rn=(0,Se.rU)(function(G){var E=G.css,h=G.token;return{modal:E($e||($e=P()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
        }
    `]))),panel:E(Pe||(Pe=P()([`
        position: relative;
        overflow: hidden;
        border: 1px solid `,`;
        border-radius: `,`px;
        background: `,`;

        &:fullscreen {
            width: 100%;
            height: 100%;
            padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left);
            box-sizing: border-box;
            display: flex;
            flex-direction: column;
            border: 0;
            border-radius: 0;

            .task-log-console {
                min-height: 0;
                height: auto;
                flex: 1;
            }
        }
    `])),h.colorBorderSecondary,h.borderRadiusLG,h.colorBgContainer),panelHeader:E(Be||(Be=P()([`
        height: 42px;
        padding: 0 9px;
        display: flex;
        flex: none;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        box-sizing: border-box;
        background: `,`;
        box-shadow: inset 0 -1px 0 `,`;
    `])),h.colorBgContainer,h.colorSplit),panelTitle:E(ze||(ze=P()([`
        min-width: 0;
        padding-left: 5px;
        overflow: hidden;
        color: `,`;
        font-size: `,`px;
        font-weight: 500;
        line-height: 20px;
        text-overflow: ellipsis;
        white-space: nowrap;
    `])),h.colorTextTertiary,h.fontSizeSM),console:E(De||(De=P()([`
        position: relative;
        height: min(480px, calc(100vh - 280px));
        min-height: 300px;
        overflow: auto;
        scrollbar-gutter: stable;
        background: `,`;

        &::-webkit-scrollbar {
            width: 8px !important;
            height: 8px !important;
        }

        &::-webkit-scrollbar-track {
            background: transparent;
        }

        &::-webkit-scrollbar-thumb {
            border: 2px solid `,`;
            border-radius: 8px;
            background: `,`;
        }

        &::-webkit-scrollbar-thumb:hover {
            background: `,`;
        }

        @media (max-width: 560px) {
            height: min(440px, calc(100vh - 250px));
            min-height: 260px;
        }
    `])),h.colorBgContainer,h.colorBgContainer,h.colorFillSecondary,h.colorFill),content:E(Ne||(Ne=P()([`
        width: max-content;
        min-width: 100%;
        min-height: 100%;
        margin: 0;
        box-sizing: border-box;
        color: `,`;
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
        font-size: 12px;
        line-height: `,`px;
        white-space: pre;
        overflow-wrap: normal;
        word-break: normal;
        tab-size: 4;
    `])),h.colorTextSecondary,T.lineHeight),virtualContent:E(Ue||(Ue=P()([`
        position: relative;
        width: max-content;
        min-width: 100%;
        box-sizing: border-box;
        color: `,`;
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
        font-size: 12px;
        line-height: `,`px;
        tab-size: 4;
    `])),h.colorTextSecondary,T.lineHeight),virtualLine:E(qe||(qe=P()([`
        position: absolute;
        left: 0;
        width: max-content;
        min-width: 100%;
        height: `,`px;
        color: inherit;
        font: inherit;
        line-height: `,`px;
        white-space: pre;
        overflow-wrap: normal;
        word-break: normal;
    `])),T.lineHeight,T.lineHeight),actions:E(Ve||(Ve=P()([`
        display: flex;
        flex: none;
        align-items: center;
        justify-content: flex-end;
        gap: 2px;

        .ant-btn {
            width: 30px;
            height: 30px;
            padding: 0;
            display: inline-flex;
            align-items: center;
            justify-content: center;
        }

        .ant-btn .anticon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            line-height: 1;
        }

    `]))),state:E(Ge||(Ge=P()([`
        min-width: 100%;
        min-height: 100%;
        box-sizing: border-box;
        display: flex;
        align-items: center;
        justify-content: center;

        .ant-empty-description {
            color: `,`;
        }
    `])),h.colorTextTertiary)}}),Ke=(0,r.forwardRef)(function(G,E){var h=G.request,N=G.showClear,Re=N===void 0?!0:N,se=G.title,Le=se===void 0?"\u4EFB\u52A1\u65E5\u5FD7":se,fe=G.body,pe=fe===void 0?{}:fe,B=rn(),H=B.styles,K=(0,r.useRef)({}),A=(0,r.useRef)(null),U=(0,r.useRef)(!1),O=(0,r.useState)(!1),ne=$()(O,2),M=ne[0],R=ne[1],me=(0,r.useState)(!1),ge=$()(me,2),re=ge[0],J=ge[1],f=nn(h,pe,A),xe=(0,r.useCallback)(function(){var k=A.current;k&&document.fullscreenElement===k&&typeof document.exitFullscreen=="function"&&document.exitFullscreen().catch(function(){}),R(!1)},[]),he=(0,r.useCallback)(function(k){if(!k){var y=A.current;y&&document.fullscreenElement===y&&typeof document.exitFullscreen=="function"&&document.exitFullscreen().catch(function(){})}A.current=k},[]),oe=(0,r.useCallback)(function(){var k=A.current;return k&&document.fullscreenElement===k?k:document.body},[]);(0,r.useEffect)(function(){var k=function(){R(!!(A.current&&document.fullscreenElement===A.current)),f.measureViewport()};return document.addEventListener("fullscreenchange",k),function(){return document.removeEventListener("fullscreenchange",k)}},[f.measureViewport]);var ce=(0,r.useCallback)(j()(c()().mark(function k(){var y,x;return c()().wrap(function(m){for(;;)switch(m.prev=m.next){case 0:if(y=A.current,!(!y||re)){m.next=3;break}return m.abrupt("return");case 3:if(!(!document.fullscreenEnabled||typeof y.requestFullscreen!="function"||typeof document.exitFullscreen!="function")){m.next=6;break}return f.message.info("\u5F53\u524D\u6D4F\u89C8\u5668\u4E0D\u652F\u6301\u5168\u5C4F"),m.abrupt("return");case 6:if(x=document.fullscreenElement===y,J(!0),m.prev=8,!x){m.next=14;break}return m.next=12,document.exitFullscreen();case 12:m.next=16;break;case 14:return m.next=16,y.requestFullscreen();case 16:if(!((!U.current||A.current!==y)&&document.fullscreenElement===y)){m.next=19;break}return m.next=19,document.exitFullscreen();case 19:m.next=24;break;case 21:m.prev=21,m.t0=m.catch(8),U.current&&A.current===y&&f.message.error(x?"\u9000\u51FA\u5168\u5C4F\u5931\u8D25":"\u8FDB\u5165\u5168\u5C4F\u5931\u8D25");case 24:return m.prev=24,J(!1),m.finish(24);case 27:case"end":return m.stop()}},k,null,[[8,21,24,27]])})),[re,f.message]);return(0,r.useImperativeHandle)(E,function(){return{open:function(y,x){if(f.open(y,x)){var q;U.current=!0,(q=K.current)===null||q===void 0||q.show()}}}},[f.open]),(0,b.jsx)(V.Ht,{ref:K,title:(0,b.jsx)(V.hE,{children:Le}),width:900,afterOpenChange:function(y){y||(U.current=!1,xe(),f.pause())},modalProps:{rootClassName:H.modal,footer:null,keyboard:!M,zIndex:1200,style:{top:100,paddingBottom:100},mask:{closable:!0},afterOpenChange:function(y){y&&f.measureViewport()},afterClose:f.reset},children:(0,b.jsxs)("div",{ref:he,className:H.panel,children:[f.messageHolder,(0,b.jsxs)("div",{className:H.panelHeader,children:[(0,b.jsx)("span",{className:H.panelTitle,children:f.fileName||"\u65E5\u5FD7\u6587\u4EF6"}),(0,b.jsxs)("div",{className:H.actions,children:[(0,b.jsx)(le.A,{title:M?"\u9000\u51FA\u5168\u5C4F":"\u5168\u5C4F",placement:M?"bottom":"top",getPopupContainer:oe,children:(0,b.jsx)(X.A,{color:"default",variant:"text","aria-label":M?"\u9000\u51FA\u5168\u5C4F":"\u5168\u5C4F","aria-pressed":M,icon:(0,b.jsx)(V.In,{type:M?"FullscreenExitOutlined":"FullscreenOutlined"}),disabled:re,onClick:function(){ce()}})}),(0,b.jsx)(le.A,{title:"\u5237\u65B0\u65E5\u5FD7",placement:M?"bottom":"top",getPopupContainer:oe,children:(0,b.jsx)(X.A,{color:"default",variant:"text","aria-label":"\u5237\u65B0\u65E5\u5FD7",icon:(0,b.jsx)(V.In,{type:"ReloadOutlined"}),loading:f.loading,disabled:f.clearing,onClick:f.refresh})}),Re&&(0,b.jsx)(le.A,{title:"\u6E05\u7406\u65E5\u5FD7",placement:M?"bottom":"top",getPopupContainer:oe,children:(0,b.jsx)(X.A,{color:"danger",variant:"text","aria-label":"\u6E05\u7406\u65E5\u5FD7",icon:(0,b.jsx)(V.In,{type:"DeleteOutlined"}),loading:f.clearing,onClick:f.clear})})]})]}),(0,b.jsx)("div",{ref:f.setConsole,className:"".concat(H.console," task-log-console"),onScroll:f.handleScroll,children:f.logs.length>0?f.virtual?(0,b.jsx)("div",{className:H.virtualContent,style:{height:f.virtualHeight},children:f.logs.slice(f.virtualStart,f.virtualEnd).map(function(k,y){var x=f.virtualStart+y;return(0,b.jsx)("div",{className:H.virtualLine,style:{top:x*T.lineHeight},children:k},x)})}):(0,b.jsx)("pre",{className:H.content,children:f.logs.join(`
`)}):f.loading?(0,b.jsx)("div",{className:H.state,children:(0,b.jsx)(Fe.A,{})}):(0,b.jsx)("div",{className:H.state,children:(0,b.jsx)(ee.A,{image:ee.A.PRESENTED_IMAGE_SIMPLE,description:"\u6682\u65E0\u6570\u636E"})})})]})})});Ke.displayName="Log";var tn=(0,r.memo)(Ke)}),14499:(function(_e,ke,p){p.d(ke,{Mg:function(){return P},S8:function(){return je},UT:function(){return Fe},ZZ:function(){return Ee},ec:function(){return Ae},lC:function(){return Se},on:function(){return le},u3:function(){return $},vq:function(){return Te}});var we=p(55420),c=p.n(we),Ce=p(98312),j=p.n(Ce),z=p(78781);function $(){return ie.apply(this,arguments)}function ie(){return ie=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.Jt)("/task/list",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),ie.apply(this,arguments)}function P(){return r.apply(this,arguments)}function r(){return r=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.Jt)("/task/info",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),r.apply(this,arguments)}function le(){return X.apply(this,arguments)}function X(){return X=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.bE)("/task/log",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),X.apply(this,arguments)}function Fe(){return ee.apply(this,arguments)}function ee(){return ee=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.bE)("/task/create",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),ee.apply(this,arguments)}function Se(){return V.apply(this,arguments)}function V(){return V=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.bE)("/task/update",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),V.apply(this,arguments)}function Te(){return W.apply(this,arguments)}function W(){return W=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.bE)("/task/delete",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),W.apply(this,arguments)}function Ee(){return D.apply(this,arguments)}function D(){return D=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.bE)("/task/run",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),D.apply(this,arguments)}function je(){return ue.apply(this,arguments)}function ue(){return ue=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.bE)("/task/stop",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),ue.apply(this,arguments)}function Ae(){return Y.apply(this,arguments)}function Y(){return Y=j()(c()().mark(function S(){var e,o=arguments;return c()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return e=o.length>0&&o[0]!==void 0?o[0]:{},a.abrupt("return",(0,z.bE)("/task/restore",e==null?void 0:e.body,e==null?void 0:e.onSuccess,e==null?void 0:e.onFail,e==null?void 0:e.onFinally));case 2:case"end":return a.stop()}},S)})),Y.apply(this,arguments)}})}]);
