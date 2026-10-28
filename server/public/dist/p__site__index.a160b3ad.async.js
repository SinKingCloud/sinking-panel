!(function(){var md=Object.defineProperty,vd=Object.defineProperties;var hd=Object.getOwnPropertyDescriptors;var wa=Object.getOwnPropertySymbols;var vs=Object.prototype.hasOwnProperty,hs=Object.prototype.propertyIsEnumerable;var ms=(_n,Rn,X)=>Rn in _n?md(_n,Rn,{enumerable:!0,configurable:!0,writable:!0,value:X}):_n[Rn]=X,bt=(_n,Rn)=>{for(var X in Rn||(Rn={}))vs.call(Rn,X)&&ms(_n,X,Rn[X]);if(wa)for(var X of wa(Rn))hs.call(Rn,X)&&ms(_n,X,Rn[X]);return _n},fs=(_n,Rn)=>vd(_n,hd(Rn));var gs=(_n,Rn)=>{var X={};for(var Un in _n)vs.call(_n,Un)&&Rn.indexOf(Un)<0&&(X[Un]=_n[Un]);if(_n!=null&&wa)for(var Un of wa(_n))Rn.indexOf(Un)<0&&hs.call(_n,Un)&&(X[Un]=_n[Un]);return X};(self.webpackChunk=self.webpackChunk||[]).push([[664],{14352:(function(_n,Rn,X){"use strict";X.r(Rn),X.d(Rn,{default:function(){return ls}});var Un=X(1974),D=X.n(Un),tr=X(11604),C=X.n(tr),i=X(96540),ce=X(62885),rt=X(9752),tt=X(8490),ar=X(9300),at=X(45647),Je=X.n(at),e=X(74848),Or=function(n){var t=n.keyword,r=n.searchField,a=n.typeId,l=n.type,h=n.status,c=n.typeItems,f=n.typeData,y=n.statusData,j=n.refreshing,F=n.onKeywordChange,B=n.onSearchFieldChange,o=n.onSearch,v=n.onTypeIdChange,N=n.onManageTypes,w=n.onTypeChange,u=n.onStatusChange,S=n.onRefresh,x=n.onCreate,m=n.onHTTPSettings,p=n.onManageCert,_=n.onManageSecret,A=i.useMemo(function(){return[{label:"\u5168\u90E8\u5206\u7C7B",value:"0"}].concat(Je()((c||[]).map(function(T){return{value:String(T.id),label:String(T.name)}})))},[c]),E=i.useMemo(function(){return[{label:"\u5168\u90E8\u7C7B\u578B",value:"all"}].concat(Je()(Object.entries(f||{}).map(function(T){var U=C()(T,2),P=U[0],M=U[1];return{value:P,label:String(M)}})))},[f]),k=i.useMemo(function(){return[{label:"\u5168\u90E8\u72B6\u6001",value:"all"}].concat(Je()(Object.entries(y||{}).map(function(T){var U=C()(T,2),P=U[0],M=U[1];return{value:P,label:String(M)}})))},[y]);return{hero:{eyebrow:"WEBSITE MANAGER",title:"\u7F51\u7AD9\u7BA1\u7406",action:{label:"\u6DFB\u52A0\u7F51\u7AD9",onClick:x}},toolbar:{search:{value:t,field:r,maxLength:100,fields:[{value:"keyword",label:"\u5168\u90E8",placeholder:"\u641C\u7D22\u540D\u79F0\u6216\u76EE\u5F55"},{value:"name",label:"\u540D\u79F0",placeholder:"\u641C\u7D22\u7F51\u7AD9\u540D\u79F0"},{value:"root",label:"\u76EE\u5F55",placeholder:"\u641C\u7D22\u7F51\u7AD9\u76EE\u5F55"}],onChange:F,onFieldChange:B,onSearch:o},refresh:{loading:j,ariaLabel:"\u5237\u65B0\u7F51\u7AD9\u5217\u8868",onClick:S},actions:[{key:"category",type:"filter",value:a||"0",options:A,icon:"FolderOutlined",ariaLabel:"\u7F51\u7AD9\u5206\u7C7B\u7B5B\u9009",matchWidth:!0,command:{label:"\u5206\u7C7B\u7BA1\u7406",icon:"SettingOutlined",onClick:N},onChange:v},{key:"type",type:"filter",value:l||"all",options:E,icon:"AppstoreOutlined",ariaLabel:"\u7F51\u7AD9\u7C7B\u578B\u7B5B\u9009",matchWidth:!0,onChange:function(U){return w(U==="all"?"":U)}},{key:"status",type:"filter",value:h||"all",options:k,icon:"FlagOutlined",ariaLabel:"\u7F51\u7AD9\u72B6\u6001\u7B5B\u9009",matchWidth:!0,onChange:function(U){return u(U==="all"?"":U)}},{key:"settings",type:"button",label:"\u670D\u52A1\u7BA1\u7406",icon:"SettingOutlined",suffixIcon:"DownOutlined",menu:{items:[{key:"cert",label:"\u8BC1\u4E66\u7BA1\u7406",icon:(0,e.jsx)(ce.In,{type:"SafetyCertificateOutlined"}),onClick:p},{key:"secret",label:"\u5BC6\u94A5\u7BA1\u7406",icon:(0,e.jsx)(ce.In,{type:"KeyOutlined"}),onClick:_}].concat(Je()(m?[{key:"http",label:"\u670D\u52A1\u8BBE\u7F6E",icon:(0,e.jsx)(ce.In,{type:"SettingOutlined"}),onClick:m}]:[]))}}]}}},ot=Or,Kn=X(88826),ge=X(50400),gn=X(57798),on=X(23455),Ye=X(63496),Te=function(n){var t=n.className,r=n.sites,a=n.loading,l=n.categoryData,h=n.typeData,c=n.statusData,f=n.sort,y=n.order,j=n.operating,F=n.onSortChange,B=n.onSettings,o=n.onAction,v=(0,i.useState)(""),N=C()(v,2),w=N[0],u=N[1],S=(0,i.useMemo)(function(){return[{title:"\u7F51\u7AD9\u540D\u79F0",dataIndex:"name",key:"name",width:190,className:"site-name-cell",sorter:!0,sortOrder:f==="name"?y==="asc"?"ascend":"descend":null,render:function(m,p){return(0,e.jsx)("div",{className:"site-name",children:(0,e.jsx)("button",{type:"button",title:p.name||"\u672A\u547D\u540D\u7F51\u7AD9",onClick:function(){return B(p,"basic")},children:p.name||"\u672A\u547D\u540D\u7F51\u7AD9"})})}},{title:"\u5206\u7C7B",dataIndex:"type_id",key:"type_id",width:110,className:"category-cell",sorter:!0,sortOrder:f==="type_id"?y==="asc"?"ascend":"descend":null,render:function(m){var p=Number(m)>0?l[String(m)]||"\u672A\u77E5\u5206\u7C7B":"\u5168\u90E8\u5206\u7C7B";return(0,e.jsx)(Kn.A,{title:p,children:(0,e.jsx)("span",{className:"site-category",children:p})})}},{title:"\u7F51\u7AD9\u7C7B\u578B",dataIndex:"type",key:"type",width:110,className:"type-cell",sorter:!0,sortOrder:f==="type"?y==="asc"?"ascend":"descend":null,render:function(m){return(0,e.jsx)("span",{className:"site-type",children:h[String(m)]||"\u672A\u77E5\u7C7B\u578B"})}},{title:"\u7F51\u7AD9\u76EE\u5F55",dataIndex:"root",key:"root",width:220,className:"root-cell",render:function(m){var p=String(m||"").trim();return p?(0,e.jsx)(Kn.A,{title:p,children:(0,e.jsx)("button",{type:"button",className:"site-root site-root-link",onClick:function(){return(0,on.PC)("".concat((0,on.jO)("file"),"?").concat(new URLSearchParams({path:p})))},children:p})}):(0,e.jsx)("span",{className:"site-root",children:"-"})}},{title:"\u8FD0\u884C\u72B6\u6001",dataIndex:"status",key:"status",width:100,className:"status-cell",sorter:!0,sortOrder:f==="status"?y==="asc"?"ascend":"descend":null,render:function(m){var p=Number(m)===0;return(0,e.jsxs)("span",{className:"site-state ".concat(p?"running":"stopped"),children:[(0,e.jsx)("i",{}),c[String(m)]||(p?"\u5DF2\u542F\u7528":"\u5DF2\u505C\u7528")]})}},{title:"\u66F4\u65B0\u65F6\u95F4",dataIndex:"update_time",key:"update_time",width:130,className:"update-cell",sorter:!0,sortOrder:f==="update_time"?y==="asc"?"ascend":"descend":null,render:function(m){var p=String(m||"");return(0,e.jsx)(Kn.A,{title:p||"-",children:(0,e.jsx)("span",{className:"update-time",children:p?(0,Ye.K)(p)||p:"-"})})}},{title:"\u64CD\u4F5C",key:"action",width:82,fixed:"right",className:"action-cell",render:function(m,p){var _=String(p.id),A=Number(p.status)===0,E=[{key:"settings",label:"\u7F51\u7AD9\u8BBE\u7F6E",onClick:function(){return B(p,"basic")}},{key:"domain",label:"\u57DF\u540D\u7BA1\u7406",onClick:function(){return B(p,"domain")}},{key:"ssl",label:"\u8BC1\u4E66\u8BBE\u7F6E",onClick:function(){return B(p,"ssl")}},{key:"log",label:"\u7F51\u7AD9\u65E5\u5FD7",onClick:function(){return B(p,"log")}},{type:"divider"},A?{key:"disable",label:"\u505C\u7528\u7F51\u7AD9",onClick:function(){return o("disable",p)}}:{key:"enable",label:"\u542F\u7528\u7F51\u7AD9",onClick:function(){return o("enable",p)}},{key:"delete",label:"\u5220\u9664\u7F51\u7AD9",danger:!0,onClick:function(){return o("delete",p)}}];return(0,e.jsx)(gn.A,{open:w===_,onOpenChange:function(T){return u(T?_:"")},menu:{items:E},children:(0,e.jsx)(ge.A,{size:"small",loading:j==null?void 0:j.has(_),children:"\u64CD\u4F5C"})})}}]},[l,o,B,w,j,y,f,c,h]);return{className:t,columns:S,dataSource:r,loading:a,onSortChange:F,rowKey:function(m){return String(m.id)},rowClassName:function(m){return w===String(m.id)?"action-menu-open":""}}},yn=Te,Mn=X(55420),b=X.n(Mn),ur=X(98312),ne=X.n(ur),Ke=X(84169),H=X.n(Ke),un=X(79273),Sn=X(40372),d=X(20873),ze=X(22817),na=X(66291),dt=X(51643),yt=X(47152),ct=X(16370),qe=X(48212),Xn=X(21612),ra=X(10662),De=X(58679),Dr=X(79508),kr=X(12185),Br=X(67064),dn=X(78781);function Aa(){return Tt.apply(this,arguments)}function Tt(){return Tt=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.Jt)("/site/list",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Tt.apply(this,arguments)}function Ea(){return Ft.apply(this,arguments)}function Ft(){return Ft=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.Jt)("/site/info",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ft.apply(this,arguments)}function ta(){return Ca.apply(this,arguments)}function Ca(){return Ca=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/create",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ca.apply(this,arguments)}function xs(){return ja.apply(this,arguments)}function ja(){return ja=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/update",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),ja.apply(this,arguments)}function bs(){return Sa.apply(this,arguments)}function Sa(){return Sa=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/delete",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Sa.apply(this,arguments)}function ys(){return Da.apply(this,arguments)}function Da(){return Da=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/enable",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Da.apply(this,arguments)}function Fs(){return ka.apply(this,arguments)}function ka(){return ka=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/disable",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),ka.apply(this,arguments)}function Ri(){return Ba.apply(this,arguments)}function Ba(){return Ba=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/log",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ba.apply(this,arguments)}function Na(){return Ta.apply(this,arguments)}function Ta(){return Ta=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/domain",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ta.apply(this,arguments)}function Ii(){return _a.apply(this,arguments)}function _a(){return _a=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/ssl",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),_a.apply(this,arguments)}function Pi(){return Ra.apply(this,arguments)}function Ra(){return Ra=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/waf",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ra.apply(this,arguments)}function Ia(){return Pa.apply(this,arguments)}function Pa(){return Pa=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/cache",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Pa.apply(this,arguments)}function zi(){return za.apply(this,arguments)}function za(){return za=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/rate",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),za.apply(this,arguments)}function Oi(){return Oa.apply(this,arguments)}function Oa(){return Oa=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/traffic",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Oa.apply(this,arguments)}function $i(){return $a.apply(this,arguments)}function $a(){return $a=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/header",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),$a.apply(this,arguments)}function Mi(){return Ma.apply(this,arguments)}function Ma(){return Ma=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/compression",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ma.apply(this,arguments)}function Hi(){return Ha.apply(this,arguments)}function Ha(){return Ha=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/redirect",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ha.apply(this,arguments)}function Li(){return La.apply(this,arguments)}function La(){return La=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/route",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),La.apply(this,arguments)}function Vi(){return Va.apply(this,arguments)}function Va(){return Va=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/static",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Va.apply(this,arguments)}function aa(){return qa.apply(this,arguments)}function qa(){return qa=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/proxy",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),qa.apply(this,arguments)}function qi(){return Wa.apply(this,arguments)}function Wa(){return Wa=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/fastcgi",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Wa.apply(this,arguments)}function ia(){return Ua.apply(this,arguments)}function Ua(){return Ua=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/site/process",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ua.apply(this,arguments)}var Nr=function(n){var t=arguments.length>1&&arguments[1]!==void 0?arguments[1]:"http",r=String(n||"").trim();if(!r||/[\r\n\0]/.test(r))throw new Error("\u8BF7\u8F93\u5165\u6709\u6548\u7684\u76EE\u6807\u5730\u5740");var a=/^unix(?:\+h2c)?\//i.test(r)||r.startsWith("/");if(a){var l=r.replace(/^unix(?:\+h2c)?\//i,"");if(!l.startsWith("/")||l==="/")throw new Error("UNIX \u5957\u63A5\u5B57\u5FC5\u987B\u586B\u5199\u7EDD\u5BF9\u6587\u4EF6\u8DEF\u5F84\uFF0C\u4F8B\u5982 /tmp/panel.sock");return{dial:"unix/".concat(l),scheme:"http",hostname:"",h2c:/^unix\+h2c\//i.test(r)}}try{var h=r.includes("://")?r:"".concat(t,"://").concat(r);if(!/^https?:\/\/[^/\\]/i.test(h)||r.includes("\\"))throw new Error;var c=new URL(h);if(!c.hostname||c.username||c.password||c.pathname!=="/"||c.search||c.hash)throw new Error;return{dial:"".concat(c.protocol,"//").concat(c.host),scheme:c.protocol.slice(0,-1),hostname:c.hostname,h2c:!1}}catch(f){throw new Error("\u8BF7\u8F93\u5165 HTTP\u3001HTTPS \u5730\u5740\u6216 UNIX \u5957\u63A5\u5B57\u8DEF\u5F84\uFF0C\u7F51\u5740\u4E0D\u80FD\u5305\u542B\u8DEF\u5F84\u548C\u53C2\u6570")}},Wi,Ui,Gi,Qi,Ki,Xi,Ji,Yi,Zi,eu,Ga=[{label:"\u9759\u6001\u7F51\u7AD9",value:0,description:"\u5DF2\u6709 HTML \u7F51\u9875\u6216\u524D\u7AEF\u6253\u5305\u540E\u7684\u6587\u4EF6",hint:"\u586B\u5199\u8BBF\u95EE\u57DF\u540D\uFF0C\u518D\u9009\u62E9\u7F51\u9875\u6587\u4EF6\u6240\u5728\u7684\u76EE\u5F55\u3002"},{label:"\u53CD\u5411\u4EE3\u7406",value:1,description:"\u5DF2\u6709\u670D\u52A1\u6B63\u5728\u8FD0\u884C\uFF0C\u5E0C\u671B\u901A\u8FC7\u57DF\u540D\u8BBF\u95EE",hint:"\u586B\u5199\u8BBF\u95EE\u57DF\u540D\uFF0C\u4EE5\u53CA\u9700\u8981\u8F6C\u53D1\u5230\u7684\u670D\u52A1\u5730\u5740\u3002"},{label:"\u901A\u7528\u7F51\u7AD9",value:2,description:"\u8FD0\u884C\u672C\u673A\u7A0B\u5E8F\uFF0C\u7531\u9762\u677F\u8D1F\u8D23\u542F\u52A8\u548C\u4FDD\u6D3B",hint:"\u9009\u62E9\u8981\u8FD0\u884C\u7684\u7A0B\u5E8F\uFF1B\u4E0D\u7ED1\u5B9A\u57DF\u540D\u4E5F\u53EF\u4EE5\u521B\u5EFA\u3002"},{label:"PHP \u7F51\u7AD9",value:3,description:"\u90E8\u7F72 PHP \u9879\u76EE\uFF0C\u9700\u5DF2\u6709 PHP-FPM \u73AF\u5883",hint:"\u586B\u5199\u8BBF\u95EE\u57DF\u540D\uFF0C\u9009\u62E9 PHP \u6587\u4EF6\u76EE\u5F55\u5E76\u8FDE\u63A5 PHP-FPM\u3002"}],ws=(0,kr.rU)(function(s,n){var t=s.css,r=s.token,a=n.compact;return{modal:t(Wi||(Wi=H()([`
        .ant-modal { max-width: calc(100vw - 24px); }
        .ant-modal-body { padding-top: `,`px; }
    `])),a?12:16),steps:t(Ui||(Ui=H()([`
        margin-bottom: `,`px;
    `])),a?20:24),content:t(Gi||(Gi=H()([`
        height: `,`px;
        max-height: calc(100dvh - 290px);
        overflow-x: hidden;
        overflow-y: auto;
        overscroll-behavior: contain;
        scrollbar-gutter: stable;
        padding: 0 2px;
        @supports not (height: 100dvh) {
            max-height: calc(100vh - 290px);
        }
        @media (max-width: 767px) {
            max-height: calc(100dvh - 240px);
            @supports not (height: 100dvh) {
                max-height: calc(100vh - 240px);
            }
        }
    `])),a?330:370),form:t(Qi||(Qi=H()([`
        .ant-form-item { margin-bottom: `,`px; }
        .ant-form-item-extra { margin-top: 4px; min-height: 0; font-size: 12px; line-height: 1.6; color: `,`; }
        .ant-input-number { width: 100%; }
        textarea.ant-input { resize: vertical; }
    `])),a?16:20,r.colorTextTertiary),intro:t(Ki||(Ki=H()([`
        margin: 0 0 `,`px;
        color: `,`;
        font-size: 12px;
        line-height: 1.6;
    `])),a?14:18,r.colorTextSecondary),typeSelect:t(Xi||(Xi=H()([`
        display: flex;
        flex-direction: column;
        gap: `,`px;
        width: 100%;
        .ant-radio-wrapper {
            align-items: center;
            margin: 0;
            padding: `,`;
            border-radius: `,`px;
            background: `,`;
            transition: background-color `,`;
        }
        .ant-radio-wrapper:hover { background: `,`; }
        .ant-radio-wrapper-checked, .ant-radio-wrapper-checked:hover { background: `,`; }
        .ant-radio-wrapper > span:last-child { min-width: 0; padding-inline-end: 0; }
        .site-type-title { display: block; font-weight: 500; color: `,`; }
        .site-type-description { display: block; margin-top: 3px; color: `,`; font-size: 12px; line-height: 1.5; }
    `])),a?6:8,a?"10px 12px":"12px 14px",r.borderRadiusLG,r.colorFillQuaternary,r.motionDurationMid,r.colorFillTertiary,r.colorPrimaryBg,r.colorText,r.colorTextTertiary),more:t(Ji||(Ji=H()([`
        margin-bottom: `,`px;
    `])),a?12:16),review:t(Yi||(Yi=H()([`
        margin-bottom: `,`px;
        padding: `,`;
        border-radius: `,`px;
        background: `,`;
        .ant-descriptions-item-label { white-space: nowrap; }
        .ant-descriptions-item-content { min-width: 0; overflow-wrap: anywhere; }
        .site-review-domains { max-height: 72px; overflow-y: auto; white-space: pre-line; }
    `])),a?16:20,a?"12px 14px":"14px 16px",r.borderRadiusLG,r.colorFillQuaternary),enabled:t(Zi||(Zi=H()([`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        padding: 4px 0;
        .site-enabled-description { margin-top: 3px; color: `,`; font-size: 12px; line-height: 1.5; }
    `])),r.colorTextTertiary),footer:t(eu||(eu=H()([`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        > div { display: flex; gap: 8px; }
    `])))}}),ua=function(n){return Array.from(new Set((Array.isArray(n)?n:String(n||"").split(/[\s,，;；]+/)).map(function(t){return String(t||"").trim()}).filter(Boolean)))},nu=(0,i.forwardRef)(function(s,n){var t,r=s.onSuccess,a=s.typeItems,l=un.A.useApp(),h=l.message,c=(0,ce.DP)(),f=Sn.Ay.useBreakpoint(),y=!!(!(c==null||(t=c.isCompactTheme)===null||t===void 0)&&t.call(c)),j=ws({compact:y}),F=j.styles,B=(0,i.useRef)({}),o=(0,i.useRef)(null),v=(0,i.useRef)(0),N=(0,i.useRef)(!1),w=d.A.useForm(),u=C()(w,1),S=u[0],x=(0,i.useState)(0),m=C()(x,2),p=m[0],_=m[1],A=(0,i.useState)(0),E=C()(A,2),k=E[0],T=E[1],U=(0,i.useState)({}),P=C()(U,2),M=P[0],I=P[1],$=(0,i.useState)([]),q=C()($,2),z=q[0],he=q[1],R=(0,i.useState)(!1),g=C()(R,2),L=g[0],J=g[1],V=(0,i.useMemo)(function(){return[{label:"\u672A\u5206\u7C7B",value:0}].concat(Je()((a||[]).map(function(ie){return{label:String(ie.name),value:Number(ie.id)}})))},[a]),ee=(0,i.useMemo)(function(){return{type:0,type_id:0,domains:"",remark:"",root:"",target:"",host:"",project_name:"",project_file:"",project_port:8080,arguments:"",enabled:!0,fastcgi_address:"127.0.0.1:9000"}},[]),fe=(0,i.useCallback)(function(){v.current+=1,N.current=!1,S.resetFields(),_(0),T(0),I({}),he([]),J(!1)},[S]),Y=(0,i.useCallback)(function(){var ie;fe(),S.setFieldsValue(ee),(ie=B.current)===null||ie===void 0||ie.show()},[S,ee,fe]),de=(0,i.useCallback)(function(){var ie;v.current+=1,N.current=!1,J(!1),(ie=B.current)===null||ie===void 0||ie.hide()},[]);(0,i.useImperativeHandle)(n,function(){return{open:Y}},[Y]),(0,i.useEffect)(function(){var ie;(ie=o.current)===null||ie===void 0||ie.scrollTo({top:0})},[k]);var xe=(0,i.useCallback)((function(){var ie=ne()(b()().mark(function Be(Z){var cn,Le,Fe,nn,Ue,tn,pe,Se,Ee,we,ue,K,ve;return b()().wrap(function(le){for(;;)switch(le.prev=le.next){case 0:if(!(k!==2||N.current)){le.next=2;break}return le.abrupt("return");case 2:return cn=++v.current,N.current=!0,J(!0),Le=ua(Z.domains),Fe=String(Z.remark||"").trim(),nn={type:p,type_id:Number(Z.type_id||0),status:Z.enabled===!1?1:0,domains:Le,name:p===2?String(Z.project_name||"").trim():Fe||Le[0]},p===0?(nn.root=String(Z.root||"").trim(),nn.static={index:["index.html","index.htm"],browse:!1,try_files:[],hide:[".git",".svn",".env",".env.*","*.sql","*.bak","*.backup"],precompressed:["zstd","gzip"]}):p===1?(Ue=Nr(Z.target),tn=String(Z.host||"").trim(),nn.proxy=D()(D()({upstreams:[{dial:Ue.dial,max_requests:0}],transport:"http",scheme:Ue.scheme},Ue.h2c?{versions:["h2c"]}:{}),{},{policy:"random"}),tn&&(nn.proxy.headers={request_set:{Host:[tn==="$host"||tn==="$http_host"?"{http.request.host}":tn]}})):p===2?(pe=String(Z.project_file||"").trim(),Se=String(Z.arguments||"").trim(),Ee=pe.includes(" ")?'"'.concat(pe.replace(/"/g,'\\"'),'"'):pe,we=Math.max(pe.lastIndexOf("/"),pe.lastIndexOf("\\")),nn.root=we>0?pe.slice(0,we):"",nn.proxy={upstreams:[{dial:"127.0.0.1:".concat(Number(Z.project_port)),max_requests:0}],transport:"http",scheme:"http",policy:"random"},nn.process={command:"".concat(Ee).concat(Se?" ".concat(Se):""),environment:{},max_retries:3,restart_delay:2e9,stop_timeout:1e10}):(ue=String(Z.fastcgi_address||"").trim(),nn.root=String(Z.root||"").trim(),nn.fastcgi={upstreams:[{dial:ue.startsWith("/")?"unix/".concat(ue):ue,max_requests:0}],transport:"fastcgi",policy:"random",split_path:[".php"],index:"index.php",try_files:[],capture_stderr:!0}),le.prev=9,le.next=12,ta({body:nn});case 12:if(K=le.sent,v.current===cn){le.next=15;break}return le.abrupt("return");case 15:(K==null?void 0:K.code)===200?((0,ar.Rp)("site"),(ve=B.current)===null||ve===void 0||ve.hide(),h.success((K==null?void 0:K.message)||"\u521B\u5EFA\u6210\u529F"),r==null||r(K==null?void 0:K.data)):h.error((K==null?void 0:K.message)||"\u521B\u5EFA\u5931\u8D25");case 16:return le.prev=16,v.current===cn&&(N.current=!1,J(!1)),le.finish(16);case 19:case"end":return le.stop()}},Be,null,[[9,,16,19]])}));return function(Be){return ie.apply(this,arguments)}})(),[h,r,p,k]),ye=(0,i.useCallback)(ne()(b()().mark(function ie(){var Be,Z,cn;return b()().wrap(function(Fe){for(;;)switch(Fe.prev=Fe.next){case 0:if(!N.current){Fe.next=2;break}return Fe.abrupt("return");case 2:if(k!==0){Fe.next=5;break}return T(1),Fe.abrupt("return");case 5:return Be=++v.current,Fe.prev=6,Fe.next=9,S.validateFields(k===1?["domains"].concat(Je()(p===1?["target"]:p===2?["project_file","project_name","project_port"]:p===3?["fastcgi_address"]:[])):void 0);case 9:if(Z=Fe.sent,v.current===Be){Fe.next=12;break}return Fe.abrupt("return");case 12:if(k!==1){Fe.next=17;break}I(S.getFieldsValue(!0)),T(2),Fe.next=19;break;case 17:return Fe.next=19,xe(Z);case 19:Fe.next=24;break;case 21:Fe.prev=21,Fe.t0=Fe.catch(6),v.current===Be&&Fe.t0!==null&&Fe.t0!==void 0&&(cn=Fe.t0.errorFields)!==null&&cn!==void 0&&cn[0]&&(T(1),window.requestAnimationFrame(function(){var nn;v.current===Be&&S.scrollToField(Fe.t0===null||Fe.t0===void 0||(nn=Fe.t0.errorFields)===null||nn===void 0||(nn=nn[0])===null||nn===void 0?void 0:nn.name,{block:"nearest",focus:!0})}));case 24:case"end":return Fe.stop()}},ie,null,[[6,21]])})),[S,p,k,xe]),Ce=(0,e.jsx)(d.A.Item,{name:"domains",label:p===2?"\u7ED1\u5B9A\u57DF\u540D\uFF08\u53EF\u9009\uFF09":"\u7ED1\u5B9A\u57DF\u540D",required:p!==2,extra:p===2?"\u4E0D\u7ED1\u5B9A\u57DF\u540D\u4E5F\u53EF\u8FD0\u884C\u7A0B\u5E8F\uFF1B\u591A\u4E2A\u57DF\u540D\u7528\u6362\u884C\u3001\u7A7A\u683C\u6216\u9017\u53F7\u5206\u9694\u3002":"\u652F\u6301\u57DF\u540D\u3001\u6CDB\u57DF\u540D\u6216 IP\uFF1B\u591A\u4E2A\u7528\u6362\u884C\u3001\u7A7A\u683C\u3001\u9017\u53F7\u6216\u5206\u53F7\u5206\u9694\u3002",rules:[{validator:(function(){var ie=ne()(b()().mark(function Z(cn,Le){var Fe;return b()().wrap(function(Ue){for(;;)switch(Ue.prev=Ue.next){case 0:if(Fe=ua(Le),!(p!==2&&!Fe.length)){Ue.next=3;break}throw new Error("\u8BF7\u81F3\u5C11\u8F93\u5165\u4E00\u4E2A\u57DF\u540D");case 3:if(!(Fe.length>1e3)){Ue.next=5;break}throw new Error("\u6BCF\u4E2A\u7F51\u7AD9\u6700\u591A\u7ED1\u5B9A 1000 \u4E2A\u57DF\u540D");case 5:if(!Fe.some(function(tn){return/[\/?#]/.test(tn)||(tn.match(/:/g)||[]).length===1})){Ue.next=7;break}throw new Error("\u53EA\u9700\u586B\u5199\u57DF\u540D\u6216 IP\uFF0C\u4E0D\u8981\u5305\u542B\u534F\u8BAE\u3001\u7AEF\u53E3\u548C\u8DEF\u5F84");case 7:case"end":return Ue.stop()}},Z)}));function Be(Z,cn){return ie.apply(this,arguments)}return Be})()}],children:(0,e.jsx)(ze.A.TextArea,{rows:2,placeholder:`example.com
www.example.com`})});return(0,e.jsxs)(ce.Ht,{ref:B,title:(0,e.jsx)(ce.hE,{children:"\u6DFB\u52A0\u7F51\u7AD9"}),width:620,onOk:ye,onCancel:de,modalProps:{forceRender:!0,rootClassName:F.modal,closable:!0,keyboard:!0,style:{top:f.md?64:24,paddingBottom:f.md?64:24},mask:{closable:!0},afterClose:fe,footer:(0,e.jsxs)("div",{className:F.footer,children:[(0,e.jsx)(ge.A,{color:"default",variant:"filled",onClick:de,children:"\u53D6\u6D88"}),(0,e.jsxs)("div",{children:[k>0&&(0,e.jsx)(ge.A,{color:"default",variant:"filled",disabled:L,onClick:function(){v.current+=1,T(k-1)},children:"\u4E0A\u4E00\u6B65"}),(0,e.jsx)(ge.A,{type:"primary",loading:L,onClick:ye,children:k===2?"\u521B\u5EFA\u7F51\u7AD9":"\u4E0B\u4E00\u6B65"})]})]})},children:[(0,e.jsx)(na.A,{className:F.steps,size:"small",current:k,responsive:!1,titlePlacement:f.sm?"horizontal":"vertical",items:[{title:"\u9009\u62E9\u7C7B\u578B"},{title:"\u7F51\u7AD9\u914D\u7F6E"},{title:"\u786E\u8BA4\u521B\u5EFA"}]}),(0,e.jsx)(d.A,{className:F.form,form:S,layout:"vertical",variant:"filled",disabled:L,initialValues:ee,onFinish:ye,onSubmitCapture:function(Be){Be.preventDefault(),Be.stopPropagation(),ye()},children:(0,e.jsxs)("div",{className:F.content,ref:o,children:[(0,e.jsxs)("div",{hidden:k!==0,children:[(0,e.jsx)("p",{className:F.intro,children:"\u6839\u636E\u4F60\u5DF2\u6709\u7684\u9879\u76EE\u9009\u62E9\uFF0C\u4E0B\u4E00\u6B65\u518D\u586B\u5199\u7F51\u7AD9\u4FE1\u606F\u3002"}),(0,e.jsx)(d.A.Item,{name:"type",children:(0,e.jsx)(dt.Ay.Group,{className:F.typeSelect,"aria-label":"\u7F51\u7AD9\u7C7B\u578B",onChange:function(Be){return _(Number(Be.target.value))},children:Ga.map(function(ie){return(0,e.jsxs)(dt.Ay,{value:ie.value,children:[(0,e.jsx)("span",{className:"site-type-title",children:ie.label}),(0,e.jsx)("span",{className:"site-type-description",children:ie.description})]},ie.value)})})})]}),(0,e.jsxs)("div",{hidden:k!==1,children:[(0,e.jsx)("p",{className:F.intro,children:Ga[p].hint}),p!==2&&Ce,p===1&&(0,e.jsx)(d.A.Item,{name:"target",label:"\u76EE\u6807\u670D\u52A1\u5730\u5740",extra:"\u57DF\u540D\u6536\u5230\u7684\u8BF7\u6C42\u5C06\u8F6C\u53D1\u5230\u8FD9\u91CC\uFF0C\u652F\u6301 HTTP\u3001HTTPS \u6216 UNIX \u5957\u63A5\u5B57\u3002",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u76EE\u6807\u670D\u52A1\u5730\u5740"},{validator:(function(){var ie=ne()(b()().mark(function Z(cn,Le){return b()().wrap(function(nn){for(;;)switch(nn.prev=nn.next){case 0:String(Le||"").trim()&&Nr(Le);case 1:case"end":return nn.stop()}},Z)}));function Be(Z,cn){return ie.apply(this,arguments)}return Be})()}],children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 http://127.0.0.1:8080"})}),p===2&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"project_file",label:"\u6267\u884C\u6587\u4EF6",extra:"\u9009\u62E9\u670D\u52A1\u5668\u4E0A\u53EF\u4EE5\u76F4\u63A5\u8FD0\u884C\u7684\u7A0B\u5E8F\u3002",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u9009\u62E9\u6267\u884C\u6587\u4EF6"}],children:(0,e.jsx)(Br.A,{mode:"file",placeholder:"\u8F93\u5165\u8DEF\u5F84\uFF0C\u6216\u70B9\u51FB\u53F3\u4FA7\u9009\u62E9\u6587\u4EF6",buttonProps:{color:"default",variant:"filled"}})}),(0,e.jsxs)(yt.A,{gutter:[16,0],children:[(0,e.jsx)(ct.A,{xs:24,sm:14,children:(0,e.jsx)(d.A.Item,{name:"project_name",label:"\u9879\u76EE\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u9879\u76EE\u540D\u79F0"}],children:(0,e.jsx)(ze.A,{maxLength:100,placeholder:"\u4F8B\u5982 API \u670D\u52A1"})})}),(0,e.jsx)(ct.A,{xs:24,sm:10,children:(0,e.jsx)(d.A.Item,{name:"project_port",label:"\u7A0B\u5E8F\u76D1\u542C\u7AEF\u53E3",extra:"\u4E0E\u7A0B\u5E8F\u81EA\u8EAB\u7684\u7AEF\u53E3\u914D\u7F6E\u4E00\u81F4\u3002",rules:[{required:!0,message:"\u8BF7\u8F93\u5165\u7A0B\u5E8F\u76D1\u542C\u7AEF\u53E3"}],children:(0,e.jsx)(qe.A,{min:1,max:65535,precision:0,placeholder:"\u4F8B\u5982 8080"})})})]}),Ce]}),(p===0||p===3)&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"root",label:"\u7F51\u7AD9\u76EE\u5F55",extra:"\u9009\u62E9\u5DF2\u6709\u7F51\u9875\u6587\u4EF6\u6240\u5728\u7684\u76EE\u5F55\uFF1B\u7559\u7A7A\u4F1A\u81EA\u52A8\u521B\u5EFA\u4E00\u4E2A\u76EE\u5F55\u3002",children:(0,e.jsx)(Br.A,{mode:"directory",placeholder:"\u81EA\u52A8\u521B\u5EFA\uFF0C\u6216\u9009\u62E9\u5DF2\u6709\u76EE\u5F55",buttonProps:{color:"default",variant:"filled"}})}),p===3&&(0,e.jsx)(d.A.Item,{name:"fastcgi_address",label:"PHP-FPM \u5730\u5740",extra:"\u586B\u5199\u5DF2\u8FD0\u884C\u7684 PHP-FPM \u670D\u52A1\u5730\u5740\uFF0C\u4E5F\u53EF\u4EE5\u4F7F\u7528\u5957\u63A5\u5B57\u8DEF\u5F84\u3002",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165 PHP-FPM \u5730\u5740"}],children:(0,e.jsx)(ze.A,{placeholder:"127.0.0.1:9000 \u6216 /run/php/php-fpm.sock"})})]}),(p===1||p===2)&&(0,e.jsx)(Xn.A,{ghost:!0,className:F.more,activeKey:z,onChange:function(Be){return he(Array.isArray(Be)?Be:[Be])},styles:{header:{padding:"4px 0",alignItems:"center"},body:{padding:"".concat(y?12:16,"px 0 0")}},expandIconPlacement:"end",items:[{key:"advanced",label:p===1?"\u81EA\u5B9A\u4E49\u8BF7\u6C42\u57DF\u540D\uFF08\u53EF\u9009\uFF09":"\u542F\u52A8\u53C2\u6570\uFF08\u53EF\u9009\uFF09",forceRender:!0,children:p===2?(0,e.jsx)(d.A.Item,{name:"arguments",label:"\u542F\u52A8\u53C2\u6570",children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 --config ./config.yaml"})}):(0,e.jsx)(d.A.Item,{name:"host",label:"\u8BF7\u6C42\u57DF\u540D\uFF08Host\uFF09",extra:"\u4E00\u822C\u65E0\u9700\u586B\u5199\uFF0C\u53EA\u6709\u76EE\u6807\u670D\u52A1\u8981\u6C42\u65F6\u624D\u9700\u8981\u8BBE\u7F6E\u3002",children:(0,e.jsx)(ze.A,{placeholder:"\u9ED8\u8BA4\u81EA\u52A8\u5904\u7406"})})}]})]}),(0,e.jsxs)("div",{hidden:k!==2,children:[(0,e.jsx)("p",{className:F.intro,children:"\u786E\u8BA4\u4EE5\u4E0B\u4FE1\u606F\u65E0\u8BEF\uFF0C\u5373\u53EF\u521B\u5EFA\u7F51\u7AD9\u3002"}),(0,e.jsx)("div",{className:F.review,children:(0,e.jsx)(ra.A,{column:1,size:"small",colon:!1,items:[{key:"type",label:"\u7F51\u7AD9\u7C7B\u578B",children:Ga[p].label}].concat(Je()(p===2?[{key:"name",label:"\u9879\u76EE\u540D\u79F0",children:M.project_name}]:[]),[{key:"domains",label:"\u7ED1\u5B9A\u57DF\u540D",children:(0,e.jsx)("div",{className:"site-review-domains",children:ua(M.domains).join(`
`)||"\u4E0D\u7ED1\u5B9A\uFF0C\u72EC\u7ACB\u8FD0\u884C"})},{key:"target",label:p===1?"\u8F6C\u53D1\u5230":p===2?"\u6267\u884C\u6587\u4EF6":"\u7F51\u7AD9\u76EE\u5F55",children:p===1?M.target:p===2?M.project_file:String(M.root||"").trim()||"\u6309\u9996\u4E2A\u57DF\u540D\u81EA\u52A8\u521B\u5EFA"}],Je()(p===2?[{key:"port",label:"\u76D1\u542C\u7AEF\u53E3",children:M.project_port}]:[]),Je()(p===3?[{key:"php",label:"PHP-FPM",children:M.fastcgi_address}]:[]))})}),(0,e.jsxs)(yt.A,{gutter:[16,0],children:[p!==2&&(0,e.jsx)(ct.A,{xs:24,sm:14,children:(0,e.jsx)(d.A.Item,{name:"remark",label:"\u7F51\u7AD9\u540D\u79F0",extra:"\u4EC5\u7528\u4E8E\u5728\u9762\u677F\u4E2D\u8BC6\u522B\u7F51\u7AD9\u3002",children:(0,e.jsx)(ze.A,{maxLength:100,placeholder:ua(M.domains)[0]||"\u7559\u7A7A\u4F7F\u7528\u9996\u4E2A\u57DF\u540D"})})}),(0,e.jsx)(ct.A,{xs:24,sm:p===2?24:10,children:(0,e.jsx)(d.A.Item,{name:"type_id",label:"\u7F51\u7AD9\u5206\u7C7B",children:(0,e.jsx)(De.A,{options:V,showSearch:!1})})})]}),(0,e.jsxs)("div",{className:F.enabled,children:[(0,e.jsxs)("div",{children:[(0,e.jsx)("label",{htmlFor:"create-site-enabled",children:"\u521B\u5EFA\u540E\u542F\u7528"}),(0,e.jsx)("div",{className:"site-enabled-description",children:p===2?"\u7F51\u7AD9\u670D\u52A1\u8FD0\u884C\u65F6\uFF0C\u81EA\u52A8\u542F\u52A8\u5E76\u4FDD\u6301\u7A0B\u5E8F\u8FD0\u884C\u3002":"\u5173\u95ED\u540E\u4EC5\u4FDD\u5B58\u914D\u7F6E\uFF0C\u7A0D\u540E\u53EF\u5728\u7F51\u7AD9\u5217\u8868\u4E2D\u542F\u7528\u3002"})]}),(0,e.jsx)(d.A.Item,{name:"enabled",valuePropName:"checked",noStyle:!0,children:(0,e.jsx)(Dr.A,{id:"create-site-enabled","aria-label":"\u521B\u5EFA\u540E\u542F\u7528"})})]})]})]})})]})});nu.displayName="SiteForm";var As=(0,i.memo)(nu),Es=X(59224),wt=X.n(Es),Cs=X(71640),pt=X.n(Cs),Gn=X(49216),mt=X(45478),la=X(69809),Lt=X(15487),js=X(21642),Ss=X(74353),br=X.n(Ss),_t=X(47577),ru=X(42736),sa=X(96885),tu,au,iu,uu,lu,su,Ds={starting:{label:"\u6B63\u5728\u542F\u52A8",color:"processing"},running:{label:"\u8FD0\u884C\u4E2D",color:"success"},stopping:{label:"\u6B63\u5728\u505C\u6B62",color:"warning"},restarting:{label:"\u7B49\u5F85\u91CD\u8BD5",color:"processing"},stopped:{label:"\u5DF2\u505C\u6B62",color:"default"},failed:{label:"\u8FD0\u884C\u5931\u8D25",color:"error"}},ks=(0,kr.rU)(function(s,n){var t=s.css,r=s.token,a=n.compact;return{toolbar:t(tu||(tu=H()([`
        min-width: 0;
        margin-bottom: `,`px;
        padding: `,`;
        border-radius: `,`px;
        background: `,`;
    `])),a?14:18,a?"10px 12px":"12px 14px",r.borderRadiusLG,r.colorFillQuaternary),row:t(au||(au=H()([`
        display: flex;
        align-items: center;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: 10px 16px;
    `]))),status:t(iu||(iu=H()([`
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 12px;

        .status-label {
            color: `,`;
            font-size: 12px;
            white-space: nowrap;
        }

        .status-value {
            --status-color: `,`;
            display: inline-flex;
            align-items: center;
            gap: 8px;
            min-height: 24px;
            color: `,`;
            font-size: 13px;
            font-weight: 500;
            white-space: nowrap;
        }

        .status-value[data-tone="success"] {
            --status-color: `,`;
        }

        .status-value[data-tone="processing"] {
            --status-color: `,`;
        }

        .status-value[data-tone="warning"] {
            --status-color: `,`;
        }

        .status-value[data-tone="error"] {
            --status-color: `,`;
        }

        .status-dot {
            width: 6px;
            height: 6px;
            flex: none;
            margin: 0 2px;
            border-radius: 50%;
            background: var(--status-color);
            box-shadow: 0 0 0 3px color-mix(in srgb, var(--status-color) 12%, transparent);
        }

        .ant-spin {
            display: inline-flex;
        }

        .ant-spin-dot {
            font-size: 12px;
        }
    `])),r.colorTextTertiary,r.colorTextQuaternary,r.colorText,r.colorSuccess,r.colorPrimary,r.colorWarning,r.colorError),actions:t(uu||(uu=H()([`
        display: flex;
        flex-wrap: wrap;
        gap: `,`px;

        .service-action.ant-btn {
            width: auto;
            min-width: max-content;
            height: `,`px;
            padding: 0 `,`px;
            gap: 5px;
            font-size: `,`px;
        }
    `])),a?5:6,a?28:32,a?8:10,a?11:12),metadata:t(lu||(lu=H()([`
        display: flex;
        align-items: center;
        gap: 8px 20px;
        flex-wrap: wrap;
        margin: `,`px 0 0;
        padding: 0;
        font-size: 12px;
        line-height: 20px;

        > div {
            display: flex;
            align-items: baseline;
            gap: 8px;
            min-width: 0;
        }

        dt {
            color: `,`;
        }

        dd {
            margin: 0;
            color: `,`;
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            font-variant-numeric: tabular-nums;
        }

        .retry-limit {
            color: `,`;
        }
    `])),a?8:10,r.colorTextTertiary,r.colorTextSecondary,r.colorTextTertiary),error:t(su||(su=H()([`
        display: flex;
        align-items: flex-start;
        gap: 7px;
        margin-top: 10px;
        padding-top: 10px;
        border-top: 1px solid `,`;
        color: `,`;
        font-size: 12px;
        line-height: 1.6;

        > .anticon {
            flex: none;
            margin-top: 3px;
        }

        > span:last-child {
            min-width: 0;
            overflow-wrap: anywhere;
            white-space: pre-wrap;
        }
    `])),r.colorSplit,r.colorError)}}),ou=function(n){var t,r=n.siteId,a=n.disabled,l=a===void 0?!1:a,h=(0,ce.DP)(),c=ks({compact:!!(!(h==null||(t=h.isCompactTheme)===null||t===void 0)&&t.call(h))}),f=c.styles,y=un.A.useApp(),j=y.message,F=(0,i.useRef)(null),B=(0,i.useRef)(0),o=(0,i.useRef)(!1),v=(0,i.useRef)(Promise.resolve()),N=(0,i.useState)(null),w=C()(N,2),u=w[0],S=w[1],x=(0,i.useState)(!0),m=C()(x,2),p=m[0],_=m[1],A=(0,i.useState)(null),E=C()(A,2),k=E[0],T=E[1],U=(0,i.useState)(""),P=C()(U,2),M=P[0],I=P[1],$=(0,i.useCallback)(function(V){return B.current===V&&F.current!==null},[]),q=(0,i.useCallback)(function(V){var ee=v.current.then(V,V);return v.current=ee.catch(function(){}),ee},[]),z=(0,i.useCallback)((function(){var V=ne()(b()().mark(function ee(fe){var Y;return b()().wrap(function(xe){for(;;)switch(xe.prev=xe.next){case 0:if($(fe)){xe.next=2;break}return xe.abrupt("return");case 2:return xe.prev=2,xe.next=5,ia({body:{action:"status",id:F.current}});case 5:if(Y=xe.sent,$(fe)){xe.next=8;break}return xe.abrupt("return");case 8:(Y==null?void 0:Y.code)===200?(S(Y==null?void 0:Y.data),I("")):I((Y==null?void 0:Y.message)||"\u83B7\u53D6\u8FDB\u7A0B\u72B6\u6001\u5931\u8D25"),xe.next=14;break;case 11:xe.prev=11,xe.t0=xe.catch(2),$(fe)&&I("\u83B7\u53D6\u8FDB\u7A0B\u72B6\u6001\u5931\u8D25");case 14:return xe.prev=14,$(fe)&&_(!1),xe.finish(14);case 17:case"end":return xe.stop()}},ee,null,[[2,11,14,17]])}));return function(ee){return V.apply(this,arguments)}})(),[$]);(0,i.useEffect)(function(){return B.current+=1,F.current=r,o.current=!1,S(null),T(null),I(""),_(!0),function(){B.current+=1,F.current=null}},[r]),(0,i.useEffect)(function(){if(!k){var V=B.current,ee=!1,fe=0,Y=(function(){var de=ne()(b()().mark(function xe(){return b()().wrap(function(Ce){for(;;)switch(Ce.prev=Ce.next){case 0:if(!(ee||!$(V))){Ce.next=2;break}return Ce.abrupt("return");case 2:if(document.visibilityState!=="visible"){Ce.next=5;break}return Ce.next=5,q(function(){return z(V)});case 5:!ee&&$(V)&&(fe=window.setTimeout(Y,1500));case 6:case"end":return Ce.stop()}},xe)}));return function(){return de.apply(this,arguments)}})();return Y(),function(){ee=!0,window.clearTimeout(fe)}}},[r,k,q,$,z]);var he=(function(){var V=ne()(b()().mark(function ee(fe){var Y,de;return b()().wrap(function(ye){for(;;)switch(ye.prev=ye.next){case 0:if(!(!F.current||o.current||l)){ye.next=2;break}return ye.abrupt("return");case 2:return Y=F.current,de=++B.current,o.current=!0,T(fe),I(""),ye.prev=7,ye.next=10,q(ne()(b()().mark(function Ce(){var ie;return b()().wrap(function(Z){for(;;)switch(Z.prev=Z.next){case 0:if($(de)){Z.next=2;break}return Z.abrupt("return");case 2:return Z.prev=2,Z.next=5,ia({body:{id:Y,action:fe}});case 5:if(ie=Z.sent,$(de)){Z.next=8;break}return Z.abrupt("return");case 8:(ie==null?void 0:ie.code)===200?S(ie==null?void 0:ie.data):j.error((ie==null?void 0:ie.message)||"\u8FDB\u7A0B\u64CD\u4F5C\u5931\u8D25"),Z.next=14;break;case 11:Z.prev=11,Z.t0=Z.catch(2),$(de)&&j.error("\u8FDB\u7A0B\u64CD\u4F5C\u5931\u8D25");case 14:case"end":return Z.stop()}},Ce,null,[[2,11]])})));case 10:return ye.prev=10,$(de)&&(o.current=!1,T(null)),ye.finish(10);case 13:case"end":return ye.stop()}},ee,null,[[7,,10,13]])}));return function(fe){return V.apply(this,arguments)}})(),R=u?Ds[u.state]:void 0,g=u&&["starting","running","restarting"].includes(u.state),L=(u==null?void 0:u.state)==="stopping",J=l||!u||p||!!k;return(0,e.jsxs)("div",{className:f.toolbar,children:[(0,e.jsxs)("div",{className:f.row,children:[(0,e.jsxs)("div",{className:f.status,"aria-live":"polite",children:[(0,e.jsx)("span",{className:"status-label",children:"\u8FD0\u884C\u72B6\u6001"}),(0,e.jsxs)("span",{className:"status-value","data-tone":u!=null&&u.retry_exhausted?"error":R==null?void 0:R.color,children:[p?(0,e.jsx)(mt.A,{size:"small"}):(0,e.jsx)("span",{className:"status-dot","aria-hidden":"true"}),p?"\u83B7\u53D6\u4E2D":u!=null&&u.retry_exhausted?"\u5DF2\u505C\u6B62\u91CD\u8BD5":(R==null?void 0:R.label)||"\u72B6\u6001\u672A\u77E5"]})]}),(0,e.jsxs)("div",{className:f.actions,role:"group","aria-label":"\u8FDB\u7A0B\u64CD\u4F5C",children:[(0,e.jsx)(ge.A,{className:"service-action",color:"default",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",icon:(0,e.jsx)(ce.In,{type:"CaretRightOutlined"}),loading:k==="start",disabled:J||!!g||L,onClick:function(){he("start")},children:"\u542F\u52A8"}),(0,e.jsx)(ge.A,{className:"service-action",color:"danger",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",icon:(0,e.jsx)(ce.In,{type:"PauseOutlined"}),loading:k==="stop",disabled:J||L||!g&&!(u!=null&&u.pid),onClick:function(){he("stop")},children:"\u505C\u6B62"}),(0,e.jsx)(ge.A,{className:"service-action",color:"default",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",icon:(0,e.jsx)(ce.In,{type:"ReloadOutlined"}),loading:k==="restart",disabled:J||L,onClick:function(){he("restart")},children:"\u91CD\u542F"})]})]}),(0,e.jsxs)("dl",{className:f.metadata,children:[(0,e.jsxs)("div",{children:[(0,e.jsx)("dt",{children:"PID"}),(0,e.jsx)("dd",{children:(u==null?void 0:u.pid)||"\u2014"})]}),(0,e.jsxs)("div",{children:[(0,e.jsx)("dt",{children:"\u91CD\u8BD5\u6B21\u6570"}),(0,e.jsx)("dd",{children:u?(0,e.jsxs)(e.Fragment,{children:[u.restart_count,(0,e.jsxs)("span",{className:"retry-limit",children:[" / ",u.max_retries]})]}):"\u2014"})]})]}),(u==null?void 0:u.error)&&(0,e.jsxs)("div",{className:f.error,children:[(0,e.jsx)(ce.In,{type:"InfoCircleOutlined"}),(0,e.jsx)("span",{children:u.error})]}),M&&(0,e.jsxs)("div",{className:f.error,role:"alert",children:[(0,e.jsx)(ce.In,{type:"InfoCircleOutlined"}),(0,e.jsx)("span",{children:M})]})]})};ou.displayName="ProcessControls";var Bs=(0,i.memo)(ou),Tr=X(11834);function Qa(){return Ka.apply(this,arguments)}function Ka(){return Ka=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.Jt)("/cert/list",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ka.apply(this,arguments)}function du(){return Xa.apply(this,arguments)}function Xa(){return Xa=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.Jt)("/cert/info",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Xa.apply(this,arguments)}function Ns(){return Ja.apply(this,arguments)}function Ja(){return Ja=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/cert/log",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ja.apply(this,arguments)}function Ts(){return Ya.apply(this,arguments)}function Ya(){return Ya=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/cert/create",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Ya.apply(this,arguments)}function _s(){return Za.apply(this,arguments)}function Za(){return Za=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/cert/update",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),Za.apply(this,arguments)}function Rs(){return ei.apply(this,arguments)}function ei(){return ei=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/cert/delete",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),ei.apply(this,arguments)}function $r(){return ni.apply(this,arguments)}function ni(){return ni=ne()(b()().mark(function s(){var n,t,r=arguments;return b()().wrap(function(l){for(;;)switch(l.prev=l.next){case 0:return t=r.length>0&&r[0]!==void 0?r[0]:{},l.abrupt("return",(0,dn.bE)("/cert/obtain",t==null?void 0:t.body,t==null?void 0:t.onSuccess,t==null?void 0:t.onFail,t==null?void 0:t.onFinally,void 0,{timeout:Number(t==null||(n=t.body)===null||n===void 0?void 0:n.type)===1?1260*1e3:35*1e3}));case 2:case"end":return l.stop()}},s)})),ni.apply(this,arguments)}function Mr(){return ri.apply(this,arguments)}function ri(){return ri=ne()(b()().mark(function s(){var n,t,r=arguments;return b()().wrap(function(l){for(;;)switch(l.prev=l.next){case 0:return t=r.length>0&&r[0]!==void 0?r[0]:{},l.abrupt("return",(0,dn.bE)("/cert/renew",t==null?void 0:t.body,t==null?void 0:t.onSuccess,t==null?void 0:t.onFail,t==null?void 0:t.onFinally,void 0,{timeout:Number(t==null||(n=t.body)===null||n===void 0?void 0:n.type)===1?1260*1e3:35*1e3}));case 2:case"end":return l.stop()}},s)})),ri.apply(this,arguments)}var Is=["styles","className","children"],Ps=["wide","full","className","children"],zs=["name","title","description","styles"],Os=["root","value","onChange"],$s=["kind","_new","proxy_upstreams","response_status","response_body"],cu,pu,mu,vu,hu,fu,gu,xu,bu,yu,Fu,wu,Au,Eu,Cu,ju,Su,Du,ku,Bu,Nu,Tu,_u,Ru,Iu,Pu,Ms="".concat((_t.A===null||_t.A===void 0?void 0:_t.A.basePath)||"/","ace"),oa=1e9,ti=["GET","POST","PUT","PATCH","DELETE","HEAD","OPTIONS"].map(function(s){return{label:s,value:s}}),zu=[{value:"{http.vars.client_ip}",label:"\u771F\u5B9E\u8BBF\u5BA2 IP",description:"\u5DF2\u6309\u53EF\u4FE1\u4EE3\u7406\u89C4\u5219\u89E3\u6790\uFF0C\u9002\u5408\u5927\u591A\u6570\u7F51\u7AD9"},{value:"{http.request.remote.host}",label:"\u8FDE\u63A5\u6765\u6E90 IP",description:"\u76F4\u63A5\u8FDE\u63A5\u5230\u670D\u52A1\u5668\u7684\u6765\u6E90\u5730\u5740"},{value:"{http.request.host}",label:"\u8BBF\u95EE\u57DF\u540D",description:"\u6309\u8BF7\u6C42\u4F7F\u7528\u7684\u57DF\u540D\u5206\u522B\u7EDF\u8BA1"},{value:"{http.request.uri.path}",label:"\u8BF7\u6C42\u8DEF\u5F84",description:"\u6309\u6BCF\u4E2A\u8BBF\u95EE\u8DEF\u5F84\u5206\u522B\u7EDF\u8BA1"},{value:"{http.request.method}",label:"\u8BF7\u6C42\u65B9\u6CD5",description:"\u6309 GET\u3001POST \u7B49\u8BF7\u6C42\u65B9\u6CD5\u5206\u522B\u7EDF\u8BA1"}],Ou=[{label:"\u8BF7\u6C42\u4FE1\u606F",options:[{label:"\u8BBF\u95EE\u5730\u5740",value:"REQUEST_URI",description:"\u5B8C\u6574\u8BBF\u95EE\u8DEF\u5F84",placeholder:"\u4F8B\u5982 /admin",operator:"contains"},{label:"\u8BF7\u6C42\u6587\u4EF6\u8DEF\u5F84",value:"REQUEST_FILENAME",description:"\u670D\u52A1\u5668\u89E3\u6790\u540E\u7684\u8BF7\u6C42\u6587\u4EF6\u8DEF\u5F84",placeholder:"\u4F8B\u5982 /www/site/index.php",operator:"contains"},{label:"\u8BF7\u6C42\u6587\u4EF6\u540D",value:"REQUEST_BASENAME",description:"\u8BF7\u6C42\u8DEF\u5F84\u672B\u5C3E\u7684\u6587\u4EF6\u540D",placeholder:"\u4F8B\u5982 index.php",operator:"equals"},{label:"\u8BF7\u6C42\u65B9\u6CD5",value:"REQUEST_METHOD",description:"GET\u3001POST \u7B49\u8BF7\u6C42\u65B9\u6CD5",placeholder:"\u4F8B\u5982 POST",operator:"equals"},{label:"\u8BF7\u6C42\u534F\u8BAE",value:"REQUEST_PROTOCOL",description:"\u5F53\u524D\u8BF7\u6C42\u4F7F\u7528\u7684\u534F\u8BAE\u7248\u672C",placeholder:"\u4F8B\u5982 HTTP/2.0",operator:"equals"},{label:"\u8BBF\u5BA2 IP",value:"REMOTE_ADDR",description:"\u5F53\u524D\u8BF7\u6C42\u7684\u6765\u6E90 IP",placeholder:"\u4F8B\u5982 192.168.1.0/24",operator:"ip"}]},{label:"\u53C2\u6570\u4E0E\u6B63\u6587",options:[{label:"\u67E5\u8BE2\u5B57\u7B26\u4E32",value:"QUERY_STRING",description:"\u95EE\u53F7\u540E\u7684\u539F\u59CB\u53C2\u6570",placeholder:"\u4F8B\u5982 debug=true",operator:"contains"},{label:"\u5168\u90E8\u53C2\u6570\u503C",value:"ARGS",description:"GET \u548C POST \u53C2\u6570\u5185\u5BB9",placeholder:"\u4F8B\u5982 union select",operator:"contains"},{label:"\u5168\u90E8\u53C2\u6570\u540D",value:"ARGS_NAMES",description:"GET \u548C POST \u53C2\u6570\u540D\u79F0",placeholder:"\u4F8B\u5982 redirect_url",operator:"equals"},{label:"GET \u53C2\u6570\u503C",value:"ARGS_GET",description:"URL \u67E5\u8BE2\u53C2\u6570\u5185\u5BB9",placeholder:"\u4F8B\u5982 debug=true",operator:"contains"},{label:"GET \u53C2\u6570\u540D",value:"ARGS_GET_NAMES",description:"URL \u67E5\u8BE2\u53C2\u6570\u540D\u79F0",placeholder:"\u4F8B\u5982 redirect_url",operator:"equals"},{label:"POST \u53C2\u6570\u503C",value:"ARGS_POST",description:"\u8868\u5355\u63D0\u4EA4\u53C2\u6570\u5185\u5BB9",placeholder:"\u4F8B\u5982 union select",operator:"contains"},{label:"POST \u53C2\u6570\u540D",value:"ARGS_POST_NAMES",description:"\u8868\u5355\u63D0\u4EA4\u53C2\u6570\u540D\u79F0",placeholder:"\u4F8B\u5982 command",operator:"equals"},{label:"\u8BF7\u6C42\u53C2\u6570\u5927\u5C0F",value:"ARGS_COMBINED_SIZE",description:"\u5168\u90E8\u53C2\u6570\u540D\u79F0\u4E0E\u5185\u5BB9\u7684\u603B\u5B57\u8282\u6570",placeholder:"\u4F8B\u5982 1048576",operator:"greater_than"},{label:"\u8BF7\u6C42\u6B63\u6587",value:"REQUEST_BODY",description:"\u672A\u7ECF\u89E3\u6790\u7684\u8BF7\u6C42\u6B63\u6587",placeholder:"\u4F8B\u5982 malicious payload",operator:"contains"},{label:"\u4E0A\u4F20\u4E34\u65F6\u6587\u4EF6",value:"FILES",description:"\u4E0A\u4F20\u6587\u4EF6\u5728\u670D\u52A1\u5668\u4E2D\u7684\u4E34\u65F6\u8DEF\u5F84",placeholder:"\u4F8B\u5982 /tmp/upload",operator:"contains"},{label:"\u4E0A\u4F20\u6587\u4EF6\u540D",value:"FILES_NAMES",description:"\u4E0A\u4F20\u6587\u4EF6\u7684\u539F\u59CB\u540D\u79F0",placeholder:"\u4F8B\u5982 .php",operator:"ends_with"}]},{label:"\u8BF7\u6C42\u5934\u4E0E Cookie",options:[{label:"\u5168\u90E8\u8BF7\u6C42\u5934",value:"REQUEST_HEADERS",description:"\u5168\u90E8\u8BF7\u6C42\u5934\u7684\u5185\u5BB9",placeholder:"\u4F8B\u5982 suspicious-value",operator:"contains"},{label:"\u8BF7\u6C42\u5934\u540D\u79F0",value:"REQUEST_HEADERS_NAMES",description:"\u5168\u90E8\u8BF7\u6C42\u5934\u7684\u540D\u79F0",placeholder:"\u4F8B\u5982 X-Debug",operator:"equals"},{label:"\u8BF7\u6C42\u57DF\u540D",value:"REQUEST_HEADERS:Host",description:"Host \u8BF7\u6C42\u5934",placeholder:"\u4F8B\u5982 example.com",operator:"equals"},{label:"\u5185\u5BB9\u7C7B\u578B",value:"REQUEST_HEADERS:Content-Type",description:"Content-Type \u8BF7\u6C42\u5934",placeholder:"\u4F8B\u5982 application/json",operator:"contains"},{label:"\u6D4F\u89C8\u5668\u6807\u8BC6",value:"REQUEST_HEADERS:User-Agent",description:"User-Agent \u8BF7\u6C42\u5934",placeholder:"\u4F8B\u5982 sqlmap",operator:"contains"},{label:"\u6765\u6E90\u9875\u9762",value:"REQUEST_HEADERS:Referer",description:"Referer \u8BF7\u6C42\u5934",placeholder:"\u4F8B\u5982 example.com",operator:"contains"},{label:"\u8DE8\u57DF\u6765\u6E90",value:"REQUEST_HEADERS:Origin",description:"Origin \u8BF7\u6C42\u5934",placeholder:"\u4F8B\u5982 https://example.com",operator:"equals"},{label:"\u8BA4\u8BC1\u4FE1\u606F",value:"REQUEST_HEADERS:Authorization",description:"Authorization \u8BF7\u6C42\u5934",placeholder:"\u4F8B\u5982 Bearer",operator:"starts_with"},{label:"\u8F6C\u53D1\u6765\u6E90 IP",value:"REQUEST_HEADERS:X-Forwarded-For",description:"X-Forwarded-For \u8BF7\u6C42\u5934",placeholder:"\u4F8B\u5982 192.168.1.1",operator:"contains"},{label:"Cookie \u5185\u5BB9",value:"REQUEST_COOKIES",description:"\u5168\u90E8 Cookie \u7684\u503C",placeholder:"\u4F8B\u5982 admin=true",operator:"contains"},{label:"Cookie \u540D\u79F0",value:"REQUEST_COOKIES_NAMES",description:"\u5168\u90E8 Cookie \u7684\u540D\u79F0",placeholder:"\u4F8B\u5982 session_id",operator:"equals"}]}],$u=[{label:"\u5305\u542B",value:"contains"},{label:"\u5B8C\u5168\u7B49\u4E8E",value:"equals"},{label:"\u5F00\u5934\u662F",value:"starts_with"},{label:"\u7ED3\u5C3E\u662F",value:"ends_with"},{label:"\u6B63\u5219\u8868\u8FBE\u5F0F",value:"regex"},{label:"IP \u6216\u7F51\u6BB5",value:"ip"},{label:"\u5B58\u5728",value:"exists"},{label:"\u5185\u5BB9\u4E0D\u4E3A\u7A7A",value:"not_empty"},{label:"\u6570\u503C\u5927\u4E8E",value:"greater_than"},{label:"\u6570\u503C\u5C0F\u4E8E",value:"less_than"}],Hs=new Map([["contains","\u4E0D\u5305\u542B"],["equals","\u4E0D\u7B49\u4E8E"],["starts_with","\u4E0D\u4EE5\u6B64\u5F00\u5934"],["ends_with","\u4E0D\u4EE5\u6B64\u7ED3\u5C3E"],["regex","\u4E0D\u5339\u914D\u6B63\u5219"],["ip","\u4E0D\u5C5E\u4E8E\u6B64\u7F51\u6BB5"],["exists","\u4E0D\u5B58\u5728"],["not_empty","\u5185\u5BB9\u4E3A\u7A7A"],["greater_than","\u5C0F\u4E8E\u6216\u7B49\u4E8E"],["less_than","\u5927\u4E8E\u6216\u7B49\u4E8E"]]),Vt=Ou.flatMap(function(s){return s.options}).map(function(s){var n=s.value==="ARGS_COMBINED_SIZE",t=$u.filter(function(r){return r.value==="ip"?s.value==="REMOTE_ADDR":["greater_than","less_than"].includes(r.value)?n:!n||r.value==="equals"});return D()(D()({},s),{},{operators:t,options:t.flatMap(function(r){return[r,{label:Hs.get(r.value),value:"not:".concat(r.value)}]})})}),Mu=new Set(["exists","not_empty"]),Hu=new Set(["contains","equals","starts_with","ends_with","regex"]),ai={loose:{paranoia_level:1,detection_paranoia_level:1,inbound_anomaly_score_threshold:10,outbound_anomaly_score_threshold:8},recommended:{paranoia_level:1,detection_paranoia_level:1,inbound_anomaly_score_threshold:5,outbound_anomaly_score_threshold:4},strict:{paranoia_level:2,detection_paranoia_level:2,inbound_anomaly_score_threshold:5,outbound_anomaly_score_threshold:4}},Lu=function(n){return{target:String((n==null?void 0:n.target)||"REQUEST_URI"),operator:String((n==null?void 0:n.operator)||"contains"),value:String((n==null?void 0:n.value)||""),negated:n!=null&&n.negated?"not":"match",ignore_case:(n==null?void 0:n.ignore_case)==null?!0:!!n.ignore_case}},Ls=function(n){return{match:(n==null?void 0:n.match)==="any"?"any":"all",conditions:(Array.isArray(n==null?void 0:n.conditions)?n.conditions:[]).map(Lu)}},Vu=(0,i.createContext)(void 0),vn=function(n){return Array.from(new Set((Array.isArray(n)?n:String(n||"").split(/[\n,]/)).map(function(t){return String(t||"").trim()}).filter(Boolean)))},Dn=function(n){return Math.max(0,Number(n||0))/oa},kn=function(n){return Math.max(0,Number(n||0))*oa},ii=function(n){return(Array.isArray(n)?n:[]).map(function(t){return t==null?void 0:t.dial}).filter(Boolean)},Vs=function(n){return vn(n).map(function(t){return{dial:t,max_requests:0}})},qu=function(n,t,r){return[].concat(Je()(Object.entries(n||{}).flatMap(function(a){var l=C()(a,2),h=l[0],c=l[1];return(Array.isArray(c)?c:[c]).map(function(f){return{operation:"add",name:h,value:String(f||"")}})})),Je()(Object.entries(t||{}).flatMap(function(a){var l=C()(a,2),h=l[0],c=l[1];return(Array.isArray(c)?c:[c]).map(function(f){return{operation:"set",name:h,value:String(f||"")}})})),Je()((Array.isArray(r)?r:[]).map(function(a){return{operation:"delete",name:String(a||""),value:""}})))},Wu=function(n){return(Array.isArray(n)?n:[]).reduce(function(t,r){var a=["add","set","delete"].includes(r==null?void 0:r.operation)?r.operation:"set",l=String((r==null?void 0:r.name)||"").trim(),h=String((r==null?void 0:r.value)||"").trim();if(!l)return t;if(a==="delete")t.delete.includes(l)||t.delete.push(l);else if(h){var c;((c=t[a])[l]||(c[l]=[])).push(h)}return t},{add:{},set:{},delete:[]})},Uu=function(n){return Object.entries(n||{}).map(function(t){var r=C()(t,2),a=r[0],l=r[1];return"".concat(a,"=").concat(l)}).join(`
`)},Gu=function(n){return String(n||"").split(`
`).reduce(function(t,r){var a=r.indexOf("=");return a>0&&(t[r.slice(0,a).trim()]=r.slice(a+1).trim()),t},{})},qs=(0,kr.rU)(function(s){var n=s.css,t=s.token,r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};return{modal:n(cu||(cu=H()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-container {
            height: `,`px;
            min-height: `,`px;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .ant-modal-body {
            min-height: 0;
            padding: 0 !important;
            display: flex;
            flex: 1 1 auto;
            overflow: hidden;
        }

        .ant-modal-body > div {
            width: 100%;
            min-height: 0;
            flex: 1 1 auto;
            overflow: hidden;
        }

        .ant-modal-header, .ant-modal-footer {
            flex: none;
        }

        .ant-modal-footer {
            margin-top: 0;
            padding-top: `,`px;
        }

    `])),r!=null&&r.compact?680:700,r!=null&&r.compact?680:700,r!=null&&r.compact?10:12),layout:n(pu||(pu=H()([`
        box-sizing: border-box;
        width: 100%;
        height: auto;
        min-width: 0;
        min-height: 0;
        display: grid;
        flex: 1 1 auto;
        grid-template-columns: `,`px minmax(0, 1fr);
        grid-template-rows: minmax(0, 1fr);
        overflow: hidden;

        @media (max-width: 760px) {
            grid-template-columns: minmax(0, 1fr);
            grid-template-rows: auto minmax(0, 1fr);
        }
    `])),r!=null&&r.compact?166:178),navigation:n(mu||(mu=H()([`
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        max-height: calc(100% - `,`px);
        margin: `,`px 0;
        padding: `,`px 8px;
        display: flex;
        align-self: start;
        overflow: hidden;
        border-radius: `,`px;
        background: `,`;

        .navigation-scroll {
            min-width: 0;
            min-height: 0;
            flex: 1;
            overflow-x: hidden;
            overflow-y: auto;
            scrollbar-color: `,` transparent;
            scrollbar-width: thin;
        }

        .navigation-grid {
            min-width: 0;
            display: flex;
            flex-direction: column;
            gap: 2px;
        }

        .navigation-item {
            width: 100%;
            min-width: 0;
            min-height: 40px;
            padding: 0 11px;
            display: flex;
            align-items: center;
            gap: 8px;
            flex: none;
            overflow: hidden;
            border: 0;
            border-radius: `,`px;
            outline: 0;
            background: transparent;
            color: `,`;
            cursor: pointer;
            font: inherit;
            font-size: 13px;
            font-weight: 400;
            text-align: left;
            transition: background-color .16s ease, color .16s ease;
        }

        .navigation-item:hover { background: `,`; }

        .navigation-item:focus-visible {
            box-shadow: 0 0 0 2px `,`;
        }

        .navigation-item[data-active="true"] {
            background: color-mix(in srgb, `," 86%, ",`);
            color: `,`;
            font-weight: 600;
        }

        .navigation-dot {
            width: 7px;
            height: 7px;
            flex: none;
            border-radius: 50%;
            background: color-mix(in srgb, var(--accent), `,` 8%);
            opacity: .62;
        }

        .navigation-item[data-active="true"] .navigation-dot {
            background: var(--accent);
            box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent), transparent 78%);
            opacity: 1;
        }

        .navigation-label {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (max-width: 760px) {
            min-width: 0;
            max-height: none;
            margin: 0;
            padding: 5px 0;
            align-self: stretch;
            overflow: hidden;
            border-right: 0;
            border-radius: 0;
            background: transparent;

            .navigation-scroll {
                width: 100%;
                max-width: 100%;
                flex: 1 1 auto;
                overflow-x: scroll;
                overflow-y: hidden;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
                -webkit-overflow-scrolling: touch;
                touch-action: pan-x;
            }

            .navigation-scroll::-webkit-scrollbar { display: none; }

            .navigation-grid {
                width: max-content;
                min-width: max-content;
                flex-direction: row;
                gap: 6px;
            }

            .navigation-item {
                width: auto;
                min-width: max-content;
                min-height: 34px;
                padding: 0 9px;
                gap: 6px;
                background: `,`;
            }
        }
    `])),r!=null&&r.compact?20:24,r!=null&&r.compact?6:8,r!=null&&r.compact?3:4,t.borderRadius,t.colorFillQuaternary,t.colorBorderSecondary,t.borderRadius,t.colorTextSecondary,t.colorFillTertiary,t.colorPrimaryBg,t.colorPrimaryBg,t.colorBgContainer,t.colorTextHeading,t.colorTextTertiary,t.colorFillQuaternary),content:n(vu||(vu=H()([`
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        height: auto;
        padding: `,`;
        overflow: auto;
        overscroll-behavior: contain;
        scrollbar-color: `,` transparent;
        scrollbar-width: thin;
        -webkit-overflow-scrolling: touch;
        touch-action: pan-y;

        .ant-collapse-ghost {
            margin-top: 2px;
            border: 0 !important;
            background: transparent !important;
        }

        .ant-collapse-ghost > .ant-collapse-item {
            border: 0 !important;
        }

        .ant-collapse-ghost > .ant-collapse-item > .ant-collapse-header {
            min-height: 0;
            padding: `,` !important;
            align-items: center !important;
            border-radius: `,`px !important;
            outline: 0 !important;
            background: transparent !important;
            color: `,` !important;
            font-size: 13px;
            font-weight: 500;
            box-shadow: none !important;
        }

        .ant-collapse-ghost > .ant-collapse-item-active > .ant-collapse-header {
            color: `,` !important;
        }

        .ant-collapse-ghost .ant-collapse-expand-icon {
            color: `,`;
        }

        .ant-collapse-ghost .ant-collapse-item-active .ant-collapse-expand-icon {
            color: `,`;
        }

        .ant-collapse-ghost .ant-collapse-body {
            margin: 1px 0 `,`px;
            padding: `,` !important;
            border-radius: `,`px;
            background: `,`;
        }

        @media (max-width: 760px) {
            min-height: 0;
            height: auto;
            padding: `,`;
        }
    `])),r!=null&&r.compact?"15px 18px 18px":"18px 22px 22px",t.colorBorderSecondary,r!=null&&r.compact?"4px 0":"5px 0",t.borderRadiusSM,t.colorTextSecondary,t.colorText,t.colorTextTertiary,t.colorPrimary,r!=null&&r.compact?3:4,r!=null&&r.compact?"5px 8px":"7px 10px",t.borderRadiusSM,t.colorFillQuaternary,r!=null&&r.compact?"12px 12px 16px":"15px 12px 18px"),contentInner:n(hu||(hu=H()([`
        width: 100%;
        max-width: 800px;
        margin: 0 auto;
    `]))),sectionHeader:n(fu||(fu=H()([`
        margin-bottom: `,`px;
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 12px;

        h3 { margin: 0; color: `,"; font-size: ",`px; font-weight: 600; line-height: 22px; }
        p { margin: 2px 0 0; color: `,`; font-size: 12px; line-height: 18px; }

        @media (max-width: 575px) {
            margin-bottom: `,`px;
        }
    `])),r!=null&&r.compact?10:13,t.colorTextHeading,r!=null&&r.compact?14:15,t.colorTextTertiary,r!=null&&r.compact?10:12),settingAction:n(gu||(gu=H()([`
        &.ant-btn {
            width: auto;
            min-width: 0;
            height: `,`px;
            padding: 0 `,`px;
            gap: 4px;
            font-size: 12px;
            line-height: 1;
            white-space: nowrap;
        }

        &.ant-btn .anticon {
            font-size: 12px;
        }
    `])),r!=null&&r.compact?24:26,r!=null&&r.compact?7:8),ruleAction:n(xu||(xu=H()([`
        &.ant-btn {
            min-width: 0;
            height: 24px;
            padding: 0 6px;
            font-size: 12px;
            line-height: 1;
        }

        &.ant-btn.ant-btn-icon-only {
            width: 24px;
            padding: 0;
        }

        &.ant-btn .anticon {
            font-size: 12px;
        }
    `]))),settingsForm:n(bu||(bu=H()([`
        .ant-form-item {
            margin: 0;
            padding: `,`;
        }

        .ant-form-item-label {
            padding: 0 0 `,`px;
            text-align: left;
            white-space: normal;
        }

        .ant-form-item-label > label {
            height: auto;
            color: `,`;
            font-size: 13px;
            font-weight: 500;
            line-height: 20px;
            white-space: normal;
        }

        .ant-form-item-control {
            min-width: 0;
            width: 100%;
            max-width: none;
        }

        .ant-form-item-control-input {
            width: 100%;
            max-width: none;
        }

        .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            width: 100%;
            max-width: `,`px;
        }

        .ant-form-item-control-input-content > :where(
            textarea.ant-input,
            .ant-select-multiple
        ),
        .setting-wide .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: `,`px;
        }

        .setting-full .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: none;
        }

        :where(input, textarea)::placeholder,
        .ant-select-placeholder {
            font-size: `,`px !important;
        }

        .ant-form-item-extra {
            margin-top: `,`px;
            color: `,`;
            font-size: 12px;
            line-height: 18px;
        }

        .ant-form-item-explain {
            margin-top: 4px;
            font-size: 11px;
        }

        .ant-collapse-body .ant-form-item {
            padding-top: `,`px;
            padding-bottom: `,`px;
        }

        .ant-collapse-body .ant-form-item-label {
            padding-bottom: 3px;
        }

        .ant-collapse-body .ant-form-item-label > label {
            font-size: 12px;
            line-height: 18px;
        }

        .ant-collapse-body .ant-form-item-extra {
            margin-top: 3px;
            font-size: 11px;
            line-height: 16px;
        }

        .ant-collapse-body :where(
            .ant-input,
            .ant-input-number-input,
            .ant-input-prefix,
            .ant-input-suffix,
            .ant-btn,
            .ant-select,
            .ant-select-selection-search-input,
            .ant-select-selection-item,
            .ant-select-placeholder
        ),
        .ant-collapse-body :where(input, textarea)::placeholder {
            font-size: 12px !important;
        }

    `])),r!=null&&r.compact?"4px 0":"6px 0",r!=null&&r.compact?5:6,t.colorText,r!=null&&r.compact?280:300,r!=null&&r.compact?380:420,r!=null&&r.compact?12:13,r!=null&&r.compact?4:5,t.colorTextTertiary,r!=null&&r.compact?2:3,r!=null&&r.compact?2:3),fieldGrid:n(yu||(yu=H()([`
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr));
        align-items: start;
        gap: `,`px 16px;
        min-width: 0;

        &.setting-grid-numeric {
            grid-template-columns: repeat(auto-fit, minmax(min(100%, 148px), 1fr));
        }

        > .ant-form-item {
            min-width: 0;
        }

        .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select
        ) {
            max-width: none;
        }
    `])),r!=null&&r.compact?4:6),formToolbar:n(Fu||(Fu=H()([`
        min-height: 32px;
        margin: -4px 0 `,`px;
        display: flex;
        justify-content: flex-end;
    `])),r!=null&&r.compact?10:14),toggleRow:n(wu||(wu=H()([`
        min-height: 0;
        margin: 0;
        padding: `,`;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 20px;

        .toggle-title { color: `,`; font-size: 13px; font-weight: 500; line-height: 20px; }
        .toggle-description { margin-top: 2px; color: `,`; font-size: 12px; line-height: 18px; }
        .ant-form-item { margin: 0; flex: none; }
    `])),r!=null&&r.compact?"7px 0":"9px 0",t.colorText,t.colorTextTertiary),fieldHeading:n(Au||(Au=H()([`
        margin: `,`;
        color: `,`;
        font-size: 13px;
        font-weight: 500;
    `])),r!=null&&r.compact?"4px 0 10px":"6px 0 12px",t.colorText),httpsTable:n(Eu||(Eu=H()([`
        min-width: 0;
        max-width: 100%;
        margin-bottom: 12px;

        .https-table-scroll {
            width: 100%;
            min-width: 0;
            max-width: 100%;
            overflow-x: auto;
            overscroll-behavior-inline: contain;
            scrollbar-width: thin;
            scrollbar-color: `,` transparent;
            -webkit-overflow-scrolling: touch;
        }

        .https-table-scroll .domain-list-heading, .https-table-scroll .domain-list { min-width: 440px; }
        .https-table-scroll .domain-list-heading, .https-table-scroll .domain-row {
            grid-template-columns: minmax(140px, 1fr) minmax(150px, 1fr) 96px;
        }

        .https-table-scroll .domain-row { align-items: center; }
        .https-domain { min-width: 0; min-height: `,`px; display: flex; align-items: center; gap: 8px; }
        .https-domain .ant-checkbox-wrapper { flex: none; }
        .domain-list-heading .https-domain { min-height: 0; }
        .https-certificate { min-width: 0; }
        .https-certificate-status { color: `,`; font-size: 12px; line-height: 20px; font-variant-numeric: tabular-nums; }
        .https-certificate-status[data-expired="true"] { color: `,`; }
        .https-certificate-status[data-expiring="true"] { color: `,`; }
        .https-operations { min-width: 0; text-align: right; }
        .https-table-scroll .domain-actions { min-height: `,`px; align-items: center; }
        .https-certificate-empty { color: `,`; }
    `])),t.colorBorderSecondary,t.controlHeight,t.colorTextSecondary,t.colorError,t.colorWarning,t.controlHeight,t.colorTextQuaternary),domainEditor:n(Cu||(Cu=H()([`
        width: 100%;
        min-width: 0;

        .domain-composer {
            min-width: 0;
            margin-bottom: `,`px;
        }

        .domain-composer-heading { margin-bottom: 8px; }
        .domain-composer-heading h4 { margin: 0; color: `,`; font-size: 13px; font-weight: 500; line-height: 20px; }

        .domain-composer .ant-form-item { min-width: 0; margin: 0; padding: 0; }
        .domain-composer .ant-form-item-control-input { min-height: 0; }
        .domain-composer .ant-form-item-control-input-content > :where(.ant-input, .ant-input-affix-wrapper) {
            width: 100%;
            max-width: none;
            min-height: `,`px;
            font-size: 13px;
        }

        .domain-composer .ant-form-item-control-input-content > textarea.ant-input {
            box-sizing: border-box;
            width: 100%;
            max-width: none;
            height: `,`px;
            min-height: `,`px;
            padding: 8px 10px;
            line-height: 20px;
            resize: vertical;
        }

        .domain-composer-footer { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px 12px; margin-top: 8px; }
        .domain-batch-status { flex: 1 1 160px; min-width: 0; color: `,`; font-size: 12px; line-height: 18px; overflow-wrap: anywhere; }
        .domain-batch-status strong { color: `,`; font-weight: 500; font-variant-numeric: tabular-nums; }
        .domain-composer-actions { display: flex; align-items: center; justify-content: flex-end; flex-wrap: nowrap; gap: 6px; margin-left: auto; }
        .domain-composer-actions > .ant-btn { flex: none; }

        .domain-toolbar { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 4px 12px; margin-bottom: 8px; }
        .domain-list-title { display: inline-flex; align-items: baseline; flex-wrap: wrap; gap: 8px; color: `,`; font-size: 13px; font-weight: 500; line-height: 20px; }

        .domain-list-container { width: 100%; min-width: 0; box-sizing: border-box; overflow: hidden; border: 1px solid `,"; border-radius: ",`px; }
        .domain-list { width: 100%; min-width: 0; border-top: 1px solid `,`; }
        .domain-list-heading, .domain-row {
            box-sizing: border-box;
            min-width: 0;
            display: grid;
            grid-template-columns: minmax(0, 1fr) 48px;
            align-items: center;
            gap: `,`px;
            padding: `,`;
        }

        .domain-list-heading { background: `,"; color: ",`; font-size: 12px; line-height: 18px; }
        .domain-list-heading > :last-child { text-align: right; }
        .domain-row { min-height: `,`px; }
        .domain-row + .domain-row { border-top: 1px solid `,`; }
        .domain-main { min-width: 0; }
        .domain-value { display: block; min-width: 0; color: `,`; font-size: 13px; line-height: 20px; overflow-wrap: anywhere; }

        .domain-actions { display: flex; align-items: center; justify-content: flex-end; flex-wrap: nowrap; gap: 4px; min-width: 0; }
        .domain-actions > .ant-btn { flex: none; }

        .domain-empty { padding: 20px 12px; text-align: center; }
        .domain-empty > .anticon { margin-bottom: 7px; color: `,`; font-size: 20px; }
        .domain-empty-title { color: `,`; font-size: 13px; }
        .domain-empty p { margin: 4px 0 0; color: `,`; font-size: 12px; line-height: 18px; }

        .domain-errors { color: `,`; font-size: 12px; line-height: 18px; }
        .domain-errors:not(:empty) { margin: 8px 0; }

        @media (max-width: 575px) {
            .domain-list-heading, .domain-row { padding-inline: 8px; column-gap: 8px; }
        }
    `])),r!=null&&r.compact?18:20,t.colorText,r!=null&&r.compact?28:32,r!=null&&r.compact?78:80,r!=null&&r.compact?78:80,t.colorTextTertiary,t.colorTextSecondary,t.colorText,t.colorBorderSecondary,t.borderRadiusSM,t.colorBorderSecondary,r!=null&&r.compact?10:12,r!=null&&r.compact?"7px 10px":"8px 12px",t.colorFillQuaternary,t.colorTextTertiary,r!=null&&r.compact?44:48,t.colorBorderSecondary,t.colorText,t.colorTextQuaternary,t.colorTextSecondary,t.colorTextTertiary,t.colorError),redirectRules:n(ju||(ju=H()([`
        display: grid;
        gap: `,`px;

        .redirect-toolbar {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 10px;
            color: `,`;
            font-size: 12px;
        }

        .redirect-card {
            min-width: 0;
            border: 1px solid `,`;
            border-radius: `,`px;
        }

        .redirect-heading {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: `,`;
            border-radius: `,`px;
            background: `,`;
        }

        .redirect-summary {
            min-width: 0;
            flex: 1;
            padding: 0;
            display: flex;
            align-items: center;
            gap: 8px;
            border: 0;
            background: transparent;
            color: `,`;
            font: inherit;
            text-align: left;
            cursor: pointer;
        }

        .redirect-summary:focus-visible { outline: 2px solid `,`; outline-offset: 3px; }
        .redirect-summary > .anticon { color: `,`; font-size: 11px; }
        .redirect-summary-text { min-width: 0; flex: 1; }
        .redirect-title { display: flex; align-items: center; gap: 7px; font-size: 13px; font-weight: 500; }
        .redirect-title > span:first-child { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .redirect-status { flex: none; color: `,`; font-size: 11px; font-weight: 400; }
        .redirect-destination { display: block; margin-top: 2px; overflow: hidden; color: `,`; font-size: 11px; line-height: 17px; text-overflow: ellipsis; white-space: nowrap; }
        .redirect-controls { flex: none; display: flex; align-items: center; gap: 5px; }
        .redirect-body { padding: `,`; }
        .redirect-body[hidden] { display: none; }
        .redirect-scope { max-width: 300px; }
        .redirect-scope .ant-segmented-item-label { min-height: 28px; padding: 0 10px; font-size: 12px; line-height: 28px; }
        .redirect-example { margin: 8px 0; padding: 9px 11px; border-radius: `,"px; background: ",`; }
        .redirect-example-title { margin-bottom: 5px; color: `,`; font-size: 12px; font-weight: 500; }
        .redirect-example-row { display: grid; grid-template-columns: 38px minmax(0, 1fr); align-items: baseline; gap: 6px; color: `,`; font-size: 11px; line-height: 19px; }
        .redirect-example-row code { color: `,`; font: inherit; overflow-wrap: anywhere; }
        .redirect-example-note { margin-top: 5px; color: `,`; font-size: 11px; line-height: 17px; }
        .redirect-empty { padding: 22px 12px; text-align: center; }
        .redirect-empty .ant-empty { margin: 0 0 12px; }

        @media (max-width: 575px) {
            .redirect-heading { padding: 9px; gap: 5px; }
            .redirect-body { padding: 6px 10px 10px; }
            .redirect-toolbar { flex-wrap: wrap; }
            .redirect-title { font-size: 12px; }
        }
    `])),r!=null&&r.compact?9:11,t.colorTextTertiary,t.colorBorderSecondary,t.borderRadius,r!=null&&r.compact?"8px 9px":"10px 12px",t.borderRadius,t.colorFillQuaternary,t.colorText,t.colorPrimary,t.colorTextTertiary,t.colorTextTertiary,t.colorTextTertiary,r!=null&&r.compact?"6px 12px 10px":"8px 14px 12px",t.borderRadiusSM,t.colorFillQuaternary,t.colorTextSecondary,t.colorTextTertiary,t.colorTextSecondary,t.colorTextTertiary),routeRules:n(Su||(Su=H()([`
        display: grid;
        gap: `,`px;

        .route-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 10px; color: `,`; font-size: 12px; }
        .route-card { min-width: 0; border: 1px solid `,"; border-radius: ",`px; scroll-margin-block: 8px; }
        .route-heading { display: flex; align-items: center; gap: 8px; padding: `,"; border-radius: ","px; background: ",`; }
        .route-summary { min-width: 0; flex: 1; display: flex; align-items: center; gap: 8px; padding: 0; border: 0; background: transparent; color: `,`; font: inherit; text-align: left; cursor: pointer; }
        .route-summary:focus-visible { outline: 2px solid `,`; outline-offset: 3px; }
        .route-summary > .anticon { flex: none; color: `,`; font-size: 11px; }
        .route-summary-text { min-width: 0; flex: 1; }
        .route-title { display: flex; align-items: center; gap: 7px; font-size: 13px; font-weight: 500; }
        .route-title > span:first-child { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .route-kind { flex: none; color: `,`; font-size: 11px; font-weight: 400; }
        .route-destination { display: block; margin-top: 2px; overflow: hidden; color: `,`; font-size: 11px; line-height: 17px; text-overflow: ellipsis; white-space: nowrap; }
        .route-body { padding: `,`; }
        .route-body[hidden] { display: none; }
        .route-empty { padding: 20px 12px; text-align: center; }
        .route-empty .ant-empty { margin: 0 0 12px; }
        .route-hint { margin: 4px 0 8px; color: `,`; font-size: 12px; line-height: 20px; }

        @media (max-width: 575px) {
            .route-heading { padding: 9px; gap: 5px; }
            .route-body { padding: 6px 10px 10px; }
            .route-toolbar { flex-wrap: wrap; }
            .route-title { font-size: 12px; }
        }
    `])),r!=null&&r.compact?9:11,t.colorTextTertiary,t.colorBorderSecondary,t.borderRadius,r!=null&&r.compact?"8px 9px":"10px 12px",t.borderRadius,t.colorFillQuaternary,t.colorText,t.colorPrimary,t.colorTextTertiary,t.colorTextTertiary,t.colorTextTertiary,r!=null&&r.compact?"6px 12px 10px":"8px 14px 12px",t.colorTextTertiary),code:n(Du||(Du=H()([`
        min-width: 0;
        overflow: hidden;
        border: 1px solid `,`;
        border-radius: `,`px;

        .ace_editor {
            border-radius: `,`px !important;
        }
    `])),t.colorBorderSecondary,t.borderRadius,t.borderRadius),wafMode:n(ku||(ku=H()([`
        max-width: min(100%, 300px);

        .ant-segmented-item-label {
            min-height: 28px;
            padding: 0 10px;
            font-size: 12px;
            line-height: 28px;
        }
    `]))),variableGuide:n(Bu||(Bu=H()([`
        width: 100%;
        max-width: 560px;
        margin: 4px 0 2px;
        padding: `,`;
        display: grid;
        gap: 5px;
        border-radius: `,`px;
        background: `,`;

        .variable-guide-title {
            margin-bottom: 1px;
            color: `,`;
            font-size: 12px;
            font-weight: 500;
        }

        .variable-guide-row {
            min-width: 0;
            display: grid;
            grid-template-columns: 146px minmax(0, 1fr);
            gap: 10px;
            color: `,`;
            font-size: 11px;
            line-height: 17px;
        }

        code {
            overflow: hidden;
            color: `,`;
            font: inherit;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (max-width: 575px) {
            max-width: none;
            .variable-guide-row { grid-template-columns: minmax(0, 1fr); gap: 0; }
        }
    `])),r!=null&&r.compact?"8px 9px":"9px 10px",t.borderRadiusSM,t.colorFillQuaternary,t.colorTextSecondary,t.colorTextTertiary,t.colorTextSecondary),wafRules:n(Nu||(Nu=H()([`
        margin-top: 2px;
        container-type: inline-size;

        .waf-rules-header {
            min-height: 32px;
            margin-bottom: 8px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 12px;
        }
        .waf-rules-title {
            color: `,`;
            font-size: 12px;
            font-weight: 500;
        }

        .waf-rule-list {
            display: grid;
            gap: `,`px;
        }

        .waf-rule-entry {
            min-width: 0;
            border: 1px solid `,`;
            border-radius: `,`px;
            scroll-margin-block: 8px;
        }

        .waf-rule-summary {
            box-sizing: border-box;
            min-height: 40px;
            padding: 4px 6px 4px 9px;
            display: grid;
            grid-template-columns: minmax(0, 1fr) auto;
            align-items: center;
            gap: 10px;
            background: `,`;
            border-radius: `,`px;
        }

        .waf-rule-select {
            min-width: 0;
            height: 32px;
            padding: 0;
            display: flex;
            align-items: center;
            gap: 7px;
            overflow: hidden;
            border: 0;
            outline: 0;
            background: transparent;
            color: inherit;
            font: inherit;
            text-align: left;
            cursor: pointer;
        }

        .waf-rule-select:focus-visible {
            border-radius: `,`px;
            box-shadow: 0 0 0 2px `,`;
        }

        .waf-rule-select > .anticon {
            flex: none;
            color: `,`;
            font-size: 11px;
        }

        .waf-rule-summary-text { min-width: 0; flex: 1; }
        .waf-rule-title { display: flex; align-items: center; gap: 7px; line-height: 18px; }

        .waf-rule-name {
            min-width: 0;
            overflow: hidden;
            color: `,`;
            font-size: 12px;
            font-weight: 500;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .waf-rule-meta {
            display: block;
            overflow: hidden;
            color: `,`;
            font-size: 11px;
            line-height: 14px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .waf-rule-actions {
            height: 32px;
            display: flex;
            align-items: center;
            gap: 3px;
            white-space: nowrap;
        }

        .waf-rule-action {
            flex: none;
            color: `,`;
            font-size: 11px;
            line-height: 18px;
            white-space: nowrap;
        }

        .waf-rule-switch {
            flex: none;
            margin: 0 2px;
        }

        .waf-rule-editor {
            padding: `,`;
        }

        .waf-rule-body[hidden] { display: none; }

        .waf-rule-name-row {
            max-width: 430px;
            display: grid;
            grid-template-columns: 64px minmax(0, 320px);
            align-items: start;
            gap: 9px;
            color: `,`;
            font-size: 12px;
            line-height: 22px;
        }

        .waf-rule-name-row > span {
            height: 24px;
            padding-top: 0;
            display: flex;
            align-items: center;
        }
        .waf-rule-name-row .ant-form-item {
            height: 24px;
            min-height: 24px;
            margin-bottom: 0;
            padding: 0;
            position: relative;
        }

        .waf-rule-name-row .ant-form-item-control-input { min-height: 24px; }
        .waf-rule-name-row .ant-input {
            font-size: 12px;
            line-height: 22px;
        }
        .waf-rule-name-row .ant-input::placeholder { font-size: 12px !important; }

        .waf-rule-sentence,
        .waf-situation-sentence {
            min-height: 24px;
            display: flex;
            align-items: center;
            flex-wrap: wrap;
            gap: 6px;
            color: `,`;
            font-size: 12px;
            line-height: 22px;
        }

        .waf-rule-sentence .ant-select,
        .waf-situation-sentence .ant-select {
            font-size: 12px;
            line-height: 22px;
        }

        .waf-rule-sentence .ant-select-content,
        .waf-situation-sentence .ant-select-content {
            font-size: 12px !important;
        }

        .waf-rule-sentence .ant-select { min-width: 104px; }
        .waf-rule-name-row,
        .waf-rule-action-row { margin-bottom: `,`px; }
        .waf-rule-match { margin: 0 0 `,`px; }
        .waf-situation-list { margin: 6px 0 10px; }
        .waf-rule-action-row:has(+ .waf-rule-match) { margin-bottom: `,`px; }
        .waf-rule-action-row:has(+ .waf-situation-list) { margin-bottom: `,`px; }
        .waf-rule-action-row + .waf-situation-list { margin-top: 0; }

        @container (min-width: 521px) {
            .waf-rule-action-row {
                display: grid;
                grid-template-columns: 64px 112px;
                gap: 9px;
            }

            .waf-rule-action-row .ant-select {
                width: 112px;
                min-width: 0;
            }
        }

        .waf-situation {
            min-width: 0;
            padding: 3px 0;
        }

        .waf-situation[data-multiple="true"] {
            padding-left: 10px;
            border-left: 1px solid `,`;
        }

        .waf-situation-heading {
            min-height: 28px;
            margin-bottom: `,`px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
        }

        .waf-situation-sentence strong {
            color: `,`;
            font-size: 12px;
            font-weight: 500;
        }

        .waf-situation-sentence .ant-select { min-width: 100px; }

        .waf-group-join {
            height: 20px;
            padding-left: 12px;
            display: flex;
            align-items: center;
            color: `,`;
            font-size: 11px;
        }

        .waf-condition-list { margin: 1px 0 3px; }

        .waf-condition {
            min-width: 0;
            padding: 3px 0 11px;
        }

        .waf-condition-row {
            min-width: 0;
            display: grid;
            grid-template-columns: 32px minmax(0, .85fr) minmax(0, .85fr) minmax(0, 1.5fr) 54px;
            align-items: start;
            gap: 6px;
        }

        .waf-condition-row .ant-form-item {
            min-width: 0;
            margin-bottom: 0;
            padding: 0;
        }

        .waf-condition-value {
            height: 24px;
            min-height: 24px;
            position: relative;
        }

        .waf-condition-row .ant-form-item-control-input {
            min-height: 24px;
        }

        .waf-condition-row :where(.ant-select, .ant-input) {
            width: 100%;
            max-width: none;
            font-size: 12px;
            line-height: 22px;
        }

        .waf-condition-row .ant-select-content,
        .waf-condition-row .ant-input::placeholder {
            font-size: 12px !important;
        }

        .waf-condition-join {
            height: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: `,`;
            font-size: 12px;
            line-height: 22px;
            white-space: nowrap;
        }

        .waf-condition-actions {
            height: 24px;
            display: flex;
            align-items: center;
            justify-content: flex-end;
            gap: 1px;
        }

        .waf-case-button {
            width: 25px;
            min-width: 25px;
            padding: 0;
            font-size: 12px;
        }

        .waf-case-button.is-active {
            font-weight: 600;
        }

        .waf-rules-empty {
            padding: `,`;
            color: `,`;
            font-size: 12px;
            text-align: center;
        }

        .waf-rules-empty .ant-empty { margin: 0 0 10px; }
        .waf-rules-empty .ant-empty-description { font-size: 12px; }

        .waf-rule-name-row .ant-form-item-additional,
        .waf-condition-value .ant-form-item-additional {
            height: 14px;
            position: absolute;
            top: 24px;
            right: 0;
            left: 0;
        }

        .waf-rule-name-row .ant-form-item-explain,
        .waf-condition-value .ant-form-item-explain {
            height: 14px;
            margin-top: 0;
            overflow: hidden;
        }

        .waf-rule-name-row .ant-form-item-explain-error,
        .waf-condition-value .ant-form-item-explain-error {
            height: 14px;
            margin-top: 0;
            overflow: hidden;
            font-size: 11px;
            line-height: 14px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @container (max-width: 492px) {
            .waf-rule-summary { grid-template-columns: minmax(0, 1fr) auto; gap: 5px; }
            .waf-rule-actions { gap: 1px; }
            .waf-rule-name-row { grid-template-columns: minmax(0, 1fr); gap: 2px; }
            .waf-rule-name-row > span { padding-top: 0; }
            .waf-rule-name-row .ant-form-item { width: 100%; }
            .waf-situation[data-multiple="true"] { padding-left: 8px; }
            .waf-condition-row {
                grid-template-columns: 30px minmax(0, 1fr) minmax(0, 1fr) 68px;
            }
            .waf-condition-value {
                grid-row: 2;
                grid-column: 2 / 5;
            }
            .waf-condition-actions {
                grid-row: 1;
                grid-column: 4;
            }
            .waf-condition-actions .ant-btn {
                width: 32px;
                min-width: 32px;
            }
        }

        @container (max-width: 380px) {
            .waf-rule-editor { padding-inline: 6px; }
            .waf-situation[data-multiple="true"] { padding-left: 6px; }
            .waf-condition-row {
                grid-template-columns: 28px minmax(0, 1fr) 68px;
            }
            .waf-condition-target {
                grid-row: 1;
                grid-column: 2;
            }
            .waf-condition-operator {
                grid-row: 2;
                grid-column: 2 / 4;
            }
            .waf-condition-value {
                grid-row: 3;
                grid-column: 2 / 4;
            }
            .waf-condition-actions {
                grid-row: 1;
                grid-column: 3;
            }
        }
    `])),t.colorTextHeading,r!=null&&r.compact?7:9,t.colorBorderSecondary,t.borderRadius,t.colorFillQuaternary,t.borderRadius,t.borderRadiusSM,t.colorPrimaryBorder,t.colorTextTertiary,t.colorText,t.colorTextTertiary,t.colorTextTertiary,r!=null&&r.compact?"9px 7px 10px":"11px 9px 12px",t.colorTextSecondary,t.colorTextSecondary,r!=null&&r.compact?14:16,r!=null&&r.compact?6:8,r!=null&&r.compact?18:20,r!=null&&r.compact?6:8,t.colorBorderSecondary,r!=null&&r.compact?6:8,t.colorText,t.colorTextSecondary,t.colorTextSecondary,r!=null&&r.compact?"16px 0":"19px 0",t.colorTextTertiary),rule:n(Tu||(Tu=H()([`
        margin-bottom: 10px;
        padding: `,`px;
        border: 1px solid transparent;
        border-radius: `,`px;
        background: `,`;

        .waf-rule-heading {
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            flex-wrap: wrap;
            gap: 10px;
        }

        .waf-rule-heading > .waf-rule-index,
        .waf-rule-heading > .ant-space {
            flex: none;
            white-space: nowrap;
        }

        .waf-rule-heading > .ant-space { margin-left: auto; }
        .waf-rule-index { color: `,`; font-size: 12px; }
        .ant-form-item:last-child { margin-bottom: 0; }
    `])),r!=null&&r.compact?11:13,t.borderRadius,t.colorFillQuaternary,t.colorTextSecondary),headerPane:n(_u||(_u=H()([`
        --header-control-height: `,`px;
        min-width: 0;

        .header-switcher {
            width: 100%;
            max-width: 320px;
            margin-bottom: `,`px;
        }

        .header-summary {
            margin: 0 0 `,`px;
            color: `,`;
            font-size: 12px;
            line-height: 18px;
        }

        .header-rule-list {
            display: flex;
            flex-direction: column;
            gap: `,`px;
        }

        .header-columns,
        .header-rule {
            display: grid;
            grid-template-columns: 94px minmax(100px, .85fr) minmax(140px, 1.15fr) var(--header-control-height);
            gap: 8px;
        }

        .header-columns {
            margin-bottom: 6px;
            padding: 0 0 0 1px;
            color: `,`;
            font-size: 12px;
            line-height: 18px;
        }

        .header-rule {
            min-width: 0;
            align-items: start;
        }

        .header-rule .ant-form-item { min-width: 0; margin: 0; padding: 0; }
        .header-rule .ant-form-item-control-input { min-height: var(--header-control-height); }
        .header-rule :where(.ant-input, .ant-select) {
            width: 100%;
            height: var(--header-control-height);
            max-width: none !important;
            font-size: 12px;
        }
        .header-rule .ant-input { padding-top: 0; padding-bottom: 0; }
        .header-rule input::placeholder,
        .header-rule .ant-select-placeholder { font-size: 12px !important; }
        .header-rule .ant-form-item-explain { margin-top: 3px; font-size: 11px; line-height: 16px; }

        .header-rule-remove.ant-btn {
            width: var(--header-control-height);
            height: var(--header-control-height);
            min-width: 0;
            padding: 0;
            align-self: start;
            font-size: 12px;
        }
        .header-rule-remove.ant-btn .anticon { font-size: 12px; }

        .header-rule-empty {
            padding: `,`;
            border-radius: `,`px;
            background: `,`;
            color: `,`;
            font-size: 12px;
            text-align: center;
        }

        .header-rule-toolbar {
            margin-top: 8px;
        }

        @media (max-width: 620px) {
            .header-switcher { max-width: none; }
            .header-columns { display: none; }
            .header-rule {
                grid-template-columns: 94px minmax(0, 1fr) var(--header-control-height);
                gap: 6px 8px;
            }
            .header-rule-list { gap: 14px; }
            .header-rule .header-value {
                grid-column: 1 / 3;
                grid-row: 2;
            }
            .header-rule-remove {
                grid-column: 3;
                grid-row: 1;
            }
        }
    `])),r!=null&&r.compact?26:28,r!=null&&r.compact?9:11,r!=null&&r.compact?10:12,t.colorTextTertiary,r!=null&&r.compact?8:10,t.colorTextSecondary,r!=null&&r.compact?"16px 12px":"20px 14px",t.borderRadius,t.colorFillQuaternary,t.colorTextTertiary),empty:n(Ru||(Ru=H()(["height: 360px; display: flex; align-items: center; justify-content: center;"]))),logGrid:n(Iu||(Iu=H()([`
        display: flex;
        flex-direction: column;
        gap: `,`px;
    `])),r!=null&&r.compact?8:10),logItem:n(Pu||(Pu=H()([`
        box-sizing: border-box;
        min-height: `,`px;
        padding: `,`;
        display: flex;
        align-items: center;
        gap: 12px;
        border-radius: `,`px;
        background: `,`;
        transition: background-color .16s ease;

        .log-icon {
            display: flex;
            width: `,`px;
            height: `,`px;
            flex: none;
            align-items: center;
            justify-content: center;
            border-radius: `,`px;
            background: `,`;
            color: `,`;
            font-size: 14px;
        }

        .log-main { min-width: 0; flex: 1; }
        .log-name { color: `,`; font-size: 12px; font-weight: 500; line-height: 18px; }
        .log-description {
            margin-top: 3px;
            color: `,`;
            font-size: 12px;
            line-height: 18px;
        }

        .log-open.ant-btn { flex: none; }

        @media (max-width: 575px) {
            gap: 9px;
            padding: 10px;
            .log-icon { width: 26px; height: 26px; font-size: 13px; }
        }
    `])),r!=null&&r.compact?56:62,r!=null&&r.compact?"10px 12px":"12px 14px",t.borderRadius,t.colorFillQuaternary,r!=null&&r.compact?28:30,r!=null&&r.compact?28:30,t.borderRadiusSM,t.colorFillTertiary,t.colorTextSecondary,t.colorText,t.colorTextTertiary)}}),Rt=function(n){var t=n.title,r=n.description,a=n.extra,l=n.styles;return(0,e.jsxs)("div",{className:l.sectionHeader,children:[(0,e.jsxs)("div",{children:[(0,e.jsx)("h3",{children:t}),r&&(0,e.jsx)("p",{children:r})]}),a]})},qt=function(n){var t=n.styles,r=n.className,a=r===void 0?"":r,l=n.children,h=pt()(n,Is);return(0,e.jsx)(d.A,D()(D()({},h),{},{layout:"vertical",variant:"filled",className:"".concat(t.settingsForm," ").concat(a).trim(),children:l}))},_r=function(n){var t=n.wide,r=n.full,a=n.className,l=a===void 0?"":a,h=n.children,c=pt()(n,Ps);return(0,e.jsx)(d.A.Item,D()(D()({},c),{},{className:"setting-field ".concat(t?"setting-wide":""," ").concat(r?"setting-full":""," ").concat(l).trim(),children:h}))},Kr=function(n){var t=n.name,r=n.title,a=n.description,l=n.styles,h=pt()(n,zs);return(0,e.jsxs)("div",{className:l.toggleRow,children:[(0,e.jsxs)("div",{children:[(0,e.jsx)("div",{className:"toggle-title",children:r}),a&&(0,e.jsx)("div",{className:"toggle-description",children:a})]}),(0,e.jsx)(d.A.Item,{name:t,valuePropName:"checked",noStyle:!0,children:(0,e.jsx)(Dr.A,D()({},h))})]})},Qu=function(n){var t=n.name,r=n.label,a=n.styles;return(0,e.jsx)(d.A.List,{name:t,children:function(h,c){var f=c.add,y=c.remove;return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:"header-summary",children:[t==="request_rules"?"\u4FEE\u6539\u53D1\u5F80\u7F51\u7AD9\u6216\u4E0A\u6E38\u670D\u52A1\u7684\u8BF7\u6C42\u5934\u3002":"\u4FEE\u6539\u8FD4\u56DE\u7ED9\u6D4F\u89C8\u5668\u7684\u54CD\u5E94\u5934\u3002","\u540D\u79F0\u4E0D\u533A\u5206\u5927\u5C0F\u5199\u3002"]}),h.length>0&&(0,e.jsxs)("div",{className:"header-columns","aria-hidden":"true",children:[(0,e.jsx)("span",{children:"\u5904\u7406\u65B9\u5F0F"}),(0,e.jsx)("span",{children:"\u540D\u79F0"}),(0,e.jsx)("span",{children:"\u5185\u5BB9"})]}),(0,e.jsxs)("div",{className:"header-rule-list",children:[h.length===0&&(0,e.jsx)("div",{className:"header-rule-empty",children:"\u6682\u65E0\u6570\u636E"}),h.map(function(j){var F=j.key,B=j.name;return(0,e.jsxs)("div",{className:"header-rule",children:[(0,e.jsx)(d.A.Item,{name:[B,"operation"],rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u64CD\u4F5C"}],children:(0,e.jsx)(De.A,{size:"small","aria-label":"".concat(r,"\u5904\u7406\u65B9\u5F0F"),placeholder:"\u5904\u7406\u65B9\u5F0F",options:[{label:"\u8BBE\u7F6E / \u8986\u76D6",value:"set",title:"\u8BBE\u7F6E\u5185\u5BB9\uFF0C\u5DF2\u6709\u540C\u540D\u5934\u65F6\u66FF\u6362\u539F\u503C"},{label:"\u8FFD\u52A0\u5185\u5BB9",value:"add",title:"\u4FDD\u7559\u539F\u6709\u5185\u5BB9\uFF0C\u518D\u8FFD\u52A0\u4E00\u4E2A\u503C"},{label:"\u79FB\u9664\u6B64\u5934",value:"delete",title:"\u79FB\u9664\u8FD9\u4E2A\u540D\u79F0\u5BF9\u5E94\u7684\u5168\u90E8\u5185\u5BB9"}]})}),(0,e.jsx)(d.A.Item,{name:[B,"name"],rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165".concat(r,"\u540D\u79F0")}],normalize:function(v){return String(v||"").trimStart()},children:(0,e.jsx)(ze.A,{size:"small","aria-label":"".concat(r,"\u540D\u79F0"),placeholder:t==="request_rules"?"\u4F8B\u5982 X-Real-IP":"\u4F8B\u5982 Cache-Control",autoComplete:"off"})}),(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(v,N){var w,u;return(v==null||(w=v[t])===null||w===void 0||(w=w[B])===null||w===void 0?void 0:w.operation)!==(N==null||(u=N[t])===null||u===void 0||(u=u[B])===null||u===void 0?void 0:u.operation)},children:function(v){var N=v.getFieldValue,w=N([t,B,"operation"])==="delete";return(0,e.jsx)(d.A.Item,{className:"header-value",name:[B,"value"],dependencies:[[t,B,"operation"]],rules:[{validator:(function(){var u=ne()(b()().mark(function x(m,p){return b()().wrap(function(A){for(;;)switch(A.prev=A.next){case 0:if(!(!w&&!String(p||"").trim())){A.next=2;break}throw new Error("\u8BF7\u8F93\u5165".concat(r,"\u5185\u5BB9"));case 2:case"end":return A.stop()}},x)}));function S(x,m){return u.apply(this,arguments)}return S})()}],children:(0,e.jsx)(ze.A,{size:"small","aria-label":"".concat(r,"\u5185\u5BB9"),disabled:w,placeholder:w?"\u65E0\u9700\u586B\u5199\u5185\u5BB9":t==="request_rules"?"\u8F93\u5165\u9700\u8981\u4F20\u9012\u7684\u5185\u5BB9":"\u4F8B\u5982 no-cache",autoComplete:"off"})})}}),(0,e.jsx)(ge.A,{color:"danger",variant:"filled",style:{borderColor:"transparent"},className:"header-rule-remove",title:"\u5220\u9664\u89C4\u5219","aria-label":"\u5220\u9664".concat(r,"\u89C4\u5219 ").concat(B+1),icon:(0,e.jsx)(ce.In,{type:"CloseOutlined"}),onClick:function(){return y(B)}})]},F)})]}),(0,e.jsx)("div",{className:"header-rule-toolbar",children:(0,e.jsxs)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:a.settingAction,size:"small",icon:(0,e.jsx)(ce.In,{type:"PlusOutlined"}),onClick:function(){return f({operation:"set",name:"",value:""})},children:["\u6DFB\u52A0",r]})})]})}})},Ws=function(n){var t=n.ruleName,r=n.groupName,a=n.fields,l=n.add,h=n.remove,c=n.styles,f=d.A.useFormInstance(),y=d.A.useWatch(["rules",t,"groups",r,"match"],{form:f,preserve:!0})||"all",j=function(B,o){return["rules",t,"groups",r,"conditions",B,o]};return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)("div",{className:"waf-condition-list",children:a.map(function(F,B){var o=F.key,v=F.name;return(0,e.jsx)("div",{className:"waf-condition",children:(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(w,u){var S,x,m=w==null||(S=w.rules)===null||S===void 0||(S=S[t])===null||S===void 0||(S=S.groups)===null||S===void 0||(S=S[r])===null||S===void 0||(S=S.conditions)===null||S===void 0?void 0:S[v],p=u==null||(x=u.rules)===null||x===void 0||(x=x[t])===null||x===void 0||(x=x.groups)===null||x===void 0||(x=x[r])===null||x===void 0||(x=x.conditions)===null||x===void 0?void 0:x[v];return(m==null?void 0:m.target)!==(p==null?void 0:p.target)||(m==null?void 0:m.operator)!==(p==null?void 0:p.operator)||(m==null?void 0:m.negated)!==(p==null?void 0:p.negated)||(m==null?void 0:m.ignore_case)!==(p==null?void 0:p.ignore_case)},children:function(w){var u=w.getFieldValue,S=String(u(j(v,"target"))||"REQUEST_URI"),x=String(u(j(v,"operator"))||"contains"),m=u(j(v,"negated"))==="not",p=!!u(j(v,"ignore_case")),_=Vt.find(function(k){return k.value===S})||Vt[0],A=S==="ARGS_COMBINED_SIZE",E=Mu.has(x);return(0,e.jsxs)("div",{className:"waf-condition-row",children:[(0,e.jsx)("span",{className:"waf-condition-join",children:B===0?"\u5982\u679C":y==="any"?"\u6216\u8005":"\u5E76\u4E14"}),(0,e.jsx)(d.A.Item,{className:"waf-condition-target",name:[v,"target"],rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u68C0\u67E5\u4F4D\u7F6E"}],children:(0,e.jsx)(De.A,{size:"small",showSearch:!0,optionFilterProp:"label",title:_.description,placeholder:"\u68C0\u67E5\u4F4D\u7F6E",options:Ou,onChange:function(T){var U=Vt.find(function(M){return M.value===T})||Vt[0],P=f.getFieldValue(j(v,"operator"));(!P||!U.operators.some(function(M){return M.value===P}))&&(f.setFieldValue(j(v,"operator"),U.operator||"contains"),f.setFieldValue(j(v,"negated"),"match"))}})}),(0,e.jsx)(d.A.Item,{name:[v,"operator"],hidden:!0,rules:[{required:!0}],children:(0,e.jsx)(ze.A,{})}),(0,e.jsx)(d.A.Item,{name:[v,"negated"],hidden:!0,children:(0,e.jsx)(ze.A,{})}),(0,e.jsx)(De.A,{className:"waf-condition-operator",size:"small",value:m?"not:".concat(x):x,placeholder:"\u5339\u914D\u65B9\u5F0F",options:_.options,onChange:function(T){var U=String(T).startsWith("not:");f.setFieldValue(j(v,"operator"),U?String(T).slice(4):T),f.setFieldValue(j(v,"negated"),U?"not":"match")}}),(0,e.jsx)(d.A.Item,{className:"waf-condition-value",name:[v,"value"],rules:E?[]:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u5339\u914D\u5185\u5BB9"}].concat(Je()(A?[{pattern:/^\d+$/,message:"\u8BF7\u8F93\u5165\u975E\u8D1F\u6574\u6570"}]:[])),children:(0,e.jsx)(ze.A,{size:"small",disabled:E,inputMode:A?"numeric":void 0,maxLength:2048,placeholder:E?"\u65E0\u9700\u586B\u5199\u5185\u5BB9":A?"\u8BF7\u8F93\u5165\u6BD4\u8F83\u6574\u6570":_.placeholder})}),(0,e.jsx)(d.A.Item,{name:[v,"ignore_case"],valuePropName:"checked",hidden:!0,children:(0,e.jsx)(Dr.A,{})}),(0,e.jsxs)("div",{className:"waf-condition-actions",children:[Hu.has(x)&&!A&&(0,e.jsx)(Kn.A,{title:p?"\u5DF2\u5FFD\u7565\u82F1\u6587\u5927\u5C0F\u5199":"\u5FFD\u7565\u82F1\u6587\u5927\u5C0F\u5199",children:(0,e.jsx)(ge.A,{color:p?"primary":"default",variant:"text",className:"".concat(c.ruleAction," waf-case-button ").concat(p?"is-active":""),size:"small","aria-label":"\u5FFD\u7565\u82F1\u6587\u5927\u5C0F\u5199","aria-pressed":p,onClick:function(){return f.setFieldValue(j(v,"ignore_case"),!p)},children:"Aa"})}),(0,e.jsx)(Kn.A,{title:"\u5220\u9664\u6761\u4EF6",children:(0,e.jsx)(ge.A,{color:"danger",variant:"text",className:c.ruleAction,size:"small","aria-label":"\u5220\u9664\u6761\u4EF6 ".concat(B+1),icon:(0,e.jsx)(ce.In,{type:"CloseOutlined"}),disabled:a.length<=1,onClick:function(){return h(v)}})})]})]})}})},o)})}),(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:c.settingAction,size:"small",icon:(0,e.jsx)(ce.In,{type:"PlusOutlined"}),onClick:function(){return l({target:"REQUEST_URI",operator:"contains",value:"",negated:"match",ignore_case:!0})},children:"\u6DFB\u52A0\u6761\u4EF6"})]})},Us=function(n){var t=n.ruleName,r=d.A.useFormInstance(),a=d.A.useWatch(["rules",t,"match"],{form:r,preserve:!0})||"any";return(0,e.jsx)("div",{className:"waf-group-join",children:a==="all"?"\u5E76\u4E14":"\u6216\u8005"})},Ku=function(n){var t,r=n.path,a=n.label,l=n.options,h=d.A.useFormInstance(),c=d.A.useWatch(r,{form:h,preserve:!0})||((t=l[0])===null||t===void 0?void 0:t.value);return(0,e.jsx)(De.A,{size:"small",value:c,"aria-label":a,options:l,onChange:function(y){return h.setFieldValue(r,y)}})},Gs=function(n){var t=n.field,r=n.styles,a=t.name;return(0,e.jsxs)("div",{className:"waf-rule-editor",children:[(0,e.jsxs)("div",{className:"waf-rule-name-row",children:[(0,e.jsx)("span",{children:"\u89C4\u5219\u540D\u79F0"}),(0,e.jsx)(d.A.Item,{name:[a,"name"],rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u89C4\u5219\u540D\u79F0"}],children:(0,e.jsx)(ze.A,{size:"small",placeholder:"\u4F8B\u5982\uFF1A\u62E6\u622A\u540E\u53F0\u626B\u63CF",maxLength:80})})]}),(0,e.jsxs)("div",{className:"waf-rule-sentence waf-rule-action-row",children:[(0,e.jsx)("span",{children:"\u547D\u4E2D\u540E"}),(0,e.jsx)(d.A.Item,{name:[a,"action"],noStyle:!0,children:(0,e.jsx)(De.A,{size:"small","aria-label":"\u547D\u4E2D\u540E\u7684\u64CD\u4F5C",options:[{label:"\u62E6\u622A\u8BBF\u95EE",value:"block"},{label:"\u4EC5\u8BB0\u65E5\u5FD7",value:"log"}]})})]}),(0,e.jsx)(d.A.List,{name:[a,"groups"],rules:[{validator:(function(){var l=ne()(b()().mark(function c(f,y){return b()().wrap(function(F){for(;;)switch(F.prev=F.next){case 0:if(!(!Array.isArray(y)||y.length===0)){F.next=2;break}throw new Error("\u8BF7\u81F3\u5C11\u6DFB\u52A0\u4E00\u79CD\u60C5\u51B5");case 2:case"end":return F.stop()}},c)}));function h(c,f){return l.apply(this,arguments)}return h})()}],children:function(h,c,f){var y=c.add,j=c.remove,F=f.errors;return(0,e.jsxs)(e.Fragment,{children:[h.length>1&&(0,e.jsxs)("div",{className:"waf-rule-sentence waf-rule-match",children:[(0,e.jsx)("span",{children:"\u7B26\u5408"}),(0,e.jsx)(Ku,{path:["rules",a,"match"],label:"\u60C5\u51B5\u5173\u7CFB",options:[{label:"\u4EFB\u4E00\u60C5\u51B5",value:"any"},{label:"\u6240\u6709\u60C5\u51B5",value:"all"}]}),(0,e.jsx)("span",{children:"\u5373\u53EF\u547D\u4E2D"})]}),(0,e.jsx)("div",{className:"waf-situation-list",children:h.map(function(B,o){var v=B.key,N=B.name;return(0,e.jsxs)("div",{className:"waf-situation-entry",children:[o>0&&(0,e.jsx)(Us,{ruleName:a}),(0,e.jsx)("div",{className:"waf-situation","data-multiple":h.length>1,children:(0,e.jsx)(d.A.List,{name:[N,"conditions"],rules:[{validator:(function(){var w=ne()(b()().mark(function S(x,m){return b()().wrap(function(_){for(;;)switch(_.prev=_.next){case 0:if(!(!Array.isArray(m)||m.length===0)){_.next=2;break}throw new Error("\u6BCF\u79CD\u60C5\u51B5\u81F3\u5C11\u9700\u8981\u4E00\u4E2A\u6761\u4EF6");case 2:case"end":return _.stop()}},S)}));function u(S,x){return w.apply(this,arguments)}return u})()}],children:function(u,S,x){var m=x.errors;return(0,e.jsxs)(e.Fragment,{children:[(h.length>1||u.length>1)&&(0,e.jsxs)("div",{className:"waf-situation-heading",children:[(0,e.jsxs)("div",{className:"waf-situation-sentence",children:[h.length>1&&(0,e.jsxs)("strong",{children:["\u60C5\u51B5 ",o+1]}),u.length>1&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)("span",{children:"\u6EE1\u8DB3"}),(0,e.jsx)(Ku,{path:["rules",a,"groups",N,"match"],label:"\u60C5\u51B5 ".concat(o+1," \u7684\u6761\u4EF6\u5173\u7CFB"),options:[{label:"\u5168\u90E8\u6761\u4EF6",value:"all"},{label:"\u4EFB\u4E00\u6761\u4EF6",value:"any"}]})]})]}),h.length>1&&(0,e.jsx)(ge.A,{color:"danger",variant:"text",className:r.ruleAction,size:"small",onClick:function(){return j(N)},children:"\u5220\u9664"})]}),(0,e.jsx)(Ws,{ruleName:a,groupName:N,fields:u,add:S.add,remove:S.remove,styles:r}),(0,e.jsx)(d.A.ErrorList,{errors:m})]})}})})]},v)})}),(0,e.jsx)(d.A.ErrorList,{errors:F}),(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:r.settingAction,size:"small",icon:(0,e.jsx)(ce.In,{type:"PlusOutlined"}),onClick:function(){return y({match:"all",conditions:[{target:"REQUEST_URI",operator:"contains",value:"",negated:"match",ignore_case:!0}]})},children:"\u6DFB\u52A0\u60C5\u51B5"})]})}})]})},Qs=(0,i.forwardRef)(function(s,n){var t=s.fields,r=s.add,a=s.remove,l=s.errors,h=s.styles,c=d.A.useFormInstance(),f=(0,i.useState)(),y=C()(f,2),j=y[0],F=y[1],B=(0,i.useRef)(!1),o=(0,i.useRef)(),v=(0,i.useRef)(null),N=t.map(function(x){return String(x.key)}).join("|"),w=t.find(function(x){return String(x.key)===String(j)});(0,i.useEffect)(function(){if(t.length===0){j!==void 0&&F(void 0);return}if(B.current){B.current=!1,o.current=t[t.length-1].key,F(t[t.length-1].key);return}(j===void 0||j!==null&&!t.some(function(x){return String(x.key)===String(j)}))&&F(t[0].key)},[j,N]),(0,i.useEffect)(function(){if(w&&o.current===w.key){var x;o.current=void 0,(x=v.current)===null||x===void 0||(x=x.querySelector('[data-rule-key="'.concat(w.key,'"]')))===null||x===void 0||x.scrollIntoView({block:"nearest",inline:"nearest"})}},[j,N]),(0,i.useImperativeHandle)(n,function(){return{reveal:function(m){var p,_=m.find(function(E){var k,T;return((k=E.name)===null||k===void 0?void 0:k[0])==="rules"&&typeof((T=E.name)===null||T===void 0?void 0:T[1])=="number"}),A=t.find(function(E){return E.name===(_==null?void 0:_.name[1])});A&&(F(A.key),(p=v.current)===null||p===void 0||(p=p.querySelector('[data-rule-key="'.concat(A.key,'"]')))===null||p===void 0||p.scrollIntoView({block:"nearest",inline:"nearest"}))}}},[t]);var u=function(m){for(var p,_=new Set((c.getFieldValue("rules")||[]).map(function(T){return String((T==null?void 0:T.name)||"").trim()})),A=m?"".concat(m.name||"\u9632\u62A4\u89C4\u5219"," \u526F\u672C"):"\u9632\u62A4\u89C4\u5219",E=1,k=m?A:"".concat(A," ").concat(E);_.has(k);)k="".concat(A," ").concat(++E);B.current=!0,r(D()(D()({},m||{}),{},{id:"".concat(Date.now(),"-").concat(Math.random().toString(36).slice(2,7)),name:k,enabled:(m==null?void 0:m.enabled)!==!1,match:(m==null?void 0:m.match)||"any",action:(m==null?void 0:m.action)||"block",groups:(m==null||(p=m.groups)===null||p===void 0?void 0:p.map(function(T){return D()(D()({},T),{},{conditions:(T.conditions||[]).map(function(U){return D()({},U)})})}))||[{match:"all",conditions:[{target:"REQUEST_URI",operator:"contains",value:"",negated:"match",ignore_case:!0}]}]}))},S=(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:h.settingAction,size:"small",icon:(0,e.jsx)(ce.In,{type:"PlusOutlined"}),onClick:function(){return u()},children:"\u6DFB\u52A0\u89C4\u5219"});return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:"waf-rules-header",children:[(0,e.jsx)("div",{className:"waf-rules-title",children:"\u81EA\u5B9A\u4E49\u9632\u62A4\u89C4\u5219"}),t.length>0&&S]}),t.length===0?(0,e.jsxs)("div",{className:"waf-rules-empty",children:[(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:"\u6682\u65E0\u6570\u636E"}),S]}):(0,e.jsx)(e.Fragment,{children:(0,e.jsx)("div",{className:"waf-rule-list",ref:v,children:t.map(function(x,m){var p=x.key===(w==null?void 0:w.key);return(0,e.jsxs)("div",{className:"waf-rule-entry","data-rule-key":x.key,children:[(0,e.jsx)("div",{className:"waf-rule-summary","data-active":p,children:(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(A,E){var k,T;return(A==null||(k=A.rules)===null||k===void 0?void 0:k[x.name])!==(E==null||(T=E.rules)===null||T===void 0?void 0:T[x.name])},children:function(A){var E,k=A.getFieldValue,T=k(["rules",x.name])||{},U=Array.isArray(T.groups)?T.groups:[],P=U.reduce(function(I,$){return I+(Array.isArray($==null?void 0:$.conditions)?$.conditions.length:0)},0),M=U.length>1?"\u7B26\u5408".concat(T.match==="all"?"\u6240\u6709":"\u4EFB\u4E00","\u60C5\u51B5"):"\u6EE1\u8DB3".concat(((E=U[0])===null||E===void 0?void 0:E.match)==="any"?"\u4EFB\u4E00":"\u5168\u90E8","\u6761\u4EF6");return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("button",{className:"waf-rule-select",type:"button","aria-expanded":p,"aria-label":"".concat(p?"\u6536\u8D77":"\u5C55\u5F00","\u9632\u62A4\u89C4\u5219 ").concat(m+1),onClick:function(){return F(p?null:x.key)},children:[(0,e.jsx)(ce.In,{type:p?"DownOutlined":"RightOutlined"}),(0,e.jsxs)("span",{className:"waf-rule-summary-text",children:[(0,e.jsxs)("span",{className:"waf-rule-title",children:[(0,e.jsx)("span",{className:"waf-rule-name",children:String(T.name||"\u672A\u547D\u540D\u89C4\u5219 ".concat(m+1))}),(0,e.jsx)("span",{className:"waf-rule-action",children:T.enabled===!1?"\u5DF2\u505C\u7528":T.action==="log"?"\u547D\u4E2D\u540E\u8BB0\u5F55":"\u547D\u4E2D\u540E\u62E6\u622A"})]}),(0,e.jsxs)("span",{className:"waf-rule-meta",children:[M," \xB7 ",U.length>1?"".concat(U.length," \u79CD\u60C5\u51B5 \xB7 "):"",P," \u4E2A\u6761\u4EF6"]})]})]}),(0,e.jsxs)("div",{className:"waf-rule-actions",children:[(0,e.jsx)(d.A.Item,{name:[x.name,"enabled"],valuePropName:"checked",noStyle:!0,children:(0,e.jsx)(Dr.A,{className:"waf-rule-switch",size:"small","aria-label":"\u542F\u7528\u89C4\u5219 ".concat(m+1)})}),(0,e.jsx)(gn.A,{menu:{items:[{key:"copy",label:"\u590D\u5236\u89C4\u5219"},{key:"delete",label:"\u5220\u9664\u89C4\u5219",danger:!0}],onClick:function($){var q=$.key;if(q==="copy"&&u(T),q==="delete"){var z,he=t[m+1]||t[m-1];p&&F((z=he==null?void 0:he.key)!==null&&z!==void 0?z:null),a(x.name)}}},children:(0,e.jsx)(ge.A,{color:"default",variant:"text",className:h.ruleAction,size:"small","aria-label":"\u9632\u62A4\u89C4\u5219 ".concat(m+1," \u64CD\u4F5C"),icon:(0,e.jsx)(ce.In,{type:"MoreOutlined"})})})]})]})}})}),(0,e.jsx)("div",{className:"waf-rule-body",hidden:!p,children:(0,e.jsx)(Gs,{field:x,styles:h})})]},x.key)})})}),(0,e.jsx)(d.A.ErrorList,{errors:l})]})}),Wt=function(n){var t=n.loading,r=n.children;return(0,e.jsx)(mt.A,{spinning:t,size:"medium",children:(0,e.jsx)("div",{style:t?{minHeight:380,visibility:"hidden"}:void 0,children:r})})},Ut=function(n,t){var r=(0,i.useContext)(Vu);return(0,i.useEffect)(function(){var a;r==null||(a=r.update)===null||a===void 0||a.call(r,r.formId,{loading:n,saving:t})},[r,n,t]),(0,i.useEffect)(function(){return function(){var a;return r==null||(a=r.update)===null||a===void 0?void 0:a.call(r,r.formId)}},[r]),r==null?void 0:r.formId},Xu=(0,i.memo)((0,i.forwardRef)(function(s,n){var t=s.formId,r=s.onClose,a=(0,i.useState)({loading:!0,saving:!1}),l=C()(a,2),h=l[0],c=l[1];return(0,i.useImperativeHandle)(n,function(){return{update:function(y){return c(y||{loading:!0,saving:!1})}}},[]),h.loading?null:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(ge.A,{onClick:r,children:"\u5173\u95ED"}),(0,e.jsx)(ge.A,{type:"primary",htmlType:"submit",form:t,loading:h.saving,children:"\u4FDD\u5B58"})]})}));Xu.displayName="SettingsFooter";var qr=(0,i.memo)(function(s){var n=s.site,t=s.request,r=s.update,a=s.title,l=s.description,h=s.styles,c=s.normalize,f=s.serialize,y=s.children,j=s.extra,F=s.onFinishFailed,B=s.hideHeader,o=B===void 0?!1:B,v=un.A.useApp(),N=v.message,w=d.A.useForm(),u=C()(w,1),S=u[0],x=(0,i.useRef)(0),m=(0,i.useState)(!0),p=C()(m,2),_=p[0],A=p[1],E=(0,i.useState)(!1),k=C()(E,2),T=k[0],U=k[1],P=Ut(_,T),M=(0,i.useCallback)(ne()(b()().mark(function $(){var q,z;return b()().wrap(function(R){for(;;)switch(R.prev=R.next){case 0:return q=++x.current,A(!0),R.prev=2,R.next=5,t({body:{action:"get",id:n.id}});case 5:if(z=R.sent,x.current===q){R.next=8;break}return R.abrupt("return");case 8:(z==null?void 0:z.code)===200?(S.resetFields(),S.setFieldsValue(c?c((z==null?void 0:z.data)||{}):(z==null?void 0:z.data)||{})):N.error((z==null?void 0:z.message)||"\u83B7\u53D6".concat(a,"\u5931\u8D25")),R.next=14;break;case 11:R.prev=11,R.t0=R.catch(2),x.current===q&&N.error("\u83B7\u53D6".concat(a,"\u5931\u8D25"));case 14:return R.prev=14,x.current===q&&A(!1),R.finish(14);case 17:case"end":return R.stop()}},$,null,[[2,11,14,17]])})),[S,N,c,t,n.id,a]);(0,i.useEffect)(function(){return M(),function(){x.current+=1}},[M]);var I=(0,i.useCallback)((function(){var $=ne()(b()().mark(function q(z){var he,R,g;return b()().wrap(function(J){for(;;)switch(J.prev=J.next){case 0:return he=++x.current,U(!0),R=!1,J.prev=3,J.next=6,r({body:{action:"set",id:n.id,config:f?f(z,S):z}});case 6:if(g=J.sent,x.current===he){J.next=9;break}return J.abrupt("return");case 9:(g==null?void 0:g.code)===200?(N.success((g==null?void 0:g.message)||"\u4FDD\u5B58\u6210\u529F"),R=!0):N.error((g==null?void 0:g.message)||"\u4FDD\u5B58\u5931\u8D25"),J.next=15;break;case 12:J.prev=12,J.t0=J.catch(3),x.current===he&&N.error("\u4FDD\u5B58\u5931\u8D25");case 15:return J.prev=15,x.current===he&&U(!1),J.finish(15);case 18:R&&x.current===he&&M();case 19:case"end":return J.stop()}},q,null,[[3,12,15,18]])}));return function(q){return $.apply(this,arguments)}})(),[S,M,N,f,n.id,r]);return(0,e.jsxs)(Wt,{loading:_,children:[!o&&(0,e.jsx)(Rt,{title:a,description:l,extra:j==null?void 0:j({reload:M,form:S}),styles:h}),o&&j&&(0,e.jsx)("div",{className:h.formToolbar,children:j({reload:M,form:S})}),(0,e.jsx)(qt,{styles:h,id:P,form:S,onFinish:I,onFinishFailed:F,children:y})]})}),Ks=function(n){var t=n.root,r=n.value,a=r===void 0?"":r,l=n.onChange,h=pt()(n,Os),c=un.A.useApp(),f=c.message,y=(0,i.useMemo)(function(){var B=String(t||"").trim();return(0,Tr.F6)(B)?(0,Tr.j7)(B):""},[t]),j=(0,i.useCallback)(function(B){if(!y)return f.warning("\u8BF7\u5148\u9009\u62E9\u7F51\u7AD9\u76EE\u5F55"),!1;var o=(0,Tr.j7)(B);if(!(0,Tr.ru)(o,y))return f.warning("\u8FD0\u884C\u76EE\u5F55\u5FC5\u987B\u4F4D\u4E8E\u7F51\u7AD9\u76EE\u5F55\u5185"),!1;var v=y.endsWith("/")?y:"".concat(y,"/"),N=(0,Tr.Mt)(o)===(0,Tr.Mt)(y)?"":o.slice(v.length);return l==null||l(N),!0},[y,f,l]),F=(0,i.useCallback)(function(B){var o=String(B||"");if((0,Tr.F6)(o)){j(o);return}l==null||l(o)},[l,j]);return(0,e.jsx)(Br.A,D()(D()({buttonProps:{color:"default",variant:"filled",style:{borderColor:"transparent"}}},h),{},{mode:"directory",value:a,initialPath:y,onChange:F,onSelect:j}))},Xs=function(n){var t=n.site,r=n.styles,a=n.typeItems,l=n.onChange,h=n.hideHeader,c=h===void 0?!1:h,f=un.A.useApp(),y=f.message,j=d.A.useForm(),F=C()(j,1),B=F[0],o=d.A.useWatch("root",B),v=(0,i.useState)(!0),N=C()(v,2),w=N[0],u=N[1],S=(0,i.useState)(!1),x=C()(S,2),m=x[0],p=x[1],_=(0,i.useMemo)(function(){return[{label:"\u5168\u90E8\u5206\u7C7B",value:0}].concat(Je()((a||[]).map(function(k){return{label:String(k.name),value:Number(k.id)}})))},[a]),A=Ut(w,m);(0,i.useEffect)(function(){var k=!0;return u(!0),Ea({body:{id:t.id}}).then(function(T){k&&(u(!1),(T==null?void 0:T.code)===200?B.setFieldsValue(D()({type_id:0},(T==null?void 0:T.data)||t)):y.error((T==null?void 0:T.message)||"\u83B7\u53D6\u7F51\u7AD9\u4FE1\u606F\u5931\u8D25"))}),function(){k=!1}},[B,y,t]);var E=(function(){var k=ne()(b()().mark(function T(U){var P,M;return b()().wrap(function($){for(;;)switch($.prev=$.next){case 0:return p(!0),P={id:t.id,name:String(U.name||"").trim(),type_id:String(U.type_id||0)},Number(t.type)!==1&&(P.root=String(U.root||"").trim(),P.run_path=String(U.run_path||"").trim()),$.next=5,xs({body:P});case 5:M=$.sent,p(!1),(M==null?void 0:M.code)===200?((0,ar.Rp)("site"),y.success((M==null?void 0:M.message)||"\u4FDD\u5B58\u6210\u529F"),l==null||l(D()(D()(D()({},t),P),{},{type_id:Number(P.type_id)}))):y.error((M==null?void 0:M.message)||"\u4FDD\u5B58\u5931\u8D25");case 8:case"end":return $.stop()}},T)}));return function(U){return k.apply(this,arguments)}})();return(0,e.jsxs)(Wt,{loading:w,children:[!c&&(0,e.jsx)(Rt,{title:"\u57FA\u7840\u8BBE\u7F6E",description:"\u4FEE\u6539\u7F51\u7AD9\u540D\u79F0\u4E0E\u7AD9\u70B9\u76EE\u5F55",styles:r}),(0,e.jsxs)(qt,{styles:r,id:A,form:B,onFinish:E,children:[(0,e.jsx)(_r,{name:"name",label:"\u7F51\u7AD9\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u7F51\u7AD9\u540D\u79F0"}],children:(0,e.jsx)(ze.A,{placeholder:"\u8BF7\u8F93\u5165\u7F51\u7AD9\u540D\u79F0",maxLength:100})}),(0,e.jsx)(_r,{name:"type_id",label:"\u7F51\u7AD9\u5206\u7C7B",children:(0,e.jsx)(De.A,{placeholder:"\u5168\u90E8\u5206\u7C7B",options:_})}),Number(t.type)!==1&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(_r,{className:"site-file-picker",name:"root",label:"\u7F51\u7AD9\u76EE\u5F55",extra:"\u7F51\u7AD9\u6587\u4EF6\u4FDD\u5B58\u4F4D\u7F6E",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u7F51\u7AD9\u76EE\u5F55"}],children:(0,e.jsx)(Br.A,{buttonProps:{color:"default",variant:"filled",style:{borderColor:"transparent"}},mode:"directory",placeholder:"\u8BF7\u8F93\u5165\u6216\u9009\u62E9\u7F51\u7AD9\u76EE\u5F55"})}),(0,e.jsx)(_r,{className:"site-file-picker",name:"run_path",label:"\u8FD0\u884C\u76EE\u5F55",dependencies:["root"],extra:"\u9009\u62E9\u7F51\u7AD9\u76EE\u5F55\u5185\u7684\u5B50\u76EE\u5F55\uFF0C\u7559\u7A7A\u5219\u76F4\u63A5\u4F7F\u7528\u7F51\u7AD9\u76EE\u5F55",rules:[{validator:(function(){var k=ne()(b()().mark(function U(P,M){var I,$;return b()().wrap(function(z){for(;;)switch(z.prev=z.next){case 0:if(I=String(M||"").trim(),$=String(B.getFieldValue("root")||"").trim(),!(!I||!(0,Tr.F6)($))){z.next=4;break}return z.abrupt("return");case 4:if(!((0,Tr.F6)(I)||!(0,Tr.ru)((0,Tr.j7)((0,Tr.e_)($,I)),(0,Tr.j7)($)))){z.next=6;break}throw new Error("\u8FD0\u884C\u76EE\u5F55\u5FC5\u987B\u4F4D\u4E8E\u7F51\u7AD9\u76EE\u5F55\u5185");case 6:case"end":return z.stop()}},U)}));function T(U,P){return k.apply(this,arguments)}return T})()}],children:(0,e.jsx)(Ks,{root:o,placeholder:"\u4F8B\u5982 public\uFF0C\u53EF\u7559\u7A7A"})})]})]})]})},Js=function(n){var t=n.site,r=n.styles,a=n.onChange,l=n.hideHeader,h=l===void 0?!1:l,c=un.A.useApp(),f=c.message,y=Number(t.type)===2,j=d.A.useForm(),F=C()(j,1),B=F[0],o=(0,i.useRef)(0),v=(0,i.useState)(!0),N=C()(v,2),w=N[0],u=N[1],S=(0,i.useState)(!1),x=C()(S,2),m=x[0],p=x[1],_=(0,i.useState)(!1),A=C()(_,2),E=A[0],k=A[1],T=(0,i.useState)(""),U=C()(T,2),P=U[0],M=U[1],I=(0,i.useState)(""),$=C()(I,2),q=$[0],z=$[1],he=Ut(w||m,E),R=(0,i.useCallback)(ne()(b()().mark(function fe(){var Y,de,xe;return b()().wrap(function(Ce){for(;;)switch(Ce.prev=Ce.next){case 0:return Y=++o.current,u(!0),p(!1),Ce.prev=3,Ce.next=6,Na({body:{action:"get",id:t.id}});case 6:if(de=Ce.sent,o.current===Y){Ce.next=9;break}return Ce.abrupt("return");case 9:(de==null?void 0:de.code)===200?(xe=Array.isArray(de==null?void 0:de.data)?de==null?void 0:de.data.filter(Boolean):[],B.resetFields(),B.setFieldsValue({domains:xe.map(function(ie){return String(ie.domain||"").trim()}).filter(Boolean)}),M(""),z("")):(p(!0),f.error((de==null?void 0:de.message)||"\u83B7\u53D6\u57DF\u540D\u5931\u8D25")),Ce.next=15;break;case 12:Ce.prev=12,Ce.t0=Ce.catch(3),o.current===Y&&(p(!0),f.error("\u83B7\u53D6\u57DF\u540D\u5931\u8D25"));case 15:return Ce.prev=15,o.current===Y&&u(!1),Ce.finish(15);case 18:case"end":return Ce.stop()}},fe,null,[[3,12,15,18]])})),[B,f,t.id]);(0,i.useEffect)(function(){return R(),function(){o.current+=1}},[R]);var g=function(Y){for(var de=Array.from(new Set(Y.split(/[\s,，;；]+/).filter(Boolean).map(function(ie){return ie.toLowerCase().replace(/\.$/,"").replace(/^\[([^\]]+)\]$/,"$1")}))),xe=0,ye=de;xe<ye.length;xe++){var Ce=ye[xe];if(!Ce)throw new Error("\u8BF7\u8F93\u5165\u6709\u6548\u7684\u57DF\u540D\u6216 IP");if(Ce.length>253)throw new Error("\u57DF\u540D\u4E0D\u80FD\u8D85\u8FC7 253 \u4E2A\u5B57\u7B26");if(/[\/@?#\[\]\\]/.test(Ce)||Ce.includes("*")&&!/^\*\.[^*]+$/.test(Ce))throw new Error("\u300C".concat(Ce,"\u300D\u683C\u5F0F\u4E0D\u6B63\u786E\uFF0C\u8BF7\u53EA\u586B\u5199\u57DF\u540D\u6216 IP\uFF0C\u4E0D\u8981\u5305\u542B\u534F\u8BAE\u548C\u8DEF\u5F84"));if((Ce.match(/:/g)||[]).length===1)throw new Error("\u57DF\u540D\u4E0D\u9700\u8981\u586B\u5199\u7AEF\u53E3")}return de},L=function(Y,de){try{var xe=g(Y),ye=new Set(de.flatMap(function(ie){try{return g(ie)}catch(Be){return[]}})),Ce=xe.filter(function(ie){return!ye.has(ie)});if(de.length+Ce.length>1e3)throw new Error("\u6BCF\u4E2A\u7F51\u7AD9\u6700\u591A\u7ED1\u5B9A 1000 \u4E2A\u57DF\u540D");return{additions:Ce,skipped:Y.split(/[\s,，;；]+/).filter(Boolean).length-Ce.length,error:""}}catch(ie){return{additions:[],skipped:0,error:ie.message}}},J=(function(){var fe=ne()(b()().mark(function Y(de){var xe,ye,Ce,ie,Be,Z,cn=arguments;return b()().wrap(function(Fe){for(;;)switch(Fe.prev=Fe.next){case 0:if(xe=cn.length>1&&cn[1]!==void 0?cn[1]:[],!(!y&&!xe.length&&!P.trim())){Fe.next=3;break}throw new Error("\u8BF7\u81F3\u5C11\u6DFB\u52A0\u4E00\u4E2A\u57DF\u540D");case 3:if(!(xe.length>1e3)){Fe.next=5;break}throw new Error("\u6BCF\u4E2A\u7F51\u7AD9\u6700\u591A\u7ED1\u5B9A 1000 \u4E2A\u57DF\u540D");case 5:ye=new Set,Ce=wt()(xe),Fe.prev=7,Ce.s();case 9:if((ie=Ce.n()).done){Fe.next=19;break}if(Be=ie.value,Z=g(Be||""),Z.length===1){Fe.next=14;break}throw new Error("\u6BCF\u9879\u8BF7\u586B\u5199\u4E00\u4E2A\u57DF\u540D\u6216 IP");case 14:if(!ye.has(Z[0])){Fe.next=16;break}throw new Error("\u57DF\u540D\u300C".concat(Z[0],"\u300D\u91CD\u590D\uFF0C\u8BF7\u5220\u9664\u91CD\u590D\u884C"));case 16:ye.add(Z[0]);case 17:Fe.next=9;break;case 19:Fe.next=24;break;case 21:Fe.prev=21,Fe.t0=Fe.catch(7),Ce.e(Fe.t0);case 24:return Fe.prev=24,Ce.f(),Fe.finish(24);case 27:case"end":return Fe.stop()}},Y,null,[[7,21,24,27]])}));return function(de){return fe.apply(this,arguments)}})(),V=function(){if(!E)try{var Y=B.getFieldValue("domains")||[],de=L(P,Y);if(de.error)throw new Error(de.error);if(!de.additions.length)throw new Error(de.skipped?"\u8FD9\u4E9B\u57DF\u540D\u5DF2\u7ECF\u5728\u5217\u8868\u4E2D":"\u8BF7\u8F93\u5165\u8981\u6DFB\u52A0\u7684\u57DF\u540D");B.setFieldsValue({domains:[].concat(Je()(Y),Je()(de.additions))}),M(""),z(""),f.info({key:he+"-save-reminder",content:"\u57DF\u540D\u5DF2\u6DFB\u52A0\u5230\u5217\u8868\uFF0C\u8BF7\u70B9\u51FB\u201C\u4FDD\u5B58\u201D\u751F\u6548"})}catch(xe){z(xe.message)}},ee=(function(){var fe=ne()(b()().mark(function Y(){var de,xe,ye,Ce,ie;return b()().wrap(function(Z){for(;;)switch(Z.prev=Z.next){case 0:if(!(E||w||m)){Z.next=2;break}return Z.abrupt("return");case 2:if(Z.prev=2,xe=B.getFieldValue("domains")||[],de=Array.from(new Set([].concat(Je()(xe.flatMap(g)),Je()(g(P))))),!(!y&&!de.length)){Z.next=7;break}throw new Error("\u8BF7\u81F3\u5C11\u6DFB\u52A0\u4E00\u4E2A\u57DF\u540D");case 7:if(!(de.length>1e3)){Z.next=9;break}throw new Error("\u6BCF\u4E2A\u7F51\u7AD9\u6700\u591A\u7ED1\u5B9A 1000 \u4E2A\u57DF\u540D");case 9:Z.next=15;break;case 11:return Z.prev=11,Z.t0=Z.catch(2),z(Z.t0.message),Z.abrupt("return");case 15:return z(""),ye=++o.current,k(!0),Ce=!1,Z.prev=19,Z.next=22,Na({body:{action:"set",id:t.id,domains:de}});case 22:if(ie=Z.sent,o.current===ye){Z.next=25;break}return Z.abrupt("return");case 25:(ie==null?void 0:ie.code)===200?(f.success({key:he+"-save-reminder",content:(ie==null?void 0:ie.message)||"\u4FDD\u5B58\u6210\u529F"}),a==null||a(),Ce=!0):f.error((ie==null?void 0:ie.message)||"\u4FDD\u5B58\u5931\u8D25"),Z.next=31;break;case 28:Z.prev=28,Z.t1=Z.catch(19),o.current===ye&&f.error("\u4FDD\u5B58\u5931\u8D25");case 31:return Z.prev=31,o.current===ye&&k(!1),Z.finish(31);case 34:Ce&&o.current===ye&&R();case 35:case"end":return Z.stop()}},Y,null,[[2,11],[19,28,31,34]])}));return function(){return fe.apply(this,arguments)}})();return m?(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:"\u57DF\u540D\u52A0\u8F7D\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5",children:(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:r.settingAction,icon:(0,e.jsx)(ce.In,{type:"ReloadOutlined"}),onClick:function(){R()},children:"\u91CD\u65B0\u52A0\u8F7D"})}):(0,e.jsxs)(Wt,{loading:w,children:[!h&&(0,e.jsx)(Rt,{title:"\u57DF\u540D\u7BA1\u7406",description:y?"\u8BBE\u7F6E\u9879\u76EE\u7684\u8BBF\u95EE\u57DF\u540D\uFF0C\u4E5F\u53EF\u4E0D\u7ED1\u5B9A\u57DF\u540D\u72EC\u7ACB\u8FD0\u884C\u3002":"\u7BA1\u7406\u7F51\u7AD9\u7684\u8BBF\u95EE\u57DF\u540D\uFF0C\u652F\u6301\u901A\u914D\u7B26\u57DF\u540D\u548C IP\u3002",styles:r}),(0,e.jsx)(qt,{styles:r,id:he,form:B,onFinish:ee,disabled:E,scrollToFirstError:{block:"nearest",focus:!0},children:(0,e.jsx)("div",{className:r.domainEditor,children:(0,e.jsx)(d.A.List,{name:"domains",rules:[{validator:J}],children:function(Y,de,xe){var ye=de.remove,Ce=xe.errors,ie=L(P,B.getFieldValue("domains")||[]);return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:"domain-composer",children:[(0,e.jsx)("div",{className:"domain-composer-heading",children:(0,e.jsx)("h4",{children:"\u6DFB\u52A0\u57DF\u540D"})}),(0,e.jsx)(d.A.Item,{validateStatus:q?"error":void 0,help:q||void 0,children:(0,e.jsx)(ze.A.TextArea,{"aria-label":"\u6279\u91CF\u6DFB\u52A0\u57DF\u540D","aria-describedby":he+"-domain-batch-status",placeholder:`example.com
www.example.com
*.example.com`,rows:3,autoComplete:"off",autoCapitalize:"none",spellCheck:!1,value:P,onChange:function(Z){M(Z.target.value),z("")}})}),(0,e.jsxs)("div",{className:"domain-composer-footer",children:[(0,e.jsx)("span",{className:"domain-batch-status",id:he+"-domain-batch-status","aria-live":"polite",children:P.trim()?ie.error?"\u8BF7\u68C0\u67E5\u8F93\u5165\u683C\u5F0F\u548C\u57DF\u540D\u6570\u91CF\u3002":(0,e.jsxs)(e.Fragment,{children:["\u5F85\u6DFB\u52A0 ",(0,e.jsx)("strong",{children:ie.additions.length})," \u4E2A\u57DF\u540D",ie.skipped>0&&"\uFF0C\u5DF2\u5FFD\u7565 ".concat(ie.skipped," \u4E2A\u91CD\u590D\u9879")]}):"\u6362\u884C\u3001\u7A7A\u683C\u3001\u9017\u53F7\u6216\u5206\u53F7\u5206\u9694"}),(0,e.jsx)("div",{className:"domain-composer-actions",children:(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",className:r.settingAction,disabled:E||!P.trim(),onClick:V,children:"\u6DFB\u52A0"})})]})]}),(0,e.jsx)("div",{className:"domain-toolbar",children:(0,e.jsx)("div",{className:"domain-list-title",children:"\u57DF\u540D\u5217\u8868"})}),(0,e.jsx)("div",{className:"domain-errors",role:"alert",children:(0,e.jsx)(d.A.ErrorList,{errors:Ce})}),(0,e.jsxs)("div",{className:"domain-list-container",children:[(0,e.jsxs)("div",{className:"domain-list-heading","aria-hidden":"true",children:[(0,e.jsx)("span",{children:"\u57DF\u540D"}),(0,e.jsx)("span",{children:"\u64CD\u4F5C"})]}),(0,e.jsxs)("div",{className:"domain-list",role:"list","aria-label":"\u57DF\u540D\u5217\u8868",children:[Y.map(function(Be){var Z=Be.key,cn=Be.name,Le=String(B.getFieldValue(["domains",cn])||"");return(0,e.jsxs)("div",{className:"domain-row",role:"listitem",children:[(0,e.jsx)("div",{className:"domain-main",children:(0,e.jsx)("span",{className:"domain-value",children:Le})}),(0,e.jsx)("div",{className:"domain-actions",children:(0,e.jsx)(ge.A,{color:"danger",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",className:r.settingAction,disabled:E||Y.length<=1,title:Y.length<=1?"\u81F3\u5C11\u4FDD\u7559\u4E00\u4E2A\u57DF\u540D":void 0,"aria-label":"\u5220\u9664\u57DF\u540D "+Le,onClick:function(){E||(B.getFieldValue("domains")||[]).length<=1||(ye(cn),f.info({key:he+"-save-reminder",content:"\u57DF\u540D\u5DF2\u4ECE\u5217\u8868\u79FB\u9664\uFF0C\u8BF7\u70B9\u51FB\u201C\u4FDD\u5B58\u201D\u751F\u6548"}))},children:"\u5220\u9664"})})]},Z)}),!Y.length&&(0,e.jsxs)("div",{className:"domain-empty",children:[(0,e.jsx)(ce.In,{type:"GlobalOutlined"}),(0,e.jsx)("div",{className:"domain-empty-title",children:"\u6682\u672A\u6DFB\u52A0\u57DF\u540D"}),(0,e.jsx)("p",{children:y?"\u9879\u76EE\u53EF\u72EC\u7ACB\u8FD0\u884C\uFF0C\u9700\u8981\u57DF\u540D\u8BBF\u95EE\u65F6\u518D\u6DFB\u52A0\u3002":"\u5728\u4E0A\u65B9\u8F93\u5165\u57DF\u540D\uFF0C\u53EF\u4E00\u6B21\u6DFB\u52A0\u591A\u6761\u3002"})]})]})]})]})}})})})]})},ui=function(n){return n.trim().toLowerCase().replace(/\.+$/,"")},Ys=function(n,t){if(n===t)return!0;if(!n.startsWith("*.")||t.startsWith("*.")||t.includes(":")||/^\d+(?:\.\d+){3}$/.test(t))return!1;var r=n.slice(1);if(!t.endsWith(r))return!1;var a=t.slice(0,-r.length);return!!a&&!a.includes(".")},Zs=function(n){var t=n.site,r=n.styles,a=n.onManageCert,l=n.onObtainCert,h=n.onDeployCert,c=n.onRenewCert,f=n.settingsSessionRef,y=n.certVersion,j=y===void 0?0:y,F=n.hideHeader,B=F===void 0?!1:F,o=un.A.useApp(),v=o.message,N=d.A.useForm(),w=C()(N,1),u=w[0],S=d.A.useWatch("bindings",{form:u,preserve:!0})||{},x=(0,i.useState)(function(){return Date.now()}),m=C()(x,2),p=m[0],_=m[1],A=(0,i.useRef)(0),E=(0,i.useRef)(0),k=(0,i.useRef)(0),T=(0,i.useState)(!0),U=C()(T,2),P=U[0],M=U[1],I=(0,i.useState)(!1),$=C()(I,2),q=$[0],z=$[1],he=(0,i.useState)(!0),R=C()(he,2),g=R[0],L=R[1],J=(0,i.useState)(!1),V=C()(J,2),ee=V[0],fe=V[1],Y=Ut(P||q,ee),de=(0,i.useState)([]),xe=C()(de,2),ye=xe[0],Ce=xe[1],ie=(0,i.useState)(new Set),Be=C()(ie,2),Z=Be[0],cn=Be[1],Le=(0,i.useState)([]),Fe=C()(Le,2),nn=Fe[0],Ue=Fe[1],tn=(0,i.useMemo)(function(){return ye.filter(function(ue){return Z.has(String(ue.id))})},[ye,Z]),pe=(0,i.useMemo)(function(){var ue=new Map,K=wt()(nn),ve;try{for(K.s();!(ve=K.n()).done;){var me=ve.value,le=Number(me.id);!Number.isNaN(le)&&!ue.has(le)&&ue.set(le,me)}}catch(Ae){K.e(Ae)}finally{K.f()}return ue},[nn]);(0,i.useEffect)(function(){var ue=window.setInterval(function(){return _(Date.now())},6e4);return function(){return window.clearInterval(ue)}},[]);var Se=(0,i.useCallback)(ne()(b()().mark(function ue(){var K,ve,me,le,Ae,re,te;return b()().wrap(function(se){for(;;)switch(se.prev=se.next){case 0:return K=++A.current,M(!0),z(!1),se.prev=3,se.next=6,Ii({body:{action:"get",id:t.id}});case 6:if(Ae=se.sent,A.current===K){se.next=9;break}return se.abrupt("return");case 9:if((Ae==null?void 0:Ae.code)===200){se.next=13;break}return z(!0),v.error((Ae==null?void 0:Ae.message)||"\u83B7\u53D6 HTTPS \u914D\u7F6E\u5931\u8D25"),se.abrupt("return");case 13:re=Array.isArray(Ae==null||(ve=Ae.data)===null||ve===void 0?void 0:ve.domains)?Ae==null||(me=Ae.data)===null||me===void 0?void 0:me.domains.filter(Boolean):[],te=(Ae==null||(le=Ae.data)===null||le===void 0?void 0:le.config)||{},Ce(re),cn(new Set),u.setFieldsValue(D()(D()({},te),{},{min_version:te.min_version||void 0,max_version:te.max_version||void 0,bindings:Object.fromEntries(re.map(function(Bn){return[String(Bn.id),Number(Bn.cert_id||0)||void 0]}))})),se.next=23;break;case 20:se.prev=20,se.t0=se.catch(3),A.current===K&&(z(!0),v.error("\u83B7\u53D6 HTTPS \u914D\u7F6E\u5931\u8D25"));case 23:return se.prev=23,A.current===K&&M(!1),se.finish(23);case 26:case"end":return se.stop()}},ue,null,[[3,20,23,26]])})),[u,v,t.id]);(0,i.useEffect)(function(){return Se(),function(){A.current+=1}},[Se]),(0,i.useEffect)(function(){var ue=++E.current;return L(!0),Qa({body:{page:1,page_size:1e3,order_by_field:"id",order_by_type:"desc"}}).then(function(K){var ve,me;E.current===ue&&((K==null?void 0:K.code)===200?Ue(Array.isArray(K==null||(ve=K.data)===null||ve===void 0?void 0:ve.list)?K==null||(me=K.data)===null||me===void 0?void 0:me.list.filter(Boolean):[]):v.error((K==null?void 0:K.message)||"\u83B7\u53D6\u8BC1\u4E66\u5217\u8868\u5931\u8D25"))}).catch(function(){E.current===ue&&v.error("\u83B7\u53D6\u8BC1\u4E66\u5217\u8868\u5931\u8D25")}).finally(function(){E.current===ue&&L(!1)}),function(){E.current+=1}},[j,v]);var Ee=function(K,ve){if(!(ee||P||q)&&!(ve==="obtain"&&!l||ve==="deploy"&&!h||ve==="renew"&&!c)){var me=Array.isArray(K)?K:[K],le=Je()(new Set(me.map(function(In){return ui(String(In.domain||""))}).filter(Boolean)));if(le.length){var Ae=A.current,re=++k.current,te=f.current,mn=new Map,se=function(Vn){if(!(A.current!==Ae||k.current!==re||f.current!==te||!(Vn!=null&&Vn.id))){var Zn=Vn.domains;if(typeof Zn=="string")try{Zn=JSON.parse(Zn)}catch(wr){Zn=[]}var Fr=Array.isArray(Zn)?Zn.filter(function(wr){return typeof wr=="string"}).map(ui):[];E.current+=1,L(!1),Ue(function(wr){return[Vn].concat(Je()(wr.filter(function(mr){return String(mr.id)!==String(Vn.id)})))});var ut=!1;me.forEach(function(wr){var mr=ui(String(wr.domain||"")),vr=Fr.includes(mr)?2:Fr.some(function(Ur){return Ys(Ur,mr)})?1:0,Wr=String(wr.id);!vr||vr<(mn.get(Wr)||0)||(mn.set(Wr,vr),Number(u.getFieldValue(["bindings",Wr]))!==Number(Vn.id)&&(ut=!0),u.setFieldValue(["bindings",Wr],Number(Vn.id)))}),_(Date.now()),ut&&v.info({key:Y+"-save-reminder",content:"\u8BC1\u4E66\u5DF2\u5173\u8054\uFF0C\u8BF7\u70B9\u51FB\u201C\u4FDD\u5B58\u201D\u751F\u6548"})}};if(ve==="deploy")h(le[0],se);else if(ve==="renew"){var Bn=Number(u.getFieldValue(["bindings",String(me[0].id)]));Bn&&c(pe.get(Bn)||{id:Bn},se,le[0])}else l(Array.isArray(K)?le:le[0],se)}}},we=(function(){var ue=ne()(b()().mark(function K(){var ve,me,le;return b()().wrap(function(re){for(;;)switch(re.prev=re.next){case 0:if(!(ee||P||q)){re.next=2;break}return re.abrupt("return");case 2:return ve=u.getFieldsValue(!0),me=++A.current,fe(!0),re.prev=5,re.next=8,Ii({body:{action:"set",id:t.id,config:{redirect_http:!!ve.redirect_http,min_version:ve.min_version||"",max_version:ve.max_version||""},bindings:ye.map(function(te){var mn;return{domain_id:te.id,cert_id:Number(((mn=ve.bindings)===null||mn===void 0?void 0:mn[String(te.id)])||0)}})}});case 8:if(le=re.sent,A.current===me){re.next=11;break}return re.abrupt("return");case 11:(le==null?void 0:le.code)===200?(v.success((le==null?void 0:le.message)||"\u4FDD\u5B58\u6210\u529F"),fe(!1),Se()):v.error((le==null?void 0:le.message)||"\u4FDD\u5B58\u5931\u8D25"),re.next=17;break;case 14:re.prev=14,re.t0=re.catch(5),A.current===me&&v.error("\u4FDD\u5B58\u5931\u8D25");case 17:return re.prev=17,A.current===me&&fe(!1),re.finish(17);case 20:case"end":return re.stop()}},K,null,[[5,14,17,20]])}));return function(){return ue.apply(this,arguments)}})();return q?(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:"HTTPS \u914D\u7F6E\u52A0\u8F7D\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5",children:(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:r.settingAction,icon:(0,e.jsx)(ce.In,{type:"ReloadOutlined"}),onClick:function(){Se()},children:"\u91CD\u65B0\u52A0\u8F7D"})}):(0,e.jsxs)(Wt,{loading:P,children:[!B&&(0,e.jsx)(Rt,{title:"HTTPS \u8BC1\u4E66",description:"\u67E5\u770B\u8BC1\u4E66\u90E8\u7F72\u72B6\u6001\uFF0C\u7533\u8BF7\u8BC1\u4E66\u540E\u4FDD\u5B58\u751F\u6548",styles:r,extra:(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",className:r.settingAction,size:"small",onClick:a,children:"\u8BC1\u4E66\u7BA1\u7406"})}),(0,e.jsxs)(qt,{styles:r,id:Y,form:u,onFinish:we,disabled:ee,children:[(0,e.jsxs)("div",{className:"".concat(r.domainEditor," ").concat(r.httpsTable),children:[(0,e.jsxs)("div",{className:"domain-toolbar",children:[(0,e.jsx)("div",{className:"domain-list-title",children:"\u57DF\u540D\u8BC1\u4E66"}),(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent",visibility:tn.length?"visible":"hidden"},htmlType:"button",className:r.settingAction,"aria-hidden":!tn.length,disabled:!tn.length||ee||P||!l,onClick:function(){return Ee(tn,"obtain")},children:"\u6279\u91CF\u7533\u8BF7"})]}),(0,e.jsx)("div",{className:"domain-list-container",children:(0,e.jsxs)("div",{className:"https-table-scroll",children:[(0,e.jsxs)("div",{className:"domain-list-heading",children:[(0,e.jsxs)("span",{className:"https-domain",children:[(0,e.jsx)(la.A,{"aria-label":"\u5168\u9009\u57DF\u540D",disabled:ee||P||!ye.length,checked:ye.length>0&&tn.length===ye.length,indeterminate:tn.length>0&&tn.length<ye.length,onChange:function(K){return cn(new Set(K.target.checked?ye.map(function(ve){return String(ve.id)}):[]))}}),(0,e.jsx)("span",{children:"\u57DF\u540D"})]}),(0,e.jsx)("span",{children:"\u8BC1\u4E66"}),(0,e.jsx)("span",{children:"\u64CD\u4F5C"})]}),(0,e.jsxs)("div",{className:"domain-list",role:"list","aria-label":"\u57DF\u540D\u8BC1\u4E66\u5217\u8868",children:[ye.map(function(ue){var K=String(ue.domain||""),ve=Number(S[String(ue.id)]||0),me=ve!==Number(ue.cert_id||0),le=pe.get(ve),Ae=le!=null&&le.expire_time?br()(String(le.expire_time)):null,re=Ae!=null&&Ae.isValid()?Ae.valueOf()-p:null,te=re!==null&&re<=0,mn=re===null?0:Math.floor(re/864e5),se=re===null?0:Math.floor(re/36e5)%24,Bn=re===null?g?"\u6B63\u5728\u52A0\u8F7D\u6709\u6548\u671F":"\u6709\u6548\u671F\u672A\u77E5":te?"\u5DF2\u8FC7\u671F":mn>0?"\u5269\u4F59 ".concat(mn," \u5929").concat(se>0?" ".concat(se," \u5C0F\u65F6"):""):se>0?"\u5269\u4F59 ".concat(se," \u5C0F\u65F6"):"\u5269\u4F59 ".concat(Math.max(1,Math.ceil(re/6e4))," \u5206\u949F");return(0,e.jsxs)("div",{className:"domain-row",role:"listitem",children:[(0,e.jsxs)("div",{className:"https-domain",children:[(0,e.jsx)(la.A,{"aria-label":"\u9009\u62E9\u57DF\u540D "+K,disabled:ee||P,checked:Z.has(String(ue.id)),onChange:function(Vn){return cn(function(Zn){var Fr=new Set(Zn);return Vn.target.checked?Fr.add(String(ue.id)):Fr.delete(String(ue.id)),Fr})}}),(0,e.jsx)("span",{className:"domain-value",children:K})]}),(0,e.jsx)("div",{className:"https-certificate",children:ve?(0,e.jsx)("span",{className:"https-certificate-status","data-expired":te||void 0,"data-expiring":!te&&re!==null&&re<=864e5||void 0,title:"".concat(me?"\u5F85\u90E8\u7F72":"\u5DF2\u90E8\u7F72").concat(Ae!=null&&Ae.isValid()?" \xB7 ".concat(String(le.expire_time).replace("T"," ").slice(0,16)," \u5230\u671F"):""),children:Bn}):(0,e.jsx)("span",{className:"https-certificate-empty",children:"\u2014"})}),(0,e.jsx)("div",{className:"https-operations",children:(0,e.jsxs)("div",{className:"domain-actions",children:[(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",className:r.settingAction,disabled:ee||(ve?g||!c:!h),"aria-label":"\u4E3A "+K+(ve?" \u7EED\u7B7E\u8BC1\u4E66":" \u90E8\u7F72\u8BC1\u4E66"),onClick:function(){return Ee(ue,ve?"renew":"deploy")},children:ve?"\u7EED\u7B7E":"\u90E8\u7F72"}),ve?(0,e.jsx)(ge.A,{color:"danger",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",className:r.settingAction,disabled:ee,"aria-label":"\u5173\u95ED "+K+" \u7684 HTTPS",onClick:function(){return u.setFieldValue(["bindings",String(ue.id)],void 0)},children:"\u5173\u95ED"}):null]})})]},ue.id)}),!ye.length&&(0,e.jsxs)("div",{className:"domain-empty",children:[(0,e.jsx)("div",{className:"domain-empty-title",children:"\u6682\u65E0\u57DF\u540D"}),(0,e.jsx)("p",{children:"\u8BF7\u5148\u5728\u57DF\u540D\u7BA1\u7406\u4E2D\u6DFB\u52A0\u57DF\u540D\u3002"})]})]})]})})]}),(0,e.jsx)(Kr,{name:"redirect_http",title:"\u81EA\u52A8\u8DF3\u8F6C HTTPS",description:"\u5F00\u542F\u540E\uFF0C\u5DF2\u90E8\u7F72\u8BC1\u4E66\u7684\u57DF\u540D\u4F1A\u5C06 HTTP \u8BBF\u95EE\u8DF3\u8F6C\u5230 HTTPS",styles:r}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u8FDE\u63A5\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"min_version",label:"\u6700\u4F4E TLS \u7248\u672C",children:(0,e.jsx)(De.A,{allowClear:!0,placeholder:"\u81EA\u52A8\u9009\u62E9\u6700\u4F4E\u7248\u672C",options:[{label:"TLS 1.2",value:"tls1.2"},{label:"TLS 1.3",value:"tls1.3"}]})}),(0,e.jsx)(d.A.Item,{name:"max_version",label:"\u6700\u9AD8 TLS \u7248\u672C",children:(0,e.jsx)(De.A,{allowClear:!0,placeholder:"\u81EA\u52A8\u9009\u62E9\u6700\u9AD8\u7248\u672C",options:[{label:"TLS 1.2",value:"tls1.2"},{label:"TLS 1.3",value:"tls1.3"}]})})]})}]})]})]})},eo=function(n){var t=n.site,r=n.styles;return(0,e.jsxs)(qr,{site:t,request:Vi,update:Vi,title:"\u9759\u6001\u7F51\u7AD9\u8BBE\u7F6E",description:"\u9ED8\u8BA4\u914D\u7F6E\u5373\u53EF\u76F4\u63A5\u6258\u7BA1 HTML \u548C\u9759\u6001\u8D44\u6E90",styles:r,normalize:function(l){return{index:Array.isArray(l.index)?l.index:["index.html","index.htm"],browse:!!l.browse,try_files:Array.isArray(l.try_files)?l.try_files:[],hide:Array.isArray(l.hide)?l.hide:[],precompressed:Array.isArray(l.precompressed)?l.precompressed:["zstd","gzip"]}},serialize:function(l){return{index:vn(l.index),browse:!!l.browse,try_files:l.browse?[]:vn(l.try_files),hide:vn(l.hide),precompressed:vn(l.precompressed)}},children:[(0,e.jsx)(d.A.Item,{name:"index",label:"\u9ED8\u8BA4\u9996\u9875",extra:"\u8BBF\u95EE\u76EE\u5F55\u65F6\u6309\u987A\u5E8F\u67E5\u627E\u8FD9\u4E9B\u9996\u9875\u6587\u4EF6",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"index.html, index.htm"})}),(0,e.jsx)(Kr,{name:"browse",title:"\u76EE\u5F55\u6D4F\u89C8",description:"\u5F00\u542F\u540E\u8BBF\u95EE\u76EE\u5F55\u663E\u793A\u6587\u4EF6\u5217\u8868\uFF0C\u4E0D\u518D\u81EA\u52A8\u6253\u5F00\u9996\u9875\u6587\u4EF6",styles:r}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u9AD8\u7EA7\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(l,h){return l.browse!==h.browse},children:function(l){var h=l.getFieldValue;return(0,e.jsx)(d.A.Item,{name:"try_files",label:"\u6587\u4EF6\u5339\u914D\u987A\u5E8F",extra:h("browse")?"\u5F00\u542F\u76EE\u5F55\u6D4F\u89C8\u65F6\u81EA\u52A8\u7559\u7A7A\uFF0C\u4FDD\u8BC1\u5B50\u76EE\u5F55\u53EF\u4EE5\u6B63\u5E38\u8FDB\u5165":"\u7559\u7A7A\u65F6\u6309\u8BF7\u6C42\u8DEF\u5F84\u76F4\u63A5\u67E5\u627E\u6587\u4EF6",children:(0,e.jsx)(De.A,{disabled:!!h("browse"),mode:"tags",tokenSeparators:[","," "],placeholder:"\u4F8B\u5982 {path}, /index.html"})})}}),(0,e.jsx)(d.A.Item,{name:"hide",label:"\u7981\u6B62\u8BBF\u95EE\u7684\u6587\u4EF6",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u4F8B\u5982 .git, .env"})}),(0,e.jsx)(d.A.Item,{name:"precompressed",label:"\u9884\u538B\u7F29\u6587\u4EF6",children:(0,e.jsx)(De.A,{mode:"multiple",placeholder:"\u8BF7\u9009\u62E9\u9884\u538B\u7F29\u683C\u5F0F",options:[{label:"Zstandard",value:"zstd"},{label:"Gzip",value:"gzip"}]})})]})}]})]})},no=function(n){var t=n.site,r=n.styles,a=(0,i.useRef)({}),l=(0,i.useCallback)(function(c){var f,y=ii(c.upstreams),j=String(y[0]||"127.0.0.1:8080"),F=/^unix(?:\+h2c)?\//.test(j),B=Array.isArray(c.versions)?c.versions:[],o=((f=c.headers)===null||f===void 0?void 0:f.request_set)||{},v=Object.keys(o).find(function(S){return S.toLowerCase()==="host"}),N=v?o[v]:"",w=Array.isArray(N)?N[0]:N,u=D()(D()({},c),{},{versions:B,policy:c.policy||"random",retries:Number(c.retries||0),target:F?j.startsWith("unix+h2c/")||B.includes("h2c")?j.replace(/^unix(?:\+h2c)?\//,"unix+h2c/"):j.replace(/^unix\//,""):/^https?:\/\//i.test(j)?j:"".concat(c.scheme||"http","://").concat(j),send_host:w==="{http.request.host}"?"$http_host":String(w||""),backup_upstreams:y.slice(1),try_duration:Dn(c.try_duration),try_interval:Dn(c.try_interval),dial_timeout:Dn(c.dial_timeout),read_timeout:Dn(c.read_timeout),write_timeout:Dn(c.write_timeout),health_interval:Dn(c.health_interval),health_timeout:Dn(c.health_timeout),tls_insecure_skip_verify:c.tls_insecure_skip_verify==null?!0:!!c.tls_insecure_skip_verify});return a.current=u,u},[]),h=(0,i.useCallback)(function(c){var f,y=D()(D()({},a.current),c),j=Nr(y.target),F=j.dial.startsWith("unix/"),B=[j].concat(Je()(vn(y.backup_upstreams).map(function(m){return Nr(m,j.scheme)}))),o=new Map,v=wt()(y.upstreams||[]),N;try{for(v.s();!(N=v.n()).done;){var w=N.value;try{o.set(Nr(w.dial,y.scheme||"http").dial,Number(w.max_requests||0))}catch(m){}}}catch(m){v.e(m)}finally{v.f()}var u=String(y.send_host||"").trim(),S=j.hostname,x=F||S==="localhost"||S.includes(":")||/^\d{1,3}(?:\.\d{1,3}){3}$/.test(S)?"{http.request.host}":S;return{upstreams:vn(B.map(function(m){return m.dial})).map(function(m){return{dial:m,max_requests:o.get(m)||0}}),scheme:j.scheme,versions:B.some(function(m){return m.h2c})?["h2c"]:(y.versions||[]).filter(function(m){return(m!=="h2c"||!F&&j.scheme==="http")&&(m!=="3"||j.scheme==="https")}),policy:y.policy||"random",retries:Number(y.retries||0),try_duration:kn(y.try_duration),try_interval:kn(y.try_interval),dial_timeout:kn(y.dial_timeout),read_timeout:kn(y.read_timeout),write_timeout:kn(y.write_timeout),tls_server_name:j.scheme==="https"?String(y.tls_server_name||"").trim():"",tls_insecure_skip_verify:j.scheme==="https"&&!!y.tls_insecure_skip_verify,health_uri:String(y.health_uri||"").trim(),health_interval:kn(y.health_interval),health_timeout:kn(y.health_timeout),health_status:Number(y.health_status||0),headers:D()(D()({},y.headers),{},{request_set:D()(D()({},Object.fromEntries(Object.entries(((f=y.headers)===null||f===void 0?void 0:f.request_set)||{}).filter(function(m){var p=C()(m,1),_=p[0];return _.toLowerCase()!=="host"}))),{},{Host:[u?u==="$http_host"?"{http.request.host}":u:x]})})}},[]);return(0,e.jsxs)(qr,{site:t,request:aa,update:aa,title:"\u53CD\u5411\u4EE3\u7406\u8BBE\u7F6E",description:"\u5C06\u7F51\u7AD9\u8BF7\u6C42\u8F6C\u53D1\u5230\u53E6\u4E00\u4E2A HTTP \u670D\u52A1\u6216 UNIX \u5957\u63A5\u5B57",styles:r,normalize:l,serialize:h,children:[(0,e.jsx)(d.A.Item,{name:"target",label:"\u76EE\u6807\u5730\u5740",extra:"\u81EA\u52A8\u8BC6\u522B HTTP\u3001HTTPS \u548C UNIX \u5957\u63A5\u5B57\uFF1B\u7F51\u5740\u4E0D\u586B\u534F\u8BAE\u65F6\u4F7F\u7528 HTTP\u3002",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u76EE\u6807\u5730\u5740"},{validator:(function(){var c=ne()(b()().mark(function y(j,F){return b()().wrap(function(o){for(;;)switch(o.prev=o.next){case 0:String(F||"").trim()&&Nr(F);case 1:case"end":return o.stop()}},y)}));function f(y,j){return c.apply(this,arguments)}return f})()}],children:(0,e.jsx)(ze.A,{placeholder:"http://127.0.0.1:15700 \u6216 /tmp/panel.sock"})}),(0,e.jsx)(d.A.Item,{name:"send_host",label:"\u53D1\u9001\u57DF\u540D\uFF08Host\uFF09",extra:"\u7559\u7A7A\u4F7F\u7528\u76EE\u6807\u57DF\u540D\uFF1B\u4EE3\u7406\u672C\u673A IP \u65F6\u53EF\u586B\u5199 $http_host \u900F\u4F20\u8BBF\u95EE\u57DF\u540D",children:(0,e.jsx)(ze.A,{placeholder:"\u81EA\u52A8\u4F7F\u7528\u76EE\u6807\u57DF\u540D"})}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u9AD8\u7EA7\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"backup_upstreams",label:"\u5907\u7528\u4E0A\u6E38",dependencies:["target"],rules:[function(c){var f=c.getFieldValue;return{validator:(function(){var y=ne()(b()().mark(function F(B,o){var v,N;return b()().wrap(function(u){for(;;)switch(u.prev=u.next){case 0:if(v=vn(o),v.length){u.next=3;break}return u.abrupt("return");case 3:if(N=Nr(f("target")),!v.some(function(S){return Nr(S,N.scheme).scheme!==N.scheme})){u.next=6;break}throw new Error("\u5907\u7528\u4E0A\u6E38\u5FC5\u987B\u4E0E\u76EE\u6807\u5730\u5740\u4F7F\u7528\u76F8\u540C\u534F\u8BAE");case 6:case"end":return u.stop()}},F)}));function j(F,B){return y.apply(this,arguments)}return j})()}}],children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u53EF\u6DFB\u52A0\u76F8\u540C\u534F\u8BAE\u7684\u5907\u7528\u5730\u5740"})}),(0,e.jsxs)("div",{className:r.fieldGrid,children:[(0,e.jsx)(d.A.Item,{name:"policy",label:"\u8D1F\u8F7D\u5747\u8861",children:(0,e.jsx)(De.A,{placeholder:"\u8BF7\u9009\u62E9\u8D1F\u8F7D\u5747\u8861\u7B56\u7565",options:[{label:"\u968F\u673A",value:"random"},{label:"\u8F6E\u8BE2",value:"round_robin"},{label:"\u6700\u5C11\u8FDE\u63A5",value:"least_conn"},{label:"\u5BA2\u6237\u7AEF IP",value:"client_ip_hash"},{label:"URI",value:"uri_hash"}]})}),(0,e.jsx)(d.A.Item,{name:"retries",label:"\u91CD\u8BD5\u6B21\u6570",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u91CD\u8BD5\u6B21\u6570",min:0,precision:0})})]}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"dial_timeout",label:"\u8FDE\u63A5\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8FDE\u63A5\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"read_timeout",label:"\u8BFB\u53D6\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BFB\u53D6\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"write_timeout",label:"\u5199\u5165\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u5199\u5165\u8D85\u65F6",min:0})})]}),(0,e.jsxs)("div",{className:r.fieldGrid,children:[(0,e.jsx)(d.A.Item,{name:"tls_server_name",label:"\u4E0A\u6E38 TLS SNI",children:(0,e.jsx)(ze.A,{placeholder:"\u9ED8\u8BA4\u4F7F\u7528\u76EE\u6807\u57DF\u540D"})}),(0,e.jsx)(d.A.Item,{name:"tls_insecure_skip_verify",label:"\u5FFD\u7565\u4E0A\u6E38\u8BC1\u4E66\u6821\u9A8C",valuePropName:"checked",children:(0,e.jsx)(Dr.A,{})})]}),(0,e.jsx)(_r,{full:!0,name:"health_uri",label:"\u5065\u5EB7\u68C0\u67E5\u8DEF\u5F84",children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 /health"})}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"health_interval",label:"\u95F4\u9694\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u68C0\u67E5\u95F4\u9694",min:0})}),(0,e.jsx)(d.A.Item,{name:"health_timeout",label:"\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u68C0\u67E5\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"health_status",label:"\u72B6\u6001\u7801",children:(0,e.jsx)(qe.A,{placeholder:"\u671F\u671B\u72B6\u6001\u7801",min:100,max:599,precision:0})})]})]})}]})]})},ro=function(n){var t=n.site,r=n.styles;return(0,e.jsxs)(qr,{site:t,request:qi,update:qi,title:"PHP \u8BBE\u7F6E",description:"\u9009\u62E9 PHP-FPM \u670D\u52A1\u5373\u53EF\u8FD0\u884C PHP \u7F51\u7AD9",styles:r,normalize:function(l){var h=ii(l.upstreams).map(function(c){return String(c).replace(/^unix(?:\+h2c)?\//,"")});return D()(D()({policy:"random",index:"index.php",split_path:[".php"],try_files:[],capture_stderr:!0},l),{},{fpm_address:h[0]||"127.0.0.1:9000",backup_upstreams:h.slice(1),dial_timeout:Dn(l.dial_timeout),read_timeout:Dn(l.read_timeout),write_timeout:Dn(l.write_timeout),health_interval:Dn(l.health_interval),health_timeout:Dn(l.health_timeout),env_lines:Uu(l.env)})},serialize:function(l){var h=[l.fpm_address].concat(Je()(vn(l.backup_upstreams))).map(function(f){var y=String(f||"").trim();return y.startsWith("/")?"unix/".concat(y):y}),c=vn(l.split_path);return{upstreams:Vs(h),policy:l.policy||"random",index:String(l.index||"index.php").trim(),split_path:c.length?c:[".php"],try_files:vn(l.try_files),capture_stderr:!!l.capture_stderr,resolve_root_symlink:!!l.resolve_root_symlink,dial_timeout:kn(l.dial_timeout),read_timeout:kn(l.read_timeout),write_timeout:kn(l.write_timeout),health_uri:String(l.health_uri||"").trim(),health_interval:kn(l.health_interval),health_timeout:kn(l.health_timeout),health_status:Number(l.health_status||0),env:Gu(l.env_lines)}},children:[(0,e.jsx)(d.A.Item,{name:"fpm_address",label:"PHP-FPM \u5730\u5740",extra:"\u53EF\u586B\u5199 TCP \u5730\u5740\u6216 UNIX \u5957\u63A5\u5B57",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165 PHP-FPM \u5730\u5740"}],children:(0,e.jsx)(ze.A,{placeholder:"127.0.0.1:9000"})}),(0,e.jsx)(d.A.Item,{name:"index",label:"\u9ED8\u8BA4\u5165\u53E3\u6587\u4EF6",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u5165\u53E3\u6587\u4EF6"}],children:(0,e.jsx)(ze.A,{placeholder:"index.php"})}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u9AD8\u7EA7\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"backup_upstreams",label:"\u5907\u7528 PHP-FPM",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u8F93\u5165\u5907\u7528 PHP-FPM \u5730\u5740"})}),(0,e.jsxs)("div",{className:r.fieldGrid,children:[(0,e.jsx)(d.A.Item,{name:"split_path",label:"\u811A\u672C\u6269\u5C55\u540D",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u4F8B\u5982 .php"})}),(0,e.jsx)(d.A.Item,{name:"policy",label:"\u8D1F\u8F7D\u5747\u8861",children:(0,e.jsx)(De.A,{placeholder:"\u8BF7\u9009\u62E9\u8D1F\u8F7D\u5747\u8861\u7B56\u7565",options:[{label:"\u968F\u673A",value:"random"},{label:"\u8F6E\u8BE2",value:"round_robin"},{label:"\u6700\u5C11\u8FDE\u63A5",value:"least_conn"}]})})]}),(0,e.jsx)(d.A.Item,{name:"try_files",label:"\u6587\u4EF6\u5339\u914D\u987A\u5E8F",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u7559\u7A7A\u4F7F\u7528\u9ED8\u8BA4\u89C4\u5219"})}),(0,e.jsxs)("div",{className:r.fieldGrid,children:[(0,e.jsx)(d.A.Item,{name:"capture_stderr",label:"\u8BB0\u5F55 PHP \u9519\u8BEF",valuePropName:"checked",children:(0,e.jsx)(Dr.A,{})}),(0,e.jsx)(d.A.Item,{name:"resolve_root_symlink",label:"\u89E3\u6790\u76EE\u5F55\u8F6F\u94FE\u63A5",valuePropName:"checked",children:(0,e.jsx)(Dr.A,{})})]}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"dial_timeout",label:"\u8FDE\u63A5\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8FDE\u63A5\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"read_timeout",label:"\u8BFB\u53D6\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BFB\u53D6\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"write_timeout",label:"\u5199\u5165\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u5199\u5165\u8D85\u65F6",min:0})})]}),(0,e.jsx)(_r,{full:!0,name:"health_uri",label:"\u5065\u5EB7\u68C0\u67E5\u8DEF\u5F84",children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 /status"})}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"health_interval",label:"\u95F4\u9694\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u68C0\u67E5\u95F4\u9694",min:0})}),(0,e.jsx)(d.A.Item,{name:"health_timeout",label:"\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u68C0\u67E5\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"health_status",label:"\u72B6\u6001\u7801",children:(0,e.jsx)(qe.A,{placeholder:"\u671F\u671B\u72B6\u6001\u7801",min:100,max:599,precision:0})})]}),(0,e.jsx)(d.A.Item,{name:"env_lines",label:"\u73AF\u5883\u53D8\u91CF",extra:"\u6BCF\u884C\u4E00\u4E2A KEY=VALUE",children:(0,e.jsx)(ze.A.TextArea,{placeholder:"\u4F8B\u5982 APP_ENV=production",autoSize:{minRows:3,maxRows:8}})})]})}]})]})},to=function(n){var t=n.site,r=n.styles,a=un.A.useApp(),l=a.message,h=d.A.useForm(),c=C()(h,1),f=c[0],y=(0,i.useRef)(0),j=(0,i.useRef)({}),F=(0,i.useRef)(null),B=(0,i.useState)(!0),o=C()(B,2),v=o[0],N=o[1],w=(0,i.useState)(!1),u=C()(w,2),S=u[0],x=u[1],m=(0,i.useState)(!1),p=C()(m,2),_=p[0],A=p[1],E=Ut(v||S,_),k=(0,i.useCallback)(ne()(b()().mark(function U(){var P,M,I,$,q,z,he,R,g,L,J,V,ee,fe,Y,de;return b()().wrap(function(ye){for(;;)switch(ye.prev=ye.next){case 0:return P=++y.current,N(!0),x(!1),ye.prev=3,ye.next=6,Promise.all([ia({body:{action:"get",id:t.id}}),aa({body:{action:"get",id:t.id}})]);case 6:if($=ye.sent,q=C()($,2),z=q[0],he=q[1],y.current===P){ye.next=12;break}return ye.abrupt("return");case 12:if(!((z==null?void 0:z.code)!==200||(he==null?void 0:he.code)!==200)){ye.next=16;break}return x(!0),l.error((z==null?void 0:z.code)!==200?(z==null?void 0:z.message)||"\u83B7\u53D6\u8FD0\u884C\u914D\u7F6E\u5931\u8D25":(he==null?void 0:he.message)||"\u83B7\u53D6\u672C\u5730\u4EE3\u7406\u5931\u8D25"),ye.abrupt("return");case 16:R=(z==null?void 0:z.data)||{},g=(he==null?void 0:he.data)||{},L=String(((M=g.upstreams)===null||M===void 0||(M=M[0])===null||M===void 0?void 0:M.dial)||"127.0.0.1:8080"),J=L.match(/:(\d+)$/),V=((I=g.headers)===null||I===void 0?void 0:I.request_set)||{},ee=Object.keys(V).find(function(Ce){return Ce.toLowerCase()==="host"}),fe=ee?V[ee]:"",Y=Array.isArray(fe)?fe[0]:fe,de={command:R.command||"",port:Number((J==null?void 0:J[1])||8080),environment_lines:Uu(R.environment),max_retries:Number(R.max_retries||3),restart_delay:Dn(R.restart_delay)||2,stop_timeout:Dn(R.stop_timeout)||10,send_host:Y==="{http.request.host}"?"$http_host":String(Y||""),policy:g.policy||"random",retries:Number(g.retries||0),try_duration:Dn(g.try_duration),try_interval:Dn(g.try_interval),dial_timeout:Dn(g.dial_timeout),read_timeout:Dn(g.read_timeout),write_timeout:Dn(g.write_timeout),response_header_timeout:Dn(g.response_header_timeout),flush_interval:Number(g.flush_interval||0)/oa,stream_timeout:Dn(g.stream_timeout),stream_close_delay:Dn(g.stream_close_delay),health_uri:g.health_uri||"",health_interval:Dn(g.health_interval),health_timeout:Dn(g.health_timeout),health_status:Number(g.health_status||0)||void 0},j.current={proxy:g,values:de},f.resetFields(),f.setFieldsValue(de),ye.next=33;break;case 30:ye.prev=30,ye.t0=ye.catch(3),y.current===P&&(x(!0),l.error("\u83B7\u53D6\u9879\u76EE\u8BBE\u7F6E\u5931\u8D25"));case 33:return ye.prev=33,y.current===P&&N(!1),ye.finish(33);case 36:case"end":return ye.stop()}},U,null,[[3,30,33,36]])})),[f,l,t.id]);(0,i.useEffect)(function(){return k(),function(){y.current+=1}},[k]);var T=(function(){var U=ne()(b()().mark(function P(){var M,I,$,q,z,he,R,g,L,J,V,ee;return b()().wrap(function(Y){for(;;)switch(Y.prev=Y.next){case 0:if(!(_||v||S)){Y.next=2;break}return Y.abrupt("return");case 2:return M=++y.current,I=D()(D()({},j.current.values),f.getFieldsValue(!0)),$=j.current.proxy||{},q=Array.isArray($.upstreams)?$.upstreams:[],z=q[0]||{},he={upstreams:[D()(D()({max_requests:0},z),{},{dial:Number(I.port)===j.current.values.port&&z.dial?z.dial:"127.0.0.1:".concat(Number(I.port))})].concat(Je()(q.slice(1))),policy:I.policy||"random",retries:Number(I.retries||0),try_duration:kn(I.try_duration),try_interval:kn(I.try_interval),dial_timeout:kn(I.dial_timeout),read_timeout:kn(I.read_timeout),write_timeout:kn(I.write_timeout),response_header_timeout:kn(I.response_header_timeout),flush_interval:Number(I.flush_interval||0)*oa,stream_timeout:kn(I.stream_timeout),stream_close_delay:kn(I.stream_close_delay),health_uri:String(I.health_uri||"").trim(),health_interval:kn(I.health_interval),health_timeout:kn(I.health_timeout),health_status:Number(I.health_status||0)},R=String(I.send_host||"").trim(),R!==j.current.values.send_host&&(he.headers={request_set:D()(D()({},Object.fromEntries(Object.entries(((g=$.headers)===null||g===void 0?void 0:g.request_set)||{}).filter(function(de){var xe=C()(de,1),ye=xe[0];return ye.toLowerCase()!=="host"}))),R?{Host:[R==="$http_host"||R==="$host"?"{http.request.host}":R]}:{})}),A(!0),L=!1,J=!1,Y.prev=13,Y.next=16,ia({body:{action:"set",id:t.id,config:{command:String(I.command||"").trim(),environment:Gu(I.environment_lines),max_retries:Number(I.max_retries),restart_delay:kn(I.restart_delay),stop_timeout:kn(I.stop_timeout)}}});case 16:if(V=Y.sent,y.current===M){Y.next=19;break}return Y.abrupt("return");case 19:if((V==null?void 0:V.code)===200){Y.next=22;break}return l.error((V==null?void 0:V.message)||"\u4FDD\u5B58\u8FD0\u884C\u914D\u7F6E\u5931\u8D25"),Y.abrupt("return");case 22:return J=!0,Y.next=25,aa({body:{action:"set",id:t.id,config:he}});case 25:if(ee=Y.sent,y.current===M){Y.next=28;break}return Y.abrupt("return");case 28:(ee==null?void 0:ee.code)===200?(l.success((ee==null?void 0:ee.message)||"\u4FDD\u5B58\u6210\u529F"),L=!0):l.error("\u8FD0\u884C\u914D\u7F6E\u5DF2\u4FDD\u5B58\uFF0C\u672C\u5730\u4EE3\u7406\u4FDD\u5B58\u5931\u8D25\uFF1A".concat((ee==null?void 0:ee.message)||"\u8BF7\u91CD\u8BD5")),Y.next=34;break;case 31:Y.prev=31,Y.t0=Y.catch(13),y.current===M&&l.error(J?"\u8FD0\u884C\u914D\u7F6E\u5DF2\u4FDD\u5B58\uFF0C\u672C\u5730\u4EE3\u7406\u4FDD\u5B58\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5":"\u4FDD\u5B58\u8FD0\u884C\u914D\u7F6E\u5931\u8D25");case 34:return Y.prev=34,y.current===M&&A(!1),Y.finish(34);case 37:L&&y.current===M&&k();case 38:case"end":return Y.stop()}},P,null,[[13,31,34,37]])}));return function(){return U.apply(this,arguments)}})();return S?(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:"\u9879\u76EE\u8BBE\u7F6E\u52A0\u8F7D\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5",children:(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:r.settingAction,icon:(0,e.jsx)(ce.In,{type:"ReloadOutlined"}),onClick:function(){k()},children:"\u91CD\u65B0\u52A0\u8F7D"})}):(0,e.jsxs)(Wt,{loading:v,children:[(0,e.jsx)(Rt,{title:"\u9879\u76EE\u8BBE\u7F6E",description:"\u8BBE\u7F6E\u9879\u76EE\u542F\u52A8\u65B9\u5F0F\u548C\u672C\u673A\u4EE3\u7406\uFF0C\u672A\u7ED1\u5B9A\u57DF\u540D\u65F6\u4E5F\u53EF\u72EC\u7ACB\u8FD0\u884C",styles:r,extra:(0,e.jsx)(ge.A,{color:"default",variant:"filled",className:r.settingAction,style:{flexShrink:0,borderColor:"transparent"},icon:(0,e.jsx)(ce.In,{type:"CodeOutlined"}),disabled:_,onClick:function(){var P;return(P=F.current)===null||P===void 0?void 0:P.open(t,{type:"process"})},children:"\u8FDB\u7A0B\u65E5\u5FD7"})}),(0,e.jsx)(Bs,{siteId:t.id,disabled:_||v},t.id),(0,e.jsxs)(qt,{styles:r,id:E,form:f,onFinish:T,disabled:_,children:[(0,e.jsx)(_r,{wide:!0,name:"command",label:"\u542F\u52A8\u547D\u4EE4",extra:"\u8BF7\u8F93\u5165\u5B8C\u6574\u6267\u884C\u547D\u4EE4\u548C\u6240\u9700\u53C2\u6570",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u542F\u52A8\u547D\u4EE4"}],children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 /www/app/server --port 8080"})}),(0,e.jsx)(_r,{name:"port",label:"\u9879\u76EE\u7AEF\u53E3",extra:"\u9879\u76EE\u5FC5\u987B\u76D1\u542C\u672C\u673A\u7684\u8FD9\u4E2A\u7AEF\u53E3",rules:[{required:!0,message:"\u8BF7\u8F93\u5165\u9879\u76EE\u7AEF\u53E3"}],children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u9879\u76EE\u7AEF\u53E3",min:1,max:65535,precision:0,style:{width:"100%"}})}),(0,e.jsx)(_r,{name:"max_retries",label:"\u5931\u8D25\u91CD\u8BD5\u6B21\u6570",extra:"\u9996\u6B21\u5931\u8D25\u540E\u6700\u591A\u91CD\u8BD5\u6307\u5B9A\u6B21\u6570\uFF0C\u8017\u5C3D\u540E\u505C\u6B62\u81EA\u52A8\u4FDD\u6D3B\uFF0C\u7F51\u7AD9\u548C\u4EE3\u7406\u914D\u7F6E\u4E0D\u53D8\u3002\u53EF\u901A\u8FC7\u4E0A\u65B9\u6309\u94AE\u542F\u52A8\u3001\u505C\u6B62\u6216\u91CD\u542F\u9879\u76EE\uFF1B\u7A33\u5B9A\u8FD0\u884C 1 \u5206\u949F\u540E\u91CD\u7F6E\u6B21\u6570\u3002",rules:[{required:!0,type:"integer",min:1,message:"\u8BF7\u8F93\u5165\u5927\u4E8E 0 \u7684\u6574\u6570"}],children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4 3 \u6B21",min:1,precision:0,style:{width:"100%"}})}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"process",label:"\u8FDB\u7A0B\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"environment_lines",label:"\u73AF\u5883\u53D8\u91CF",extra:"\u6BCF\u884C\u4E00\u4E2A KEY=VALUE",children:(0,e.jsx)(ze.A.TextArea,{placeholder:"\u4F8B\u5982 NODE_ENV=production",autoSize:{minRows:3,maxRows:8}})}),(0,e.jsxs)("div",{className:r.fieldGrid,children:[(0,e.jsx)(d.A.Item,{name:"restart_delay",label:"\u5F02\u5E38\u91CD\u542F\u7B49\u5F85\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u91CD\u542F\u7B49\u5F85\u65F6\u95F4",min:0})}),(0,e.jsx)(d.A.Item,{name:"stop_timeout",label:"\u505C\u6B62\u7B49\u5F85\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u505C\u6B62\u7B49\u5F85\u65F6\u95F4",min:1})})]})]})},{key:"proxy",label:"\u672C\u673A\u53CD\u5411\u4EE3\u7406",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(_r,{wide:!0,name:"send_host",label:"\u53D1\u9001\u57DF\u540D\uFF08Host\uFF09",extra:"\u7559\u7A7A\u900F\u4F20\u8BBF\u95EE\u57DF\u540D\uFF0C\u4E5F\u53EF\u586B\u5199\u6307\u5B9A\u57DF\u540D\u6216 $http_host\u3002",children:(0,e.jsx)(ze.A,{placeholder:"\u9ED8\u8BA4\u900F\u4F20\u8BBF\u95EE\u57DF\u540D"})}),(0,e.jsx)(_r,{name:"policy",label:"\u8D1F\u8F7D\u5747\u8861",children:(0,e.jsx)(De.A,{placeholder:"\u8BF7\u9009\u62E9\u8D1F\u8F7D\u5747\u8861\u7B56\u7565",options:[{label:"\u968F\u673A",value:"random"},{label:"\u8F6E\u8BE2",value:"round_robin"},{label:"\u6700\u5C11\u8FDE\u63A5",value:"least_conn"},{label:"\u5BA2\u6237\u7AEF IP",value:"client_ip_hash"},{label:"URI",value:"uri_hash"}]})}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"retries",label:"\u8BF7\u6C42\u91CD\u8BD5\u6B21\u6570",extra:"\u4E0A\u6E38\u8BF7\u6C42\u5931\u8D25\u65F6\u7684\u91CD\u8BD5\u6B21\u6570\uFF0C\u4E0E\u8FDB\u7A0B\u4FDD\u6D3B\u6B21\u6570\u72EC\u7ACB\u3002",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u4E0D\u91CD\u8BD5",min:0,precision:0})}),(0,e.jsx)(d.A.Item,{name:"try_interval",label:"\u91CD\u8BD5\u95F4\u9694\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u95F4\u9694",min:0})}),(0,e.jsx)(d.A.Item,{name:"try_duration",label:"\u91CD\u8BD5\u603B\u65F6\u957F\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"0 \u6309\u91CD\u8BD5\u6B21\u6570",min:0})})]}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"dial_timeout",label:"\u8FDE\u63A5\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"read_timeout",label:"\u8BFB\u53D6\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"write_timeout",label:"\u5199\u5165\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u8D85\u65F6",min:0})})]}),(0,e.jsx)(_r,{name:"response_header_timeout",label:"\u54CD\u5E94\u5934\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u7B49\u5F85\u4E0A\u6E38\u54CD\u5E94\u5934\u7684\u65F6\u95F4",min:0})}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"flush_interval",label:"\u5237\u65B0\u95F4\u9694\uFF08\u79D2\uFF09",extra:"0 \u4F7F\u7528\u9ED8\u8BA4\u7B56\u7565\uFF0C\u8D1F\u6570\u8868\u793A\u7ACB\u5373\u5237\u65B0\u54CD\u5E94\u3002",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u5237\u65B0\u7B56\u7565"})}),(0,e.jsx)(d.A.Item,{name:"stream_timeout",label:"\u6D41\u5F0F\u8D85\u65F6\uFF08\u79D2\uFF09",extra:"\u6D41\u5F0F\u8FDE\u63A5\u7684\u6700\u957F\u6301\u7EED\u65F6\u95F4\uFF0C0 \u4E3A\u4E0D\u9650\u5236\u3002",children:(0,e.jsx)(qe.A,{placeholder:"0 \u4E3A\u4E0D\u9650\u5236",min:0})}),(0,e.jsx)(d.A.Item,{name:"stream_close_delay",label:"\u6D41\u5F0F\u5173\u95ED\u5EF6\u8FDF\uFF08\u79D2\uFF09",extra:"\u914D\u7F6E\u91CD\u8F7D\u540E\uFF0C\u5EF6\u8FDF\u5173\u95ED\u5DF2\u6709\u6D41\u5F0F\u8FDE\u63A5\u3002",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u7ACB\u5373\u5173\u95ED",min:0})})]}),(0,e.jsx)(_r,{full:!0,name:"health_uri",label:"\u5065\u5EB7\u68C0\u67E5\u8DEF\u5F84",extra:"\u7559\u7A7A\u5173\u95ED\u4E3B\u52A8\u68C0\u67E5\uFF1B\u914D\u7F6E\u540E\u5B9A\u671F\u8BF7\u6C42\u6B64\u8DEF\u5F84\uFF0C\u5931\u8D25\u65F6\u6682\u505C\u5411\u8BE5\u4E0A\u6E38\u8F6C\u53D1\u8BF7\u6C42\u3002",rules:[{validator:(function(){var U=ne()(b()().mark(function M(I,$){var q;return b()().wrap(function(he){for(;;)switch(he.prev=he.next){case 0:if(q=String($||"").trim(),!(q&&!q.startsWith("/"))){he.next=3;break}throw new Error("\u5065\u5EB7\u68C0\u67E5\u8DEF\u5F84\u987B\u4EE5 / \u5F00\u5934");case 3:case"end":return he.stop()}},M)}));function P(M,I){return U.apply(this,arguments)}return P})()}],children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 /health"})}),(0,e.jsxs)("div",{className:"".concat(r.fieldGrid," setting-grid-numeric"),children:[(0,e.jsx)(d.A.Item,{name:"health_interval",label:"\u68C0\u67E5\u95F4\u9694\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u95F4\u9694",min:0})}),(0,e.jsx)(d.A.Item,{name:"health_timeout",label:"\u68C0\u67E5\u8D85\u65F6\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4\u8D85\u65F6",min:0})}),(0,e.jsx)(d.A.Item,{name:"health_status",label:"\u671F\u671B\u72B6\u6001\u7801",children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4 200",min:100,max:599,precision:0})})]})]})}]})]}),(0,e.jsx)(sa.A,{ref:F,request:Ri,title:"\u8FDB\u7A0B\u65E5\u5FD7"})]})},ao=function(n){var t,r,a,l=n.site,h=n.styles,c=n.hideHeader,f=c===void 0?!1:c,y=(0,ce.DP)(),j=!!(!(y==null||(t=y.isCompactTheme)===null||t===void 0)&&t.call(y)),F=!!(!(y==null||(r=y.isDarkMode)===null||r===void 0)&&r.call(y)||!(y==null||(a=y.isDarkTheme)===null||a===void 0)&&a.call(y)),B=(0,i.useRef)(null);return(0,e.jsxs)(qr,{site:l,request:Pi,update:Pi,title:"Web \u9632\u706B\u5899",description:"\u8BC6\u522B SQL \u6CE8\u5165\u3001\u8DE8\u7AD9\u811A\u672C\u3001\u6076\u610F\u626B\u63CF\u7B49\u5E38\u89C1\u653B\u51FB",styles:h,hideHeader:f,onFinishFailed:function(v){var N,w=v.errorFields;return(N=B.current)===null||N===void 0?void 0:N.reveal(w)},normalize:function(v){var N,w={enabled:!0,paranoia_level:1,detection_paranoia_level:1,inbound_anomaly_score_threshold:5,outbound_anomaly_score_threshold:4,sampling_percentage:100},u=v.owasp||{},S=Array.isArray(v.rules)&&v.rules.some(function(E){return(E==null?void 0:E.enabled)&&((E==null?void 0:E.directive)||Array.isArray(E==null?void 0:E.conditions)&&E.conditions.length||Array.isArray(E==null?void 0:E.groups)&&E.groups.length)})||Array.isArray(v.directives)&&v.directives.some(function(E){return String(E||"").trim()}),x=!!(u.enabled||u.paranoia_level||u.inbound_anomaly_score_threshold),m=x||S?D()(D()({},w),u):w;!m.enabled&&!S&&(m.enabled=!0);var p=((N=Object.entries(ai).find(function(E){var k=C()(E,2),T=k[1];return Object.entries(T).every(function(U){var P=C()(U,2),M=P[0],I=P[1];return Number(m[M])===I})}))===null||N===void 0?void 0:N[0])||"custom",_=[],A=[];return(Array.isArray(v.rules)?v.rules:[]).forEach(function(E){var k=String((E==null?void 0:E.directive)||"").trim();if(Array.isArray(E==null?void 0:E.groups)&&E.groups.length||Array.isArray(E==null?void 0:E.conditions)&&E.conditions.length){var T=Array.isArray(E==null?void 0:E.groups)&&E.groups.length?E.groups.map(Ls):[{match:E.match==="all"?"all":"any",conditions:E.conditions.map(Lu)}];_.push({id:String(E.id||""),name:String(E.name||""),enabled:!!E.enabled,match:E.match==="all"?"all":"any",action:E.action==="log"?"log":"block",groups:T});return}A.push(D()(D()({},E),{},{directive:k}))}),D()(D()({},v),{},{enabled:!!v.enabled,mode:v.mode||"block",block_page:String(v.block_page||""),audit_log:v.audit_log==null?!0:!!v.audit_log,request_body_limit_mb:Number(v.request_body_limit||16*1024*1024)/1024/1024,owasp:m,protection_level:p,rules:_,expert_rules:A,directives_text:(Array.isArray(v.directives)?v.directives:[]).join(`
`)})},serialize:function(v,N){var w=N.getFieldsValue(!0),u=D()(D()({},w.owasp||{}),ai[w.protection_level]||{});u.detection_paranoia_level=Math.max(Number(u.paranoia_level||1),Number(u.detection_paranoia_level||0));var S=(Array.isArray(w.rules)?w.rules:[]).map(function(x){return{id:String(x.id||"").trim(),name:String(x.name||"").trim(),enabled:!!x.enabled,match:x.match==="all"?"all":"any",action:x.action==="log"?"log":"block",groups:(Array.isArray(x.groups)?x.groups:[]).map(function(m){return{match:(m==null?void 0:m.match)==="any"?"any":"all",conditions:(Array.isArray(m==null?void 0:m.conditions)?m.conditions:[]).map(function(p){var _=String((p==null?void 0:p.operator)||"contains"),A=String((p==null?void 0:p.target)||"").trim();return{target:A,operator:_,value:Mu.has(_)?"":String((p==null?void 0:p.value)||"").trim(),negated:(p==null?void 0:p.negated)==="not",ignore_case:A!=="ARGS_COMBINED_SIZE"&&Hu.has(_)&&!!(p!=null&&p.ignore_case)}})}}),conditions:[],directive:""}});return(Array.isArray(w.expert_rules)?w.expert_rules:[]).forEach(function(x){return S.push({id:String(x.id||"").trim(),name:String(x.name||"").trim(),enabled:!!x.enabled,directive:String(x.directive||"").trim()})}),{enabled:!!w.enabled,mode:w.mode||"block",block_page:String(w.block_page||""),audit_log:!!w.audit_log,request_body_limit:Math.round(Number(w.request_body_limit_mb||0)*1024*1024),owasp:u,directives:String(w.directives_text||"").split(`
`).map(function(x){return x.trim()}).filter(Boolean),rules:S}},children:[(0,e.jsx)(Kr,{name:"enabled",title:"Web \u9632\u706B\u5899",description:"\u81EA\u52A8\u68C0\u67E5\u8FDB\u5165\u7F51\u7AD9\u7684\u8BF7\u6C42\uFF0C\u8BC6\u522B\u5E38\u89C1\u653B\u51FB\u884C\u4E3A",styles:h}),(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(v,N){return v.enabled!==N.enabled},children:function(v){var N=v.getFieldValue;return N("enabled")?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(Kr,{name:["owasp","enabled"],title:"\u542F\u7528\u5185\u7F6E\u901A\u7528\u89C4\u5219",description:"\u8986\u76D6\u5E38\u89C1\u6CE8\u5165\u3001\u8DE8\u7AD9\u811A\u672C\u3001\u534F\u8BAE\u5F02\u5E38\u548C\u6076\u610F\u6587\u4EF6\u8BF7\u6C42",styles:h}),(0,e.jsx)(d.A.Item,{name:"mode",label:"\u9632\u62A4\u6A21\u5F0F",extra:"\u5EFA\u8BAE\u5148\u89C2\u5BDF\u65E5\u5FD7\uFF0C\u786E\u8BA4\u4E1A\u52A1\u6B63\u5E38\u540E\u518D\u62E6\u622A",children:(0,e.jsx)(Lt.A,{className:h.wafMode,options:[{label:"\u62E6\u622A\u653B\u51FB",value:"block"},{label:"\u4EC5\u89C2\u5BDF",value:"detection"}]})}),(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(u,S){var x,m;return((x=u.owasp)===null||x===void 0?void 0:x.enabled)!==((m=S.owasp)===null||m===void 0?void 0:m.enabled)||u.protection_level!==S.protection_level},children:function(u){var S=u.getFieldValue,x=u.setFieldsValue;return S(["owasp","enabled"])?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"protection_level",label:"\u9632\u62A4\u5F3A\u5EA6",extra:"\u5747\u8861\u6A21\u5F0F\u517C\u987E\u5B89\u5168\u6027\u548C\u517C\u5BB9\u6027",children:(0,e.jsx)(Lt.A,{className:h.wafMode,options:[{label:"\u57FA\u7840",value:"loose"},{label:"\u5747\u8861",value:"recommended"},{label:"\u4E25\u683C",value:"strict"},{label:"\u81EA\u5B9A\u4E49",value:"custom"}],onChange:function(p){var _=ai[String(p)];_&&x({owasp:D()(D()({},S("owasp")),_)})}})}),S("protection_level")==="custom"&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:["owasp","paranoia_level"],label:"\u89C4\u5219\u654F\u611F\u7EA7\u522B",extra:"1 \u5230 4 \u7EA7\uFF0C\u7EA7\u522B\u8D8A\u9AD8\u68C0\u67E5\u8D8A\u4E25\u683C\uFF0C\u4E5F\u8D8A\u5BB9\u6613\u8BEF\u62E6\u622A\u3002",rules:[{required:!0,type:"integer",min:1,max:4,message:"\u8BF7\u8F93\u5165 1 \u5230 4 \u7684\u6574\u6570"}],children:(0,e.jsx)(qe.A,{placeholder:"1 - 4",min:1,max:4,precision:0,style:{width:"100%"}})}),(0,e.jsx)(d.A.Item,{name:["owasp","inbound_anomaly_score_threshold"],label:"\u8BF7\u6C42\u62E6\u622A\u9608\u503C\uFF08\u5206\uFF09",extra:"\u8BBF\u5BA2\u8BF7\u6C42\u547D\u4E2D\u89C4\u5219\u4F1A\u7D2F\u52A0\u5F02\u5E38\u5206\u6570\uFF0C\u8FBE\u5230\u6216\u8D85\u8FC7\u6B64\u503C\u65F6\u62E6\u622A\u8BF7\u6C42\u3002\u8D8A\u5C0F\u8D8A\u4E25\u683C\uFF0C\u5747\u8861\u503C\u4E3A 5\u3002",rules:[{required:!0,type:"integer",min:1,message:"\u8BF7\u8F93\u5165\u5927\u4E8E 0 \u7684\u6574\u6570"}],children:(0,e.jsx)(qe.A,{placeholder:"\u5747\u8861\u503C 5",min:1,precision:0,style:{width:"100%"}})}),(0,e.jsx)(d.A.Item,{name:["owasp","outbound_anomaly_score_threshold"],label:"\u54CD\u5E94\u62E6\u622A\u9608\u503C\uFF08\u5206\uFF09",extra:"\u7F51\u7AD9\u8FD4\u56DE\u7684\u54CD\u5E94\u547D\u4E2D\u89C4\u5219\u4F1A\u7D2F\u52A0\u5F02\u5E38\u5206\u6570\uFF0C\u8FBE\u5230\u6216\u8D85\u8FC7\u6B64\u503C\u65F6\u62E6\u622A\u54CD\u5E94\u3002\u8D8A\u5C0F\u8D8A\u4E25\u683C\uFF0C\u5747\u8861\u503C\u4E3A 4\u3002",rules:[{required:!0,type:"integer",min:1,message:"\u8BF7\u8F93\u5165\u5927\u4E8E 0 \u7684\u6574\u6570"}],children:(0,e.jsx)(qe.A,{placeholder:"\u5747\u8861\u503C 4",min:1,precision:0,style:{width:"100%"}})}),(0,e.jsx)("div",{className:"ant-form-item-extra",children:"\u201C\u4EC5\u89C2\u5BDF\u201D\u6A21\u5F0F\u53EA\u8BB0\u5F55\u547D\u4E2D\u60C5\u51B5\uFF0C\u4E0D\u6267\u884C\u62E6\u622A\u3002"})]})]}):null}}),(0,e.jsx)(d.A.Item,{name:"request_body_limit_mb",label:"\u8BF7\u6C42\u5185\u5BB9\u68C0\u67E5\u4E0A\u9650\uFF08MB\uFF09",extra:"\u8D85\u8FC7\u6B64\u5927\u5C0F\u7684\u8BF7\u6C42\u6309\u9632\u62A4\u6A21\u5F0F\u5904\u7406",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u68C0\u67E5\u4E0A\u9650",min:0,precision:0,style:{width:"100%",maxWidth:360}})}),(0,e.jsx)(Kr,{name:"audit_log",title:"\u8BB0\u5F55\u547D\u4E2D\u8BE6\u60C5",description:"\u4FDD\u7559\u547D\u4E2D\u89C4\u5219\u3001\u8BBF\u5BA2\u6765\u6E90\u4E0E\u5904\u7F6E\u7ED3\u679C\uFF0C\u7528\u4E8E\u6392\u67E5\u8BEF\u62E6\u622A",styles:h}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"block_page",label:"\u62E6\u622A\u9875\u9762",children:(0,e.jsx)(d.A.Item,{name:"block_page",label:"\u9875\u9762\u5185\u5BB9",extra:"\u652F\u6301\u5B8C\u6574 HTML\uFF1B\u7559\u7A7A\u4F7F\u7528\u7CFB\u7EDF\u9ED8\u8BA4\u9875\u9762\u3002",children:(0,e.jsx)(ru.A,{mode:"html",theme:F?"one_dark":"github",width:"100%",height:j?220:260,fontSize:12,tabSize:4,showPrintMargin:!1,showLineNumbers:!0,wrapEnabled:!0,acePath:Ms,placeholder:"\u7559\u7A7A\u4F7F\u7528\u7CFB\u7EDF\u9ED8\u8BA4\u9875\u9762",className:h.code})})}]}),(0,e.jsx)("div",{className:h.wafRules,children:(0,e.jsx)(d.A.List,{name:"rules",validateTrigger:"onSubmit",rules:[{validator:(function(){var w=ne()(b()().mark(function S(x,m){var p,_,A,E,k,T,U,P,M,I,$,q,z;return b()().wrap(function(R){for(;;)switch(R.prev=R.next){case 0:p=wt()((Array.isArray(m)?m:[]).entries()),R.prev=1,p.s();case 3:if((_=p.n()).done){R.next=46;break}if(A=C()(_.value,2),E=A[0],k=A[1],!(!["any","all"].includes(k==null?void 0:k.match)||!["block","log"].includes(k==null?void 0:k.action))){R.next=7;break}throw new Error("\u89C4\u5219 ".concat(E+1," \u7684\u6267\u884C\u65B9\u5F0F\u65E0\u6548"));case 7:if(!(!Array.isArray(k==null?void 0:k.groups)||k.groups.length===0)){R.next=9;break}throw new Error("\u89C4\u5219 ".concat(E+1," \u81F3\u5C11\u9700\u8981\u4E00\u79CD\u60C5\u51B5"));case 9:T=wt()(k.groups.entries()),R.prev=10,T.s();case 12:if((U=T.n()).done){R.next=36;break}if(P=C()(U.value,2),M=P[0],I=P[1],["any","all"].includes(I==null?void 0:I.match)){R.next=16;break}throw new Error("\u89C4\u5219 ".concat(E+1," \u7684\u60C5\u51B5\u5173\u7CFB\u65E0\u6548"));case 16:if(!(!Array.isArray(I==null?void 0:I.conditions)||I.conditions.length===0)){R.next=18;break}throw new Error("\u89C4\u5219 ".concat(E+1," \u7684\u60C5\u51B5 ").concat(M+1," \u81F3\u5C11\u9700\u8981\u4E00\u4E2A\u6761\u4EF6"));case 18:$=wt()(I.conditions),R.prev=19,z=b()().mark(function g(){var L,J,V,ee;return b()().wrap(function(Y){for(;;)switch(Y.prev=Y.next){case 0:if(L=q.value,J=String((L==null?void 0:L.target)||""),V=String((L==null?void 0:L.operator)||""),!(!Vt.some(function(de){return de.value===J})||!$u.some(function(de){return de.value===V}))){Y.next=5;break}throw new Error("\u89C4\u5219 ".concat(E+1," \u5B58\u5728\u672A\u5B8C\u6210\u7684\u6761\u4EF6"));case 5:if(ee=J==="ARGS_COMBINED_SIZE",!(V==="ip"&&J!=="REMOTE_ADDR"||["greater_than","less_than"].includes(V)&&!ee||ee&&!["equals","greater_than","less_than"].includes(V))){Y.next=8;break}throw new Error("\u89C4\u5219 ".concat(E+1," \u7684\u5339\u914D\u65B9\u5F0F\u4E0D\u9002\u7528\u4E8E\u68C0\u67E5\u4F4D\u7F6E"));case 8:case"end":return Y.stop()}},g)}),$.s();case 22:if((q=$.n()).done){R.next=26;break}return R.delegateYield(z(),"t0",24);case 24:R.next=22;break;case 26:R.next=31;break;case 28:R.prev=28,R.t1=R.catch(19),$.e(R.t1);case 31:return R.prev=31,$.f(),R.finish(31);case 34:R.next=12;break;case 36:R.next=41;break;case 38:R.prev=38,R.t2=R.catch(10),T.e(R.t2);case 41:return R.prev=41,T.f(),R.finish(41);case 44:R.next=3;break;case 46:R.next=51;break;case 48:R.prev=48,R.t3=R.catch(1),p.e(R.t3);case 51:return R.prev=51,p.f(),R.finish(51);case 54:case"end":return R.stop()}},S,null,[[1,48,51,54],[10,38,41,44],[19,28,31,34]])}));function u(S,x){return w.apply(this,arguments)}return u})()}],children:function(u,S,x){var m=S.add,p=S.remove,_=x.errors;return(0,e.jsx)(Qs,{ref:B,fields:u,add:m,remove:p,errors:_,styles:h})}})}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"expert",label:"\u4E13\u5BB6\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"directives_text",label:"\u5E95\u5C42 Coraza \u6307\u4EE4",extra:"\u6BCF\u884C\u4E00\u6761\u539F\u751F Coraza \u914D\u7F6E\uFF1B\u8BED\u6CD5\u9519\u8BEF\u4F1A\u5BFC\u81F4\u7F51\u7AD9\u914D\u7F6E\u65E0\u6CD5\u751F\u6548\uFF0C\u666E\u901A\u7528\u6237\u8BF7\u4FDD\u6301\u4E3A\u7A7A",children:(0,e.jsx)(ze.A.TextArea,{placeholder:"\u666E\u901A\u7528\u6237\u65E0\u9700\u586B\u5199",autoSize:{minRows:3,maxRows:10}})}),(0,e.jsx)(d.A.List,{name:"expert_rules",children:function(u,S){var x=S.remove;return u.length?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)("div",{className:"waf-rules-hint",style:{marginBottom:8},children:"\u4EE5\u4E0B\u65E7\u89C4\u5219\u65E0\u6CD5\u8F6C\u6362\u6210\u53EF\u89C6\u5316\u8868\u5355\uFF0C\u5C06\u6309\u539F\u5185\u5BB9\u4FDD\u7559\u3002"}),u.map(function(m){var p=m.key,_=m.name;return(0,e.jsxs)("div",{className:h.rule,children:[(0,e.jsx)(d.A.Item,{name:[_,"name"],label:"\u89C4\u5219\u540D\u79F0",children:(0,e.jsx)(ze.A,{placeholder:"\u8BF7\u8F93\u5165\u89C4\u5219\u540D\u79F0"})}),(0,e.jsx)(d.A.Item,{name:[_,"id"],label:"\u89C4\u5219 ID",children:(0,e.jsx)(ze.A,{placeholder:"\u8BF7\u8F93\u5165\u89C4\u5219 ID"})}),(0,e.jsx)(d.A.Item,{name:[_,"enabled"],label:"\u542F\u7528",valuePropName:"checked",children:(0,e.jsx)(Dr.A,{})}),(0,e.jsx)(d.A.Item,{name:[_,"directive"],label:"\u539F\u59CB\u89C4\u5219",children:(0,e.jsx)(ze.A.TextArea,{placeholder:"\u8BF7\u8F93\u5165\u539F\u59CB Coraza \u89C4\u5219",autoSize:{minRows:2,maxRows:6}})}),(0,e.jsx)(ge.A,{color:"danger",variant:"text",className:h.ruleAction,size:"small",onClick:function(){return x(_)},children:"\u5220\u9664\u65E7\u89C4\u5219"})]},p)})]}):null}})]})}]})]}):null}})]})},io=function(n){var t=n.site,r=n.styles,a=n.hideHeader,l=a===void 0?!1:a;return(0,e.jsxs)(qr,{site:t,request:zi,update:zi,title:"\u8BBF\u95EE\u9891\u7387",description:"\u9650\u5236\u5355\u4E2A\u8BBF\u5BA2\u5728\u6307\u5B9A\u65F6\u95F4\u5185\u7684\u8BF7\u6C42\u6B21\u6570",styles:r,hideHeader:l,normalize:function(c){return D()(D()({},c),{},{enabled:!!c.enabled,key:!c.key||c.key==="{http.request.client_ip}"?"{http.vars.client_ip}":c.key,window:Dn(c.window)||60,max_events:Number(c.max_events||100),paths:Array.isArray(c.paths)?c.paths:[],methods:Array.isArray(c.methods)?c.methods:[],ipv4_prefix:c.ipv4_prefix==null?32:c.ipv4_prefix,ipv6_prefix:c.ipv6_prefix==null?128:c.ipv6_prefix})},serialize:function(c){return D()(D()({},c),{},{methods:vn(c.methods).filter(function(f){return f!=="*"}),paths:vn(c.paths),window:kn(c.window)})},children:[(0,e.jsx)(Kr,{name:"enabled",title:"\u8BBF\u95EE\u9891\u7387\u9650\u5236",description:"\u9ED8\u8BA4\u6309\u8BBF\u5BA2 IP \u7EDF\u8BA1\u6307\u5B9A\u65F6\u95F4\u5185\u7684\u8BF7\u6C42\u6B21\u6570",styles:r}),(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(c,f){return c.enabled!==f.enabled},children:function(c){var f=c.getFieldValue;return f("enabled")?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"window",label:"\u7EDF\u8BA1\u5468\u671F\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u7EDF\u8BA1\u5468\u671F",min:1,style:{width:"100%"}})}),(0,e.jsx)(d.A.Item,{name:"max_events",label:"\u6700\u591A\u8BF7\u6C42\u6B21\u6570",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u6700\u591A\u8BF7\u6C42\u6B21\u6570",min:1,precision:0,style:{width:"100%"}})}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u9AD8\u7EA7\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"paths",label:"\u9650\u5236\u8DEF\u5F84",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u7559\u7A7A\u8868\u793A\u6574\u4E2A\u7F51\u7AD9"})}),(0,e.jsx)(d.A.Item,{name:"methods",label:"\u8BF7\u6C42\u65B9\u6CD5",children:(0,e.jsx)(De.A,{mode:"multiple",placeholder:"\u7559\u7A7A\u8868\u793A\u5168\u90E8\u65B9\u6CD5",options:ti})}),(0,e.jsx)(d.A.Item,{name:"key",label:"\u7EDF\u8BA1\u5BF9\u8C61",extra:"\u53EF\u9009\u62E9\u5E38\u7528\u5BF9\u8C61\uFF0C\u4E5F\u53EF\u4EE5\u7EC4\u5408\u591A\u4E2A\u53D8\u91CF\u6216\u76F4\u63A5\u8F93\u5165\u56FA\u5B9A\u524D\u7F00\u3002",children:(0,e.jsx)(js.A,{placeholder:"\u8BF7\u9009\u62E9\u6216\u8F93\u5165\u7EDF\u8BA1\u5BF9\u8C61",options:zu.map(function(y){return{value:y.value,label:"".concat(y.label," \xB7 ").concat(y.value)}}),filterOption:function(j,F){return String((F==null?void 0:F.label)||"").toLowerCase().includes(j.toLowerCase())}})}),(0,e.jsxs)("div",{className:r.variableGuide,children:[(0,e.jsx)("div",{className:"variable-guide-title",children:"\u5E38\u7528\u53D8\u91CF"}),zu.map(function(y){return(0,e.jsxs)("div",{className:"variable-guide-row",children:[(0,e.jsx)("code",{children:y.value}),(0,e.jsx)("span",{children:y.description})]},y.value)}),(0,e.jsxs)("div",{className:"variable-guide-row",children:[(0,e.jsx)("code",{children:"{http.request.header.<\u540D\u79F0>}"}),(0,e.jsx)("span",{children:"\u6309\u6307\u5B9A\u8BF7\u6C42\u5934\u5185\u5BB9\u7EDF\u8BA1\uFF0C\u4F8B\u5982 API Key"})]}),(0,e.jsxs)("div",{className:"variable-guide-row",children:[(0,e.jsx)("code",{children:"{http.request.cookie.<\u540D\u79F0>}"}),(0,e.jsx)("span",{children:"\u6309\u6307\u5B9A Cookie \u5185\u5BB9\u7EDF\u8BA1"})]})]}),(0,e.jsx)(d.A.Item,{name:"ipv4_prefix",label:"IPv4 \u524D\u7F00",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165 IPv4 \u524D\u7F00\u957F\u5EA6",min:0,max:32,precision:0,style:{width:"100%"}})}),(0,e.jsx)(d.A.Item,{name:"ipv6_prefix",label:"IPv6 \u524D\u7F00",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165 IPv6 \u524D\u7F00\u957F\u5EA6",min:0,max:128,precision:0,style:{width:"100%"}})})]})}]})]}):null}})]})},uo=function(n){var t=n.site,r=n.styles,a=n.hideHeader,l=a===void 0?!1:a;return(0,e.jsxs)(qr,{site:t,request:Oi,update:Oi,title:"\u5E76\u53D1\u4E0E\u901F\u5EA6",description:"\u9650\u5236\u7AD9\u70B9\u5E76\u53D1\u91CF\u548C\u5355\u4E2A\u8BF7\u6C42\u7684\u4F20\u8F93\u901F\u5EA6",styles:r,hideHeader:l,normalize:function(c){var f=!c.enabled&&!c.max_connections&&!c.max_connections_per_ip&&!c.rate_per_request;return D()(D()({enabled:!1},c),{},{max_connections:f?1e3:Number(c.max_connections||0),max_connections_per_ip:f?50:Number(c.max_connections_per_ip||0),rate_per_request_kb:Number(c.rate_per_request||0)/1024})},serialize:function(c){return{enabled:!!c.enabled,max_connections:Number(c.max_connections||0),max_connections_per_ip:Number(c.max_connections_per_ip||0),rate_per_request:Math.round(Number(c.rate_per_request_kb||0)*1024)}},children:[(0,e.jsx)(Kr,{name:"enabled",title:"\u5E76\u53D1\u4E0E\u901F\u5EA6\u9650\u5236",description:"\u63A7\u5236\u7F51\u7AD9\u603B\u5E76\u53D1\u3001\u5355\u4E2A\u8BBF\u5BA2\u5E76\u53D1\u548C\u54CD\u5E94\u4F20\u8F93\u901F\u5EA6",styles:r}),(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(c,f){return c.enabled!==f.enabled},children:function(c){var f=c.getFieldValue;return f("enabled")?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"max_connections",label:"\u7F51\u7AD9\u6700\u5927\u5E76\u53D1",extra:"0 \u8868\u793A\u4E0D\u9650",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u7F51\u7AD9\u6700\u5927\u5E76\u53D1\u6570",min:0,precision:0,style:{width:"100%"}})}),(0,e.jsx)(d.A.Item,{name:"max_connections_per_ip",label:"\u5355 IP \u6700\u5927\u5E76\u53D1",extra:"0 \u8868\u793A\u4E0D\u9650",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u5355 IP \u6700\u5927\u5E76\u53D1\u6570",min:0,precision:0,style:{width:"100%"}})}),(0,e.jsx)(d.A.Item,{name:"rate_per_request_kb",label:"\u5355\u8BF7\u6C42\u901F\u5EA6\uFF08KB/s\uFF09",extra:"0 \u8868\u793A\u4E0D\u9650",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u5355\u8BF7\u6C42\u901F\u5EA6",min:0,precision:0,style:{width:"100%"}})})]}):null}})]})},lo=function(n){var t=n.site,r=n.styles,a=n.hideHeader,l=a===void 0?!1:a,h=un.A.useApp(),c=h.message,f=h.modal,y=function(F){var B=F.reload;return(0,e.jsx)(ge.A,{color:"danger",variant:"filled",style:{borderColor:"transparent"},className:r.settingAction,size:"small",icon:(0,e.jsx)(ce.In,{type:"ClearOutlined"}),onClick:function(){return f.confirm({title:"\u6E05\u7406\u7F51\u7AD9\u7F13\u5B58",content:"\u786E\u5B9A\u6E05\u7406\u8FD9\u4E2A\u7F51\u7AD9\u7684\u5168\u90E8\u54CD\u5E94\u7F13\u5B58\u5417\uFF1F",okText:"\u6E05\u7406",okButtonProps:{danger:!0},onOk:(function(){var v=ne()(b()().mark(function w(){var u;return b()().wrap(function(x){for(;;)switch(x.prev=x.next){case 0:return x.next=2,Ia({body:{action:"clear",id:t.id}});case 2:u=x.sent,(u==null?void 0:u.code)===200?(c.success((u==null?void 0:u.message)||"\u6E05\u7406\u6210\u529F"),B()):c.error((u==null?void 0:u.message)||"\u6E05\u7406\u5931\u8D25");case 4:case"end":return x.stop()}},w)}));function N(){return v.apply(this,arguments)}return N})()})},children:"\u6E05\u7406\u7F13\u5B58"})};return(0,e.jsxs)(qr,{site:t,request:Ia,update:Ia,title:"\u54CD\u5E94\u7F13\u5B58",description:"\u7F13\u5B58\u5E38\u7528\u54CD\u5E94\uFF0C\u51CF\u5C11\u78C1\u76D8\u6216\u4E0A\u6E38\u670D\u52A1\u538B\u529B",styles:r,hideHeader:l,extra:y,normalize:function(F){return D()(D()({},F),{},{enabled:!!F.enabled,ttl:Dn(F.ttl)||120,stale:Dn(F.stale)||30,methods:Array.isArray(F.methods)&&F.methods.length?F.methods:["GET","HEAD"],key_headers:Array.isArray(F.key_headers)?F.key_headers:[],default_cache_control:F.default_cache_control||"public, max-age=120",max_size_mb:Number(F.max_size||256*1024*1024)/1024/1024})},serialize:function(F){return{enabled:!!F.enabled,methods:vn(F.methods).filter(function(B){return B!=="*"}),key_headers:vn(F.key_headers),ttl:kn(F.ttl),stale:kn(F.stale),max_size:Math.round(Number(F.max_size_mb||0)*1024*1024),default_cache_control:String(F.default_cache_control||"").trim()}},children:[(0,e.jsx)(Kr,{name:"enabled",title:"\u54CD\u5E94\u7F13\u5B58",description:"\u7F13\u5B58\u5E38\u7528\u54CD\u5E94\u4EE5\u964D\u4F4E\u78C1\u76D8\u6216\u4E0A\u6E38\u670D\u52A1\u538B\u529B\uFF0C\u52A8\u6001\u7F51\u7AD9\u8BF7\u8C28\u614E\u5F00\u542F",styles:r}),(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(F,B){return F.enabled!==B.enabled},children:function(F){var B=F.getFieldValue;return B("enabled")?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:r.fieldGrid,children:[(0,e.jsx)(d.A.Item,{name:"ttl",label:"\u7F13\u5B58\u65F6\u95F4\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u7F13\u5B58\u65F6\u95F4",min:1})}),(0,e.jsx)(d.A.Item,{name:"max_size_mb",label:"\u5BB9\u91CF\u4E0A\u9650\uFF08MB\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u5BB9\u91CF\u4E0A\u9650",min:1,precision:0})})]}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u9AD8\u7EA7\u8BBE\u7F6E",children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"stale",label:"\u8FC7\u671F\u5BBD\u9650\uFF08\u79D2\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u8FC7\u671F\u5BBD\u9650\u65F6\u95F4",min:0,style:{width:"100%"}})}),(0,e.jsx)(d.A.Item,{name:"methods",label:"\u7F13\u5B58\u8BF7\u6C42\u65B9\u6CD5",children:(0,e.jsx)(De.A,{mode:"multiple",placeholder:"\u8BF7\u9009\u62E9\u7F13\u5B58\u8BF7\u6C42\u65B9\u6CD5",options:ti})}),(0,e.jsx)(d.A.Item,{name:"key_headers",label:"\u53C2\u4E0E\u7F13\u5B58\u952E\u7684\u8BF7\u6C42\u5934",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u8F93\u5165\u8BF7\u6C42\u5934\u540D\u79F0\u540E\u56DE\u8F66"})}),(0,e.jsx)(d.A.Item,{name:"default_cache_control",label:"\u9ED8\u8BA4 Cache-Control",children:(0,e.jsx)(ze.A,{placeholder:"public, max-age=120"})})]})}]})]}):null}})]})},so=function(n){var t=n.site,r=n.styles,a=n.hideHeader,l=a===void 0?!1:a;return(0,e.jsxs)(qr,{site:t,request:Mi,update:Mi,title:"\u54CD\u5E94\u538B\u7F29",description:"\u81EA\u52A8\u538B\u7F29\u6587\u672C\u3001\u811A\u672C\u548C\u6837\u5F0F\u6587\u4EF6",styles:r,hideHeader:l,normalize:function(c){return D()(D()({enabled:!0,min_length:1024},c),{},{algorithms:Array.isArray(c.algorithms)&&c.algorithms.length?c.algorithms:["zstd","gzip"]})},children:[(0,e.jsx)(Kr,{name:"enabled",title:"\u54CD\u5E94\u538B\u7F29",description:"\u538B\u7F29\u6587\u672C\u3001\u811A\u672C\u548C\u6837\u5F0F\u6587\u4EF6\uFF0C\u63A8\u8350\u4FDD\u6301\u5F00\u542F",styles:r}),(0,e.jsx)(d.A.Item,{noStyle:!0,shouldUpdate:function(c,f){return c.enabled!==f.enabled},children:function(c){var f=c.getFieldValue;return f("enabled")?(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u9AD8\u7EA7\u8BBE\u7F6E",children:(0,e.jsxs)("div",{className:r.fieldGrid,children:[(0,e.jsx)(d.A.Item,{name:"algorithms",label:"\u538B\u7F29\u7B97\u6CD5",children:(0,e.jsx)(De.A,{mode:"multiple",placeholder:"\u8BF7\u9009\u62E9\u538B\u7F29\u7B97\u6CD5",options:[{label:"Zstandard",value:"zstd"},{label:"Gzip",value:"gzip"}]})}),(0,e.jsx)(d.A.Item,{name:"min_length",label:"\u6700\u5C0F\u54CD\u5E94\uFF08\u5B57\u8282\uFF09",children:(0,e.jsx)(qe.A,{placeholder:"\u8BF7\u8F93\u5165\u6700\u5C0F\u54CD\u5E94\u5B57\u8282\u6570",min:0,precision:0})})]})}]}):null}})]})},oo=function(n){var t=n.site,r=n.styles,a=n.hideHeader,l=a===void 0?!1:a,h=(0,i.useState)("request"),c=C()(h,2),f=c[0],y=c[1],j=(0,i.useCallback)(function(B){return{request_rules:qu(B.request_add,B.request_set,B.request_delete),response_rules:qu(B.response_add,B.response_set,B.response_delete)}},[]),F=(0,i.useCallback)(function(B){var o=Wu(B.request_rules),v=Wu(B.response_rules);return{request_add:o.add,request_set:o.set,request_delete:o.delete,response_add:v.add,response_set:v.set,response_delete:v.delete}},[]);return(0,e.jsx)(qr,{site:t,request:$i,update:$i,title:"\u8BF7\u6C42\u4E0E\u54CD\u5E94\u5934",description:"\u6309\u89C4\u5219\u8FFD\u52A0\u3001\u8986\u76D6\u6216\u5220\u9664\u8BF7\u6C42\u5934\u548C\u54CD\u5E94\u5934",styles:r,hideHeader:l,normalize:j,serialize:F,onFinishFailed:function(o){var v,N=o.errorFields,w=N==null||(v=N[0])===null||v===void 0||(v=v.name)===null||v===void 0?void 0:v[0];(w==="request_rules"||w==="response_rules")&&y(w==="request_rules"?"request":"response")},children:(0,e.jsxs)("div",{className:r.headerPane,children:[(0,e.jsx)(Lt.A,{className:"header-switcher",block:!0,value:f,onChange:y,options:[{label:"\u8BF7\u6C42\u5934",value:"request"},{label:"\u54CD\u5E94\u5934",value:"response"}]}),(0,e.jsx)("div",{hidden:f!=="request",children:(0,e.jsx)(Qu,{name:"request_rules",label:"\u8BF7\u6C42\u5934",styles:r})}),(0,e.jsx)("div",{hidden:f!=="response",children:(0,e.jsx)(Qu,{name:"response_rules",label:"\u54CD\u5E94\u5934",styles:r})})]})})},co=function(n){var t,r=n.field,a=n.index,l=n.count,h=n.move,c=n.remove,f=n.domainOptions,y=n.errors,j=n.styles,F=d.A.useFormInstance(),B=["redirects",r.name],o=d.A.useWatch(B,F)||F.getFieldValue(B)||{},v=(0,i.useState)(a===0||!o.target),N=C()(v,2),w=N[0],u=N[1],S=(0,i.useRef)(null),x=vn(o.domains),m=o.path_scope==="custom",p=m?vn(o.paths):[],_=m?String(p[0]||"/old/*"):"/news/123",A=_.replace(/\*/g,"example"),E=String(x[0]||((t=f[0])===null||t===void 0?void 0:t.value)||"example.com").replace(/^\*\./,"www."),k=String(o.target||"").trim(),T="".concat(E).concat(A,"?from=site"),U=k?o.preserve_uri?"".concat(k.replace(/\/+$/,"")).concat(A,"?from=site"):k:"\u586B\u5199\u76EE\u6807\u5730\u5740\u540E\u663E\u793A",P=[301,308].includes(Number(o.status))?"\u6C38\u4E45\u8DF3\u8F6C":"\u4E34\u65F6\u8DF3\u8F6C",M="".concat(x.length?x.join("\u3001"):"\u5168\u90E8\u57DF\u540D"," \xB7 ").concat(m?p.join("\u3001")||"\u5F85\u9009\u62E9\u8DEF\u5F84":"\u6240\u6709\u8DEF\u5F84");return(0,i.useEffect)(function(){var I;a>0&&!o.target&&((I=S.current)===null||I===void 0||I.scrollIntoView({block:"start",inline:"nearest",behavior:"smooth"}))},[]),(0,i.useEffect)(function(){if(y.some(function(q){var z;return((z=q.name)===null||z===void 0?void 0:z[1])===r.name})){var I,$;u(!0),((I=y[0])===null||I===void 0||(I=I.name)===null||I===void 0?void 0:I[1])===r.name&&(($=S.current)===null||$===void 0||$.scrollIntoView({block:"start",inline:"nearest"}))}},[y,r.name]),(0,e.jsxs)("div",{className:"redirect-card",children:[(0,e.jsxs)("div",{className:"redirect-heading",ref:S,children:[(0,e.jsxs)("button",{type:"button",className:"redirect-summary","aria-expanded":w,"aria-label":"".concat(w?"\u6536\u8D77":"\u5C55\u5F00","\u91CD\u5B9A\u5411\u89C4\u5219 ").concat(a+1),onClick:function(){return u(!w)},children:[(0,e.jsx)(ce.In,{type:w?"DownOutlined":"RightOutlined"}),(0,e.jsxs)("span",{className:"redirect-summary-text",children:[(0,e.jsxs)("span",{className:"redirect-title",children:[(0,e.jsx)("span",{children:o.name||"\u91CD\u5B9A\u5411 ".concat(a+1)}),(0,e.jsx)("span",{className:"redirect-status",children:o.enabled===!1?"\u5DF2\u505C\u7528":P})]}),(0,e.jsxs)("span",{className:"redirect-destination",title:"".concat(M," \u2192 ").concat(k||"\u5F85\u586B\u5199\u76EE\u6807\u5730\u5740"),children:[M," \u2192 ",k||"\u5F85\u586B\u5199\u76EE\u6807\u5730\u5740"]})]})]}),(0,e.jsxs)("div",{className:"redirect-controls",children:[(0,e.jsx)(d.A.Item,{name:[r.name,"enabled"],valuePropName:"checked",noStyle:!0,children:(0,e.jsx)(Dr.A,{size:"small","aria-label":"\u542F\u7528\u91CD\u5B9A\u5411\u89C4\u5219 ".concat(a+1)})}),(0,e.jsx)(gn.A,{menu:{items:[{key:"up",label:"\u4E0A\u79FB\u89C4\u5219",disabled:a===0},{key:"down",label:"\u4E0B\u79FB\u89C4\u5219",disabled:a===l-1},{type:"divider"},{key:"delete",label:"\u5220\u9664\u89C4\u5219",danger:!0}],onClick:function($){var q=$.key;q==="up"&&h(r.name,r.name-1),q==="down"&&h(r.name,r.name+1),q==="delete"&&c(r.name)}},children:(0,e.jsx)(ge.A,{color:"default",variant:"text",className:j.ruleAction,size:"small","aria-label":"\u91CD\u5B9A\u5411\u89C4\u5219 ".concat(a+1," \u64CD\u4F5C"),icon:(0,e.jsx)(ce.In,{type:"MoreOutlined"})})})]})]}),(0,e.jsxs)("div",{className:"redirect-body",hidden:!w,children:[(0,e.jsx)(d.A.Item,{name:[r.name,"name"],label:"\u89C4\u5219\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u89C4\u5219\u540D\u79F0"},function(I){var $=I.getFieldValue;return{validator:(function(){var q=ne()(b()().mark(function he(R,g){var L;return b()().wrap(function(V){for(;;)switch(V.prev=V.next){case 0:if(L=String(g||"").trim().toLowerCase(),!(L&&($("redirects")||[]).some(function(ee,fe){return fe!==r.name&&String((ee==null?void 0:ee.name)||"").trim().toLowerCase()===L}))){V.next=3;break}throw new Error("\u89C4\u5219\u540D\u79F0\u5DF2\u5B58\u5728\uFF0C\u8BF7\u6362\u4E00\u4E2A\u540D\u79F0");case 3:case"end":return V.stop()}},he)}));function z(he,R){return q.apply(this,arguments)}return z})()}}],children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982\uFF1A\u65E7\u57DF\u540D\u8DF3\u8F6C\uFF0C\u65B9\u4FBF\u4EE5\u540E\u8FA8\u8BA4",maxLength:100})}),(0,e.jsx)(d.A.Item,{name:[r.name,"path_scope"],label:"\u54EA\u4E9B\u8BBF\u95EE\u9700\u8981\u8DF3\u8F6C",children:(0,e.jsx)(Lt.A,{className:"redirect-scope",options:[{label:"\u6574\u4E2A\u7F51\u7AD9",value:"all"},{label:"\u6307\u5B9A\u8DEF\u5F84",value:"custom"}]})}),m&&(0,e.jsx)(d.A.Item,{name:[r.name,"paths"],label:"\u8BBF\u95EE\u8DEF\u5F84",extra:"\u4F8B\u5982 /old \u53EA\u5339\u914D\u8BE5\u5730\u5740\uFF1B/old/* \u5339\u914D\u5176\u4E0B\u7684\u9875\u9762\u3002\u53EF\u6DFB\u52A0\u591A\u4E2A\uFF0C\u6309\u56DE\u8F66\u786E\u8BA4\u3002",rules:[{validator:(function(){var I=ne()(b()().mark(function q(z,he){var R;return b()().wrap(function(L){for(;;)switch(L.prev=L.next){case 0:if(R=vn(he),R.length){L.next=3;break}throw new Error("\u8BF7\u81F3\u5C11\u586B\u5199\u4E00\u4E2A\u8BBF\u95EE\u8DEF\u5F84");case 3:if(!R.some(function(J){return!J.startsWith("/")||/[\r\n]/.test(J)})){L.next=5;break}throw new Error("\u8BBF\u95EE\u8DEF\u5F84\u5FC5\u987B\u4EE5 / \u5F00\u5934\u4E14\u4E0D\u80FD\u5305\u542B\u6362\u884C");case 5:case"end":return L.stop()}},q)}));function $(q,z){return I.apply(this,arguments)}return $})()}],children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u4F8B\u5982 /old \u6216 /old/*"})}),(0,e.jsx)(d.A.Item,{name:[r.name,"domains"],label:"\u751F\u6548\u57DF\u540D",extra:"\u4E0D\u9009\u62E9\u5219\u5BF9\u672C\u7AD9\u5168\u90E8\u57DF\u540D\u751F\u6548\uFF0C\u4E5F\u4F1A\u5305\u542B\u4EE5\u540E\u65B0\u589E\u7684\u57DF\u540D\u3002",children:(0,e.jsx)(De.A,{mode:"multiple",showSearch:!1,placeholder:"\u672C\u7AD9\u5168\u90E8\u57DF\u540D",options:f})}),(0,e.jsx)(d.A.Item,{name:[r.name,"target"],label:"\u8DF3\u8F6C\u5230\u54EA\u91CC",className:"setting-wide",extra:"\u586B\u5199\u5B8C\u6574\u7F51\u5740\uFF0C\u6216\u7528 / \u5F00\u5934\u8DF3\u8F6C\u5230\u672C\u7AD9\u9875\u9762\u3002",dependencies:[["redirects",r.name,"preserve_uri"]],rules:[{required:!0,whitespace:!0,message:"\u8BF7\u586B\u5199\u8DF3\u8F6C\u540E\u7684\u76EE\u6807\u5730\u5740"},function(I){var $=I.getFieldValue;return{validator:(function(){var q=ne()(b()().mark(function he(R,g){var L,J,V;return b()().wrap(function(fe){for(;;)switch(fe.prev=fe.next){case 0:if(L=String(g||"").trim(),L){fe.next=3;break}return fe.abrupt("return");case 3:if(!/[\r\n{}]/.test(L)){fe.next=5;break}throw new Error("\u76EE\u6807\u5730\u5740\u5305\u542B\u65E0\u6548\u5B57\u7B26");case 5:if(J=!!$(["redirects",r.name,"preserve_uri"]),!/^\/(?!\/)/.test(L)){fe.next=14;break}if(!L.includes("\\")){fe.next=9;break}throw new Error("\u7AD9\u5185\u8DEF\u5F84\u4E0D\u80FD\u5305\u542B\u53CD\u659C\u6760");case 9:if(!(J&&L==="/")){fe.next=11;break}throw new Error("\u8DF3\u8F6C\u5230\u9996\u9875\u65F6\uFF0C\u8BF7\u9009\u62E9\u300C\u53EA\u8DF3\u8F6C\u5230\u76EE\u6807\u300D");case 11:if(!(J&&/[?#]/.test(L))){fe.next=13;break}throw new Error("\u76EE\u6807\u5DF2\u5305\u542B\u53C2\u6570\uFF0C\u8BF7\u9009\u62E9\u300C\u53EA\u8DF3\u8F6C\u5230\u76EE\u6807\u300D");case 13:return fe.abrupt("return");case 14:if(/^https?:\/\/[^/?#\\\s]/i.test(L)){fe.next=16;break}throw new Error("\u8BF7\u8F93\u5165\u5B8C\u6574\u7F51\u5740\uFF08\u542B http:// \u6216 https://\uFF09\u6216 / \u5F00\u5934\u7684\u7AD9\u5185\u8DEF\u5F84");case 16:fe.prev=16,V=new URL(L),fe.next=23;break;case 20:throw fe.prev=20,fe.t0=fe.catch(16),new Error("\u8BF7\u8F93\u5165\u5B8C\u6574\u7F51\u5740\uFF08\u542B http:// \u6216 https://\uFF09\u6216 / \u5F00\u5934\u7684\u7AD9\u5185\u8DEF\u5F84");case 23:if(!(!["http:","https:"].includes(V.protocol)||V.username||V.password)){fe.next=25;break}throw new Error("\u8BF7\u8F93\u5165 HTTP\u3001HTTPS \u7F51\u5740\u6216\u7AD9\u5185\u8DEF\u5F84");case 25:if(!(J&&(V.search||V.hash))){fe.next=27;break}throw new Error("\u76EE\u6807\u5DF2\u5305\u542B\u53C2\u6570\u6216\u951A\u70B9\uFF0C\u8BF7\u9009\u62E9\u300C\u53EA\u8DF3\u8F6C\u5230\u76EE\u6807\u300D");case 27:case"end":return fe.stop()}},he,null,[[16,20]])}));function z(he,R){return q.apply(this,arguments)}return z})()}}],children:(0,e.jsx)(ze.A,{placeholder:"https://new.example.com \u6216 /new-page"})}),(0,e.jsx)(d.A.Item,{name:[r.name,"preserve_uri"],label:"\u8DF3\u8F6C\u540E\u7684\u8DEF\u5F84",getValueProps:function($){return{value:$?"preserve":"target"}},getValueFromEvent:function($){return $==="preserve"},children:(0,e.jsx)(De.A,{placeholder:"\u8BF7\u9009\u62E9\u8DEF\u5F84\u5904\u7406\u65B9\u5F0F",options:[{label:"\u4FDD\u7559\u539F\u8DEF\u5F84\u548C\u53C2\u6570",value:"preserve"},{label:"\u53EA\u8DF3\u8F6C\u5230\u76EE\u6807",value:"target"}]})}),(0,e.jsxs)("div",{className:"redirect-example","aria-live":"polite",children:[(0,e.jsx)("div",{className:"redirect-example-title",children:"\u8DF3\u8F6C\u793A\u4F8B"}),(0,e.jsxs)("div",{className:"redirect-example-row",children:[(0,e.jsx)("span",{children:"\u8BBF\u95EE"}),(0,e.jsx)("code",{children:T})]}),(0,e.jsxs)("div",{className:"redirect-example-row",children:[(0,e.jsx)("span",{children:"\u8DF3\u8F6C\u5230"}),(0,e.jsx)("code",{children:U})]}),(0,e.jsx)("div",{className:"redirect-example-note",children:o.preserve_uri?"\u5728\u76EE\u6807\u5730\u5740\u540E\u8FFD\u52A0\u5B8C\u6574\u539F\u8DEF\u5F84\u548C\u53C2\u6570\uFF0C\u4E0D\u4F1A\u79FB\u9664\u5339\u914D\u7684\u8DEF\u5F84\u524D\u7F00\u3002":"\u76F4\u63A5\u4F7F\u7528\u76EE\u6807\u5730\u5740\uFF0C\u4E0D\u9644\u52A0\u539F\u8DEF\u5F84\u548C\u53C2\u6570\u3002"})]}),(0,e.jsx)(Xn.A,{ghost:!0,expandIconPlacement:"end",items:[{key:"advanced",label:"\u66F4\u591A\u8BBE\u7F6E \xB7 ".concat(P,"\uFF08").concat(o.status||302,"\uFF09"),forceRender:!0,children:(0,e.jsx)(e.Fragment,{children:(0,e.jsx)(d.A.Item,{name:[r.name,"status"],label:"\u8DF3\u8F6C\u65B9\u5F0F",extra:P==="\u6C38\u4E45\u8DF3\u8F6C"?"\u9002\u5408\u57DF\u540D\u8FC1\u79FB\u3001\u65E7\u94FE\u63A5\u66FF\u6362\uFF0C\u6D4F\u89C8\u5668\u53EF\u80FD\u8BB0\u4F4F\u8DF3\u8F6C\u7ED3\u679C\u3002":"\u9002\u5408\u4E34\u65F6\u8C03\u6574\u6216\u6D4B\u8BD5\uFF0C\u4EE5\u540E\u4ECD\u53EF\u6539\u56DE\u539F\u5730\u5740\u3002",children:(0,e.jsx)(De.A,{placeholder:"\u8BF7\u9009\u62E9\u8DF3\u8F6C\u65B9\u5F0F",options:[{label:"\u4E34\u65F6\u8DF3\u8F6C\uFF08302\uFF09",value:302},{label:"\u6C38\u4E45\u8DF3\u8F6C\uFF08301\uFF09",value:301},{label:"\u4E34\u65F6\u8DF3\u8F6C\uFF0C\u4FDD\u7559\u8BF7\u6C42\u65B9\u6CD5\uFF08307\uFF09",value:307},{label:"\u6C38\u4E45\u8DF3\u8F6C\uFF0C\u4FDD\u7559\u8BF7\u6C42\u65B9\u6CD5\uFF08308\uFF09",value:308}]})})})}]})]})]})},po=function(n){var t=n.fields,r=n.add,a=n.remove,l=n.move,h=n.domainOptions,c=n.errors,f=n.onChange,y=n.styles,j=d.A.useFormInstance(),F=function(){for(var v=new Set((j.getFieldValue("redirects")||[]).map(function(w){return String((w==null?void 0:w.name)||"").trim().toLowerCase()})),N=1;v.has("\u91CD\u5B9A\u5411 ".concat(N));)N+=1;f(),r({enabled:!0,name:"\u91CD\u5B9A\u5411 ".concat(N),domains:[],path_scope:"all",paths:[],target:"",status:302,preserve_uri:!0})},B=(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:y.settingAction,size:"small",icon:(0,e.jsx)(ce.In,{type:"PlusOutlined"}),onClick:F,children:"\u6DFB\u52A0\u89C4\u5219"});return(0,e.jsx)("div",{className:y.redirectRules,children:t.length?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:"redirect-toolbar",children:[(0,e.jsx)("span",{children:"\u4ECE\u4E0A\u5230\u4E0B\u5339\u914D\uFF0C\u547D\u4E2D\u7B2C\u4E00\u6761\u540E\u8DF3\u8F6C"}),B]}),t.map(function(o,v){return(0,e.jsx)(co,{field:o,index:v,count:t.length,move:function(w,u){f(),l(w,u)},remove:function(w){f(),a(w)},domainOptions:h,errors:c,styles:y},o.key)})]}):(0,e.jsxs)("div",{className:"redirect-empty",children:[(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:"\u6682\u65E0\u6570\u636E"}),B]})})},mo=function(n){var t=n.site,r=n.styles,a=un.A.useApp(),l=a.message,h=(0,i.useState)([]),c=C()(h,2),f=c[0],y=c[1],j=(0,i.useState)([]),F=C()(j,2),B=F[0],o=F[1],v=(0,i.useCallback)(function(w){return{redirects:(Array.isArray(w)?w:[]).map(function(u){return D()(D()({},u),{},{enabled:u.enabled!==!1,domains:Array.isArray(u.domains)?u.domains:[],path_scope:!Array.isArray(u.paths)||!u.paths.length||u.paths.includes("/*")?"all":"custom",paths:Array.isArray(u.paths)?u.paths:[],status:[301,302,307,308].includes(Number(u.status))?Number(u.status):302,preserve_uri:u.preserve_uri!==!1})})}},[]),N=(0,i.useCallback)(function(w){return(Array.isArray(w.redirects)?w.redirects:[]).map(function(u){return{enabled:!!u.enabled,name:String(u.name||"").trim(),domains:vn(u.domains),paths:u.path_scope==="all"?["/*"]:vn(u.paths),target:String(u.target||"").trim(),status:Number(u.status||302),preserve_uri:!!u.preserve_uri}})},[]);return(0,i.useEffect)(function(){var w=!0;return y([]),Na({body:{action:"get",id:t.id}}).then(function(u){w&&((u==null?void 0:u.code)===200?y((Array.isArray(u==null?void 0:u.data)?u==null?void 0:u.data.filter(Boolean):[]).map(function(S){return{label:S.domain,value:S.domain}})):l.error((u==null?void 0:u.message)||"\u83B7\u53D6\u7F51\u7AD9\u57DF\u540D\u5931\u8D25"))}).catch(function(){w&&l.error("\u83B7\u53D6\u7F51\u7AD9\u57DF\u540D\u5931\u8D25")}),function(){w=!1}},[l,t.id]),(0,e.jsx)(qr,{site:t,request:Hi,update:Hi,title:"\u91CD\u5B9A\u5411",description:"\u5C06\u8BBF\u5BA2\u4ECE\u65E7\u57DF\u540D\u6216\u65E7\u9875\u9762\uFF0C\u81EA\u52A8\u5F15\u5BFC\u5230\u65B0\u7684\u5730\u5740",styles:r,normalize:v,serialize:N,onFinishFailed:function(u){var S=u.errorFields;return o(S)},children:(0,e.jsx)(d.A.List,{name:"redirects",children:function(u,S){var x=S.add,m=S.remove,p=S.move;return(0,e.jsx)(po,{fields:u,add:x,remove:m,move:p,domainOptions:f,errors:B,onChange:function(){return o([])},styles:r})}})})},vo=function(n){var t,r,a=n.field,l=n.index,h=n.count,c=n.move,f=n.remove,y=n.copy,j=n.errors,F=n.styles,B=d.A.useFormInstance(),o=["routes",a.name],v=d.A.useWatch(o,{form:B,preserve:!0})||B.getFieldValue(o)||{},N=(0,i.useState)(l===0||!!v._new),w=C()(N,2),u=w[0],S=w[1],x=(0,i.useState)([]),m=C()(x,2),p=m[0],_=m[1],A=(0,i.useRef)(null),E=v.kind||"proxy",k=E==="proxy"?"\u53CD\u5411\u4EE3\u7406":E==="response"?"\u56FA\u5B9A\u54CD\u5E94":"\u9759\u6001\u6587\u4EF6",T=vn(v.paths),U=E==="proxy"?vn(v.proxy_upstreams).join("\u3001")||"\u5F85\u586B\u5199\u76EE\u6807\u670D\u52A1":E==="response"?"\u8FD4\u56DE ".concat(v.response_status||200," \xB7 ").concat(v.response_body||"\u7A7A\u767D\u5185\u5BB9"):v.root||"\u5F85\u9009\u62E9\u6587\u4EF6\u76EE\u5F55",P="".concat(T.length?T.join("\u3001"):"\u5168\u90E8\u8DEF\u5F84"," \u2192 ").concat(U);return(0,i.useEffect)(function(){var M;v._new&&((M=A.current)===null||M===void 0||M.scrollIntoView({block:"nearest",inline:"nearest"}))},[]),(0,i.useEffect)(function(){var M,I,$=j.filter(function(q){var z;return((z=q.name)===null||z===void 0?void 0:z[1])===a.name});$.length&&(S(!0),$.some(function(q){var z;return["methods","strip_prefix","rewrite","index","try_files"].includes((z=q.name)===null||z===void 0?void 0:z[2])})&&_(["advanced"]),((M=j[0])===null||M===void 0||(M=M.name)===null||M===void 0?void 0:M[1])===a.name&&((I=A.current)===null||I===void 0||I.scrollIntoView({block:"nearest",inline:"nearest"})))},[j,a.name]),(0,e.jsxs)("div",{className:"route-card",children:[(0,e.jsxs)("div",{className:"route-heading",ref:A,children:[(0,e.jsxs)("button",{className:"route-summary",type:"button","aria-expanded":u,"aria-label":"".concat(u?"\u6536\u8D77":"\u5C55\u5F00","\u8DEF\u7531 ").concat(l+1),onClick:function(){return S(!u)},children:[(0,e.jsx)(ce.In,{type:u?"DownOutlined":"RightOutlined"}),(0,e.jsxs)("span",{className:"route-summary-text",children:[(0,e.jsxs)("span",{className:"route-title",children:[(0,e.jsx)("span",{children:v.name||"\u8DEF\u7531 ".concat(l+1)}),(0,e.jsx)("span",{className:"route-kind",children:k})]}),(0,e.jsx)("span",{className:"route-destination",title:P,children:P})]})]}),(0,e.jsx)(gn.A,{menu:{items:[{key:"up",label:"\u4E0A\u79FB\u89C4\u5219",disabled:l===0},{key:"down",label:"\u4E0B\u79FB\u89C4\u5219",disabled:l===h-1},{key:"copy",label:"\u590D\u5236\u89C4\u5219"},{type:"divider"},{key:"delete",label:"\u5220\u9664\u89C4\u5219",danger:!0}],onClick:function(I){var $=I.key;$==="up"&&c(a.name,a.name-1),$==="down"&&c(a.name,a.name+1),$==="copy"&&y(v),$==="delete"&&f(a.name)}},children:(0,e.jsx)(ge.A,{color:"default",variant:"text",className:F.ruleAction,size:"small","aria-label":"\u8DEF\u7531 ".concat(l+1," \u64CD\u4F5C"),icon:(0,e.jsx)(ce.In,{type:"MoreOutlined"})})})]}),(0,e.jsxs)("div",{className:"route-body",hidden:!u,children:[(0,e.jsx)(d.A.Item,{name:[a.name,"name"],label:"\u89C4\u5219\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u89C4\u5219\u540D\u79F0"}],children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982\uFF1A\u63A5\u53E3\u8F6C\u53D1\u3001\u4E0B\u8F7D\u76EE\u5F55",maxLength:100})}),(0,e.jsx)(d.A.Item,{name:[a.name,"paths"],label:"\u54EA\u4E9B\u8BBF\u95EE\u9700\u8981\u5904\u7406",extra:"\u4F8B\u5982 /api \u53EA\u5339\u914D\u8BE5\u5730\u5740\uFF1B/api/* \u5339\u914D\u5176\u4E0B\u7684\u9875\u9762\u3002\u7559\u7A7A\u8868\u793A\u5168\u90E8\u8DEF\u5F84\u3002",dependencies:[["routes",a.name,"methods"]],rules:[function(M){var I=M.getFieldValue;return{validator:(function(){var $=ne()(b()().mark(function z(he,R){var g;return b()().wrap(function(J){for(;;)switch(J.prev=J.next){case 0:if(g=vn(R),!g.some(function(V){return!V.startsWith("/")||/[\r\n]/.test(V)})){J.next=3;break}throw new Error("\u8BBF\u95EE\u8DEF\u5F84\u5FC5\u987B\u4EE5 / \u5F00\u5934\u4E14\u4E0D\u80FD\u5305\u542B\u6362\u884C");case 3:if(!(!g.length&&!vn(I(["routes",a.name,"methods"])).length&&l<h-1)){J.next=5;break}throw new Error("\u5339\u914D\u5168\u90E8\u8BBF\u95EE\u7684\u89C4\u5219\u8BF7\u79FB\u5230\u6700\u540E");case 5:case"end":return J.stop()}},z)}));function q(z,he){return $.apply(this,arguments)}return q})()}}],children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u4F8B\u5982 /api/*\uFF0C\u6309\u56DE\u8F66\u6DFB\u52A0"})}),(0,e.jsx)(d.A.Item,{name:[a.name,"kind"],label:"\u5982\u4F55\u5904\u7406\u8BBF\u95EE",children:(0,e.jsx)(De.A,{showSearch:!1,placeholder:"\u8BF7\u9009\u62E9\u5904\u7406\u65B9\u5F0F",options:[{label:"\u8F6C\u53D1\u5230\u5176\u4ED6\u670D\u52A1",value:"proxy"},{label:"\u63D0\u4F9B\u76EE\u5F55\u4E2D\u7684\u6587\u4EF6",value:"static"},{label:"\u8FD4\u56DE\u56FA\u5B9A\u5185\u5BB9",value:"response"}]})}),E==="proxy"?(0,e.jsx)(e.Fragment,{children:(0,e.jsx)(d.A.Item,{name:[a.name,"proxy_upstreams"],label:"\u76EE\u6807\u670D\u52A1",extra:((t=v.proxy)===null||t===void 0?void 0:t.transport)==="fastcgi"?"\u586B\u5199 PHP-FPM \u5730\u5740\u6216\u5957\u63A5\u5B57\u8DEF\u5F84\uFF0C\u4FDD\u7559\u5DF2\u6709 FastCGI \u5904\u7406\u914D\u7F6E\u3002":"\u81EA\u52A8\u8BC6\u522B HTTP\u3001HTTPS \u548C\u5957\u63A5\u5B57\u3002\u53EF\u586B\u5199\u591A\u4E2A\u5730\u5740\uFF0C\u6309\u56DE\u8F66\u6DFB\u52A0\uFF0C\u6240\u6709\u76EE\u6807\u5E94\u4F7F\u7528\u76F8\u540C\u534F\u8BAE\u3002",rules:[{validator:(function(){var M=ne()(b()().mark(function $(q,z){var he,R,g,L;return b()().wrap(function(V){for(;;)switch(V.prev=V.next){case 0:if(R=vn(z),R.length){V.next=3;break}throw new Error("\u8BF7\u81F3\u5C11\u586B\u5199\u4E00\u4E2A\u76EE\u6807\u670D\u52A1\u5730\u5740");case 3:if(((he=v.proxy)===null||he===void 0?void 0:he.transport)!=="fastcgi"){V.next=7;break}if(!R.some(function(ee){return/[\r\n\0]/.test(ee)||ee.includes("://")})){V.next=6;break}throw new Error("FastCGI \u5730\u5740\u4E0D\u80FD\u5E26 HTTP \u6216 HTTPS \u534F\u8BAE");case 6:return V.abrupt("return");case 7:if(g=Nr(R[0]),L=R.map(function(ee){return Nr(ee,g.scheme)}),!L.some(function(ee){return ee.scheme!==L[0].scheme})){V.next=11;break}throw new Error("\u6240\u6709\u76EE\u6807\u670D\u52A1\u9700\u8981\u4F7F\u7528\u76F8\u540C\u534F\u8BAE");case 11:case"end":return V.stop()}},$)}));function I($,q){return M.apply(this,arguments)}return I})()}],children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:((r=v.proxy)===null||r===void 0?void 0:r.transport)==="fastcgi"?"\u4F8B\u5982 127.0.0.1:9000 \u6216 /tmp/php.sock":"\u4F8B\u5982 https://example.com \u6216 /tmp/app.sock"})})}):E==="response"?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{className:"setting-wide",name:[a.name,"response_body"],label:"\u8FD4\u56DE\u5185\u5BB9",extra:"\u53EF\u586B\u5199\u6587\u672C\u6216 HTML\uFF0C\u7559\u7A7A\u5219\u53EA\u8FD4\u56DE\u72B6\u6001\u7801\u3002",children:(0,e.jsx)(ze.A.TextArea,{placeholder:"\u4F8B\u5982\uFF1A\u670D\u52A1\u7EF4\u62A4\u4E2D\uFF0C\u8BF7\u7A0D\u540E\u518D\u8BD5",autoSize:{minRows:3,maxRows:8}})}),(0,e.jsx)(d.A.Item,{name:[a.name,"response_status"],label:"\u54CD\u5E94\u72B6\u6001\u7801",rules:[{type:"number",min:100,max:599,message:"\u8BF7\u8F93\u5165 100 \u81F3 599 \u7684\u72B6\u6001\u7801"}],children:(0,e.jsx)(qe.A,{placeholder:"\u9ED8\u8BA4 200\uFF0C\u53EF\u586B\u5199 100 \u81F3 599",min:100,max:599,precision:0,style:{width:"100%"}})})]}):(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{className:"site-file-picker",name:[a.name,"root"],label:"\u6587\u4EF6\u6240\u5728\u76EE\u5F55",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u9009\u62E9\u6587\u4EF6\u6240\u5728\u76EE\u5F55"}],children:(0,e.jsx)(Br.A,{buttonProps:{color:"default",variant:"filled",style:{borderColor:"transparent"}},mode:"directory",placeholder:"\u8BF7\u8F93\u5165\u6216\u9009\u62E9\u6587\u4EF6\u76EE\u5F55"})}),(0,e.jsx)(Kr,{name:[a.name,"browse"],title:"\u5141\u8BB8\u6D4F\u89C8\u76EE\u5F55",description:"\u76F4\u63A5\u663E\u793A\u6587\u4EF6\u5217\u8868\uFF0C\u4E0D\u518D\u81EA\u52A8\u6253\u5F00\u9996\u9875\u6587\u4EF6",styles:F})]}),(0,e.jsx)(Xn.A,{ghost:!0,activeKey:p,onChange:function(I){return _(Array.isArray(I)?I:[I])},expandIconPlacement:"end",items:[{key:"advanced",label:"\u66F4\u591A\u8BBE\u7F6E",forceRender:!0,children:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:[a.name,"methods"],label:"\u8BF7\u6C42\u65B9\u6CD5",extra:"\u901A\u5E38\u65E0\u9700\u9009\u62E9\uFF0C\u7559\u7A7A\u5373\u53EF\u5904\u7406\u6240\u6709\u8BF7\u6C42\u65B9\u6CD5\u3002",children:(0,e.jsx)(De.A,{mode:"multiple",showSearch:!1,placeholder:"\u5168\u90E8\u8BF7\u6C42\u65B9\u6CD5",options:ti})}),(0,e.jsx)(d.A.Item,{name:[a.name,"strip_prefix"],label:"\u79FB\u9664\u8DEF\u5F84\u524D\u7F00",extra:"\u4F8B\u5982\u586B\u5199 /api\uFF0C\u8BBF\u95EE /api/users \u65F6\u4F1A\u6309 /users \u5904\u7406\u3002",rules:[{validator:(function(){var M=ne()(b()().mark(function $(q,z){return b()().wrap(function(R){for(;;)switch(R.prev=R.next){case 0:if(!(z&&!String(z).trim().startsWith("/"))){R.next=2;break}throw new Error("\u8DEF\u5F84\u524D\u7F00\u5FC5\u987B\u4EE5 / \u5F00\u5934");case 2:case"end":return R.stop()}},$)}));function I($,q){return M.apply(this,arguments)}return I})()}],children:(0,e.jsx)(ze.A,{placeholder:"\u53EF\u9009\uFF0C\u4F8B\u5982 /api"})}),(0,e.jsx)(d.A.Item,{name:[a.name,"rewrite"],label:"\u6539\u5199\u8BBF\u95EE\u8DEF\u5F84",extra:"\u5728\u79FB\u9664\u524D\u7F00\u540E\u6267\u884C\u3002\u53EA\u6539\u53D8\u5185\u90E8\u5904\u7406\u8DEF\u5F84\uFF0C\u4E0D\u4F1A\u8BA9\u6D4F\u89C8\u5668\u8DF3\u8F6C\u3002",children:(0,e.jsx)(ze.A,{placeholder:"\u53EF\u9009\uFF0C\u4F8B\u5982 /index.html \u6216 /v1{http.request.uri}"})}),E==="static"&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:[a.name,"index"],label:"\u9ED8\u8BA4\u9996\u9875",extra:"\u6309\u586B\u5199\u987A\u5E8F\u67E5\u627E\u9996\u9875\u6587\u4EF6\uFF0C\u53EF\u7559\u7A7A\u3002",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],placeholder:"\u4F8B\u5982 index.html\u3001index.htm"})}),(0,e.jsx)(d.A.Item,{name:[a.name,"try_files"],label:"\u6587\u4EF6\u67E5\u627E\u987A\u5E8F",extra:"\u4F8B\u5982\u5148\u67E5\u627E\u8BF7\u6C42\u6587\u4EF6\uFF0C\u627E\u4E0D\u5230\u65F6\u8FD4\u56DE /index.html\uFF0C\u9002\u5408\u524D\u7AEF\u5355\u9875\u5E94\u7528\u3002",children:(0,e.jsx)(De.A,{mode:"tags",placeholder:"\u4F8B\u5982 {http.request.uri.path}\u3001/index.html"})})]})]})}]})]})]})},ho=(0,i.forwardRef)(function(s,n){var t=s.fields,r=s.add,a=s.remove,l=s.move,h=s.styles,c=d.A.useFormInstance(),f=(0,i.useState)([]),y=C()(f,2),j=y[0],F=y[1];(0,i.useImperativeHandle)(n,function(){return{reveal:F,getValues:function(){return c.getFieldValue("routes")}}},[c]);var B=function(N){for(var w=new Set((c.getFieldValue("routes")||[]).map(function(m){return String((m==null?void 0:m.name)||"").trim()})),u=N?"".concat(N.name||"\u8DEF\u7531"," \u526F\u672C"):"\u8DEF\u7531",S=1,x=N?u:"".concat(u," ").concat(S);w.has(x);)x="".concat(u," ").concat(++S);F([]),r(N?D()(D()({},JSON.parse(JSON.stringify(N))),{},{name:x,_new:!0}):{name:x,_new:!0,methods:[],paths:["/api/*"],kind:"proxy",response_status:200})},o=(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:h.settingAction,size:"small",icon:(0,e.jsx)(ce.In,{type:"PlusOutlined"}),onClick:function(){return B()},children:"\u6DFB\u52A0\u89C4\u5219"});return(0,e.jsx)("div",{className:h.routeRules,children:t.length?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:"route-toolbar",children:[(0,e.jsx)("span",{children:"\u4ECE\u4E0A\u5230\u4E0B\u5339\u914D\uFF0C\u547D\u4E2D\u540E\u6309\u5BF9\u5E94\u65B9\u5F0F\u5904\u7406"}),o]}),t.map(function(v,N){return(0,e.jsx)(vo,{field:v,index:N,count:t.length,errors:j,move:function(u,S){F([]),l(u,S)},remove:function(u){F([]),a(u)},copy:B,styles:h},v.key)})]}):(0,e.jsxs)("div",{className:"route-empty",children:[(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:"\u6682\u65E0\u6570\u636E"}),o]})})}),fo=function(n){var t=n.site,r=n.styles,a=n.hideHeader,l=a===void 0?!1:a,h=(0,i.useRef)(null),c=(0,i.useCallback)(function(y){return{routes:(Array.isArray(y)?y:[]).map(function(j){var F,B,o;return D()(D()({},j),{},{kind:j.proxy?"proxy":j.response?"response":"static",proxy_upstreams:ii((F=j.proxy)===null||F===void 0?void 0:F.upstreams).map(function(v){var N,w,u,S=String(v);return((N=j.proxy)===null||N===void 0?void 0:N.transport)==="fastcgi"?S:/^unix(?:\+h2c)?\//.test(S)?(w=j.proxy)!==null&&w!==void 0&&(w=w.versions)!==null&&w!==void 0&&w.includes("h2c")?S.replace(/^unix(?:\+h2c)?\//,"unix+h2c/"):S.replace(/^unix\//,""):/^https?:\/\//i.test(S)||S.startsWith("/")?S:"".concat(((u=j.proxy)===null||u===void 0?void 0:u.scheme)||"http","://").concat(S)}),response_status:((B=j.response)===null||B===void 0?void 0:B.status)||200,response_body:((o=j.response)===null||o===void 0?void 0:o.body)||""})})}},[]),f=(0,i.useCallback)(function(y){var j;return(((j=h.current)===null||j===void 0?void 0:j.getValues())||y.routes||[]).map(function(F){var B=F.kind,o=F._new,v=F.proxy_upstreams,N=F.response_status,w=F.response_body,u=pt()(F,$s),S=D()(D()({},u),{},{name:String(F.name||"").trim(),paths:vn(F.paths),methods:vn(F.methods).filter(function(k){return k!=="*"}),strip_prefix:String(F.strip_prefix||"").trim(),rewrite:F.rewrite||"",root:"",index:[],browse:!1,try_files:[],hide:[],precompressed:[],proxy:null,response:null});if(B==="proxy"){var x=u.proxy||{},m=x.transport==="fastcgi",p=vn(v),_=m?x.scheme:Nr(p[0]).scheme,A=p.map(function(k){return m?{dial:k.startsWith("/")?"unix/".concat(k):k,scheme:x.scheme||"http",h2c:!1}:Nr(k,_)});S.proxy=D()(D()({},x),{},{upstreams:A.map(function(k){var T=k.dial,U=(x.upstreams||[]).find(function(P){if(m)return P.dial===T||P.dial===T.replace(/^unix\//,"");try{return Nr(P.dial,x.scheme||"http").dial===T}catch(M){return P.dial===T}});return D()(D()({max_requests:0},U),{},{dial:(U==null?void 0:U.dial)||T})}),transport:x.transport||"http",scheme:_,policy:x.policy||"random"},m?{}:{versions:A.some(function(k){return k.h2c})?["h2c"]:(x.versions||[]).filter(function(k){var T;return(k!=="h2c"||_==="http"&&!((T=A[0])!==null&&T!==void 0&&T.dial.startsWith("unix/")))&&(k!=="3"||_==="https")}),tls_server_name:_==="https"&&x.tls_server_name||"",tls_insecure_skip_verify:_==="https"?x.tls_insecure_skip_verify:!1})}else if(B==="response"){var E;S.response=D()(D()({},u.response),{},{status:Number(N||200),body:w||"",headers:((E=u.response)===null||E===void 0?void 0:E.headers)||{}})}else S.root=String(F.root||"").trim(),S.index=vn(F.index),S.browse=!!F.browse,S.try_files=vn(F.try_files),S.hide=vn(F.hide),S.precompressed=vn(F.precompressed);return S})},[]);return(0,e.jsx)(qr,{site:t,request:Li,update:Li,title:"\u81EA\u5B9A\u4E49\u8DEF\u7531",description:"\u8BA9\u4E0D\u540C\u8BBF\u95EE\u8DEF\u5F84\u8F6C\u53D1\u5230\u670D\u52A1\u3001\u8BFB\u53D6\u6587\u4EF6\u6216\u8FD4\u56DE\u6307\u5B9A\u5185\u5BB9",styles:r,hideHeader:l,normalize:c,serialize:f,onFinishFailed:function(j){var F,B=j.errorFields;return(F=h.current)===null||F===void 0?void 0:F.reveal(B)},children:(0,e.jsx)(d.A.List,{name:"routes",children:function(j,F){var B=F.add,o=F.remove,v=F.move;return(0,e.jsx)(ho,{ref:h,fields:j,add:B,remove:o,move:v,styles:r})}})})},go=function(n){var t=n.site,r=n.styles,a=(0,i.useRef)({}),l=[{type:"access",name:"\u8BBF\u95EE\u65E5\u5FD7",icon:"FileTextOutlined",hint:"\u8BBF\u95EE\u8BF7\u6C42\u3001\u54CD\u5E94\u72B6\u6001\u4E0E\u8017\u65F6"},{type:"waf",name:"\u9632\u706B\u5899\u65E5\u5FD7",icon:"SafetyCertificateOutlined",hint:"\u89C4\u5219\u547D\u4E2D\u4E0E\u62E6\u622A\u7ED3\u679C"}];return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(Rt,{title:"\u7F51\u7AD9\u65E5\u5FD7",description:"\u9009\u62E9\u65E5\u5FD7\u67E5\u770B\u5185\u5BB9\uFF0C\u53EF\u5728\u65E5\u5FD7\u7A97\u53E3\u4E2D\u5237\u65B0\u6216\u6E05\u7406",styles:r}),(0,e.jsx)("div",{className:r.logGrid,children:l.map(function(h){return(0,e.jsxs)("div",{className:r.logItem,children:[(0,e.jsx)("span",{className:"log-icon","aria-hidden":"true",children:(0,e.jsx)(ce.In,{type:h.icon})}),(0,e.jsxs)("div",{className:"log-main",children:[(0,e.jsx)("div",{className:"log-name",children:h.name}),(0,e.jsx)("div",{className:"log-description",children:h.hint})]}),(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},className:"".concat(r.settingAction," log-open"),size:"small","aria-label":"\u67E5\u770B".concat(h.name),onClick:function(){var f;return(f=a.current)===null||f===void 0?void 0:f.open(t,{type:h.type})},children:"\u67E5\u770B"})]},h.type)})}),(0,e.jsx)(sa.A,{ref:a,request:Ri,title:"\u7F51\u7AD9\u65E5\u5FD7"})]})},Ju=(0,i.forwardRef)(function(s,n){var t,r=s.onSuccess,a=s.onManageCert,l=s.onObtainCert,h=s.onDeployCert,c=s.onRenewCert,f=s.certVersion,y=f===void 0?0:f,j=s.typeItems,F=(0,ce.DP)(),B=!!(!(F==null||(t=F.isCompactTheme)===null||t===void 0)&&t.call(F)),o=Sn.Ay.useBreakpoint(),v=qs({compact:B}),N=v.styles,w=(0,i.useRef)({}),u=(0,i.useRef)(0),S=(0,i.useRef)(null),x=(0,i.useRef)(null),m=(0,i.useRef)({}),p=(0,i.useState)(),_=C()(p,2),A=_[0],E=_[1],k=(0,i.useState)("basic"),T=C()(k,2),U=T[0],P=T[1],M="site-settings-".concat(String((A==null?void 0:A.id)||"none"),"-").concat(U),I=(0,i.useRef)(M);I.current=M;var $=(0,i.useCallback)(function(g,L){var J,V;g===I.current&&((J=m.current)===null||J===void 0||(V=J.update)===null||V===void 0||V.call(J,L))},[]),q=(0,i.useMemo)(function(){return{formId:M,update:$}},[M,$]),z=(0,i.useCallback)(function(g){var L,J=arguments.length>1&&arguments[1]!==void 0?arguments[1]:"basic";u.current+=1;var V={common:"basic",safety:"waf",performance:"cache",advanced:"header"},ee=V[J]||J,fe=["basic","domain","ssl","runtime","redirect","waf","rate","traffic","cache","compression","log","header","route"].includes(ee)?ee:"basic";E(g),P(fe),(L=w.current)===null||L===void 0||L.show()},[]);(0,i.useImperativeHandle)(n,function(){return{open:z}},[z]);var he=(0,i.useMemo)(function(){if(!A)return[];var g={0:"\u9759\u6001\u8BBE\u7F6E",1:"\u53CD\u5411\u4EE3\u7406",2:"\u9879\u76EE\u8BBE\u7F6E",3:"PHP \u8BBE\u7F6E"};return[{key:"basic",label:"\u57FA\u7840\u4FE1\u606F",color:"#1677ff"},{key:"domain",label:"\u57DF\u540D\u7BA1\u7406",color:"#13a8a8"},{key:"ssl",label:"HTTPS",color:"#52a832"},{key:"runtime",label:g[Number(A.type)]||"\u8FD0\u884C\u8BBE\u7F6E",color:"#722ed1"},{key:"redirect",label:"\u91CD\u5B9A\u5411",color:"#389e0d"},{key:"waf",label:"Web \u9632\u706B\u5899",color:"#d4380d"},{key:"rate",label:"\u8BBF\u95EE\u9891\u7387",color:"#d46b08"},{key:"traffic",label:"\u5E76\u53D1\u9650\u901F",color:"#d4a106"},{key:"cache",label:"\u54CD\u5E94\u7F13\u5B58",color:"#2f54c9"},{key:"compression",label:"\u54CD\u5E94\u538B\u7F29",color:"#08979c"},{key:"log",label:"\u7F51\u7AD9\u65E5\u5FD7",color:"#597ef7"},{key:"header",label:"\u8BF7\u6C42\u4E0E\u54CD\u5E94\u5934",color:"#c41d7f"},{key:"route",label:"\u81EA\u5B9A\u4E49\u8DEF\u7531",color:"#531dab"}]},[A]);(0,i.useEffect)(function(){var g=S.current;if(g){var L=0,J=function(){cancelAnimationFrame(L),L=requestAnimationFrame(function(){var Y=g.querySelector('[data-section="'.concat(U,'"]'));Y==null||Y.scrollIntoView({behavior:"auto",block:"nearest",inline:"nearest"})})},V=function(Y){var de=Math.max(0,g.scrollWidth-g.clientWidth);if(!(de<=0)){var xe=Math.abs(Y.deltaX)>Math.abs(Y.deltaY)?Y.deltaX:Y.deltaY,ye=xe<0?g.scrollLeft>0:g.scrollLeft<de;ye&&(Y.preventDefault(),g.scrollLeft+=xe)}},ee=new ResizeObserver(J);return g.addEventListener("wheel",V,{passive:!1}),ee.observe(g),J(),function(){cancelAnimationFrame(L),ee.disconnect(),g.removeEventListener("wheel",V)}}},[U,A]),(0,i.useEffect)(function(){var g;(g=x.current)===null||g===void 0||g.scrollTo({top:0,behavior:"auto"})},[U,A==null?void 0:A.id]);var R=(0,i.useMemo)(function(){if(!A)return null;var g={site:A,styles:N,typeItems:j};switch(U){case"basic":return(0,e.jsx)(Xs,D()(D()({},g),{},{onChange:function(J){J&&E(J),r==null||r()}}));case"domain":return(0,e.jsx)(Js,D()(D()({},g),{},{onChange:r}));case"ssl":return(0,e.jsx)(Zs,D()(D()({},g),{},{settingsSessionRef:u,certVersion:y,onManageCert:a,onObtainCert:l,onDeployCert:h,onRenewCert:c}),A.id);case"redirect":return(0,e.jsx)(mo,D()({},g));case"runtime":return Number(A.type)===0?(0,e.jsx)(eo,D()({},g)):Number(A.type)===3?(0,e.jsx)(ro,D()({},g)):Number(A.type)===2?(0,e.jsx)(to,D()({},g)):(0,e.jsx)(no,D()({},g));case"waf":return(0,e.jsx)(ao,D()({},g));case"rate":return(0,e.jsx)(io,D()({},g));case"traffic":return(0,e.jsx)(uo,D()({},g));case"cache":return(0,e.jsx)(lo,D()({},g));case"compression":return(0,e.jsx)(so,D()({},g));case"log":return(0,e.jsx)(go,D()({},g));case"header":return(0,e.jsx)(oo,D()({},g));case"route":return(0,e.jsx)(fo,D()({},g));default:return null}},[U,y,a,l,h,c,r,A,N,j]);return(0,e.jsx)(ce.Ht,{ref:w,title:(0,e.jsx)(ce.hE,{children:"\u7F51\u7AD9\u8BBE\u7F6E"}),width:B?840:800,afterOpenChange:function(L){L||(u.current+=1)},modalProps:{rootClassName:N.modal,footer:U==="log"?null:(0,e.jsx)(Xu,{ref:m,formId:M,onClose:function(){var L;return(L=w.current)===null||L===void 0?void 0:L.hide()}},M),forceRender:!0,closable:!0,keyboard:!0,style:{top:o.md?100:24,paddingBottom:o.md?100:24},mask:{closable:!0},afterClose:function(){E(void 0),P("basic")}},children:(0,e.jsx)(Vu.Provider,{value:q,children:(0,e.jsxs)("div",{className:N.layout,children:[(0,e.jsx)("aside",{className:N.navigation,"aria-label":"\u7F51\u7AD9\u8BBE\u7F6E\u5206\u7C7B",children:(0,e.jsx)("div",{className:"navigation-scroll",ref:S,children:(0,e.jsx)("div",{className:"navigation-grid",children:he.map(function(g){return(0,e.jsxs)("button",{type:"button",className:"navigation-item","data-section":g.key,"data-active":U===g.key,"aria-current":U===g.key?"page":void 0,style:{"--accent":g.color},onClick:function(){U!==g.key&&P(g.key)},children:[(0,e.jsx)("span",{className:"navigation-dot"}),(0,e.jsx)("span",{className:"navigation-label",children:g.label})]},g.key)})})})}),(0,e.jsx)("main",{ref:x,className:N.content,children:(0,e.jsx)("div",{className:N.contentInner,children:R})})]})})})});Ju.displayName="Settings";var xo=(0,i.memo)(Ju),bo=X(25542),yo=X.n(bo),Fo=X(68328),Yu=X.n(Fo),wo=X(97087),Gt=X(24816),it=X(72523);function Ao(){return li.apply(this,arguments)}function li(){return li=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.Jt)("/secret/list",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),li.apply(this,arguments)}function Eo(){return si.apply(this,arguments)}function si(){return si=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.Jt)("/secret/info",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),si.apply(this,arguments)}function Co(){return oi.apply(this,arguments)}function oi(){return oi=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/secret/create",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),oi.apply(this,arguments)}function jo(){return di.apply(this,arguments)}function di(){return di=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/secret/update",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),di.apply(this,arguments)}function So(){return ci.apply(this,arguments)}function ci(){return ci=ne()(b()().mark(function s(){var n,t=arguments;return b()().wrap(function(a){for(;;)switch(a.prev=a.next){case 0:return n=t.length>0&&t[0]!==void 0?t[0]:{},a.abrupt("return",(0,dn.bE)("/secret/delete",n==null?void 0:n.body,n==null?void 0:n.onSuccess,n==null?void 0:n.onFail,n==null?void 0:n.onFinally));case 2:case"end":return a.stop()}},s)})),ci.apply(this,arguments)}var Do=X(63859),pi=X.n(Do),ko=X(82591),Zu,el,Bo=(0,kr.rU)(function(s){var n=s.css;return{modal:n(Zu||(Zu=H()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-body {
            max-height: calc(100dvh - 290px);
            overflow-y: auto;
            overscroll-behavior: contain;
        }

        @supports not (height: 100dvh) {
            .ant-modal-body {
                max-height: calc(100vh - 290px);
            }
        }
        @media (max-width: 767px) {
            .ant-modal-body {
                max-height: calc(100dvh - 138px);
            }

            @supports not (height: 100dvh) {
                .ant-modal-body {
                    max-height: calc(100vh - 138px);
                }
            }
        }
    `]))),loading:n(el||(el=H()([`
        min-height: 240px;
        display: flex;
        align-items: center;
        justify-content: center;
    `])))}}),nl=(0,i.forwardRef)(function(s,n){var t,r=s.zIndex,a=s.onSuccess,l=un.A.useApp(),h=l.message,c=(0,ce.DP)(),f=!!(!(c==null||(t=c.isCompactTheme)===null||t===void 0)&&t.call(c)),y=Sn.Ay.useBreakpoint(),j=Bo(),F=j.styles,B=(0,i.useRef)(null),o=(0,i.useRef)(0),v=(0,i.useRef)(!0),N=(0,i.useRef)(!1),w=(0,i.useRef)(!1),u=(0,i.useRef)({}),S=d.A.useForm(),x=C()(S,1),m=x[0],p=(0,i.useState)(),_=C()(p,2),A=_[0],E=_[1],k=(0,i.useState)({}),T=C()(k,2),U=T[0],P=T[1],M=(0,i.useState)([]),I=C()(M,2),$=I[0],q=I[1],z=(0,i.useState)(!1),he=C()(z,2),R=he[0],g=he[1],L=(0,i.useState)(!1),J=C()(L,2),V=J[0],ee=J[1],fe=(0,i.useState)(""),Y=C()(fe,2),de=Y[0],xe=Y[1],ye=(0,i.useState)(!1),Ce=C()(ye,2),ie=Ce[0],Be=Ce[1],Z=A!==void 0,cn=(0,i.useMemo)(function(){return Object.entries(U).map(function(Ue){var tn=C()(Ue,2),pe=tn[0],Se=tn[1];return{value:Number(pe),label:Se}})},[U]),Le=(0,i.useCallback)(function(){var Ue;o.current+=1,N.current=!1,w.current=!1,u.current={},(Ue=B.current)===null||Ue===void 0||Ue.hide()},[]),Fe=(0,i.useCallback)(function(Ue){var tn,pe=++o.current;N.current=!0,w.current=!1,u.current={},m.resetFields(),E(Ue),P({}),q([]),g(!0),ee(!0),xe(""),Be(!1),(tn=B.current)===null||tn===void 0||tn.show(),ne()(b()().mark(function Se(){var Ee,we,ue,K,ve,me,le,Ae,re,te;return b()().wrap(function(se){for(;;)switch(se.prev=se.next){case 0:return se.prev=0,se.next=3,Promise.all([(0,it.GV)({body:{name:"secret"}}),Ue?Eo({body:{id:Ue.id}}):Promise.resolve(void 0)]);case 3:if(ue=se.sent,K=C()(ue,2),ve=K[0],me=K[1],!(o.current!==pe||!N.current)){se.next=9;break}return se.abrupt("return");case 9:if(!((ve==null?void 0:ve.code)!==200||!(ve!=null&&(Ee=ve.data)!==null&&Ee!==void 0&&Ee.provider))){se.next=11;break}throw new Error((ve==null?void 0:ve.message)||"\u83B7\u53D6\u5BC6\u94A5\u914D\u7F6E\u5931\u8D25");case 11:if(!(Ue&&((me==null?void 0:me.code)!==200||!(me!=null&&me.data)))){se.next=13;break}throw new Error((me==null?void 0:me.message)||"\u83B7\u53D6\u5BC6\u94A5\u8BE6\u60C5\u5931\u8D25");case 13:if(le=ve==null?void 0:ve.data,Ae=me==null?void 0:me.data,re=Number(Ae?Ae==null?void 0:Ae.provider:Object.keys((le==null?void 0:le.provider)||{})[0]),te=Ae?Ae==null?void 0:Ae.data:{},!(!(le!=null&&(we=le.provider)!==null&&we!==void 0&&we[String(re)])||!te||pi()(te)!=="object"||Array.isArray(te)||Ae&&Object.keys(te).length===0||Object.values(te).some(function(Bn){return typeof Bn!="string"}))){se.next=19;break}throw new Error("\u5BC6\u94A5\u914D\u7F6E\u4E0D\u53EF\u7528\uFF0C\u8BF7\u91CD\u65B0\u52A0\u8F7D");case 19:u.current=Ae?D()({},te):{},P((le==null?void 0:le.provider)||{}),q(Object.keys(te)),m.setFieldsValue({name:(Ae==null?void 0:Ae.name)||"",provider:re,credentials:D()({},te)}),se.next=28;break;case 25:se.prev=25,se.t0=se.catch(0),o.current===pe&&N.current&&xe(se.t0 instanceof Error?se.t0.message:"\u83B7\u53D6\u5BC6\u94A5\u914D\u7F6E\u5931\u8D25");case 28:return se.prev=28,o.current===pe&&ee(!1),se.finish(28);case 31:case"end":return se.stop()}},Se,null,[[0,25,28,31]])}))()},[m]);(0,i.useImperativeHandle)(n,function(){return{open:Fe,close:Le}},[Le,Fe]),(0,i.useEffect)(function(){return v.current=!0,function(){v.current=!1,o.current+=1,N.current=!1,u.current={}}},[]);var nn=(0,i.useCallback)((function(){var Ue=ne()(b()().mark(function tn(pe){var Se,Ee,we,ue,K,ve,me,le;return b()().wrap(function(re){for(;;)switch(re.prev=re.next){case 0:if(!(!N.current||w.current||V||de||Z&&$.length===0)){re.next=2;break}return re.abrupt("return");case 2:if(Se=Number(pe.provider),!(!Z&&!U[String(Se)])){re.next=6;break}return h.error("\u8BF7\u9009\u62E9\u5BC6\u94A5\u5382\u5546"),re.abrupt("return");case 6:return Ee=++o.current,we={},Z&&$.forEach(function(te){var mn,se=String(((mn=pe.credentials)===null||mn===void 0?void 0:mn[te])||"").trim();se!==u.current[te]&&(we[te]=se)}),ue=D()(D()(D()({name:pe.name.trim()},Z?{}:{provider:Se}),Object.keys(we).length>0?{data:JSON.stringify(we)}:{}),A?{id:A.id}:{}),w.current=!0,Be(!0),re.prev=12,re.next=15,(Z?jo:Co)({body:ue});case 15:if(me=re.sent,v.current){re.next=18;break}return re.abrupt("return");case 18:if((me==null?void 0:me.code)===200){re.next=21;break}return o.current===Ee&&h.error((me==null?void 0:me.message)||"\u4FDD\u5B58\u5BC6\u94A5\u5931\u8D25"),re.abrupt("return");case 21:o.current===Ee&&(h.success((me==null?void 0:me.message)||"\u4FDD\u5B58\u6210\u529F"),A?Le():Fe({id:Number(me==null||(le=me.data)===null||le===void 0?void 0:le.id)})),a({action:Z?"update":"create",id:(K=A==null?void 0:A.id)!==null&&K!==void 0?K:Number(me==null||(ve=me.data)===null||ve===void 0?void 0:ve.id)}),re.next=28;break;case 25:re.prev=25,re.t0=re.catch(12),o.current===Ee&&N.current&&h.error("\u4FDD\u5B58\u5BC6\u94A5\u5931\u8D25");case 28:return re.prev=28,o.current===Ee&&(w.current=!1,Be(!1)),re.finish(28);case 31:case"end":return re.stop()}},tn,null,[[12,25,28,31]])}));return function(tn){return Ue.apply(this,arguments)}})(),[Le,Z,$,de,V,h,a,Fe,U,A]);return(0,e.jsx)(ce.Ht,{ref:B,title:(0,e.jsx)(ce.hE,{children:Z?"\u7F16\u8F91\u5BC6\u94A5":"\u6DFB\u52A0\u5BC6\u94A5"}),width:Z?f?380:400:350,onOk:m.submit,onCancel:Le,modalProps:{rootClassName:F.modal,zIndex:r,forceRender:!0,confirmLoading:ie,okButtonProps:{disabled:V||!!de},okText:Z?"\u4FDD\u5B58":"\u6DFB\u52A0",cancelText:"\u53D6\u6D88",footer:V||de?null:void 0,mask:{closable:!0},style:{top:y.md?100:24,paddingBottom:y.md?100:24},focusable:{focusTriggerAfterClose:!1},afterClose:function(){N.current||(m.resetFields(),E(void 0),P({}),q([]),g(!1),ee(!1),xe(""),Be(!1))}},children:(0,e.jsx)(d.A,{form:m,layout:"vertical",variant:"filled",onFinish:nn,children:R&&(V?(0,e.jsx)("div",{className:F.loading,children:(0,e.jsx)(mt.A,{})}):de?(0,e.jsx)("div",{className:F.loading,children:(0,e.jsxs)(ko.A,{orientation:"vertical",align:"center",children:[(0,e.jsx)(Gt.A.Text,{type:"danger",children:de}),(0,e.jsx)(ge.A,{color:"default",variant:"filled",onClick:function(){return Fe(A)},children:"\u91CD\u65B0\u52A0\u8F7D"})]})}):(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"name",label:"\u5BC6\u94A5\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u5BC6\u94A5\u540D\u79F0"},{max:100,message:"\u5BC6\u94A5\u540D\u79F0\u4E0D\u80FD\u8D85\u8FC7 100 \u4E2A\u5B57\u7B26"}],children:(0,e.jsx)(ze.A,{maxLength:100,placeholder:"\u8BF7\u8F93\u5165\u5BC6\u94A5\u540D\u79F0",autoComplete:"off"})}),!Z&&(0,e.jsx)(d.A.Item,{name:"provider",label:"\u5BC6\u94A5\u5382\u5546",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u5BC6\u94A5\u5382\u5546"}],children:(0,e.jsx)(De.A,{options:cn,placeholder:"\u8BF7\u9009\u62E9\u5BC6\u94A5\u5382\u5546",disabled:ie})}),Z&&$.map(function(Ue){return(0,e.jsx)(d.A.Item,{preserve:!1,name:["credentials",Ue],label:Ue,rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165 ".concat(Ue)}],children:(0,e.jsx)(ze.A,{placeholder:"\u8BF7\u8F93\u5165 ".concat(Ue),autoComplete:"off"})},Ue)})]}))})})});nl.displayName="SecretForm";var No=(0,i.memo)(nl),rl,To=(0,kr.rU)(function(s){var n=s.css,t=s.token;return{table:n(rl||(rl=H()([`
        .secret-name,
        .secret-provider,
        .secret-time {
            display: block;
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            line-height: 20px;
        }
        .secret-name {
            padding: 0;
            border: 0;
            background: transparent;
            color: inherit;
            font: inherit;
            text-align: left;
            cursor: pointer;
        }
        .secret-name:hover,
        .secret-name:focus-visible { color: `,`; }
        .secret-time { color: `,`; font-variant-numeric: tabular-nums; }
    `])),t.colorPrimary,t.colorTextSecondary)}}),tl=(0,i.forwardRef)(function(s,n){var t,r=s.zIndex,a=r===void 0?1050:r,l=s.onChange,h=un.A.useApp(),c=h.message,f=h.modal,y=(0,ce.DP)(),j=!!(!(y==null||(t=y.isCompactTheme)===null||t===void 0)&&t.call(y)),F=Sn.Ay.useBreakpoint(),B=To(),o=B.styles,v=(0,i.useRef)(null),N=(0,i.useRef)(null),w=(0,i.useRef)(null),u=(0,i.useRef)(!0),S=(0,i.useRef)(!1),x=(0,i.useRef)(0),m=(0,i.useRef)(0),p=(0,i.useRef)(l);p.current=l;var _=(0,i.useState)(!1),A=C()(_,2),E=A[0],k=A[1],T=(0,i.useState)(0),U=C()(T,2),P=U[0],M=U[1],I=(0,i.useState)(""),$=C()(I,2),q=$[0],z=$[1],he=(0,i.useState)({}),R=C()(he,2),g=R[0],L=R[1],J=(0,i.useState)(""),V=C()(J,2),ee=V[0],fe=V[1],Y=(0,i.useState)([]),de=C()(Y,2),xe=de[0],ye=de[1],Ce=(0,i.useMemo)(function(){return[{value:"",label:"\u5168\u90E8\u5382\u5546"}].concat(Je()(Object.entries(g).map(function(pe){var Se=C()(pe,2),Ee=Se[0],we=Se[1];return{value:Ee,label:we}})))},[g]),ie=(0,i.useCallback)(ne()(b()().mark(function pe(){var Se,Ee,we;return b()().wrap(function(K){for(;;)switch(K.prev=K.next){case 0:return Se=++m.current,K.prev=1,K.next=4,(0,it.GV)({body:{name:"secret"}});case 4:if(we=K.sent,!(!u.current||m.current!==Se)){K.next=7;break}return K.abrupt("return");case 7:if((we==null?void 0:we.code)===200){K.next=10;break}return c.error((we==null?void 0:we.message)||"\u83B7\u53D6\u5BC6\u94A5\u5382\u5546\u5931\u8D25"),K.abrupt("return");case 10:L((we==null||(Ee=we.data)===null||Ee===void 0?void 0:Ee.provider)||{}),K.next=16;break;case 13:K.prev=13,K.t0=K.catch(1),u.current&&m.current===Se&&c.error("\u83B7\u53D6\u5BC6\u94A5\u5382\u5546\u5931\u8D25");case 16:case"end":return K.stop()}},pe,null,[[1,13]])})),[c]),Be=(0,i.useCallback)(function(){var pe,Se;x.current+=1,m.current+=1,S.current=!1,(pe=N.current)===null||pe===void 0||pe.close(),(Se=w.current)===null||Se===void 0||Se.destroy(),w.current=null,k(!1),z(""),L({}),fe("")},[]),Z=(0,i.useCallback)(function(){var pe,Se;(pe=N.current)===null||pe===void 0||pe.close(),(Se=w.current)===null||Se===void 0||Se.destroy(),w.current=null,x.current+=1,S.current=!0,M(x.current),z(""),L({}),fe(""),k(!0),ie()},[ie]),cn=(0,i.useCallback)(function(){if(S.current){var pe;(pe=v.current)===null||pe===void 0||pe.reload(),ie()}},[ie]),Le=(0,i.useCallback)(function(){var pe;(pe=N.current)===null||pe===void 0||pe.open()},[]);(0,i.useImperativeHandle)(n,function(){return{open:Z,close:Be,refresh:cn,openCreate:Le}},[Be,Le,cn,Z]),(0,i.useEffect)(function(){return u.current=!0,function(){var pe;u.current=!1,S.current=!1,x.current+=1,m.current+=1,(pe=w.current)===null||pe===void 0||pe.destroy()}},[]);var Fe=(0,i.useCallback)((function(){var pe=ne()(b()().mark(function Se(Ee){var we,ue,K,ve,me,le,Ae,re,te,mn;return b()().wrap(function(Bn){for(;;)switch(Bn.prev=Bn.next){case 0:return ve=Ee.current,me=Ee.pageSize,le=Ee.keyword,Ae=Ee.provider,Bn.next=3,Ao({body:{page:ve,page_size:me,keyword:typeof le=="string"?le.trim():"",provider:Ae,order_by_field:"id",order_by_type:"desc"}});case 3:if(re=Bn.sent,te=(we=re==null||(ue=re.data)===null||ue===void 0?void 0:ue.list)!==null&&we!==void 0?we:[],!((re==null?void 0:re.code)!==200||!Array.isArray(te))){Bn.next=7;break}throw new Error((re==null?void 0:re.message)||"\u83B7\u53D6\u5BC6\u94A5\u5217\u8868\u5931\u8D25");case 7:return mn=te.filter(function(In){return In==null?void 0:In.id}).map(function(In){return{id:Number(In.id),name:In.name,provider:Number(In.provider),create_time:In.create_time,update_time:In.update_time}}),Bn.abrupt("return",{data:mn,total:Number((re==null||(K=re.data)===null||K===void 0?void 0:K.total)||0),success:!0});case 9:case"end":return Bn.stop()}},Se)}));return function(Se){return pe.apply(this,arguments)}})(),[]),nn=(0,i.useCallback)(function(pe){if(u.current){if(S.current){var Se;(Se=v.current)===null||Se===void 0||Se.reload()}try{var Ee;Promise.resolve((Ee=p.current)===null||Ee===void 0?void 0:Ee.call(p,pe)).catch(function(){})}catch(we){}}},[]),Ue=(0,i.useCallback)(function(pe){var Se,Ee=x.current;(Se=w.current)===null||Se===void 0||Se.destroy();var we=f.confirm({title:"\u5220\u9664\u5BC6\u94A5",content:"\u786E\u5B9A\u5220\u9664\u5BC6\u94A5\u201C".concat(pe.name,"\u201D\u5417\uFF1F\u5C06\u89E3\u9664\u8BC1\u4E66\u5173\u8054\u5E76\u5173\u95ED\u81EA\u52A8\u7EED\u7B7E\u3002"),zIndex:a+30,okText:"\u5220\u9664",cancelText:"\u53D6\u6D88",okButtonProps:{danger:!0},mask:{closable:!0},onOk:(function(){var ue=ne()(b()().mark(function ve(){var me;return b()().wrap(function(Ae){for(;;)switch(Ae.prev=Ae.next){case 0:return ye(function(re){return[].concat(Je()(re),[pe.id])}),Ae.prev=1,Ae.next=4,So({body:{id:pe.id}});case 4:if(me=Ae.sent,(me==null?void 0:me.code)===200){Ae.next=7;break}throw new Error((me==null?void 0:me.message)||"\u5220\u9664\u5BC6\u94A5\u5931\u8D25");case 7:u.current&&x.current===Ee&&c.success((me==null?void 0:me.message)||"\u5220\u9664\u6210\u529F"),nn({action:"delete",id:pe.id}),Ae.next=15;break;case 11:throw Ae.prev=11,Ae.t0=Ae.catch(1),u.current&&x.current===Ee&&c.error(Ae.t0 instanceof Error?Ae.t0.message:"\u5220\u9664\u5BC6\u94A5\u5931\u8D25"),Ae.t0;case 15:return Ae.prev=15,u.current&&ye(function(re){return re.filter(function(te){return te!==pe.id})}),Ae.finish(15);case 18:case"end":return Ae.stop()}},ve,null,[[1,11,15,18]])}));function K(){return ue.apply(this,arguments)}return K})(),afterClose:function(){w.current===we&&(w.current=null)}});w.current=we,Promise.resolve(we)},[nn,c,f,a]),tn=(0,i.useMemo)(function(){return[{title:"\u5BC6\u94A5\u540D\u79F0",dataIndex:"name",key:"name",width:190,render:function(Se,Ee){return(0,e.jsx)("button",{type:"button",className:"secret-name",title:Se,disabled:xe.includes(Ee.id),onClick:function(){var ue;return(ue=N.current)===null||ue===void 0?void 0:ue.open(Ee)},children:Se})}},{title:"\u5382\u5546",dataIndex:"provider",key:"provider",width:125,render:function(Se){return(0,e.jsx)("span",{className:"secret-provider",children:g[Se]||"\u672A\u77E5\u5382\u5546"})}},{title:"\u66F4\u65B0\u65F6\u95F4",dataIndex:"update_time",key:"update_time",width:160,render:function(Se){return(0,e.jsx)(Kn.A,{title:Se,children:(0,e.jsx)("span",{className:"secret-time",children:Se?String(Se).slice(0,16):"\u2014"})})}},{title:"\u64CD\u4F5C",key:"action",width:82,fixed:"right",className:"action-cell",render:function(Se,Ee){return(0,e.jsx)(gn.A,{disabled:xe.includes(Ee.id),menu:{items:[{key:"edit",label:"\u7F16\u8F91\u5BC6\u94A5"},{key:"delete",label:"\u5220\u9664\u5BC6\u94A5",danger:!0}],onClick:function(ue){var K,ve=ue.key;return ve==="edit"?(K=N.current)===null||K===void 0?void 0:K.open(Ee):Ue(Ee)}},children:(0,e.jsx)(ge.A,{size:"small",autoInsertSpace:!1,disabled:xe.includes(Ee.id),"aria-label":"\u64CD\u4F5C\u5BC6\u94A5 ".concat(Ee.name),children:"\u64CD\u4F5C"})})}}]},[Ue,xe,g]);return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(ce.oK,{ref:v,open:E,onCancel:Be,title:(0,e.jsx)(ce.hE,{children:"\u5BC6\u94A5\u7BA1\u7406"}),width:j?700:740,minHeight:j?440:480,className:o.table,scrollMode:"modal",columns:tn,rowKey:"id",scroll:{x:557},request:Fe,requestEnabled:E,params:{provider:q},defaultPageSize:10,onLoad:function(){return fe("")},onRequestError:function(Se){fe(Se.message),c.error(Se.message)},emptyContent:(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:ee||"\u6682\u65E0\u6570\u636E"}),toolbar:{search:{placeholder:"\u641C\u7D22\u5BC6\u94A5\u540D\u79F0",ariaLabel:"\u641C\u7D22\u5BC6\u94A5",maxLength:100,debounce:300},refresh:!0,actions:[{key:"provider",type:"filter",value:q,options:Ce,icon:"CloudOutlined",ariaLabel:"\u7B5B\u9009\u5BC6\u94A5\u5382\u5546",onChange:z},{key:"create",type:"button",label:"\u6DFB\u52A0",icon:"PlusOutlined",onClick:Le}]},pagination:{unit:"\u6761"},modalProps:{zIndex:a,destroyOnHidden:!0,style:{top:F.md?100:24,paddingBottom:F.md?100:24,maxWidth:"calc(100vw - 24px)"}}},P),(0,e.jsx)(No,{ref:N,zIndex:a+20,onSuccess:nn})]})});tl.displayName="SecretManager";var mi=(0,i.memo)(tl),da=X(55270),_o=X(52363),al,il,ul,ll,sl,Ro=(0,kr.rU)(function(s){var n=s.css,t=s.token,r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};return{modal:n(al||(al=H()([`
        .ant-modal { max-width: calc(100vw - 24px); outline: none; }
        .ant-modal-body { min-width: 0; padding: 0 !important; overflow: hidden; }
    `]))),content:n(il||(il=H()([`
        box-sizing: border-box;
        width: 100%;
        min-width: 0;
        max-height: calc(100dvh - 290px);
        padding: 4px;
        overflow-x: hidden;
        overflow-y: auto;
        scrollbar-gutter: stable;
        overscroll-behavior: contain;
        @supports not (max-height: 100dvh) { max-height: calc(100vh - 290px); }
        @media (max-width: 767px) {
            max-height: calc(100dvh - 164px);
            @supports not (max-height: 100dvh) { max-height: calc(100vh - 164px); }
        }
    `]))),summary:n(ul||(ul=H()([`
        margin-bottom: `,`px;
        .verification-tools { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; }
        .verification-expiry { color: `,"; font-size: ",`px; white-space: nowrap; }
        .verification-actions { display: flex; align-items: center; gap: 6px; }
        .verification-actions .ant-btn { height: 22px; padding-inline: 6px; font-size: `,`px; }
        .verification-error { margin-top: 10px; color: `,"; font-size: ",`px; line-height: 1.6; overflow-wrap: anywhere; }
    `])),r!=null&&r.compact?14:18,t.colorTextTertiary,t.fontSizeSM,t.fontSizeSM,t.colorError,t.fontSizeSM),record:n(ll||(ll=H()([`
        min-width: 0;
        & + & { margin-top: `,"px; padding-top: ","px; border-top: 1px solid ",`; }
        .verification-heading { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: `,`px; }
        .verification-heading > span { min-width: 0; color: `,"; font-size: ",`px; }
        .verification-heading > .ant-btn { flex: none; height: 22px; padding-inline: 6px; gap: 4px; font-size: `,`px; }
        .ant-form-item { min-width: 0; margin-bottom: `,`px; }
        .ant-form-item:last-child { margin-bottom: 0; }
        .ant-form-item-label { padding-bottom: `,`px; }
        .verification-type { margin-inline-start: 6px; color: `,"; font-size: ",`px; font-weight: 400; }
        .verification-field { position: relative; min-width: 0; }
        textarea.ant-input { max-width: 100%; padding-inline-end: 32px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: `,`px; line-height: 1.65; overflow-wrap: anywhere; resize: none; }
        .verification-copy { position: absolute; top: `,"px; inset-inline-end: ","px; font-size: ",`px; line-height: 22px; }
        .verification-copy .ant-typography-copy { display: inline-flex; align-items: center; justify-content: center; width: 22px; height: 22px; margin-inline-start: 0; }
    `])),r!=null&&r.compact?18:22,r!=null&&r.compact?18:22,t.colorBorderSecondary,r!=null&&r.compact?12:16,t.colorText,t.fontSize,t.fontSizeSM,r!=null&&r.compact?12:16,r!=null&&r.compact?4:6,t.colorTextSecondary,t.fontSizeSM,t.fontSizeSM,r!=null&&r.compact?2:4,r!=null&&r.compact?4:6,t.fontSizeSM),hint:n(sl||(sl=H()([`
        color: `,`;
        font-size: `,`px;
        line-height: 1.6;
    `])),t.colorTextSecondary,t.fontSizeSM)}}),ol=(0,i.forwardRef)(function(s,n){var t,r,a=s.session,l=s.busy,h=s.onVerify,c=s.onRefresh,f=s.onSkip,y=s.afterClose,j=(0,ce.DP)(),F=!!(!(j==null||(t=j.isCompactTheme)===null||t===void 0)&&t.call(j)),B=Sn.Ay.useBreakpoint(),o=Ro({compact:F}),v=o.styles,N=(0,i.useRef)(null),w=(0,i.useState)(null),u=C()(w,2),S=u[0],x=u[1],m=a||S,p=(0,i.useMemo)(function(){var _,A;return Array.isArray(m==null||(_=m.data)===null||_===void 0?void 0:_.challenges)?m==null||(A=m.data)===null||A===void 0?void 0:A.challenges.filter(Boolean):[]},[m]);return(0,i.useEffect)(function(){a&&x(a)},[a]),(0,i.useImperativeHandle)(n,function(){return{show:function(){var A;return(A=N.current)===null||A===void 0?void 0:A.show()},hide:function(){var A;return(A=N.current)===null||A===void 0?void 0:A.hide()}}},[]),(0,e.jsx)(ce.Ht,{ref:N,title:(0,e.jsx)(ce.hE,{children:"\u9A8C\u8BC1\u4FE1\u606F"}),width:F?380:400,onOk:function(){a&&!l&&h()},onCancel:function(){var A;l||(A=N.current)===null||A===void 0||A.hide()},modalProps:{rootClassName:v.modal,zIndex:1050,footer:[(0,e.jsx)(ge.A,{disabled:l,onClick:function(){var A;return(A=N.current)===null||A===void 0?void 0:A.hide()},children:"\u8FD4\u56DE"},"back"),(0,e.jsx)(ge.A,{type:"primary",loading:l,disabled:!a,onClick:h,children:"\u9A8C\u8BC1"},"verify")],forceRender:!0,closable:!l,keyboard:!l,mask:{closable:!l},style:{top:B.md?100:24,paddingBottom:B.md?100:24},focusable:{focusTriggerAfterClose:!1},afterClose:function(){a||x(null),y==null||y()}},children:m&&(0,e.jsxs)("div",{className:v.content,"aria-busy":l,children:[(0,e.jsxs)("div",{className:v.summary,children:[(0,e.jsxs)("div",{className:"verification-tools",children:[(0,e.jsxs)("span",{className:"verification-expiry",children:["\u6709\u6548\u81F3 ",br().unix(m==null||(r=m.data)===null||r===void 0?void 0:r.expires_at).format("MM-DD HH:mm")]}),(0,e.jsxs)("div",{className:"verification-actions",children:[(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",disabled:l||!a,onClick:c,children:"\u91CD\u65B0\u83B7\u53D6"}),(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",disabled:l||!a,onClick:f,children:"\u8DF3\u8FC7\u6B64\u9879"})]})]}),(m==null?void 0:m.error)&&(0,e.jsx)("div",{className:"verification-error",role:"alert",children:m==null?void 0:m.error})]}),p.length===0&&(0,e.jsx)("div",{className:v.hint,children:"\u65E0\u9700\u6DFB\u52A0\u9A8C\u8BC1\u4FE1\u606F\uFF0C\u53EF\u76F4\u63A5\u9A8C\u8BC1\u3002"}),p.map(function(_,A){return(0,e.jsxs)("section",{className:v.record,children:[_.type==="HTTP"&&(0,e.jsxs)("div",{className:"verification-heading",children:[(0,e.jsxs)("span",{children:["\u90E8\u7F72\u9A8C\u8BC1\u6587\u4EF6",p.length>1?" ".concat(A+1):""]}),(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",icon:(0,e.jsx)(ce.In,{type:"DownloadOutlined"}),onClick:function(){var k=URL.createObjectURL(new Blob([_.value],{type:"application/octet-stream"})),T=document.createElement("a");T.href=k,T.download=_.name,T.click(),window.setTimeout(function(){return URL.revokeObjectURL(k)},1e3)},children:"\u4E0B\u8F7D\u6587\u4EF6"})]}),(0,e.jsx)(d.A,{component:"div",layout:"vertical",variant:"filled",children:(_.type==="HTTP"?[{label:"\u6587\u4EF6\u540D",value:_.name},{label:"\u90E8\u7F72\u8DEF\u5F84",value:_.path||""},{label:"\u6587\u4EF6\u5185\u5BB9",value:_.value}]:[{label:"\u8BB0\u5F55\u540D\u79F0",value:_.name},{label:"\u8BB0\u5F55\u503C",value:_.value}]).map(function(E){var k=E.label,T=E.value;return(0,e.jsx)(d.A.Item,{label:_.type!=="HTTP"&&k==="\u8BB0\u5F55\u540D\u79F0"?(0,e.jsxs)("span",{children:[k,(0,e.jsxs)("span",{className:"verification-type",children:["TXT",p.length>1?" ".concat(A+1):""]})]}):k,children:(0,e.jsxs)("div",{className:"verification-field",children:[(0,e.jsx)(ze.A.TextArea,{value:T,readOnly:!0,variant:"filled",autoSize:{minRows:1},spellCheck:!1,"aria-label":k}),(0,e.jsx)(Gt.A.Text,{className:"verification-copy",copyable:{text:T,tooltips:["\u590D\u5236".concat(k),"\u5DF2\u590D\u5236"]}})]})},k)})})]},_.id||A)})]})})});ol.displayName="BatchVerification";var Io=ol,dl,cl,pl,ml,vl,Po=(0,kr.rU)(function(s,n){var t=s.css,r=s.token;return{item:t(dl||(dl=H()([`
        min-width: 0;
        scroll-margin-block: 8px;
        & + & { margin-top: 4px; }
        .ant-form-item { min-width: 0; margin-bottom: 0; }
        .ant-form-item-control-input { min-height: 0; }
        .ant-select { width: 100%; }
        .ant-form-item-explain { font-size: 12px; overflow-wrap: anywhere; }
    `]))),summary:t(cl||(cl=H()([`
        &.ant-collapse-header {
            gap: 10px;
            min-height: `,`px;
            padding: 8px 10px;
            border-radius: `,`px;
            transition: background `,`;
            &:hover, &[aria-expanded="true"] { background: `,`; }
            &:focus-visible { outline: 2px solid `,`; outline-offset: -2px; }
        }
        .batch-title { display: flex; align-items: center; gap: 8px; }
        .batch-domain { min-width: 0; color: `,"; font-size: ",`px; font-weight: 500; line-height: 20px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .batch-kind { flex: none; color: `,`; font-size: 12px; }
        .ant-collapse-arrow { color: `,`; }
        .batch-meta { display: block; color: `,`; font-size: 12px; line-height: 18px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .batch-missing { color: `,`; font-size: 12px; white-space: nowrap; }
    `])),n.compact?48:52,r.borderRadius,r.motionDurationMid,r.colorFillQuaternary,r.colorPrimaryBorder,r.colorText,n.compact?12:13,r.colorTextTertiary,r.colorTextTertiary,r.colorTextTertiary,r.colorWarningText),fields:t(pl||(pl=H()([`
        display: grid;
        gap: `,`px;
        min-width: 0;
        padding: `,`px 10px;
        .ant-form-item-row { flex-wrap: nowrap; }
        .ant-form-item-label { flex: 0 0 `,`px; padding: 0; text-align: left; }
        .ant-form-item-label > label { height: `,"px; color: ","; font-size: ",`px; font-weight: 400; }
        .ant-form-item-control { flex: 1; min-width: 0; }
    `])),n.compact?8:10,n.compact?10:12,n.compact?72:80,n.compact?24:28,r.colorTextSecondary,n.compact?12:13),secret:t(ml||(ml=H()([`
        display: flex;
        align-items: flex-start;
        gap: 6px;
        min-width: 0;
        > .ant-form-item { flex: 1; }
        > .ant-btn { flex: none; }
    `]))),error:t(vl||(vl=H()([`
        padding: 4px 10px;
        color: `,`;
        font-size: 12px;
        line-height: 1.6;
        overflow-wrap: anywhere;
    `])),r.colorError)}}),zo=function(n){var t,r,a,l=n.index,h=n.domain,c=n.renew,f=c===void 0?!1:c,y=n.form,j=n.disabled,F=n.expanded,B=n.onToggle,o=n.secretOptions,v=n.secretLoading,N=n.secretError,w=n.onReloadSecrets,u=n.onManageSecrets,S=n.error,x=(0,ce.DP)(),m=!!(!(x==null||(t=x.isCompactTheme)===null||t===void 0)&&t.call(x)),p=Po({compact:m}),_=p.styles,A=d.A.useWatch(["items",l],{form:y,preserve:!0})||y.getFieldValue(["items",l])||{},E=h.startsWith("*."),k=h.replace(/^\[([^\]]+)\]$/,"$1"),T=k.includes(":")||/^(?:\d{1,3}\.){3}\d{1,3}$/.test(k)&&k.split(".").every(function(q){return Number(q)<=255}),U=Number((r=A.type)!==null&&r!==void 0?r:2)===1,P=!U&&(E||!T&&A.challenge==="dns"),M=!v&&!N&&o.length===0,I=(a=o.find(function(q){return q.value===Number(A.secret_id)}))===null||a===void 0?void 0:a.label,$=[E?"DNS":T?"HTTP":String(A.challenge||"http").toUpperCase(),U?"\u624B\u52A8\u9A8C\u8BC1":"\u81EA\u52A8\u9A8C\u8BC1",f&&!A.ca?"\u6CBF\u7528\u539F\u73AF\u5883":A.ca==="staging"?"\u6D4B\u8BD5\u73AF\u5883":"\u6B63\u5F0F\u73AF\u5883"].concat(Je()(U?[]:[Number(A.auto_renew)===1?"\u81EA\u52A8\u7EED\u7B7E":"\u4E0D\u81EA\u52A8\u7EED\u7B7E"]),Je()(P&&I?[I]:[])).join(" \xB7 ");return(0,e.jsxs)("section",{className:_.item,role:"listitem","aria-label":"".concat(h," \u7684").concat(f?"\u7EED\u7B7E":"\u7533\u8BF7","\u8BBE\u7F6E"),children:[(0,e.jsx)(Xn.A,{ghost:!0,size:"small",expandIconPlacement:"end",activeKey:F?["settings"]:[],onChange:function(){return B(l)},collapsible:j?"disabled":void 0,destroyOnHidden:!1,classNames:{header:_.summary},styles:{header:{alignItems:"center"},title:{minWidth:0},body:{padding:0},icon:{paddingInline:0}},items:[{key:"settings",forceRender:!0,onKeyDown:function(z){!j&&z.key===" "&&z.target===z.currentTarget.querySelector(".ant-collapse-header")&&(z.preventDefault(),B(l))},label:(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("span",{className:"batch-title",children:[(0,e.jsx)("span",{className:"batch-domain",title:h,children:h}),(E||T)&&(0,e.jsx)("span",{className:"batch-kind",children:E?"\u6CDB\u57DF\u540D":"IP"}),f&&(0,e.jsx)("span",{className:"batch-kind",children:"\u7EED\u7B7E"})]}),(0,e.jsx)("span",{className:"batch-meta",title:$,children:$})]}),extra:P&&!A.secret_id?(0,e.jsx)("span",{className:"batch-missing",children:"\u9009\u62E9\u5BC6\u94A5"}):void 0,children:(0,e.jsxs)("div",{className:_.fields,children:[(0,e.jsx)(d.A.Item,{name:["items",l,"challenge"],label:"\u9A8C\u8BC1\u7C7B\u578B",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u9A8C\u8BC1\u7C7B\u578B"}],children:(0,e.jsx)(De.A,{size:"small",variant:"filled","aria-label":"".concat(h," \u7684\u9A8C\u8BC1\u7C7B\u578B"),disabled:j||E||T,options:[{label:"HTTP",value:"http",disabled:E},{label:"DNS",value:"dns",disabled:T}],onChange:function(z){(z!=="dns"||U)&&y.setFieldValue(["items",l,"secret_id"],void 0)}})}),(0,e.jsx)(d.A.Item,{name:["items",l,"type"],label:"\u9A8C\u8BC1\u6A21\u5F0F",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u9A8C\u8BC1\u6A21\u5F0F"}],children:(0,e.jsx)(De.A,{size:"small",variant:"filled","aria-label":"".concat(h," \u7684\u9A8C\u8BC1\u6A21\u5F0F"),disabled:j,options:[{label:"\u81EA\u52A8",value:2},{label:"\u624B\u52A8",value:1}],onChange:function(z){y.setFieldValue(["items",l,"auto_renew"],Number(z)===1?0:1),(Number(z)===1||A.challenge!=="dns")&&y.setFieldValue(["items",l,"secret_id"],void 0)}})}),!U&&(0,e.jsx)(d.A.Item,{name:["items",l,"auto_renew"],label:"\u81EA\u52A8\u7EED\u7B7E",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u662F\u5426\u81EA\u52A8\u7EED\u7B7E"}],children:(0,e.jsx)(De.A,{size:"small",variant:"filled","aria-label":"".concat(h," \u7684\u81EA\u52A8\u7EED\u7B7E"),disabled:j,options:[{label:"\u5173\u95ED",value:0},{label:"\u5F00\u542F",value:1}]})}),(0,e.jsx)(d.A.Item,{name:["items",l,"ca"],label:"\u7B7E\u53D1\u73AF\u5883",rules:[{required:!f,message:"\u8BF7\u9009\u62E9\u7B7E\u53D1\u73AF\u5883"}],children:(0,e.jsx)(De.A,{size:"small",variant:"filled","aria-label":"".concat(h," \u7684\u7B7E\u53D1\u73AF\u5883"),disabled:j,allowClear:f,placeholder:f?"\u6CBF\u7528\u539F\u7B7E\u53D1\u73AF\u5883":"\u9009\u62E9\u7B7E\u53D1\u73AF\u5883",options:[{label:"\u6B63\u5F0F\u73AF\u5883",value:"production"},{label:"\u6D4B\u8BD5\u73AF\u5883",value:"staging"}]})}),P&&(0,e.jsx)(d.A.Item,{label:"\u5173\u8054\u5BC6\u94A5",children:(0,e.jsxs)("div",{className:_.secret,children:[(0,e.jsx)(d.A.Item,{name:["items",l,"secret_id"],rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u4E91\u5BC6\u94A5"}],children:(0,e.jsx)(De.A,{size:"small",variant:"filled","aria-label":"".concat(h," \u7684\u5173\u8054\u5BC6\u94A5"),showSearch:!1,loading:v,disabled:j||v,placeholder:N?"\u5BC6\u94A5\u52A0\u8F7D\u5931\u8D25":!v&&!o.length?"\u6682\u65E0\u5BC6\u94A5":"\u9009\u62E9\u4E91\u5BC6\u94A5",options:o,notFoundContent:v?(0,e.jsx)(mt.A,{size:"small"}):"\u6682\u65E0\u53EF\u7528\u5BC6\u94A5"})}),N&&(0,e.jsx)(Kn.A,{title:"\u91CD\u65B0\u52A0\u8F7D\u5BC6\u94A5",children:(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",htmlType:"button","aria-label":"".concat(h," \u91CD\u65B0\u52A0\u8F7D\u5BC6\u94A5"),disabled:j||v,icon:(0,e.jsx)(ce.In,{type:"ReloadOutlined"}),onClick:w})}),(0,e.jsx)(Kn.A,{title:M?"\u6DFB\u52A0\u5BC6\u94A5":"\u7BA1\u7406\u5BC6\u94A5",children:(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",htmlType:"button","aria-label":"".concat(h," ").concat(M?"\u6DFB\u52A0\u5BC6\u94A5":"\u7BA1\u7406\u5BC6\u94A5"),disabled:j,icon:(0,e.jsx)(ce.In,{type:M?"PlusOutlined":"KeyOutlined"}),onClick:function(){return u(l,M)}})})]})})]})}]}),S&&(0,e.jsx)("div",{className:_.error,role:"alert",children:S})]})},Oo=(0,i.memo)(zo),hl,fl,gl,xl,bl,yl,$o={pending:"\u5F85\u7533\u8BF7",running:"\u7533\u8BF7\u4E2D",waiting:"\u5F85\u9A8C\u8BC1",success:"\u6210\u529F",failed:"\u5931\u8D25",skipped:"\u5DF2\u8DF3\u8FC7"},At=function(n){var t=n.replace(/^\[([^\]]+)\]$/,"$1");if(/^(?:0|[1-9]\d{0,2})(?:\.(?:0|[1-9]\d{0,2})){3}$/.test(t))return t.split(".").every(function(r){return Number(r)<=255});if(!t.includes(":")||!/^[\da-f:.]+$/i.test(t))return!1;try{return new URL("http://[".concat(t,"]")).hostname.startsWith("[")}catch(r){return!1}},Mo=(0,kr.rU)(function(s,n){var t=s.css,r=s.token;return{modal:t(hl||(hl=H()([`
        .ant-modal { max-width: calc(100vw - 24px); outline: none; }
        .ant-modal-body { padding: 0 !important; overflow: hidden; }
        .ant-modal-footer { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
        .ant-modal-footer > .ant-btn { margin-inline-start: 0 !important; }
    `]))),shell:t(fl||(fl=H()([`
        min-height: min(`,`px, calc(100dvh - 290px));
        max-height: calc(100dvh - 290px);
        padding: 4px;
        overflow-x: hidden;
        overflow-y: auto;
        scrollbar-gutter: stable;
        overscroll-behavior: contain;
        @supports not (max-height: 100dvh) { max-height: calc(100vh - 290px); }
        @media (max-width: 760px) {
            min-height: min(`,`px, calc(100dvh - 210px));
            max-height: calc(100dvh - 210px);
            @supports not (max-height: 100dvh) { max-height: calc(100vh - 210px); }
        }
    `])),n.compact?380:410,n.compact?380:410),loading:t(gl||(gl=H()([`
        min-height: 220px;
        display: flex;
        align-items: center;
        justify-content: center;
    `]))),progress:t(xl||(xl=H()([`
        margin-bottom: `,`px;
        .progress-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 6px 12px; margin-bottom: 8px; }
        .progress-title { color: `,`; font-weight: 500; }
        .progress-count, .progress-note { color: `,"; font-size: ",`px; }
        .progress-stats { display: flex; flex-wrap: wrap; gap: 6px 16px; margin-top: 8px; color: `,"; font-size: ",`px; }
        .progress-success { color: `,`; }
        .progress-failed { color: `,`; }
        .progress-note { margin-top: 8px; overflow-wrap: anywhere; }
    `])),n.compact?14:18,r.colorText,r.colorTextSecondary,r.fontSizeSM,r.colorTextSecondary,r.fontSizeSM,r.colorSuccess,r.colorError),taskHeader:t(bl||(bl=H()([`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        margin-bottom: 10px;
        color: `,`;
        font-size: `,`px;
        .ant-btn { flex: none; font-size: `,`px; }
    `])),r.colorTextSecondary,r.fontSizeSM,r.fontSizeSM),taskList:t(yl||(yl=H()([`
        margin: 0;
        padding: 0;
        list-style: none;
        > li {
            padding: 10px;
        }
        > li + li { margin-top: 4px; }
        .task-top { display: flex; align-items: center; gap: 10px; }
        .task-identity { flex: 1; min-width: 0; }
        .task-domain { display: block; font-size: `,`px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .task-method { display: flex; flex-wrap: wrap; gap: 2px 8px; margin-top: 2px; color: `,"; font-size: ",`px; overflow-wrap: anywhere; }
        .task-status { flex: none; font-size: 12px; white-space: nowrap; color: `,`; }
        .task-status.success { color: `,`; }
        .task-status.failed { color: `,`; }
        .task-status.running, .task-status.waiting { color: `,`; }
        .task-error { margin-top: 5px; color: `,`; font-size: 12px; overflow-wrap: anywhere; }
        .task-actions { display: flex; justify-content: flex-end; flex: none; width: 48px; }
        .task-actions .ant-btn { font-size: `,`px; }
        .task-status.running { display: inline-flex; align-items: center; gap: 5px; }
    `])),n.compact?12:13,r.colorTextTertiary,r.fontSizeSM,r.colorTextTertiary,r.colorSuccess,r.colorError,r.colorPrimary,r.colorError,r.fontSizeSM)}}),Fl=(0,i.forwardRef)(function(s,n){var t,r,a=s.onChange,l=s.secretVersion,h=l===void 0?0:l,c=un.A.useApp(),f=c.message,y=(0,ce.DP)(),j=!!(!(y==null||(t=y.isCompactTheme)===null||t===void 0)&&t.call(y)),F=Sn.Ay.useBreakpoint(),B=Mo({compact:j}),o=B.styles,v=d.A.useForm(),N=C()(v,1),w=N[0],u=(0,i.useRef)(null),S=(0,i.useRef)(null),x=(0,i.useRef)(null),m=(0,i.useRef)(null),p=(0,i.useRef)(!1),_=(0,i.useRef)(!1),A=(0,i.useRef)(0),E=(0,i.useRef)(!1),k=(0,i.useRef)(!1),T=(0,i.useRef)(!1),U=(0,i.useRef)(null),P=(0,i.useRef)([]),M=(0,i.useRef)([]),I=(0,i.useRef)(new Map),$=(0,i.useRef)(null),q=(0,i.useRef)(null),z=(0,i.useRef)(void 0),he=(0,i.useRef)(a);he.current=a;var R=(0,i.useRef)(0),g=(0,i.useRef)(null),L=(0,i.useRef)(null),J=(0,i.useState)([]),V=C()(J,2),ee=V[0],fe=V[1],Y=(0,i.useState)(!0),de=C()(Y,2),xe=de[0],ye=de[1],Ce=(0,i.useState)(!1),ie=C()(Ce,2),Be=ie[0],Z=ie[1],cn=(0,i.useState)(""),Le=C()(cn,2),Fe=Le[0],nn=Le[1],Ue=(0,i.useState)(!1),tn=C()(Ue,2),pe=tn[0],Se=tn[1],Ee=(0,i.useState)(!1),we=C()(Ee,2),ue=we[0],K=we[1],ve=(0,i.useState)(!1),me=C()(ve,2),le=me[0],Ae=me[1],re=(0,i.useState)(!1),te=C()(re,2),mn=te[0],se=te[1],Bn=(0,i.useState)(!0),In=C()(Bn,2),Vn=In[0],Zn=In[1],Fr=(0,i.useState)(null),ut=C()(Fr,2),wr=ut[0],mr=ut[1],vr=(0,i.useState)(null),Wr=C()(vr,2),Ur=Wr[0],Cn=Wr[1],zt=(0,i.useState)(null),lt=C()(zt,2),zn=lt[0],W=lt[1],je=(0,i.useState)([]),rn=C()(je,2),Me=rn[0],Ve=rn[1],Oe=(0,i.useState)(!1),sn=C()(Oe,2),Re=sn[0],lr=sn[1],Fn=(0,i.useState)(!1),er=C()(Fn,2),Qn=er[0],sr=er[1],$e=(0,i.useMemo)(function(){var oe={pending:0,running:0,waiting:0,failed:0,success:0,skipped:0,renewal:0},ae=wt()(ee),_e;try{for(ae.s();!(_e=ae.n()).done;){var He=_e.value;oe[He.status]++,He.certificate_id&&oe.renewal++}}catch(xn){ae.e(xn)}finally{ae.f()}return oe},[ee]),nr=ee.length-$e.success-$e.skipped,Xr=mn&&!pe&&!zn&&$e.pending===0,Jr=pe||!!zn||Be||!!Fe,ir=$e.renewal===ee.length&&ee.length>0?"\u7EED\u7B7E":$e.renewal>0?"\u5904\u7406":"\u7533\u8BF7",Rr=ee.length-$e.skipped,Qt=$e.success+$e.failed,hr=zn?ee[zn.index]:void 0,We=(0,i.useMemo)(function(){return zn?{domain:(hr==null?void 0:hr.domain)||"",data:zn.data,error:hr==null?void 0:hr.error}:null},[zn,hr==null?void 0:hr.domain,hr==null?void 0:hr.error]);(0,i.useEffect)(function(){if(Ur){var oe=0,ae=function _e(){var He;if(!(!p.current||Ur.version!==A.current)){if((He=u.current)!==null&&He!==void 0&&He.querySelector(".ant-motion-collapse")){oe=window.requestAnimationFrame(_e);return}w.scrollToField(Ur.field,{block:"center",focus:!0}),Cn(null)}};return oe=window.requestAnimationFrame(ae),function(){return window.cancelAnimationFrame(oe)}}},[w,Ur]);var Ar=(0,i.useCallback)(function(oe,ae){P.current=P.current.map(function(_e,He){return He===oe?D()(D()({},_e),ae):_e}),fe(P.current)},[]),Gr=(0,i.useCallback)(function(){var oe,ae=q.current;q.current=null,ae&&(ae.method==="renew"?Mr:$r)({body:{type:1,action:"cancel",session_id:ae==null||(oe=ae.data)===null||oe===void 0?void 0:oe.session_id,id:ae.id}}).catch(function(){})},[]),Yr=(0,i.useCallback)(ne()(b()().mark(function oe(){var ae,_e,He,xn,an,Hn,hn,fn;return b()().wrap(function(wn){for(;;)switch(wn.prev=wn.next){case 0:return ae=A.current,_e=++R.current,lr(!0),sr(!1),wn.prev=4,wn.next=7,(0,it.GV)({body:{name:"secret"}});case 7:if(an=wn.sent,!(!p.current||ae!==A.current||_e!==R.current)){wn.next=10;break}return wn.abrupt("return");case 10:if(!((an==null?void 0:an.code)!==200||!(an!=null&&(He=an.data)!==null&&He!==void 0&&He.name))){wn.next=12;break}throw new Error("\u83B7\u53D6\u5BC6\u94A5\u5931\u8D25");case 12:Hn=Object.entries((an==null||(xn=an.data)===null||xn===void 0?void 0:xn.name)||{}).map(function(qn){var ln=C()(qn,2),On=ln[0],fr=ln[1];return{value:Number(On),label:String(fr)}}),Ve(Hn),hn=new Set(Hn.map(function(qn){return qn.value})),fn=w.getFieldValue("items")||[],fn.forEach(function(qn,ln){var On=Number(qn==null?void 0:qn.secret_id);On&&!hn.has(On)&&w.setFieldValue(["items",ln,"secret_id"],void 0)}),wn.next=22;break;case 19:wn.prev=19,wn.t0=wn.catch(4),p.current&&ae===A.current&&_e===R.current&&sr(!0);case 22:return wn.prev=22,p.current&&ae===A.current&&_e===R.current&&lr(!1),wn.finish(22);case 25:case"end":return wn.stop()}},oe,null,[[4,19,22,25]])})),[w]),Hr=(0,i.useCallback)(function(){var oe,ae,_e;if(Cn(null),U.current=null,E.current){T.current=!0,k.current=!0,K(!0),Ae(!0);return}A.current+=1,R.current+=1,p.current=!1,_.current=!1,E.current=!1,k.current=!1,T.current=!1,g.current=null,L.current=null,z.current=void 0,M.current=[],I.current.clear(),Gr(),(oe=m.current)===null||oe===void 0||oe.close(),(ae=x.current)===null||ae===void 0||ae.hide(),(_e=S.current)===null||_e===void 0||_e.hide(),Se(!1),K(!1),Ae(!1),W(null),lr(!1),Z(!1),nn("")},[Gr]),Ct=(0,i.useCallback)(function(oe){if(!(!p.current||E.current||q.current)){_.current&&P.current.forEach(function(hn,fn){I.current.set(hn.domain,{id:hn.certificate_id,values:D()({},w.getFieldValue(["items",fn]))})});var ae=++A.current;R.current+=1,g.current=null,L.current=null,_.current=!1,ye(oe),nn(""),mr(null),Cn(null);var _e=M.current,He=_e;if(oe){var xn=new Map,an=new Map;_e.forEach(function(hn){if(!(hn.startsWith("*.")||hn.includes("..")||At(hn))){var fn=hn.indexOf(".");if(!(fn<=0)){var or=hn.slice(fn+1);if(or.includes(".")){var wn="*."+or;xn.set(hn,wn),an.set(wn,(an.get(wn)||0)+1)}}}});var Hn=new Set(_e.map(function(hn){var fn=xn.get(hn);return fn&&(an.get(fn)||0)>1?fn:hn}));He=Je()(Hn).filter(function(hn){if(hn.startsWith("*.")||At(hn))return!0;var fn=hn.indexOf(".");return fn<0||!Hn.has("*."+hn.slice(fn+1))})}He.length&&(P.current=He.map(function(hn){return{domain:hn,status:"pending"}}),$.current=null,fe(P.current),sr(!1),lr(!1),Ve([]),Z(!0),w.resetFields(),ne()(b()().mark(function hn(){var fn,or,wn;return b()().wrap(function(ln){for(;;)switch(ln.prev=ln.next){case 0:return fn=!1,ln.prev=1,or=new Array(He.length),wn=0,ln.next=6,Promise.all(Array.from({length:Math.min(4,He.length)},ne()(b()().mark(function On(){var fr,rr;return b()().wrap(function(En){for(;;)switch(En.prev=En.next){case 0:fr=b()().mark(function Er(){var Cr,Wn,jr,Zr,Ir,gr,jn,Jn;return b()().wrap(function(Ln){for(;;)switch(Ln.prev=Ln.next){case 0:if(Cr=wn++,!(Cr>=He.length)){Ln.next=3;break}return Ln.abrupt("return",{v:void 0});case 3:Wn=He[Cr].replace(/^\[([^\]]+)\]$/,"$1"),jr=1;case 5:if(!(!fn&&p.current&&ae===A.current)){Ln.next=22;break}return Ln.next=8,Qa({body:{domain:Wn,page:jr,page_size:100,order_by_field:"id",order_by_type:"desc"}});case 8:if(jn=Ln.sent,!(fn||!p.current||ae!==A.current)){Ln.next=11;break}return Ln.abrupt("return",{v:void 0});case 11:if(!((jn==null?void 0:jn.code)!==200||!Array.isArray(jn==null||(Zr=jn.data)===null||Zr===void 0?void 0:Zr.list))){Ln.next=13;break}throw new Error((jn==null?void 0:jn.message)||"\u83B7\u53D6\u5DF2\u6709\u8BC1\u4E66\u5931\u8D25");case 13:if(Jn=jn==null||(Ir=jn.data)===null||Ir===void 0?void 0:Ir.list,or[Cr]=Jn==null?void 0:Jn.find(function(Ze){if(!(Ze!=null&&Ze.id)||![1,2].includes(Number(Ze==null?void 0:Ze.type)))return!1;var en=Ze==null?void 0:Ze.domains;if(typeof en=="string")try{en=JSON.parse(en)}catch(Pn){return!1}if(!Array.isArray(en))return!1;var xr=Je()(new Set(en.map(function(Pn){return String(Pn).trim().toLowerCase().replace(/\.+$/,"").replace(/^\[([^\]]+)\]$/,"$1")})));return xr.length===1&&xr[0]===Wn}),!(or[Cr]||jr*100>=Number((jn==null||(gr=jn.data)===null||gr===void 0?void 0:gr.total)||0))){Ln.next=17;break}return Ln.abrupt("break",22);case 17:if(Jn!=null&&Jn.length){Ln.next=19;break}throw new Error("\u83B7\u53D6\u5DF2\u6709\u8BC1\u4E66\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5");case 19:jr++,Ln.next=5;break;case 22:case"end":return Ln.stop()}},Er)});case 1:if(!(!fn&&p.current&&ae===A.current)){En.next=8;break}return En.delegateYield(fr(),"t0",3);case 3:if(rr=En.t0,!rr){En.next=6;break}return En.abrupt("return",rr.v);case 6:En.next=1;break;case 8:case"end":return En.stop()}},On)}))));case 6:if(!(!p.current||ae!==A.current)){ln.next=8;break}return ln.abrupt("return");case 8:P.current=He.map(function(On,fr){var rr;return{domain:On,status:"pending",certificate_id:(rr=or[fr])===null||rr===void 0?void 0:rr.id}}),fe(P.current),w.setFieldsValue({items:He.map(function(On,fr){var rr,An=or[fr],En=I.current.get(On),Er=Number(An==null?void 0:An.type)===1?1:2,Cr=At(On)?"http":On.startsWith("*.")||(An==null?void 0:An.challenge)==="dns"?"dns":"http",Wn=D()({challenge:Cr,type:Er,secret_id:Er===2&&Cr==="dns"&&Number((An==null?void 0:An.secret_id)||0)||void 0,auto_renew:Er===1?0:Number((rr=An==null?void 0:An.auto_renew)!==null&&rr!==void 0?rr:1),ca:An?void 0:"production"},En==null?void 0:En.values);return En&&String(En.id||"")!==String((An==null?void 0:An.id)||"")&&(Wn.ca=An?void 0:"production"),At(On)?Wn.challenge="http":On.startsWith("*.")&&(Wn.challenge="dns"),Number(Wn.type)===1&&(Wn.auto_renew=0),(Number(Wn.type)===1||Wn.challenge!=="dns")&&(Wn.secret_id=void 0),!An&&!Wn.ca&&(Wn.ca="production"),Wn})}),_.current=!0,Yr(),ln.next=19;break;case 15:ln.prev=15,ln.t0=ln.catch(1),fn=!0,p.current&&ae===A.current&&nn(ln.t0 instanceof Error?ln.t0.message:"\u83B7\u53D6\u5DF2\u6709\u8BC1\u4E66\u5931\u8D25");case 19:return ln.prev=19,p.current&&ae===A.current&&Z(!1),ln.finish(19);case 22:case"end":return ln.stop()}},hn,null,[[1,15,19,22]])}))())}},[w,Yr]),jt=(0,i.useCallback)(function(oe,ae){var _e;if(E.current){Hr(),U.current=function(){return jt(oe,ae)};return}Hr(),M.current=Je()(new Set(oe.map(function(He){return He.trim().toLowerCase().replace(/\.+$/,"")}).filter(Boolean))),M.current.length&&(p.current=!0,z.current=ae,se(!1),Zn(!0),Ct(!0),(_e=S.current)===null||_e===void 0||_e.show())},[Hr,Ct]);(0,i.useImperativeHandle)(n,function(){return{open:jt,close:Hr}},[jt,Hr]),(0,i.useEffect)(function(){p.current&&Yr()},[Yr,h]),(0,i.useEffect)(function(){return function(){p.current=!1,_.current=!1,A.current+=1,R.current+=1,Gr()}},[Gr]);var ht=(0,i.useCallback)(function(oe,ae,_e){var He;(He=x.current)===null||He===void 0||He.hide(),q.current=null,W(null),Ar(oe,{status:"success",error:void 0});for(var xn=z.current,an=he.current,Hn=0,hn=[function(){return xn==null?void 0:xn(ae)},function(){return an==null?void 0:an()}];Hn<hn.length;Hn++){var fn=hn[Hn];if(!p.current||_e!==A.current)break;try{Promise.resolve(fn()).catch(function(){})}catch(or){}}},[Ar]),St=(0,i.useCallback)(ne()(b()().mark(function oe(){var ae,_e,He,xn,an,Hn,hn,fn,or,wn,qn,ln,On,fr,rr,An,En,Er,Cr,Wn,jr,Zr,Ir,gr,jn,Jn,Dt,Ln,Ze,en,xr,Pn,kt,Yn,Pr,Sr,Lr=arguments;return b()().wrap(function(Ne){for(;;)switch(Ne.prev=Ne.next){case 0:if(ae=Lr.length>0&&Lr[0]!==void 0?Lr[0]:!1,_e=Lr.length>1&&Lr[1]!==void 0?Lr[1]:!1,He=Lr.length>2&&Lr[2]!==void 0?Lr[2]:!1,!(!p.current||!_.current||E.current||q.current&&!_e)){Ne.next=5;break}return Ne.abrupt("return");case 5:if(xn=A.current,an=function(){return p.current&&xn===A.current},E.current=!0,Se(!0),k.current=!1,K(!1),Ne.prev=11,Hn=P.current.flatMap(function(dr,cr){return(ae!==!1?dr.status==="failed"&&(ae===!0||ae===cr):dr.status==="pending"&&(!He||!$.current||$.current.includes(cr)))?[cr]:[]}),_e){Ne.next=16;break}return Ne.next=16,w.validateFields(Hn.flatMap(function(dr){var cr=w.getFieldValue(["items",dr])||{},et=P.current[dr],Xt=Number(cr.type)!==1,fi=Xt&&!At(et.domain)&&(et.domain.startsWith("*.")||cr.challenge==="dns");return["challenge","type"].concat(Je()(et.certificate_id?[]:["ca"]),Je()(Xt?["auto_renew"]:[]),Je()(fi?["secret_id"]:[])).map(function(xa){return["items",dr,xa]})}));case 16:if(an()){Ne.next=18;break}return Ne.abrupt("return");case 18:hn=w.getFieldsValue(!0).items||[],ae!==!1?($.current=P.current.flatMap(function(dr,cr){return dr.status==="failed"&&(ae===!0||ae===cr)?[cr]:[]}),P.current=P.current.map(function(dr,cr){var et;return(et=$.current)!==null&&et!==void 0&&et.includes(cr)?D()(D()({},dr),{},{status:"pending",error:void 0}):dr}),fe(P.current)):!_e&&!He&&($.current=null),se(!0),Zn(!1),fn=_e?q.current:null;case 23:if(!an()){Ne.next=94;break}if(!(k.current&&!fn)){Ne.next=26;break}return Ne.abrupt("break",94);case 26:if(qn=(or=(wn=fn)===null||wn===void 0?void 0:wn.index)!==null&&or!==void 0?or:P.current.findIndex(function(dr,cr){return dr.status==="pending"&&(!$.current||$.current.includes(cr))}),!(qn<0)){Ne.next=29;break}return Ne.abrupt("break",94);case 29:if(ln=P.current[qn],On=hn[qn]||{},fr=At(ln.domain),rr=fr?"http":ln.domain.startsWith("*.")?"dns":On.challenge,An=Number(On.type)===1,En=ln.certificate_id?"renew":"obtain",Er=D()(D()(D()({},ln.certificate_id?{id:ln.certificate_id}:{name:ln.domain.slice(0,100)}),{},{action:"start",type:An?1:2,domain:fr?ln.domain.replace(/^\[([^\]]+)\]$/,"$1"):ln.domain},On.ca?{ca:On.ca}:{}),{},{challenge:rr,secret_id:!An&&rr==="dns"?Number(On.secret_id||0):0,auto_renew:An?0:Number(On.auto_renew||0)}),Ar(qn,{status:"running",error:void 0,challenge:rr,type:An?1:2,ca:On.ca,secret_id:Er.secret_id,auto_renew:Er.auto_renew}),Ne.prev=37,Ir=void 0,!An){Ne.next=69;break}if(Jn=fn,Jn){Ne.next=59;break}return Ne.next=44,(En==="renew"?Mr:$r)({body:Er});case 44:if(Ze=Ne.sent,en=Ze==null?void 0:Ze.data,an()){Ne.next=49;break}return en!=null&&en.session_id&&(En==="renew"?Mr:$r)({body:{type:1,action:"cancel",session_id:en==null?void 0:en.session_id,id:ln.certificate_id}}).catch(function(){}),Ne.abrupt("return");case 49:if(!((Ze==null?void 0:Ze.code)!==200||!(en!=null&&en.session_id)||!Array.isArray(en==null?void 0:en.challenges)||!Number.isFinite(en==null?void 0:en.expires_at))){Ne.next=52;break}throw en!=null&&en.session_id&&(En==="renew"?Mr:$r)({body:{type:1,action:"cancel",session_id:en==null?void 0:en.session_id,id:ln.certificate_id}}).catch(function(){}),new Error((Ze==null?void 0:Ze.code)!==200?(Ze==null?void 0:Ze.message)||"\u83B7\u53D6\u9A8C\u8BC1\u4FE1\u606F\u5931\u8D25\uFF0C\u8BF7\u91CD\u65B0\u7533\u8BF7":"\u9A8C\u8BC1\u4FE1\u606F\u65E0\u6548\uFF0C\u8BF7\u91CD\u65B0\u7533\u8BF7");case 52:if(Jn={index:qn,data:en,method:En,id:ln.certificate_id},q.current=Jn,W(Jn),T.current||(Dt=x.current)===null||Dt===void 0||Dt.show(),!((en==null||(Ln=en.challenges)===null||Ln===void 0?void 0:Ln.length)>0||k.current)){Ne.next=59;break}return Ar(qn,{status:"waiting"}),Ne.abrupt("return");case 59:if(!(((gr=Jn)===null||gr===void 0||(gr=gr.data)===null||gr===void 0?void 0:gr.expires_at)<=Math.floor(Date.now()/1e3))){Ne.next=64;break}throw Gr(),(xr=x.current)===null||xr===void 0||xr.hide(),W(null),new Error("\u9A8C\u8BC1\u4FE1\u606F\u5DF2\u8FC7\u671F\uFF0C\u8BF7\u91CD\u8BD5\u83B7\u53D6\u65B0\u7684\u9A8C\u8BC1\u4FE1\u606F");case 64:return Ne.next=66,(Jn.method==="renew"?Mr:$r)({body:{type:1,action:"submit",session_id:(jn=Jn)===null||jn===void 0||(jn=jn.data)===null||jn===void 0?void 0:jn.session_id,id:Jn.id}});case 66:Ir=Ne.sent,Ne.next=72;break;case 69:return Ne.next=71,(ln.certificate_id?Mr:$r)({body:Er});case 71:Ir=Ne.sent;case 72:if(an()){Ne.next=74;break}return Ne.abrupt("return");case 74:if(((Cr=Ir)===null||Cr===void 0?void 0:Cr.code)===200){Ne.next=76;break}throw new Error(((Wn=Ir)===null||Wn===void 0?void 0:Wn.message)||(ln.certificate_id?"\u7EED\u7B7E\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5":"\u7533\u8BF7\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5"));case 76:if((jr=Ir)!==null&&jr!==void 0&&(jr=jr.data)!==null&&jr!==void 0&&jr.id){Ne.next=78;break}throw new Error("\u672A\u6536\u5230\u5B8C\u6574\u8BC1\u4E66\u4FE1\u606F\uFF0C\u8BF7\u5148\u5728\u8BC1\u4E66\u7BA1\u7406\u4E2D\u68C0\u67E5\u7ED3\u679C");case 78:ht(qn,(Zr=Ir)===null||Zr===void 0?void 0:Zr.data,xn),fn=null,Ne.next=92;break;case 82:if(Ne.prev=82,Ne.t0=Ne.catch(37),an()){Ne.next=86;break}return Ne.abrupt("return");case 86:if(kt=Ne.t0 instanceof Error?Ne.t0.message:ln.certificate_id?"\u7EED\u7B7E\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5":"\u7533\u8BF7\u5931\u8D25\uFF0C\u8BF7\u91CD\u8BD5",((Pn=q.current)===null||Pn===void 0?void 0:Pn.index)!==qn){Ne.next=90;break}return Ar(qn,{status:"waiting",error:kt}),Ne.abrupt("return");case 90:Ar(qn,{status:"failed",error:kt}),fn=null;case 92:Ne.next=23;break;case 94:Ne.next=99;break;case 96:Ne.prev=96,Ne.t1=Ne.catch(11),an()&&(Zn(!0),Ne.t1&&pi()(Ne.t1)==="object"&&"errorFields"in Ne.t1?(Pr=(Yn=Ne.t1.errorFields)===null||Yn===void 0||(Yn=Yn[0])===null||Yn===void 0?void 0:Yn.name,(Pr==null?void 0:Pr[0])==="items"&&typeof Pr[1]=="number"&&mr(Pr[1]),Pr&&Cn({field:Pr,version:A.current})):f.error("\u8BF7\u68C0\u67E5\u7533\u8BF7\u914D\u7F6E\u540E\u91CD\u8BD5"));case 99:return Ne.prev=99,an()&&(E.current=!1,Se(!1),T.current&&(Sr=U.current,Hr(),Sr==null||Sr())),Ne.finish(99);case 102:case"end":return Ne.stop()}},oe,null,[[11,96,99,102],[37,82]])})),[Gr,Hr,ht,w,f,Ar]),ha=(0,i.useCallback)(function(oe){var ae,_e,He=arguments.length>1&&arguments[1]!==void 0?arguments[1]:!1;E.current||q.current||(g.current={version:A.current,index:oe},L.current=null,He?(ae=m.current)===null||ae===void 0||ae.openCreate():(_e=m.current)===null||_e===void 0||_e.open())},[]),Qr=(0,i.useCallback)(function(oe){Cn(null),mr(function(ae){return ae===oe?null:oe})},[]),fa=(0,i.useCallback)(function(oe){try{var ae;Promise.resolve((ae=he.current)===null||ae===void 0?void 0:ae.call(he)).catch(function(){})}catch(an){}if(p.current){var _e=g.current;if((_e==null?void 0:_e.version)===A.current&&oe.action==="create"&&(L.current=oe.id),(_e==null?void 0:_e.version)===A.current&&oe.action==="update"&&L.current===oe.id&&!E.current&&!q.current){var He,xn=w.getFieldValue(["items",_e.index]);Number(xn==null?void 0:xn.type)===2&&(xn==null?void 0:xn.challenge)==="dns"&&["pending","failed"].includes((He=P.current[_e.index])===null||He===void 0?void 0:He.status)&&w.setFieldValue(["items",_e.index,"secret_id"],oe.id),L.current=null}oe.action==="delete"&&L.current===oe.id&&(L.current=null),Yr()}},[w,Yr]),ga=function(){var ae,_e=arguments.length>0&&arguments[0]!==void 0?arguments[0]:!1;if(!(E.current||!q.current)){var He=q.current.index;(ae=x.current)===null||ae===void 0||ae.hide(),Gr(),W(null),Ar(He,{status:_e?"skipped":"pending",error:void 0}),St(!1,!1,!0)}};return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(ce.Ht,{ref:S,title:(0,e.jsx)(ce.hE,{children:"\u7533\u8BF7\u8BC1\u4E66"}),width:j?560:600,onCancel:Hr,modalProps:{rootClassName:o.modal,zIndex:1030,forceRender:!0,closable:!le,keyboard:!le,mask:{closable:!pe&&!zn},style:{top:F.md?100:24,paddingBottom:F.md?100:24},focusable:{focusTriggerAfterClose:!1},footer:Vn?[(0,e.jsx)(ge.A,{disabled:pe,onClick:function(){return mn?Zn(!1):Hr()},children:mn?"\u8FD4\u56DE":"\u53D6\u6D88"},"back"),(0,e.jsx)(ge.A,{type:"primary",loading:pe,disabled:Be||!!Fe||nr===0,onClick:w.submit,children:mn?$e.pending>0?"\u7EE7\u7EED".concat(ir):"\u91CD\u8BD5":ir==="\u5904\u7406"?"\u5F00\u59CB":ir},"apply")]:[(0,e.jsx)(ge.A,{loading:le,onClick:Hr,children:Xr&&!$e.failed?"\u5B8C\u6210":"\u5173\u95ED"},"close"),pe?(0,e.jsx)(ge.A,{type:"primary",disabled:ue,onClick:function(){k.current=!0,K(!0)},children:ue?"\u6682\u505C\u4E2D":"\u6682\u505C"},"pause"):zn?(0,e.jsx)(ge.A,{type:"primary",onClick:function(){var ae;return(ae=x.current)===null||ae===void 0?void 0:ae.show()},children:"\u9A8C\u8BC1"},"verify"):$e.pending>0||$e.failed>0?(0,e.jsx)(ge.A,{type:"primary",onClick:function(){St($e.pending===0&&$e.failed>0)},children:$e.pending>0?"\u7EE7\u7EED".concat(ir):"\u91CD\u8BD5"},"continue"):null]},children:(0,e.jsxs)("div",{className:o.shell,style:Vn?{minHeight:0}:void 0,children:[Vn&&(0,e.jsxs)("div",{className:o.taskHeader,children:[(0,e.jsx)("span",{children:"\u5408\u5E76\u57DF\u540D"}),(0,e.jsx)(Dr.A,{size:"small",checked:xe,disabled:mn||pe||!!zn,"aria-label":"\u5408\u5E76\u57DF\u540D",onChange:function(ae){mn||Ct(ae)}})]}),Be?(0,e.jsx)("div",{className:o.loading,children:(0,e.jsx)(mt.A,{})}):Fe?(0,e.jsx)("div",{className:o.loading,children:(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:Fe,children:(0,e.jsx)(ge.A,{color:"default",variant:"filled",size:"small",onClick:function(){return Ct(xe)},children:"\u91CD\u8BD5"})})}):null,(0,e.jsx)(da.Ay,{componentSize:"small",theme:{token:{fontSize:j?12:13,fontSizeSM:12,controlHeightSM:j?24:28},components:{Form:{labelFontSize:j?12:13},Select:{optionFontSize:j?12:13},Button:{contentFontSizeSM:j?12:13}}},children:(0,e.jsx)(d.A,{form:w,layout:"horizontal",colon:!1,requiredMark:!1,variant:"filled",disabled:Jr,style:Vn&&!Be&&!Fe?void 0:{display:"none"},onFinishFailed:function(ae){var _e,He=ae.errorFields,xn=ae.outOfDate;if(!(!p.current||xn)){var an=(_e=He[0])===null||_e===void 0?void 0:_e.name;an&&(an[0]==="items"&&typeof an[1]=="number"&&mr(an[1]),Cn({field:an,version:A.current}))}},onFinish:function(){!p.current||!_.current||E.current||q.current||(P.current=P.current.map(function(ae){return ae.status==="failed"?D()(D()({},ae),{},{status:"pending",error:void 0}):ae}),fe(P.current),St())},children:(0,e.jsx)("div",{ref:u,role:"list","aria-label":"\u57DF\u540D\u7533\u8BF7\u8BBE\u7F6E",children:ee.map(function(oe,ae){return!["success","skipped"].includes(oe.status)&&(0,e.jsx)(Oo,{index:ae,domain:oe.domain,renew:!!oe.certificate_id,form:w,disabled:Jr,expanded:wr===ae,onToggle:Qr,secretOptions:Me,secretLoading:Re,secretError:Qn,onReloadSecrets:Yr,onManageSecrets:ha,error:oe.error},"".concat(ae,":").concat(oe.domain))})})})}),!Vn&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:o.progress,"aria-live":"polite",children:[(0,e.jsxs)("div",{className:"progress-heading",children:[(0,e.jsx)("span",{className:"progress-title",children:le?"\u6B63\u5728\u7ED3\u675F\u4EFB\u52A1":pe?ue?"\u6B63\u5728\u6682\u505C":zn?"\u6B63\u5728\u9A8C\u8BC1":"\u6B63\u5728".concat(ir):zn?"\u7B49\u5F85\u624B\u52A8\u9A8C\u8BC1":$e.pending>0?"\u7B49\u5F85\u7EE7\u7EED".concat(ir):$e.failed>0?$e.success>0?"\u90E8\u5206\u8BC1\u4E66".concat(ir,"\u5931\u8D25"):"".concat(ir,"\u5931\u8D25"):$e.success>0?"".concat(ir,"\u5B8C\u6210"):"\u672A\u5904\u7406\u8BC1\u4E66"}),(0,e.jsxs)("span",{className:"progress-count",children:["\u5DF2\u5904\u7406 ",Qt," / ",Rr]})]}),(0,e.jsx)(_o.A,{percent:Rr>0?Math.round(Qt/Rr*100):0,success:{percent:Rr>0?$e.success/Rr*100:0},status:pe?"active":$e.failed>0?"exception":"normal",showInfo:!1,size:"small"}),(0,e.jsxs)("div",{className:"progress-stats",children:[(0,e.jsxs)("span",{className:"progress-success",children:["\u6210\u529F ",$e.success]}),(0,e.jsxs)("span",{className:$e.failed?"progress-failed":void 0,children:["\u5931\u8D25 ",$e.failed]}),(0,e.jsxs)("span",{children:["\u5F85\u5904\u7406 ",Math.max(0,Rr-Qt)]}),$e.skipped>0&&(0,e.jsxs)("span",{children:["\u8DF3\u8FC7 ",$e.skipped]})]}),(le||ue||zn)&&(0,e.jsx)("div",{className:"progress-note",children:le?"\u5F53\u524D\u57DF\u540D\u5904\u7406\u7ED3\u675F\u540E\u5173\u95ED\uFF0C\u540E\u7EED\u57DF\u540D\u4E0D\u518D\u5904\u7406\u3002":pe&&ue?"\u5F53\u524D\u57DF\u540D\u5904\u7406\u7ED3\u675F\u540E\u6682\u505C\u3002":zn?"\u8BF7\u5B8C\u6210 ".concat((r=ee[zn.index])===null||r===void 0?void 0:r.domain," \u7684\u9A8C\u8BC1\uFF0C\u9A8C\u8BC1\u6210\u529F\u540E\u7EE7\u7EED\u4E0B\u4E00\u9879\u3002"):"\u53EF\u4EE5\u8C03\u6574\u8BBE\u7F6E\u540E\u7EE7\u7EED\uFF0C\u5DF2\u6210\u529F\u7684\u8BC1\u4E66\u4E0D\u4F1A\u91CD\u590D\u5904\u7406\u3002"})]}),(0,e.jsxs)("div",{className:o.taskHeader,children:[(0,e.jsx)("span",{children:"\u57DF\u540D\u4E0E\u9A8C\u8BC1\u8BBE\u7F6E"}),!Jr&&nr>0&&(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",onClick:function(){return Zn(!0)},children:"\u4FEE\u6539\u8BBE\u7F6E"})]}),(0,e.jsx)("ul",{className:o.taskList,"aria-label":"\u7533\u8BF7\u8FDB\u5EA6",children:ee.map(function(oe,ae){var _e,He=oe.status==="success"?oe:w.getFieldValue(["items",ae])||oe,xn=At(oe.domain)?"http":oe.domain.startsWith("*.")?"dns":He.challenge||"http";return(0,e.jsxs)("li",{children:[(0,e.jsxs)("div",{className:"task-top",children:[(0,e.jsxs)("div",{className:"task-identity",children:[(0,e.jsx)(Kn.A,{title:oe.domain,children:(0,e.jsx)("span",{className:"task-domain",children:oe.domain})}),(0,e.jsxs)("div",{className:"task-method",children:[(0,e.jsx)("span",{children:oe.certificate_id?"\u7EED\u7B7E":"\u7533\u8BF7"}),(0,e.jsxs)("span",{children:[xn.toUpperCase()," \xB7 ",Number(He.type)===1?"\u624B\u52A8":"\u81EA\u52A8"]}),(0,e.jsx)("span",{children:oe.certificate_id&&!He.ca?"\u6CBF\u7528\u539F\u73AF\u5883":He.ca==="staging"?"\u6D4B\u8BD5\u73AF\u5883":"\u6B63\u5F0F\u73AF\u5883"}),Number(He.type)!==1&&xn==="dns"&&(0,e.jsx)("span",{children:((_e=Me.find(function(an){return an.value===Number(He.secret_id)}))===null||_e===void 0?void 0:_e.label)||(He.secret_id?"\u5BC6\u94A5 #".concat(He.secret_id):"\u672A\u9009\u62E9\u5BC6\u94A5")})]})]}),(0,e.jsxs)("span",{className:"task-status ".concat(oe.status),children:[oe.status==="running"&&(0,e.jsx)(ce.In,{type:"LoadingOutlined",spin:!0}),oe.status==="running"&&(zn==null?void 0:zn.index)===ae?"\u9A8C\u8BC1\u4E2D":oe.certificate_id&&oe.status==="pending"?"\u5F85\u7EED\u7B7E":oe.certificate_id&&oe.status==="running"?"\u7EED\u7B7E\u4E2D":$o[oe.status]]}),(0,e.jsxs)("span",{className:"task-actions",children:[!pe&&(zn==null?void 0:zn.index)===ae&&(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",onClick:function(){var Hn;return(Hn=x.current)===null||Hn===void 0?void 0:Hn.show()},children:"\u9A8C\u8BC1"}),!Jr&&oe.status==="failed"&&(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",onClick:function(){St(ae)},"aria-label":"\u91CD\u8BD5 ".concat(oe.domain),children:"\u91CD\u8BD5"}),!Jr&&["pending","skipped"].includes(oe.status)&&(0,e.jsx)(ge.A,{size:"small",type:"text",onClick:function(){return Ar(ae,{status:oe.status==="skipped"?"pending":"skipped",error:void 0})},children:oe.status==="skipped"?"\u6062\u590D":"\u8DF3\u8FC7"})]})]}),oe.error&&(0,e.jsx)("div",{className:"task-error",role:"alert",children:oe.error})]},oe.domain)})})]})]})}),(0,e.jsx)(Io,{ref:x,session:We,busy:pe,onVerify:function(){St(!1,!0)},onRefresh:function(){return ga()},onSkip:function(){return ga(!0)}}),(0,e.jsx)(mi,{ref:m,onChange:fa})]})});Fl.displayName="BatchCert";var Ho=(0,i.memo)(Fl),wl,Al,El,Cl,jl,Sl,Dl,kl,Bl,Nl,Tl,_l,Rl,Il,Lo=["id"],Vo=function(n){var t=n.body||{},r=t.id,a=pt()(t,Lo);return Ns(D()(D()({},n),{},{body:D()({action:"read"},a)}))},vt=function(n){var t=String(n||"").trim().replace(/^\[([^\]]+)\]$/,"$1");if(/^(?:0|[1-9]\d{0,2})(?:\.(?:0|[1-9]\d{0,2})){3}$/.test(t))return t.split(".").every(function(r){return Number(r)<=255});if(!t.includes(":")||!/^[\da-f:.]+$/i.test(t))return!1;try{return new URL("http://[".concat(t,"]")).hostname.startsWith("[")}catch(r){return!1}},It=function(n){var t=Number((n==null?void 0:n.type)||0),r=[];try{var a=Array.isArray(n==null?void 0:n.domains)?n==null?void 0:n.domains:JSON.parse((n==null?void 0:n.domains)||"[]");Array.isArray(a)&&(r=a.filter(function(l){return typeof l=="string"&&l.trim()}))}catch(l){}return D()(D()({},n),{},{domains:r,type:t,challenge:t===1||t===2?(n==null?void 0:n.challenge)||"http":"",secret_id:Number((n==null?void 0:n.secret_id)||0),auto_renew:Number((n==null?void 0:n.auto_renew)||0)})},qo=(0,kr.rU)(function(s){var n=s.css,t=s.token,r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};return{modal:n(wl||(wl=H()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-body {
            padding: 0 !important;
            overflow: hidden;
        }
    `]))),certificateTable:n(Al||(Al=H()([`
        .record-name,
        .record-domains,
        .record-type,
        .record-secret {
            display: block;
            min-width: 0;
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            line-height: 20px;
        }

        .record-name {
            padding: 0;
            border: 0;
            background: transparent;
            color: inherit;
            font: inherit;
            text-align: left;
        }

        button.record-name {
            cursor: pointer;
        }

        button.record-name:hover, button.record-name:focus-visible {
            color: `,`;
        }

        .record-secret {
            color: `,`;
        }

        .record-issuance {
            display: flex;
            align-items: center;
            gap: 8px;
            min-width: 0;
            height: 20px;
        }

        .record-issuance .record-type {
            flex: none;
        }

        .record-issuance .record-secret {
            flex: 0 1 auto;
        }

        .record-renewal {
            flex: none;
            color: `,`;
            white-space: nowrap;
        }

        .record-domains {
            color: `,`;
        }

        .record-state {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            line-height: 20px;
            white-space: nowrap;
            font-variant-numeric: tabular-nums;
        }

        .record-state i {
            width: 6px;
            height: 6px;
            flex: none;
            border-radius: 50%;
            background: currentColor;
        }

        .record-state.valid {
            color: `,`;
        }

        .record-state.expiring {
            color: `,`;
        }

        .record-state.expired {
            color: `,`;
        }

        .record-state.unknown {
            color: `,`;
        }
    `])),t.colorPrimary,t.colorTextSecondary,t.colorTextSecondary,t.colorTextSecondary,t.colorSuccess,t.colorWarning,t.colorError,t.colorTextTertiary),loading:n(El||(El=H()([`
        display: flex;
        flex: 1;
        min-height: var(--certificate-form-min-height, 180px);
        align-items: center;
        justify-content: center;
    `]))),formShell:n(Cl||(Cl=H()([`
        box-sizing: border-box;
        width: 100%;
        min-height: var(--certificate-form-min-height, 0px);
        max-height: calc(100dvh - 290px);
        flex-shrink: 0;
        margin: 0 auto;
        padding: 4px 4px;
        scrollbar-gutter: stable;
        overflow-x: hidden;
        overflow-y: auto;
        overscroll-behavior: contain;

        @supports not (max-height: 100dvh) {
            max-height: calc(100vh - 320px);
        }

        @media (max-width: 767px) {
            max-height: calc(100dvh - 174px);

            @supports not (max-height: 100dvh) {
                max-height: calc(100vh - 174px);
            }
        }

    `]))),formSection:n(jl||(jl=H()([`
        & + & {
            margin-top: `,`px;
            padding-top: 0;
        }

        .ant-form-item:last-child {
            margin-bottom: 0;
        }
    `])),r!=null&&r.compact?15:18),formHints:n(Sl||(Sl=H()([`
        .ant-form-item-extra,
        .ant-form-item-extra .ant-btn {
            font-size: `,`px;
        }
    `])),t.fontSizeSM),pemGrid:n(Dl||(Dl=H()([`
        .ant-form-item {
            margin-bottom: 0;
        }

        @media (max-width: 767px) {
            .ant-col + .ant-col {
                margin-top: 14px;
            }
        }
    `]))),editor:n(kl||(kl=H()([`
        &.ant-input {
            height: `,`px;
            overflow-y: auto;
            resize: none;
        }

        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
        font-size: 12px !important;
        line-height: 1.55;
        tab-size: 2;
    `])),r!=null&&r.compact?180:200),manualShell:n(Bl||(Bl=H()([`
        width: 100%;
        overflow: hidden;
    `]))),manualContent:n(Nl||(Nl=H()([`
        box-sizing: border-box;
        width: 100%;
        padding: 4px;
    `]))),verificationRecord:n(Tl||(Tl=H()([`
        min-width: 0;

        & + & {
            margin-top: `,`px;
            padding-top: `,`px;
            border-top: 1px solid `,`;
        }

        .verification-heading {
            display: flex;
            align-items: flex-start;
            justify-content: space-between;
            gap: 12px;
            margin-bottom: `,`px;
        }

        .verification-heading-text {
            display: flex;
            align-items: center;
            flex-wrap: wrap;
            gap: 4px 12px;
            min-width: 0;
            line-height: 22px;
        }

        .verification-title {
            color: `,`;
            font-size: `,`px;
            font-weight: 500;
        }

        .verification-expiry {
            color: `,`;
            font-size: 12px;
            white-space: nowrap;
        }

        .verification-heading > .ant-btn {
            flex: none;
            height: 22px;
            padding-inline: 6px;
            gap: 4px;
            font-size: 12px;
        }

        .verification-fields .ant-form-item {
            min-width: 0;
            margin-bottom: `,`px;
        }

        .verification-fields .ant-form-item:last-child {
            margin-bottom: 0;
        }

        .verification-fields .ant-form-item-label {
            padding-bottom: `,`px;
        }

        .verification-fields .ant-form-item-label > label {
            width: 100%;
        }

        .verification-field-label {
            display: flex;
            align-items: baseline;
            justify-content: space-between;
            flex-wrap: wrap;
            gap: 2px 12px;
            width: 100%;
        }

        .verification-field-label .verification-expiry {
            color: `,`;
            font-weight: 400;
        }

        .verification-type {
            margin-inline-start: 6px;
            color: `,`;
            font-size: 12px;
            font-weight: 400;
        }

        .verification-field-content {
            position: relative;
            min-width: 0;
        }

        .verification-fields textarea.ant-input {
            padding-inline-end: 32px;
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            font-size: 12px;
            line-height: 1.65;
            overflow-wrap: anywhere;
            resize: none;
        }

        .verification-copy {
            position: absolute;
            top: `,`px;
            inset-inline-end: `,`px;
            font-size: 12px;
            line-height: 22px;
        }

        .verification-copy .ant-typography-copy {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 22px;
            height: 22px;
            margin-inline-start: 0;
        }
    `])),r!=null&&r.compact?18:22,r!=null&&r.compact?18:22,t.colorBorderSecondary,r!=null&&r.compact?12:16,t.colorText,t.fontSize,t.colorTextSecondary,r!=null&&r.compact?12:16,r!=null&&r.compact?4:6,t.colorTextTertiary,t.colorTextSecondary,r!=null&&r.compact?2:4,r!=null&&r.compact?4:6),details:n(_l||(_l=H()([`
        margin-bottom: `,`px;
        padding: `,`;
        border-radius: `,`px;
        background: `,`;
        font-size: `,`px;
    `])),r!=null&&r.compact?12:16,r!=null&&r.compact?"8px 10px":"10px 12px",t.borderRadius,t.colorFillQuaternary,t.fontSizeSM),fieldLabel:n(Rl||(Rl=H()([`
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;

        .ant-btn {
            height: 22px;
            padding: 0 4px;
            font-size: 12px;
        }
    `]))),pemField:n(Il||(Il=H()([`
        .ant-form-item-label > label {
            width: 100%;
        }
    `])))}}),Pl=(0,i.forwardRef)(function(s,n){var t,r,a=s.onChange,l=s.secretVersion,h=l===void 0?0:l,c=un.A.useApp(),f=c.message,y=c.modal,j=(0,ce.DP)(),F=!!(!(j==null||(t=j.isCompactTheme)===null||t===void 0)&&t.call(j)),B=Sn.Ay.useBreakpoint(),o=qo({compact:F}),v=o.styles,N=(0,i.useRef)(null),w=(0,i.useRef)(null),u=(0,i.useRef)(null),S=(0,i.useRef)(null),x=(0,i.useRef)(null),m=(0,i.useRef)(null),p=(0,i.useRef)(null),_=(0,i.useRef)(null),A=(0,i.useRef)(null),E=(0,i.useRef)(h),k=(0,i.useRef)(0),T=(0,i.useRef)(0),U=(0,i.useRef)(0),P=(0,i.useRef)(!1),M=(0,i.useRef)(!1),I=(0,i.useRef)(null),$=(0,i.useRef)(null),q=(0,i.useRef)(null),z=(0,i.useRef)(""),he=d.A.useForm(),R=C()(he,1),g=R[0],L=(0,i.useState)(!1),J=C()(L,2),V=J[0],ee=J[1],fe=(0,i.useState)(!1),Y=C()(fe,2),de=Y[0],xe=Y[1],ye=(0,i.useState)(!1),Ce=C()(ye,2),ie=Ce[0],Be=Ce[1],Z=(0,i.useState)(""),cn=C()(Z,2),Le=cn[0],Fe=cn[1],nn=(0,i.useState)(!1),Ue=C()(nn,2),tn=Ue[0],pe=Ue[1],Se=(0,i.useState)(!1),Ee=C()(Se,2),we=Ee[0],ue=Ee[1],K=(0,i.useState)(!1),ve=C()(K,2),me=ve[0],le=ve[1],Ae=(0,i.useState)("import"),re=C()(Ae,2),te=re[0],mn=re[1],se=te==="view",Bn=se||te==="import"||te==="edit",In=(0,i.useState)(""),Vn=C()(In,2),Zn=Vn[0],Fr=Vn[1],ut=(0,i.useState)("http"),wr=C()(ut,2),mr=wr[0],vr=wr[1],Wr=(0,i.useState)(null),Ur=C()(Wr,2),Cn=Ur[0],zt=Ur[1],lt=(0,i.useMemo)(function(){return Array.isArray(Cn==null?void 0:Cn.challenges)?Cn==null?void 0:Cn.challenges.filter(Boolean):[]},[Cn]),zn=(0,i.useState)([]),W=C()(zn,2),je=W[0],rn=W[1],Me=(0,i.useState)(!1),Ve=C()(Me,2),Oe=Ve[0],sn=Ve[1],Re=(0,i.useState)(!1),lr=C()(Re,2),Fn=lr[0],er=lr[1],Qn=(0,i.useState)(!1),sr=C()(Qn,2),$e=sr[0],nr=sr[1],Xr=(0,i.useState)(!1),Jr=C()(Xr,2),ir=Jr[0],Rr=Jr[1],Qt=(0,i.useState)(),hr=C()(Qt,2),We=hr[0],Ar=hr[1],Gr=(We==null?void 0:We.type)===0,Yr=(te==="import"||te==="obtain"||te==="edit"&&Gr?440:se?Gr?480:520:600)-(F?40:0),Hr=(0,i.useState)([]),Ct=C()(Hr,2),jt=Ct[0],ht=Ct[1],St=(0,i.useState)(1),ha=C()(St,2),Qr=ha[0],fa=ha[1],ga=(0,i.useState)(10),oe=C()(ga,2),ae=oe[0],_e=oe[1],He=(0,i.useState)(0),xn=C()(He,2),an=xn[0],Hn=xn[1],hn=(0,i.useState)(""),fn=C()(hn,2),or=fn[0],wn=fn[1],qn=(0,i.useState)("keyword"),ln=C()(qn,2),On=ln[0],fr=ln[1],rr=(0,i.useState)(""),An=C()(rr,2),En=An[0],Er=An[1],Cr=(0,i.useState)("keyword"),Wn=C()(Cr,2),jr=Wn[0],Zr=Wn[1],Ir=d.A.useWatch("secret_id",g),gr=Number((r=d.A.useWatch("type",g))!==null&&r!==void 0?r:2),jn=mr==="dns"&&gr===2,Jn=jn&&!Ir,Dt=gr===1,Ln=(0,i.useMemo)(function(){return Object.fromEntries(je.map(function(O){var G=O.value,Q=O.label;return[G,Q]}))},[je]);(0,i.useLayoutEffect)(function(){var O=S.current,G=O==null?void 0:O.parentElement;if(!(!Cn||!O||!G)){var Q=function(){var Ge="".concat(O.offsetHeight,"px");G.style.height!==Ge&&(G.style.height=Ge)};Q();var be=new ResizeObserver(Q);return be.observe(O),function(){return be.disconnect()}}},[Cn]);var Ze=(0,i.useCallback)(ne()(b()().mark(function O(){var G,Q,be,ke,Ge,Pe,Xe,pn,Qe,bn=arguments;return b()().wrap(function(Nn){for(;;)switch(Nn.prev=Nn.next){case 0:return G=bn.length>0&&bn[0]!==void 0?bn[0]:Qr,Q=bn.length>1&&bn[1]!==void 0?bn[1]:ae,be=bn.length>2&&bn[2]!==void 0?bn[2]:En,ke=bn.length>3&&bn[3]!==void 0?bn[3]:jr,Ge=++k.current,Er(be.trim()),Zr(ke),xe(!0),Be(!1),Nn.prev=9,Nn.next=12,Qa({body:D()(D()({page:G,page_size:Q},be.trim()?Yu()({},ke,be.trim()):{}),{},{order_by_field:"id",order_by_type:"desc"},q.current?{domain:q.current.domain,usable:1}:{})});case 12:if(Pe=Nn.sent,k.current===Ge){Nn.next=15;break}return Nn.abrupt("return");case 15:if((Pe==null?void 0:Pe.code)===200){Nn.next=20;break}return Be(!0),q.current&&(ht([]),Hn(0)),f.error((Pe==null?void 0:Pe.message)||"\u83B7\u53D6\u8BC1\u4E66\u5217\u8868\u5931\u8D25"),Nn.abrupt("return");case 20:Xe=(Pe==null?void 0:Pe.data)||{},pn=Array.isArray(Xe==null?void 0:Xe.list)?Xe==null?void 0:Xe.list.filter(function(Vr){return Vr==null?void 0:Vr.id}).map(It):[],Qe=Number((Xe==null?void 0:Xe.total)||0),ht(pn),Hn(Qe),fa(Number((Xe==null?void 0:Xe.page)||G)),_e(Number((Xe==null?void 0:Xe.page_size)||Q)),Nn.next=32;break;case 29:Nn.prev=29,Nn.t0=Nn.catch(9),k.current===Ge&&(Be(!0),q.current&&(ht([]),Hn(0)),f.error("\u83B7\u53D6\u8BC1\u4E66\u5217\u8868\u5931\u8D25"));case 32:return Nn.prev=32,k.current===Ge&&xe(!1),Nn.finish(32);case 35:case"end":return Nn.stop()}},O,null,[[9,29,32,35]])})),[jr,En,f,Qr,ae]),en=(0,i.useCallback)(ne()(b()().mark(function O(){var G,Q,be,ke,Ge,Pe;return b()().wrap(function(pn){for(;;)switch(pn.prev=pn.next){case 0:return G=++U.current,sn(!0),er(!1),pn.prev=3,pn.next=6,(0,it.GV)({body:{name:"secret"}});case 6:if(ke=pn.sent,U.current===G){pn.next=9;break}return pn.abrupt("return");case 9:if(!((ke==null?void 0:ke.code)!==200||!(ke!=null&&(Q=ke.data)!==null&&Q!==void 0&&Q.name))){pn.next=12;break}return er(!0),pn.abrupt("return");case 12:if(Ge=Object.entries((ke==null||(be=ke.data)===null||be===void 0?void 0:be.name)||{}).map(function(Qe){var bn=C()(Qe,2),pr=bn[0],Nn=bn[1];return{value:Number(pr),label:String(Nn)}}),rn(Ge),P.current){pn.next=16;break}return pn.abrupt("return");case 16:Pe=g.getFieldValue("secret_id"),Pe&&!Ge.some(function(Qe){return Qe.value===Pe})&&g.setFieldsValue({secret_id:void 0,auto_renew:0}),pn.next=23;break;case 20:pn.prev=20,pn.t0=pn.catch(3),U.current===G&&er(!0);case 23:return pn.prev=23,U.current===G&&sn(!1),pn.finish(23);case 26:case"end":return pn.stop()}},O,null,[[3,20,23,26]])})),[g]);(0,i.useEffect)(function(){E.current!==h&&(E.current=h,(V||P.current&&te!=="import")&&en(),V&&Ze())},[V,Ze,en,te,h]);var xr=(0,i.useCallback)(function(){var O=I.current;if(I.current=null,O){var G;(O.method==="renew"?Mr:$r)({body:{type:1,action:"cancel",session_id:O==null||(G=O.data)===null||G===void 0?void 0:G.session_id,id:O.id}}).catch(function(){})}},[]),Pn=(0,i.useCallback)(function(){var O,G,Q=arguments.length>0&&arguments[0]!==void 0?arguments[0]:!1,be=Q&&P.current;if(T.current+=1,xr(),M.current=!1,pe(!1),ue(!1),(O=u.current)===null||O===void 0||O.hide(),be){var ke;(ke=w.current)===null||ke===void 0||ke.show();return}_.current=null,A.current=null,U.current+=1,P.current=!1,$.current=null,z.current="",(G=w.current)===null||G===void 0||G.hide(),le(!1),sn(!1)},[xr]),kt=(0,i.useCallback)(function(){var O,G,Q=arguments.length>0&&arguments[0]!==void 0?arguments[0]:!1;_.current=P.current?T.current:null,A.current=null,Q?(O=m.current)===null||O===void 0||O.openCreate():(G=m.current)===null||G===void 0||G.open()},[]),Yn=(0,i.useCallback)(function(O,G){var Q,be,ke;if((Q=x.current)===null||Q===void 0||Q.close(),T.current+=1,xr(),(be=u.current)===null||be===void 0||be.hide(),U.current+=1,P.current=!0,M.current=!1,$.current=null,mn(O),Fr(""),Ar(G),ue(!1),pe(!1),g.resetFields(),z.current="",vr("http"),sn(!1),er(!1),nr(!1),Rr(!1),O==="obtain"){var Ge,Pe=((Ge=q.current)===null||Ge===void 0?void 0:Ge.domain)||"",Xe=Pe.startsWith("*."),pn=Xe?"dns":"http";Fr(Pe),nr(Xe),Rr(vt(Pe)),vr(pn),z.current=Pe.slice(0,100),g.setFieldsValue({domain:Pe,name:z.current,ca:"production",challenge:pn,type:2,auto_renew:1})}else if(O==="import"&&q.current)g.setFieldValue("name",q.current.domain.slice(0,100));else if(O==="renew"){var Qe,bn,pr=String((G==null||(Qe=G.domains)===null||Qe===void 0?void 0:Qe[0])||""),Nn=pr.trim().startsWith("*."),Vr=vt(pr),ba=String((G==null?void 0:G.challenge)||"http"),ft=Number((bn=G==null?void 0:G.type)!==null&&bn!==void 0?bn:2),gt=Vr?"http":Nn?"dns":ba,Bt=Number((G==null?void 0:G.secret_id)||0);vr(gt),nr(Nn),Rr(Vr),g.setFieldsValue({domain:pr,challenge:gt,type:ft,secret_id:ft===2&&gt==="dns"&&Bt?Bt:void 0,auto_renew:ft!==2||gt==="dns"&&!Bt?0:Number((G==null?void 0:G.auto_renew)||0)})}if(le(!0),(ke=w.current)===null||ke===void 0||ke.show(),(O==="obtain"||O==="renew")&&en(),!(O!=="edit"&&O!=="view")){var Nt=++T.current;ue(!0),du({body:{id:G==null?void 0:G.id}}).then(function(zr){if(T.current===Nt){if((zr==null?void 0:zr.code)!==200){f.error((zr==null?void 0:zr.message)||"\u83B7\u53D6\u8BC1\u4E66\u8BE6\u60C5\u5931\u8D25"),Pn();return}var $n=It(zr==null?void 0:zr.data);Ar($n),g.setFieldsValue({name:$n.name,certificate:$n.certificate,private_key:$n.private_key,challenge:$n.challenge,type:$n.type,secret_id:$n.type===2&&$n.challenge==="dns"&&$n.secret_id||void 0,auto_renew:$n.type!==2||$n.challenge==="dns"&&!$n.secret_id?0:$n.auto_renew}),vr($n.challenge),nr($n.domains.some(function(Jt){return Jt.startsWith("*.")})),Rr($n.domains.some(vt)),($n.type===1||$n.type===2)&&en()}}).catch(function(){T.current===Nt&&(f.error("\u83B7\u53D6\u8BC1\u4E66\u8BE6\u60C5\u5931\u8D25"),Pn())}).finally(function(){T.current===Nt&&ue(!1)})}},[xr,Pn,g,en,f]),Pr=(0,i.useCallback)(function(O){var G,Q;(G=x.current)===null||G===void 0||G.close(),Pn(),q.current=O||null,Fe((O==null?void 0:O.domain)||""),k.current+=1,ee(!0),xe(!0),Be(!1),ht([]),Hn(0),fa(1),wn(""),fr("keyword"),Er(""),Zr("keyword"),(Q=N.current)===null||Q===void 0||Q.show(),en(),V&&Ze(1,ae,"","keyword")},[V,Pn,Ze,en,ae]),Sr=(0,i.useCallback)(function(){var O,G,Q;(O=x.current)===null||O===void 0||O.close(),(G=m.current)===null||G===void 0||G.close(),Pn(),q.current=null,Fe(""),k.current+=1,ee(!1),xe(!1),Be(!1),ht([]),Hn(0),(Q=N.current)===null||Q===void 0||Q.hide()},[Pn]),Lr=(0,i.useCallback)(function(O,G){var Q;if(q.current=null,Fe(""),k.current+=1,ee(!1),xe(!1),(Q=N.current)===null||Q===void 0||Q.hide(),Array.isArray(O)){var be;Pn(),(be=x.current)===null||be===void 0||be.open(O,G);return}var ke=O.trim();if(ke){Yn("obtain");var Ge=ke.slice(0,100),Pe=ke.startsWith("*.");$.current={domain:ke,onObtained:G},z.current=Ge,Fr(ke),nr(Pe),Rr(vt(ke)),vr(Pe?"dns":"http"),g.setFieldsValue({domain:ke,name:Ge,challenge:Pe?"dns":"http"})}},[Pn,g,Yn]),Kt=(0,i.useCallback)(function(O,G){var Q=O.trim();Q&&Pr({domain:Q,onSelected:G})},[Pr]),Ne=(0,i.useCallback)((function(){var O=ne()(b()().mark(function G(Q,be,ke){var Ge,Pe,Xe;return b()().wrap(function(Qe){for(;;)switch(Qe.prev=Qe.next){case 0:if(Sr(),Ge=It(Q),Q!=null&&Q.domains){Qe.next=21;break}return Yn("renew"),ue(!0),Pe=++T.current,Qe.prev=6,Qe.next=9,du({body:{id:Q==null?void 0:Q.id}});case 9:if(Xe=Qe.sent,!(T.current!==Pe||!P.current)){Qe.next=12;break}return Qe.abrupt("return");case 12:if((Xe==null?void 0:Xe.code)===200){Qe.next=14;break}throw new Error((Xe==null?void 0:Xe.message)||"\u83B7\u53D6\u8BC1\u4E66\u8BE6\u60C5\u5931\u8D25");case 14:Ge=It(Xe==null?void 0:Xe.data),Qe.next=21;break;case 17:return Qe.prev=17,Qe.t0=Qe.catch(6),T.current===Pe&&(f.error(Qe.t0 instanceof Error?Qe.t0.message:"\u83B7\u53D6\u8BC1\u4E66\u8BE6\u60C5\u5931\u8D25"),Pn()),Qe.abrupt("return");case 21:if([1,2].includes(Ge.type)){Qe.next=25;break}return f.info("\u5BFC\u5165\u8BC1\u4E66\u65E0\u6CD5\u76F4\u63A5\u7EED\u7B7E\uFF0C\u8BF7\u7533\u8BF7\u6216\u5BFC\u5165\u65B0\u8BC1\u4E66\u3002"),Kt(ke||Ge.domains[0]||"",be||function(){}),Qe.abrupt("return");case 25:Yn("renew",Ge),$.current={domain:"",onObtained:be};case 27:case"end":return Qe.stop()}},G,null,[[6,17]])}));return function(G,Q,be){return O.apply(this,arguments)}})(),[Sr,Pn,f,Kt,Yn]);(0,i.useImperativeHandle)(n,function(){return{open:Pr,close:Sr,openObtain:Lr,openDeploy:Kt,openRenew:Ne,refresh:Ze}},[Sr,Ze,Pr,Lr,Kt,Ne]),(0,i.useEffect)(function(){return function(){k.current+=1,T.current+=1,U.current+=1,P.current=!1,$.current=null,q.current=null,xr()}},[xr]);var dr=(0,i.useCallback)((function(){var O=ne()(b()().mark(function G(Q){var be,ke,Ge,Pe,Xe,pn,Qe,bn,pr,Nn,Vr,ba,ft,gt,Bt,Nt,zr,$n,Jt,Yt,gi,ya,Ot,$t,xt,xi,bi,yi,Mt,Zt,nt,Tn,Fi,wi,Ai,Ei,ea,cs,Ci,Ht,ps,ji,Si,Di,ki,Bi,Ni,Ti,_i;return b()().wrap(function(Ie){for(;;)switch(Ie.prev=Ie.next){case 0:if(!(!P.current||M.current||we||te==="view")){Ie.next=2;break}return Ie.abrupt("return");case 2:if(be=++T.current,ke=$.current,Ge=q.current,M.current=!0,pe(!0),Ie.prev=7,te!=="import"){Ie.next=14;break}return Ie.next=11,Ts({body:Q});case 11:Qe=Ie.sent,Ie.next=76;break;case 14:if(te!=="edit"){Ie.next=27;break}for(bn={id:We==null?void 0:We.id},pr=0,Nn=["name","certificate","private_key"];pr<Nn.length;pr++)Vr=Nn[pr],Q[Vr]!==(We==null?void 0:We[Vr])&&(bn[Vr]=Q[Vr]);if((We==null?void 0:We.type)===1||(We==null?void 0:We.type)===2)for(ba={challenge:Q.challenge,type:Number(Q.type),secret_id:Number(Q.type)===2&&Q.challenge==="dns"?Number(Q.secret_id||0):0,auto_renew:Number(Q.type)!==2||Q.challenge==="dns"&&!Q.secret_id?0:Number(Q.auto_renew||0)},ft=0,gt=Object.entries(ba);ft<gt.length;ft++)Bt=C()(gt[ft],2),Nt=Bt[0],zr=Bt[1],zr!==We[Nt]&&(bn[Nt]=String(zr));if(Object.keys(bn).length!==1){Ie.next=22;break}return Pn(),f.info("\u6CA1\u6709\u9700\u8981\u4FDD\u5B58\u7684\u4FEE\u6539"),Ie.abrupt("return");case 22:return Ie.next=24,_s({body:bn});case 24:Qe=Ie.sent,Ie.next=76;break;case 27:if($n=String((ke==null?void 0:ke.domain)||(te==="obtain"?Ge==null?void 0:Ge.domain:"")||Q.domain||"").trim(),Jt=vt($n),Yt=Number(Q.type)===1,gi=Jt?"http":$n.startsWith("*.")?"dns":Q.challenge||"http",ya=D()(D()({action:"start",name:Q.name,domain:Jt?$n.replace(/^\[([^\]]+)\]$/,"$1"):$n},Q.ca?{ca:Q.ca}:{}),{},{challenge:gi,type:Yt?1:2,secret_id:!Yt&&gi==="dns"?Number(Q.secret_id):0,auto_renew:Yt?0:Number(Q.auto_renew||0)}),!Yt){Ie.next=66;break}if(xt=I.current,xt){Ie.next=56;break}return Mt=te==="renew"?"renew":"obtain",Zt=Mt==="renew"?We==null?void 0:We.id:void 0,Ie.next=39,(Mt==="renew"?Mr:$r)({body:D()(D()({},ya),{},{id:Zt})});case 39:if(nt=Ie.sent,Tn=nt==null?void 0:nt.data,T.current===be){Ie.next=44;break}return Tn!=null&&Tn.session_id&&(Mt==="renew"?Mr:$r)({body:{type:1,action:"cancel",session_id:Tn==null?void 0:Tn.session_id,id:Zt}}).catch(function(){}),Ie.abrupt("return");case 44:if((nt==null?void 0:nt.code)===200){Ie.next=46;break}throw new Error((nt==null?void 0:nt.message)||"\u83B7\u53D6\u9A8C\u8BC1\u4FE1\u606F\u5931\u8D25");case 46:if(!(!(Tn!=null&&Tn.session_id)||!Array.isArray(Tn==null?void 0:Tn.challenges)||!Number.isFinite(Tn==null?void 0:Tn.expires_at))){Ie.next=49;break}throw Tn!=null&&Tn.session_id&&(Mt==="renew"?Mr:$r)({body:{type:1,action:"cancel",session_id:Tn==null?void 0:Tn.session_id,id:Zt}}).catch(function(){}),new Error("\u9A8C\u8BC1\u4FE1\u606F\u65E0\u6548\uFF0C\u8BF7\u91CD\u65B0\u7533\u8BF7");case 49:if(xt={data:Tn,method:Mt,id:Zt},I.current=xt,zt(Tn),(xi=w.current)===null||xi===void 0||xi.hide(),(bi=u.current)===null||bi===void 0||bi.show(),!((Tn==null||(yi=Tn.challenges)===null||yi===void 0?void 0:yi.length)>0)){Ie.next=56;break}return Ie.abrupt("return");case 56:if(!(((Ot=xt)===null||Ot===void 0||(Ot=Ot.data)===null||Ot===void 0?void 0:Ot.expires_at)<=Math.floor(Date.now()/1e3))){Ie.next=61;break}throw xr(),(Fi=u.current)===null||Fi===void 0||Fi.hide(),(wi=w.current)===null||wi===void 0||wi.show(),new Error("\u9A8C\u8BC1\u4F1A\u8BDD\u5DF2\u8FC7\u671F\uFF0C\u8BF7\u91CD\u65B0\u7533\u8BF7");case 61:return Ie.next=63,(xt.method==="renew"?Mr:$r)({body:{type:1,action:"submit",session_id:($t=xt)===null||$t===void 0||($t=$t.data)===null||$t===void 0?void 0:$t.session_id,id:xt.id}});case 63:Qe=Ie.sent,Ie.next=76;break;case 66:if(te!=="obtain"){Ie.next=72;break}return Ie.next=69,$r({body:ya});case 69:Ie.t0=Ie.sent,Ie.next=75;break;case 72:return Ie.next=74,Mr({body:D()(D()({},ya),{},{id:We==null?void 0:We.id})});case 74:Ie.t0=Ie.sent;case 75:Qe=Ie.t0;case 76:if(T.current===be){Ie.next=78;break}return Ie.abrupt("return");case 78:if(((Pe=Qe)===null||Pe===void 0?void 0:Pe.code)===200){Ie.next=81;break}return f.error(((Ai=Qe)===null||Ai===void 0?void 0:Ai.message)||"\u64CD\u4F5C\u5931\u8D25"),Ie.abrupt("return");case 81:I.current=null,$.current=null,P.current=!1,le(!1),(Xe=w.current)===null||Xe===void 0||Xe.hide(),(pn=u.current)===null||pn===void 0||pn.hide();try{Promise.resolve(a==null?void 0:a()).catch(function(){})}catch(st){}if(T.current===be){Ie.next=90;break}return Ie.abrupt("return");case 90:if(!(Ge&&q.current===Ge&&(te==="import"||te==="obtain"))){Ie.next=105;break}if(ea=It((Ei=Qe)===null||Ei===void 0?void 0:Ei.data),cs=[Ge.domain].concat(Je()(ea.domains)).map(function(st){var Fa=String(st).trim().toLowerCase().replace(/\.$/,"").replace(/^\[([^\]]+)\]$/,"$1");try{return new URL("http://".concat(Fa.includes(":")?"[".concat(Fa,"]"):Fa)).hostname.replace(/^\[([^\]]+)\]$/,"$1")}catch(gd){return Fa}}),Ci=yo()(cs),Ht=Ci[0],ps=Ci.slice(1),ji=ps.some(function(st){return st===Ht||st.startsWith("*.")&&!Ht.startsWith("*.")&&!vt(Ht)&&Ht.split(".").length===st.split(".").length&&Ht.endsWith(st.slice(1))}),Si=br()(),Di=br()(String(ea.start_time||"")),ki=br()(String(ea.expire_time||"")),!(ji&&Di.isValid()&&ki.isValid()&&!Di.isAfter(Si)&&ki.isAfter(Si))){Ie.next=102;break}f.success(((Bi=Qe)===null||Bi===void 0?void 0:Bi.message)||"\u64CD\u4F5C\u6210\u529F"),Sr();try{Promise.resolve(Ge.onSelected(ea)).catch(function(){})}catch(st){}return Ie.abrupt("return");case 102:f.warning(ji?"\u8BC1\u4E66\u5DF2\u4FDD\u5B58\uFF0C\u4F46\u4E0D\u5728\u6709\u6548\u671F\u5185\uFF0C\u8BF7\u9009\u62E9\u5176\u4ED6\u8BC1\u4E66\u3002":"\u8BC1\u4E66\u5DF2\u4FDD\u5B58\uFF0C\u4F46\u4E0D\u9002\u7528\u4E8E\u5F53\u524D\u57DF\u540D\uFF0C\u8BF7\u9009\u62E9\u5176\u4ED6\u8BC1\u4E66\u3002"),Ie.next=106;break;case 105:f.success(((Ni=Qe)===null||Ni===void 0?void 0:Ni.message)||"\u64CD\u4F5C\u6210\u529F");case 106:V&&Ze(Qr,ae,En);try{T.current===be&&Promise.resolve(ke==null||(Ti=ke.onObtained)===null||Ti===void 0?void 0:Ti.call(ke,It((_i=Qe)===null||_i===void 0?void 0:_i.data))).catch(function(){})}catch(st){}Ie.next=113;break;case 110:Ie.prev=110,Ie.t1=Ie.catch(7),T.current===be&&f.error(Ie.t1 instanceof Error?Ie.t1.message:"\u64CD\u4F5C\u5931\u8D25");case 113:return Ie.prev=113,T.current===be&&(M.current=!1,pe(!1)),Ie.finish(113);case 116:case"end":return Ie.stop()}},G,null,[[7,110,113,116]])}));return function(G){return O.apply(this,arguments)}})(),[V,En,xr,Sr,Pn,we,Ze,f,te,a,Qr,ae,We]),cr=(0,i.useCallback)((function(){var O=ne()(b()().mark(function G(Q){var be,ke;return b()().wrap(function(Pe){for(;;)switch(Pe.prev=Pe.next){case 0:return Pe.prev=0,Pe.next=3,Rs({body:{id:Q.id}});case 3:if(be=Pe.sent,(be==null?void 0:be.code)===200){Pe.next=7;break}return f.error((be==null?void 0:be.message)||"\u5220\u9664\u8BC1\u4E66\u5931\u8D25"),Pe.abrupt("return");case 7:f.success((be==null?void 0:be.message)||"\u5220\u9664\u6210\u529F"),a==null||a(),ke=jt.length===1&&Qr>1?Qr-1:Qr,Ze(ke,ae,En),Pe.next=16;break;case 13:Pe.prev=13,Pe.t0=Pe.catch(0),f.error("\u5220\u9664\u8BC1\u4E66\u5931\u8D25");case 16:case"end":return Pe.stop()}},G,null,[[0,13]])}));return function(G){return O.apply(this,arguments)}})(),[En,jt.length,Ze,f,a,Qr,ae]),et=(0,i.useCallback)(function(O){y.confirm({title:"\u5220\u9664\u8BC1\u4E66",content:"\u786E\u5B9A\u5220\u9664\u8BC1\u4E66\u201C".concat((O==null?void 0:O.name)||"\u672A\u547D\u540D\u8BC1\u4E66","\u201D\u5417\uFF1F\u88AB\u7F51\u7AD9\u4F7F\u7528\u7684\u8BC1\u4E66\u65E0\u6CD5\u5220\u9664\u3002"),zIndex:1040,okText:"\u5220\u9664",cancelText:"\u53D6\u6D88",okButtonProps:{danger:!0},mask:{closable:!0},onOk:function(){return cr(O)}})},[y,cr]),Xt=(0,i.useCallback)(function(O){return Le?(0,e.jsx)(ge.A,{size:"small",autoInsertSpace:!1,onClick:function(){var Q=q.current;if(Q){var be=br()(),ke=br()(String(O.start_time||"")),Ge=br()(String(O.expire_time||""));if(!ke.isValid()||ke.isAfter(be)||!Ge.isValid()||!Ge.isAfter(be)){f.warning("\u8BC1\u4E66\u4E0D\u5728\u6709\u6548\u671F\u5185\uFF0C\u8BF7\u9009\u62E9\u5176\u4ED6\u8BC1\u4E66\u3002"),Ze();return}Sr();try{Promise.resolve(Q.onSelected(O)).catch(function(){})}catch(Pe){}}},children:"\u90E8\u7F72"}):(0,e.jsx)(gn.A,{menu:{items:[{key:"view",label:"\u67E5\u770B\u8BC1\u4E66"},{key:"edit",label:"\u7F16\u8F91\u8BC1\u4E66"},{key:"renew",label:"\u7EED\u7B7E\u8BC1\u4E66",disabled:![1,2].includes(Number(O.type))},{type:"divider"},{key:"delete",label:"\u5220\u9664\u8BC1\u4E66",danger:!0}],onClick:function(Q){var be=Q.key;be==="view"||be==="edit"?Yn(be,O):be==="renew"&&[1,2].includes(Number(O.type))?Yn("renew",O):be==="delete"&&et(O)}},children:(0,e.jsx)(ge.A,{size:"small",autoInsertSpace:!1,"aria-label":"\u64CD\u4F5C\u8BC1\u4E66 ".concat((O==null?void 0:O.name)||""),children:"\u64CD\u4F5C"})})},[Sr,et,Le,Ze,f,Yn]),fi=(0,i.useMemo)(function(){return[{title:"\u8BC1\u4E66\u540D\u79F0",dataIndex:"name",key:"name",width:160,render:function(G,Q){return Le?(0,e.jsx)("span",{className:"record-name",title:G,children:G||"\u672A\u547D\u540D\u8BC1\u4E66"}):(0,e.jsx)("button",{type:"button",className:"record-name",title:G,onClick:function(){return Yn("view",Q)},children:G||"\u672A\u547D\u540D\u8BC1\u4E66"})}},{title:"\u9002\u7528\u57DF\u540D",dataIndex:"domains",key:"domains",width:190,render:function(G){var Q=G.join(", ")||"\u2014";return(0,e.jsx)(Kn.A,{title:Q,children:(0,e.jsx)("span",{className:"record-domains",children:Q})})}},{title:"\u7B7E\u53D1\u4FE1\u606F",key:"issuance",width:220,render:function(G,Q){var be=Q.secret_id?Ln[Q.secret_id]||"\u5BC6\u94A5 #".concat(Q.secret_id):"\u672A\u5173\u8054\u5BC6\u94A5";return(0,e.jsxs)("div",{className:"record-issuance",children:[(0,e.jsx)("span",{className:"record-type",children:Q.type===2?"\u81EA\u52A8":Q.type===1?"\u624B\u52A8":"\u5BFC\u5165\u8BC1\u4E66"}),(Q.type===1||Q.type===2)&&(0,e.jsx)("span",{className:"record-type",children:Q.challenge==="dns"?"DNS":"HTTP"}),Q.type===2&&Q.challenge==="dns"&&(0,e.jsx)(Kn.A,{title:Q.secret_id?"\u5173\u8054\u5BC6\u94A5\uFF1A".concat(be):be,children:(0,e.jsx)("span",{className:"record-secret",children:be})}),Q.type===2&&(0,e.jsx)("span",{className:"record-renewal",children:Q.auto_renew===1?"\u81EA\u52A8\u7EED\u7B7E":"\u624B\u52A8\u7EED\u7B7E"})]})}},{title:"\u6709\u6548\u671F",key:"expiry",width:145,render:function(G,Q){var be=br()(),ke=br()(String(Q.expire_time||"")),Ge=br()(String(Q.start_time||"")),Pe=ke.isValid(),Xe=Pe&&!ke.isAfter(be),pn=Ge.isValid()&&Ge.isAfter(be),Qe=Pe?ke.diff(be,"day"):0,bn=Pe?Xe?"expired":pn||Qe<=7?"expiring":"valid":"unknown",pr=Pe?Xe?"\u5DF2\u8FC7\u671F":pn?"\u5C1A\u672A\u751F\u6548":Qe>0?"\u5269\u4F59 ".concat(Qe," \u5929"):"\u4E0D\u8DB3 1 \u5929":"\u6709\u6548\u671F\u672A\u77E5";return(0,e.jsx)(Kn.A,{title:Pe?"".concat(String(Q.expire_time).replace("T"," ").slice(0,16)," \u5230\u671F"):void 0,children:(0,e.jsxs)("span",{className:"record-state ".concat(bn),children:[(0,e.jsx)("i",{}),pr]})})}},{title:"\u64CD\u4F5C",key:"action",width:82,fixed:"right",className:"action-cell",render:function(G,Q){return Xt(Q)}}]},[Le,Yn,Xt,Ln]),xa=(0,i.useCallback)(function(O){var G=String(g.getFieldValue(O)||"");if(!G.trim()){f.warning(O==="certificate"?"\u8BC1\u4E66\u5185\u5BB9\u4E3A\u7A7A":"\u79C1\u94A5\u5185\u5BB9\u4E3A\u7A7A");return}var Q=URL.createObjectURL(new Blob([G],{type:"application/x-pem-file"})),be=document.createElement("a");be.href=Q,be.download=O==="certificate"?"fullchain.pem":"private.key",be.click(),window.setTimeout(function(){return URL.revokeObjectURL(Q)},1e3)},[g,f]),cd=(0,i.useCallback)(function(O){if(a==null||a(),V&&Ze(),V||P.current&&te!=="import"){_.current===T.current&&O.action==="create"&&(A.current=O.id);var G=O.action==="update"&&A.current===O.id&&_.current===T.current&&g.getFieldValue("challenge")==="dns"&&Number(g.getFieldValue("type"))===2;G&&g.setFieldValue("secret_id",O.id),(G||O.action==="delete"&&A.current===O.id)&&(A.current=null),en()}},[V,g,Ze,en,te,a]),pd={view:"\u67E5\u770B\u8BC1\u4E66",import:"\u5BFC\u5165\u8BC1\u4E66",edit:"\u7F16\u8F91\u8BC1\u4E66",obtain:"\u7533\u8BF7\u8BC1\u4E66",renew:"\u7EED\u7B7E\u8BC1\u4E66"}[te]||"\u8BC1\u4E66\u914D\u7F6E",ss=(0,e.jsx)(d.A.Item,{extra:te==="edit"?"\u7528\u4E8E\u4E0B\u6B21\u7EED\u7B7E\u9A8C\u8BC1\uFF0C\u4E0D\u6539\u53D8\u5F53\u524D\u8BC1\u4E66\u3002":Dt?mr==="http"?"\u81EA\u884C\u5C06\u9A8C\u8BC1\u6587\u4EF6\u653E\u5230\u76EE\u6807\u7F51\u7AD9\uFF0C\u516C\u7F51\u9700\u80FD\u901A\u8FC7 80 \u7AEF\u53E3\u8BBF\u95EE\u3002":"\u81EA\u884C\u6DFB\u52A0 TXT \u89E3\u6790\u8BB0\u5F55\uFF0C\u65E0\u9700\u63D0\u4F9B\u4E91\u5BC6\u94A5\u3002":ir?"\u516C\u7F51 IP \u7684 80 \u7AEF\u53E3\u9700\u80FD\u8BBF\u95EE\u672C\u673A\u9A8C\u8BC1\u670D\u52A1\u3002":mr==="dns"?"\u81EA\u52A8\u6DFB\u52A0 DNS \u9A8C\u8BC1\u8BB0\u5F55\uFF0C\u7F51\u7AD9\u65E0\u9700\u4E0A\u7EBF\u3002":"\u57DF\u540D\u9700\u89E3\u6790\u5230\u672C\u673A\uFF0C\u516C\u7F51\u53EF\u8BBF\u95EE\u9A8C\u8BC1\u7AEF\u53E3\uFF08\u9ED8\u8BA4 80\uFF09\u3002",children:(0,e.jsxs)(wo.A,{gap:F?12:16,children:[(0,e.jsx)(d.A.Item,{name:"challenge",label:"\u9A8C\u8BC1\u7C7B\u578B",style:{flex:1,minWidth:0,marginBottom:0},rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u9A8C\u8BC1\u7C7B\u578B"}],children:(0,e.jsx)(De.A,{options:[{label:"HTTP",value:"http",disabled:$e},{label:"DNS",value:"dns",disabled:ir}]})}),(0,e.jsx)(d.A.Item,{name:"type",label:"\u9A8C\u8BC1\u6A21\u5F0F",style:{flex:1,minWidth:0,marginBottom:0},rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u9A8C\u8BC1\u6A21\u5F0F"}],children:(0,e.jsx)(De.A,{options:[{label:"\u81EA\u52A8",value:2},{label:"\u624B\u52A8",value:1}]})})]})}),os=(0,e.jsx)(d.A.Item,{preserve:!1,name:"secret_id",label:"\u5173\u8054\u5BC6\u94A5",extra:(0,e.jsxs)(e.Fragment,{children:[Fn?(0,e.jsxs)("span",{children:["\u5BC6\u94A5\u5217\u8868\u52A0\u8F7D\u5931\u8D25\uFF0C",(0,e.jsx)(ge.A,{type:"link",size:"small",htmlType:"button",onClick:function(){en()},children:"\u91CD\u8BD5"})]}):!Oe&&je.length===0?(0,e.jsxs)("span",{children:["\u6682\u65E0\u5BC6\u94A5\uFF0C\u8BF7\u5148",(0,e.jsx)(ge.A,{type:"link",size:"small",htmlType:"button",onClick:function(){return kt(!0)},children:"\u6DFB\u52A0\u5BC6\u94A5"}),"\u5E76\u586B\u5199\u5BC6\u94A5\u4FE1\u606F\u3002"]}):te==="edit"?"DNS \u7EED\u7B7E\u524D\u9700\u5173\u8054\u5BC6\u94A5\u3002":"\u8BF7\u9009\u62E9\u5177\u6709\u8BE5\u57DF\u540D DNS \u89E3\u6790\u6743\u9650\u7684\u5BC6\u94A5\u3002",(0,e.jsx)(ge.A,{type:"link",size:"small",htmlType:"button",onClick:function(){return kt()},children:"\u7BA1\u7406\u5BC6\u94A5"})]}),rules:te==="edit"?[]:[{required:!0,message:"\u8BF7\u9009\u62E9\u4E91\u5BC6\u94A5"}],children:(0,e.jsx)(De.A,{allowClear:te==="edit",loading:Oe,disabled:Oe,showSearch:!1,placeholder:te==="edit"?"\u6682\u4E0D\u5173\u8054\u5BC6\u94A5":"\u8BF7\u9009\u62E9\u4E91\u5BC6\u94A5",options:je,notFoundContent:Oe?(0,e.jsx)(mt.A,{size:"small"}):"\u6682\u65E0\u53EF\u7528\u5BC6\u94A5"})}),ds=(0,e.jsx)(d.A.Item,{name:"auto_renew",label:"\u81EA\u52A8\u7EED\u7B7E",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u662F\u5426\u81EA\u52A8\u7EED\u7B7E"}],children:(0,e.jsx)(De.A,{options:[{label:"\u5173\u95ED",value:0},{label:"\u5F00\u542F",value:1,disabled:te!=="obtain"&&Jn}]})});return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(ce.oK,{ref:N,title:(0,e.jsx)(ce.hE,{children:Le?"\u90E8\u7F72\u8BC1\u4E66":"\u8BC1\u4E66\u7BA1\u7406"}),width:F?800:910,onCancel:Sr,className:v.certificateTable,minHeight:F?440:480,scrollMode:"modal",columns:fi,dataSource:jt,loading:de,emptyContent:ie?(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:"\u83B7\u53D6\u8BC1\u4E66\u5217\u8868\u5931\u8D25",children:(0,e.jsx)(ge.A,{color:"default",variant:"filled",size:"small",onClick:function(){Ze()},children:"\u91CD\u8BD5"})}):void 0,onReload:Ze,rowKey:function(G){return String(G.id)},scroll:{x:907},toolbar:{search:{value:or,field:On,maxLength:100,fields:[{value:"keyword",label:"\u5168\u90E8",placeholder:"\u641C\u7D22\u540D\u79F0\u6216\u57DF\u540D"},{value:"name",label:"\u540D\u79F0",placeholder:"\u641C\u7D22\u8BC1\u4E66\u540D\u79F0"},{value:"domains",label:"\u57DF\u540D",placeholder:"\u641C\u7D22\u8BC1\u4E66\u57DF\u540D",maxLength:253}],ariaLabel:"\u641C\u7D22\u8BC1\u4E66",debounce:V&&!me?300:void 0,onChange:wn,onFieldChange:fr,onSearch:function(G,Q){Ze(1,ae,G,Q)}},refresh:{loading:de,disabled:de,ariaLabel:"\u5237\u65B0\u8BC1\u4E66\u5217\u8868",tooltip:"\u5237\u65B0",onClick:function(){Ze()}},actions:[{key:"import",type:"button",label:"\u5BFC\u5165",icon:"ImportOutlined",onClick:function(){return Yn("import")}},{key:"obtain",type:"button",label:"\u7533\u8BF7",icon:"PlusOutlined",onClick:function(){return Yn("obtain")}},{key:"secret",type:"button",label:"\u5BC6\u94A5",icon:"KeyOutlined",onClick:function(){return kt()}},{key:"log",type:"button",label:"\u65E5\u5FD7",icon:"FileTextOutlined",hidden:!!Le,onClick:function(){var G;return(G=p.current)===null||G===void 0?void 0:G.open({id:"cert-auto-renew",name:"\u8BC1\u4E66\u81EA\u52A8\u7EED\u7B7E"})}}]},pagination:{page:Qr,pageSize:ae,total:an,unit:"\u5F20",disabled:de,onChange:function(G,Q){Ze(Q===ae?G:1,Q,En)}},modalProps:{zIndex:1010,footer:null,closable:!0,keyboard:!0,mask:{closable:!0},style:{top:B.md?100:24,paddingBottom:B.md?100:24,maxWidth:"calc(100vw - 24px)"},focusable:{focusTriggerAfterClose:!1}}}),(0,e.jsx)(ce.Ht,{ref:w,title:(0,e.jsx)(ce.hE,{children:pd}),width:te==="import"||te==="edit"||se&&Gr?F?540:580:se?F?650:700:F?460:500,onOk:se?function(){return Pn()}:g.submit,onCancel:function(){return Pn()},modalProps:{rootClassName:v.modal,zIndex:1030,styles:Bn||te==="obtain"?{container:{minHeight:Yr,display:"flex",flexDirection:"column"},body:{flex:1,display:"flex",flexDirection:"column"}}:void 0,footer:we?null:se?[(0,e.jsx)(ge.A,{onClick:function(){return Pn()},children:"\u5173\u95ED"},"close")]:[(0,e.jsx)(ge.A,{onClick:function(){return Pn()},children:"\u53D6\u6D88"},"cancel"),(0,e.jsx)(ge.A,{type:"primary",loading:tn,disabled:(te==="obtain"||te==="renew")&&jn&&(Oe||Fn),onClick:g.submit,children:te==="import"?"\u5BFC\u5165":te==="obtain"?"\u7533\u8BF7":te==="renew"?"\u7EED\u7B7E":"\u4FDD\u5B58"},"submit")],forceRender:!0,closable:!0,keyboard:!0,mask:{closable:!0},style:{top:B.md?100:24,paddingBottom:B.md?100:24},focusable:{focusTriggerAfterClose:!1},afterClose:function(){P.current||(g.resetFields(),Ar(void 0))}},children:(0,e.jsxs)(d.A,{className:v.formHints,style:Bn||te==="obtain"?{"--certificate-form-min-height":"".concat(Yr-120,"px"),flex:1,display:"flex",flexDirection:"column"}:void 0,form:g,layout:"vertical",variant:"filled",disabled:tn,onFinish:dr,onValuesChange:function(G){if(!(te==="import"||se)){if(te==="obtain"&&Object.prototype.hasOwnProperty.call(G,"name")&&(z.current=""),Object.prototype.hasOwnProperty.call(G,"domain")){var Q=String(G.domain||"").trim(),be=Q.startsWith("*."),ke=vt(Q),Ge=String(g.getFieldValue("challenge")||"http"),Pe=ke?"http":be?"dns":Ge;if(nr(be),Rr(ke),vr(Pe),g.setFieldValue("challenge",Pe),te==="obtain"){var Xe=String(g.getFieldValue("name")||"");(!Xe||Xe===z.current)&&(z.current=Q.slice(0,100),g.setFieldValue("name",z.current))}}if(Object.prototype.hasOwnProperty.call(G,"challenge")){var pn=String(G.challenge||"http"),Qe=ir?"http":$e?"dns":pn;vr(Qe),g.setFieldValue("challenge",Qe)}if(["domain","challenge","type","secret_id"].some(function(Nn){return Object.prototype.hasOwnProperty.call(G,Nn)})){var bn=Number(g.getFieldValue("type"))===2,pr=g.getFieldValue("challenge");(!bn||pr!=="dns")&&g.setFieldValue("secret_id",void 0),!bn||te!=="obtain"&&pr==="dns"&&!g.getFieldValue("secret_id")?g.setFieldValue("auto_renew",0):te==="obtain"&&Object.prototype.hasOwnProperty.call(G,"type")&&g.setFieldValue("auto_renew",1)}}},children:[we&&(0,e.jsx)("div",{className:v.loading,children:(0,e.jsx)(mt.A,{size:"medium"})}),(0,e.jsxs)("div",{className:v.formShell,style:we?{display:"none"}:void 0,children:[Bn&&(0,e.jsxs)(e.Fragment,{children:[se&&(We==null?void 0:We.certificate)&&(0,e.jsx)(ra.A,{className:v.details,size:"small",column:{xs:1,sm:2},colon:!1,styles:{label:{fontSize:"inherit",flexShrink:0,whiteSpace:"nowrap"},content:{fontSize:"inherit",minWidth:0,overflowWrap:"anywhere",whiteSpace:"normal"}},items:[{key:"domains",label:"\u9002\u7528\u57DF\u540D",children:(We.domains||[]).join("\u3001")||"\u6682\u65E0\u57DF\u540D"},{key:"type",label:"\u7B7E\u53D1\u65B9\u5F0F",children:We.type===1?"\u624B\u52A8\u7533\u8BF7":We.type===2?"\u81EA\u52A8\u7533\u8BF7":"\u5BFC\u5165\u8BC1\u4E66"},{key:"start_time",label:"\u751F\u6548\u65F6\u95F4",children:We.start_time?String(We.start_time).replace("T"," ").slice(0,19):"\u672A\u77E5"},{key:"expire_time",label:"\u5230\u671F\u65F6\u95F4",children:We.expire_time?String(We.expire_time).replace("T"," ").slice(0,19):"\u672A\u77E5"}].concat(Je()([1,2].includes(We.type)?[{key:"challenge",label:"\u9A8C\u8BC1\u7C7B\u578B",children:String(We.challenge||"http").toUpperCase()},{key:"auto_renew",label:"\u81EA\u52A8\u7EED\u7B7E",children:We.auto_renew===1?"\u5F00\u542F":"\u5173\u95ED"}].concat(Je()(We.challenge==="dns"?[{key:"secret_id",label:"\u5173\u8054\u5BC6\u94A5",children:We.secret_id?Ln[We.secret_id]||"\u5BC6\u94A5 #".concat(We.secret_id):"\u672A\u5173\u8054"}]:[])):[]))}),(0,e.jsx)("section",{className:v.formSection,children:(0,e.jsx)(d.A.Item,{name:"name",label:"\u8BC1\u4E66\u540D\u79F0",rules:se?[]:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u8BC1\u4E66\u540D\u79F0"},{max:100,message:"\u8BC1\u4E66\u540D\u79F0\u4E0D\u80FD\u8D85\u8FC7 100 \u4E2A\u5B57\u7B26"}],children:(0,e.jsx)(ze.A,{readOnly:se,prefix:(0,e.jsx)(ce.In,{type:"SafetyCertificateOutlined"}),maxLength:100,placeholder:"\u4F8B\u5982\uFF1Aexample.com \u6B63\u5F0F\u8BC1\u4E66",autoComplete:"off"})})}),te==="edit"&&((We==null?void 0:We.type)===1||(We==null?void 0:We.type)===2)&&(0,e.jsxs)("section",{className:v.formSection,children:[ss,jn&&os,gr===2&&ds]}),(0,e.jsx)("section",{className:v.formSection,children:(0,e.jsxs)(yt.A,{className:v.pemGrid,gutter:[16,0],children:[(0,e.jsx)(ct.A,{xs:24,md:12,children:(0,e.jsx)(d.A.Item,{className:v.pemField,name:"certificate",label:(0,e.jsxs)("span",{className:v.fieldLabel,children:[(0,e.jsx)("span",{children:"\u8BC1\u4E66\u5185\u5BB9"}),(se||te==="edit")&&(0,e.jsx)(ge.A,{type:"text",icon:(0,e.jsx)(ce.In,{type:"DownloadOutlined"}),onClick:function(G){G.preventDefault(),xa("certificate")},children:"\u4E0B\u8F7D"})]}),rules:se?[]:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165 PEM \u8BC1\u4E66\u5185\u5BB9"}],children:(0,e.jsx)(ze.A.TextArea,{readOnly:se,className:v.editor,spellCheck:!1,placeholder:`\u7C98\u8D34\u5B8C\u6574\u8BC1\u4E66\u94FE\uFF08PEM \u683C\u5F0F\uFF09
-----BEGIN CERTIFICATE-----`})})}),(0,e.jsx)(ct.A,{xs:24,md:12,children:(0,e.jsx)(d.A.Item,{className:v.pemField,name:"private_key",label:(0,e.jsxs)("span",{className:v.fieldLabel,children:[(0,e.jsx)("span",{children:"\u79C1\u94A5\u5185\u5BB9"}),(se||te==="edit")&&(0,e.jsx)(ge.A,{type:"text",icon:(0,e.jsx)(ce.In,{type:"DownloadOutlined"}),onClick:function(G){G.preventDefault(),xa("private_key")},children:"\u4E0B\u8F7D"})]}),rules:se?[]:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165 PEM \u79C1\u94A5\u5185\u5BB9"}],children:(0,e.jsx)(ze.A.TextArea,{readOnly:se,className:v.editor,spellCheck:!1,placeholder:`\u7C98\u8D34\u4E0E\u8BC1\u4E66\u5339\u914D\u7684\u79C1\u94A5
-----BEGIN PRIVATE KEY-----`})})})]})})]}),(te==="obtain"||te==="renew")&&(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(d.A.Item,{name:"domain",label:"\u57DF\u540D\u6216 IP",extra:ir?"\u516C\u7F51 IP \u4F7F\u7528 HTTP \u9A8C\u8BC1\uFF0C\u8BF7\u7559\u610F\u8BC1\u4E66\u6709\u6548\u671F\u3002":$e?"\u6CDB\u57DF\u540D\u5DF2\u81EA\u52A8\u5207\u6362\u4E3A DNS \u9A8C\u8BC1\u3002":void 0,rules:[{required:!0,whitespace:!0,message:"\u8BF7\u9009\u62E9\u6216\u8F93\u5165\u57DF\u540D\u6216 IP"},{validator:(function(){var O=ne()(b()().mark(function Q(be,ke){var Ge;return b()().wrap(function(Xe){for(;;)switch(Xe.prev=Xe.next){case 0:if(Ge=String(ke||"").trim(),!(Ge&&(/[\[\]:]/.test(Ge)||/^[\d.]+$/.test(Ge))&&!vt(Ge))){Xe.next=3;break}throw new Error("\u8BF7\u8F93\u5165\u6709\u6548 IP \u5730\u5740\uFF0C\u57DF\u540D\u65E0\u9700\u586B\u5199\u7AEF\u53E3");case 3:case"end":return Xe.stop()}},Q)}));function G(Q,be){return O.apply(this,arguments)}return G})()}],children:te==="renew"?(0,e.jsx)(De.A,{options:((We==null?void 0:We.domains)||[]).map(function(O){return{label:O,value:O}}),placeholder:"\u8BF7\u9009\u62E9\u7EED\u7B7E\u57DF\u540D\u6216 IP"}):(0,e.jsx)(ze.A,{readOnly:!!Zn,placeholder:"\u57DF\u540D\u3001\u516C\u7F51 IPv4 \u6216 IPv6",autoComplete:"off"})}),te==="obtain"&&(0,e.jsx)(d.A.Item,{name:"name",label:"\u8BC1\u4E66\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u8BC1\u4E66\u540D\u79F0"},{max:100,message:"\u8BC1\u4E66\u540D\u79F0\u4E0D\u80FD\u8D85\u8FC7 100 \u4E2A\u5B57\u7B26"}],children:(0,e.jsx)(ze.A,{maxLength:100,placeholder:"\u9ED8\u8BA4\u4F7F\u7528\u8BC1\u4E66\u57DF\u540D",autoComplete:"off"})}),ss,jn&&os,gr===2&&ds,(0,e.jsx)(d.A.Item,{name:"ca",label:"\u7B7E\u53D1\u73AF\u5883",extra:te==="renew"?"\u7559\u7A7A\u6CBF\u7528\u539F\u7B7E\u53D1\u73AF\u5883\uFF1B\u6D4B\u8BD5\u8BC1\u4E66\u4E0D\u53D7\u6D4F\u89C8\u5668\u4FE1\u4EFB\u3002":void 0,rules:[{required:te==="obtain"}],children:(0,e.jsx)(De.A,{allowClear:te==="renew",placeholder:te==="renew"?"\u6CBF\u7528\u539F\u7B7E\u53D1\u73AF\u5883":"\u8BF7\u9009\u62E9\u7B7E\u53D1\u73AF\u5883",options:[{label:"\u6B63\u5F0F\u73AF\u5883",value:"production"},{label:"\u6D4B\u8BD5\u73AF\u5883",value:"staging"}]})}),!Dt&&ir&&(0,e.jsx)(Gt.A.Paragraph,{type:"secondary",children:"\u79C1\u6709 IP \u8BF7\u5BFC\u5165\u5305\u542B\u5BF9\u5E94 IP \u5730\u5740\u7684\u8BC1\u4E66\u3002"})]})]})]})}),(0,e.jsx)(ce.Ht,{ref:u,title:(0,e.jsx)(ce.hE,{children:"\u9A8C\u8BC1\u4FE1\u606F"}),width:F?380:400,onOk:g.submit,onCancel:function(){return Pn(!0)},modalProps:{rootClassName:v.modal,zIndex:1030,footer:[(0,e.jsx)(ge.A,{onClick:function(){return Pn(!0)},children:"\u53D6\u6D88"},"cancel"),(0,e.jsx)(ge.A,{type:"primary",loading:tn,onClick:g.submit,children:"\u9A8C\u8BC1"},"submit")],forceRender:!0,closable:!0,keyboard:!0,mask:{closable:!0},style:{top:B.md?100:24,paddingBottom:B.md?100:24},focusable:{focusTriggerAfterClose:!1},afterClose:function(){I.current||zt(null),P.current||(g.resetFields(),Ar(void 0))}},children:Cn&&(0,e.jsx)("div",{className:v.manualShell,"aria-busy":tn,children:(0,e.jsxs)("div",{className:v.manualContent,ref:S,children:[lt.length===0&&(0,e.jsxs)("section",{className:v.verificationRecord,children:[(0,e.jsx)("div",{className:"verification-heading",children:(0,e.jsxs)("div",{className:"verification-heading-text",children:[(0,e.jsx)("span",{className:"verification-title",children:"\u63D0\u4EA4\u9A8C\u8BC1"}),(0,e.jsxs)("span",{className:"verification-expiry",children:["\u6709\u6548\u81F3 ",br().unix(Cn==null?void 0:Cn.expires_at).format("MM-DD HH:mm")]})]})}),(0,e.jsx)(Gt.A.Text,{type:"secondary",children:"\u65E0\u9700\u6DFB\u52A0\u9A8C\u8BC1\u4FE1\u606F\uFF0C\u70B9\u51FB\u201C\u9A8C\u8BC1\u201D\u7EE7\u7EED\u3002"})]}),lt.map(function(O,G){return(0,e.jsxs)("section",{className:v.verificationRecord,children:[O.type==="HTTP"&&(0,e.jsxs)("div",{className:"verification-heading",children:[(0,e.jsxs)("div",{className:"verification-heading-text",children:[(0,e.jsxs)("span",{className:"verification-title",children:["\u90E8\u7F72\u9A8C\u8BC1\u6587\u4EF6",lt.length>1?" ".concat(G+1):""]}),G===0&&(0,e.jsxs)("span",{className:"verification-expiry",children:["\u6709\u6548\u81F3 ",br().unix(Cn==null?void 0:Cn.expires_at).format("MM-DD HH:mm")]})]}),(0,e.jsx)(ge.A,{size:"small",color:"default",variant:"filled",icon:(0,e.jsx)(ce.In,{type:"DownloadOutlined"}),onClick:function(){var be=URL.createObjectURL(new Blob([O.value],{type:"application/octet-stream"})),ke=document.createElement("a");ke.href=be,ke.download=O.name,ke.click(),window.setTimeout(function(){return URL.revokeObjectURL(be)},1e3)},children:"\u4E0B\u8F7D\u6587\u4EF6"})]}),(0,e.jsx)(d.A,{component:"div",layout:"vertical",variant:"filled",className:"verification-fields",children:(O.type==="HTTP"?[{label:"\u6587\u4EF6\u540D",value:O.name},{label:"\u90E8\u7F72\u8DEF\u5F84",value:O.path||""},{label:"\u6587\u4EF6\u5185\u5BB9",value:O.value}]:[{label:"\u8BB0\u5F55\u540D\u79F0",value:O.name},{label:"\u8BB0\u5F55\u503C",value:O.value}]).map(function(Q){var be=Q.label,ke=Q.value;return(0,e.jsx)(d.A.Item,{label:O.type!=="HTTP"&&be==="\u8BB0\u5F55\u540D\u79F0"?(0,e.jsxs)("span",{className:"verification-field-label",children:[(0,e.jsxs)("span",{children:[be,(0,e.jsxs)("span",{className:"verification-type",children:["TXT",lt.length>1?" ".concat(G+1):""]})]}),G===0&&(0,e.jsxs)("span",{className:"verification-expiry",children:["\u6709\u6548\u81F3 ",br().unix(Cn==null?void 0:Cn.expires_at).format("MM-DD HH:mm")]})]}):be,children:(0,e.jsxs)("div",{className:"verification-field-content",children:[(0,e.jsx)(ze.A.TextArea,{value:ke,readOnly:!0,variant:"filled",autoSize:{minRows:1},spellCheck:!1,"aria-label":be}),(0,e.jsx)(Gt.A.Text,{className:"verification-copy",copyable:{text:ke,tooltips:["\u590D\u5236".concat(be),"\u5DF2\u590D\u5236"]}})]})},be)})})]},O.id)})]})})}),(0,e.jsx)(Ho,{ref:x,onChange:a,secretVersion:h}),(0,e.jsx)(sa.A,{ref:p,request:Vo,title:"\u7EED\u7B7E\u65E5\u5FD7"}),(0,e.jsx)(mi,{ref:m,onChange:cd})]})});Pl.displayName="CertManager";var Wo=(0,i.memo)(Pl),zl,Ol,$l,Ml,Hl,Ll,Vl,ql,Wl,Ul,Gl,Ql,Kl,Xl,Jl,Yl,Zl,es,ns,rs,Uo=["styles","title","description","wide","className","children"],Go=["id","action"],Qo="".concat((_t.A===null||_t.A===void 0?void 0:_t.A.basePath)||"/","ace"),ca=[["read_timeout","\u8BFB\u53D6\u8BF7\u6C42\u8D85\u65F6","\u9650\u5236\u8BFB\u53D6\u5B8C\u6574\u8BF7\u6C42\u7684\u6700\u957F\u65F6\u95F4\uFF0C0 \u8868\u793A\u4E0D\u9650\u5236\u3002"],["read_header_timeout","\u8BF7\u6C42\u5934\u8D85\u65F6","\u9632\u6B62\u5BA2\u6237\u7AEF\u957F\u65F6\u95F4\u5360\u7528\u8FDE\u63A5\uFF0C0 \u4F7F\u7528\u9ED8\u8BA4\u503C 10 \u79D2\u3002"],["write_timeout","\u54CD\u5E94\u5199\u5165\u8D85\u65F6","\u9650\u5236\u5B8C\u6574\u54CD\u5E94\u7684\u5199\u5165\u65F6\u95F4\uFF0C0 \u8868\u793A\u4E0D\u9650\u5236\u3002"],["idle_timeout","\u7A7A\u95F2\u8FDE\u63A5\u8D85\u65F6","\u8FDE\u63A5\u590D\u7528\u671F\u95F4\u5141\u8BB8\u4FDD\u6301\u7A7A\u95F2\u7684\u65F6\u95F4\uFF0C0 \u4F7F\u7528\u9ED8\u8BA4\u503C 300 \u79D2\u3002"],["grace_period","\u4F18\u96C5\u5173\u95ED\u65F6\u95F4","\u505C\u6B62\u6216\u91CD\u542F\u65F6\u7B49\u5F85\u73B0\u6709\u8BF7\u6C42\u5B8C\u6210\uFF0C0 \u4F7F\u7528\u9ED8\u8BA4\u503C 10 \u79D2\u3002"]],vi=[["data_path","\u8FD0\u884C\u6570\u636E\u76EE\u5F55","directory","\u4FDD\u5B58 HTTP \u5185\u6838\u8FD0\u884C\u6570\u636E\uFF0C\u7559\u7A7A\u4F7F\u7528 temp/server/data\u3002"],["cache_path","\u54CD\u5E94\u7F13\u5B58\u76EE\u5F55","directory","\u4FDD\u5B58\u7F51\u7AD9\u54CD\u5E94\u7F13\u5B58\uFF0C\u7559\u7A7A\u4F7F\u7528 temp/server/cache\u3002"],["log_path","HTTP \u8FD0\u884C\u65E5\u5FD7","file","\u8BB0\u5F55\u670D\u52A1\u542F\u52A8\u548C\u8FD0\u884C\u5F02\u5E38\uFF1B\u7559\u7A7A\u4F7F\u7528 temp/server/server.log\uFF0C\u586B\u5199 - \u53EF\u5173\u95ED\u3002"],["waf_log_path","WAF \u65E5\u5FD7","file","\u6C47\u603B\u9632\u706B\u5899\u68C0\u6D4B\u4E0E\u62E6\u622A\u4E8B\u4EF6\uFF0C\u7559\u7A7A\u4F7F\u7528 temp/server/waf.log\u3002"],["config_path","\u914D\u7F6E\u5FEB\u7167\u6587\u4EF6","file","\u4FDD\u5B58\u5B9E\u9645\u8FD0\u884C\u914D\u7F6E\u7528\u4E8E\u6392\u969C\uFF0C\u7559\u7A7A\u4F7F\u7528 temp/server/config.json\u3002"]],Ko=[{key:"service",icon:"CloudServerOutlined",label:"\u670D\u52A1\u7BA1\u7406"},{key:"basic",icon:"DashboardOutlined",label:"\u5E38\u7528\u8BBE\u7F6E"},{key:"network",icon:"GlobalOutlined",label:"\u76D1\u542C\u7BA1\u7406"},{key:"runtime",icon:"ControlOutlined",label:"\u8FD0\u884C\u53C2\u6570"},{key:"certificate",icon:"SafetyCertificateOutlined",label:"\u8BC1\u4E66\u7533\u8BF7"},{key:"pages",icon:"FileTextOutlined",label:"\u9ED8\u8BA4\u9875\u9762"}],pa=[{key:"initial_page",label:"\u521D\u59CB\u9875\u9762",description:"\u9759\u6001\u6216 PHP \u7F51\u7AD9\u5C1A\u672A\u90E8\u7F72\u9996\u9875\u65F6\u663E\u793A",status:200},{key:"not_found_page",label:"404 \u9875\u9762",description:"\u7AD9\u70B9\u5185\u8D44\u6E90\u4E0D\u5B58\u5728\u65F6\u663E\u793A",status:404},{key:"site_not_found_page",label:"\u7F51\u7AD9\u4E0D\u5B58\u5728",description:"\u8BBF\u95EE\u7684\u57DF\u540D\u6CA1\u6709\u7ED1\u5B9A\u7F51\u7AD9\u65F6\u663E\u793A",status:404},{key:"site_disabled_page",label:"\u7F51\u7AD9\u505C\u7528",description:"\u8BBF\u95EE\u5DF2\u505C\u7528\u7684\u7F51\u7AD9\u65F6\u663E\u793A",status:503}],Xo=[{key:"start",label:"\u542F\u52A8",icon:"CaretRightOutlined",title:"\u542F\u52A8 HTTP \u670D\u52A1",content:"\u786E\u5B9A\u542F\u52A8 HTTP \u670D\u52A1\u5417\uFF1F",okText:"\u542F\u52A8"},{key:"stop",label:"\u505C\u6B62",icon:"PauseOutlined",title:"\u505C\u6B62 HTTP \u670D\u52A1",content:"\u505C\u6B62\u540E\u6240\u6709\u7F51\u7AD9\u5C06\u6682\u65F6\u65E0\u6CD5\u8BBF\u95EE\uFF0C\u786E\u5B9A\u7EE7\u7EED\u5417\uFF1F",okText:"\u505C\u6B62",danger:!0},{key:"restart",label:"\u91CD\u542F",icon:"ReloadOutlined",title:"\u91CD\u542F HTTP \u670D\u52A1",content:"\u91CD\u542F\u671F\u95F4\u7F51\u7AD9\u4F1A\u77ED\u6682\u4E2D\u65AD\uFF0C\u786E\u5B9A\u7EE7\u7EED\u5417\uFF1F",okText:"\u91CD\u542F"},{key:"sync",label:"\u540C\u6B65",icon:"SyncOutlined",title:"\u540C\u6B65 HTTP \u914D\u7F6E",content:"\u786E\u5B9A\u6839\u636E\u5F53\u524D\u7F51\u7AD9\u914D\u7F6E\u91CD\u65B0\u540C\u6B65 HTTP \u670D\u52A1\u5417\uFF1F",okText:"\u540C\u6B65"}],Jo=[["1xx \u4FE1\u606F\u54CD\u5E94",[[100,"Continue"],[101,"Switching Protocols"],[102,"Processing"],[103,"Early Hints"],[104,"Upload Resumption Supported"]]],["2xx \u6210\u529F\u54CD\u5E94",[[200,"OK"],[201,"Created"],[202,"Accepted"],[203,"Non-Authoritative Information"],[204,"No Content"],[205,"Reset Content"],[206,"Partial Content"],[207,"Multi-Status"],[208,"Already Reported"],[226,"IM Used"]]],["3xx \u91CD\u5B9A\u5411",[[300,"Multiple Choices"],[301,"Moved Permanently"],[302,"Found"],[303,"See Other"],[304,"Not Modified"],[305,"Use Proxy"],[306,"Unused"],[307,"Temporary Redirect"],[308,"Permanent Redirect"]]],["4xx \u5BA2\u6237\u7AEF\u9519\u8BEF",[[400,"Bad Request"],[401,"Unauthorized"],[402,"Payment Required"],[403,"Forbidden"],[404,"Not Found"],[405,"Method Not Allowed"],[406,"Not Acceptable"],[407,"Proxy Authentication Required"],[408,"Request Timeout"],[409,"Conflict"],[410,"Gone"],[411,"Length Required"],[412,"Precondition Failed"],[413,"Content Too Large"],[414,"URI Too Long"],[415,"Unsupported Media Type"],[416,"Range Not Satisfiable"],[417,"Expectation Failed"],[418,"Unused"],[421,"Misdirected Request"],[422,"Unprocessable Content"],[423,"Locked"],[424,"Failed Dependency"],[425,"Too Early"],[426,"Upgrade Required"],[428,"Precondition Required"],[429,"Too Many Requests"],[431,"Request Header Fields Too Large"],[451,"Unavailable For Legal Reasons"]]],["5xx \u670D\u52A1\u7AEF\u9519\u8BEF",[[500,"Internal Server Error"],[501,"Not Implemented"],[502,"Bad Gateway"],[503,"Service Unavailable"],[504,"Gateway Timeout"],[505,"HTTP Version Not Supported"],[506,"Variant Also Negotiates"],[507,"Insufficient Storage"],[508,"Loop Detected"],[510,"Not Extended (Obsoleted)"],[511,"Network Authentication Required"]]]].map(function(s){var n=C()(s,2),t=n[0],r=n[1];return{label:t,options:r.map(function(a){var l=C()(a,2),h=l[0],c=l[1];return{value:h,label:"".concat(h," ").concat(c)}})}}),Et=function(n){return(Array.isArray(n)?n:[]).map(function(t){var r=String(t||"").trim();return/^\d+$/.test(r)?":".concat(r.replace(/^0+(?=\d)/,"")):r})},ts=function(n,t){var r=Et(t).find(function(a){var l=a.match(/^:(\d+)$/),h=Number((l==null?void 0:l[1])||0);return l&&(h<1||h>65535)});return r?Promise.reject(new Error("\u76D1\u542C\u7AEF\u53E3\u8303\u56F4\u4E3A 1-65535")):Promise.resolve()},hi={token:{fontSize:12,fontSizeSM:12},components:{Button:{contentFontSize:12,contentFontSizeSM:12},Input:{inputFontSize:12,inputFontSizeSM:12},Select:{optionFontSize:12}}},Yo=function(n){var t=Number(n||0);return Number.isFinite(t)?t/1e9:0},Zo=function(n){var t=Number(n||0);return Number.isFinite(t)?Math.round(t*1e9):0},ed=function(n){return Object.entries(n&&pi()(n)==="object"?n:{}).flatMap(function(t){var r=C()(t,2),a=r[0],l=r[1];return(Array.isArray(l)?l:[l]).map(function(h){return{name:a,value:String(h||"")}})})},nd=function(n){return(Array.isArray(n)?n:[]).reduce(function(t,r){var a=String((r==null?void 0:r.name)||"").trim(),l=String((r==null?void 0:r.value)||"").trim();return a&&l&&(t[a]||(t[a]=[])).push(l),t},{})},Pt=function(n){var t=n.title,r=n.description,a=n.styles;return(0,e.jsxs)("div",{className:a.sectionHeading,children:[(0,e.jsx)("h3",{children:t}),r&&(0,e.jsx)("p",{children:r})]})},yr=function(n){var t=n.styles,r=n.title,a=n.description,l=n.wide,h=n.className,c=h===void 0?"":h,f=n.children,y=pt()(n,Uo);return(0,e.jsx)(d.A.Item,D()(D()({},y),{},{className:"".concat(t.settingItem," setting-field ").concat(l?"setting-wide":""," ").concat(c).trim(),label:r,extra:a,colon:!1,children:f}))},ma=function(n){var t=n.styles,r=n.title,a=n.description,l=n.wide,h=n.children;return(0,e.jsxs)("div",{className:"".concat(t.settingRow," setting-row ").concat(l?"setting-wide":""),children:[(0,e.jsx)("div",{className:"field-title",children:r}),(0,e.jsx)("div",{className:"setting-control",children:h}),a&&(0,e.jsx)("div",{className:"field-description",children:a})]})},va=function(n){var t=n.title,r=n.description;return(0,e.jsxs)("span",{className:"disclosure-label",children:[(0,e.jsx)("span",{children:t}),r&&(0,e.jsx)("small",{children:r})]})},rd=function(n){var t=n.name,r=n.styles;return(0,e.jsx)(d.A.List,{name:[t,"headers"],children:function(l,h){var c=h.add,f=h.remove;return(0,e.jsxs)("div",{className:r.headerEditor,children:[l.map(function(y){var j=y.key,F=y.name;return(0,e.jsxs)("div",{className:"header-row",children:[(0,e.jsx)(d.A.Item,{name:[F,"name"],rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u54CD\u5E94\u5934\u540D\u79F0"}],children:(0,e.jsx)(ze.A,{placeholder:"\u540D\u79F0",autoComplete:"off"})}),(0,e.jsx)(d.A.Item,{name:[F,"value"],rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u54CD\u5E94\u5934\u5185\u5BB9"}],children:(0,e.jsx)(ze.A,{placeholder:"\u5185\u5BB9",autoComplete:"off"})}),(0,e.jsx)(ge.A,{className:"header-remove",color:"danger",variant:"filled",style:{borderColor:"transparent"},"aria-label":"\u5220\u9664\u54CD\u5E94\u5934",icon:(0,e.jsx)(ce.In,{type:"CloseOutlined"}),onClick:function(){return f(F)}})]},j)}),(0,e.jsx)(ge.A,{color:"default",variant:"filled",style:{borderColor:"transparent"},icon:(0,e.jsx)(ce.In,{type:"PlusOutlined"}),onClick:function(){return c({name:"",value:""})},children:"\u6DFB\u52A0\u54CD\u5E94\u5934"})]})}})},td=(0,kr.rU)(function(s){var n=s.css,t=s.token,r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};return{modal:n(zl||(zl=H()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-container {
            height: `,`px;
            min-height: `,`px;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .ant-modal-header, .ant-modal-footer {
            flex: none;
        }

        .ant-modal-body {
            min-height: 0;
            padding: 0 !important;
            display: flex;
            flex: 1 1 auto;
            overflow: hidden;
        }

        .ant-modal-body > div,
        .ant-modal-body > div > .ant-spin,
        .ant-modal-body > div > .ant-spin > .ant-spin-container,
        .ant-modal-body > div > .ant-spin-nested-loading,
        .ant-modal-body > div > .ant-spin-nested-loading > .ant-spin-container {
            width: 100%;
            min-height: 0;
            display: flex;
            flex: 1 1 auto;
            flex-direction: column;
            overflow: hidden;
        }

        .ant-modal-footer {
            margin-top: 0;
            padding-top: `,`px;
        }
    `])),r!=null&&r.compact?480:520,r!=null&&r.compact?480:520,r!=null&&r.compact?10:12),workbench:n(Ol||(Ol=H()([`
        box-sizing: border-box;
        width: 100%;
        height: auto;
        min-width: 0;
        min-height: 0;
        display: grid;
        flex: 1 1 auto;
        grid-template-columns: `,`px minmax(0, 1fr);
        grid-template-rows: minmax(0, 1fr);
        overflow: hidden;

        &[data-section="service"] {
            min-height: `,`px;
        }

        @media (max-width: 760px) {
            grid-template-columns: minmax(0, 1fr);
            grid-template-rows: auto minmax(0, 1fr);
        }
    `])),r!=null&&r.compact?156:166,r!=null&&r.compact?250:280),navigation:n($l||($l=H()([`
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        max-height: calc(100% - `,`px);
        margin: `,`px 0;
        padding: `,`px;
        display: flex;
        align-self: start;
        overflow: hidden;
        border-radius: `,`px;
        background: `,`;

        .navigation-scroll {
            min-width: 0;
            min-height: 0;
            flex: 1;
            overflow-x: hidden;
            overflow-y: auto;
            scrollbar-color: `,` transparent;
            scrollbar-width: thin;
        }

        .navigation-grid {
            min-width: 0;
            display: flex;
            flex-direction: column;
            gap: 3px;
        }

        .navigation-item {
            width: 100%;
            min-width: 0;
            min-height: `,`px;
            padding: 0 10px;
            display: flex;
            align-items: center;
            gap: 8px;
            flex: none;
            overflow: hidden;
            border: 0;
            border-radius: `,`px;
            outline: 0;
            appearance: none;
            background: transparent;
            color: `,`;
            cursor: pointer;
            font: inherit;
            font-size: `,`px;
            text-align: left;
            transition: background-color .16s ease, color .16s ease;
        }

        .navigation-item:hover {
            background: `,`;
        }

        .navigation-item:focus-visible {
            box-shadow: 0 0 0 2px `,`;
        }

        .navigation-item[data-active="true"] {
            background: color-mix(in srgb, `," 84%, ",`);
            color: `,`;
            font-weight: 600;
        }

        .navigation-item .anticon {
            width: 15px;
            flex: none;
            color: `,`;
            font-size: 15px;
            line-height: 1;
        }

        .navigation-item[data-active="true"] .anticon {
            color: `,`;
        }

        .navigation-label {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (max-width: 760px) {
            width: 100%;
            max-height: none;
            margin: 0;
            padding: 5px 0;
            align-self: stretch;
            overflow: hidden;
            border-radius: 0;
            background: transparent;

            .navigation-scroll {
                width: 100%;
                max-width: 100%;
                flex: 1 1 auto;
                overflow-x: scroll;
                overflow-y: hidden;
                overscroll-behavior-inline: contain;
                scrollbar-width: none;
                -webkit-overflow-scrolling: touch;
                touch-action: pan-x;
            }

            .navigation-scroll::-webkit-scrollbar { display: none; }

            .navigation-grid {
                width: max-content;
                min-width: max-content;
                flex-direction: row;
                gap: 6px;
            }

            .navigation-item {
                width: auto;
                min-width: max-content;
                min-height: `,`px;
                padding: 0 9px;
                gap: 6px;
                background: `,`;
            }
        }
    `])),r!=null&&r.compact?20:24,r!=null&&r.compact?10:12,r!=null&&r.compact?6:7,t.borderRadius,t.colorFillQuaternary,t.colorBorderSecondary,r!=null&&r.compact?34:38,t.borderRadius,t.colorTextSecondary,r!=null&&r.compact?12:13,t.colorFillTertiary,t.colorPrimaryBg,t.colorPrimaryBg,t.colorBgContainer,t.colorTextHeading,t.colorTextTertiary,t.colorPrimary,r!=null&&r.compact?30:34,t.colorFillQuaternary),mainPane:n(Ml||(Ml=H()([`
        min-width: 0;
        min-height: 0;
        height: auto;
        display: flex;
        overflow: hidden;
        container-name: http-service-pane;
        container-type: inline-size;
    `]))),content:n(Hl||(Hl=H()([`
        box-sizing: border-box;
        min-width: 0;
        min-height: 0;
        height: auto;
        padding: `,`;
        flex: 1 1 auto;
        overflow: auto;
        overscroll-behavior: contain;
        scrollbar-color: `,` transparent;
        scrollbar-width: thin;
        -webkit-overflow-scrolling: touch;
        touch-action: pan-y;

        .ant-input-number, .ant-select {
            width: 100%;
        }

        .field-title {
            color: `,`;
            font-size: 13px;
            font-weight: 500;
        }

        .field-description {
            color: `,`;
            font-size: 12px;
            font-weight: 400;
            line-height: 18px;
        }

        .setting-field :where(input, textarea)::placeholder,
        .setting-field .ant-select-placeholder {
            font-size: `,`px !important;
        }

        @media (max-width: 760px) {
            padding: `,`;
        }
    `])),r!=null&&r.compact?"15px 18px 18px":"18px 22px 22px",t.colorBorderSecondary,t.colorText,t.colorTextTertiary,r!=null&&r.compact?12:13,r!=null&&r.compact?"12px 12px 16px":"15px 12px 18px"),contentInner:n(Ll||(Ll=H()([`
        width: 100%;
        max-width: 700px;
        margin: 0 auto;
    `]))),section:n(Vl||(Vl=H()([`
        min-width: 0;
        padding-bottom: 2px;
    `]))),sectionHeading:n(ql||(ql=H()([`
        margin-bottom: `,`px;

        h3 {
            margin: 0;
            color: `,`;
            font-size: `,`px;
            font-weight: 600;
            line-height: 22px;
        }

        p {
            margin: 2px 0 0;
            color: `,`;
            font-size: 12px;
            line-height: 18px;
        }
    `])),r!=null&&r.compact?10:13,t.colorTextHeading,r!=null&&r.compact?14:15,t.colorTextTertiary),settingItem:n(Wl||(Wl=H()([`
        && {
            margin: 0;
            padding: `,`;
        }

        && .ant-form-item-row {
            display: block;
        }

        && .ant-form-item-label {
            padding: 0 0 `,`px;
            text-align: left;
            white-space: normal;
        }

        && .ant-form-item-label > label {
            height: auto;
            color: `,`;
            font-size: 13px;
            font-weight: 500;
            line-height: 20px;
            white-space: normal;
        }

        && .ant-form-item-label > label::after {
            display: none;
        }

        && .ant-form-item-control {
            min-width: 0;
            width: 100%;
            max-width: none;
        }

        && .ant-form-item-control-input {
            width: 100%;
            max-width: none;
        }

        && .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            width: 100%;
            max-width: `,`px;
        }

        &&.setting-wide .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: `,`px;
        }

        && .ant-form-item-explain {
            margin-top: 4px;
            font-size: 11px;
        }

        && .ant-form-item-extra {
            margin-top: `,`px;
            color: `,`;
            font-size: 12px;
            line-height: 18px;
        }
    `])),r!=null&&r.compact?"4px 0":"6px 0",r!=null&&r.compact?5:6,t.colorText,r!=null&&r.compact?240:260,r!=null&&r.compact?300:320,r!=null&&r.compact?4:5,t.colorTextTertiary),settingRow:n(Ul||(Ul=H()([`
        min-width: 0;
        padding: `,`;

        .field-title {
            margin-bottom: `,`px;
            line-height: 20px;
        }

        .setting-control {
            min-width: 0;
            width: 100%;
            max-width: none;
        }

        .field-description {
            margin-top: `,`px;
        }
    `])),r!=null&&r.compact?"7px 0":"9px 0",r!=null&&r.compact?5:6,r!=null&&r.compact?4:5),disclosure:n(Gl||(Gl=H()([`
        margin-top: 2px;
        border: 0 !important;
        background: transparent !important;

        .ant-collapse-item {
            border: 0 !important;
        }

        .ant-collapse-header {
            min-height: 0;
            padding: `,` !important;
            align-items: center !important;
            color: `,` !important;
            border-radius: `,`px !important;
            transition: color .16s ease, background-color .16s ease !important;
        }

        .ant-collapse-header:hover,
        .ant-collapse-header:focus,
        .ant-collapse-header:focus-visible {
            outline: 0 !important;
            background: transparent !important;
            box-shadow: none !important;
        }

        .ant-collapse-item-active > .ant-collapse-header {
            color: `,` !important;
        }

        .ant-collapse-expand-icon {
            color: `,`;
        }

        .ant-collapse-item-active .ant-collapse-expand-icon {
            color: `,`;
        }

        .ant-collapse-body {
            margin: 1px 0 `,`px;
            padding: `,` !important;
            border-radius: `,`px;
            background: `,`;
        }

        .ant-collapse-body .ant-form-item-label {
            padding-bottom: 3px !important;
        }

        .ant-collapse-body .ant-form-item-label > label,
        .ant-collapse-body .field-title {
            font-size: 12px !important;
            line-height: 18px;
        }

        .ant-collapse-body .ant-form-item-extra,
        .ant-collapse-body .field-description {
            margin-top: 3px !important;
            font-size: 11px !important;
            line-height: 16px;
        }

        .ant-collapse-body :where(
            .ant-input,
            .ant-input-number-input,
            .ant-input-prefix,
            .ant-input-suffix,
            .ant-btn,
            .ant-select,
            .ant-select-selection-search-input,
            .ant-select-selection-item,
            .ant-select-selection-placeholder,
            .ant-select-placeholder
        ) {
            font-size: 12px !important;
        }

        .ant-collapse-body :where(input, textarea)::placeholder,
        .ant-collapse-body .ant-select-selection-search-input::placeholder,
        .ant-collapse-body .setting-field .ant-select-placeholder {
            font-size: 12px !important;
        }

        .ant-collapse-body :where(.ant-btn, .ant-input-prefix, .ant-input-suffix, .ant-select-arrow) .anticon,
        .ant-collapse-body > .anticon {
            font-size: 12px !important;
        }

        .ant-collapse-body > .setting-field,
        .ant-collapse-body > .setting-row {
            padding-top: `,`px;
            padding-bottom: `,`px;
        }

        .ant-collapse-body > .setting-field .ant-form-item-control-input-content > :where(
            .ant-input,
            .ant-input-affix-wrapper,
            .ant-input-number,
            .ant-select,
            .ant-space-compact
        ) {
            max-width: `,`px;
        }

        @media (max-width: 760px) {
            .ant-collapse-body > .setting-field .ant-form-item-control-input-content > :where(
                .ant-input,
                .ant-input-affix-wrapper,
                .ant-input-number,
                .ant-select,
                .ant-space-compact
            ) {
                max-width: none;
            }
        }

        .ant-collapse-body > .setting-field:first-child,
        .ant-collapse-body > .setting-row:first-child {
            padding-top: 0;
        }

        .ant-collapse-body > .setting-field:last-child,
        .ant-collapse-body > .setting-row:last-child {
            padding-bottom: 0;
        }

        .disclosure-label {
            min-width: 0;
            display: grid;
            gap: 2px;
            font-size: 13px;
            font-weight: 500;
            line-height: 18px;
        }

        .disclosure-label small {
            color: `,`;
            font-size: 12px;
            font-weight: 400;
            line-height: 18px;
        }
    `])),r!=null&&r.compact?"4px 0":"5px 0",t.colorTextSecondary,t.borderRadiusSM,t.colorText,t.colorTextTertiary,t.colorPrimary,r!=null&&r.compact?3:4,r!=null&&r.compact?"4px 7px":"6px 9px",t.borderRadiusSM,t.colorFillQuaternary,r!=null&&r.compact?2:3,r!=null&&r.compact?2:3,r!=null&&r.compact?230:250,t.colorTextTertiary),serviceActions:n(Ql||(Ql=H()([`
        display: flex;
        flex-wrap: wrap;
        gap: `,`px;

        .service-action.ant-btn {
            width: auto;
            min-width: max-content;
            height: `,`px;
            padding: 0 `,`px;
            gap: 5px;
            font-size: `,`px;
        }
    `])),r!=null&&r.compact?5:6,r!=null&&r.compact?28:32,r!=null&&r.compact?8:10,r!=null&&r.compact?11:12),serviceStatus:n(Kl||(Kl=H()([`
        min-height: `,`px;
        display: inline-flex;
        align-items: center;
        gap: 8px;
        color: `,`;
        font-size: 13px;

        .status-dot {
            width: 7px;
            height: 7px;
            flex: none;
            border-radius: 50%;
            background: `,`;
            box-shadow: 0 0 0 3px color-mix(in srgb, `,`, transparent 84%);
        }

        &[data-running="true"] {
            color: `,`;
        }

        &[data-running="true"] .status-dot {
            background: `,`;
            box-shadow: 0 0 0 3px color-mix(in srgb, `,`, transparent 84%);
        }
    `])),r!=null&&r.compact?28:32,t.colorTextSecondary,t.colorTextQuaternary,t.colorTextQuaternary,t.colorSuccess,t.colorSuccess,t.colorSuccess),infoNote:n(Xl||(Xl=H()([`
        margin-top: `,`px;
        padding: `,`;
        display: flex;
        align-items: flex-start;
        gap: 7px;
        border-left: 2px solid `,`;
        background: `,`;
        color: `,`;
        font-size: 12px;
        line-height: 18px;

        .anticon {
            margin-top: 2px;
            flex: none;
            color: `,`;
        }
    `])),r!=null&&r.compact?10:12,r!=null&&r.compact?"8px 9px":"9px 10px",t.colorPrimary,t.colorFillQuaternary,t.colorTextSecondary,t.colorPrimary),errorState:n(Jl||(Jl=H()([`
        min-height: `,`px;
        height: 100%;
        padding: 24px;
        display: flex;
        align-items: center;
        justify-content: center;

        .ant-empty-description {
            color: `,`;
        }
    `])),r!=null&&r.compact?280:320,t.colorTextSecondary),protocolGroup:n(Yl||(Yl=H()([`
        display: flex;
        flex-wrap: wrap;
        gap: `,`px;

        .ant-checkbox-wrapper {
            min-height: `,`px;
            margin-inline-start: 0;
            padding: 0 `,`px;
            display: inline-flex;
            align-items: center;
            border-radius: `,`px;
            background: `,`;
            color: `,`;
            font-size: 12px;
            transition: background-color .16s ease, color .16s ease;
        }

        .ant-checkbox-wrapper:hover {
            background: `,`;
            color: `,`;
        }

        .ant-checkbox-wrapper-checked {
            background: `,`;
            color: `,`;
        }

        .ant-checkbox-wrapper-checked:hover {
            background: `,`;
        }
    `])),r!=null&&r.compact?5:6,r!=null&&r.compact?28:32,r!=null&&r.compact?8:10,t.borderRadiusSM,t.colorFillQuaternary,t.colorTextSecondary,t.colorFillSecondary,t.colorText,t.colorPrimaryBg,t.colorText,t.colorPrimaryBgHover),pageSwitcher:n(Zl||(Zl=H()([`
        width: 100%;
        margin-bottom: `,`px;

        .ant-segmented-item-label {
            padding-inline: `,`px;
            font-size: 12px;
        }

        @media (max-width: 760px) {
            overflow-x: auto;
            scrollbar-width: none;
            -webkit-overflow-scrolling: touch;

            &::-webkit-scrollbar { display: none; }

            &.ant-segmented-block .ant-segmented-group {
                width: max-content;
                min-width: 100%;
            }

            &.ant-segmented-block .ant-segmented-item {
                min-width: max-content;
                flex: 1 0 auto;
            }
        }
    `])),r!=null&&r.compact?10:12,r!=null&&r.compact?8:10),pageEditor:n(es||(es=H()([`
        min-width: 0;

        .editor-field {
            margin: `,`;
        }

        .editor-label {
            margin-bottom: 7px;
            display: grid;
            gap: 2px;
        }

        .editor-label strong {
            color: `,`;
            font-size: 13px;
            font-weight: 500;
        }

        .editor-label span {
            color: `,`;
            font-size: 12px;
            line-height: 18px;
        }

        .editor-field > .ant-form-item {
            margin: 0;
        }
    `])),r!=null&&r.compact?"10px 0":"12px 0",t.colorText,t.colorTextTertiary),headerEditor:n(ns||(ns=H()([`
        --header-control-height: 26px;
        display: grid;
        gap: 6px;

        .header-row {
            display: grid;
            grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) var(--header-control-height);
            gap: 6px;
            align-items: start;
        }

        .header-row .ant-form-item {
            min-width: 0;
            margin: 0;
            padding: 0;
        }

        .header-row .ant-form-item-control-input {
            min-height: var(--header-control-height);
        }

        .header-row .ant-input {
            height: var(--header-control-height);
            padding-top: 0;
            padding-bottom: 0;
        }

        .header-remove.ant-btn {
            width: var(--header-control-height);
            height: var(--header-control-height);
            min-width: 0;
            padding: 0;
            align-self: start;
        }

        @container http-service-pane (max-width: 360px) {
            .header-row {
                grid-template-columns: minmax(0, 1fr) var(--header-control-height);
            }

            .header-row .ant-form-item:nth-child(2) {
                grid-column: 1 / -1;
                grid-row: 2;
            }

            .header-remove {
                grid-column: 2;
                grid-row: 1;
            }
        }
    `]))),code:n(rs||(rs=H()([`
        overflow: hidden;
        border: 1px solid `,`;
        border-radius: `,`px;

        .ace_editor {
            border-radius: `,`px !important;
        }
    `])),t.colorBorderSecondary,t.borderRadius,t.borderRadius)}}),as=(0,i.forwardRef)(function(s,n){var t,r,a,l=un.A.useApp(),h=l.message,c=l.modal,f=(0,ce.DP)(),y=!!(!(f==null||(t=f.isCompactTheme)===null||t===void 0)&&t.call(f)),j=!!(!(f==null||(r=f.isDarkMode)===null||r===void 0)&&r.call(f)||!(f==null||(a=f.isDarkTheme)===null||a===void 0)&&a.call(f)),F=Sn.Ay.useBreakpoint(),B=td({compact:y}),o=B.styles,v=(0,i.useRef)(null),N=(0,i.useRef)(null),w=(0,i.useRef)(null),u=(0,i.useRef)(0),S=d.A.useForm(),x=C()(S,1),m=x[0],p=(0,i.useState)(!1),_=C()(p,2),A=_[0],E=_[1],k=(0,i.useState)(!1),T=C()(k,2),U=T[0],P=T[1],M=(0,i.useState)(!1),I=C()(M,2),$=I[0],q=I[1],z=(0,i.useState)(!1),he=C()(z,2),R=he[0],g=he[1],L=(0,i.useState)(""),J=C()(L,2),V=J[0],ee=J[1],fe=(0,i.useState)(""),Y=C()(fe,2),de=Y[0],xe=Y[1],ye=(0,i.useState)([]),Ce=C()(ye,2),ie=Ce[0],Be=Ce[1],Z=(0,i.useState)("service"),cn=C()(Z,2),Le=cn[0],Fe=cn[1],nn=(0,i.useState)("initial_page"),Ue=C()(nn,2),tn=Ue[0],pe=Ue[1],Se=(0,i.useState)([]),Ee=C()(Se,2),we=Ee[0],ue=Ee[1],K=(0,i.useState)([]),ve=C()(K,2),me=ve[0],le=ve[1],Ae=(0,i.useState)([]),re=C()(Ae,2),te=re[0],mn=re[1],se=(0,i.useCallback)(function(){var W=arguments.length>0&&arguments[0]!==void 0?arguments[0]:{},je=W.body||{},rn=je.id,Me=je.action,Ve=pt()(je,Go);return(0,it.Ez)(D()(D()({},W),{},{body:D()(D()({},Ve),{},{action:"log",log_action:Me==="clear"?"clear":"read"})}))},[]),Bn=(0,i.useCallback)(function(W){var je=D()({},W||{});ca.forEach(function(Me){var Ve=C()(Me,1),Oe=Ve[0];je[Oe]=Yo(W==null?void 0:W[Oe])}),vi.forEach(function(Me){var Ve,Oe=C()(Me,1),sn=Oe[0];je[sn]=String((Ve=W==null?void 0:W[sn])!==null&&Ve!==void 0?Ve:"")}),je.http_listen=Et(W==null?void 0:W.http_listen),je.https_listen=Et(W==null?void 0:W.https_listen);var rn=Number((W==null?void 0:W.max_header_bytes)||0);je.max_header_bytes=Number.isFinite(rn)?rn/1024:0,pa.forEach(function(Me){var Ve,Oe,sn,Re=Me.key,lr=Me.status;je[Re]={status:Number((W==null||(Ve=W[Re])===null||Ve===void 0?void 0:Ve.status)||lr),body:String((W==null||(Oe=W[Re])===null||Oe===void 0?void 0:Oe.body)||""),headers:ed(W==null||(sn=W[Re])===null||sn===void 0?void 0:sn.headers)}}),je.default_site=W!=null&&W.default_site?String(W.default_site):void 0,m.setFieldsValue(je)},[m]),In=(0,i.useCallback)(ne()(b()().mark(function W(){var je,rn,Me,Ve,Oe,sn,Re,lr,Fn,er,Qn=arguments;return b()().wrap(function($e){for(;;)switch($e.prev=$e.next){case 0:return je=Qn.length>0&&Qn[0]!==void 0?Qn[0]:!0,rn=Qn.length>1&&Qn[1]!==void 0?Qn[1]:!1,Me=++u.current,je&&P(!0),xe(""),$e.prev=5,$e.next=8,Promise.all([(0,it.Ez)({body:{action:"get"}}),rn?(0,it.GV)({body:{name:"site"}}):Promise.resolve(null)]);case 8:if(Ve=$e.sent,Oe=C()(Ve,2),sn=Oe[0],Re=Oe[1],u.current===Me){$e.next=14;break}return $e.abrupt("return");case 14:if((sn==null?void 0:sn.code)===200){$e.next=19;break}return lr=(sn==null?void 0:sn.message)||"\u83B7\u53D6 HTTP \u670D\u52A1\u914D\u7F6E\u5931\u8D25",xe(lr),h.error(lr),$e.abrupt("return");case 19:Fn=(sn==null?void 0:sn.data)||{},g(!!(Fn!=null&&Fn.running)),rn&&((Re==null?void 0:Re.code)===200?Be(Object.entries((Re==null||(er=Re.data)===null||er===void 0?void 0:er.name)||{}).map(function(nr){var Xr=C()(nr,2),Jr=Xr[0],ir=Xr[1];return{label:String(ir),value:String(Jr)}}).sort(function(nr,Xr){return nr.label.localeCompare(Xr.label,"zh-CN")})):(Be([]),h.error((Re==null?void 0:Re.message)||"\u83B7\u53D6\u9ED8\u8BA4\u7AD9\u70B9\u9009\u9879\u5931\u8D25"))),Bn((Fn==null?void 0:Fn.config)||{}),$e.next=28;break;case 25:$e.prev=25,$e.t0=$e.catch(5),u.current===Me&&(xe("\u65E0\u6CD5\u8FDE\u63A5\u670D\u52A1\u5668\uFF0C\u8BF7\u68C0\u67E5\u7F51\u7EDC\u540E\u91CD\u8BD5"),h.error("\u83B7\u53D6 HTTP \u670D\u52A1\u914D\u7F6E\u5931\u8D25"));case 28:return $e.prev=28,u.current===Me&&je&&P(!1),$e.finish(28);case 31:case"end":return $e.stop()}},W,null,[[5,25,28,31]])})),[Bn,h]),Vn=(0,i.useCallback)(function(){var W;u.current+=1,E(!0),P(!0),ee(""),xe(""),Fe("service"),pe("initial_page"),ue([]),le([]),mn([]),(W=v.current)===null||W===void 0||W.show()},[]);(0,i.useEffect)(function(){A&&(m.resetFields(),In(!0,!0))},[A,m,In]),(0,i.useEffect)(function(){var W=w.current;if(!(!A||!W)){var je=requestAnimationFrame(function(){var Ve=W.querySelector('[data-section="'.concat(Le,'"]'));Ve==null||Ve.scrollIntoView({behavior:"smooth",block:"nearest",inline:"nearest"})}),rn=function(Oe){var sn=Math.max(0,W.scrollWidth-W.clientWidth);if(!(sn<=0)){var Re=Math.abs(Oe.deltaX)>Math.abs(Oe.deltaY)?Oe.deltaX:Oe.deltaY,lr=Re<0?W.scrollLeft>0:W.scrollLeft<sn;lr&&(Oe.preventDefault(),W.scrollLeft+=Re)}},Me=new ResizeObserver(function(){cancelAnimationFrame(je),je=requestAnimationFrame(function(){var Ve=W.querySelector('[data-section="'.concat(Le,'"]'));Ve==null||Ve.scrollIntoView({behavior:"auto",block:"nearest",inline:"nearest"})})});return W.addEventListener("wheel",rn,{passive:!1}),Me.observe(W),function(){cancelAnimationFrame(je),Me.disconnect(),W.removeEventListener("wheel",rn)}}},[A,Le]),(0,i.useImperativeHandle)(n,function(){return{open:Vn,refresh:function(){return A?In(!1,!1):Promise.resolve()}}},[A,In,Vn]);var Zn=(0,i.useCallback)((function(){var W=ne()(b()().mark(function je(rn){var Me,Ve,Oe;return b()().wrap(function(Re){for(;;)switch(Re.prev=Re.next){case 0:if(!V){Re.next=2;break}return Re.abrupt("return");case 2:return Me=++u.current,ee(rn),Re.prev=4,Re.next=7,(0,it.Ez)({body:{action:rn}});case 7:if(Oe=Re.sent,u.current===Me){Re.next=10;break}return Re.abrupt("return");case 10:if((Oe==null?void 0:Oe.code)===200){Re.next=13;break}return h.error((Oe==null?void 0:Oe.message)||"\u64CD\u4F5C\u5931\u8D25"),Re.abrupt("return");case 13:g(!!(!(Oe==null||(Ve=Oe.data)===null||Ve===void 0)&&Ve.running)),h.success((Oe==null?void 0:Oe.message)||"\u64CD\u4F5C\u6210\u529F"),Re.next=20;break;case 17:Re.prev=17,Re.t0=Re.catch(4),u.current===Me&&h.error("\u64CD\u4F5C\u5931\u8D25");case 20:return Re.prev=20,u.current===Me&&ee(""),Re.finish(20);case 23:case"end":return Re.stop()}},je,null,[[4,17,20,23]])}));return function(je){return W.apply(this,arguments)}})(),[V,h]),Fr=(0,i.useCallback)((function(){var W=ne()(b()().mark(function je(rn){var Me,Ve,Oe,sn,Re;return b()().wrap(function(Fn){for(;;)switch(Fn.prev=Fn.next){case 0:return Me=D()(D()({},m.getFieldsValue(!0)),rn),Ve=D()(D()({},Me),{},{default_site:String(Me.default_site||"")}),pa.forEach(function(er){var Qn,sr,$e,nr=er.key,Xr=er.status;Ve[nr]={status:Number(((Qn=Me[nr])===null||Qn===void 0?void 0:Qn.status)||Xr),body:String(((sr=Me[nr])===null||sr===void 0?void 0:sr.body)||""),headers:nd(($e=Me[nr])===null||$e===void 0?void 0:$e.headers)}}),ca.forEach(function(er){var Qn=C()(er,1),sr=Qn[0];Ve[sr]=Zo(Me[sr])}),vi.forEach(function(er){var Qn,sr=C()(er,1),$e=sr[0];Ve[$e]=String((Qn=Me[$e])!==null&&Qn!==void 0?Qn:"")}),Ve.http_listen=Et(Me.http_listen),Ve.https_listen=Et(Me.https_listen),Ve.max_header_bytes=Math.round(Number(Me.max_header_bytes||0)*1024),Oe=++u.current,q(!0),Fn.prev=10,Fn.next=13,(0,it.Ez)({body:{action:"set",config:Ve}});case 13:if(Re=Fn.sent,u.current===Oe){Fn.next=16;break}return Fn.abrupt("return");case 16:if((Re==null?void 0:Re.code)===200){Fn.next=19;break}return h.error((Re==null?void 0:Re.message)||"\u4FDD\u5B58 HTTP \u670D\u52A1\u914D\u7F6E\u5931\u8D25"),Fn.abrupt("return");case 19:g(!!(!(Re==null||(sn=Re.data)===null||sn===void 0)&&sn.running)),h.success((Re==null?void 0:Re.message)||"\u4FDD\u5B58\u6210\u529F"),Fn.next=26;break;case 23:Fn.prev=23,Fn.t0=Fn.catch(10),u.current===Oe&&h.error("\u4FDD\u5B58 HTTP \u670D\u52A1\u914D\u7F6E\u5931\u8D25");case 26:return Fn.prev=26,u.current===Oe&&q(!1),Fn.finish(26);case 29:case"end":return Fn.stop()}},je,null,[[10,23,26,29]])}));return function(je){return W.apply(this,arguments)}})(),[m,h]),ut=(0,i.useCallback)(function(){var W;u.current+=1,(W=v.current)===null||W===void 0||W.hide()},[]),wr=(0,i.useCallback)(function(W,je){var rn=Array.isArray(je)?je:[];if(rn.length===0)return Promise.reject(new Error("\u8BF7\u81F3\u5C11\u542F\u7528\u4E00\u79CD HTTP \u534F\u8BAE"));if((rn.includes("h2")||rn.includes("h2c"))&&!rn.includes("h1"))return Promise.reject(new Error("\u542F\u7528 HTTP/2 \u6216 h2c \u65F6\u5FC5\u987B\u540C\u65F6\u542F\u7528 HTTP/1.1"));var Me=m.getFieldValue("http_listen");return Array.isArray(Me)&&Me.length>0&&!rn.some(function(Ve){return Ve==="h1"||Ve==="h2c"})?Promise.reject(new Error("\u542F\u7528 HTTP \u76D1\u542C\u65F6\u81F3\u5C11\u9700\u8981 HTTP/1.1 \u6216 HTTP/2 \u660E\u6587")):Promise.resolve()},[m]),mr=(0,i.useCallback)(function(W,je){var rn=m.getFieldValue("protocols");return Array.isArray(je)&&je.length>0&&(!Array.isArray(rn)||!rn.some(function(Me){return Me==="h1"||Me==="h2c"}))?Promise.reject(new Error("HTTP \u76D1\u542C\u81F3\u5C11\u9700\u8981\u542F\u7528 HTTP/1.1 \u6216 HTTP/2 \u660E\u6587")):Promise.resolve()},[m]),vr=Le==="service"&&(0,e.jsxs)("div",{className:o.section,children:[(0,e.jsx)(Pt,{styles:o,title:"\u670D\u52A1\u7BA1\u7406",description:"\u7BA1\u7406\u627F\u8F7D\u5168\u90E8\u7F51\u7AD9\u7684 HTTP \u670D\u52A1\uFF0C\u64CD\u4F5C\u524D\u4F1A\u518D\u6B21\u786E\u8BA4\u3002"}),(0,e.jsx)(ma,{styles:o,title:"\u8FD0\u884C\u72B6\u6001",description:"HTTP \u670D\u52A1\u8D1F\u8D23\u7F51\u7AD9\u76D1\u542C\u3001\u8F6C\u53D1\u3001\u7F13\u5B58\u548C\u9632\u706B\u5899\u914D\u7F6E\u3002",children:(0,e.jsxs)("span",{className:o.serviceStatus,"data-running":R,children:[(0,e.jsx)("i",{className:"status-dot"}),R?"\u8FD0\u884C\u4E2D":"\u5DF2\u505C\u6B62"]})}),(0,e.jsx)(ma,{styles:o,wide:!0,title:"\u670D\u52A1\u64CD\u4F5C",description:"\u91CD\u542F\u4F1A\u77ED\u6682\u4E2D\u65AD\u8BBF\u95EE\uFF1B\u540C\u6B65\u4EC5\u91CD\u65B0\u52A0\u8F7D\u5F53\u524D\u7F51\u7AD9\u914D\u7F6E\u3002",children:(0,e.jsx)("div",{className:o.serviceActions,role:"toolbar","aria-label":"HTTP \u670D\u52A1\u64CD\u4F5C",children:Xo.map(function(W){return(0,e.jsx)(ge.A,{className:"service-action",color:W.key==="stop"?"danger":"default",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",icon:(0,e.jsx)(ce.In,{type:W.icon}),disabled:$||!!V||(W.key==="start"?R:W.key==="stop"?!R:!1),loading:V===W.key,onClick:function(){return c.confirm({title:W.title,content:W.content,okText:W.okText,cancelText:"\u53D6\u6D88",okButtonProps:W.danger?{danger:!0,type:"default"}:void 0,mask:{closable:!0},onOk:function(){return Zn(W.key)}})},children:W.label},W.key)})})}),(0,e.jsx)(ma,{styles:o,title:"\u8FD0\u884C\u65E5\u5FD7",description:"\u67E5\u770B\u670D\u52A1\u542F\u52A8\u3001\u914D\u7F6E\u52A0\u8F7D\u548C\u8FD0\u884C\u5F02\u5E38\u8BB0\u5F55\u3002",children:(0,e.jsx)("div",{className:o.serviceActions,children:(0,e.jsx)(ge.A,{className:"service-action",color:"default",variant:"filled",style:{borderColor:"transparent"},htmlType:"button",icon:(0,e.jsx)(ce.In,{type:"FileTextOutlined"}),disabled:!!V,onClick:function(){var je;return(je=N.current)===null||je===void 0?void 0:je.open({id:"server",name:"HTTP \u670D\u52A1"})},children:"\u65E5\u5FD7"})})})]}),Wr=Le==="basic"&&(0,e.jsxs)("div",{className:o.section,children:[(0,e.jsx)(Pt,{styles:o,title:"\u5E38\u7528\u8BBE\u7F6E",description:"\u8FD9\u4E9B\u914D\u7F6E\u4FDD\u5B58\u540E\u4F1A\u7ACB\u5373\u5E94\u7528\u5230\u5F53\u524D HTTP \u670D\u52A1\u3002"}),(0,e.jsx)(yr,{styles:o,name:"default_site",title:"\u9ED8\u8BA4\u7AD9\u70B9",description:"\u57DF\u540D\u672A\u7ED1\u5B9A\u7F51\u7AD9\u65F6\u8F6C\u53D1\u5230\u8FD9\u91CC\uFF0C\u7559\u7A7A\u663E\u793A\u201C\u7F51\u7AD9\u4E0D\u5B58\u5728\u201D\u9875\u9762\u3002",children:(0,e.jsx)(De.A,{allowClear:!0,showSearch:!1,options:ie,placeholder:"\u8BF7\u9009\u62E9\u9ED8\u8BA4\u7AD9\u70B9\uFF0C\u7559\u7A7A\u8868\u793A\u5173\u95ED"})}),(0,e.jsx)(yr,{styles:o,name:"protocols",title:"\u542F\u7528\u534F\u8BAE",description:"HTTP/1.1 \u517C\u5BB9\u6027\u6700\u597D\uFF1BHTTP/2 \u548C HTTP/3 \u53EF\u63D0\u5347\u5E76\u53D1\u4F20\u8F93\u6548\u7387\u3002",rules:[{validator:wr}],children:(0,e.jsx)(la.A.Group,{className:o.protocolGroup,options:[{label:"HTTP/1.1",value:"h1"},{label:"HTTP/2",value:"h2"},{label:"HTTP/2 \u660E\u6587",value:"h2c"},{label:"HTTP/3",value:"h3"}]})}),(0,e.jsx)(yr,{styles:o,name:"log_level",title:"\u65E5\u5FD7\u7EA7\u522B",description:"\u53EA\u63A7\u5236 HTTP \u670D\u52A1\u8FD0\u884C\u65E5\u5FD7\uFF1B\u65E5\u5E38\u4F7F\u7528 INFO\uFF0CDEBUG \u4EC5\u7528\u4E8E\u6392\u969C\u3002",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u65E5\u5FD7\u7EA7\u522B"}],children:(0,e.jsx)(De.A,{options:["DEBUG","INFO","WARN","ERROR","PANIC","FATAL"].map(function(W){return{label:W,value:W}}),placeholder:"\u8BF7\u9009\u62E9\u65E5\u5FD7\u7EA7\u522B"})})]}),Ur=Le==="network"&&(0,e.jsxs)("div",{className:o.section,children:[(0,e.jsx)(Pt,{styles:o,title:"\u76D1\u542C\u7BA1\u7406",description:"\u914D\u7F6E\u7F51\u7AD9\u5BF9\u5916\u63D0\u4F9B\u670D\u52A1\u7684\u5730\u5740\uFF0C\u4EE5\u53CA\u7ECF\u8FC7 CDN \u6216\u53CD\u5411\u4EE3\u7406\u65F6\u7684\u5BA2\u6237\u7AEF IP \u8BC6\u522B\u3002"}),(0,e.jsx)(yr,{styles:o,wide:!0,name:"http_listen",title:"HTTP \u76D1\u542C\u5730\u5740",description:"\u652F\u6301\u591A\u4E2A IP:\u7AEF\u53E3\uFF0C\u4F8B\u5982 :80\uFF1B\u5220\u9664\u5168\u90E8\u5730\u5740\u4F1A\u5173\u95ED HTTP \u76D1\u542C\u3002",normalize:Et,rules:[{validator:ts},{validator:mr}],children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],open:!1,placeholder:"\u8BF7\u8F93\u5165\u76D1\u542C\u5730\u5740\uFF0C\u4F8B\u5982 :80"})}),(0,e.jsx)(yr,{styles:o,wide:!0,name:"https_listen",title:"HTTPS \u76D1\u542C\u5730\u5740",description:"\u652F\u6301\u591A\u4E2A IP:\u7AEF\u53E3\uFF0C\u4F8B\u5982 :443\uFF1B\u7F51\u7AD9\u9700\u8981\u7ED1\u5B9A\u8BC1\u4E66\u624D\u80FD\u6B63\u5E38\u63D0\u4F9B HTTPS\u3002",normalize:Et,rules:[{validator:ts}],children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],open:!1,placeholder:"\u8BF7\u8F93\u5165\u76D1\u542C\u5730\u5740\uFF0C\u4F8B\u5982 :443"})}),(0,e.jsx)(Xn.A,{ghost:!0,className:o.disclosure,expandIconPlacement:"end",activeKey:we,onChange:function(je){return ue((Array.isArray(je)?je:[je]).map(String))},items:[{key:"proxy",label:(0,e.jsx)(va,{title:"\u53EF\u4FE1\u4EE3\u7406\uFF08\u9AD8\u7EA7\uFF09",description:"\u4EC5\u5728\u670D\u52A1\u5668\u524D\u65B9\u4F7F\u7528 CDN\u3001\u8D1F\u8F7D\u5747\u8861\u6216\u53CD\u5411\u4EE3\u7406\u65F6\u914D\u7F6E"}),children:(0,e.jsxs)(da.Ay,{componentSize:"small",theme:hi,children:[(0,e.jsx)(yr,{styles:o,wide:!0,name:"trusted_proxies",title:"\u53EF\u4FE1\u4EE3\u7406",description:"\u53EA\u586B\u5199\u4F60\u80FD\u63A7\u5236\u7684\u4EE3\u7406\u51FA\u53E3 IP \u6216 CIDR\uFF0C\u914D\u7F6E\u9519\u8BEF\u53EF\u80FD\u5BFC\u81F4\u5BA2\u6237\u7AEF IP \u88AB\u4F2A\u9020\u3002",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],open:!1,placeholder:"\u8BF7\u8F93\u5165 IP \u6216 CIDR\uFF0C\u4F8B\u5982 127.0.0.1/32"})}),(0,e.jsx)(yr,{styles:o,wide:!0,name:"client_ip_headers",title:"\u5BA2\u6237\u7AEF IP \u8BF7\u6C42\u5934",description:"\u6309\u987A\u5E8F\u8BFB\u53D6\u4EE3\u7406\u5199\u5165\u7684\u771F\u5B9E IP \u8BF7\u6C42\u5934\uFF0C\u5E38\u7528 X-Forwarded-For\u3002",children:(0,e.jsx)(De.A,{mode:"tags",tokenSeparators:[","," "],open:!1,placeholder:"\u8BF7\u8F93\u5165\u8BF7\u6C42\u5934\uFF0C\u4F8B\u5982 X-Forwarded-For"})}),(0,e.jsx)(ma,{styles:o,title:"\u4E25\u683C\u89E3\u6790\u4EE3\u7406\u94FE",description:"\u4ECE\u4EE3\u7406\u94FE\u53F3\u4FA7\u5F00\u59CB\u89E3\u6790\uFF0C\u53EA\u4FE1\u4EFB\u8FDE\u7EED\u7ECF\u8FC7\u53EF\u4FE1\u4EE3\u7406\u7684\u5730\u5740\u3002",children:(0,e.jsx)(d.A.Item,{name:"trusted_proxies_strict",valuePropName:"checked",noStyle:!0,children:(0,e.jsx)(Dr.A,{})})})]})}]})]}),Cn=Le==="runtime"&&(0,e.jsxs)("div",{className:o.section,children:[(0,e.jsx)(Pt,{styles:o,title:"\u8FD0\u884C\u53C2\u6570",description:"\u9ED8\u8BA4\u503C\u9002\u5408\u5927\u591A\u6570\u7F51\u7AD9\uFF0C\u4EC5\u5728\u660E\u786E\u9047\u5230\u8D85\u65F6\u6216\u8BF7\u6C42\u5934\u9650\u5236\u65F6\u8C03\u6574\u3002"}),(0,e.jsx)(yr,{styles:o,name:"max_header_bytes",title:"\u8BF7\u6C42\u5934\u5927\u5C0F",description:"\u5355\u4E2A\u8BF7\u6C42\u5141\u8BB8\u7684\u8BF7\u6C42\u5934\u4E0A\u9650\uFF0C\u9ED8\u8BA4 1024 KB\u3002",rules:[{type:"number",min:0,message:"\u4E0D\u80FD\u5C0F\u4E8E 0"}],children:(0,e.jsx)(qe.A,{min:0,precision:0,suffix:"KB",placeholder:"\u8BF7\u8F93\u5165\u8BF7\u6C42\u5934\u5927\u5C0F"})}),(0,e.jsx)(Xn.A,{ghost:!0,className:o.disclosure,expandIconPlacement:"end",activeKey:me,onChange:function(je){return le((Array.isArray(je)?je:[je]).map(String))},items:[{key:"timeouts",label:(0,e.jsx)(va,{title:"\u8FDE\u63A5\u4E0E\u8D85\u65F6",description:"5 \u9879 \xB7 \u6570\u503C\u5355\u4F4D\u5747\u4E3A\u79D2\uFF0C0 \u7684\u5177\u4F53\u542B\u4E49\u89C1\u5404\u9879\u8BF4\u660E"}),children:(0,e.jsx)(da.Ay,{componentSize:"small",theme:hi,children:ca.map(function(W){var je=C()(W,3),rn=je[0],Me=je[1],Ve=je[2];return(0,e.jsx)(yr,{styles:o,name:rn,title:Me,description:Ve,rules:[{type:"number",min:0,transform:function(sn){return sn===""?void 0:Number(sn)},message:"\u8BF7\u8F93\u5165\u4E0D\u5C0F\u4E8E 0 \u7684\u6570\u5B57"}],children:(0,e.jsx)(ze.A,{suffix:"\u79D2",inputMode:"decimal",placeholder:"\u8BF7\u8F93\u5165\u79D2\u6570",autoComplete:"off"})},rn)})})},{key:"paths",label:(0,e.jsx)(va,{title:"\u6570\u636E\u4E0E\u65E5\u5FD7\u8DEF\u5F84",description:"5 \u9879 \xB7 \u7559\u7A7A\u65F6\u81EA\u52A8\u4F7F\u7528\u7CFB\u7EDF\u9ED8\u8BA4\u8DEF\u5F84"}),children:(0,e.jsx)(e.Fragment,{children:vi.map(function(W){var je=C()(W,4),rn=je[0],Me=je[1],Ve=je[2],Oe=je[3];return(0,e.jsx)(yr,{styles:o,wide:!0,name:rn,title:Me,description:Oe,children:(0,e.jsx)(Br.A,{mode:Ve,size:"small",buttonProps:{color:"default",variant:"filled",style:{borderColor:"transparent"}},placeholder:"\u8BF7\u8F93\u5165\u6216\u9009\u62E9".concat(Me,"\uFF0C\u7559\u7A7A\u4F7F\u7528\u9ED8\u8BA4\u8DEF\u5F84"),autoComplete:"off"})},rn)})})}]})]}),zt=Le==="certificate"&&(0,e.jsxs)("div",{className:o.section,children:[(0,e.jsx)(Pt,{styles:o,title:"\u8BC1\u4E66\u7533\u8BF7",description:"\u8FD9\u91CC\u4EC5\u914D\u7F6E\u5168\u5C40 ACME HTTP-01 \u9A8C\u8BC1\u53C2\u6570\uFF0C\u8BC1\u4E66\u7533\u8BF7\u548C\u7F51\u7AD9\u7ED1\u5B9A\u4ECD\u5728\u8BC1\u4E66\u7BA1\u7406\u4E2D\u5B8C\u6210\u3002"}),(0,e.jsx)(yr,{styles:o,name:"acme_email",title:"\u8D26\u6237\u90AE\u7BB1",description:"\u7528\u4E8E ACME \u8D26\u6237\u6CE8\u518C\u548C\u8BC1\u4E66\u5230\u671F\u63D0\u9192\uFF0C\u53EF\u7559\u7A7A\u3002",rules:[{type:"email",message:"\u90AE\u7BB1\u683C\u5F0F\u4E0D\u6B63\u786E"}],children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 admin@example.com\uFF0C\u53EF\u7559\u7A7A",autoComplete:"off"})}),(0,e.jsx)(yr,{styles:o,name:"http_challenge_host",title:"\u9A8C\u8BC1\u76D1\u542C\u5730\u5740",description:"\u9650\u5236\u9A8C\u8BC1\u670D\u52A1\u76D1\u542C\u7684\u672C\u673A\u5730\u5740\uFF0C\u7559\u7A7A\u8868\u793A\u76D1\u542C\u6240\u6709\u53EF\u7528\u5730\u5740\u3002",children:(0,e.jsx)(ze.A,{placeholder:"\u4F8B\u5982 0.0.0.0\uFF0C\u7559\u7A7A\u76D1\u542C\u6240\u6709\u5730\u5740",autoComplete:"off"})}),(0,e.jsx)(yr,{styles:o,name:"http_challenge_port",title:"\u9A8C\u8BC1\u7AEF\u53E3",description:"HTTP-01 \u9A8C\u8BC1\u4F7F\u7528\u7684\u672C\u673A\u7AEF\u53E3\uFF0C\u9ED8\u8BA4 80\uFF0C\u5E76\u9700\u80FD\u88AB\u516C\u7F51\u8BBF\u95EE\u3002",rules:[{type:"number",min:1,max:65535,message:"\u7AEF\u53E3\u8303\u56F4\u4E3A 1-65535"}],children:(0,e.jsx)(qe.A,{min:1,max:65535,precision:0,placeholder:"\u8BF7\u8F93\u5165\u9A8C\u8BC1\u7AEF\u53E3",style:{width:"100%"}})}),(0,e.jsxs)("div",{className:o.infoNote,children:[(0,e.jsx)(ce.In,{type:"InfoCircleOutlined"}),"ACME \u8D26\u6237\u4E0E\u7B7E\u53D1\u6570\u636E\u7531\u7CFB\u7EDF\u4FDD\u5B58\u5728 data/acme\uFF1B\u4FEE\u6539\u8FD0\u884C\u6570\u636E\u6216\u7F13\u5B58\u76EE\u5F55\u4E0D\u4F1A\u5F71\u54CD\u8D26\u6237\u4FE1\u606F\u3002"]})]}),lt=Le==="pages"&&(0,e.jsxs)("div",{className:o.section,children:[(0,e.jsx)(Pt,{styles:o,title:"\u9ED8\u8BA4\u9875\u9762",description:"\u81EA\u5B9A\u4E49\u7F51\u7AD9\u521D\u59CB\u9875\u9762\u3001\u8D44\u6E90\u4E0D\u5B58\u5728\u3001\u57DF\u540D\u672A\u7ED1\u5B9A\u548C\u7F51\u7AD9\u505C\u7528\u65F6\u7684\u9875\u9762\u5185\u5BB9\u3002"}),(0,e.jsx)(Lt.A,{block:!0,className:o.pageSwitcher,value:tn,onChange:function(je){return pe(String(je))},options:pa.map(function(W){return{label:W.label,value:W.key}})}),pa.filter(function(W){return tn===W.key}).map(function(W){return(0,e.jsxs)("div",{className:o.pageEditor,children:[(0,e.jsxs)("div",{className:"editor-field",children:[(0,e.jsxs)("div",{className:"editor-label",children:[(0,e.jsx)("strong",{children:"\u9875\u9762\u5185\u5BB9"}),(0,e.jsx)("span",{children:"\u652F\u6301\u5B8C\u6574 HTML\uFF1B\u7559\u7A7A\u5373\u53EF\u6062\u590D\u7CFB\u7EDF\u5185\u7F6E\u9875\u9762\u3002"})]}),(0,e.jsx)(d.A.Item,{name:[W.key,"body"],children:(0,e.jsx)(ru.A,{mode:"html",theme:j?"one_dark":"github",width:"100%",height:y?220:260,fontSize:12,tabSize:4,showPrintMargin:!1,showLineNumbers:!0,wrapEnabled:!0,acePath:Qo,placeholder:"\u7559\u7A7A\u5373\u53EF\u6062\u590D\u7CFB\u7EDF\u5185\u7F6E\u9875\u9762",className:o.code})})]}),(0,e.jsx)(yr,{styles:o,name:[W.key,"status"],title:"\u54CD\u5E94\u72B6\u6001\u7801",description:"".concat(W.description,"\uFF0C\u7559\u7A7A\u4F7F\u7528\u9ED8\u8BA4\u72B6\u6001\u7801 ").concat(W.status,"\u3002"),children:(0,e.jsx)(De.A,{allowClear:!0,showSearch:!0,optionFilterProp:"label",options:Jo,placeholder:"\u8BF7\u9009\u62E9\u54CD\u5E94\u72B6\u6001\u7801\uFF0C\u7559\u7A7A\u4F7F\u7528 ".concat(W.status)})}),(0,e.jsx)(Xn.A,{ghost:!0,className:o.disclosure,expandIconPlacement:"end",activeKey:te,onChange:function(rn){return mn((Array.isArray(rn)?rn:[rn]).map(String))},items:[{key:"headers",label:(0,e.jsx)(va,{title:"\u54CD\u5E94\u5934\uFF08\u9AD8\u7EA7\uFF09",description:"\u8BBE\u7F6E\u6B64\u9875\u9762\u989D\u5916\u8FD4\u56DE\u7684 HTTP \u54CD\u5E94\u5934"}),children:(0,e.jsx)(da.Ay,{componentSize:"small",theme:hi,children:(0,e.jsx)(rd,{name:W.key,styles:o})})}]})]},W.key)})]}),zn={service:vr,basic:Wr,network:Ur,runtime:Cn,certificate:zt,pages:lt};return(0,e.jsxs)(e.Fragment,{children:[(0,e.jsx)(ce.Ht,{ref:v,title:(0,e.jsx)(ce.hE,{children:"\u670D\u52A1\u8BBE\u7F6E"}),width:y?"620px":"680px",onOk:m.submit,onCancel:ut,modalProps:{rootClassName:o.modal,confirmLoading:$||U,okText:"\u4FDD\u5B58",cancelText:"\u5173\u95ED",okButtonProps:Le==="service"?{style:{display:"none"}}:void 0,footer:U||de?null:void 0,forceRender:!0,closable:!0,keyboard:!0,mask:{closable:!0},focusable:{focusTriggerAfterClose:!1},style:{top:F.md?100:24,paddingBottom:F.md?100:24},afterClose:function(){u.current+=1,E(!1),P(!1),q(!1),ee(""),xe(""),Be([]),Fe("service"),pe("initial_page"),ue([]),le([]),mn([])}},children:(0,e.jsx)("div",{style:A?void 0:{display:"none"},children:(0,e.jsx)(mt.A,{spinning:U,size:"medium",children:de?(0,e.jsx)("div",{className:o.errorState,children:(0,e.jsx)(Gn.A,{image:Gn.A.PRESENTED_IMAGE_SIMPLE,description:de,children:(0,e.jsx)(ge.A,{type:"primary",onClick:function(){In(!0,!0)},children:"\u91CD\u65B0\u52A0\u8F7D"})})}):(0,e.jsxs)(d.A,{form:m,className:"http-service-workbench ".concat(o.workbench),"data-section":Le,layout:"vertical",variant:"filled",onFinish:Fr,onFinishFailed:function(je){var rn=je.errorFields,Me=rn[0];if(Me){var Ve=Me.name.map(String);ca.some(function(Oe){var sn=C()(Oe,1),Re=sn[0];return Re===Ve[0]})&&le(function(Oe){return Array.from(new Set([].concat(Je()(Oe),["timeouts"])))}),Ve.includes("headers")&&mn(function(Oe){return Array.from(new Set([].concat(Je()(Oe),["headers"])))}),requestAnimationFrame(function(){return requestAnimationFrame(function(){m.scrollToField(Me.name,{block:"center"})})})}},style:U?{visibility:"hidden"}:void 0,children:[(0,e.jsx)("aside",{className:o.navigation,"aria-label":"\u670D\u52A1\u8BBE\u7F6E\u5206\u7C7B",children:(0,e.jsx)("div",{className:"navigation-scroll",ref:w,children:(0,e.jsx)("div",{className:"navigation-grid",children:Ko.map(function(W){return(0,e.jsxs)("button",{type:"button",className:"navigation-item","data-section":W.key,"data-active":Le===W.key,"aria-current":Le===W.key?"page":void 0,onClick:function(){return Fe(W.key)},children:[(0,e.jsx)(ce.In,{type:W.icon}),(0,e.jsx)("span",{className:"navigation-label",children:W.label})]},W.key)})})})}),(0,e.jsx)("div",{className:o.mainPane,children:(0,e.jsx)("main",{className:o.content,children:(0,e.jsx)("div",{className:o.contentInner,children:zn[Le]})})})]})})})}),(0,e.jsx)(sa.A,{ref:N,request:se,title:"\u670D\u52A1\u65E5\u5FD7"})]})});as.displayName="HTTPService";var ad=(0,i.memo)(as),is={enable:{request:ys,title:"\u542F\u7528\u7F51\u7AD9",content:function(n){return"\u786E\u5B9A\u542F\u7528\u7F51\u7AD9\u201C".concat(n,"\u201D\u5417\uFF1F")},okText:"\u542F\u7528",success:"\u7F51\u7AD9\u5DF2\u542F\u7528",filter:function(n){return Number(n==null?void 0:n.status)!==0}},disable:{request:Fs,title:"\u505C\u7528\u7F51\u7AD9",content:function(n){return"\u786E\u5B9A\u505C\u7528\u7F51\u7AD9\u201C".concat(n,"\u201D\u5417\uFF1F\u505C\u7528\u540E\u5C06\u65E0\u6CD5\u7EE7\u7EED\u8BBF\u95EE\u3002")},okText:"\u505C\u7528",success:"\u7F51\u7AD9\u5DF2\u505C\u7528",filter:function(n){return Number(n==null?void 0:n.status)===0}},delete:{request:bs,title:"\u5220\u9664\u7F51\u7AD9",content:function(n){return"\u786E\u5B9A\u5220\u9664\u7F51\u7AD9\u201C".concat(n,"\u201D\u5417\uFF1F\u5220\u9664\u540E\u65E0\u6CD5\u6062\u590D\u3002")},okText:"\u5220\u9664",success:"\u7F51\u7AD9\u5DF2\u5220\u9664",filter:function(){return!0},danger:!0}},id=function(){var n=un.A.useApp(),t=n.message,r=n.modal,a=(0,i.useRef)(0),l=(0,i.useRef)(new Set),h=(0,i.useState)([]),c=C()(h,2),f=c[0],y=c[1],j=(0,i.useState)(0),F=C()(j,2),B=F[0],o=F[1],v=(0,i.useState)(""),N=C()(v,2),w=N[0],u=N[1],S=(0,i.useState)("keyword"),x=C()(S,2),m=x[0],p=x[1],_=(0,i.useState)({page:1,pageSize:10,keyword:"",field:"keyword",typeId:"0",type:"",status:"",sort:"id",order:"desc"}),A=C()(_,2),E=A[0],k=A[1],T=(0,i.useState)(!0),U=C()(T,2),P=U[0],M=U[1],I=(0,i.useState)(!1),$=C()(I,2),q=$[0],z=$[1],he=(0,i.useState)(0),R=C()(he,2),g=R[0],L=R[1],J=(0,i.useState)(function(){return new Set}),V=C()(J,2),ee=V[0],fe=V[1],Y=E.page,de=E.pageSize,xe=E.typeId,ye=E.type,Ce=E.status,ie=E.sort,Be=E.order;(0,i.useEffect)(function(){var Ee=++a.current,we=!0,ue=!1;return M(!0),Aa({body:D()(D()({page:Y,page_size:de,order_by_field:ie,order_by_type:Be},E.keyword?Yu()({},E.field,E.keyword):{}),{},{type_id:xe,type:ye,status:Ce})}).then(function(K){if(!(!we||a.current!==Ee)){if((K==null?void 0:K.code)!==200){t.error((K==null?void 0:K.message)||"\u83B7\u53D6\u7F51\u7AD9\u5217\u8868\u5931\u8D25");return}var ve=(K==null?void 0:K.data)||{},me=Array.isArray(ve==null?void 0:ve.list)?ve==null?void 0:ve.list.filter(Boolean):[],le=Number((ve==null?void 0:ve.total)||0);if(me.length===0&&le>0&&Y>1){ue=!0,k(function(Ae){return D()(D()({},Ae),{},{page:Math.max(1,Math.min(Y-1,Math.ceil(le/de)))})});return}y(me),o(le)}}).catch(function(){we&&a.current===Ee&&t.error("\u83B7\u53D6\u7F51\u7AD9\u5217\u8868\u5931\u8D25")}).finally(function(){we&&a.current===Ee&&!ue&&(M(!1),z(!0))}),function(){we=!1}},[t,Be,Y,de,E.field,E.keyword,g,ie,Ce,ye,xe]);var Z=(0,i.useCallback)(function(){a.current+=1,L(function(Ee){return Ee+1})},[]),cn=(0,i.useCallback)(function(Ee){var we=arguments.length>1&&arguments[1]!==void 0?arguments[1]:m;u(Ee),k(function(ue){return D()(D()({},ue),{},{page:1,keyword:Ee.trim(),field:we})})},[m]),Le=(0,i.useCallback)(function(Ee){k(function(we){return D()(D()({},we),{},{page:1,type:Ee})})},[]),Fe=(0,i.useCallback)(function(Ee){k(function(we){return D()(D()({},we),{},{page:1,typeId:Ee})})},[]),nn=(0,i.useCallback)(function(Ee){k(function(we){return D()(D()({},we),{},{page:1,status:Ee})})},[]),Ue=(0,i.useCallback)(function(Ee,we){k(function(ue){return D()(D()({},ue),{},{page:1,sort:we?Ee:"id",order:we==="ascend"?"asc":"desc"})})},[]),tn=(0,i.useCallback)(function(Ee,we){k(function(ue){return D()(D()({},ue),{},{page:we===ue.pageSize?Ee:1,pageSize:we})})},[]),pe=(0,i.useCallback)((function(){var Ee=ne()(b()().mark(function we(ue,K){var ve,me,le,Ae,re,te=arguments;return b()().wrap(function(se){for(;;)switch(se.prev=se.next){case 0:if(ve=te.length>2&&te[2]!==void 0?te[2]:{},me=is[ue],le=K==null?void 0:K.id,Ae=String(le!=null?le:""),!(!me||le===void 0||le===null||!me.filter(K))){se.next=7;break}return t.warning("\u5F53\u524D\u7F51\u7AD9\u72B6\u6001\u4E0D\u652F\u6301\u6B64\u64CD\u4F5C"),se.abrupt("return");case 7:if(!l.current.has(Ae)){se.next=9;break}return se.abrupt("return");case 9:return l.current.add(Ae),fe(new Set(l.current)),se.prev=11,se.next=14,me.request({body:D()({id:le},ve)});case 14:if(re=se.sent,(re==null?void 0:re.code)===200){se.next=19;break}return t.error((re==null?void 0:re.message)||"\u64CD\u4F5C\u5931\u8D25"),ue==="delete"&&Z(),se.abrupt("return");case 19:(0,ar.Rp)("site"),t.success((re==null?void 0:re.message)||me.success),Z(),se.next=27;break;case 24:se.prev=24,se.t0=se.catch(11),t.error("\u64CD\u4F5C\u5931\u8D25");case 27:return se.prev=27,l.current.delete(Ae),fe(new Set(l.current)),se.finish(27);case 31:case"end":return se.stop()}},we,null,[[11,24,27,31]])}));return function(we,ue){return Ee.apply(this,arguments)}})(),[t,Z]),Se=(0,i.useCallback)(function(Ee,we){var ue=is[Ee];if(ue){var K=!1;r.confirm({title:ue.title,content:Ee==="delete"?(0,i.createElement)("div",null,(0,i.createElement)("div",null,ue.content((we==null?void 0:we.name)||"-")),(0,i.createElement)("div",{style:{marginTop:12}},(0,i.createElement)(la.A,{onChange:function(me){K=me.target.checked}},"\u5C06\u7F51\u7AD9\u6839\u76EE\u5F55\u79FB\u81F3\u56DE\u6536\u7AD9"))):ue.content((we==null?void 0:we.name)||"-"),okText:ue.okText,cancelText:"\u53D6\u6D88",okButtonProps:ue.danger?{danger:!0,type:"default"}:void 0,mask:{closable:!0},onOk:function(){if(Ee!=="delete")return pe(Ee,we);r.confirm({title:"\u518D\u6B21\u786E\u8BA4",content:K?"\u786E\u5B9A\u5220\u9664\u7F51\u7AD9\u5E76\u5C06\u7F51\u7AD9\u6839\u76EE\u5F55\u79FB\u81F3\u56DE\u6536\u7AD9\u5417\uFF1F":"\u786E\u5B9A\u5220\u9664\u7F51\u7AD9\u5417\uFF1F",okText:"\u5220\u9664",cancelText:"\u53D6\u6D88",okButtonProps:{danger:!0,type:"default"},mask:{closable:!0},onOk:function(){return pe(Ee,we,{delete_root:K})}})}})}},[pe,r]);return{sites:f,total:B,page:Y,pageSize:de,keyword:w,searchField:m,typeId:xe,type:ye,status:Ce,sort:ie,order:Be,loading:P,initialized:q,operating:ee,reload:Z,changeKeyword:u,changeSearchField:p,search:cn,changeTypeId:Fe,changeType:Le,changeStatus:nn,changeSort:Ue,changePage:tn,confirmAction:Se}},ud=id,us,ld=(0,kr.rU)(function(s){var n=s.css,t=s.token;return{siteTable:n(us||(us=H()([`
        .root-cell {
            width: 220px;
            max-width: 220px;
            min-width: 0;
        }

        .site-name,
        .site-name > button,
        .site-category,
        .site-type,
        .site-root,
        .update-time {
            min-width: 0;
        }

        .site-name > button {
            display: block;
            max-width: 100%;
            padding: 0;
            overflow: hidden;
            border: 0;
            outline: none;
            background: transparent;
            color: inherit;
            cursor: pointer;
            font: inherit;
            line-height: 20px;
            text-align: left;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .site-name > button:hover,
        .site-name > button:focus-visible {
            color: `,`;
        }

        .site-category,
        .site-type,
        .site-root,
        .update-time {
            display: block;
            overflow: hidden;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .site-root {
            width: 100%;
            color: `,`;
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: `,`px;
        }

        .site-root-link {
            padding: 0;
            border: 0;
            outline: none;
            background: transparent;
            cursor: pointer;
            text-align: left;
        }

        .site-root-link:hover,
        .site-root-link:focus-visible {
            color: `,`;
        }

        .site-state {
            display: inline-flex;
            align-items: center;
            gap: 5px;
            line-height: 20px;
            white-space: nowrap;
        }

        .site-state i {
            width: 6px;
            height: 6px;
            flex: none;
            border-radius: 50%;
            background: currentColor;
        }

        .site-state.running {
            color: `,`;
        }

        .site-state.stopped {
            color: `,`;
        }
    `])),t.colorPrimary,t.colorTextSecondary,t.fontSizeSM,t.colorPrimary,t.colorSuccess,t.colorWarning)}}),sd=ld,od={0:"\u9759\u6001\u7F51\u7AD9",1:"\u53CD\u5411\u4EE3\u7406",2:"\u901A\u7528\u7F51\u7AD9",3:"PHP\u7F51\u7AD9"},dd={0:"\u5DF2\u542F\u7528",1:"\u5DF2\u505C\u7528"},ls=(function(){var s=sd(),n=s.styles,t=(0,ar.Ay)("site"),r=C()(t,2),a=r[0],l=r[1],h=(0,i.useRef)({}),c=(0,i.useRef)({}),f=(0,i.useRef)(null),y=(0,i.useRef)(null),j=(0,i.useRef)({}),F=(0,i.useRef)({}),B=(0,i.useRef)(0),o=ud(),v=(0,i.useState)(0),N=C()(v,2),w=N[0],u=N[1],S=(0,i.useState)(0),x=C()(S,2),m=x[0],p=x[1],_=(0,i.useState)(null),A=C()(_,2),E=A[0],k=A[1],T=(a==null?void 0:a.kind)||od,U=(a==null?void 0:a.status)||dd,P=(0,i.useMemo)(function(){return E===null?(a==null?void 0:a.type)||{}:Object.fromEntries(E.map(function(g){return[String(g.id),String(g.name)]}))},[a==null?void 0:a.type,E]);(0,i.useEffect)(function(){var g=!0,L=B.current;return(0,tt.Jt)({body:{module:"site",order_by_field:"sort",order_by_type:"asc"}}).then(function(J){var V,ee;!g||B.current!==L||(J==null?void 0:J.code)!==200||k(Array.isArray(J==null||(V=J.data)===null||V===void 0?void 0:V.list)?J==null||(ee=J.data)===null||ee===void 0?void 0:ee.list.filter(Boolean):[])}),function(){g=!1}},[]);var M=(0,i.useCallback)(function(g,L){var J;return(J=c.current)===null||J===void 0?void 0:J.open(g,L)},[]),I=(0,i.useCallback)(function(){var g;return(g=F.current)===null||g===void 0?void 0:g.open()},[]),$=(0,i.useCallback)(function(){return u(function(g){return g+1})},[]),q=(0,i.useCallback)(function(){p(function(g){return g+1}),u(function(g){return g+1})},[]),z=(0,i.useCallback)(function(g){B.current+=1;var L=Object.fromEntries(g.map(function(J){return[String(J.id),J.name]}));k(g),o.typeId!=="0"&&!L[o.typeId]&&o.changeTypeId("0")},[o.changeTypeId,o.typeId]),he=ot({keyword:o.keyword,searchField:o.searchField,typeId:o.typeId,type:o.type,status:o.status,typeItems:E||void 0,typeData:T,statusData:U,refreshing:o.loading,onKeywordChange:o.changeKeyword,onSearchFieldChange:o.changeSearchField,onSearch:o.search,onTypeIdChange:o.changeTypeId,onManageTypes:I,onTypeChange:o.changeType,onStatusChange:o.changeStatus,onRefresh:o.reload,onCreate:function(){var L;return(L=h.current)===null||L===void 0?void 0:L.open()},onHTTPSettings:function(){var L;return(L=j.current)===null||L===void 0?void 0:L.open()},onManageCert:function(){var L;return(L=f.current)===null||L===void 0?void 0:L.open()},onManageSecret:function(){var L;return(L=y.current)===null||L===void 0?void 0:L.open()}}),R=yn({className:n.siteTable,sites:o.sites,loading:o.loading,categoryData:P,typeData:T,statusData:U,sort:o.sort,order:o.order,operating:o.operating,onSortChange:o.changeSort,onSettings:M,onAction:o.confirmAction});return(0,e.jsxs)(ce.nB,{loading:l,children:[(0,e.jsx)(ce.II,D()(D()(D()({},he),R),{},{ariaLabel:"\u7F51\u7AD9\u7BA1\u7406",paginationAffix:{offsetBottom:15},empty:o.initialized&&o.sites.length===0,pagination:{page:o.page,pageSize:o.pageSize,total:o.total,unit:"\u6761",onChange:o.changePage}})),(0,e.jsx)(As,{ref:h,typeItems:E||void 0,onSuccess:o.reload}),(0,e.jsx)(xo,{ref:c,certVersion:w,typeItems:E||void 0,onSuccess:o.reload,onObtainCert:function(L,J){var V;return(V=f.current)===null||V===void 0?void 0:V.openObtain(L,J)},onDeployCert:function(L,J){var V;return(V=f.current)===null||V===void 0?void 0:V.openDeploy(L,J)},onRenewCert:function(L,J,V){var ee;return(ee=f.current)===null||ee===void 0?void 0:ee.openRenew(L,J,V)},onManageCert:function(){var L;return(L=f.current)===null||L===void 0?void 0:L.open()}}),(0,e.jsx)(Wo,{ref:f,secretVersion:m,onChange:$}),(0,e.jsx)(mi,{ref:y,onChange:q}),(0,e.jsx)(ad,{ref:j}),(0,e.jsx)(rt.A,{ref:F,module:"site",onChange:z,onMutation:o.reload})]})});X.dn(ls)}),23455:(function(_n,Rn,X){"use strict";X.d(Rn,{R7:function(){return Kn},zg:function(){return ge},jO:function(){return at},xz:function(){return ot},PC:function(){return Je}});var Un=X(59224),D=X.n(Un),tr=X(39058),C=[{path:"/",title:"\u6570\u636E\u6982\u89C8",name:"index",icon:"icon-home",hideInMenu:!1,hideBreadCrumb:!1,component:"@/pages/index"},{path:"site",title:"\u7F51\u7AD9\u7BA1\u7406",name:"site",icon:"GlobalOutlined",hideInMenu:!1,component:"@/pages/site"},{path:"file",title:"\u6587\u4EF6\u7BA1\u7406",name:"file",icon:"FolderOutlined",hideInMenu:!1,component:"@/pages/file"},{path:"task",title:"\u8BA1\u5212\u4EFB\u52A1",name:"task",icon:"ScheduleOutlined",hideInMenu:!1,component:"@/pages/task"},{path:"server",title:"\u7EC8\u7AEF\u7BA1\u7406",name:"server",icon:"CodeOutlined",hideInMenu:!1,component:"@/pages/server"},{path:"database",title:"\u6570\u636E\u7BA1\u7406",name:"database",icon:"DatabaseOutlined",hideInMenu:!1,component:"@/pages/database"},{path:"monitor",title:"\u76D1\u63A7\u7BA1\u7406",name:"monitor",icon:"LineChartOutlined",hideInMenu:!1,component:"@/pages/monitor"},{path:"software",title:"\u8F6F\u4EF6\u7BA1\u7406",name:"software",icon:"AppstoreOutlined",hideInMenu:!1,component:"@/pages/software"},{path:"log",title:"\u64CD\u4F5C\u65E5\u5FD7",name:"log",icon:"FileTextOutlined",hideInMenu:!1,component:"@/pages/log"},{path:"setting",title:"\u7CFB\u7EDF\u8BBE\u7F6E",name:"setting",icon:"SettingOutlined",hideInMenu:!1,component:"@/pages/setting"},{path:"login",component:"@/pages/login",name:"login",title:"\u5E10\u53F7\u767B\u5F55",auth:!1,hideInMenu:!0},{path:"500",component:"@/pages/500",name:"error",title:"\u670D\u52A1\u5668\u9519\u8BEF",auth:!1,hideInMenu:!0},{path:"403",component:"@/pages/403",name:"notAllowed",title:"\u65E0\u6743\u9650",auth:!1,hideInMenu:!0},{path:"*",component:"@/pages/404",name:"notFound",title:"\u9875\u9762\u4E0D\u5B58\u5728",auth:!1,hideInMenu:!0}],i=X(62885),ce=X(96540),rt=X(74848);function tt(gn,on){var Ye=arguments.length>2&&arguments[2]!==void 0?arguments[2]:{},Te=arguments.length>3&&arguments[3]!==void 0?arguments[3]:"",yn=D()(gn),Mn;try{for(yn.s();!(Mn=yn.n()).done;){var b=Mn.value,ur=Te+(Te&&!Te.endsWith("/")?"/":"")+(b.path.startsWith("/")?b.path.slice(1):b.path);if(Ye&&Object.keys(Ye).length>0)for(var ne in Ye)b.path.includes(":".concat(ne))&&(ur=ur.replace(":".concat(ne),Ye[ne]));if(b.name===on)return ur.startsWith("/")?ur:"/".concat(ur);if(b.routes){var Ke=tt(b.routes,on,Ye,ur);if(Ke)return Ke}}}catch(H){yn.e(H)}finally{yn.f()}return null}var ar={};function at(gn){var on=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};if(Object.keys(on).length<=0&&(ar==null?void 0:ar[gn])!=null)return ar[gn];var Ye=tt(C,gn,on,"");return Object.keys(on).length<=0&&(ar[gn]=Ye),Ye}function Je(gn){var on=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};tr.b6===null||tr.b6===void 0||tr.b6.push(at(gn,on)||gn||"/")}function e(gn){var on=arguments.length>1&&arguments[1]!==void 0?arguments[1]:"/",Ye=arguments.length>2&&arguments[2]!==void 0?arguments[2]:!0;return(gn||[]).map(function(Te){var yn;if((Te==null?void 0:Te.path)!==""&&!(Ye&&Te!==null&&Te!==void 0&&Te.hideInMenu)){var Mn={label:Te==null?void 0:Te.title,path:Te==null?void 0:Te.path,name:Te==null?void 0:Te.name};if(Te!=null&&Te.icon&&typeof(Te==null?void 0:Te.icon)=="string"&&(Te==null?void 0:Te.icon)!="0"&&(Mn.icon=(0,rt.jsx)(i.In,{type:Te==null?void 0:Te.icon})),on!="/"&&!(Te!=null&&(yn=Te.path)!==null&&yn!==void 0&&yn.startsWith("/"))?Mn.key="".concat(on,"/").concat(Te.path):on=="/"&&Te.path!="/"?Mn.key="/".concat(Te.path):Mn.key="".concat(Te.path),Te!=null&&Te.routes){var b=e(Te==null?void 0:Te.routes,Mn.key,Ye).filter(Boolean);b.length>0?Mn.children=b:Mn.children=[]}return Mn}}).filter(Boolean)}var Or={};function ot(gn,on){if(!on||!gn||gn.length<=0)return[];if((Or==null?void 0:Or[on])!=null)return Or[on];var Ye=[];function Te(yn,Mn){var b;if(yn!=null&&yn.children&&(yn==null||(b=yn.children)===null||b===void 0?void 0:b.length)>0){if((yn==null?void 0:yn.name)===Mn)return Ye.push(yn),!0}else{var ur=D()(yn),ne;try{for(ur.s();!(ne=ur.n()).done;){var Ke=ne.value;if((Ke==null?void 0:Ke.name)===Mn||Ke!=null&&Ke.children&&Te(Ke==null?void 0:Ke.children,Mn))return Ye.push(Ke),!0}}catch(H){ur.e(H)}finally{ur.f()}}return!1}return Te(gn,on),Ye=Ye.reverse(),Ye.length>0&&(Or[on]=Ye),Ye}function Kn(){var gn=arguments.length>0&&arguments[0]!==void 0?arguments[0]:!0;return e(C,"/",gn)}function ge(gn){var on=D()(gn),Ye;try{for(on.s();!(Ye=on.n()).done;){var Te=Ye.value;if(Te)if(Te.hasOwnProperty("children")){var yn=ge(Te.children);if(yn!==null)return yn}else return Te}}catch(Mn){on.e(Mn)}finally{on.f()}return null}}),63496:(function(_n,Rn,X){"use strict";X.d(Rn,{K:function(){return Un}});function Un(D){if(D){var tr=D.split(" ",2),C=((tr==null?void 0:tr[0])||"").replace(/\//g,"-").split("-",3),i=((tr==null?void 0:tr[1])||"").split(":",3),ce=new Date(parseInt(C[0],10)||0,(parseInt(C[1],10)||1)-1,parseInt(C[2],10)||0,parseInt(i[0],10)||0,parseInt(i[1],10)||0,parseInt(i[2],10)||0).getTime(),rt=1e3*60,tt=rt*60,ar=tt*24,at=ar*7,Je=ar*30;if(Number.isFinite(ce)){var e=Date.now()-ce,Or=e/rt,ot=e/tt,Kn=e/ar,ge=e/at,gn=e/Je,on;if(e<=rt)on="\u521A\u521A";else if(Or<60)on=parseInt(String(Or))+"\u5206\u949F\u524D";else if(ot<24)on=parseInt(String(ot))+"\u5C0F\u65F6\u524D";else if(Kn<7)on=parseInt(String(Kn))+"\u5929\u524D";else if(Kn<30)on=parseInt(String(ge))+"\u5468\u524D";else if(gn<6)on=parseInt(String(gn))+"\u6708\u524D";else{var Ye=new Date;Ye.setTime(ce);var Te=Ye.getFullYear(),yn=Ye.getMonth()+1<10?"0"+(Ye.getMonth()+1):Ye.getMonth()+1,Mn=Ye.getDate()<10?"0"+Ye.getDate():Ye.getDate(),b=Ye.getHours()<10?"0"+Ye.getHours():Ye.getHours(),ur=Ye.getMinutes()<10?"0"+Ye.getMinutes():Ye.getMinutes(),ne=Ye.getSeconds()<10?"0"+Ye.getSeconds():Ye.getSeconds();on=Te+"-"+yn+"-"+Mn+" "+b+":"+ur+":"+ne}return on}}}}),97087:(function(_n,Rn,X){"use strict";X.d(Rn,{A:function(){return ne}});var Un=X(96540),D=X(65437),tr=X(34164),C=X(42704),i=X(42283),ce=X(62279),rt=X(37358),tt=X(11214);const ar=["wrap","nowrap","wrap-reverse"],at=["flex-start","flex-end","start","end","center","space-between","space-around","space-evenly","stretch","normal","left","right"],Je=["center","start","end","flex-start","flex-end","self-start","self-end","baseline","normal","stretch"],e=(Ke,H)=>{const un=H.wrap===!0?"wrap":H.wrap;return{[`${Ke}-wrap-${un}`]:un&&ar.includes(un)}},Or=(Ke,H)=>{const un={};return Je.forEach(Sn=>{un[`${Ke}-align-${Sn}`]=H.align===Sn}),un[`${Ke}-align-stretch`]=!H.align&&!!H.vertical,un},ot=(Ke,H)=>{const un={};return at.forEach(Sn=>{un[`${Ke}-justify-${Sn}`]=H.justify===Sn}),un};var ge=(Ke,H)=>(0,tr.$)(bt(bt(bt({},e(Ke,H)),Or(Ke,H)),ot(Ke,H)));const gn=Ke=>{const{componentCls:H}=Ke;return{[H]:{display:"flex",margin:0,padding:0,"&-vertical":{flexDirection:"column"},"&-rtl":{direction:"rtl"},"&:empty":{display:"none"}}}},on=Ke=>{const{componentCls:H}=Ke;return{[H]:{"&-gap-small":{gap:Ke.flexGapSM},"&-gap-medium, &-gap-middle":{gap:Ke.flexGap},"&-gap-large":{gap:Ke.flexGapLG}}}},Ye=Ke=>{const{componentCls:H}=Ke,un={};return ar.forEach(Sn=>{un[`${H}-wrap-${Sn}`]={flexWrap:Sn}}),un},Te=Ke=>{const{componentCls:H}=Ke,un={};return Je.forEach(Sn=>{un[`${H}-align-${Sn}`]={alignItems:Sn}}),un},yn=Ke=>{const{componentCls:H}=Ke,un={};return at.forEach(Sn=>{un[`${H}-justify-${Sn}`]={justifyContent:Sn}}),un},Mn=()=>({});var b=(0,rt.OF)("Flex",Ke=>{const{paddingXS:H,padding:un,paddingLG:Sn}=Ke,d=(0,tt.oX)(Ke,{flexGapSM:H,flexGap:un,flexGapLG:Sn});return[gn(d),on(d),Ye(d),Te(d),yn(d)]},Mn,{resetStyle:!1}),ne=Un.forwardRef((Ke,H)=>{const ta=Ke,{prefixCls:un,rootClassName:Sn,className:d,style:ze,flex:na,gap:dt,vertical:yt,orientation:ct,component:qe="div",children:Xn}=ta,ra=gs(ta,["prefixCls","rootClassName","className","style","flex","gap","vertical","orientation","component","children"]),{flex:De,direction:Dr,getPrefixCls:kr}=Un.useContext(ce.QO),Br=kr("flex",un),[dn,Aa]=b(Br),[,Tt]=(0,i.M)(ct,yt!=null?yt:De==null?void 0:De.vertical),Ea=(0,tr.$)(d,Sn,De==null?void 0:De.className,Br,dn,Aa,ge(Br,fs(bt({},Ke),{vertical:Tt})),{[`${Br}-rtl`]:Dr==="rtl",[`${Br}-gap-${dt}`]:(0,C.X)(dt),[`${Br}-vertical`]:Tt}),Ft=bt(bt({},De==null?void 0:De.style),ze);return(0,D.R)(na)&&(Ft.flex=na),(0,D.R)(dt)&&!(0,C.X)(dt)&&(Ft.gap=dt),Un.createElement(qe,bt({ref:H,className:Ea,style:Ft},(0,D.cJ)(ra,["justify","wrap","align"])),Xn)})}),25542:(function(_n,Rn,X){var Un=X(39806),D=X(33032),tr=X(34911),C=X(99911);function i(ce){return Un(ce)||D(ce)||tr(ce)||C()}_n.exports=i,_n.exports.__esModule=!0,_n.exports.default=_n.exports})}]);
}());