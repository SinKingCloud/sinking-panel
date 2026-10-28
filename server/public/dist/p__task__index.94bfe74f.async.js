(self.webpackChunk=self.webpackChunk||[]).push([[306],{72955:(function(x,Y,s){"use strict";s.r(Y),s.d(Y,{default:function(){return yn}});var J=s(1974),j=s.n(J),T=s(11604),v=s.n(T),t=s(96540),D=s(62885),ne=s(9300),se=s(9752),ye=s(8490),Ce=s(55420),Z=s.n(Ce),q=s(98312),Se=s.n(q),Re=s(45647),ge=s.n(Re),Ve=s(84169),ae=s.n(Ve),ue=s(79273),C=s(20873),Ue=s(45478),ke=s(47152),ie=s(16370),ve=s(22817),xe=s(58679),Ae=s(12185),Ge=s(47577),Xe=s(42736),xn=s(34241),be=s(14499),bn=s(63859),Me=s.n(bn),Pe=function(e,a,r){if(/^\d+$/.test(e)){var i=Number(e);return i>=a&&i<=r?i:void 0}},We=function(e,a){var r=e.match(/^\*\/([1-9]\d*)$/);if(r){var i=Number(r[1]);return i<=a?i:void 0}},Je=function(e){var a=String(e||"").trim().split(/\s+/);if(a.length!==6)return{mode:"custom"};var r=v()(a,6),i=r[0],m=r[1],f=r[2],y=r[3],c=r[4],l=r[5],p=i==="*"?1:We(i,59);if(p&&m==="*"&&f==="*"&&y==="*"&&c==="*"&&l==="*")return{mode:"second",interval:p};var W=m==="*"?1:We(m,59);if(i==="0"&&W&&f==="*"&&y==="*"&&c==="*"&&l==="*")return{mode:"minute",interval:W};var w=Pe(m,0,59),z=f==="*"?1:We(f,23);if(i==="0"&&w!==void 0&&z&&y==="*"&&c==="*"&&l==="*")return{mode:"hour",interval:z,minute:w};var u=Pe(f,0,23);if(i!=="0"||w===void 0||u===void 0||c!=="*")return{mode:"custom"};if(y==="*"&&l==="*")return{mode:"day",minute:w,hour:u};var R=Pe(l,0,6);if(y==="*"&&R!==void 0)return{mode:"week",minute:w,hour:u,weekday:R};var U=Pe(y,1,31);return U!==void 0&&l==="*"?{mode:"month",minute:w,hour:u,day:U}:{mode:"custom"}},Sn=["\u5468\u65E5","\u5468\u4E00","\u5468\u4E8C","\u5468\u4E09","\u5468\u56DB","\u5468\u4E94","\u5468\u516D"],qe=function(e){return String(Number(e||0)).padStart(2,"0")},_e=function(e){var a="".concat(qe(e.hour),":").concat(qe(e.minute));switch(e.mode){case"second":return"\u6BCF ".concat(e.interval," \u79D2\u6267\u884C\u4E00\u6B21");case"minute":return"\u6BCF ".concat(e.interval," \u5206\u949F\u6267\u884C\u4E00\u6B21");case"hour":return e.minute?"\u6BCF ".concat(e.interval," \u5C0F\u65F6\u7684\u7B2C ").concat(e.minute," \u5206\u949F\u6267\u884C"):"\u6BCF ".concat(e.interval," \u5C0F\u65F6\u6267\u884C\u4E00\u6B21");case"day":return"\u6BCF\u5929 ".concat(a," \u6267\u884C");case"week":return"\u6BCF".concat(Sn[e.weekday||0]," ").concat(a," \u6267\u884C");case"month":return"\u6BCF\u6708 ".concat(e.day," \u65E5 ").concat(a," \u6267\u884C");default:return"\u81EA\u5B9A\u4E49\u6267\u884C\u5468\u671F"}},Cn=function(e){return _e(Je(e))},$e=function(e){var a=arguments.length>1&&arguments[1]!==void 0?arguments[1]:!1,r=e;if(typeof e=="string")try{r=JSON.parse(e)}catch(y){if(a)throw new Error("HTTP \u8BF7\u6C42\u811A\u672C\u5FC5\u987B\u662F JSON \u683C\u5F0F");r={url:e}}if(a){if(!r||Me()(r)!=="object"||Array.isArray(r)||typeof r.url!="string"||r.method!==void 0&&(typeof r.method!="string"||!/^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/.test(r.method.trim()))||r.body!==void 0&&typeof r.body!="string")throw new Error("HTTP \u8BF7\u6C42\u811A\u672C\u683C\u5F0F\u4E0D\u6B63\u786E");try{var i=new URL(r.url.trim());if(i.protocol!=="http:"&&i.protocol!=="https:"||!i.hostname)throw new Error}catch(y){throw new Error("\u8BF7\u8F93\u5165\u6709\u6548\u7684 HTTP \u6216 HTTPS \u5730\u5740")}if(r.headers!==void 0&&(!r.headers||Me()(r.headers)!=="object"||Array.isArray(r.headers)||Object.keys(r.headers).length>100||Object.entries(r.headers).some(function(y){var c=v()(y,2),l=c[0],p=c[1];return!/^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/.test(l)||typeof p!="string"})))throw new Error("HTTP \u8BF7\u6C42\u5934\u683C\u5F0F\u4E0D\u6B63\u786E")}(!r||Me()(r)!=="object"||Array.isArray(r))&&(r={});var m=r.headers&&Me()(r.headers)==="object"&&!Array.isArray(r.headers)?Object.entries(r.headers).map(function(y){var c=v()(y,2),l=c[0],p=c[1];return{name:l,value:String(p!=null?p:"")}}):[],f={method:String(r.method||"GET").trim().toUpperCase(),url:String(r.url||""),headers:m,body:String(r.body||"")};return a&&Ze(f),f},Ze=function(e){var a,r={},i=new Set;return((e==null?void 0:e.headers)||[]).forEach(function(m){var f,y=String((m==null?void 0:m.name)||"").trim(),c=y.toLowerCase();if(!y)throw new Error("\u8BF7\u6C42\u5934\u540D\u79F0\u4E0D\u80FD\u4E3A\u7A7A");if(i.has(c))throw new Error("\u8BF7\u6C42\u5934\u540D\u79F0\u4E0D\u80FD\u91CD\u590D");i.add(c);var l=String((f=m==null?void 0:m.value)!==null&&f!==void 0?f:"");if(c==="content-length"||c==="transfer-encoding"||c==="trailer")throw new Error("\u4E0D\u652F\u6301\u81EA\u5B9A\u4E49\u4F20\u8F93\u5C42\u8BF7\u6C42\u5934");if(c==="host"&&!l.trim())throw new Error("Host\u8BF7\u6C42\u5934\u4E0D\u80FD\u4E3A\u7A7A");if(/\r|\n/.test(l))throw new Error("\u8BF7\u6C42\u5934\u503C\u4E0D\u80FD\u6362\u884C");if(/^[^\u0000-\u0008\u000A-\u001F\u007F]*$/.test(l)===!1)throw new Error("\u8BF7\u6C42\u5934\u503C\u5305\u542B\u975E\u6CD5\u5B57\u7B26");r[y]=l}),JSON.stringify({method:String((e==null?void 0:e.method)||"GET").trim().toUpperCase(),url:String((e==null?void 0:e.url)||"").trim(),headers:r,body:String((a=e==null?void 0:e.body)!==null&&a!==void 0?a:"")})},Qe=s(50400),n=s(74848),wn=["GET","POST","PUT","PATCH","DELETE","HEAD","OPTIONS"].map(function(h){return{label:h,value:h}}),jn=function(){return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsxs)(ke.A,{gutter:[12,0],children:[(0,n.jsx)(ie.A,{xs:24,sm:7,children:(0,n.jsx)(C.A.Item,{name:["request","method"],label:"\u8BF7\u6C42\u65B9\u5F0F",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u8BF7\u6C42\u65B9\u5F0F"},{pattern:/^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/,message:"\u8BF7\u6C42\u65B9\u5F0F\u683C\u5F0F\u4E0D\u5408\u6CD5"}],children:(0,n.jsx)(xe.A,{options:wn})})}),(0,n.jsx)(ie.A,{xs:24,sm:17,children:(0,n.jsx)(C.A.Item,{name:["request","url"],label:"\u8BF7\u6C42\u5730\u5740",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u8BF7\u6C42\u5730\u5740"},{validator:(function(){var e=Se()(Z()().mark(function r(i,m){var f;return Z()().wrap(function(c){for(;;)switch(c.prev=c.next){case 0:if(String(m||"").trim()){c.next=2;break}return c.abrupt("return");case 2:if(c.prev=2,f=new URL(String(m).trim()),!((f.protocol==="http:"||f.protocol==="https:")&&f.hostname)){c.next=6;break}return c.abrupt("return");case 6:c.next=11;break;case 8:throw c.prev=8,c.t0=c.catch(2),new Error("\u8BF7\u8F93\u5165\u6709\u6548\u7684HTTP\u6216HTTPS\u5730\u5740");case 11:throw new Error("\u8BF7\u8F93\u5165\u6709\u6548\u7684HTTP\u6216HTTPS\u5730\u5740");case 12:case"end":return c.stop()}},r,null,[[2,8]])}));function a(r,i){return e.apply(this,arguments)}return a})()}],children:(0,n.jsx)(ve.A,{placeholder:"https://example.com/api"})})})]}),(0,n.jsx)(C.A.Item,{label:"\u8BF7\u6C42\u5934",children:(0,n.jsx)(C.A.List,{name:["request","headers"],rules:[{validator:(function(){var e=Se()(Z()().mark(function r(i){var m,f,y=arguments;return Z()().wrap(function(l){for(;;)switch(l.prev=l.next){case 0:if(m=y.length>1&&y[1]!==void 0?y[1]:[],!(m.length>100)){l.next=3;break}throw new Error("\u8BF7\u6C42\u5934\u6570\u91CF\u4E0D\u80FD\u8D85\u8FC7100\u4E2A");case 3:if(f=m.map(function(p){return String((p==null?void 0:p.name)||"").trim().toLowerCase()}).filter(Boolean),new Set(f).size===f.length){l.next=6;break}throw new Error("\u8BF7\u6C42\u5934\u540D\u79F0\u4E0D\u80FD\u91CD\u590D");case 6:if(!f.some(function(p){return p==="content-length"||p==="transfer-encoding"||p==="trailer"})){l.next=8;break}throw new Error("\u4E0D\u652F\u6301\u81EA\u5B9A\u4E49\u4F20\u8F93\u5C42\u8BF7\u6C42\u5934");case 8:if(!m.some(function(p){return String((p==null?void 0:p.name)||"").trim().toLowerCase()==="host"&&!String((p==null?void 0:p.value)||"").trim()})){l.next=10;break}throw new Error("Host\u8BF7\u6C42\u5934\u4E0D\u80FD\u4E3A\u7A7A");case 10:case"end":return l.stop()}},r)}));function a(r){return e.apply(this,arguments)}return a})()}],children:function(a,r,i){var m=r.add,f=r.remove,y=i.errors;return(0,n.jsxs)("div",{children:[a.map(function(c){return(0,n.jsxs)(ke.A,{gutter:[8,0],wrap:!1,children:[(0,n.jsx)(ie.A,{flex:"1 1 0",children:(0,n.jsx)(C.A.Item,{name:[c.name,"name"],rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u540D\u79F0"},{pattern:/^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/,message:"\u540D\u79F0\u683C\u5F0F\u4E0D\u5408\u6CD5"}],children:(0,n.jsx)(ve.A,{placeholder:"\u540D\u79F0"})})}),(0,n.jsx)(ie.A,{flex:"1 1 0",children:(0,n.jsx)(C.A.Item,{name:[c.name,"value"],rules:[{pattern:/^[^\u0000-\u0008\u000A-\u001F\u007F]*$/,message:"\u8BF7\u6C42\u5934\u503C\u5305\u542B\u975E\u6CD5\u5B57\u7B26"}],children:(0,n.jsx)(ve.A,{placeholder:"\u503C"})})}),(0,n.jsx)(ie.A,{flex:"none",children:(0,n.jsx)(Qe.A,{type:"text",danger:!0,icon:(0,n.jsx)(D.In,{type:"DeleteOutlined"}),"aria-label":"\u5220\u9664\u8BF7\u6C42\u5934",onClick:function(){return f(c.name)}})})]},c.key)}),(0,n.jsx)(Qe.A,{type:"dashed",block:!0,icon:(0,n.jsx)(D.In,{type:"PlusOutlined"}),onClick:function(){return m({name:"",value:""})},children:"\u6DFB\u52A0\u8BF7\u6C42\u5934"}),(0,n.jsx)(C.A.ErrorList,{errors:y})]})}})}),(0,n.jsx)(C.A.Item,{name:["request","body"],label:"\u8BF7\u6C42\u4F53",children:(0,n.jsx)(ve.A.TextArea,{autoSize:{minRows:3,maxRows:8},placeholder:"\u8BF7\u8F93\u5165\u8BF7\u6C42\u5185\u5BB9"})})]})},Tn=jn,Fe=s(48212),Ye=s(15622),en=s(24816),kn=s(74353),An=s.n(kn),nn,tn,rn,an,Fn=[{label:"\u6BCF\u79D2",value:"second"},{label:"\u6BCF\u5206",value:"minute"},{label:"\u6BCF\u5C0F\u65F6",value:"hour"},{label:"\u6BCF\u5929",value:"day"},{label:"\u6BCF\u5468",value:"week"},{label:"\u6BCF\u6708",value:"month"},{label:"\u81EA\u5B9A\u4E49",value:"custom"}],En=[{label:"\u5468\u4E00",value:1},{label:"\u5468\u4E8C",value:2},{label:"\u5468\u4E09",value:3},{label:"\u5468\u56DB",value:4},{label:"\u5468\u4E94",value:5},{label:"\u5468\u516D",value:6},{label:"\u5468\u65E5",value:0}],un={second:"*/5 * * * * *",minute:"0 */5 * * * *",hour:"0 0 */1 * * *",day:"0 0 2 * * *",week:"0 0 2 * * 1",month:"0 0 2 1 * *"},In=(0,Ae.rU)(function(h){var e=h.css,a=h.token,r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{},i=!!(r!=null&&r.isCompactMode);return{schedule:e(nn||(nn=ae()([`
        min-width: 0;
        container-type: inline-size;
    `]))),editor:e(tn||(tn=ae()([`
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 8px;

        .cycle-type {
            width: 132px;
            flex: none;
        }

        .cycle-value {
            min-width: 0;
            flex: 1;
        }

        .schedule-number {
            width: 96px;
            flex: none;
        }

        .schedule-minute {
            width: 88px;
            flex: none;
        }

        .schedule-time {
            width: 120px;
            flex: none;
        }

        .schedule-weekday {
            width: 104px;
            flex: none;
        }

        @container (max-width: 520px) {
            flex-direction: column;
            align-items: stretch;

            .cycle-type {
                width: 100%;
            }

            .schedule-number,
            .schedule-minute {
                width: 60px;
            }

            .schedule-time {
                width: 88px;
            }

            .schedule-weekday {
                width: 74px;
            }
        }
    `]))),inlineControl:e(rn||(rn=ae()([`
        min-width: 0;
        min-height: `,`px;
        display: flex;
        align-items: center;
        gap: 8px;
        white-space: nowrap;

        > span {
            flex: none;
        }

        @container (max-width: 520px) {
            gap: 3px;
            flex-wrap: wrap;
            font-size: 12px;
            white-space: nowrap;

            .schedule-number,
            .schedule-minute,
            .schedule-time,
            .schedule-weekday {
                flex: none;
            }
        }
    `])),a.controlHeight),preview:e(an||(an=ae()([`
        min-width: 0;
        min-height: `,`px;
        margin-top: `,`px;
        padding: `,`;
        box-sizing: border-box;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: `,`px;
        border-radius: `,`px;
        background: `,`;

        .schedule-description.ant-typography {
            min-width: 0;
            flex: 1 1 auto;
            margin: 0;
            display: inline-flex;
            align-items: center;
            gap: `,`px;
            overflow: hidden;
            color: `,`;
            font-size: `,`px;
            line-height: 20px;
            white-space: nowrap;
            text-overflow: ellipsis;
        }

        .schedule-description .anticon {
            flex: none;
            color: `,`;
            font-size: `,`px;
        }

        .schedule-expression.ant-typography {
            min-width: 0;
            flex: 0 1 auto;
            margin: 0;
            padding-inline-start: `,`px;
            display: inline-flex;
            align-items: center;
            gap: `,`px;
            overflow: hidden;
            border-inline-start: 1px solid `,`;
            color: `,`;
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: `,`px;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .schedule-expression .expression-value {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .schedule-expression .ant-typography-copy {
            flex: none;
            margin-inline-start: 0;
            color: `,`;
            font-family: `,`;
            transition: color .16s ease;
        }

        .schedule-expression .ant-typography-copy:hover {
            color: `,`;
        }

    `])),i?32:38,i?6:8,i?"6px 9px":"8px 11px",i?10:12,a.borderRadius,a.colorFillQuaternary,i?6:7,a.colorTextSecondary,i?11:12,a.colorTextTertiary,i?11:12,i?10:12,i?5:6,a.colorSplit,a.colorTextTertiary,i?11:12,a.colorTextTertiary,a.fontFamily,a.colorPrimary)}}),sn=(0,t.forwardRef)(function(h,e){var a,r,i=h.value,m=h.onChange,f=(0,D.DP)(),y=(f==null||(a=f.isCompactTheme)===null||a===void 0?void 0:a.call(f))||!1,c=In({isCompactMode:y}),l=c.styles,p=(0,t.useState)(function(){return Je(i).mode}),W=v()(p,2),w=W[0],z=W[1],u=(0,t.useMemo)(function(){return Je(i)},[i]),R=(0,t.useMemo)(function(){return _e(u)},[u]),U=(0,t.useMemo)(function(){return An()().startOf("day").hour(u.hour||0).minute(u.minute||0)},[u.hour,u.minute]);(0,t.useEffect)(function(){z(u.mode)},[u.mode]);var P=function(o){return m==null?void 0:m(o)},M=function(o){z(o),un[o]&&P(un[o])},ee=function(o,S){if(o){var G=o.minute(),X=o.hour();if(S==="day")P("0 ".concat(G," ").concat(X," * * *"));else if(S==="week"){var g;P("0 ".concat(G," ").concat(X," * * ").concat((g=u.weekday)!==null&&g!==void 0?g:1))}else P("0 ".concat(G," ").concat(X," ").concat(u.day||1," * *"))}};return(0,n.jsxs)("div",{ref:e,className:l.schedule,children:[(0,n.jsxs)("div",{className:l.editor,children:[(0,n.jsx)(xe.A,{className:"cycle-type",value:w,options:Fn,onChange:function(o){return M(String(o))}}),(0,n.jsxs)("div",{className:"cycle-value",children:[w==="second"&&(0,n.jsxs)("div",{className:l.inlineControl,children:[(0,n.jsx)("span",{children:"\u6BCF"}),(0,n.jsx)(Fe.A,{className:"schedule-number",min:1,max:59,value:u.interval||5,onChange:function(o){return P("*/".concat(o||1," * * * * *"))}}),(0,n.jsx)("span",{children:"\u79D2"})]}),w==="minute"&&(0,n.jsxs)("div",{className:l.inlineControl,children:[(0,n.jsx)("span",{children:"\u6BCF"}),(0,n.jsx)(Fe.A,{className:"schedule-number",min:1,max:59,value:u.interval||5,onChange:function(o){return P("0 */".concat(o||1," * * * *"))}}),(0,n.jsx)("span",{children:"\u5206\u949F"})]}),w==="hour"&&(0,n.jsxs)("div",{className:l.inlineControl,children:[(0,n.jsx)("span",{children:"\u6BCF"}),(0,n.jsx)(Fe.A,{className:"schedule-number",min:1,max:23,value:u.interval||1,onChange:function(o){return P("0 ".concat(u.minute||0," */").concat(o||1," * * *"))}}),(0,n.jsx)("span",{children:"\u5C0F\u65F6"}),(0,n.jsx)(Fe.A,{className:"schedule-minute",min:0,max:59,value:u.minute||0,onChange:function(o){return P("0 ".concat(o||0," */").concat(u.interval||1," * * *"))}}),(0,n.jsx)("span",{children:"\u5206\u949F"})]}),w==="day"&&(0,n.jsxs)("div",{className:l.inlineControl,children:[(0,n.jsx)("span",{children:"\u6BCF\u5929"}),(0,n.jsx)(Ye.A,{className:"schedule-time",value:U,format:"HH:mm",minuteStep:1,allowClear:!1,onChange:function(o){return ee(o,"day")}})]}),w==="week"&&(0,n.jsxs)("div",{className:l.inlineControl,children:[(0,n.jsx)("span",{children:"\u6BCF\u5468"}),(0,n.jsx)(xe.A,{className:"schedule-weekday",value:(r=u.weekday)!==null&&r!==void 0?r:1,options:En,onChange:function(o){return P("0 ".concat(u.minute||0," ").concat(u.hour||0," * * ").concat(o))}}),(0,n.jsx)(Ye.A,{className:"schedule-time",value:U,format:"HH:mm",minuteStep:1,allowClear:!1,onChange:function(o){return ee(o,"week")}})]}),w==="month"&&(0,n.jsxs)("div",{className:l.inlineControl,children:[(0,n.jsx)("span",{children:"\u6BCF\u6708"}),(0,n.jsx)(Fe.A,{className:"schedule-number",min:1,max:31,value:u.day||1,onChange:function(o){return P("0 ".concat(u.minute||0," ").concat(u.hour||0," ").concat(o||1," * *"))}}),(0,n.jsx)("span",{children:"\u65E5"}),(0,n.jsx)(Ye.A,{className:"schedule-time",value:U,format:"HH:mm",minuteStep:1,allowClear:!1,onChange:function(o){return ee(o,"month")}})]}),w==="custom"&&(0,n.jsx)(ve.A,{value:String(i||""),placeholder:"\u79D2 \u5206 \u65F6 \u65E5 \u6708 \u661F\u671F",onChange:function(o){return P(o.target.value)}})]})]}),(0,n.jsxs)("div",{className:l.preview,children:[(0,n.jsxs)(en.A.Text,{type:"secondary",className:"schedule-description",children:[(0,n.jsx)(D.In,{type:"ClockCircleOutlined"}),R]}),(0,n.jsx)(en.A.Text,{type:"secondary",className:"schedule-expression",copyable:{text:String(i||"")},children:(0,n.jsx)("span",{className:"expression-value",children:String(i||"-")})})]})]})});sn.displayName="Schedule";var Nn=(0,t.memo)(sn),on,ln,cn,dn,On="".concat((Ge.A===null||Ge.A===void 0?void 0:Ge.A.basePath)||"/","ace"),Bn=[{label:"\u8FD0\u884C",value:0},{label:"\u6682\u505C",value:1}],Dn=(0,Ae.rU)(function(h){var e=h.css,a=h.token;return{modal:e(on||(on=ae()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }
    `]))),editor:e(ln||(ln=ae()([`
        overflow: hidden;
        border: 1px solid `,`;
        border-radius: `,`px;

        .ace_editor {
            border-radius: `,`px !important;
        }
    `])),a.colorBorderSecondary,a.borderRadius,a.borderRadius),loading:e(cn||(cn=ae()([`
        height: 380px;
        display: flex;
        align-items: center;
        justify-content: center;
    `]))),scriptToolbar:e(dn||(dn=ae()([`
        display: flex;
        justify-content: flex-end;
        margin-bottom: 8px;
    `])))}}),pn=(0,t.forwardRef)(function(h,e){var a,r,i=h.typeData,m=h.typeItems,f=h.execTypeData,y=h.onSuccess,c=ue.A.useApp(),l=c.message,p=(0,D.DP)(),W=!!(!(p==null||(a=p.isCompactTheme)===null||a===void 0)&&a.call(p)),w=Dn(),z=w.styles,u=(0,t.useRef)({}),R=(0,t.useRef)(0),U=C.A.useForm(),P=v()(U,1),M=P[0],ee=(0,t.useState)(),b=v()(ee,2),o=b[0],S=b[1],G=(0,t.useState)(0),X=v()(G,2),g=X[0],F=X[1],L=(0,t.useState)(!1),k=v()(L,2),I=k[0],Q=k[1],oe=(0,t.useState)(!1),Le=v()(oe,2),Ee=Le[0],we=Le[1],He=(0,t.useState)(!1),ze=v()(He,2),Ie=ze[0],me=ze[1],te=o!=null,Ke=(0,t.useMemo)(function(){return[{label:"\u5168\u90E8\u5206\u7C7B",value:0}].concat(ge()(m?m.map(function(A){return{label:String(A.name),value:Number(A.id)}}):Object.entries(i||{}).filter(function(A){var E=v()(A,1),N=E[0];return N!=="0"}).map(function(A){var E=v()(A,2),N=E[0],$=E[1];return{label:String($),value:Number(N)}})))},[i,m]),fe=(0,t.useMemo)(function(){return Object.entries(f||{}).map(function(A){var E=v()(A,2),N=E[0],$=E[1];return{label:String($),value:Number(N)}})},[f]),Ne=(0,t.useCallback)(function(){R.current+=1,M.resetFields(),S(void 0),F(0),Q(!1),we(!1),me(!1)},[M]),je=(0,t.useCallback)(function(){var A,E,N=Number((A=fe[0])===null||A===void 0?void 0:A.value);if(!Number.isFinite(N)){l.error("\u6267\u884C\u65B9\u5F0F\u5C1A\u672A\u52A0\u8F7D");return}R.current+=1,S(void 0),F(N),Q(!0),we(!1),M.resetFields(),M.setFieldsValue({name:"",type_id:0,exec_type:N,spec:"0 */5 * * * *",script:"",request:$e({method:"GET"}),status:0}),(E=u.current)===null||E===void 0||E.show()},[fe,M,l]),Te=(0,t.useCallback)(function(A){var E;if(!((A==null?void 0:A.id)===void 0||(A==null?void 0:A.id)===null)){if(fe.length===0){l.error("\u6267\u884C\u65B9\u5F0F\u5C1A\u672A\u52A0\u8F7D");return}var N=++R.current;S(A.id),F(Number(A.exec_type||0)),Q(!0),we(!0),M.resetFields(),(E=u.current)===null||E===void 0||E.show(),(0,be.Mg)({body:{id:A.id}}).then(function($){if(R.current===N){if(($==null?void 0:$.code)!==200){var le;(le=u.current)===null||le===void 0||le.hide(),l.error(($==null?void 0:$.message)||"\u83B7\u53D6\u4EFB\u52A1\u8BE6\u60C5\u5931\u8D25");return}var B=($==null?void 0:$.data)||{},ce=Number((B==null?void 0:B.exec_type)||0);F(ce),M.setFieldsValue({name:B==null?void 0:B.name,type_id:Number((B==null?void 0:B.type_id)||0),exec_type:ce,spec:B==null?void 0:B.spec,script:ce===0?String((B==null?void 0:B.script)||""):"",request:ce===1?$e(B==null?void 0:B.script):void 0,status:Number(B==null?void 0:B.status)})}}).catch(function(){if(R.current===N){var $;($=u.current)===null||$===void 0||$.hide(),l.error("\u83B7\u53D6\u4EFB\u52A1\u8BE6\u60C5\u5931\u8D25")}}).finally(function(){R.current===N&&we(!1)})}},[fe.length,M,l]);(0,t.useImperativeHandle)(e,function(){return{open:function(E){return E?Te(E):je()}}},[je,Te]);var Oe=(0,t.useCallback)((function(){var A=Se()(Z()().mark(function E(N){var $,le,B,ce,Be,re;return Z()().wrap(function(d){for(;;)switch(d.prev=d.next){case 0:if($=++R.current,le=Number(N.exec_type),B=String(N.script||""),le!==1){d.next=12;break}d.prev=4,B=Ze(N.request),d.next=12;break;case 8:return d.prev=8,d.t0=d.catch(4),l.error((d.t0===null||d.t0===void 0?void 0:d.t0.message)||"\u8BF7\u6C42\u914D\u7F6E\u4E0D\u5408\u6CD5"),d.abrupt("return");case 12:return ce={name:String(N.name||"").trim(),type_id:te?String(Number(N.type_id||0)):Number(N.type_id||0),exec_type:te?String(le):le,spec:String(N.spec||"").trim(),script:B,status:te?String(Number(N.status)):Number(N.status)},te&&(ce.ids=[o]),me(!0),d.prev=15,d.next=18,(te?be.lC:be.UT)({body:ce});case 18:if(re=d.sent,R.current===$){d.next=21;break}return d.abrupt("return");case 21:if((re==null?void 0:re.code)===200){d.next=24;break}return l.error((re==null?void 0:re.message)||"\u64CD\u4F5C\u5931\u8D25"),d.abrupt("return");case 24:(Be=u.current)===null||Be===void 0||Be.hide(),l.success((re==null?void 0:re.message)||"\u64CD\u4F5C\u6210\u529F"),y==null||y(),d.next=32;break;case 29:d.prev=29,d.t1=d.catch(15),R.current===$&&l.error("\u64CD\u4F5C\u5931\u8D25");case 32:return d.prev=32,R.current===$&&me(!1),d.finish(32);case 35:case"end":return d.stop()}},E,null,[[4,8],[15,29,32,35]])}));return function(E){return A.apply(this,arguments)}})(),[te,l,y,o]);return(0,n.jsx)(D.Ht,{ref:u,title:(0,n.jsx)(D.hE,{children:te?"\u7F16\u8F91\u4EFB\u52A1":"\u6DFB\u52A0\u4EFB\u52A1"}),onOk:M.submit,width:680,modalProps:{rootClassName:z.modal,confirmLoading:Ie||Ee,forceRender:!0,okText:te?"\u4FDD\u5B58":"\u521B\u5EFA",cancelText:"\u53D6\u6D88",footer:Ee?null:void 0,style:{top:100,paddingBottom:100},mask:{closable:!0},afterClose:Ne},children:(0,n.jsx)(C.A,{form:M,layout:"vertical",variant:"filled",onFinish:Oe,children:I&&(Ee?(0,n.jsx)("div",{className:z.loading,children:(0,n.jsx)(Ue.A,{})}):(0,n.jsxs)(n.Fragment,{children:[(0,n.jsxs)(ke.A,{gutter:[16,0],children:[(0,n.jsx)(ie.A,{xs:24,sm:12,children:(0,n.jsx)(C.A.Item,{name:"name",label:"\u4EFB\u52A1\u540D\u79F0",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u4EFB\u52A1\u540D\u79F0"}],children:(0,n.jsx)(ve.A,{placeholder:"\u8BF7\u8F93\u5165\u4EFB\u52A1\u540D\u79F0"})})}),(0,n.jsx)(ie.A,{xs:24,sm:12,children:(0,n.jsx)(C.A.Item,{name:"type_id",label:"\u4EFB\u52A1\u5206\u7C7B",children:(0,n.jsx)(xe.A,{placeholder:"\u5168\u90E8\u5206\u7C7B",options:Ke})})}),(0,n.jsx)(ie.A,{xs:24,sm:12,children:(0,n.jsx)(C.A.Item,{name:"exec_type",label:"\u6267\u884C\u65B9\u5F0F",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u6267\u884C\u65B9\u5F0F"}],children:(0,n.jsx)(xe.A,{options:fe,onChange:function(E){var N=Number(E);F(N),M.setFieldsValue({script:"",request:N===1?$e({method:"GET"}):void 0})}})})}),(0,n.jsx)(ie.A,{xs:24,sm:12,children:(0,n.jsx)(C.A.Item,{name:"status",label:"\u4EFB\u52A1\u72B6\u6001",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u4EFB\u52A1\u72B6\u6001"}],children:(0,n.jsx)(xe.A,{options:Bn})})})]}),(0,n.jsx)(C.A.Item,{name:"spec",label:"\u6267\u884C\u5468\u671F",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8BBE\u7F6E\u6267\u884C\u5468\u671F"}],children:(0,n.jsx)(Nn,{})}),(g===0||g===1)&&(0,n.jsx)("div",{className:z.scriptToolbar,children:(0,n.jsx)(xn.A,{module:g===0?"ssh":"request",mode:g===0?"sh":"json",Editor:Xe.A,disabled:Ie,value:function(){if(g===0)return String(M.getFieldValue("script")||"");try{return JSON.stringify(JSON.parse(Ze(M.getFieldValue("request"))),null,2)}catch(E){return""}},onUse:function(E){g===0?M.setFieldValue("script",E):M.setFieldsValue({request:$e(E,!0)})}},g)}),g===0?(0,n.jsx)(C.A.Item,{name:"script",label:"\u4EFB\u52A1\u5185\u5BB9",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u4EFB\u52A1\u5185\u5BB9"}],children:(0,n.jsx)(Xe.A,{mode:"sh",theme:p!=null&&(r=p.isDarkTheme)!==null&&r!==void 0&&r.call(p)?"monokai":"chrome",width:"100%",height:280,fontSize:W?12:14,showPrintMargin:!1,wrapEnabled:!0,acePath:On,className:z.editor})}):g===1?(0,n.jsx)(Tn,{}):(0,n.jsx)(C.A.Item,{name:"script",label:"\u4EFB\u52A1\u5185\u5BB9",rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u4EFB\u52A1\u5185\u5BB9"}],children:(0,n.jsx)(ve.A,{placeholder:"\u8BF7\u8F93\u5165\u4EFB\u52A1\u5185\u5BB9",maxLength:2e3})})]}))})})});pn.displayName="Form";var Rn=(0,t.memo)(pn),Mn=function(e){var a,r,i,m=e.keyword,f=e.typeId,y=e.execType,c=e.status,l=e.typeData,p=e.typeItems,W=e.execTypeData,w=e.statusData,z=e.refreshing,u=e.onKeywordChange,R=e.onSearch,U=e.onTypeIdChange,P=e.onManageTypes,M=e.onExecTypeChange,ee=e.onStatusChange,b=e.onRefresh,o=e.onCreate,S=t.useMemo(function(){return[{label:"\u5168\u90E8\u5206\u7C7B",value:"0"}].concat(ge()(p?p.map(function(k){return{value:String(k.id),label:String(k.name)}}):Object.entries(l||{}).filter(function(k){var I=v()(k,1),Q=I[0];return Q!=="0"}).map(function(k){var I=v()(k,2),Q=I[0],oe=I[1];return{value:Q,label:String(oe)}})))},[l,p]),G=t.useMemo(function(){return[{label:"\u5168\u90E8\u65B9\u5F0F",value:""}].concat(ge()(Object.entries(W||{}).map(function(k){var I=v()(k,2),Q=I[0],oe=I[1];return{value:Q,label:String(oe)}})))},[W]),X=t.useMemo(function(){return[{label:"\u5168\u90E8\u72B6\u6001",value:""}].concat(ge()(Object.entries(w||{}).map(function(k){var I=v()(k,2),Q=I[0],oe=I[1];return{value:Q,label:String(oe)}})))},[w]),g=((a=S.find(function(k){return k.value===f}))===null||a===void 0?void 0:a.label)||S[0].label,F=((r=G.find(function(k){return k.value===y}))===null||r===void 0?void 0:r.label)||G[0].label,L=((i=X.find(function(k){return k.value===c}))===null||i===void 0?void 0:i.label)||X[0].label;return{hero:{eyebrow:"TASK SCHEDULER",title:"\u8BA1\u5212\u4EFB\u52A1",action:{label:"\u6DFB\u52A0\u4EFB\u52A1",onClick:o}},toolbar:{search:{value:m,placeholder:"\u641C\u7D22\u4EFB\u52A1\u540D\u79F0",onChange:u,onSearch:R},refresh:{loading:z,ariaLabel:"\u5237\u65B0\u8BA1\u5212\u4EFB\u52A1\u5217\u8868",onClick:b},actions:[{key:"category",type:"filter",value:f||"0",options:S,icon:"FolderOutlined",ariaLabel:"\u4EFB\u52A1\u5206\u7C7B\u7B5B\u9009",title:g,command:{label:"\u5206\u7C7B\u7BA1\u7406",icon:"SettingOutlined",onClick:P},onChange:U},{key:"exec-type",type:"filter",value:y||"all",options:G.map(function(k){var I=k.label,Q=k.value;return{label:I,value:Q||"all"}}),icon:"CodeOutlined",ariaLabel:"\u6267\u884C\u65B9\u5F0F\u7B5B\u9009",title:F,onChange:function(I){return M(I==="all"?"":I)}},{key:"status",type:"filter",value:c||"all",options:X.map(function(k){var I=k.label,Q=k.value;return{label:I,value:Q||"all"}}),icon:"FlagOutlined",ariaLabel:"\u4EFB\u52A1\u72B6\u6001\u7B5B\u9009",title:L,onChange:function(I){return ee(I==="all"?"":I)}}]}}},Pn=Mn,$n=s(96885),vn=s(88826),Ln=s(57798),Hn=s(63496),zn=function(e){var a=String(e||"");return!a||a.startsWith("0001-01-01")?"\u5C1A\u672A\u8FD0\u884C":a},Kn=function(e){var a=e.className,r=e.tasks,i=e.loading,m=e.typeData,f=e.execTypeData,y=e.statusData,c=e.sort,l=e.order,p=e.operating,W=e.onSortChange,w=e.onAction,z=e.onEdit,u=e.onLog,R=(0,t.useState)(""),U=v()(R,2),P=U[0],M=U[1],ee=(0,t.useMemo)(function(){return[{title:"\u4EFB\u52A1\u540D\u79F0",dataIndex:"name",key:"name",width:180,className:"task-name-cell",render:function(o,S){return(0,n.jsx)("div",{className:"task-name",children:(0,n.jsx)("button",{type:"button",onClick:function(){return u(S)},children:S.name||"\u672A\u547D\u540D\u4EFB\u52A1"})})}},{title:"\u5206\u7C7B",dataIndex:"type_id",key:"type_id",width:110,className:"category-cell",sorter:!0,sortOrder:c==="type_id"?l==="asc"?"ascend":"descend":null,render:function(o){var S=Number(o)===0?"\u5168\u90E8\u5206\u7C7B":m[String(o)]||"\u672A\u77E5\u5206\u7C7B";return(0,n.jsx)(vn.A,{title:S,children:(0,n.jsx)("span",{className:"task-category",children:S})})}},{title:"\u6267\u884C\u65B9\u5F0F",dataIndex:"exec_type",key:"exec_type",width:100,className:"exec-type-cell",sorter:!0,sortOrder:c==="exec_type"?l==="asc"?"ascend":"descend":null,render:function(o){return(0,n.jsx)("span",{className:"task-exec-type",children:f[String(o)]||"-"})}},{title:"\u6267\u884C\u5468\u671F",dataIndex:"spec",key:"spec",width:150,className:"schedule-cell",render:function(o){return(0,n.jsx)("span",{className:"schedule-value",children:Cn(o)})}},{title:"\u72B6\u6001",dataIndex:"status",key:"status",width:80,className:"status-cell",sorter:!0,sortOrder:c==="status"?l==="asc"?"ascend":"descend":null,render:function(o){var S=Number(o)===0;return(0,n.jsxs)("span",{className:"task-state ".concat(S?"running":"paused"),children:[(0,n.jsx)("i",{}),y[String(o)]||(S?"\u8FD0\u884C":"\u6682\u505C")]})}},{title:"\u6700\u8FD1\u6267\u884C",dataIndex:"run_time",key:"run_time",width:110,className:"runtime-cell",sorter:!0,sortOrder:c==="run_time"?l==="asc"?"ascend":"descend":null,render:function(o){var S=zn(o);return(0,n.jsx)(vn.A,{title:S,children:(0,n.jsx)("span",{className:"runtime-value",children:S==="\u5C1A\u672A\u8FD0\u884C"?S:(0,Hn.K)(S)||S})})}},{title:"\u64CD\u4F5C",key:"action",width:80,fixed:"right",className:"action-cell",render:function(o,S){var G=String(S.id),X=Number(S.status)===0,g=[{key:"edit",label:"\u7F16\u8F91\u4EFB\u52A1",onClick:function(){return z(S)}},{key:"log",label:"\u4EFB\u52A1\u65E5\u5FD7",onClick:function(){return u(S)}},{key:"run",label:"\u7ACB\u5373\u6267\u884C",onClick:function(){return w("run",S)}},X?{key:"stop",label:"\u6682\u505C\u4EFB\u52A1",onClick:function(){return w("stop",S)}}:{key:"restore",label:"\u6062\u590D\u4EFB\u52A1",onClick:function(){return w("restore",S)}},{type:"divider"},{key:"delete",label:"\u5220\u9664\u4EFB\u52A1",danger:!0,onClick:function(){return w("delete",S)}}];return(0,n.jsx)(Ln.A,{open:P===G,onOpenChange:function(L){return M(L?G:"")},menu:{items:g},children:(0,n.jsx)(Qe.A,{size:"small",loading:p.has(G),children:"\u64CD\u4F5C"})})}}]},[f,w,z,u,P,p,l,c,y,m]);return{className:a,columns:ee,dataSource:r,loading:i,onSortChange:W,rowKey:function(o){return String(o.id)},rowClassName:function(o){return P===String(o.id)?"action-menu-open":""}}},Vn=Kn,mn={run:{request:be.ZZ,title:"\u6267\u884C\u4EFB\u52A1",content:function(e){return"\u786E\u5B9A\u7ACB\u5373\u6267\u884C\u4EFB\u52A1\u201C".concat(e,"\u201D\u5417\uFF1F")},okText:"\u6267\u884C",success:"\u4EFB\u52A1\u5DF2\u63D0\u4EA4\u6267\u884C",empty:"\u8BF7\u9009\u62E9\u8981\u6267\u884C\u7684\u4EFB\u52A1",filter:function(){return!0}},stop:{request:be.S8,title:"\u6682\u505C\u4EFB\u52A1",content:function(e){return"\u786E\u5B9A\u6682\u505C\u4EFB\u52A1\u201C".concat(e,"\u201D\u5417\uFF1F")},okText:"\u6682\u505C",success:"\u4EFB\u52A1\u5DF2\u6682\u505C",loading:"\u6B63\u5728\u6682\u505C\u4EFB\u52A1...",empty:"\u8BF7\u9009\u62E9\u8FD0\u884C\u4E2D\u7684\u4EFB\u52A1",filter:function(e){return Number(e==null?void 0:e.status)===0}},restore:{request:be.ec,title:"\u6062\u590D\u4EFB\u52A1",content:function(e){return"\u786E\u5B9A\u6062\u590D\u4EFB\u52A1\u201C".concat(e,"\u201D\u5417\uFF1F")},okText:"\u6062\u590D",success:"\u4EFB\u52A1\u5DF2\u6062\u590D",loading:"\u6B63\u5728\u6062\u590D\u4EFB\u52A1...",empty:"\u8BF7\u9009\u62E9\u5DF2\u6682\u505C\u7684\u4EFB\u52A1",filter:function(e){return Number(e==null?void 0:e.status)===1}},delete:{request:be.vq,title:"\u5220\u9664\u4EFB\u52A1",content:function(e){return"\u786E\u5B9A\u5220\u9664\u4EFB\u52A1\u201C".concat(e,"\u201D\u5417\uFF1F")},okText:"\u786E\u5B9A",success:"\u5220\u9664\u6210\u529F",empty:"\u8BF7\u9009\u62E9\u8981\u5220\u9664\u7684\u4EFB\u52A1",filter:function(){return!0},danger:!0}},Un=function(){var e=ue.A.useApp(),a=e.message,r=e.modal,i=(0,t.useRef)(0),m=(0,t.useRef)(new Set),f=(0,t.useState)([]),y=v()(f,2),c=y[0],l=y[1],p=(0,t.useState)(0),W=v()(p,2),w=W[0],z=W[1],u=(0,t.useState)(""),R=v()(u,2),U=R[0],P=R[1],M=(0,t.useState)({page:1,pageSize:10,name:"",typeId:"0",execType:"",status:"",sort:"id",order:"desc"}),ee=v()(M,2),b=ee[0],o=ee[1],S=(0,t.useState)(!0),G=v()(S,2),X=G[0],g=G[1],F=(0,t.useState)(!1),L=v()(F,2),k=L[0],I=L[1],Q=(0,t.useState)(0),oe=v()(Q,2),Le=oe[0],Ee=oe[1],we=(0,t.useState)(function(){return new Set}),He=v()(we,2),ze=He[0],Ie=He[1],me=b.page,te=b.pageSize,Ke=b.name,fe=b.typeId,Ne=b.execType,je=b.status,Te=b.sort,Oe=b.order;(0,t.useEffect)(function(){var d=++i.current,O=!0,K=!1;return g(!0),(0,be.u3)({body:{page:me,page_size:te,order_by_field:Te,order_by_type:Oe,name:Ke,type_id:fe,exec_type:Ne,status:je}}).then(function(_){if(!(!O||i.current!==d)){if((_==null?void 0:_.code)!==200){a.error((_==null?void 0:_.message)||"\u83B7\u53D6\u8BA1\u5212\u4EFB\u52A1\u5931\u8D25");return}var H=(_==null?void 0:_.data)||{},de=Array.isArray(H==null?void 0:H.list)?H==null?void 0:H.list.filter(Boolean):[],he=Number((H==null?void 0:H.total)||0);if(de.length===0&&he>0&&me>1){K=!0,o(function(De){return j()(j()({},De),{},{page:Math.max(1,Math.min(me-1,Math.ceil(he/te)))})});return}l(de),z(he)}}).catch(function(){O&&i.current===d&&a.error("\u83B7\u53D6\u8BA1\u5212\u4EFB\u52A1\u5931\u8D25")}).finally(function(){O&&i.current===d&&!K&&(g(!1),I(!0))}),function(){O=!1}},[Ne,a,Ke,Oe,me,te,Le,Te,je,fe]);var A=(0,t.useCallback)(function(){i.current+=1,Ee(function(d){return d+1})},[]),E=(0,t.useCallback)(function(d){P(d),d||o(function(O){return j()(j()({},O),{},{page:1,name:""})})},[]),N=(0,t.useCallback)(function(d){o(function(O){return j()(j()({},O),{},{page:1,name:d.trim()})})},[]),$=(0,t.useCallback)(function(d){o(function(O){return j()(j()({},O),{},{page:1,typeId:d})})},[]),le=(0,t.useCallback)(function(d){o(function(O){return j()(j()({},O),{},{page:1,execType:d})})},[]),B=(0,t.useCallback)(function(d){o(function(O){return j()(j()({},O),{},{page:1,status:d})})},[]),ce=(0,t.useCallback)(function(d,O){o(function(K){return j()(j()({},K),{},{page:1,sort:O?d:"id",order:O==="ascend"?"asc":"desc"})})},[]),Be=(0,t.useCallback)(function(d,O){o(function(K){return j()(j()({},K),{},{page:O===K.pageSize?d:1,pageSize:O})})},[]),re=(0,t.useCallback)((function(){var d=Se()(Z()().mark(function O(K,_){var H,de,he,De,pe;return Z()().wrap(function(V){for(;;)switch(V.prev=V.next){case 0:if(H=mn[K],de=_==null?void 0:_.id,he=String(de!=null?de:""),!(!H||de===void 0||de===null||!H.filter(_))){V.next=6;break}return a.warning((H==null?void 0:H.empty)||"\u64CD\u4F5C\u53C2\u6570\u4E0D\u6B63\u786E"),V.abrupt("return");case 6:if(!m.current.has(he)){V.next=8;break}return V.abrupt("return");case 8:return m.current.add(he),Ie(new Set(m.current)),De="task-operation-".concat(K,"-").concat(he),H.loading&&a.loading({key:De,content:H.loading,duration:0}),V.prev=12,V.next=15,H.request({body:{ids:[de]}});case 15:if(pe=V.sent,(pe==null?void 0:pe.code)===200){V.next=19;break}return a.error((pe==null?void 0:pe.message)||"\u64CD\u4F5C\u5931\u8D25"),V.abrupt("return");case 19:if(a.success((pe==null?void 0:pe.message)||H.success),!(K==="stop"||K==="restore")){V.next=23;break}return A(),V.abrupt("return");case 23:K==="delete"&&A(),V.next=29;break;case 26:V.prev=26,V.t0=V.catch(12),a.error("\u64CD\u4F5C\u5931\u8D25");case 29:return V.prev=29,H.loading&&a.destroy(De),m.current.delete(he),Ie(new Set(m.current)),V.finish(29);case 34:case"end":return V.stop()}},O,null,[[12,26,29,34]])}));return function(O,K){return d.apply(this,arguments)}})(),[a,A]),gn=(0,t.useCallback)(function(d,O){var K=mn[d];K&&r.confirm({title:K.title,content:K.content((O==null?void 0:O.name)||"-"),okText:K.okText,cancelText:"\u53D6\u6D88",okButtonProps:K.danger?{danger:!0,type:"default"}:void 0,mask:{closable:!0},onOk:function(){return re(d,O)}})},[re,r]);return{tasks:c,total:w,page:me,pageSize:te,keyword:U,typeId:fe,execType:Ne,status:je,sort:Te,order:Oe,loading:X,initialized:k,operating:ze,reload:A,changeKeyword:E,search:N,changeTypeId:$,changeExecType:le,changeStatus:B,changeSort:ce,changePage:Be,confirmAction:gn}},Gn=Un,fn,Wn=(0,Ae.rU)(function(h){var e=h.css,a=h.token;return{taskTable:e(fn||(fn=ae()([`
        .task-name {
            min-width: 0;
        }

        .task-name > button {
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

        .task-name > button:hover,
        .task-name > button:focus-visible {
            color: `,`;
        }

        .task-state {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            line-height: 20px;
        }

        .task-state i {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: currentColor;
        }

        .task-state.running {
            color: `,`;
        }

        .task-state.paused {
            color: `,`;
        }

        .task-category,
        .task-exec-type,
        .schedule-value,
        .runtime-value {
            display: block;
            min-width: 0;
            overflow: hidden;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
    `])),a.colorPrimary,a.colorSuccess,a.colorWarning)}}),Jn=Wn,hn={},Zn={0:"\u7CFB\u7EDF\u811A\u672C",1:"HTTP\u8BF7\u6C42"},yn=(function(){var h=Jn(),e=h.styles,a=(0,ne.Ay)("task"),r=v()(a,2),i=r[0],m=r[1],f=(0,t.useRef)({}),y=(0,t.useRef)({}),c=(0,t.useRef)({}),l=(0,t.useRef)(0),p=(0,t.useState)(null),W=v()(p,2),w=W[0],z=W[1],u=Gn(),R=i!=null&&i.exec_type&&Object.keys(i.exec_type).length>0?i.exec_type:Zn,U=(0,t.useMemo)(function(){var g;return w===null?(g=i==null?void 0:i.type)!==null&&g!==void 0?g:hn:Object.fromEntries(w.map(function(F){return[String(F.id),F.name]}))},[i==null?void 0:i.type,w]),P=(i==null?void 0:i.status)||hn;(0,t.useEffect)(function(){var g=!0,F=l.current;return(0,ye.Jt)({body:{module:"task",order_by_field:"sort",order_by_type:"asc"}}).then(function(L){var k,I;!g||l.current!==F||(L==null?void 0:L.code)!==200||z(Array.isArray(L==null||(k=L.data)===null||k===void 0?void 0:k.list)?L==null||(I=L.data)===null||I===void 0?void 0:I.list.filter(Boolean):[])}),function(){g=!1}},[]);var M=(0,t.useCallback)(function(){var g;return(g=f.current)===null||g===void 0?void 0:g.open()},[]),ee=(0,t.useCallback)(function(g){var F;return(F=f.current)===null||F===void 0?void 0:F.open(g)},[]),b=(0,t.useCallback)(function(g){var F;return(F=y.current)===null||F===void 0?void 0:F.open(g)},[]),o=(0,t.useCallback)(function(){var g;return(g=c.current)===null||g===void 0?void 0:g.open()},[]),S=(0,t.useCallback)(function(g){l.current+=1;var F=Object.fromEntries(g.map(function(L){return[String(L.id),L.name]}));z(g),u.typeId!=="0"&&!F[u.typeId]&&u.changeTypeId("0")},[u.changeTypeId,u.typeId]),G=Pn({keyword:u.keyword,typeId:u.typeId,execType:u.execType,status:u.status,typeData:U,typeItems:w||void 0,execTypeData:R,statusData:P,refreshing:u.loading,onKeywordChange:u.changeKeyword,onSearch:u.search,onTypeIdChange:u.changeTypeId,onManageTypes:o,onExecTypeChange:u.changeExecType,onStatusChange:u.changeStatus,onRefresh:u.reload,onCreate:M}),X=Vn({className:e.taskTable,tasks:u.tasks,loading:u.loading,typeData:U,execTypeData:R,statusData:P,sort:u.sort,order:u.order,operating:u.operating,onSortChange:u.changeSort,onAction:u.confirmAction,onEdit:ee,onLog:b});return(0,n.jsxs)(D.nB,{loading:m,children:[(0,n.jsx)(D.II,j()(j()(j()({},G),X),{},{ariaLabel:"\u8BA1\u5212\u4EFB\u52A1",empty:u.initialized&&u.tasks.length===0,pagination:{page:u.page,pageSize:u.pageSize,total:u.total,unit:"\u6761",onChange:u.changePage}})),(0,n.jsx)(Rn,{ref:f,typeData:U,typeItems:w||void 0,execTypeData:R,onSuccess:u.reload}),(0,n.jsx)($n.A,{ref:y}),(0,n.jsx)(se.A,{ref:c,module:"task",onChange:S,onMutation:u.reload})]})});s.dn(yn)}),63496:(function(x,Y,s){"use strict";s.d(Y,{K:function(){return J}});function J(j){if(j){var T=j.split(" ",2),v=((T==null?void 0:T[0])||"").replace(/\//g,"-").split("-",3),t=((T==null?void 0:T[1])||"").split(":",3),D=new Date(parseInt(v[0],10)||0,(parseInt(v[1],10)||1)-1,parseInt(v[2],10)||0,parseInt(t[0],10)||0,parseInt(t[1],10)||0,parseInt(t[2],10)||0).getTime(),ne=1e3*60,se=ne*60,ye=se*24,Ce=ye*7,Z=ye*30;if(Number.isFinite(D)){var q=Date.now()-D,Se=q/ne,Re=q/se,ge=q/ye,Ve=q/Ce,ae=q/Z,ue;if(q<=ne)ue="\u521A\u521A";else if(Se<60)ue=parseInt(String(Se))+"\u5206\u949F\u524D";else if(Re<24)ue=parseInt(String(Re))+"\u5C0F\u65F6\u524D";else if(ge<7)ue=parseInt(String(ge))+"\u5929\u524D";else if(ge<30)ue=parseInt(String(Ve))+"\u5468\u524D";else if(ae<6)ue=parseInt(String(ae))+"\u6708\u524D";else{var C=new Date;C.setTime(D);var Ue=C.getFullYear(),ke=C.getMonth()+1<10?"0"+(C.getMonth()+1):C.getMonth()+1,ie=C.getDate()<10?"0"+C.getDate():C.getDate(),ve=C.getHours()<10?"0"+C.getHours():C.getHours(),xe=C.getMinutes()<10?"0"+C.getMinutes():C.getMinutes(),Ae=C.getSeconds()<10?"0"+C.getSeconds():C.getSeconds();ue=Ue+"-"+ke+"-"+ie+" "+ve+":"+xe+":"+Ae}return ue}}}}),41162:(function(x,Y,s){var J=s(44814);function j(T){if(Array.isArray(T))return J(T)}x.exports=j,x.exports.__esModule=!0,x.exports.default=x.exports}),59224:(function(x,Y,s){var J=s(34911);function j(T,v){var t=typeof Symbol!="undefined"&&T[Symbol.iterator]||T["@@iterator"];if(!t){if(Array.isArray(T)||(t=J(T))||v&&T&&typeof T.length=="number"){t&&(T=t);var D=0,ne=function(){};return{s:ne,n:function(){return D>=T.length?{done:!0}:{done:!1,value:T[D++]}},e:function(q){throw q},f:ne}}throw new TypeError(`Invalid attempt to iterate non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}var se=!0,ye=!1,Ce;return{s:function(){t=t.call(T)},n:function(){var q=t.next();return se=q.done,q},e:function(q){ye=!0,Ce=q},f:function(){try{!se&&t.return!=null&&t.return()}finally{if(ye)throw Ce}}}}x.exports=j,x.exports.__esModule=!0,x.exports.default=x.exports}),33032:(function(x){function Y(s){if(typeof Symbol!="undefined"&&s[Symbol.iterator]!=null||s["@@iterator"]!=null)return Array.from(s)}x.exports=Y,x.exports.__esModule=!0,x.exports.default=x.exports}),30922:(function(x){function Y(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}x.exports=Y,x.exports.__esModule=!0,x.exports.default=x.exports}),71640:(function(x,Y,s){var J=s(59276);function j(T,v){if(T==null)return{};var t=J(T,v),D,ne;if(Object.getOwnPropertySymbols){var se=Object.getOwnPropertySymbols(T);for(ne=0;ne<se.length;ne++)D=se[ne],!(v.indexOf(D)>=0)&&Object.prototype.propertyIsEnumerable.call(T,D)&&(t[D]=T[D])}return t}x.exports=j,x.exports.__esModule=!0,x.exports.default=x.exports}),59276:(function(x){function Y(s,J){if(s==null)return{};var j={},T=Object.keys(s),v,t;for(t=0;t<T.length;t++)v=T[t],!(J.indexOf(v)>=0)&&(j[v]=s[v]);return j}x.exports=Y,x.exports.__esModule=!0,x.exports.default=x.exports}),84169:(function(x){function Y(s,J){return J||(J=s.slice(0)),Object.freeze(Object.defineProperties(s,{raw:{value:Object.freeze(J)}}))}x.exports=Y,x.exports.__esModule=!0,x.exports.default=x.exports}),45647:(function(x,Y,s){var J=s(41162),j=s(33032),T=s(34911),v=s(30922);function t(D){return J(D)||j(D)||T(D)||v()}x.exports=t,x.exports.__esModule=!0,x.exports.default=x.exports})}]);
