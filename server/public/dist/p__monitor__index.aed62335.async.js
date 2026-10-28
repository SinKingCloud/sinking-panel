(self.webpackChunk=self.webpackChunk||[]).push([[335],{38221:(function(se,le,t){"use strict";t.r(le),t.d(le,{default:function(){return Je}});var re=t(55420),U=t.n(re),y=t(1974),g=t.n(y),p=t(98312),S=t.n(p),X=t(11604),d=t.n(X),n=t(96540),he=t(59522),Q=t(40372),Z=t(79273),qe=t(65552),_e=t(15487),en=t(74353),oe=t.n(en),q=t(62885),je=t(72523),ue=t(2904),nn=t(66585),Ae=t(45478),Fe=t(49216),tn=t(84169),j=t.n(tn),we=t(12185),De,Ee,Te,Oe,rn=(0,we.rU)(function(m){var r=m.css,a=m.token;return{panel:r(De||(De=j()([`
        container-type: inline-size;
        container-name: monitor-chart;
        min-width: 0;
        height: 100%;
        box-sizing: border-box;
        padding: 16px 18px 12px;
        overflow: hidden;
        border: 1px solid `,`;
        border-radius: `,`px;
        background: `,`;
        font-variant-numeric: tabular-nums;

        @media (max-width: 575px) {
            padding: 14px 12px 10px;
        }
    `])),a.colorBorderSecondary,a.borderRadiusLG,a.colorBgContainer),heading:r(Ee||(Ee=j()([`
        display: flex;
        align-items: center;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: 4px 12px;
        min-height: 22px;

        h3 {
            margin: 0;
            color: `,`;
            font-size: 14px;
            font-weight: 600;
            line-height: 22px;
        }

        .sample-time {
            color: `,`;
            font-size: 11px;
            line-height: 18px;
            white-space: nowrap;
        }

        .sample-time time { margin-inline-start: 3px; }
    `])),a.colorTextHeading,a.colorTextTertiary),summary:r(Te||(Te=j()([`
        display: flex;
        align-items: center;
        gap: 12px;
        min-height: 58px;
        padding-top: 8px;

        .summary-metrics {
            min-width: 0;
            display: grid;
            flex: 1;
            gap: 12px;
        }

        .summary-metric {
            min-width: 0;
            display: flex;
            align-items: baseline;
            flex-wrap: wrap;
            gap: 2px 9px;
        }

        .metric-label {
            min-width: 0;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            color: `,`;
            font-size: 11px;
            line-height: 18px;
            white-space: nowrap;
        }

        .metric-label i {
            width: 6px;
            height: 6px;
            flex: none;
            border-radius: 50%;
            background: var(--series-color);
        }

        .metric-value {
            display: inline-flex;
            align-items: baseline;
            gap: 3px;
            min-width: 0;
            white-space: nowrap;
        }

        .metric-value strong {
            color: `,`;
            font-size: 21px;
            font-weight: 500;
            letter-spacing: -.3px;
            line-height: 28px;
        }

        .metric-value small {
            color: `,`;
            font-size: 11px;
            line-height: 18px;
        }

        .summary-detail {
            min-width: 0;
            color: `,`;
            font-size: 11px;
            line-height: 18px;
            text-align: right;
        }

        @container monitor-chart (max-width: 440px) {
            .summary-metric { flex-direction: column; align-items: flex-start; gap: 0; }
            .metric-value strong { font-size: 20px; }
        }
    `])),a.colorTextSecondary,a.colorTextHeading,a.colorTextSecondary,a.colorTextSecondary),plot:r(Oe||(Oe=j()([`
        min-width: 0;
        margin-top: 2px;

        .chart-unit {
            height: 17px;
            color: `,`;
            font-size: 10px;
            line-height: 17px;
            text-align: right;
        }

        .chart-empty { display: flex; align-items: center; justify-content: center; }
        .ant-empty-description { color: `,`; font-size: 12px; }
    `])),a.colorTextTertiary,a.colorTextTertiary)}}),e=t(74848),Ce=["B/s","KB/s","MB/s","GB/s","TB/s"],Me=function(r){return r.toLocaleString("zh-CN",{maximumFractionDigits:2})},an=function(r,a){if(typeof r!="number"||!Number.isFinite(r))return{value:"--",suffix:""};if(a==="bytes"){var _=(0,ue.X)(r).split(" "),T=d()(_,2),s=T[0],b=T[1];return{value:s,suffix:"".concat(b,"/s")}}return{value:r.toFixed(a==="percent"?1:2),suffix:a==="percent"?"%":""}},Be=n.memo(function(m){var r,a,_=m.title,T=m.subtitle,s=m.history,b=m.series,M=m.unit,ce=m.loading,fe=rn(),x=fe.styles,W=he.A.useToken(),w=W.token,D=Q.Ay.useBreakpoint(),u=D.sm===!1?190:210,h=(0,q.DP)(),G=!!(!(h==null||(r=h.isDarkMode)===null||r===void 0)&&r.call(h)||!(h==null||(a=h.isDarkTheme)===null||a===void 0)&&a.call(h)),K=(0,n.useRef)(null),pe=(0,n.useRef)(null),k=s==null?void 0:s.points[s.points.length-1];(0,n.useEffect)(function(){if(!(!K.current||typeof ResizeObserver=="undefined")){var O=Math.round(K.current.getBoundingClientRect().width),R=0,Y=new ResizeObserver(function(B){var de=d()(B,1),ie=de[0],N=Math.round(ie.contentRect.width);N<=0||N===O||(O=N,window.clearTimeout(R),R=window.setTimeout(function(){var E,l;return(E=pe.current)===null||E===void 0||(E=E.chart)===null||E===void 0||(l=E.forceFit)===null||l===void 0?void 0:l.call(E)},80))});return Y.observe(K.current),function(){Y.disconnect(),window.clearTimeout(R)}}},[]);var V=(0,n.useMemo)(function(){var O=(s==null?void 0:s.points)||[],R=[],Y=0;O.forEach(function(l){b.forEach(function(o){var P=l[o.key],ee=typeof P=="number"&&Number.isFinite(P);ee&&(Y=Math.max(Y,P)),R.push({time:l.time,value:ee?P:null,series:o.label,samples:l.samples})})});var B=0;if(M==="bytes")for(;B<Ce.length-1&&Y>=Math.pow(1024,B+1);)B+=1;var de=M==="percent"?"%":M==="load"?"\u8D1F\u8F7D":Ce[B],ie=function(o){return o===null||!Number.isFinite(o)?"\u65E0\u6570\u636E":M==="bytes"?"".concat((0,ue.X)(o),"/s"):"".concat(Me(o)).concat(M==="percent"?"%":"")},N=function(o){var P=oe()(o instanceof Date?o:Number(o));return P.format(((s==null?void 0:s.end)||0)-((s==null?void 0:s.start)||0)>864e5?"MM-DD HH:mm":"HH:mm")},E={data:R,xField:"time",yField:"value",colorField:"series",seriesField:"series",shape:"smooth",connectNulls:!1,paddingTop:6,insetBottom:3,scale:{x:{type:"time",domain:[s==null?void 0:s.start,s==null?void 0:s.end],nice:!1},y:{domainMin:0,domainMax:M==="percent"?100:Y>0?void 0:1,nice:M!=="percent"},color:{domain:b.map(function(l){return l.label}),range:b.map(function(l){return l.color})}},axis:{x:{title:!1,line:!1,tick:!1,tickCount:4,labelFill:w.colorTextSecondary,labelFontSize:11,labelOpacity:1,labelAutoHide:!0,labelAutoRotate:!1,labelFormatter:N,grid:!1},y:{title:!1,line:!1,tick:!1,tickCount:4,labelFill:w.colorTextSecondary,labelFontSize:11,labelOpacity:1,labelFormatter:function(o){return Me(Number(o)/(M==="bytes"?Math.pow(1024,B):1))},grid:!0,gridLineDash:[3,4],gridLineWidth:1,gridStroke:w.colorBorderSecondary,gridStrokeOpacity:.8}},legend:!1,tooltip:{title:function(o){return"".concat(oe()(o==null?void 0:o.time).format("YYYY-MM-DD HH:mm")).concat((o==null?void 0:o.samples)===0?"\uFF08\u65E0\u91C7\u6837\uFF09":"")},items:[function(l){var o;return{name:l==null?void 0:l.series,color:(o=b.find(function(P){return P.label===(l==null?void 0:l.series)}))===null||o===void 0?void 0:o.color,value:ie(l==null?void 0:l.value)}}]},interaction:{tooltip:{shared:!0,series:!0,crosshairs:!0}},point:O.length<=2?{size:3,shape:"point"}:!1,area:{data:R.filter(function(l){var o;return l.series===((o=b[0])===null||o===void 0?void 0:o.label)}),shape:"smooth",connectNulls:!1,style:{fillOpacity:G?.1:.055,lineWidth:0},tooltip:!1,zIndex:-1},style:{lineWidth:1.8,lineCap:"round",lineJoin:"round"},theme:{type:G?"classicDark":"classic"},autoFit:!0,animate:!1};return{config:E,displayUnit:de}},[s,b,M,G,w.colorTextSecondary,w.colorBorderSecondary]),ae=(0,n.useMemo)(function(){return(0,e.jsx)(nn.A,g()(g()({ref:pe},V.config),{},{height:u,containerStyle:{height:u,width:"100%"}}))},[V.config,u]);return(0,e.jsxs)("section",{ref:K,className:x.panel,"aria-label":_,children:[(0,e.jsxs)("div",{className:x.heading,children:[(0,e.jsx)(q.hE,{size:"small",children:(0,e.jsx)("h3",{children:_})}),(0,e.jsx)("span",{className:"sample-time",title:k?"".concat(k.samples>0?"\u6240\u9009\u8303\u56F4\u672B\u4E2A\u65F6\u6BB5":"\u5F53\u524D\u65F6\u6BB5\u65E0\u91C7\u6837","\uFF1A").concat(oe()(k.time).format("YYYY-MM-DD HH:mm")):void 0,children:k?(0,e.jsxs)(e.Fragment,{children:[k.samples>0?"\u7EDF\u8BA1\u65F6\u6BB5":"\u65E0\u91C7\u6837"," ",(0,e.jsx)("time",{dateTime:new Date(k.time).toISOString(),children:oe()(k.time).format("MM-DD HH:mm")})]}):"\u6682\u65E0\u91C7\u6837"})]}),(0,e.jsxs)("div",{className:x.summary,children:[(0,e.jsx)("div",{className:"summary-metrics",style:{gridTemplateColumns:"repeat(".concat(b.length,", minmax(0, 1fr))")},children:b.map(function(O){var R=an(k==null?void 0:k[O.key],M);return(0,e.jsxs)("div",{className:"summary-metric",style:{"--series-color":O.color},children:[(0,e.jsxs)("span",{className:"metric-label",children:[(0,e.jsx)("i",{}),b.length===1?"\u4F7F\u7528\u7387":O.label]}),(0,e.jsxs)("span",{className:"metric-value",children:[(0,e.jsx)("strong",{children:R.value}),R.suffix&&(0,e.jsx)("small",{children:R.suffix})]})]},O.key)})}),T&&(0,e.jsx)("span",{className:"summary-detail",children:T})]}),(0,e.jsxs)("div",{className:x.plot,children:[(0,e.jsx)("div",{className:"chart-unit",children:V.displayUnit}),(0,e.jsx)(Ae.A,{spinning:ce,children:s!=null&&s.points.length?ae:(0,e.jsx)("div",{className:"chart-empty",style:{height:u},children:(0,e.jsx)(Fe.A,{image:Fe.A.PRESENTED_IMAGE_SIMPLE,description:ce?"\u6B63\u5728\u52A0\u8F7D\u76D1\u63A7\u6570\u636E":"\u6240\u9009\u65F6\u95F4\u6682\u65E0\u76D1\u63A7\u6570\u636E"})})})]})]})});Be.displayName="MonitorChart";var sn=Be,ye=t(20873),ln=t(50400),on=t(79508),un=t(48212),Ne,Pe,ze,Le,Ie,Ye,cn=(0,we.rU)(function(m){var r=m.css,a=m.token;return{form:r(Ne||(Ne=j()([`
        padding-top: 4px;

        .ant-form-item { margin-bottom: 20px; }
        .ant-form-item-extra { margin-top: 6px; color: `,`; font-size: 12px; }
    `])),a.colorTextTertiary),enabledRow:r(Pe||(Pe=j()([`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 20px;
        margin-bottom: 22px;

        .setting-copy { min-width: 0; }
        label { color: `,"; font-size: ",`px; line-height: 22px; }
        p { margin: 5px 0 0; color: `,`; font-size: 12px; line-height: 20px; }
        .ant-switch { flex: none; }
    `])),a.colorText,a.fontSize,a.colorTextTertiary),state:r(ze||(ze=j()([`
        min-height: 190px;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 12px;
        color: `,`;
        font-size: 12px;
        line-height: 20px;

        .state-icon { color: `,`; font-size: 22px; }
        .state-error { max-width: 100%; text-align: center; overflow-wrap: anywhere; }
    `])),a.colorTextSecondary,a.colorError),storage:r(Le||(Le=j()([`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        padding-top: 2px;
        color: `,`;
        font-size: 12px;
        line-height: 20px;
        font-variant-numeric: tabular-nums;
    `])),a.colorTextSecondary),warning:r(Ie||(Ie=j()([`
        margin-top: 12px;
        color: `,`;
        font-size: 12px;
        line-height: 20px;
        overflow-wrap: anywhere;
    `])),a.colorWarningText),error:r(Ye||(Ye=j()([`
        margin-top: 12px;
        color: `,`;
        font-size: 12px;
        line-height: 20px;
        overflow-wrap: anywhere;
    `])),a.colorError)}}),$e=(0,n.forwardRef)(function(m,r){var a=m.onLoaded,_=m.onSuccess,T=cn(),s=T.styles,b=Z.A.useApp(),M=b.message,ce=ye.A.useForm(),fe=d()(ce,1),x=fe[0],W=(0,n.useRef)(null),w=(0,n.useRef)(!1),D=(0,n.useRef)(!1),u=(0,n.useRef)(0),h=(0,n.useRef)(!1),G=(0,n.useRef)(!1),K=(0,n.useRef)({onLoaded:a,onSuccess:_});K.current={onLoaded:a,onSuccess:_};var pe=(0,n.useState)(),k=d()(pe,2),V=k[0],ae=k[1],O=(0,n.useState)(!1),R=d()(O,2),Y=R[0],B=R[1],de=(0,n.useState)(!1),ie=d()(de,2),N=ie[0],E=ie[1],l=(0,n.useState)(""),o=d()(l,2),P=o[0],ee=o[1],C=(0,n.useState)(""),ge=d()(C,2),Se=ge[0],me=ge[1];(0,n.useEffect)(function(){return w.current=!0,function(){w.current=!1,D.current=!1,u.current+=1}},[]);var ve=(0,n.useCallback)(S()(U()().mark(function $(){var A,te,z,v,J;return U()().wrap(function(c){for(;;)switch(c.prev=c.next){case 0:if(!(!D.current||h.current)){c.next=2;break}return c.abrupt("return");case 2:return A=++u.current,G.current=!1,B(!0),ee(""),me(""),ae(void 0),x.resetFields(),c.prev=9,c.next=12,(0,je.S3)({body:{action:"settings"}});case 12:if(v=c.sent,!(!w.current||!D.current||A!==u.current)){c.next=15;break}return c.abrupt("return");case 15:if(!((v==null?void 0:v.code)!==200||!v.data)){c.next=17;break}throw new Error((v==null?void 0:v.message)||"\u8BFB\u53D6\u76D1\u63A7\u8BBE\u7F6E\u5931\u8D25");case 17:J=v.data,ae(J),x.setFieldsValue({enabled:J.enabled,retention_days:J.retention_days}),G.current=!0,(te=(z=K.current).onLoaded)===null||te===void 0||te.call(z,J),c.next=27;break;case 24:c.prev=24,c.t0=c.catch(9),w.current&&D.current&&A===u.current&&ee(c.t0 instanceof Error?c.t0.message:"\u8BFB\u53D6\u76D1\u63A7\u8BBE\u7F6E\u5931\u8D25");case 27:return c.prev=27,w.current&&D.current&&A===u.current&&B(!1),c.finish(27);case 30:case"end":return c.stop()}},$,null,[[9,24,27,30]])})),[x]),ne=(0,n.useCallback)(function(){var $;h.current||(D.current=!1,G.current=!1,u.current+=1,($=W.current)===null||$===void 0||$.hide(),x.resetFields(),ae(void 0),B(!1),ee(""),me(""))},[x]);(0,n.useImperativeHandle)(r,function(){return{open:function(){var A;!w.current||h.current||(D.current=!0,(A=W.current)===null||A===void 0||A.show(),ve())}}},[ve]);var ke=(function(){var $=S()(U()().mark(function A(te){var z,v,J,L,c;return U()().wrap(function(f){for(;;)switch(f.prev=f.next){case 0:if(!(h.current||!G.current||!D.current)){f.next=2;break}return f.abrupt("return");case 2:return z=u.current,h.current=!0,E(!0),me(""),f.prev=6,f.next=9,(0,je.S3)({body:{action:"set",enabled:te.enabled,retention_days:te.retention_days}});case 9:if(L=f.sent,!(!w.current||!D.current||z!==u.current)){f.next=12;break}return f.abrupt("return");case 12:if(!((L==null?void 0:L.code)!==200||!L.data)){f.next=14;break}throw new Error((L==null?void 0:L.message)||"\u4FDD\u5B58\u76D1\u63A7\u8BBE\u7F6E\u5931\u8D25");case 14:c=L.data,h.current=!1,E(!1),ne(),M.success(L.message||"\u76D1\u63A7\u8BBE\u7F6E\u5DF2\u4FDD\u5B58"),(v=(J=K.current).onSuccess)===null||v===void 0||v.call(J,c),f.next=25;break;case 22:f.prev=22,f.t0=f.catch(6),w.current&&D.current&&z===u.current&&me(f.t0 instanceof Error?f.t0.message:"\u4FDD\u5B58\u76D1\u63A7\u8BBE\u7F6E\u5931\u8D25");case 25:return f.prev=25,w.current&&D.current&&z===u.current&&(h.current=!1,E(!1)),f.finish(25);case 28:case"end":return f.stop()}},A,null,[[6,22,25,28]])}));return function(te){return $.apply(this,arguments)}})();return(0,e.jsx)(q.Ht,{ref:W,title:(0,e.jsx)(q.hE,{children:"\u76D1\u63A7\u8BBE\u7F6E"}),width:440,onOk:function(){G.current&&!h.current&&x.submit()},onCancel:ne,modalProps:{forceRender:!0,confirmLoading:N,okText:"\u4FDD\u5B58",cancelText:"\u53D6\u6D88",okButtonProps:{disabled:Y||!V||!!P},cancelButtonProps:{disabled:N},closable:!N,keyboard:!N,mask:{closable:!N}},children:(0,e.jsx)(ye.A,{form:x,layout:"vertical",variant:"filled",className:s.form,disabled:Y||N,onFinish:ke,requiredMark:!1,children:Y?(0,e.jsxs)("div",{className:s.state,role:"status",children:[(0,e.jsx)(Ae.A,{}),(0,e.jsx)("span",{children:"\u6B63\u5728\u8BFB\u53D6\u76D1\u63A7\u8BBE\u7F6E"})]}):P?(0,e.jsxs)("div",{className:s.state,children:[(0,e.jsx)(q.In,{type:"ExclamationCircleOutlined",className:"state-icon"}),(0,e.jsx)("span",{className:"state-error",role:"alert",children:P}),(0,e.jsx)(ln.A,{size:"small",icon:(0,e.jsx)(q.In,{type:"ReloadOutlined"}),onClick:function(){ve()},children:"\u91CD\u65B0\u52A0\u8F7D"})]}):V?(0,e.jsxs)(e.Fragment,{children:[(0,e.jsxs)("div",{className:s.enabledRow,children:[(0,e.jsxs)("div",{className:"setting-copy",children:[(0,e.jsx)("label",{htmlFor:"monitor-settings-enabled",children:"\u76D1\u63A7\u91C7\u96C6"}),(0,e.jsx)("p",{children:"\u5173\u95ED\u540E\u505C\u6B62\u91C7\u96C6\uFF0C\u5DF2\u4FDD\u5B58\u7684\u5386\u53F2\u4ECD\u53EF\u67E5\u770B\u3002"})]}),(0,e.jsx)(ye.A.Item,{name:"enabled",valuePropName:"checked",noStyle:!0,rules:[{required:!0,type:"boolean",message:"\u8BF7\u9009\u62E9\u76D1\u63A7\u91C7\u96C6\u72B6\u6001"}],children:(0,e.jsx)(on.A,{id:"monitor-settings-enabled","aria-label":"\u5F00\u542F\u76D1\u63A7"})})]}),(0,e.jsx)(ye.A.Item,{name:"retention_days",label:"\u5386\u53F2\u4FDD\u5B58\u5929\u6570",extra:"\u4FDD\u7559 1\u2013365 \u5929\uFF0C\u5230\u671F\u7684\u5386\u53F2\u6570\u636E\u81EA\u52A8\u6E05\u7406\u3002",rules:[{required:!0,message:"\u8BF7\u8F93\u5165\u5386\u53F2\u4FDD\u5B58\u5929\u6570"},{type:"integer",min:1,max:365,message:"\u8BF7\u8F93\u5165 1\u2013365 \u4E4B\u95F4\u7684\u6574\u6570"}],children:(0,e.jsx)(un.A,{min:1,max:365,precision:0,suffix:"\u5929","aria-label":"\u5386\u53F2\u4FDD\u5B58\u5929\u6570",style:{width:"100%"}})}),(0,e.jsxs)("div",{className:s.storage,children:[(0,e.jsx)("span",{children:"\u5386\u53F2\u6570\u636E\u5360\u7528"}),(0,e.jsx)("span",{children:(0,ue.X)(V.storage_bytes)})]}),V.last_error&&(0,e.jsx)("div",{className:s.warning,role:"status",children:V.last_error}),Se&&(0,e.jsx)("div",{className:s.error,role:"alert",children:Se})]}):null})})});$e.displayName="MonitorSettingsModal";var dn=$e,He,Ue,Xe,We,Ge,Ke,Ve,mn=(0,we.rU)(function(m){var r=m.css,a=m.token;return{page:r(He||(He=j()([`
        font-variant-numeric: tabular-nums;

        @container data-table-workspace (max-width: 560px) {
            .ui-table-toolbar.with-content { flex-wrap: wrap; gap: 8px; }
            .ui-table-toolbar .command-content { flex-basis: 112px; }
            .ui-table-toolbar.with-content .command-actions { margin-inline-start: auto; }
        }
    `]))),timeFilter:r(Ue||(Ue=j()([`
        position: relative;
        display: inline-flex;
        align-items: center;
        max-width: 100%;

        .ant-segmented { font-size: 12px; }
        .ant-segmented-item-label { padding-inline: 12px; }

        @container data-table-workspace (max-width: 380px) {
            .ant-segmented-item-label { padding-inline: 7px; }
        }
    `]))),mobile:r(Xe||(Xe=j()([`
        .ui-table-toolbar .command-actions-scroll { position: relative; }
        .ui-table-toolbar .command-actions-scroll > .ant-btn:last-child { margin-inline-start: auto; }
        .ui-table-toolbar .command-actions-scroll > .ant-picker {
            inset-inline-start: 0;
            inset-inline-end: auto;
            width: 112px;
        }
    `]))),dateAnchor:r(We||(We=j()([`
        && {
            position: absolute;
            inset: 0 0 0 auto;
            width: 72px;
            min-width: 0;
            height: 100%;
            padding: 0;
            overflow: hidden;
            border: 0;
            opacity: 0;
            pointer-events: none;
        }
    `]))),datePopup:r(Ge||(Ge=j()([`
        && .ant-picker-range-arrow {
            left: clamp(`,"px, calc(var(--arrow-x, 50%) - ","px), calc(100% - ",`px)) !important;
            right: auto;
            padding-inline: 0;
            transition: none;
            &::before { inset-inline-start: 0; }
        }

        && .ant-picker-panels > :last-child:not(:first-child) { display: none; }
        && .ant-picker-panels > :first-child .ant-picker-header-next-btn,
        && .ant-picker-panels > :first-child .ant-picker-header-super-next-btn { visibility: visible !important; }
    `])),a.borderRadiusLG,a.sizePopupArrow/2,a.borderRadiusLG+a.sizePopupArrow),problem:r(Ke||(Ke=j()([`
        margin-inline-start: 5px;
        color: `,`;
    `])),a.colorWarningText),charts:r(Ve||(Ve=j()([`
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 16px;
        min-width: 0;

        @container data-table-workspace (max-width: 860px) {
            grid-template-columns: minmax(0, 1fr);
            gap: 12px;
        }
    `])))}}),fn=[{value:"1",label:"1 \u5C0F\u65F6"},{value:"6",label:"6 \u5C0F\u65F6"},{value:"24",label:"24 \u5C0F\u65F6"},{value:"168",label:"7 \u5929"}],Je=(function(){var m=mn(),r=m.styles,a=m.cx,_=he.A.useToken(),T=_.token,s=Q.Ay.useBreakpoint(),b=s.md===!1,M=Z.A.useApp(),ce=M.message,fe=(0,n.useRef)(null),x=(0,n.useRef)(!1),W=(0,n.useRef)(0),w=(0,n.useState)(),D=d()(w,2),u=D[0],h=D[1],G=(0,n.useState)(""),K=d()(G,2),pe=K[0],k=K[1],V=(0,n.useState)(!0),ae=d()(V,2),O=ae[0],R=ae[1],Y=(0,n.useState)(!0),B=d()(Y,2),de=B[0],ie=B[1],N=(0,n.useState)("1"),E=d()(N,2),l=E[0],o=E[1],P=(0,n.useState)(null),ee=d()(P,2),C=ee[0],ge=ee[1],Se=(0,n.useState)(!1),me=d()(Se,2),ve=me[0],ne=me[1],ke=(0,n.useState)(),$=d()(ke,2),A=$[0],te=$[1],z=(0,n.useCallback)(S()(U()().mark(function H(){var i,be,xe,I;return U()().wrap(function(F){for(;;)switch(F.prev=F.next){case 0:return i=++W.current,be=C?Math.min(C[1].endOf("day").valueOf(),Date.now()):Date.now(),xe=C?C[0].startOf("day").valueOf():be-Number(l)*36e5,R(!0),k(""),F.prev=5,F.next=8,(0,je.S3)({body:{action:"history",start:xe,end:be}});case 8:if(I=F.sent,!(!x.current||i!==W.current)){F.next=11;break}return F.abrupt("return");case 11:if(!((I==null?void 0:I.code)!==200||!I.data)){F.next=13;break}throw new Error((I==null?void 0:I.message)||"\u8BFB\u53D6\u76D1\u63A7\u5386\u53F2\u5931\u8D25");case 13:te(g()(g()({},I.data),{},{points:Array.isArray(I.data.points)?I.data.points:[]})),k(""),F.next=20;break;case 17:F.prev=17,F.t0=F.catch(5),x.current&&i===W.current&&k(F.t0 instanceof Error?F.t0.message:"\u8BFB\u53D6\u76D1\u63A7\u5386\u53F2\u5931\u8D25");case 20:return F.prev=20,x.current&&i===W.current&&(R(!1),ie(!1)),F.finish(20);case 23:case"end":return F.stop()}},H,null,[[5,17,20,23]])})),[l,C]);(0,n.useEffect)(function(){return x.current=!0,function(){x.current=!1,W.current+=1}},[]),(0,n.useEffect)(function(){return z(),function(){W.current+=1}},[z]),(0,n.useEffect)(function(){if(!(C&&!C[1].isSame(oe()(),"day")||(u==null?void 0:u.enabled)===!1)){var H=window.setInterval(function(){document.visibilityState==="visible"&&z()},6e4);return function(){return window.clearInterval(H)}}},[z,C,u==null?void 0:u.enabled]);var v=A==null?void 0:A.points[A.points.length-1],J=(0,n.useMemo)(function(){return[{title:"\u7CFB\u7EDF\u8D1F\u8F7D",subtitle:"",unit:"load",series:[{key:"load1",label:"1 \u5206\u949F",color:T.colorPrimary},{key:"load5",label:"5 \u5206\u949F",color:"#50b6b2"},{key:"load15",label:"15 \u5206\u949F",color:"#d8a55b"}]},{title:"CPU \u4F7F\u7528\u7387",subtitle:"",unit:"percent",series:[{key:"cpu",label:"\u5E73\u5747\u4F7F\u7528\u7387",color:T.colorPrimary},{key:"cpu_max",label:"\u5CF0\u503C\u4F7F\u7528\u7387",color:"#50b6b2"}]},{title:"\u5185\u5B58\u4F7F\u7528\u7387",subtitle:v?"".concat((0,ue.X)(v.memory_used)," / ").concat((0,ue.X)(v.memory_total)):"-- / --",unit:"percent",series:[{key:"memory",label:"\u5185\u5B58\u4F7F\u7528\u7387",color:T.colorPrimary}]},{title:"\u78C1\u76D8\u4F7F\u7528\u7387",subtitle:v?"".concat((0,ue.X)(v.disk_used)," / ").concat((0,ue.X)(v.disk_total)):"-- / --",unit:"percent",series:[{key:"disk_usage",label:"\u78C1\u76D8\u4F7F\u7528\u7387",color:T.colorPrimary}]},{title:"\u78C1\u76D8 I/O",subtitle:"",unit:"bytes",series:[{key:"disk_read",label:"\u8BFB\u53D6",color:T.colorPrimary},{key:"disk_write",label:"\u5199\u5165",color:"#50b6b2"}]},{title:"\u7F51\u7EDC\u6D41\u91CF",subtitle:"",unit:"bytes",series:[{key:"net_recv",label:"\u63A5\u6536",color:T.colorPrimary},{key:"net_sent",label:"\u53D1\u9001",color:"#50b6b2"}]}]},[v,T.colorPrimary]),L=[pe,u==null?void 0:u.last_error].filter(Boolean).join("\uFF1B"),c=ve||C?"custom":l,Re=[].concat(fn,[{value:"custom",label:"\u81EA\u5B9A\u4E49",className:"monitor-custom-period",title:C?"".concat(C[0].format("YYYY-MM-DD")," \u81F3 ").concat(C[1].format("YYYY-MM-DD")):"\u81EA\u5B9A\u4E49\u65E5\u671F"}]),f=function(i){if(i==="custom"){ne(!0);return}ne(!1),o(i),ge(null)},pn=function(i){if(i){var be=oe()();if(i.some(function(Ze){return Ze.isAfter(be,"day")})){ce.warning("\u76D1\u63A7\u65E5\u671F\u4E0D\u80FD\u665A\u4E8E\u4ECA\u5929");return}var xe=i[0].startOf("day"),I=i[1].endOf("day");if(I.isBefore(xe)||I.valueOf()-xe.valueOf()>365*864e5){ce.warning("\u8BF7\u9009\u62E9\u8D77\u59CB\u65E5\u671F\u4E0D\u665A\u4E8E\u7ED3\u675F\u65E5\u671F\u3001\u8DE8\u5EA6\u4E0D\u8D85\u8FC7 365 \u5929\u7684\u8303\u56F4");return}}ge(i),ne(!1)},Qe=(0,e.jsx)(qe.A.RangePicker,{className:r.dateAnchor,classNames:{popup:{root:r.datePopup}},value:C,open:ve,onOpenChange:ne,tabIndex:-1,allowClear:!1,inputReadOnly:!0,placement:"bottomRight",popupAlign:{overflow:{adjustX:!0,adjustY:!0,shiftX:!0}},format:"YYYY-MM-DD","aria-label":"\u81EA\u5B9A\u4E49\u76D1\u63A7\u65E5\u671F",placeholder:["\u5F00\u59CB\u65E5\u671F","\u7ED3\u675F\u65E5\u671F"],disabledDate:function(i){return i.isAfter(oe()(),"day")},onChange:function(i){i!=null&&i[0]&&i[1]&&pn([i[0],i[1]])}});return(0,e.jsxs)(q.nB,{loading:de,children:[(0,e.jsx)(q.II,{rootClassName:a(r.page,b&&r.mobile),ariaLabel:"\u76D1\u63A7\u7BA1\u7406",search:!1,pagination:!1,hero:{eyebrow:"SYSTEM MONITOR",title:"\u76D1\u63A7\u7BA1\u7406"},filters:b?[{type:"select",name:"period",label:"\u76D1\u63A7\u65F6\u95F4\u8303\u56F4",icon:"ClockCircleOutlined",allLabel:!1,value:c,options:Re,onChange:f}]:void 0,toolbar:{left:b?void 0:(0,e.jsxs)("div",{className:r.timeFilter,children:[(0,e.jsx)(_e.A,{"aria-label":"\u76D1\u63A7\u65F6\u95F4\u8303\u56F4",value:c,options:Re,onChange:f,onClick:function(i){i.target.closest(".monitor-custom-period")&&ne(!0)},onKeyDownCapture:function(i){(i.key==="Enter"||i.key===" ")&&i.target.closest(".monitor-custom-period")&&(i.preventDefault(),ne(!0))}}),Qe]}),actions:[{key:"custom-range",type:"custom",hidden:!b,render:Qe},{key:"settings",type:"button",label:(0,e.jsxs)(e.Fragment,{children:["\u76D1\u63A7\u8BBE\u7F6E",L&&(0,e.jsx)(q.In,{type:"ExclamationCircleOutlined",className:r.problem})]}),icon:"SettingOutlined",tooltip:L||void 0,onClick:function(){var i;return(i=fe.current)===null||i===void 0?void 0:i.open()}}],refresh:{ariaLabel:"\u5237\u65B0\u76D1\u63A7\u6570\u636E",loading:O,onClick:function(){z()}}},tableRender:function(){return(0,e.jsx)("div",{className:r.charts,children:J.map(function(i){return(0,e.jsx)(sn,g()(g()({},i),{},{history:A,loading:O}),i.title)})})}}),(0,e.jsx)(dn,{ref:fe,onLoaded:h,onSuccess:function(i){h(i),z()}})]})});t.dn(Je)}),2904:(function(se,le,t){"use strict";t.d(le,{P:function(){return re},X:function(){return U}});function re(){var y=arguments.length>0&&arguments[0]!==void 0?arguments[0]:16,g="abcdefghijklmnopqrstuvwxyz";g+=g.toUpperCase(),g+="0123456789";for(var p="",S=0;S<y;S++){var X=Math.floor(Math.random()*g.length);p+=g[X]}return p}var U=function(g){var p=arguments.length>1&&arguments[1]!==void 0?arguments[1]:1024,S=Number(g||0);if(S<=0)return"0 B";for(var X=["B","KB","MB","GB","TB"],d=S,n=0;d>=p&&n<X.length-1;)d=d/p,n+=1;return"".concat(d.toFixed(n===0?0:2)," ").concat(X[n])}}),59224:(function(se,le,t){var re=t(34911);function U(y,g){var p=typeof Symbol!="undefined"&&y[Symbol.iterator]||y["@@iterator"];if(!p){if(Array.isArray(y)||(p=re(y))||g&&y&&typeof y.length=="number"){p&&(y=p);var S=0,X=function(){};return{s:X,n:function(){return S>=y.length?{done:!0}:{done:!1,value:y[S++]}},e:function(Z){throw Z},f:X}}throw new TypeError(`Invalid attempt to iterate non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}var d=!0,n=!1,he;return{s:function(){p=p.call(y)},n:function(){var Z=p.next();return d=Z.done,Z},e:function(Z){n=!0,he=Z},f:function(){try{!d&&p.return!=null&&p.return()}finally{if(n)throw he}}}}se.exports=U,se.exports.__esModule=!0,se.exports.default=se.exports}),39874:(function(se,le,t){"use strict";t.d(le,{A:function(){return g}});var re=t(53954),U=t(52176),y=t(56822);function g(p,S,X){return S=(0,re.A)(S),(0,y.A)(p,(0,U.A)()?Reflect.construct(S,X||[],(0,re.A)(p).constructor):S.apply(p,X))}})}]);
