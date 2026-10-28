(self.webpackChunk=self.webpackChunk||[]).push([[765],{14283:(function(j,Z,a){"use strict";a.r(Z),a.d(Z,{default:function(){return We}});var s=a(96540),ue=a(47152),se=a(16370),ve=a(79181),ne=a(62885),W=a(39058),Ke=a(55420),A=a.n(Ke),Je=a(98312),X=a.n(Je),Qe=a(11604),S=a.n(Qe),ge=a(79273),x=a(20873),re=a(22817),pe=a(72523),me=a(50400),e=a(74848),ce=(function(o){var n=o.styles,t=o.saving,r=o.onReset;return(0,e.jsxs)("div",{className:n.formActions,children:[(0,e.jsx)(me.A,{onClick:r,children:"\u91CD\u7F6E"}),(0,e.jsx)(me.A,{type:"primary",htmlType:"submit",loading:t,children:"\u4FDD\u5B58"})]})});a.dn(ce);var K=(function(o){var n=o.styles,t=o.message,r=o.onRetry;return(0,e.jsxs)("div",{className:n.errorState,children:[(0,e.jsx)(ne.In,{type:"ExclamationCircleOutlined"}),(0,e.jsx)("span",{children:t||"\u914D\u7F6E\u52A0\u8F7D\u5931\u8D25"}),(0,e.jsx)(me.A,{size:"small",icon:(0,e.jsx)(ne.In,{type:"ReloadOutlined"}),onClick:r,children:"\u91CD\u65B0\u52A0\u8F7D"})]})});a.dn(K);var Ye=a(45478),de=(function(o){var n=o.styles;return(0,e.jsx)("div",{className:n.loading,children:(0,e.jsx)(Ye.A,{})})});a.dn(de);var je=(function(o){var n=o.styles,t=ge.A.useApp(),r=t.message,i=(0,W.fn)("user"),l=x.A.useForm(),c=S()(l,1),N=c[0],E=(0,s.useState)(!0),F=S()(E,2),G=F[0],$=F[1],L=(0,s.useState)(!1),J=S()(L,2),V=J[0],R=J[1],I=(0,s.useState)(""),M=S()(I,2),q=M[0],ie=M[1],U=(0,s.useRef)(!0),C=(0,s.useRef)(!1),H=(0,s.useRef)(0),Q=(0,s.useRef)(0),B=(0,s.useRef)(""),_=(0,s.useCallback)(X()(A()().mark(function z(){var T,k,d,y;return A()().wrap(function(p){for(;;)switch(p.prev=p.next){case 0:return k=++H.current,$(!0),ie(""),p.next=5,(0,pe.zq)({body:{action:"account"}});case 5:if(d=p.sent,!(!U.current||k!==H.current)){p.next=8;break}return p.abrupt("return");case 8:if((d==null?void 0:d.code)===200){p.next=12;break}return ie((d==null?void 0:d.message)||"\u767B\u5F55\u4FE1\u606F\u52A0\u8F7D\u5931\u8D25"),$(!1),p.abrupt("return");case 12:y=String((d==null||(T=d.data)===null||T===void 0?void 0:T.account)||""),B.current=y,N.setFieldsValue({account:y,password:"",confirm:""}),$(!1);case 16:case"end":return p.stop()}},z)})),[N]);(0,s.useEffect)(function(){return U.current=!0,_(),function(){U.current=!1,H.current+=1,Q.current+=1}},[_]);var te=(0,s.useCallback)(function(){N.setFieldsValue({account:B.current,password:"",confirm:""})},[N]),ae=(0,s.useCallback)((function(){var z=X()(A()().mark(function T(k){var d,y,v,p,b,P,g,h;return A()().wrap(function(m){for(;;)switch(m.prev=m.next){case 0:if(!C.current){m.next=2;break}return m.abrupt("return");case 2:if(y=String(k.account||""),v=y.trim(),p=String(k.password||""),b={action:"update"},y!==B.current&&(b.account=v),p&&(b.password=p),!(!b.account&&!b.password)){m.next=11;break}return r.info("\u767B\u5F55\u4FE1\u606F\u6CA1\u6709\u53D8\u5316"),m.abrupt("return");case 11:return C.current=!0,P=++Q.current,R(!0),m.next=16,(0,pe.zq)({body:b});case 16:if(g=m.sent,!(!U.current||P!==Q.current)){m.next=19;break}return m.abrupt("return");case 19:if((g==null?void 0:g.code)===200){m.next=24;break}return r.error((g==null?void 0:g.message)||"\u767B\u5F55\u4FE1\u606F\u4FDD\u5B58\u5931\u8D25"),C.current=!1,R(!1),m.abrupt("return");case 24:h=b.account||B.current,B.current=h,N.setFieldsValue({account:h,password:"",confirm:""}),i==null||(d=i.refreshWebUser)===null||d===void 0||d.call(i),r.success((g==null?void 0:g.message)||"\u767B\u5F55\u4FE1\u606F\u4FDD\u5B58\u6210\u529F"),C.current=!1,R(!1);case 31:case"end":return m.stop()}},T)}));return function(T){return z.apply(this,arguments)}})(),[N,r,i]);return(0,e.jsx)(x.A,{form:N,layout:"vertical",variant:"filled",onFinish:ae,children:G?(0,e.jsx)(de,{styles:n}):q?(0,e.jsx)(K,{styles:n,message:q,onRetry:_}):(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(x.A.Item,{name:"account",label:"\u767B\u5F55\u8D26\u53F7",tooltip:"\u7528\u4E8E\u767B\u5F55\u9762\u677F\u7684\u8D26\u53F7",className:n.formField,rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u767B\u5F55\u8D26\u53F7"},{validator:function(T,k){return!String(k!=null?k:"").trim()||k===B.current?Promise.resolve():/^[A-Za-z0-9]+$/.test(k||"")?String(k).length>20?Promise.reject(new K("\u767B\u5F55\u8D26\u53F7\u4E0D\u80FD\u8D85\u8FC720\u4E2A\u5B57\u7B26")):Promise.resolve():Promise.reject(new K("\u767B\u5F55\u8D26\u53F7\u53EA\u80FD\u5305\u542B\u5B57\u6BCD\u548C\u6570\u5B57"))}}],children:(0,e.jsx)(re.A,{placeholder:"\u8BF7\u8F93\u5165\u767B\u5F55\u8D26\u53F7",maxLength:20,autoComplete:"username"})}),(0,e.jsx)(x.A.Item,{name:"password",label:"\u65B0\u5BC6\u7801",tooltip:"\u4E0D\u4FEE\u6539\u5BC6\u7801\u65F6\u8BF7\u7559\u7A7A",className:n.formField,rules:[{min:6,message:"\u5BC6\u7801\u81F3\u5C116\u4E2A\u5B57\u7B26"},{max:20,message:"\u5BC6\u7801\u4E0D\u80FD\u8D85\u8FC720\u4E2A\u5B57\u7B26"}],children:(0,e.jsx)(re.A.Password,{placeholder:"\u4E0D\u4FEE\u6539\u8BF7\u7559\u7A7A",maxLength:20,autoComplete:"new-password"})}),(0,e.jsx)(x.A.Item,{name:"confirm",label:"\u786E\u8BA4\u5BC6\u7801",tooltip:"\u518D\u6B21\u8F93\u5165\u65B0\u5BC6\u7801",className:n.formField,dependencies:["password"],rules:[function(z){var T=z.getFieldValue;return{validator:function(d,y){var v=T("password");return!v&&!y?Promise.resolve():v?y===v?Promise.resolve():Promise.reject(new K("\u4E24\u6B21\u8F93\u5165\u7684\u5BC6\u7801\u4E0D\u4E00\u81F4")):Promise.reject(new K("\u8BF7\u5148\u8F93\u5165\u65B0\u5BC6\u7801"))}}}],children:(0,e.jsx)(re.A.Password,{placeholder:"\u518D\u6B21\u8F93\u5165\u65B0\u5BC6\u7801",maxLength:20,autoComplete:"new-password"})}),(0,e.jsx)(ce,{styles:n,saving:V,onReset:te})]})})});a.dn(je);var oe=a(58679),Ze=a(97087),Xe=a(14039),qe=a(47577),we=a(78781);function _e(){return he.apply(this,arguments)}function he(){return he=X()(A()().mark(function o(){var n,t=arguments;return A()().wrap(function(i){for(;;)switch(i.prev=i.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},i.abrupt("return",(0,we.Jt)("/config/get",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return i.stop()}},o)})),he.apply(this,arguments)}function en(){return xe.apply(this,arguments)}function xe(){return xe=X()(A()().mark(function o(){var n,t=arguments;return A()().wrap(function(i){for(;;)switch(i.prev=i.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},i.abrupt("return",(0,we.bE)("/config/set",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return i.stop()}},o)})),xe.apply(this,arguments)}var nn=function(n){return typeof n=="boolean"?n?"1":"0":String(n!=null?n:"")},ye=(function(o,n){var t=arguments.length>2&&arguments[2]!==void 0?arguments[2]:function(d){return d},r=ge.A.useApp(),i=r.message,l=(0,W.fn)("web"),c=x.A.useForm(),N=S()(c,1),E=N[0],F=(0,s.useState)(!0),G=S()(F,2),$=G[0],L=G[1],J=(0,s.useState)(!1),V=S()(J,2),R=V[0],I=V[1],M=(0,s.useState)(""),q=S()(M,2),ie=q[0],U=q[1],C=(0,s.useRef)(!0),H=(0,s.useRef)(!1),Q=(0,s.useRef)(0),B=(0,s.useRef)(0),_=(0,s.useRef)({}),te=(0,s.useRef)(n),ae=(0,s.useRef)(t);te.current=n,ae.current=t;var z=(0,s.useCallback)(X()(A()().mark(function d(){var y,v,p,b,P;return A()().wrap(function(h){for(;;)switch(h.prev=h.next){case 0:return y=++Q.current,L(!0),U(""),h.next=5,_e({body:{group:o}});case 5:if(v=h.sent,!(!C.current||y!==Q.current)){h.next=8;break}return h.abrupt("return");case 8:if((v==null?void 0:v.code)===200){h.next=12;break}return U((v==null?void 0:v.message)||"\u914D\u7F6E\u52A0\u8F7D\u5931\u8D25"),L(!1),h.abrupt("return");case 12:p=(v==null?void 0:v.data)||{},b=Object.keys(te.current).reduce(function(O,m){var f="".concat(o,".").concat(m),w=p[f];return O[m]=w===void 0?te.current[m]:w,O},{}),P=ae.current(b),_.current=P,E.setFieldsValue(P),L(!1);case 18:case"end":return h.stop()}},d)})),[E,o]);(0,s.useEffect)(function(){return C.current=!0,z(),function(){C.current=!1,Q.current+=1,B.current+=1}},[z]);var T=(0,s.useCallback)((function(){var d=X()(A()().mark(function y(v){var p,b,P,g,h,O;return A()().wrap(function(f){for(;;)switch(f.prev=f.next){case 0:if(!H.current){f.next=2;break}return f.abrupt("return",!1);case 2:return H.current=!0,b=++B.current,I(!0),P=ae.current(v),f.next=8,en({body:{configs:Object.entries(P).map(function(w){var le=S()(w,2),u=le[0],ee=le[1];return{key:"".concat(o,".").concat(u),value:nn(ee)}})}});case 8:if(g=f.sent,!(!C.current||b!==B.current)){f.next=11;break}return f.abrupt("return",!1);case 11:if((g==null?void 0:g.code)===200){f.next=16;break}return i.error((g==null?void 0:g.message)||"\u914D\u7F6E\u4FDD\u5B58\u5931\u8D25"),H.current=!1,I(!1),f.abrupt("return",!1);case 16:return i.success((g==null?void 0:g.message)||"\u914D\u7F6E\u4FDD\u5B58\u6210\u529F"),f.next=19,z();case 19:return f.next=21,l==null||(p=l.getInfo)===null||p===void 0?void 0:p.call(l);case 21:return h=f.sent,C.current&&b===B.current&&h&&(l==null||(O=l.setInfo)===null||O===void 0||O.call(l,h)),C.current&&b===B.current&&(H.current=!1,I(!1)),f.abrupt("return",!0);case 25:case"end":return f.stop()}},y)}));return function(y){return d.apply(this,arguments)}})(),[o,z,i,l]),k=(0,s.useCallback)(function(){E.setFieldsValue(_.current)},[E]);return{form:E,loading:$,saving:R,error:ie,load:z,save:T,reset:k}});a.dn(ye);var Fe=function(n){return n===!0||n===1||n==="1"||n==="true"},Ce=[{value:"1",label:"\u5F00\u542F"},{value:"0",label:"\u5173\u95ED"}],rn=[{value:"top",label:"\u4E0A\u4E0B\u5E03\u5C40"},{value:"left",label:"\u5DE6\u53F3\u5E03\u5C40"}],tn=[{value:"light",label:"\u4EAE\u8272\u6A21\u5F0F"},{value:"dark",label:"\u6697\u8272\u6A21\u5F0F"}],an=Array.from({length:16},function(o,n){return{value:String(n),label:"".concat(n,"px").concat(n===0?" (\u65E0\u5706\u89D2)":"")}}),ke=function(n){return{layout:n.layout==="top"?"top":"left",watermark:Fe(n.watermark)?"1":"0",theme:n.theme==="light"?"light":"dark",compact:Fe(n.compact)?"1":"0",color:n.color===void 0||n.color===null?qe.A.color:String(n.color),radius:String(Math.min(15,Math.max(0,Number(n.radius)||0)))}},Ae=(function(o){var n=o.styles,t=o.info,r=t==null?void 0:t.ui,i=(0,s.useMemo)(function(){return ke({layout:r==null?void 0:r.layout,watermark:r==null?void 0:r.watermark,theme:r==null?void 0:r.theme,compact:r==null?void 0:r.compact,color:r==null?void 0:r.color,radius:r==null?void 0:r.radius})},[r==null?void 0:r.color,r==null?void 0:r.compact,r==null?void 0:r.layout,r==null?void 0:r.radius,r==null?void 0:r.theme,r==null?void 0:r.watermark]),l=ye("ui",i,ke);return(0,e.jsx)("div",{className:n.uiWrapper,children:(0,e.jsx)(x.A,{form:l.form,layout:"vertical",variant:"filled",onFinish:l.save,children:l.loading?(0,e.jsx)(de,{styles:n}):l.error?(0,e.jsx)(K,{styles:n,message:l.error,onRetry:l.load}):(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(x.A.Item,{name:"compact",label:"\u7D27\u51D1\u6A21\u5F0F",tooltip:"\u63A7\u5236\u754C\u9762\u5143\u7D20\u95F4\u8DDD\uFF0C\u5F00\u542F\u540E\u754C\u9762\u66F4\u7D27\u51D1",className:n.formField,rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u7D27\u51D1\u6A21\u5F0F"}],children:(0,e.jsx)(oe.A,{placeholder:"\u8BF7\u9009\u62E9\u7D27\u51D1\u6A21\u5F0F\u662F\u5426\u5F00\u542F",options:Ce})}),(0,e.jsx)(x.A.Item,{name:"layout",label:"\u754C\u9762\u5E03\u5C40",tooltip:"\u9009\u62E9\u754C\u9762\u7684\u6574\u4F53\u5E03\u5C40\u65B9\u5F0F\uFF0C\u4E0A\u4E0B\u5E03\u5C40\u6216\u5DE6\u53F3\u5E03\u5C40",className:n.formField,rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u754C\u9762\u5E03\u5C40"}],children:(0,e.jsx)(oe.A,{placeholder:"\u8BF7\u9009\u62E9\u754C\u9762\u5E03\u5C40\u65B9\u5F0F",options:rn})}),(0,e.jsx)(x.A.Item,{name:"theme",label:"\u83DC\u5355\u4E3B\u9898",tooltip:"\u9009\u62E9\u83DC\u5355\u7684\u4E3B\u9898\u6A21\u5F0F\uFF0C\u4EAE\u8272\u6216\u6697\u8272",className:n.formField,rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u83DC\u5355\u4E3B\u9898"}],children:(0,e.jsx)(oe.A,{placeholder:"\u8BF7\u9009\u62E9\u83DC\u5355\u4E3B\u9898\u6A21\u5F0F",options:tn})}),(0,e.jsx)(x.A.Item,{name:"watermark",label:"\u754C\u9762\u6C34\u5370",tooltip:"\u63A7\u5236\u754C\u9762\u662F\u5426\u663E\u793A\u8D26\u6237\u6C34\u5370",className:n.formField,rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u754C\u9762\u6C34\u5370"}],children:(0,e.jsx)(oe.A,{placeholder:"\u8BF7\u9009\u62E9\u662F\u5426\u663E\u793A\u754C\u9762\u6C34\u5370",options:Ce})}),(0,e.jsx)(x.A.Item,{name:"radius",label:"\u4E3B\u9898\u5706\u89D2",tooltip:"\u8C03\u6574\u754C\u9762\u5143\u7D20\u7684\u5706\u89D2\u5927\u5C0F",className:n.formField,rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u4E3B\u9898\u5706\u89D2"}],children:(0,e.jsx)(oe.A,{placeholder:"\u8BF7\u9009\u62E9\u4E3B\u9898\u5706\u89D2\u5927\u5C0F",options:an})}),(0,e.jsx)(x.A.Item,{label:"\u4E3B\u9898\u989C\u8272",tooltip:"\u9009\u62E9\u754C\u9762\u7684\u4E3B\u9898\u989C\u8272",className:n.formField,children:(0,e.jsxs)(Ze.A,{gap:8,align:"center",children:[(0,e.jsx)(x.A.Item,{name:"color",noStyle:!0,getValueFromEvent:function(N,E){return E},children:(0,e.jsx)(Xe.A,{showText:!0,format:"rgb",defaultFormat:"rgb"})}),(0,e.jsx)(x.A.Item,{shouldUpdate:!0,noStyle:!0,children:function(N){var E=N.setFieldValue;return(0,e.jsx)(me.A,{onClick:function(){return E("color","")},children:"\u6E05\u7A7A"})}})]})}),(0,e.jsx)(ce,{styles:n,saving:l.saving,onReset:l.reset})]})})})});a.dn(Ae);var Se=(function(o){var n=o.styles,t=o.info,r=(0,s.useMemo)(function(){return{name:(t==null?void 0:t.name)||"Sinking Panel",title:(t==null?void 0:t.title)||"Sinking Panel"}},[t==null?void 0:t.name,t==null?void 0:t.title]),i=ye("web",r);return(0,e.jsx)(x.A,{form:i.form,layout:"vertical",variant:"filled",onFinish:i.save,children:i.loading?(0,e.jsx)(de,{styles:n}):i.error?(0,e.jsx)(K,{styles:n,message:i.error,onRetry:i.load}):(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(x.A.Item,{name:"name",label:"\u7F51\u7AD9\u540D\u79F0",tooltip:"\u7F51\u7AD9\u540D\u79F0",className:n.formField,rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u7F51\u7AD9\u540D\u79F0"},{max:50,message:"\u7F51\u7AD9\u540D\u79F0\u4E0D\u80FD\u8D85\u8FC750\u4E2A\u5B57\u7B26"}],children:(0,e.jsx)(re.A,{placeholder:"\u8BF7\u8F93\u5165\u7F51\u7AD9\u540D\u79F0",maxLength:50})}),(0,e.jsx)(x.A.Item,{name:"title",label:"\u7F51\u7AD9\u6807\u9898",tooltip:"\u7F51\u7AD9\u6807\u9898",className:n.formField,rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u7F51\u7AD9\u6807\u9898"},{max:100,message:"\u7F51\u7AD9\u6807\u9898\u4E0D\u80FD\u8D85\u8FC7100\u4E2A\u5B57\u7B26"}],children:(0,e.jsx)(re.A,{placeholder:"\u8BF7\u8F93\u5165\u7F51\u7AD9\u6807\u9898",maxLength:100})}),(0,e.jsx)(ce,{styles:n,saving:i.saving,onReset:i.reset})]})})});a.dn(Se);var sn=a(45647),on=a.n(sn),Ne=(function(o){var n=o.styles,t=ge.A.useApp(),r=t.message,i=x.A.useForm(),l=S()(i,1),c=l[0],N=(0,s.useState)(!0),E=S()(N,2),F=E[0],G=E[1],$=(0,s.useState)(!1),L=S()($,2),J=L[0],V=L[1],R=(0,s.useState)(""),I=S()(R,2),M=I[0],q=I[1],ie=(0,s.useState)(""),U=S()(ie,2),C=U[0],H=U[1],Q=(0,s.useState)([]),B=S()(Q,2),_=B[0],te=B[1],ae=(0,s.useState)(""),z=S()(ae,2),T=z[0],k=z[1],d=(0,s.useRef)(!0),y=(0,s.useRef)(!1),v=(0,s.useRef)(0),p=(0,s.useRef)(0),b=(0,s.useCallback)(X()(A()().mark(function h(){var O,m,f,w,le,u,ee,fe;return A()().wrap(function(Y){for(;;)switch(Y.prev=Y.next){case 0:return le=++v.current,G(!0),q(""),Y.next=5,(0,pe.tN)({body:{action:"get"}});case 5:if(u=Y.sent,!(!d.current||le!==v.current)){Y.next=8;break}return Y.abrupt("return");case 8:if((u==null?void 0:u.code)===200){Y.next=12;break}return q((u==null?void 0:u.message)||"\u65F6\u533A\u4FE1\u606F\u52A0\u8F7D\u5931\u8D25"),G(!1),Y.abrupt("return");case 12:ee=String((u==null||(O=u.data)===null||O===void 0?void 0:O.timezone)||""),fe=Array.isArray(u==null||(m=u.data)===null||m===void 0?void 0:m.timezones)?u==null||(f=u.data)===null||f===void 0?void 0:f.timezones.filter(function(vn){return typeof vn=="string"}):[],H(ee),te(ee&&!fe.includes(ee)?[ee].concat(on()(fe)):fe),k(String((u==null||(w=u.data)===null||w===void 0?void 0:w.time)||"").replace("T"," ")),c.setFieldsValue({timezone:ee||void 0}),G(!1);case 19:case"end":return Y.stop()}},h)})),[c]);(0,s.useEffect)(function(){return d.current=!0,b(),function(){d.current=!1,v.current+=1,p.current+=1}},[b]);var P=(0,s.useCallback)(function(){c.setFieldsValue({timezone:C||void 0})},[c,C]),g=(0,s.useCallback)((function(){var h=X()(A()().mark(function O(m){var f,w;return A()().wrap(function(u){for(;;)switch(u.prev=u.next){case 0:if(!y.current){u.next=2;break}return u.abrupt("return");case 2:if(m.timezone!==C){u.next=5;break}return r.info("\u65F6\u533A\u6CA1\u6709\u53D8\u5316"),u.abrupt("return");case 5:return y.current=!0,f=++p.current,V(!0),u.next=10,(0,pe.tN)({body:{action:"set",timezone:m.timezone}});case 10:if(w=u.sent,!(!d.current||f!==p.current)){u.next=13;break}return u.abrupt("return");case 13:if((w==null?void 0:w.code)===200){u.next=18;break}return r.error((w==null?void 0:w.message)||"\u65F6\u533A\u4FDD\u5B58\u5931\u8D25"),y.current=!1,V(!1),u.abrupt("return");case 18:return r.success((w==null?void 0:w.message)||"\u65F6\u533A\u4FDD\u5B58\u6210\u529F"),u.next=21,b();case 21:d.current&&f===p.current&&(y.current=!1,V(!1));case 22:case"end":return u.stop()}},O)}));return function(O){return h.apply(this,arguments)}})(),[b,r,C]);return(0,e.jsx)(x.A,{form:c,layout:"vertical",variant:"filled",disabled:J,onFinish:g,children:F?(0,e.jsx)(de,{styles:n}):M?(0,e.jsx)(K,{styles:n,message:M,onRetry:b}):(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(x.A.Item,{label:"\u5F53\u524D\u65F6\u533A",className:n.formField,children:(0,e.jsx)(re.A,{value:C||"\u2014",readOnly:!0})}),(0,e.jsx)(x.A.Item,{label:"\u670D\u52A1\u5668\u65F6\u95F4",className:n.formField,children:(0,e.jsx)(re.A,{value:T||"\u2014",readOnly:!0})}),(0,e.jsx)(x.A.Item,{name:"timezone",label:"\u65F6\u533A",className:n.formField,rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u65F6\u533A"}],children:(0,e.jsx)(oe.A,{showSearch:{optionFilterProp:"label"},placeholder:"\u641C\u7D22\u5E76\u9009\u62E9\u65F6\u533A",options:_.map(function(h){return{label:h,value:h}})})}),(0,e.jsx)(ce,{styles:n,saving:J,onReset:P})]})})});a.dn(Ne);var ln=a(84169),D=a.n(ln),Ee=a(12185),De,un=(0,Ee.rU)(function(o){var n=o.css,t=o.token;return{graphic:n(De||(De=D()([`
        width: 100%;
        height: 100%;
        display: block;
        overflow: visible;
        color: `,`;

        .flow-line {
            fill: none;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .flow-primary {
            stroke-width: 1.6;
        }

        .flow-secondary {
            stroke: currentColor;
            stroke-width: 1;
            stroke-opacity: .16;
        }

        .flow-dash {
            stroke: currentColor;
            stroke-width: 1;
            stroke-dasharray: 3 10;
            stroke-opacity: .22;
        }

        .flow-signal {
            stroke: currentColor;
            stroke-width: 2.1;
            stroke-dasharray: 15 190;
            stroke-opacity: .48;
            animation: heroSignal 16s linear infinite;
        }

        .node-halo {
            fill: currentColor;
            fill-opacity: .055;
            stroke: currentColor;
            stroke-opacity: .13;
        }

        .node-core {
            fill: currentColor;
            fill-opacity: .52;
        }

        .flow-scene {
            transform-box: fill-box;
            transform-origin: center;
            animation: heroDrift 16s ease-in-out infinite;
        }

        .flow-node-a,
        .flow-node-c {
            animation: heroNode 7s ease-in-out infinite;
        }

        .flow-node-c {
            animation-delay: -3.5s;
        }

        .setting-orbit,
        .setting-trace {
            fill: none;
            stroke: currentColor;
            stroke-linecap: round;
        }

        .setting-orbit {
            stroke-dasharray: 4 10;
            stroke-opacity: .16;
        }

        .setting-trace {
            stroke-width: 1.3;
        }

        .setting-sheet {
            fill: currentColor;
            fill-opacity: .04;
            stroke: currentColor;
            stroke-opacity: .18;
        }

        .setting-sheet.front {
            fill-opacity: .09;
            stroke-opacity: .34;
        }

        .setting-chip {
            fill: currentColor;
            fill-opacity: .3;
        }

        .setting-chip.muted {
            fill-opacity: .1;
        }

        .setting-point-halo {
            fill: currentColor;
            fill-opacity: .045;
            stroke: currentColor;
            stroke-opacity: .12;
        }

        .setting-point-core {
            fill: currentColor;
            fill-opacity: .42;
        }

        .setting-stack {
            transform-box: fill-box;
            transform-origin: center;
            animation: settingFloat 18s ease-in-out infinite;
        }

        .setting-point-a,
        .setting-point-b {
            animation: heroNode 8s ease-in-out infinite;
        }

        .setting-point-b {
            animation-delay: -4s;
        }

        @keyframes heroSignal {
            to {
                stroke-dashoffset: -205;
            }
        }

        @keyframes heroDrift {
            0%, 100% {
                transform: translate3d(0, 0, 0);
            }
            50% {
                transform: translate3d(6px, -3px, 0);
            }
        }

        @keyframes heroNode {
            0%, 100% {
                opacity: .62;
            }
            50% {
                opacity: 1;
            }
        }

        @keyframes settingFloat {
            0%, 100% {
                transform: translate3d(0, 0, 0);
            }
            50% {
                transform: translate3d(-4px, 3px, 0);
            }
        }

        @media (prefers-reduced-motion: reduce) {
            .flow-scene,
            .flow-signal,
            .flow-node-a,
            .flow-node-c,
            .setting-stack,
            .setting-point-a,
            .setting-point-b {
                animation: none;
            }
        }
    `])),t.colorPrimary)}}),cn=function(n){var t=n.styles,r=n.id;return(0,e.jsxs)("svg",{className:"".concat(t.graphic," hero-flow-graphic"),viewBox:"0 0 480 132","aria-hidden":"true",focusable:"false",children:[(0,e.jsxs)("defs",{children:[(0,e.jsxs)("linearGradient",{id:"".concat(r,"-flow"),x1:"38",y1:"90",x2:"456",y2:"52",gradientUnits:"userSpaceOnUse",children:[(0,e.jsx)("stop",{stopColor:"currentColor",stopOpacity:"0"}),(0,e.jsx)("stop",{offset:".24",stopColor:"currentColor",stopOpacity:".24"}),(0,e.jsx)("stop",{offset:".58",stopColor:"currentColor",stopOpacity:".7"}),(0,e.jsx)("stop",{offset:"1",stopColor:"currentColor",stopOpacity:".04"})]}),(0,e.jsxs)("radialGradient",{id:"".concat(r,"-aura"),cx:"0",cy:"0",r:"1",gradientTransform:"translate(322 63) rotate(90) scale(60 132)",gradientUnits:"userSpaceOnUse",children:[(0,e.jsx)("stop",{stopColor:"currentColor",stopOpacity:".12"}),(0,e.jsx)("stop",{offset:".55",stopColor:"currentColor",stopOpacity:".035"}),(0,e.jsx)("stop",{offset:"1",stopColor:"currentColor",stopOpacity:"0"})]})]}),(0,e.jsxs)("g",{className:"flow-scene",children:[(0,e.jsx)("ellipse",{className:"flow-aura",cx:"322",cy:"63",rx:"132",ry:"60",fill:"url(#".concat(r,"-aura)")}),(0,e.jsx)("path",{className:"flow-line flow-secondary flow-detail",d:"M34 108C116 87 139 29 229 35S346 107 468 52"}),(0,e.jsx)("path",{className:"flow-line flow-dash flow-detail",d:"M60 30C143 28 174 84 255 86S371 38 454 72"}),(0,e.jsx)("path",{className:"flow-line flow-primary",stroke:"url(#".concat(r,"-flow)"),d:"M26 88C106 84 143 49 215 51S324 94 457 58"}),(0,e.jsx)("path",{className:"flow-line flow-signal",d:"M26 88C106 84 143 49 215 51S324 94 457 58"}),(0,e.jsxs)("g",{className:"flow-node flow-node-a",transform:"translate(154 57)",children:[(0,e.jsx)("circle",{className:"node-halo",r:"13"}),(0,e.jsx)("circle",{className:"node-core",r:"3.4"})]}),(0,e.jsxs)("g",{className:"flow-node flow-node-b flow-secondary",transform:"translate(282 78)",children:[(0,e.jsx)("circle",{className:"node-halo",r:"10"}),(0,e.jsx)("circle",{className:"node-core",r:"2.8"})]}),(0,e.jsxs)("g",{className:"flow-node flow-node-c",transform:"translate(390 67)",children:[(0,e.jsx)("circle",{className:"node-halo",r:"14"}),(0,e.jsx)("circle",{className:"node-core",r:"3.6"})]})]})]})},dn=function(n){var t=n.styles,r=n.id;return(0,e.jsxs)("svg",{className:"".concat(t.graphic," hero-setting-graphic"),viewBox:"0 0 480 132","aria-hidden":"true",focusable:"false",children:[(0,e.jsxs)("defs",{children:[(0,e.jsxs)("linearGradient",{id:"".concat(r,"-edge"),x1:"76",y1:"101",x2:"438",y2:"28",gradientUnits:"userSpaceOnUse",children:[(0,e.jsx)("stop",{stopColor:"currentColor",stopOpacity:"0"}),(0,e.jsx)("stop",{offset:".44",stopColor:"currentColor",stopOpacity:".28"}),(0,e.jsx)("stop",{offset:"1",stopColor:"currentColor",stopOpacity:".04"})]}),(0,e.jsxs)("radialGradient",{id:"".concat(r,"-aura"),cx:"0",cy:"0",r:"1",gradientTransform:"translate(316 64) rotate(90) scale(60 138)",gradientUnits:"userSpaceOnUse",children:[(0,e.jsx)("stop",{stopColor:"currentColor",stopOpacity:".11"}),(0,e.jsx)("stop",{offset:".62",stopColor:"currentColor",stopOpacity:".03"}),(0,e.jsx)("stop",{offset:"1",stopColor:"currentColor",stopOpacity:"0"})]})]}),(0,e.jsx)("ellipse",{className:"setting-aura",cx:"316",cy:"64",rx:"138",ry:"60",fill:"url(#".concat(r,"-aura)")}),(0,e.jsx)("path",{className:"setting-orbit setting-detail",d:"M70 102C139 77 175 31 254 30S370 77 454 43"}),(0,e.jsx)("path",{className:"setting-trace",stroke:"url(#".concat(r,"-edge)"),d:"M96 93C163 86 193 63 238 63"}),(0,e.jsxs)("g",{className:"setting-stack",children:[(0,e.jsx)("rect",{className:"setting-sheet setting-detail",x:"238",y:"23",width:"142",height:"82",rx:"22",transform:"rotate(-7 309 64)"}),(0,e.jsx)("rect",{className:"setting-sheet setting-detail",x:"270",y:"22",width:"132",height:"84",rx:"22",transform:"rotate(6 336 64)"}),(0,e.jsx)("rect",{className:"setting-sheet front",x:"250",y:"27",width:"146",height:"78",rx:"21"}),(0,e.jsx)("rect",{className:"setting-chip",x:"270",y:"46",width:"42",height:"8",rx:"4"}),(0,e.jsx)("rect",{className:"setting-chip muted",x:"321",y:"46",width:"54",height:"8",rx:"4"}),(0,e.jsx)("rect",{className:"setting-chip muted",x:"270",y:"64",width:"68",height:"8",rx:"4"}),(0,e.jsx)("rect",{className:"setting-chip",x:"347",y:"64",width:"28",height:"8",rx:"4"}),(0,e.jsx)("rect",{className:"setting-chip muted",x:"270",y:"82",width:"32",height:"8",rx:"4"}),(0,e.jsx)("rect",{className:"setting-chip muted",x:"311",y:"82",width:"64",height:"8",rx:"4"})]}),(0,e.jsxs)("g",{className:"setting-point setting-point-a",transform:"translate(181 75)",children:[(0,e.jsx)("circle",{className:"setting-point-halo",r:"12"}),(0,e.jsx)("circle",{className:"setting-point-core",r:"3.2"})]}),(0,e.jsxs)("g",{className:"setting-point setting-point-b setting-detail",transform:"translate(430 38)",children:[(0,e.jsx)("circle",{className:"setting-point-halo",r:"9"}),(0,e.jsx)("circle",{className:"setting-point-core",r:"2.6"})]})]})},pn=function(n){var t=n.variant,r=t===void 0?"task":t,i=un(),l=i.styles,c=s.useId().replace(/:/g,"");return r==="setting"?(0,e.jsx)(dn,{styles:l,id:"setting-".concat(c)}):(0,e.jsx)(cn,{styles:l,id:"task-".concat(c)})},mn=s.memo(pn),Re,Be,ze,Oe,Te,Ie,Me,$e,Pe,Ge,Le,Ve,Ue,He,fn=(0,Ee.rU)(function(o){var n=o.css,t=o.token,r=o.isDarkMode,i=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{},l=!!(i!=null&&i.isCompactMode),c=typeof(i==null?void 0:i.isDarkMode)=="boolean"?i.isDarkMode:!!r;return{page:n(Re||(Re=D()([`
            width: 100%;
            max-width: 1450px;
            margin: 0 auto;

            > .ant-col {
                min-width: 0;
            }
        `]))),workspace:n(Be||(Be=D()([`
            container-name: setting-workspace;
            container-type: inline-size;
            min-width: 0;
            overflow: hidden;
            border-radius: `,`px;
            background: `,`;
            box-shadow: `,`;
        `])),t.borderRadiusLG,t.colorBgContainer,t.boxShadowTertiary),hero:n(ze||(ze=D()([`
            position: relative;
            min-height: `,`px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            overflow: hidden;
            isolation: isolate;
            background: linear-gradient(
                118deg,
                color-mix(in srgb, `,", "," ",`%) 0%,
                `,` 58%,
                color-mix(in srgb, #13c2c2, `," ",`%) 100%
            );

            &::before {
                position: absolute;
                z-index: 0;
                inset: 0;
                background-image:
                    linear-gradient(`,` 1px, transparent 1px),
                    linear-gradient(90deg, `,` 1px, transparent 1px);
                background-size: 24px 24px;
                content: "";
                -webkit-mask-image: linear-gradient(90deg, transparent 36%, rgba(0, 0, 0, .2) 58%, #000 100%);
                mask-image: linear-gradient(90deg, transparent 36%, rgba(0, 0, 0, .2) 58%, #000 100%);
                pointer-events: none;
            }

            &::after {
                position: absolute;
                z-index: 0;
                inset: -45% -10% -55% 48%;
                background:
                    radial-gradient(ellipse at 48% 32%, color-mix(in srgb, `,`, transparent 84%) 0%, transparent 60%),
                    radial-gradient(ellipse at 76% 72%, color-mix(in srgb, #13c2c2, transparent 89%) 0%, transparent 58%);
                content: "";
                filter: blur(28px);
                opacity: `,`;
                pointer-events: none;
            }

            .hero-copy {
                position: relative;
                z-index: 2;
                min-width: 0;
                max-width: calc(100% - 178px);
                padding: `,`;
            }

            .eyebrow {
                display: flex;
                align-items: center;
                gap: 7px;
                color: color-mix(in srgb, `,", ",` 62%);
                font-size: 9px;
                font-weight: 600;
                line-height: 16px;
                letter-spacing: 0;
            }

            .eyebrow .status-dot {
                width: 5px;
                height: 5px;
                flex: none;
                border-radius: 50%;
                background: color-mix(in srgb, `,`, #fff 26%);
                box-shadow: 0 0 0 4px color-mix(in srgb, `,`, transparent 90%);
            }

            h1 {
                margin: 3px 0 0;
                color: `,`;
                font-size: `,`px;
                font-weight: 600;
                line-height: `,`px;
                letter-spacing: 0;
            }

            .hero-visual {
                position: absolute;
                z-index: 1;
                top: 0;
                right: 82px;
                width: 52%;
                height: 100%;
                opacity: `,`;
                filter: `,`;
                -webkit-mask-image: linear-gradient(90deg, transparent 0%, rgba(0, 0, 0, .38) 28%, #000 52%);
                mask-image: linear-gradient(90deg, transparent 0%, rgba(0, 0, 0, .38) 28%, #000 52%);
                pointer-events: none;
            }

            @container setting-workspace (max-width: 600px) {
                min-height: `,`px;

                .hero-copy {
                    max-width: calc(100% - 145px);
                    padding: 16px 14px;
                }

                h1 {
                    font-size: 18px;
                    line-height: 25px;
                }

                .hero-visual {
                    right: 42px;
                    width: 68%;
                    opacity: `,`;
                }

                .hero-visual .flow-secondary,
                .hero-visual .flow-detail {
                    display: none;
                }
            }

            @container setting-workspace (max-width: 430px) {
                .hero-visual {
                    right: 16px;
                    width: 76%;
                    opacity: `,`;
                }

                .hero-visual .flow-node-a {
                    display: none;
                }
            }

            @container setting-workspace (max-width: 360px) {
                .hero-visual {
                    display: none;
                }
            }
        `])),l?104:118,t.colorPrimary,t.colorBgContainer,c?94:93,t.colorBgContainer,t.colorBgContainer,c?97:96,c?"rgba(255,255,255,0.018)":"rgba(72,132,202,0.035)",c?"rgba(255,255,255,0.018)":"rgba(72,132,202,0.035)",t.colorPrimary,c?.34:.52,l?"14px 18px":"18px 22px",t.colorPrimary,t.colorTextTertiary,t.colorPrimary,t.colorPrimary,t.colorTextHeading,l?19:20,l?26:28,c?.12:.25,c?"blur(.8px) saturate(.78)":"blur(.25px)",l?104:116,c?.07:.14,c?.045:.09),settingsMain:n(Oe||(Oe=D()([`
            width: 100%;
            height: 100%;
            padding: `,`px;
            display: flex;
            box-sizing: border-box;

            .ant-menu-light.ant-menu-inline .ant-menu-item::after {
                top: 19%;
                right: 6px;
                height: 60%;
                border-top-left-radius: 15px;
                border-bottom-left-radius: 15px;
            }

            @container setting-workspace (max-width: 760px) {
                flex-direction: column;
                padding: `,`px;
            }
        `])),t.paddingLG,l?10:12),leftMenu:n(Te||(Te=D()([`
            width: 224px;
            flex: none;

            .ant-menu-item {
                border-radius: `,`px;
                font-size: `,`px;
                font-weight: bold;
            }

            @container setting-workspace (max-width: 760px) {
                width: 100%;
                margin-bottom: 10px;
                overflow: hidden;
                border: none;
            }
        `])),t.borderRadius,t.fontSize),menuScroll:n(Ie||(Ie=D()([`
            min-width: 0;

            @container setting-workspace (max-width: 760px) {
                width: 100%;
                overflow-x: auto;
                overflow-y: hidden;
                overscroll-behavior-x: contain;
                scrollbar-width: none;
                touch-action: pan-x;
                -webkit-overflow-scrolling: touch;

                &::-webkit-scrollbar {
                    display: none;
                }
            }
        `]))),menu:n(Me||(Me=D()([`
            border-inline-end: none !important;

            &:focus,
            &:focus-visible {
                outline: none !important;
            }

            @container setting-workspace (max-width: 760px) {
                display: flex !important;
                flex-direction: row;
                flex-wrap: nowrap;
                gap: 6px;
                width: max-content;
                min-width: 100%;
                padding: 0;
                background: transparent;

                .ant-menu-item {
                    width: auto !important;
                    min-width: 96px;
                    height: 34px;
                    min-height: 34px;
                    margin: 0 !important;
                    padding-inline: 12px !important;
                    display: flex;
                    flex: 0 0 auto;
                    align-items: center;
                    justify-content: center;
                    background: `,`;
                    color: `,`;
                    font-weight: 400;
                    line-height: 34px;
                    white-space: nowrap;
                }

                .ant-menu-item::after {
                    display: none !important;
                }

                .ant-menu-item-selected {
                    background: `,` !important;
                    color: `,` !important;
                    font-weight: 600;
                }
            }
        `])),t.colorFillQuaternary,t.colorTextSecondary,t.colorPrimaryBg,t.colorTextHeading),right:n($e||($e=D()([`
            min-width: 0;
            flex: 1;
            margin-left: 1px;
            padding: 8px 40px;
            border-left: 1px solid `,`;

            @container setting-workspace (max-width: 760px) {
                margin-left: 0;
                padding: 10px;
                border-left: none;
            }
        `])),t.colorSplit),title:n(Pe||(Pe=D()([`
            margin-bottom: 20px;
            color: `,`;
            font-size: `,`px;
            font-weight: bolder;
            line-height: 28px;
        `])),t.colorTextHeading,t.fontSizeHeading4),loading:n(Ge||(Ge=D()([`
            width: 100%;
            min-height: 240px;
            display: flex;
            align-items: center;
            justify-content: center;
        `]))),uiWrapper:n(Le||(Le=D()([`
            width: 100%;
            max-width: 720px;
        `]))),formField:n(Ve||(Ve=D()([`
            width: 100%;
            max-width: `,`px;
        `])),l?300:350),formActions:n(Ue||(Ue=D()([`
            margin-top: `,`px;
            display: flex;
            align-items: center;
            gap: `,`px;
        `])),t.marginLG,t.marginSM),errorState:n(He||(He=D()([`
            min-height: 240px;
            padding: 24px;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: `,`px;
            color: `,`;
            font-size: `,`px;
            text-align: center;

            > .anticon {
                color: `,`;
                font-size: `,`px;
            }
        `])),t.marginSM,t.colorTextSecondary,t.fontSize,t.colorError,t.fontSizeHeading3)}}),be=[{key:"web",title:"\u7F51\u7AD9\u8BBE\u7F6E",icon:"GlobalOutlined",component:Se},{key:"ui",title:"\u754C\u9762\u8BBE\u7F6E",icon:"LayoutOutlined",component:Ae},{key:"account",title:"\u767B\u5F55\u8BBE\u7F6E",icon:"SafetyCertificateOutlined",component:je},{key:"timezone",title:"\u65F6\u533A\u8BBE\u7F6E",icon:"ClockCircleOutlined",component:Ne}],We=(function(){var o,n,t,r=(0,ne.DP)(),i=(0,W.fn)("web"),l=(0,W.zy)(),c=(r==null||(o=r.isCompactTheme)===null||o===void 0?void 0:o.call(r))||!1,N=!!(!(r==null||(n=r.isDarkMode)===null||n===void 0)&&n.call(r)||!(r==null||(t=r.isDarkTheme)===null||t===void 0)&&t.call(r)),E=fn({isCompactMode:c,isDarkMode:N}),F=E.styles,G=new URLSearchParams(l.search).get("tab")||"web",$=be.find(function(R){return R.key===G})||be[0],L=$.key,J=$.component,V=function(I){var M=new URLSearchParams(l.search);M.set("tab",I),W.b6.replace({pathname:l.pathname,search:"?".concat(M.toString())})};return(0,e.jsx)(ne.nB,{children:(0,e.jsx)(ue.A,{className:F.page,gutter:[0,c?10:12],children:(0,e.jsx)(se.A,{span:24,children:(0,e.jsxs)("section",{className:F.workspace,children:[(0,e.jsxs)("section",{className:F.hero,children:[(0,e.jsxs)("div",{className:"hero-copy",children:[(0,e.jsxs)("div",{className:"eyebrow",children:[(0,e.jsx)("span",{className:"status-dot"}),"SYSTEM SETTINGS"]}),(0,e.jsx)("h1",{children:"\u7CFB\u7EDF\u8BBE\u7F6E"})]}),(0,e.jsx)("div",{className:"hero-visual",children:(0,e.jsx)(mn,{variant:"task"})})]}),(0,e.jsxs)("div",{className:F.settingsMain,children:[(0,e.jsx)("div",{className:F.leftMenu,children:(0,e.jsx)("div",{className:F.menuScroll,children:(0,e.jsx)(ve.A,{className:F.menu,mode:"inline",selectedKeys:[L],onClick:function(I){var M=I.key;return V(M)},items:be.map(function(R){return{key:R.key,label:R.title,icon:(0,e.jsx)(ne.In,{type:R.icon})}})})})}),(0,e.jsxs)("section",{className:F.right,children:[(0,e.jsx)("div",{className:F.title,children:$.title}),(0,e.jsx)(J,{styles:F,info:i==null?void 0:i.info})]})]})]})})})})});a.dn(We)}),41162:(function(j,Z,a){var s=a(44814);function ue(se){if(Array.isArray(se))return s(se)}j.exports=ue,j.exports.__esModule=!0,j.exports.default=j.exports}),33032:(function(j){function Z(a){if(typeof Symbol!="undefined"&&a[Symbol.iterator]!=null||a["@@iterator"]!=null)return Array.from(a)}j.exports=Z,j.exports.__esModule=!0,j.exports.default=j.exports}),30922:(function(j){function Z(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}j.exports=Z,j.exports.__esModule=!0,j.exports.default=j.exports}),45647:(function(j,Z,a){var s=a(41162),ue=a(33032),se=a(34911),ve=a(30922);function ne(W){return s(W)||ue(W)||se(W)||ve()}j.exports=ne,j.exports.__esModule=!0,j.exports.default=j.exports})}]);
