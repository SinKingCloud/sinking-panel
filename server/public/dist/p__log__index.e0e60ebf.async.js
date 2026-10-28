(self.webpackChunk=self.webpackChunk||[]).push([[487],{16540:(function(l,m,e){"use strict";e.r(m),e.d(m,{default:function(){return d}});var v=e(55420),y=e.n(v),x=e(98312),E=e.n(x),L=e(11604),j=e.n(L),U=e(96540),B=e(79273),$=e(20873),W=e(24816),O=e(88826),K=e(48212),A=e(62885),c=e(72523),I=e(9300),u=e(99998),g=e(84169),b=e.n(g),T=e(12185),P,F=(0,T.rU)(function(i){var R=i.css,s=i.token;return{logTable:R(P||(P=b()([`
        min-width: 0;

        .log-copy {
            display: block;
            min-width: 0;
            max-width: 100%;
            margin: 0;
            color: inherit;
            font: inherit;
            line-height: 20px;
            overflow-wrap: anywhere;
        }

        .log-copy .ant-typography-copy {
            margin-inline-start: 5px;
            color: `,`;
            font-size: 10px;
            line-height: 1;
            opacity: 0;
            vertical-align: 0;
            transition: color .16s ease, opacity .16s ease;
        }

        .log-copy:hover .ant-typography-copy,
        .log-copy .ant-typography-copy:focus-visible {
            color: `,`;
            opacity: 1;
        }

        .log-type {
            max-width: 100%;
            display: inline-flex;
            align-items: center;
            gap: 5px;
            overflow: hidden;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .log-type i {
            width: 6px;
            height: 6px;
            flex: none;
            border-radius: 50%;
            background: `,`;
        }

        .log-type.is-0 i {
            background: `,`;
        }

        .log-type.is-1 i {
            background: `,`;
        }

        .log-type.is-2 i {
            background: `,`;
        }

        .log-type.is-3 i {
            background: `,`;
        }

        .log-type.is-4 i {
            background: `,`;
        }

        .log-text,
        .log-time {
            min-width: 0;
            display: block;
            overflow: hidden;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        @media (hover: none) {
            .log-copy .ant-typography-copy {
                opacity: .55;
            }
        }
    `])),s.colorTextQuaternary,s.colorPrimary,s.colorTextQuaternary,s.colorSuccess,s.colorPrimary,s.colorError,s.colorWarning,s.colorSuccess)}}),o=F,r=e(74848),n={},d=(function(){var i=o(),R=i.styles,s=B.A.useApp(),p=s.message,z=(0,I.Ay)("log"),_=j()(z,2),f=_[0],X=_[1],D=(0,U.useRef)(null),w=(0,U.useRef)(null),k=(0,U.useRef)(0),S=(0,U.useRef)(!1),ne=$.A.useForm(),J=j()(ne,1),C=J[0],Q=(0,U.useState)(!1),N=j()(Q,2),te=N[0],q=N[1],re=(f==null?void 0:f.type)||n;(0,U.useEffect)(function(){return function(){k.current+=1}},[]);var ae=(0,U.useMemo)(function(){return[{title:"\u64CD\u4F5CIP",dataIndex:"ip",key:"ip",width:150,render:function(a){return(0,r.jsx)(W.A.Text,{className:"log-copy",copyable:a?{text:a}:!1,children:a||"-"})}},{title:"IP\u5F52\u5C5E\u5730",dataIndex:"location",key:"location",width:220,render:function(a){return(0,r.jsx)(W.A.Text,{className:"log-copy",copyable:a?{text:a}:!1,children:a||"-"})}},{title:"\u64CD\u4F5C\u7C7B\u578B",dataIndex:"type",key:"type",width:100,render:function(a){return(0,r.jsxs)("span",{className:"log-type is-".concat(a),children:[(0,r.jsx)("i",{}),re[String(a)]||"\u672A\u77E5\u7C7B\u578B"]})}},{title:"\u64CD\u4F5C\u6807\u9898",dataIndex:"title",key:"title",width:150,render:function(a){return(0,r.jsx)(O.A,{title:a||"-",children:(0,r.jsx)("span",{className:"log-text",children:a||"-"})})}},{title:"\u64CD\u4F5C\u5185\u5BB9",dataIndex:"content",key:"content",width:260,render:function(a){return(0,r.jsx)(O.A,{title:a||"-",children:(0,r.jsx)("span",{className:"log-text",children:a||"-"})})}},{title:"\u64CD\u4F5C\u65F6\u95F4",dataIndex:"create_time",key:"create_time",width:170,sorter:!0,render:function(a){return(0,r.jsx)(O.A,{title:a||"-",children:(0,r.jsx)("span",{className:"log-time",children:a||"-"})})}}]},[re]),se=(function(){var M=E()(y()().mark(function a(Z){var V,Y,G,h;return y()().wrap(function(t){for(;;)switch(t.prev=t.next){case 0:if(!S.current){t.next=2;break}return t.abrupt("return");case 2:return V=++k.current,S.current=!0,q(!0),t.prev=5,t.next=8,(0,c.XE)({body:{action:"clear",day:String(Z.day)}});case 8:if(h=t.sent,k.current===V){t.next=11;break}return t.abrupt("return");case 11:if((h==null?void 0:h.code)===200){t.next=14;break}return p.error((h==null?void 0:h.message)||"\u6E05\u7406\u5931\u8D25"),t.abrupt("return");case 14:(Y=w.current)===null||Y===void 0||Y.hide(),(G=D.current)===null||G===void 0||G.reload(),p.success((h==null?void 0:h.message)||"\u6E05\u7406\u6210\u529F"),t.next=22;break;case 19:t.prev=19,t.t0=t.catch(5),k.current===V&&p.error("\u6E05\u7406\u5931\u8D25");case 22:return t.prev=22,k.current===V&&(S.current=!1,q(!1)),t.finish(22);case 25:case"end":return t.stop()}},a,null,[[5,19,22,25]])}));return function(Z){return M.apply(this,arguments)}})();return(0,r.jsxs)(A.nB,{loading:X,children:[(0,r.jsx)(A.II,{ref:D,ariaLabel:"\u64CD\u4F5C\u65E5\u5FD7",className:R.logTable,columns:ae,rowKey:"id",request:(function(){var M=E()(y()().mark(function a(Z,V){var Y,G,h,ee,t;return y()().wrap(function(H){for(;;)switch(H.prev=H.next){case 0:return H.next=2,(0,c.XE)({body:(0,u.tI)(Z,V)});case 2:if(t=H.sent,(t==null?void 0:t.code)===200){H.next=5;break}throw new Error((t==null?void 0:t.message)||"\u83B7\u53D6\u5217\u8868\u5931\u8D25");case 5:return H.abrupt("return",{data:Array.isArray(t==null||(Y=t.data)===null||Y===void 0?void 0:Y.list)?t==null||(G=t.data)===null||G===void 0?void 0:G.list.filter(Boolean):[],total:(h=t==null||(ee=t.data)===null||ee===void 0?void 0:ee.total)!==null&&h!==void 0?h:0,success:!0});case 6:case"end":return H.stop()}},a)}));return function(a,Z){return M.apply(this,arguments)}})(),defaultPageSize:20,defaultSort:{id:"descend"},paginationAffix:{offsetBottom:15},hero:{eyebrow:"OPERATION AUDIT",title:"\u64CD\u4F5C\u65E5\u5FD7",action:{label:"\u6E05\u7406\u65E5\u5FD7",onClick:function(){var a;k.current+=1,S.current=!1,q(!1),C.resetFields(),C.setFieldsValue({day:30}),(a=w.current)===null||a===void 0||a.show()}}},search:{placeholder:"\u641C\u7D22\u65E5\u5FD7",maxLength:200,fields:[{value:"keyword",label:"\u5168\u90E8",placeholder:"\u641C\u7D22 IP\u3001\u5F52\u5C5E\u5730\u3001\u6807\u9898\u6216\u5185\u5BB9"},{value:"title",label:"\u6807\u9898",placeholder:"\u641C\u7D22\u6807\u9898"},{value:"content",label:"\u5185\u5BB9",placeholder:"\u641C\u7D22\u5185\u5BB9"},{value:"location",label:"\u5F52\u5C5E\u5730",placeholder:"\u641C\u7D22\u5F52\u5C5E\u5730",maxLength:100},{value:"ip",label:"IP",placeholder:"\u641C\u7D22 IP"}]},filters:[{type:"select",name:"type",label:"\u65E5\u5FD7\u7C7B\u578B\u7B5B\u9009",icon:"TagsOutlined",allLabel:"\u5168\u90E8\u7C7B\u578B",valueEnum:re},{type:"date",name:"create_time",label:"\u64CD\u4F5C\u65F6\u95F4\u7B5B\u9009"}],toolbar:{refresh:{ariaLabel:"\u5237\u65B0\u64CD\u4F5C\u65E5\u5FD7\u5217\u8868"}}}),(0,r.jsx)(A.Ht,{ref:w,title:(0,r.jsx)(A.hE,{children:"\u6E05\u7406\u65E5\u5FD7"}),width:320,onOk:C.submit,onCancel:function(){var a;k.current+=1,(a=w.current)===null||a===void 0||a.hide()},modalProps:{confirmLoading:te,okText:"\u63D0\u4EA4",cancelText:"\u53D6\u6D88",cancelButtonProps:{disabled:!1},closable:!0,keyboard:!0,forceRender:!0,style:{top:100,paddingBottom:100},mask:{closable:!0},afterClose:function(){k.current+=1,S.current=!1,C.resetFields(),q(!1)}},children:(0,r.jsx)($.A,{form:C,layout:"vertical",onFinish:se,children:(0,r.jsx)($.A.Item,{name:"day",label:"\u4FDD\u7559\u5929\u6570",rules:[{required:!0,message:"\u8BF7\u8F93\u5165\u4FDD\u7559\u5929\u6570"}],children:(0,r.jsx)(K.A,{min:1,precision:0,step:1,style:{width:"100%"},placeholder:"\u8BF7\u8F93\u5165\u4FDD\u7559\u5929\u6570"})})})})]})});e.dn(d)}),9300:(function(l,m,e){"use strict";e.d(m,{Rp:function(){return b}});var v=e(1974),y=e.n(v),x=e(55420),E=e.n(x),L=e(11604),j=e.n(L),U=e(98312),B=e.n(U),$=e(45647),W=e.n($),O=e(96540),K=e(72523),A={};function c(o){var r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:[],n=Array.isArray(o)?o:[o].concat(W()(r));return Array.from(new Set(n.map(function(d){return d==null?void 0:d.trim()}).filter(Boolean)))}function I(o){return o.reduce(function(r,n){return A[n]&&(r[n]=A[n]),r},{})}function u(o){return g.apply(this,arguments)}function g(){return g=B()(E()().mark(function o(r){var n,d,i;return E()().wrap(function(s){for(;;)switch(s.prev=s.next){case 0:if(!(r.length<=0)){s.next=2;break}return s.abrupt("return",{});case 2:if(n=I(r),d=r.filter(function(p){return!n[p]}),!(d.length<=0)){s.next=6;break}return s.abrupt("return",n);case 6:return s.next=8,Promise.all(d.map((function(){var p=B()(E()().mark(function z(_){var f;return E()().wrap(function(D){for(;;)switch(D.prev=D.next){case 0:return D.next=2,(0,K.GV)({body:{name:_}});case 2:if(f=D.sent,(f==null?void 0:f.code)===200){D.next=5;break}throw new Error((f==null?void 0:f.message)||"\u83B7\u53D6\u679A\u4E3E\u6570\u636E\u5931\u8D25");case 5:return D.abrupt("return",{name:_,data:(f==null?void 0:f.data)||{}});case 6:case"end":return D.stop()}},z)}));return function(z){return p.apply(this,arguments)}})()));case 8:return i=s.sent,i==null||i.forEach(function(p){A[p.name]=p.data}),s.abrupt("return",y()(y()({},n),I(d)));case 11:case"end":return s.stop()}},o)})),g.apply(this,arguments)}function b(o){if(o){delete A[o];return}Object.keys(A).forEach(function(r){delete A[r]})}function T(o){return P.apply(this,arguments)}function P(){return P=B()(E()().mark(function o(r){var n,d,i,R,s,p=arguments;return E()().wrap(function(_){for(;;)switch(_.prev=_.next){case 0:for(n=p.length,d=new Array(n>1?n-1:0),i=1;i<n;i++)d[i-1]=p[i];if(R=c(r,d),!(R.length<=0)){_.next=4;break}return _.abrupt("return",{});case 4:return _.next=6,u(R);case 6:if(s=_.sent,!(Array.isArray(r)||d.length>0)){_.next=9;break}return _.abrupt("return",s);case 9:return _.abrupt("return",(s==null?void 0:s[R[0]])||{});case 10:case"end":return _.stop()}},o)})),P.apply(this,arguments)}function F(o){for(var r=arguments.length,n=new Array(r>1?r-1:0),d=1;d<r;d++)n[d-1]=arguments[d];var i=Array.isArray(o)||n.length>0,R=(0,O.useMemo)(function(){return c(o,n).join(",")},[Array.isArray(o)?o.join(","):o,n.join(",")]),s=(0,O.useState)({}),p=j()(s,2),z=p[0],_=p[1],f=(0,O.useState)(!0),X=j()(f,2),D=X[0],w=X[1],k=(0,O.useState)(null),S=j()(k,2),ne=S[0],J=S[1];return(0,O.useEffect)(function(){var C=c(o,n);if(C.length<=0){_({}),w(!1),J(null);return}var Q=!0;return w(!0),J(null),T(i?C:C[0]).then(function(N){Q&&_(N)}).catch(function(N){Q&&(_({}),J(N instanceof Error?N:new Error("\u83B7\u53D6\u679A\u4E3E\u6570\u636E\u5931\u8D25")))}).finally(function(){Q&&w(!1)}),function(){Q=!1}},[i,R]),[z,D,ne]}m.Ay=F}),99998:(function(l,m,e){"use strict";e.d(m,{bQ:function(){return O},tI:function(){return W}});var v,y=e(55420),x=e.n(y),E=e(1974),L=e.n(E),j=e(68328),U=e.n(j),B=e(98312),$=e.n(B);function W(c,I){function u(g){for(var b in g){var T={order_by_field:b};return g[b]=="ascend"?T.order_by_type="asc":T.order_by_type="desc",T}return{}}return c.page=c.current,c.page_size=c.pageSize,delete c.current,delete c.pageSize,Object.assign(c,u(I))}function O(c,I,u){return K.apply(this,arguments)}function K(){return K=$()(x()().mark(function c(I,u,g){var b,T,P,F,o,r,n;return x()().wrap(function(i){for(;;)switch(i.prev=i.next){case 0:return r=W(I,u),i.next=3,g({body:L()({},r)});case 3:return n=i.sent,i.abrupt("return",{data:(n==null||(b=n.data)===null||b===void 0?void 0:b.list)===void 0||(n==null||(T=n.data)===null||T===void 0?void 0:T.list)===null||(n==null||(P=n.data)===null||P===void 0||(P=P.list)===null||P===void 0?void 0:P.length)<=0?[]:n==null||(F=n.data)===null||F===void 0?void 0:F.list,success:(n==null?void 0:n.code)===200,total:n==null||(o=n.data)===null||o===void 0?void 0:o.total});case 5:case"end":return i.stop()}},c)})),K.apply(this,arguments)}var A=function(I){return function(u){var g,b;return v(v({},"".concat(I,"_start"),u!=null&&(g=u[0])!==null&&g!==void 0&&g.format?u[0].format("YYYY-MM-DD HH:mm:ss"):u==null?void 0:u[0]),"".concat(I,"_end"),u!=null&&(b=u[1])!==null&&b!==void 0&&b.format?u[1].format("YYYY-MM-DD HH:mm:ss"):u==null?void 0:u[1])}}}),41162:(function(l,m,e){var v=e(44814);function y(x){if(Array.isArray(x))return v(x)}l.exports=y,l.exports.__esModule=!0,l.exports.default=l.exports}),33032:(function(l){function m(e){if(typeof Symbol!="undefined"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}l.exports=m,l.exports.__esModule=!0,l.exports.default=l.exports}),30922:(function(l){function m(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}l.exports=m,l.exports.__esModule=!0,l.exports.default=l.exports}),84169:(function(l){function m(e,v){return v||(v=e.slice(0)),Object.freeze(Object.defineProperties(e,{raw:{value:Object.freeze(v)}}))}l.exports=m,l.exports.__esModule=!0,l.exports.default=l.exports}),45647:(function(l,m,e){var v=e(41162),y=e(33032),x=e(34911),E=e(30922);function L(j){return v(j)||y(j)||x(j)||E()}l.exports=L,l.exports.__esModule=!0,l.exports.default=l.exports})}]);
