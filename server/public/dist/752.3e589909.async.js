"use strict";(self.webpackChunk=self.webpackChunk||[]).push([[752],{9752:(function(In,De,_){var je=_(1974),se=_.n(je),Ne=_(55420),B=_.n(Ne),J=_(98312),Q=_.n(J),oe=_(84169),R=_.n(oe),Ve=_(11604),L=_.n(Ve),u=_(96540),_e=_(40961),re=_(79273),ue=_(50400),fe=_(49216),ae=_(20873),ve=_(22817),pe=_(58679),G=_(45478),t=_(12185),D=_(62885),j=_(57798),m=_(8490),y=_(9300),c=_(74848),U,te,A,N,de,O,W,Re,V,Ce,He,We,Qe,Me,ke,ge,be,wn=function(e,i,r){var l=e.slice(),o=l.splice(i,1),w=L()(o,1),s=w[0];return l.splice(r,0,s),l},Pn=180,Dn="cubic-bezier(.2, .8, .2, 1)",Ae=new Map,Ze=function(e){return Array.from(e.querySelectorAll("[data-type-id]"))},un=function(e,i){var r=new Map;return Ze(e).forEach(function(l){var o=Number(l.dataset.typeId);Number.isFinite(o)&&(!i||i.has(o))&&r.set(o,l.getBoundingClientRect())}),r},ln=function(e){Array.from(Ae.entries()).forEach(function(i){var r=L()(i,2),l=r[0],o=r[1];e.contains(l)&&(o.cancel(),Ae.delete(l))})},jn=function(e,i){var r=Ze(e).filter(function(l){return i.has(Number(l.dataset.typeId))});return r.forEach(function(l){return l.style.setProperty("transition","none")}),function(){return r.forEach(function(l){return l.style.removeProperty("transition")})}},Xn=function(e,i){var r,l;if(!((r=(l=window).matchMedia)!==null&&r!==void 0&&r.call(l,"(prefers-reduced-motion: reduce)").matches)){var o=new Map;Ze(e).forEach(function(s){var k=Number(s.dataset.typeId);i.has(k)&&o.set(k,s)});var w=[];i.forEach(function(s,k){var h=o.get(k);if(h){var M=h.getBoundingClientRect(),x=s.left-M.left,T=s.top-M.top;(Math.abs(x)>=.5||Math.abs(T)>=.5)&&w.push({deltaX:x,deltaY:T,element:h})}}),w.forEach(function(s){var k,h=s.deltaX,M=s.deltaY,x=s.element;(k=Ae.get(x))===null||k===void 0||k.cancel();var T=x.animate([{transform:"translate3d(".concat(h,"px, ").concat(M,"px, 0)")},{transform:"translate3d(0, 0, 0)"}],{duration:Pn,easing:Dn});Ae.set(x,T);var Z=function(){Ae.get(x)===T&&Ae.delete(x)};T.addEventListener("finish",Z,{once:!0}),T.addEventListener("cancel",Z,{once:!0})})}},Rn=function(e){var i=e.overIndex<0?e.activeIndex:e.overIndex,r=Math.min(e.activeIndex,i),l=Math.max(e.activeIndex,i);return new Set(e.rows.slice(r,l+1).map(function(o){return o.id}))},Cn=function(e,i,r){if(i<0||r<0)return new Set;var l=Math.min(i,r),o=Math.max(i,r);return new Set(e.slice(l,o+1).map(function(w){return w.id}))},$n=function(e,i){for(var r=new Set,l=Math.max(e.length,i.length),o=0;o<l;o+=1){var w,s,k=(w=e[o])===null||w===void 0?void 0:w.id,h=(s=i[o])===null||s===void 0?void 0:s.id;k!==h&&(k!==void 0&&r.add(k),h!==void 0&&r.add(h))}return r},Gn=function(e,i){for(var r=0,l=e.length-1;r<l;){var o=Math.floor((r+l)/2);e[o].center<i?r=o+1:l=o}if(r>0){var w=r-1;if(Math.abs(e[w].center-i)<=Math.abs(e[r].center-i))return w}return r},qe=function(e){for(var i=e.overIndex<0?e.activeIndex:e.overIndex,r=Math.min(e.activeIndex,i),l=Math.max(e.activeIndex,i),o=r;o<=l;o+=1){var w,s=(w=e.rows[o])===null||w===void 0?void 0:w.element;s&&(s.style.removeProperty("transform"),s.classList.remove(e.classes.dragging,e.classes.shifted,e.classes.dropBefore,e.classes.dropAfter))}},en=function(e){var i=e.handle,r=e.pointerId;if(!(typeof i.hasPointerCapture!="function"||typeof i.releasePointerCapture!="function"))try{i.hasPointerCapture(r)&&i.releasePointerCapture(r)}catch(l){}},zn=function(e,i){for(var r=e.rows[e.activeIndex],l=r.height+e.rowGap,o=e.overIndex,w=o<0?e.activeIndex:o,s=i<0?e.activeIndex:i,k=Math.min(w,s),h=Math.max(w,s),M=k;M<=h;M+=1)if(M!==e.activeIndex){var x=o>e.activeIndex&&M>e.activeIndex&&M<=o?-l:o>=0&&o<e.activeIndex&&M>=o&&M<e.activeIndex?l:0,T=i>e.activeIndex&&M>e.activeIndex&&M<=i?-l:i>=0&&i<e.activeIndex&&M>=i&&M<e.activeIndex?l:0;if(x!==T){var Z=e.rows[M].element;Z.style.transform=T===0?"":"translate3d(0, ".concat(T,"px, 0)"),Z.classList.toggle(e.classes.shifted,T!==0)}}o>=0&&o!==e.activeIndex&&e.rows[o].element.classList.remove(e.classes.dropBefore,e.classes.dropAfter),i>=0&&i!==e.activeIndex&&e.rows[i].element.classList.add(i<e.activeIndex?e.classes.dropBefore:e.classes.dropAfter),e.overIndex=i},Mn=function g(e,i){if(e.frameId=0,!(e.ended||!e.activated||!e.list||!e.listRect)){var r=e.lastFrameTime===0?16:Math.min(i-e.lastFrameTime,32);e.lastFrameTime=i;var l=40,o=520,w=e.list,s=e.listRect,k=e.clientX>=s.left-20&&e.clientX<=s.right+20&&e.clientY>=s.top-20&&e.clientY<=s.bottom+20,h=0;k&&e.clientY<s.top+l?h=-o*Math.min(1,(s.top+l-e.clientY)/l):k&&e.clientY>s.bottom-l&&(h=o*Math.min(1,(e.clientY-s.bottom+l)/l));var M=w.scrollTop;h!==0&&(w.scrollTop+=h*r/1e3);var x=w.scrollTop!==M,T=e.clientY-e.startY+w.scrollTop-e.startScrollTop;e.rows[e.activeIndex].element.style.transform="translate3d(0, ".concat(T,"px, 0)");var Z=e.clientX<s.left-20||e.clientX>s.right+20||e.clientY<s.top-20||e.clientY>s.bottom+20,he=Z?-1:Gn(e.rows,e.rows[e.activeIndex].center+T);he!==e.overIndex&&zn(e,he),x&&!e.ended&&(e.frameId=window.requestAnimationFrame(function(C){return g(e,C)}))}},Jn=function(e){e.frameId===0&&!e.ended&&(e.frameId=window.requestAnimationFrame(function(i){return Mn(e,i)}))},Vn=(0,t.rU)(function(g,e){var i=g.css,r=g.token;return{modal:i(U||(U=R()([`
        .ant-modal {
            max-width: calc(100vw - 24px);
            top: clamp(12px, 8vh, 72px);
            padding-bottom: 12px;
        }

        .ant-modal-body {
            max-height: calc(100dvh - 104px);
            overflow-x: hidden;
            overflow-y: auto;
        }
    `]))),content:i(te||(te=R()([`
        display: flex;
        min-width: 0;
        flex-direction: column;
        gap: `,`px;
    `])),e.compact?6:8),listShell:i(A||(A=R()([`
        display: flex;
        min-width: 0;
        min-height: 0;
        flex-direction: column;
        gap: `,`px;
        padding-top: `,`px;
    `])),e.compact?4:6,e.compact?4:6),list:i(N||(N=R()([`
        display: flex;
        min-width: 0;
        max-height: clamp(
                `,`px,
                calc(100dvh - `,`px),
                `,`px
        );
        flex-direction: column;
        gap: `,`px;
        overflow-x: hidden;
        overflow-y: auto;
        padding-block: 2px;
        scrollbar-width: thin;
    `])),e.compact?120:132,e.compact?176:192,e.compact?300:340,e.compact?3:4),listLocked:i(de||(de=R()([`
        [data-drag-handle],
        [data-row-action] {
            pointer-events: none;
            opacity: .45;
        }
    `]))),loading:i(O||(O=R()([`
        display: flex;
        min-height: `,`px;
        align-items: center;
        justify-content: center;
    `])),e.compact?156:176),empty:i(W||(W=R()([`
        display: flex;
        min-height: `,`px;
        align-items: center;
        justify-content: center;

        .ant-empty {
            margin-block: 0;
        }
    `])),e.compact?156:176),item:i(Re||(Re=R()([`
        position: relative;
        display: grid;
        min-width: 0;
        min-height: `,`px;
        grid-template-columns: `,`px minmax(0, 1fr) auto;
        align-items: center;
        padding: `,`;
        border-radius: `,`px;
        background: `,`;
        transition: transform `,"ms ",`,
        background-color 120ms ease,
        box-shadow 120ms ease;

        &:hover,
        &:focus-within {
            background: `,`;
        }

        @media (pointer: coarse) {
            min-height: 44px;
        }
    `])),e.compact?38:44,e.compact?32:36,e.compact?"3px 6px 3px 4px":"4px 8px 4px 5px",r.borderRadiusSM,r.colorFillQuaternary,Pn,Dn,r.colorFillTertiary),dragging:i(V||(V=R()([`
        z-index: 2;
        background: `,`;
        box-shadow: `,`;
        transition: background-color 120ms ease, box-shadow 120ms ease;
        will-change: transform;
    `])),r.colorPrimaryBg,r.boxShadowTertiary),shifted:i(Ce||(Ce=R()([`
        will-change: transform;
    `]))),dropBefore:i(He||(He=R()([`
        &::before {
            position: absolute;
            z-index: 3;
            top: 0;
            right: 8px;
            left: 8px;
            height: 2px;
            border-radius: 1px;
            background: `,`;
            content: "";
            pointer-events: none;
        }
    `])),r.colorPrimary),dropAfter:i(We||(We=R()([`
        &::after {
            position: absolute;
            z-index: 3;
            right: 8px;
            bottom: 0;
            left: 8px;
            height: 2px;
            border-radius: 1px;
            background: `,`;
            content: "";
            pointer-events: none;
        }
    `])),r.colorPrimary),handle:i(Qe||(Qe=R()([`
        display: inline-flex;
        width: `,`px;
        height: `,`px;
        align-items: center;
        justify-content: center;
        padding: 0;
        border: 0;
        border-radius: `,`px;
        outline: none;
        background: transparent;
        color: `,`;
        cursor: grab;
        touch-action: none;
        user-select: none;

        &:hover,
        &:focus-visible {
            background: `,`;
            color: `,`;
        }

        &:active {
            cursor: grabbing;
        }

        &:disabled {
            background: transparent;
            color: `,`;
            cursor: default;
            opacity: .45;
        }

        @media (pointer: coarse) {
            width: 34px;
            height: 34px;
        }
    `])),e.compact?28:32,e.compact?28:32,r.borderRadiusSM,r.colorTextQuaternary,r.colorFillTertiary,r.colorTextSecondary,r.colorTextQuaternary),name:i(Me||(Me=R()([`
        display: block;
        min-width: 0;
        overflow: hidden;
        color: `,`;
        font-size: `,`px;
        font-weight: 400;
        text-overflow: ellipsis;
        white-space: nowrap;
    `])),r.colorText,e.compact?12:13),actions:i(ke||(ke=R()([`
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 2px;
    `]))),scope:i(ge||(ge=R()([`
        margin-inline: 6px;
        color: `,`;
        font-size: `,`px;
        white-space: nowrap;
    `])),r.colorTextSecondary,e.compact?11:12),actionButton:i(be||(be=R()([`
        width: `,`px;
        min-width: `,`px;
        height: `,`px;
        padding: 0;

        @media (pointer: coarse) {
            width: 34px;
            min-width: 34px;
            height: 34px;
        }
    `])),e.compact?27:30,e.compact?27:30,e.compact?27:30)}}),kn=(0,u.memo)(function(g){var e=g.actionDisabled,i=g.actionButtonClassName,r=g.actionsClassName,l=g.disabled,o=g.handleClassName,w=g.itemClassName,s=g.nameClassName,k=g.operating,h=g.record,M=g.scopeLabel,x=g.scopeClassName,T=g.onEdit,Z=g.onRemove,he=g.onPointerDown,C=g.onCancelDrag,nn=g.onKeyboardMove;return(0,c.jsxs)("div",{"data-type-id":String(h.id),className:w,children:[(0,c.jsx)("button",{type:"button",className:o,"aria-label":"\u62D6\u52A8\u5206\u7C7B","data-drag-handle":!0,disabled:l,onPointerDown:function(q){return he(h.id,q)},onKeyDown:function(q){if(q.key==="Escape"){C();return}q.key!=="ArrowUp"&&q.key!=="ArrowDown"||(q.preventDefault(),nn(h.id,q.key==="ArrowUp"?-1:1))},children:(0,c.jsx)(D.In,{type:"HolderOutlined"})}),(0,c.jsx)("span",{className:s,title:h.name,children:h.name}),(0,c.jsxs)("div",{className:r,"data-row-action":!0,children:[M&&(0,c.jsx)("span",{className:x,children:M}),(0,c.jsx)(j.A,{menu:{items:[{key:"edit",label:"\u7F16\u8F91\u5206\u7C7B"},{type:"divider"},{key:"delete",label:"\u5220\u9664\u5206\u7C7B",danger:!0}],onClick:function(q){var rn=q.key;rn==="edit"?T(h):rn==="delete"&&Z(h)}},children:(0,c.jsx)(ue.A,{className:i,type:"text","aria-label":"\u7BA1\u7406\u5206\u7C7B ".concat(h.name),icon:(0,c.jsx)(D.In,{type:"MoreOutlined"}),loading:k,disabled:e})})]})]})});kn.displayName="DraggableRow";var An=(0,u.forwardRef)(function(g,e){var i,r=g.module,l=g.scopes,o=g.zIndex,w=g.onChange,s=g.onMutation,k=re.A.useApp(),h=k.message,M=k.modal,x=(0,D.DP)(),T=!!(!(x==null||(i=x.isCompactTheme)===null||i===void 0)&&i.call(x)),Z=(0,u.useMemo)(function(){return{compact:T}},[T]),he=Vn(Z),C=he.styles,nn=(0,u.useRef)({dragging:C.dragging,dropAfter:C.dropAfter,dropBefore:C.dropBefore,shifted:C.shifted});nn.current={dragging:C.dragging,dropAfter:C.dropAfter,dropBefore:C.dropBefore,shifted:C.shifted};var Ye=(0,u.useRef)({}),q=ae.A.useForm(),rn=L()(q,1),F=rn[0],Te=(0,u.useRef)(0),Y=(0,u.useRef)(0),K=(0,u.useRef)(!1),ee=(0,u.useRef)(!1),Oe=(0,u.useRef)(0),Be=(0,u.useRef)(!1),Xe=(0,u.useRef)(!0),X=(0,u.useRef)(null),Ee=(0,u.useRef)(null),Hn=(0,u.useState)([]),Tn=L()(Hn,2),xe=Tn[0],Se=Tn[1],ce=(0,u.useRef)(xe);ce.current=xe;var Qn=(0,u.useState)(!1),On=L()(Qn,2),Le=On[0],sn=On[1],Zn=(0,u.useState)(!1),Bn=L()(Zn,2),an=Bn[0],Fe=Bn[1],qn=(0,u.useState)(!1),xn=L()(qn,2),er=xn[0],ye=xn[1],nr=(0,u.useState)(),Sn=L()(nr,2),Ie=Sn[0],Ue=Sn[1],rr=(0,u.useState)(!1),Ln=L()(rr,2),on=Ln[0],$e=Ln[1],ar=(0,u.useState)(!1),Fn=L()(ar,2),tr=Fn[0],dn=Fn[1],ir=(0,u.useState)(null),Un=L()(ir,2),cn=Un[0],mn=Un[1],ur=(0,u.useState)(null),lr=L()(ur,2),Ge=lr[1],_n=(0,u.useRef)(Le),fn=(0,u.useRef)(an);_n.current=Le,fn.current=an;var we=Ie!==void 0,Kn=on||tr||cn!==null,vn=Kn||Le||an,Nn=on||cn!==null||an,pn=(0,u.useRef)(vn);pn.current=vn,(0,u.useEffect)(function(){return Xe.current=!0,function(){Xe.current=!1,Y.current+=1,Te.current+=1,Oe.current+=1}},[]);var z=(0,u.useCallback)(function(){var a=X.current;X.current=null,a!=null&&a.frameId&&window.cancelAnimationFrame(a.frameId),a&&(a.ended=!0,qe(a)),a&&en(a),Ge(null)},[]),ne=(0,u.useCallback)(function(a,n,f){if(!n&&(f==null?void 0:f.size)===0){ce.current=a,Se(a);return}var v=Ee.current,P=n!=null?n:v?un(v,f):void 0;v&&ln(v),(0,_e.flushSync)(function(){ce.current=a,Ge(null),Se(a)}),v&&P&&Xn(v,P)},[]);(0,u.useEffect)(function(){var a=function(){return z()};return window.addEventListener("blur",a),function(){window.removeEventListener("blur",a);var n=X.current;X.current=null,n!=null&&n.frameId&&window.cancelAnimationFrame(n.frameId),n&&(n.ended=!0,qe(n)),n&&en(n),Ee.current&&ln(Ee.current)}},[z]);var $=(0,u.useCallback)(Q()(B()().mark(function a(){var n,f,v,P,b,d,p,I=arguments;return B()().wrap(function(S){for(;;)switch(S.prev=S.next){case 0:return n=I.length>0&&I[0]!==void 0?I[0]:!1,f=I.length>1&&I[1]!==void 0?I[1]:!1,v=++Te.current,sn(!0),S.prev=4,S.next=7,(0,m.Jt)({body:{module:r,order_by_field:"sort",order_by_type:"asc"}});case 7:if(d=S.sent,!(Te.current!==v||!d)){S.next=10;break}return S.abrupt("return",void 0);case 10:if((d==null?void 0:d.code)===200){S.next=13;break}return h.error((d==null?void 0:d.message)||"\u83B7\u53D6\u5206\u7C7B\u5931\u8D25"),S.abrupt("return",void 0);case 13:return p=Array.isArray(d==null||(P=d.data)===null||P===void 0?void 0:P.list)?d==null||(b=d.data)===null||b===void 0?void 0:b.list.filter(function(le){return le==null?void 0:le.id}):[],f?ne(p,void 0,$n(ce.current,p)):(ce.current=p,Se(p)),Be.current=!1,n&&(w==null||w(p)),S.abrupt("return",p);case 18:return S.prev=18,Te.current===v&&sn(!1),S.finish(18);case 21:case"end":return S.stop()}},a,null,[[4,,18,21]])})),[ne,h,r,w]),Wn=(0,u.useCallback)(function(){K.current||ee.current||(z(),Fe(!1),ye(!1),F.resetFields(),Ue(void 0),(0,y.Rp)(r),$(!0))},[z,F,$,r]),sr=(0,u.useCallback)(function(){var a=Be.current;Y.current+=1,Te.current+=1,Oe.current+=1,K.current=!1,ee.current=!1,z(),F.resetFields(),Se([]),sn(!1),Fe(!1),ye(!1),Ue(void 0),$e(!1),dn(!1),mn(null),a&&Xe.current&&((0,y.Rp)(r),$(!0))},[z,F,$,r]),Yn=(0,u.useCallback)(function(){var a;Y.current+=1,Te.current+=1,Oe.current+=1,K.current=!1,ee.current=!1,F.resetFields(),Fe(!1),ye(!1),Ue(void 0),Se([]),(a=Ye.current)===null||a===void 0||a.show(),(0,y.Rp)(r),$(!0)},[F,$,r]);(0,u.useImperativeHandle)(e,function(){return{open:Yn,refresh:Wn}},[Yn,Wn]);var Ke=(0,u.useCallback)(Q()(B()().mark(function a(){return B()().wrap(function(f){for(;;)switch(f.prev=f.next){case 0:return(0,y.Rp)(r),f.next=3,$(!0);case 3:case"end":return f.stop()}},a)})),[$,r]),or=(0,u.useCallback)(function(){z(),Ue(void 0),F.resetFields(),F.setFieldsValue({name:"",module:r}),Fe(!0),ye(!0)},[z,F,r]),dr=(0,u.useCallback)(function(a){K.current||ee.current||X.current||_n.current||fn.current||(z(),Ue(a),F.resetFields(),F.setFieldsValue({name:a.name}),Fe(!0),ye(!0))},[z,F]),cr=(0,u.useCallback)(function(){Y.current+=1,K.current=!1,$e(!1),ye(!1)},[]),mr=(0,u.useCallback)(function(){K.current=!1,F.resetFields(),Ue(void 0),Fe(!1),$e(!1)},[F]),_r=(0,u.useCallback)((function(){var a=Q()(B()().mark(function n(f){var v,P,b,d;return B()().wrap(function(I){for(;;)switch(I.prev=I.next){case 0:if(!K.current){I.next=2;break}return I.abrupt("return");case 2:return v=Y.current,P=f.name.trim(),b=(Ie==null?void 0:Ie.module)||f.module||r,K.current=!0,$e(!0),I.prev=7,I.next=10,Ie?(0,m.Gg)({body:{ids:[Ie.id],name:P}}):(0,m.rv)({body:{module:b,name:P}});case 10:if(d=I.sent,Y.current===v){I.next=18;break}if((d==null?void 0:d.code)!==200){I.next=17;break}return(0,y.Rp)(b),s==null||s(),I.next=17,Ke();case 17:return I.abrupt("return");case 18:if(d){I.next=20;break}return I.abrupt("return");case 20:if((d==null?void 0:d.code)===200){I.next=23;break}return h.error((d==null?void 0:d.message)||(we?"\u4FEE\u6539\u5206\u7C7B\u5931\u8D25":"\u6DFB\u52A0\u5206\u7C7B\u5931\u8D25")),I.abrupt("return");case 23:return ye(!1),(0,y.Rp)(b),h.success((d==null?void 0:d.message)||(we?"\u4FEE\u6539\u6210\u529F":"\u6DFB\u52A0\u6210\u529F")),s==null||s(),I.next=29,Ke();case 29:return I.prev=29,Y.current===v&&(K.current=!1,$e(!1)),I.finish(29);case 32:case"end":return I.stop()}},n,null,[[7,,29,32]])}));return function(n){return a.apply(this,arguments)}})(),[we,Ie,h,r,s,Ke]),fr=(0,u.useCallback)(function(a){if(!(K.current||ee.current||X.current||_n.current||fn.current)){var n=Y.current;M.confirm({title:"\u5220\u9664\u5206\u7C7B",zIndex:o===void 0?void 0:o+20,content:"\u5220\u9664\u201C".concat(a.name,"\u201D\u540E\uFF0C\u5173\u8054\u9879\u5C06\u6062\u590D\u4E3A\u5168\u90E8\u5206\u7C7B\u3002"),okText:"\u5220\u9664",cancelText:"\u53D6\u6D88",okButtonProps:{danger:!0,type:"default"},mask:{closable:!0},onOk:(function(){var f=Q()(B()().mark(function P(){var b;return B()().wrap(function(p){for(;;)switch(p.prev=p.next){case 0:if(!(Y.current!==n||K.current)){p.next=2;break}return p.abrupt("return");case 2:return K.current=!0,mn(a.id),p.prev=4,p.next=7,(0,m.SJ)({body:{ids:[a.id]}});case 7:if(b=p.sent,Y.current===n){p.next=15;break}if((b==null?void 0:b.code)!==200){p.next=14;break}return(0,y.Rp)(r),s==null||s(),p.next=14,Ke();case 14:return p.abrupt("return");case 15:if(b){p.next=17;break}return p.abrupt("return");case 17:if((b==null?void 0:b.code)===200){p.next=20;break}return h.error((b==null?void 0:b.message)||"\u5220\u9664\u5206\u7C7B\u5931\u8D25"),p.abrupt("return");case 20:return h.success((b==null?void 0:b.message)||"\u5220\u9664\u6210\u529F"),s==null||s(),p.next=24,Ke();case 24:return p.prev=24,Y.current===n&&(K.current=!1,mn(null)),p.finish(24);case 27:case"end":return p.stop()}},P,null,[[4,,24,27]])}));function v(){return f.apply(this,arguments)}return v})()})}},[h,M,r,s,Ke,o]),tn=(0,u.useCallback)((function(){var a=Q()(B()().mark(function n(f,v,P){var b,d,p,I,Pe,S,le,ie,ze,yn,Je,H;return B()().wrap(function(E){for(;;)switch(E.prev=E.next){case 0:if(b=Y.current,d=f.findIndex(function(me){return me.id===P}),p=v.findIndex(function(me){return me.id===P}),I=Cn(f,d,p),Pe=function(){return ne(f,void 0,I)},!(K.current||ee.current)){E.next=8;break}return Pe(),E.abrupt("return");case 8:if(!(d<0||p<0)){E.next=11;break}return Pe(),E.abrupt("return");case 11:if(d!==p){E.next=13;break}return E.abrupt("return");case 13:return S=f[d],le=f[p],ie=Number(le.sort)+(p<d?-1:1),ze=v.map(function(me){return me.id===S.id?se()(se()({},me),{},{sort:ie}):me}),ce.current=ze,Se(ze),yn=++Oe.current,ee.current=!0,dn(!0),Je=!1,E.prev=23,E.next=26,(0,m.Gg)({body:{ids:[S.id],sort:String(ie)}});case 26:if(H=E.sent,Je=(H==null?void 0:H.code)===200,Y.current===b){E.next=34;break}if(!Je){E.next=33;break}return(0,y.Rp)(r),E.next=33,$(!0,!0);case 33:return E.abrupt("return");case 34:if(Je&&(Be.current=!0),Oe.current===yn){E.next=37;break}return E.abrupt("return");case 37:if(!(!H||(H==null?void 0:H.code)!==200)){E.next=44;break}return Pe(),H&&h.error((H==null?void 0:H.message)||"\u6392\u5E8F\u5931\u8D25"),(0,y.Rp)(r),E.next=43,$(!0,!0);case 43:return E.abrupt("return");case 44:return(0,y.Rp)(r),E.next=47,$(!0,!0);case 47:if(E.prev=47,Y.current!==b){E.next=69;break}if(Oe.current!==yn){E.next=59;break}if(ee.current=!1,dn(!1),!Be.current){E.next=57;break}if((0,y.Rp)(r),!Xe.current){E.next=57;break}return E.next=57,$(!0,!0);case 57:E.next=69;break;case 59:if(!Je){E.next=69;break}if(Be.current=!0,(0,y.Rp)(r),!ee.current){E.next=66;break}Be.current=!0,E.next=69;break;case 66:if(!Xe.current){E.next=69;break}return E.next=69,$(!0,!0);case 69:return E.finish(47);case 70:case"end":return E.stop()}},n,null,[[23,,47,70]])}));return function(n,f,v){return a.apply(this,arguments)}})(),[ne,$,h,r]),vr=(0,u.useCallback)(function(a,n){if(!(X.current||!n.isPrimary||pn.current||K.current||ee.current||n.pointerType==="mouse"&&n.button!==0)){Ee.current&&ln(Ee.current),n.preventDefault();var f=n.currentTarget,v=ce.current.slice();X.current={activeId:a,activeIndex:-1,activated:!1,classes:se()({},nn.current),clientX:n.clientX,clientY:n.clientY,ended:!1,frameId:0,handle:f,lastFrameTime:0,list:null,listRect:null,originItems:v,overIndex:-1,pointerId:n.pointerId,rowGap:0,rows:[],startScrollTop:0,startX:n.clientX,startY:n.clientY};try{f.setPointerCapture(n.pointerId)}catch(P){}}},[]),gn=(0,u.useCallback)(function(a){var n=X.current;if(!(!n||n.pointerId!==a.pointerId)){if(n.clientX=a.clientX,n.clientY=a.clientY,!n.activated){var f=Math.hypot(a.clientX-n.startX,a.clientY-n.startY);if(f<6)return;var v=Ee.current;if(!v){z();return}var P=v.getBoundingClientRect(),b=Ze(v),d=b.findIndex(function(ie){return Number(ie.dataset.typeId)===n.activeId});if(d<0){z();return}var p=Number.parseFloat(window.getComputedStyle(v).rowGap)||0,I=b[d].getBoundingClientRect(),Pe=I.top-P.top+v.scrollTop,S=I.height+p,le=b.map(function(ie,ze){return{center:Pe+(ze-d)*S+I.height/2,element:ie,height:I.height,id:Number(ie.dataset.typeId)}}).filter(function(ie){return Number.isFinite(ie.id)});n.activated=!0,n.activeIndex=d,n.list=v,n.listRect=P,n.overIndex=d,n.rowGap=p,n.rows=le,n.startScrollTop=v.scrollTop,le[d].element.classList.add(n.classes.dragging),Ge(n.activeId)}a.preventDefault(),Jn(n)}},[z]),bn=(0,u.useCallback)(function(){var a=X.current;if(a){var n=Rn(a),f=a.activated&&a.list?un(a.list,n):void 0,v=a.activated&&a.list?jn(a.list,n):void 0;X.current=null,a.ended=!0,a.frameId&&window.cancelAnimationFrame(a.frameId),en(a),qe(a),f?ne(a.originItems,f):Ge(null),v==null||v()}},[ne]),hn=(0,u.useCallback)(function(a){var n=X.current;!n||n.pointerId!==a.pointerId||bn()},[bn]),En=(0,u.useCallback)(function(a){var n=X.current;if(!(!n||n.pointerId!==a.pointerId)){n.clientX=a.clientX,n.clientY=a.clientY,n.frameId&&(window.cancelAnimationFrame(n.frameId),n.frameId=0),n.activated&&(Mn(n,performance.now()),n.frameId&&(window.cancelAnimationFrame(n.frameId),n.frameId=0));var f=Rn(n),v=n.activated&&n.list?un(n.list,f):void 0,P=n.activated&&n.list?jn(n.list,f):void 0,b=n.activeIndex,d=n.overIndex;if(X.current=null,n.ended=!0,en(n),qe(n),!n.activated||b<0||d<0||b===d){v?ne(n.originItems,v):Ge(null),P==null||P();return}var p=wn(n.originItems,b,d);ne(p,v),P==null||P(),tn(n.originItems,p,n.activeId)}},[ne,tn]);(0,u.useEffect)(function(){return window.addEventListener("pointermove",gn,{passive:!1}),window.addEventListener("pointerup",En),window.addEventListener("pointercancel",hn),function(){window.removeEventListener("pointermove",gn),window.removeEventListener("pointerup",En),window.removeEventListener("pointercancel",hn)}},[hn,En,gn]);var pr=(0,u.useCallback)(function(a,n){if(!(pn.current||K.current||ee.current)){var f=ce.current,v=f.findIndex(function(p){return p.id===a}),P=v+n;if(!(v<0||P<0||P>=f.length)){var b=wn(f,v,P),d=Cn(f,v,P);ne(b,void 0,d),tn(f,b,a)}}},[ne,tn]);return(0,c.jsx)(D.Ht,{ref:Ye,title:(0,c.jsx)(D.hE,{children:"\u5206\u7C7B\u7BA1\u7406"}),onOk:or,width:480,modalProps:{rootClassName:C.modal,zIndex:o,okText:"\u6DFB\u52A0",cancelText:"\u53D6\u6D88",closable:!0,keyboard:!0,forceRender:!0,footer:Le&&xe.length===0?null:void 0,okButtonProps:{disabled:Le||Kn},mask:{closable:!0},afterClose:sr},children:(0,c.jsxs)(c.Fragment,{children:[(0,c.jsx)("div",{className:C.content,children:(0,c.jsx)("div",{className:C.listShell,children:(0,c.jsx)("div",{ref:Ee,className:[C.list,vn?C.listLocked:""].filter(Boolean).join(" "),children:Le&&xe.length===0?(0,c.jsx)("div",{className:C.loading,children:(0,c.jsx)(G.A,{size:"medium"})}):xe.length===0?(0,c.jsx)("div",{className:C.empty,children:(0,c.jsx)(fe.A,{image:fe.A.PRESENTED_IMAGE_SIMPLE,description:"\u6682\u65E0\u6570\u636E"})}):(0,c.jsx)(c.Fragment,{children:xe.map(function(a){var n;return(0,c.jsx)(kn,{record:a,disabled:Nn,actionDisabled:Nn,actionButtonClassName:C.actionButton,actionsClassName:C.actions,handleClassName:C.handle,itemClassName:C.item,nameClassName:C.name,scopeLabel:l==null||(n=l.find(function(f){return f.value===a.module}))===null||n===void 0?void 0:n.label,scopeClassName:C.scope,operating:cn===a.id,onEdit:dr,onRemove:fr,onPointerDown:vr,onCancelDrag:bn,onKeyboardMove:pr},a.id)})})})})}),(0,c.jsx)(D.Ht,{title:(0,c.jsx)(D.hE,{children:we?"\u7F16\u8F91\u5206\u7C7B":"\u6DFB\u52A0\u5206\u7C7B"}),width:"400px",onOk:F.submit,onCancel:cr,okText:we?"\u4FDD\u5B58":"\u6DFB\u52A0",afterClose:mr,modalProps:{rootClassName:C.modal,zIndex:o===void 0?void 0:o+10,open:er,cancelText:"\u53D6\u6D88",confirmLoading:on,closable:!0,keyboard:!0,mask:{closable:!0},forceRender:!0,focusable:{focusTriggerAfterClose:!1},styles:{body:{paddingTop:T?10:15}}},children:(0,c.jsxs)(ae.A,{form:F,layout:"vertical",variant:"filled",onFinish:_r,children:[(0,c.jsx)(ae.A.Item,{name:"name",label:!we&&l&&l.length>1?"\u5206\u7C7B\u540D\u79F0":void 0,rules:[{required:!0,whitespace:!0,message:"\u8BF7\u8F93\u5165\u5206\u7C7B\u540D\u79F0"},{max:50,message:"\u5206\u7C7B\u540D\u79F0\u4E0D\u80FD\u8D85\u8FC7 50 \u4E2A\u5B57\u7B26"}],children:(0,c.jsx)(ve.A,{maxLength:50,"aria-label":"\u5206\u7C7B\u540D\u79F0",placeholder:"\u8BF7\u8F93\u5165\u5206\u7C7B\u540D\u79F0",autoComplete:"off",onPressEnter:F.submit})}),!we&&l&&l.length>1&&(0,c.jsx)(ae.A.Item,{name:"module",label:"\u4FDD\u5B58\u8303\u56F4",rules:[{required:!0,message:"\u8BF7\u9009\u62E9\u4FDD\u5B58\u8303\u56F4"}],children:(0,c.jsx)(pe.A,{placeholder:"\u8BF7\u9009\u62E9\u4FDD\u5B58\u8303\u56F4",options:l})})]})})]})})});An.displayName="TypeManager",De.A=(0,u.memo)(An)}),8490:(function(In,De,_){_.d(De,{Gg:function(){return fe},Jt:function(){return u},SJ:function(){return ve},rv:function(){return re}});var je,se,Ne,B=_(55420),J=_.n(B),Q=_(98312),oe=_.n(Q),R=_(78781);function Ve(){return L.apply(this,arguments)}function L(){return L=se(je().mark(function G(){var t,D=arguments;return je().wrap(function(m){for(;;)switch(m.prev=m.next){case 0:return t=D.length>0&&D[0]!==void 0?D[0]:{},m.abrupt("return",Ne("/type/enum",t==null?void 0:t.body,t==null?void 0:t.onSuccess,t==null?void 0:t.onFail,t==null?void 0:t.onFinally));case 2:case"end":return m.stop()}},G)})),L.apply(this,arguments)}function u(){return _e.apply(this,arguments)}function _e(){return _e=oe()(J()().mark(function G(){var t,D=arguments;return J()().wrap(function(m){for(;;)switch(m.prev=m.next){case 0:return t=D.length>0&&D[0]!==void 0?D[0]:{},m.abrupt("return",(0,R.Jt)("/type/list",t==null?void 0:t.body,t==null?void 0:t.onSuccess,t==null?void 0:t.onFail,t==null?void 0:t.onFinally));case 2:case"end":return m.stop()}},G)})),_e.apply(this,arguments)}function re(){return ue.apply(this,arguments)}function ue(){return ue=oe()(J()().mark(function G(){var t,D=arguments;return J()().wrap(function(m){for(;;)switch(m.prev=m.next){case 0:return t=D.length>0&&D[0]!==void 0?D[0]:{},m.abrupt("return",(0,R.bE)("/type/create",t==null?void 0:t.body,t==null?void 0:t.onSuccess,t==null?void 0:t.onFail,t==null?void 0:t.onFinally));case 2:case"end":return m.stop()}},G)})),ue.apply(this,arguments)}function fe(){return ae.apply(this,arguments)}function ae(){return ae=oe()(J()().mark(function G(){var t,D=arguments;return J()().wrap(function(m){for(;;)switch(m.prev=m.next){case 0:return t=D.length>0&&D[0]!==void 0?D[0]:{},m.abrupt("return",(0,R.bE)("/type/update",t==null?void 0:t.body,t==null?void 0:t.onSuccess,t==null?void 0:t.onFail,t==null?void 0:t.onFinally));case 2:case"end":return m.stop()}},G)})),ae.apply(this,arguments)}function ve(){return pe.apply(this,arguments)}function pe(){return pe=oe()(J()().mark(function G(){var t,D=arguments;return J()().wrap(function(m){for(;;)switch(m.prev=m.next){case 0:return t=D.length>0&&D[0]!==void 0?D[0]:{},m.abrupt("return",(0,R.bE)("/type/delete",t==null?void 0:t.body,t==null?void 0:t.onSuccess,t==null?void 0:t.onFail,t==null?void 0:t.onFinally));case 2:case"end":return m.stop()}},G)})),pe.apply(this,arguments)}}),9300:(function(In,De,_){_.d(De,{Rp:function(){return pe}});var je=_(1974),se=_.n(je),Ne=_(55420),B=_.n(Ne),J=_(11604),Q=_.n(J),oe=_(98312),R=_.n(oe),Ve=_(45647),L=_.n(Ve),u=_(96540),_e=_(72523),re={};function ue(j){var m=arguments.length>1&&arguments[1]!==void 0?arguments[1]:[],y=Array.isArray(j)?j:[j].concat(L()(m));return Array.from(new Set(y.map(function(c){return c==null?void 0:c.trim()}).filter(Boolean)))}function fe(j){return j.reduce(function(m,y){return re[y]&&(m[y]=re[y]),m},{})}function ae(j){return ve.apply(this,arguments)}function ve(){return ve=R()(B()().mark(function j(m){var y,c,U;return B()().wrap(function(A){for(;;)switch(A.prev=A.next){case 0:if(!(m.length<=0)){A.next=2;break}return A.abrupt("return",{});case 2:if(y=fe(m),c=m.filter(function(N){return!y[N]}),!(c.length<=0)){A.next=6;break}return A.abrupt("return",y);case 6:return A.next=8,Promise.all(c.map((function(){var N=R()(B()().mark(function de(O){var W;return B()().wrap(function(V){for(;;)switch(V.prev=V.next){case 0:return V.next=2,(0,_e.GV)({body:{name:O}});case 2:if(W=V.sent,(W==null?void 0:W.code)===200){V.next=5;break}throw new Error((W==null?void 0:W.message)||"\u83B7\u53D6\u679A\u4E3E\u6570\u636E\u5931\u8D25");case 5:return V.abrupt("return",{name:O,data:(W==null?void 0:W.data)||{}});case 6:case"end":return V.stop()}},de)}));return function(de){return N.apply(this,arguments)}})()));case 8:return U=A.sent,U==null||U.forEach(function(N){re[N.name]=N.data}),A.abrupt("return",se()(se()({},y),fe(c)));case 11:case"end":return A.stop()}},j)})),ve.apply(this,arguments)}function pe(j){if(j){delete re[j];return}Object.keys(re).forEach(function(m){delete re[m]})}function G(j){return t.apply(this,arguments)}function t(){return t=R()(B()().mark(function j(m){var y,c,U,te,A,N=arguments;return B()().wrap(function(O){for(;;)switch(O.prev=O.next){case 0:for(y=N.length,c=new Array(y>1?y-1:0),U=1;U<y;U++)c[U-1]=N[U];if(te=ue(m,c),!(te.length<=0)){O.next=4;break}return O.abrupt("return",{});case 4:return O.next=6,ae(te);case 6:if(A=O.sent,!(Array.isArray(m)||c.length>0)){O.next=9;break}return O.abrupt("return",A);case 9:return O.abrupt("return",(A==null?void 0:A[te[0]])||{});case 10:case"end":return O.stop()}},j)})),t.apply(this,arguments)}function D(j){for(var m=arguments.length,y=new Array(m>1?m-1:0),c=1;c<m;c++)y[c-1]=arguments[c];var U=Array.isArray(j)||y.length>0,te=(0,u.useMemo)(function(){return ue(j,y).join(",")},[Array.isArray(j)?j.join(","):j,y.join(",")]),A=(0,u.useState)({}),N=Q()(A,2),de=N[0],O=N[1],W=(0,u.useState)(!0),Re=Q()(W,2),V=Re[0],Ce=Re[1],He=(0,u.useState)(null),We=Q()(He,2),Qe=We[0],Me=We[1];return(0,u.useEffect)(function(){var ke=ue(j,y);if(ke.length<=0){O({}),Ce(!1),Me(null);return}var ge=!0;return Ce(!0),Me(null),G(U?ke:ke[0]).then(function(be){ge&&O(be)}).catch(function(be){ge&&(O({}),Me(be instanceof Error?be:new Error("\u83B7\u53D6\u679A\u4E3E\u6570\u636E\u5931\u8D25")))}).finally(function(){ge&&Ce(!1)}),function(){ge=!1}},[U,te]),[de,V,Qe]}De.Ay=D})}]);
