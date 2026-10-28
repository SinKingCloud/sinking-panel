"use strict";(self.webpackChunk=self.webpackChunk||[]).push([[868],{58487:(function(mn,Re,l){l.d(Re,{A:function(){return yn}});var Ie=l(55420),L=l.n(Ie),Le=l(98312),P=l.n(Le),We=l(11604),E=l.n(We),e=l(96540),je=l(40961),Ae=l(79273),W=l(59522),Oe=l(88826),O=l(50400),Z=l(55270),_=l(45478),b=l(62885),f=l(84169),H=l.n(f),Te=l(12185),x=l(96804),He=l(72480),o=l(74848),q,ve=l(97856).Terminal,hn=(0,Te.rU)(function(c,K){var T=c.css;return{terminal:T(q||(q=H()([`
        width: 100%;
        height: 100%;
        min-width: 0;
        min-height: 0;
        overflow: hidden;
        flex: 1;
        background: `,`;

        .xterm {
            height: 100%;
            padding: `,`px;
            box-sizing: border-box;
        }

        .xterm-viewport {
            background: `,` !important;
            scrollbar-width: thin;
            scrollbar-color: rgba(255, 255, 255, .18) transparent;
        }

        .xterm-viewport::-webkit-scrollbar {
            width: 8px !important;
            height: 8px !important;
        }

        .xterm-viewport::-webkit-scrollbar-thumb {
            border-radius: 4px;
            background: rgba(255, 255, 255, .18);
        }
    `])),K.background,K.compact?10:12,K.background)}}),Xe=(0,e.forwardRef)(function(c,K){var T,U,R=c.socket,ee=c.active,J=ee===void 0?!0:ee,ne=c.compact,re=ne===void 0?!1:ne,Ne=c.background,me=c.accentColor,Fe=c.className,N=c.style,Be=W.A.useToken(),y=Be.token,v=(0,b.DP)(),te=!!(!(v==null||(T=v.isDarkMode)===null||T===void 0)&&T.call(v)||!(v==null||(U=v.isDarkTheme)===null||U===void 0)&&U.call(v)),G=Ne||(te?"rgb(15, 15, 15)":"#000000"),he=me||(te?"#c3ccd6":y.colorPrimary),Pe=me||te?"rgba(195, 204, 214, .28)":y.colorPrimary,$e=(0,e.useMemo)(function(){return{compact:re,background:G}},[re,G]),oe=hn($e),Ue=oe.styles,ae=(0,e.useRef)(null),m=(0,e.useRef)(void 0),ie=(0,e.useRef)(void 0),$=(0,e.useRef)(void 0),pe=(0,e.useRef)(new TextEncoder),V=(0,e.useRef)(0),ce=(0,e.useRef)("");$.current=R;var Y=(0,e.useMemo)(function(){return{background:G,foreground:"#dedede",cursor:he,cursorAccent:G,selectionBackground:Pe,selectionForeground:"#ffffff",black:"#111111",red:"#ff7b72",green:"#7ee787",yellow:"#e3b341",blue:"#79c0ff",magenta:"#d2a8ff",cyan:"#a5d6ff",white:"#ededed",brightBlack:"#707070",brightRed:"#ffa198",brightGreen:"#aff5b4",brightYellow:"#f2cc60",brightBlue:"#a5d6ff",brightMagenta:"#e2c5ff",brightCyan:"#b6e3ff",brightWhite:"#ffffff"}},[Pe,he,G,y.colorPrimary]),De=(0,e.useRef)(Y);De.current=Y;var ge=(0,e.useCallback)(function(){var a=$.current,r=m.current;if(!(!r||(a==null?void 0:a.readyState)!==WebSocket.OPEN)){var d="".concat(r.cols,"|").concat(r.rows);if(ce.current!==d){ce.current=d;try{a.send(pe.current.encode(JSON.stringify({event:"resize",content:d})))}catch(k){ce.current=""}}}},[]),le=(0,e.useCallback)(function(){var a=ae.current,r=ie.current;if(!(!a||!r||a.clientWidth<=0||a.clientHeight<=0))try{r.fit(),ge()}catch(d){}},[ge]),p=(0,e.useCallback)(function(){var a=arguments.length>0&&arguments[0]!==void 0?arguments[0]:!1;cancelAnimationFrame(V.current),V.current=requestAnimationFrame(function(){V.current=requestAnimationFrame(function(){if(le(),a){var r;(r=m.current)===null||r===void 0||r.focus()}})})},[le]),be=(0,e.useCallback)(function(a){var r=$.current;if(!a||(r==null?void 0:r.readyState)!==WebSocket.OPEN)return!1;try{var d;return r.send(pe.current.encode(JSON.stringify({event:"write",content:a}))),(d=m.current)===null||d===void 0||d.focus(),!0}catch(k){return!1}},[]),xe=(0,e.useCallback)(function(a){var r=$.current,d=m.current;return!a||!d||(r==null?void 0:r.readyState)!==WebSocket.OPEN?!1:(d.paste(a),d.focus(),!0)},[]);return(0,e.useImperativeHandle)(K,function(){return{clear:function(){var r;return(r=m.current)===null||r===void 0?void 0:r.clear()},reset:function(){var r;return(r=m.current)===null||r===void 0?void 0:r.reset()},write:function(r,d){var k;return(k=m.current)===null||k===void 0?void 0:k.write(r,d)},focus:function(){var r;return(r=m.current)===null||r===void 0?void 0:r.focus()},fit:p,getSize:function(){var r,d;return le(),{cols:((r=m.current)===null||r===void 0?void 0:r.cols)||80,rows:((d=m.current)===null||d===void 0?void 0:d.rows)||24}},sendInput:be,paste:xe}},[le,xe,p,be]),(0,e.useEffect)(function(){var a=ae.current;if(a){var r=new ve({cursorBlink:!0,cursorStyle:"block",fontFamily:"ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, monospace",fontSize:12,lineHeight:1.25,logLevel:"off",scrollback:1e4,theme:De.current}),d=new x.Y;r.loadAddon(d),r.open(a),m.current=r,ie.current=d;var k=r.onData(be),ke=new ResizeObserver(function(){return p()});return ke.observe(a),p(),function(){cancelAnimationFrame(V.current),ke.disconnect(),k.dispose(),r.dispose(),m.current=void 0,ie.current=void 0}}},[]),(0,e.useEffect)(function(){var a=window.visualViewport;if(a){var r=function(){return p()};return a.addEventListener("resize",r),a.addEventListener("scroll",r),function(){a.removeEventListener("resize",r),a.removeEventListener("scroll",r)}}},[p]),(0,e.useEffect)(function(){var a=m.current;a&&(a.options.theme=Y,a.options.fontSize=12,p())},[re,p,Y]),(0,e.useEffect)(function(){if(ce.current="",!!R){var a=function(){$.current===R&&p(J)};return R.readyState===WebSocket.OPEN?a():R.addEventListener("open",a),function(){return R.removeEventListener("open",a)}}},[J,p,R]),(0,e.useEffect)(function(){J&&p(!0)},[J,p]),(0,o.jsx)("div",{ref:ae,className:[Ue.terminal,Fe].filter(Boolean).join(" "),style:N})});Xe.displayName="Terminal";var pn=(0,e.memo)(Xe),gn=l(30734),Ke=l(47577),bn=l(5309),xn={idle:"\u672A\u8FDE\u63A5",connecting:"\u8FDE\u63A5\u4E2D",connected:"\u5DF2\u8FDE\u63A5",disconnected:"\u5DF2\u65AD\u5F00",error:"\u8FDE\u63A5\u5931\u8D25"},kn=new TextDecoder,Ze=(0,e.forwardRef)(function(c,K){var T=c.path,U=c.params,R=c.binaryControl,ee=R===void 0?!1:R,J=c.acknowledge,ne=J===void 0?!1:J,re=c.title,Ne=re===void 0?"\u7EC8\u7AEF":re,me=c.endpoint,Fe=c.active,N=Fe===void 0?!0:Fe,Be=c.loading,y=Be===void 0?!1:Be,v=c.disabledReason,te=c.disabledIcon,G=te===void 0?"SettingOutlined":te,he=c.resetKey,Pe=he===void 0?0:he,$e=c.compact,oe=$e===void 0?!1:$e,Ue=c.styles,ae=c.showHeader,m=ae===void 0?!0:ae,ie=c.toolbarContainer,$=c.showStatus,pe=$===void 0?!0:$,V=c.showClear,ce=V===void 0?!0:V,Y=c.showReconnect,De=Y===void 0?!1:Y,ge=c.showConnect,le=ge===void 0?!0:ge,p=c.showFullscreen,be=p===void 0?!0:p,xe=c.autoConnect,a=xe===void 0?!1:xe,r=c.terminalBackground,d=c.terminalAccent,k=c.errorAction,ke=c.onStatusChange,ye=c.onClose,_e=c.showClose,wn=_e===void 0?!!ye:_e,qe=Ae.A.useApp(),Me=qe.message,Sn=qe.modal,Cn=(0,gn.A)({compact:oe,terminalBackground:r}),En=Cn.styles,we=Ue||En,Rn=(0,e.useMemo)(function(){return{algorithm:oe?[W.A.darkAlgorithm,W.A.compactAlgorithm]:W.A.darkAlgorithm}},[oe]),Q=(0,e.useRef)(null),w=(0,e.useRef)(null),Se=(0,e.useRef)(void 0),se=(0,e.useRef)(0),D=(0,e.useRef)(!1),Je=(0,e.useRef)(null),ue=(0,e.useRef)({}),jn=(0,e.useState)(),en=E()(jn,2),de=en[0],An=en[1],On=(0,e.useState)(),nn=E()(On,2),Tn=nn[0],rn=nn[1],Nn=(0,e.useState)("idle"),tn=E()(Nn,2),F=tn[0],Fn=tn[1],Bn=(0,e.useState)(null),on=E()(Bn,2),Ce=on[0],Pn=on[1],$n=(0,e.useState)(!1),an=E()($n,2),Ge=an[0],Dn=an[1],Mn=(0,e.useState)(!1),cn=E()(Mn,2),Ve=cn[0],ln=cn[1],M=JSON.stringify([T,U,ee,ne,Pe]),fe=(0,e.useRef)({path:T,params:U,binaryControl:ee,acknowledge:ne,connectionKey:M,active:N,loading:y,disabledReason:v,onStatusChange:ke});fe.current={path:T,params:U,binaryControl:ee,acknowledge:ne,connectionKey:M,active:N,loading:y,disabledReason:v,onStatusChange:ke};var j=(0,e.useCallback)(function(t){var n,u;D.current&&(Fn(t),(n=(u=fe.current).onStatusChange)===null||n===void 0||n.call(u,t))},[]),A=(0,e.useCallback)(function(t){Je.current=t,D.current&&Pn(t)},[]),S=(0,e.useCallback)(function(t){se.current+=1,window.clearTimeout(ue.current.timeout),window.clearInterval(ue.current.heartbeat),ue.current={};var n=Se.current;Se.current=void 0,D.current&&rn(void 0),n&&(n.onopen=null,n.onmessage=null,n.onerror=null,n.onclose=null,(n.readyState===WebSocket.CONNECTING||n.readyState===WebSocket.OPEN)&&n.close()),t&&j(t)},[j]),B=(0,e.useCallback)(function(){var t=arguments.length>0&&arguments[0]!==void 0?arguments[0]:!1,n=fe.current;if(!(n.loading||n.disabledReason||!w.current)){var u=Se.current;!t&&((u==null?void 0:u.readyState)===WebSocket.CONNECTING||(u==null?void 0:u.readyState)===WebSocket.OPEN)||(S(),A(null),j("connecting"),An({key:n.connectionKey,number:se.current}))}},[A,j,S]);(0,e.useImperativeHandle)(K,function(){return{clear:function(){var n,u;(n=w.current)===null||n===void 0||n.reset(),(u=w.current)===null||u===void 0||u.clear()},insertCommand:function(n){var u;return((u=w.current)===null||u===void 0?void 0:u.paste(n))||!1},connect:function(){return B()},reconnect:function(){return B(!0)}}},[B]),(0,e.useEffect)(function(){return D.current=!0,function(){D.current=!1,S()}},[S]),(0,e.useEffect)(function(){S(),A(null),j("idle")},[M,A,j,S]),(0,e.useEffect)(function(){if(!(!de||de.key!==M)){var t=w.current;if(t){var n=fe.current,u=(0,bn.dJ)();if(!u.token){A({code:"authorization_failed",message:"\u767B\u5F55\u72B6\u6001\u5DF2\u5931\u6548\uFF0C\u8BF7\u91CD\u65B0\u767B\u5F55"}),j("error");return}var i=++se.current,h;try{var C=new URL((Ke.A===null||Ke.A===void 0?void 0:Ke.A.gateway)||window.location.origin,window.location.origin);C.protocol=C.protocol==="https:"||C.protocol==="wss:"?"wss:":"ws:",C.hash="",C.search="",C.pathname="".concat(C.pathname.replace(/\/$/,""),"/").concat(n.path.replace(/^\//,"")).replace(/\/{2,}/g,"/"),Object.entries(n.params).forEach(function(z){var s=E()(z,2),g=s[0],I=s[1];return C.searchParams.set(g,String(I))});var un=t.getSize(),Wn=un.cols,Hn=un.rows;C.searchParams.set("width",String(Wn)),C.searchParams.set("height",String(Hn));var Kn=btoa(JSON.stringify(u)).replace(/\+/g,"-").replace(/\//g,"_").replace(/=+$/,"");h=new WebSocket(C.toString(),Kn)}catch(z){A({code:"connection_failed",message:"\u65E0\u6CD5\u521B\u5EFA\u7EC8\u7AEF\u8FDE\u63A5\uFF0C\u8BF7\u7A0D\u540E\u91CD\u8BD5"}),j("error");return}Se.current=h,h.binaryType="arraybuffer";var dn=!1,X=function(){return D.current&&fe.current.connectionKey===M&&se.current===i&&Se.current===h},ze=function(s,g){X()&&(A({code:s,message:g}),S("error"))};ue.current.timeout=window.setTimeout(function(){return ze("connection_timeout","\u7EC8\u7AEF\u8FDE\u63A5\u8D85\u65F6\uFF0C\u8BF7\u91CD\u8BD5")},3e4);var fn=function(s){if((s==null?void 0:s.event)==="ready"){var g;return dn=!0,window.clearTimeout(ue.current.timeout),A(null),rn(h),j("connected"),(g=w.current)===null||g===void 0||g.fit(fe.current.active),!0}return(s==null?void 0:s.event)==="error"?(ze(String(s.code||"connection_failed"),String(s.content||"\u7EC8\u7AEF\u8FDE\u63A5\u5931\u8D25")),!0):(s==null?void 0:s.event)==="pong"},vn=function(s){var g;if(X()){var I=new Uint8Array(s);if(n.binaryControl&&I[0]===123&&I[I.length-1]===125)try{if(fn(JSON.parse(kn.decode(I))))return}catch(Un){}(g=w.current)===null||g===void 0||g.write(I,n.acknowledge?function(){X()&&h.readyState===WebSocket.OPEN&&h.send(JSON.stringify({event:"ack"}))}:void 0)}};return h.onmessage=function(z){var s=z.data;if(X())if(typeof s=="string"){var g;try{if(fn(JSON.parse(s)))return}catch(I){}n.binaryControl&&((g=w.current)===null||g===void 0||g.write(s))}else s instanceof ArrayBuffer?vn(s):s instanceof Blob&&s.arrayBuffer().then(vn).catch(function(){return ze("output_failed","\u8BFB\u53D6\u7EC8\u7AEF\u8F93\u51FA\u5931\u8D25")})},h.onopen=function(){X()&&(ue.current.heartbeat=window.setInterval(function(){X()&&h.readyState===WebSocket.OPEN&&h.send(JSON.stringify({event:"ping"}))},2e4))},h.onerror=function(){return ze("connection_failed","\u7EC8\u7AEF\u8FDE\u63A5\u5931\u8D25\uFF0C\u8BF7\u68C0\u67E5\u9762\u677F\u7F51\u7EDC")},h.onclose=function(){X()&&(dn||A(Je.current||{code:"connection_closed",message:"\u7EC8\u7AEF\u8FDE\u63A5\u5DF2\u5173\u95ED\uFF0C\u8BF7\u786E\u8BA4\u670D\u52A1\u6B63\u5728\u8FD0\u884C"}),S(Je.current?"error":"disconnected"))},function(){return S()}}}},[de,M,A,j,S]),(0,e.useEffect)(function(){a&&N&&!y&&!v&&B()},[N,a,M,B,v,y]),(0,e.useEffect)(function(){var t=Q.current,n=function(){var i,h=document.fullscreenElement===t;Dn(h),(i=w.current)===null||i===void 0||i.fit(h)};return document.addEventListener("fullscreenchange",n),function(){document.removeEventListener("fullscreenchange",n),t&&document.fullscreenElement===t&&document.exitFullscreen().catch(function(){})}},[]);var zn=(0,e.useCallback)(P()(L()().mark(function t(){var n;return L()().wrap(function(i){for(;;)switch(i.prev=i.next){case 0:if(n=Q.current,!(!n||Ve)){i.next=3;break}return i.abrupt("return");case 3:if(!(!document.fullscreenEnabled||!n.requestFullscreen||!document.exitFullscreen)){i.next=6;break}return Me.info("\u5F53\u524D\u6D4F\u89C8\u5668\u4E0D\u652F\u6301\u5168\u5C4F"),i.abrupt("return");case 6:if(ln(!0),i.prev=7,document.fullscreenElement!==n){i.next=13;break}return i.next=11,document.exitFullscreen();case 11:i.next=15;break;case 13:return i.next=15,n.requestFullscreen();case 15:if(!(!n.isConnected&&document.fullscreenElement===n)){i.next=18;break}return i.next=18,document.exitFullscreen();case 18:i.next=23;break;case 20:i.prev=20,i.t0=i.catch(7),n.isConnected&&Me.error("\u5207\u6362\u5168\u5C4F\u5931\u8D25");case 23:return i.prev=23,n.isConnected&&ln(!1),i.finish(23);case 26:case"end":return i.stop()}},t,null,[[7,20,23,26]])})),[Ve,Me]),In=(function(){var t=P()(L()().mark(function n(){return L()().wrap(function(i){for(;;)switch(i.prev=i.next){case 0:if(!(document.fullscreenElement===Q.current&&document.exitFullscreen)){i.next=10;break}return i.prev=1,i.next=4,document.exitFullscreen();case 4:i.next=10;break;case 6:return i.prev=6,i.t0=i.catch(1),Me.error("\u9000\u51FA\u5168\u5C4F\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5"),i.abrupt("return");case 10:ye==null||ye();case 11:case"end":return i.stop()}},n,null,[[1,6]])}));return function(){return t.apply(this,arguments)}})(),Ee=F==="connected"||F==="connecting",Ye=Ee?"":v,Ln=Ee?"\u65AD\u5F00\u8FDE\u63A5":v||"\u8FDE\u63A5\u7EC8\u7AEF",Qe=[{key:"clear",visible:ce,title:"\u6E05\u7406",icon:"ClearOutlined",onClick:function(){var n;return(n=w.current)===null||n===void 0?void 0:n.clear()}},{key:"reconnect",visible:De,title:"\u91CD\u8FDE",icon:"ReloadOutlined",disabled:y||!!v,onClick:function(){return B(!0)}},{key:"connect",visible:le,title:Ln,icon:Ee?"DisconnectOutlined":"LinkOutlined",disabled:!Ee&&(y||!!v),onClick:function(){if(!Ee){B();return}var n=se.current;Sn.confirm({title:"\u65AD\u5F00\u8FDE\u63A5",content:"\u786E\u5B9A\u65AD\u5F00\u7EC8\u7AEF\u201C".concat(Ne,"\u201D\u7684\u8FDE\u63A5\u5417\uFF1F"),okText:"\u65AD\u5F00",cancelText:"\u53D6\u6D88",okButtonProps:{danger:!0,type:"default"},mask:{closable:!0},getContainer:document.fullscreenElement===Q.current&&Q.current||document.body,onOk:function(){D.current&&se.current===n&&S("disconnected")}})}},{key:"fullscreen",visible:be,title:Ge?"\u9000\u51FA\u5168\u5C4F":"\u5168\u5C4F",icon:Ge?"FullscreenExitOutlined":"FullscreenOutlined",disabled:Ve,onClick:function(){zn()}},{key:"close",visible:wn&&!!ye,title:"\u5173\u95ED",icon:"CloseOutlined",onClick:function(){In()}}].filter(function(t){return t.visible}),sn=(0,o.jsxs)(o.Fragment,{children:[pe&&(0,o.jsx)("span",{className:"status-text ".concat(F),children:xn[F]}),Qe.map(function(t){return(0,o.jsx)(Oe.A,{title:t.title,children:(0,o.jsx)(O.A,{type:"text","aria-label":t.title,disabled:t.disabled,icon:(0,o.jsx)(b.In,{type:t.icon}),onClick:t.onClick})},t.key)})]});return(0,o.jsxs)("section",{className:"".concat(we.consolePane," ").concat(N?"":"session-hidden"),"aria-hidden":!N,children:[m&&(0,o.jsxs)("header",{className:we.consoleHeader,children:[(0,o.jsxs)("div",{className:"console-title",children:[pe&&(0,o.jsx)("span",{className:"status-dot ".concat(F)}),(0,o.jsxs)("div",{children:[(0,o.jsx)("div",{className:"console-name",children:Ne}),me&&(0,o.jsx)("div",{className:"console-endpoint",children:me})]})]}),(0,o.jsx)("div",{className:"console-actions",children:sn})]}),!m&&ie&&(0,je.createPortal)(sn,ie),(0,o.jsx)(Z.Ay,{theme:Rn,children:(0,o.jsx)("div",{className:we.terminalBody,children:(0,o.jsxs)("div",{className:we.terminalScreen,ref:Q,children:[(0,o.jsx)(pn,{ref:w,socket:Tn,active:N,compact:oe,background:r,accentColor:d},"".concat(M,":").concat((de==null?void 0:de.number)||0)),(y||Ye||F!=="connected")&&(0,o.jsx)("div",{className:"".concat(we.terminalOverlay," ").concat(F),children:y?(0,o.jsx)(O.A,{type:"text",disabled:!0,children:v||"\u6B63\u5728\u52A0\u8F7D\u8FDE\u63A5"}):Ye?(0,o.jsx)(O.A,{type:"text",disabled:!0,icon:(0,o.jsx)(b.In,{type:G}),children:Ye}):F==="connecting"?(0,o.jsx)(_.A,{size:"small",description:"\u6B63\u5728\u8FDE\u63A5..."}):F==="error"?(0,o.jsxs)("div",{className:"terminal-error",children:[(0,o.jsxs)("div",{className:"terminal-error-title",children:[(0,o.jsx)(b.In,{type:"WarningOutlined"}),(0,o.jsx)("span",{children:"\u8FDE\u63A5\u5931\u8D25"})]}),(0,o.jsx)("div",{className:"terminal-error-message",children:(Ce==null?void 0:Ce.message)||"\u7EC8\u7AEF\u8FDE\u63A5\u5931\u8D25\uFF0C\u8BF7\u7A0D\u540E\u91CD\u8BD5"}),(0,o.jsxs)("div",{className:"terminal-error-actions",children:[Ce&&(k==null?void 0:k.codes.includes(Ce.code))&&(0,o.jsx)(O.A,{color:"primary",variant:"solid",onClick:function(){var n=function(){D.current&&k.onClick()};document.fullscreenElement===Q.current&&document.exitFullscreen?document.exitFullscreen().then(n).catch(n):n()},children:k.label}),(0,o.jsx)(O.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},onClick:function(){return B()},children:"\u91CD\u65B0\u8FDE\u63A5"})]})]}):(0,o.jsx)(O.A,{color:"primary",variant:"solid",icon:(0,o.jsx)(b.In,{type:"LinkOutlined"}),onClick:function(){return B()},children:"\u8FDE\u63A5\u7EC8\u7AEF"})}),Ge&&Qe.length>0&&(0,o.jsx)("div",{className:"fullscreen-actions",children:Qe.map(function(t){return(0,o.jsx)(O.A,{type:"text",title:t.title,"aria-label":t.title,disabled:t.disabled,icon:(0,o.jsx)(b.In,{type:t.icon}),onClick:t.onClick},t.key)})})]})})})]})});Ze.displayName="TerminalSession";var yn=(0,e.memo)(Ze)}),30734:(function(mn,Re,l){l.d(Re,{x:function(){return Oe}});var Ie=l(1974),L=l.n(Ie),Le=l(84169),P=l.n(Le),We=l(12185),E,e,je,Ae,W,Oe=function(_){var b=_.css,f=_.token,H=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{},Te=H.compact,x=Te===void 0?!1:Te,He=H.terminalBackground,o=He===void 0?"#000000":He,q=x?46:50,ve=x?42:46;return{consolePane:b(E||(E=P()([`
            min-width: 0;
            min-height: 0;
            display: flex;
            flex-direction: column;
            background: `,`;

            &.session-hidden {
                display: none;
            }

            @container terminal-workspace (max-width: 760px) {
                grid-column: 1;
                grid-row: 3;
                width: 100%;
                max-width: 100%;
                height: auto;
                min-height: max(600px, calc(var(--terminal-page-height) - `,`px));
                overflow: hidden;
                box-sizing: border-box;
            }

            @media (max-width: 780px) and (max-height: 600px) and (orientation: landscape) {
                min-height: 440px;
            }

            @media (min-width: 781px) and (max-height: 600px) and (orientation: landscape) {
                min-height: 300px;
            }
        `])),f.colorBgContainer,x?220:244),consoleHeader:b(e||(e=P()([`
            min-width: 0;
            height: `,`px;
            min-height: `,`px;
            padding: 0 `,`px;
            display: flex;
            flex: 0 0 `,`px;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
            box-sizing: border-box;
            border-bottom: 1px solid `,`;

            .console-title {
                min-width: 0;
                display: flex;
                align-items: center;
                gap: 9px;
            }

            .console-title > .status-dot {
                width: 7px;
                height: 7px;
                flex: none;
                border-radius: 50%;
                background: `,`;
            }

            .console-title > .status-dot.connected {
                background: `,`;
                box-shadow: 0 0 0 3px `,`;
            }

            .console-title > .status-dot.connecting {
                background: `,`;
                box-shadow: 0 0 0 3px `,`;
            }

            .console-title > .status-dot.error {
                background: `,`;
                box-shadow: 0 0 0 3px `,`;
            }

            .console-title > div {
                min-width: 0;
            }

            .console-name,
            .console-endpoint {
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }

            .console-name {
                color: `,`;
                font-size: `,`px;
                font-weight: 600;
                line-height: `,`px;
            }

            .console-endpoint {
                color: `,`;
                font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
                font-size: `,`px;
                line-height: `,`px;
            }

            .console-actions {
                display: flex;
                flex: none;
                align-items: center;
                gap: 2px;
            }

            .console-actions .status-text {
                margin-right: 5px;
                color: `,`;
                font-size: 11px;
                white-space: nowrap;
            }

            .console-actions .status-text.connected {
                color: `,`;
            }

            .console-actions .status-text.connecting {
                color: `,`;
            }

            .console-actions .status-text.error {
                color: `,`;
            }

            .console-actions .ant-btn {
                width: `,`px;
                height: `,`px;
                padding: 0;
            }

            @container terminal-workspace (max-width: 520px) {
                .console-endpoint,
                .console-actions .status-text {
                    display: none;
                }
            }

            @container terminal-workspace (max-width: 760px) {
                height: `,`px;
                min-height: `,`px;
                flex-basis: `,`px;
            }
        `])),q,q,x?11:13,q,f.colorSplit,f.colorTextQuaternary,f.colorSuccess,f.colorSuccessBg,f.colorWarning,f.colorWarningBg,f.colorError,f.colorErrorBg,f.colorTextHeading,x?13:14,x?19:20,f.colorTextTertiary,x?11:12,x?15:16,f.colorTextTertiary,f.colorSuccessText,f.colorWarningText,f.colorErrorText,x?30:32,x?30:32,ve,ve,ve),terminalBody:b(je||(je=P()([`
            min-width: 0;
            min-height: 0;
            padding: 0;
            display: flex;
            flex: 1;
            background: `,`;
        `])),o),terminalScreen:b(Ae||(Ae=P()([`
            position: relative;
            width: 100%;
            min-width: 0;
            min-height: 0;
            overflow: hidden;
            display: flex;
            flex: 1;
            border: 0;
            border-radius: 0;
            background: `,`;

            &:fullscreen {
                width: 100vw;
                height: 100dvh;
                min-width: 0;
                min-height: 0;
                border: 0;
                border-radius: 0;
                background: `,`;
            }

            &:fullscreen::backdrop {
                background: `,`;
            }

            .fullscreen-actions {
                position: absolute;
                z-index: 3;
                top: 10px;
                right: 12px;
                padding: 3px;
                display: flex;
                align-items: center;
                gap: 2px;
                border: 1px solid rgba(255, 255, 255, .08);
                border-radius: `,`px;
                background: rgba(18, 18, 18, .82);
                backdrop-filter: blur(8px);
            }

            .fullscreen-actions .ant-btn {
                width: 30px;
                height: 30px;
                padding: 0;
            }

            @supports not (height: 100dvh) {
                &:fullscreen {
                    height: 100vh;
                }
            }
        `])),o,o,o,f.borderRadius),terminalOverlay:b(W||(W=P()([`
            position: absolute;
            z-index: 2;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: inherit;
            background: color-mix(in srgb, `,`, transparent 10%);
            pointer-events: auto;

            &.connecting {
                color: #dedede;
            }

            .ant-spin-description {
                color: #a8a8a8;
                font-size: 12px;
            }

            .terminal-error {
                width: min(360px, calc(100% - 32px));
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: 10px;
                color: rgba(255, 255, 255, .86);
                text-align: center;
            }

            .terminal-error-title {
                display: inline-flex;
                align-items: center;
                gap: 7px;
                color: #fff;
                font-size: 14px;
                font-weight: 600;
            }

            .terminal-error-title .anticon {
                color: `,`;
                font-size: 15px;
            }

            .terminal-error-message {
                max-width: 100%;
                color: rgba(255, 255, 255, .58);
                font-size: 12px;
                line-height: 20px;
                overflow-wrap: anywhere;
            }

            .terminal-error-actions {
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 8px;
            }

            .terminal-error-actions .ant-btn {
                height: `,`px;
                padding-inline: 13px;
            }
        `])),o,f.colorWarning,x?28:30)}},O=(0,We.rU)(function(Z){var _=Z.css,b=Z.token,f=Z.isDarkMode,H=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};return Oe({css:_,token:b},L()(L()({},H),{},{terminalBackground:H.terminalBackground||(f?"rgb(15, 15, 15)":"#000000")}))});Re.A=O})}]);
