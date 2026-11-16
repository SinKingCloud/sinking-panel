import {useCallback, useEffect, useLayoutEffect, useRef, useState} from "react";
import type {RefObject} from "react";

// 多个弹窗同时全屏时，共用祖先的样式要等最后一个退出后再恢复。
const styleLocks = new WeakMap<HTMLElement, Map<string, {value: string; priority: string; count: number}>>();
const pageStack: symbol[] = [];
let nativeOwner: {key: symbol; request: number} | null = null;

const useFullscreen = (targetRef: RefObject<HTMLElement | null>, {active = true, pageStyles = true} = {}) => {
    const [nativeFullscreen, setNativeFullscreen] = useState(false);
    const [pageFullscreen, setPageFullscreen] = useState(false);
    const [pending, setPending] = useState(false);
    const mounted = useRef(false);
    const activeRef = useRef(active);
    const pageRef = useRef(false);
    const pendingRef = useRef(false);
    const request = useRef(0);
    const owner = useRef<HTMLElement | null>(null);
    const nativeRequest = useRef<typeof nativeOwner>(null);
    const key = useRef(Symbol());
    activeRef.current = active;
    pageRef.current = pageFullscreen;

    const exit = useCallback(async () => {
        const current = ++request.current;
        const entering = nativeRequest.current && nativeOwner === nativeRequest.current;
        pendingRef.current = false;
        pageRef.current = false;
        if (mounted.current) {
            setPageFullscreen(false);
            setPending(false);
        }
        const target = owner.current;
        if (target && nativeOwner?.key === key.current && document.fullscreenElement === target) {
            try { await document.exitFullscreen(); } catch { /* 保持实际全屏状态，允许再次退出。 */ }
        }
        if (current !== request.current) return;
        const native = Boolean(target && nativeOwner?.key === key.current && document.fullscreenElement === target);
        if (mounted.current) setNativeFullscreen(native);
        if (!entering && !native && owner.current === target) {
            owner.current = null;
            if (nativeOwner?.key === key.current) nativeOwner = null;
        }
    }, []);

    const toggle = useCallback(async () => {
        const target = targetRef.current;
        if (!mounted.current || !activeRef.current || !target || pendingRef.current) return;
        if (pageRef.current || (owner.current === target && nativeOwner?.key === key.current && document.fullscreenElement === target)) {
            await exit();
            return;
        }
        const current = ++request.current;
        const ticket = {key: key.current, request: current};
        let requested = false;
        owner.current = target;
        pendingRef.current = true;
        setPending(true);
        const valid = () => mounted.current && activeRef.current && current === request.current
            && targetRef.current === target && target.isConnected;
        try {
            if (document.fullscreenEnabled && typeof target.requestFullscreen === "function" && typeof document.exitFullscreen === "function") {
                requested = true;
                nativeOwner = ticket;
                nativeRequest.current = ticket;
                try { await target.requestFullscreen(); } catch { /* 权限或浏览器限制时降级为网页全屏。 */ }
            }
            if (!valid()) {
                if ((nativeOwner === ticket || !nativeOwner) && document.fullscreenElement === target) {
                    try { await document.exitFullscreen(); } catch { /* 节点移除时浏览器会自动退出。 */ }
                }
                if (nativeOwner === ticket && document.fullscreenElement !== target) nativeOwner = null;
                return;
            }
            // 多个管理器使用同一个 documentElement，旧请求不能接管新窗口的全屏。
            if (requested && nativeOwner !== ticket) return;
            const native = nativeOwner === ticket && document.fullscreenElement === target;
            setNativeFullscreen(native);
            if (!native && nativeOwner === ticket) nativeOwner = null;
            pageRef.current = !native;
            setPageFullscreen(!native);
        } finally {
            if (nativeRequest.current === ticket) nativeRequest.current = null;
            if (mounted.current && current === request.current) {
                pendingRef.current = false;
                setPending(false);
            }
        }
    }, [exit, targetRef]);

    useLayoutEffect(() => {
        mounted.current = true;
        const change = () => {
            const native = Boolean(owner.current && nativeOwner?.key === key.current && document.fullscreenElement === owner.current);
            setNativeFullscreen(native && activeRef.current && nativeOwner?.request === request.current);
            if (!native && !pendingRef.current && !nativeRequest.current) {
                owner.current = null;
                if (nativeOwner?.key === key.current) nativeOwner = null;
            }
        };
        document.addEventListener("fullscreenchange", change);
        return () => {
            mounted.current = false;
            document.removeEventListener("fullscreenchange", change);
            void exit();
        };
    }, [exit]);

    useLayoutEffect(() => {
        if (!active) void exit();
    }, [active, exit]);

    useLayoutEffect(() => {
        const target = targetRef.current;
        if (!pageFullscreen || !target) return;
        const id = key.current;
        const releases: (() => void)[] = [];
        const apply = (element: HTMLElement, values: Record<string, string>) => {
            let locks = styleLocks.get(element);
            if (!locks) { locks = new Map(); styleLocks.set(element, locks); }
            for (const [property, value] of Object.entries(values)) {
                let lock = locks.get(property);
                if (!lock) {
                    // 弹窗已锁定滚动时由弹窗负责恢复，避免关闭后重新写回 hidden。
                    if (property === "overflow" && element.style.getPropertyValue(property) === value) continue;
                    lock = {value: element.style.getPropertyValue(property), priority: element.style.getPropertyPriority(property), count: 0};
                    locks.set(property, lock);
                    element.style.setProperty(property, value, "important");
                }
                lock.count += 1;
                const saved = lock;
                releases.push(() => {
                    if (--saved.count) return;
                    if (saved.value) element.style.setProperty(property, saved.value, saved.priority);
                    else element.style.removeProperty(property);
                    locks.delete(property);
                    if (!locks.size) styleLocks.delete(element);
                });
            }
        };
        pageStack.push(id);
        if (pageStyles) {
            // 不移动 DOM，避免终端断开或媒体重载；移除祖先对 fixed 定位的限制。
            for (let element = target.parentElement; element; element = element.parentElement) {
                apply(element, {
                    transform: "none", translate: "none", rotate: "none", scale: "none", perspective: "none",
                    filter: "none", "backdrop-filter": "none", "-webkit-backdrop-filter": "none",
                    contain: "none", "container-type": "normal", "content-visibility": "visible", "will-change": "auto",
                });
            }
            apply(target, {
                position: "fixed", inset: "0", width: "100vw", height: CSS.supports("height", "100dvh") ? "100dvh" : "100vh",
                "max-width": "none", "max-height": "none", "min-height": "0", margin: "0", "box-sizing": "border-box", "z-index": "2000",
            });
        }
        apply(document.documentElement, {overflow: "hidden"});
        apply(document.body, {overflow: "hidden"});
        return () => {
            const index = pageStack.indexOf(id);
            if (index >= 0) pageStack.splice(index, 1);
            releases.reverse().forEach((release) => release());
        };
    }, [pageFullscreen, pageStyles, targetRef]);

    useEffect(() => {
        if (!pageFullscreen) return;
        const escape = (event: KeyboardEvent) => {
            if (event.key !== "Escape" || pageStack[pageStack.length - 1] !== key.current) return;
            event.preventDefault();
            event.stopImmediatePropagation();
            void exit();
        };
        document.addEventListener("keydown", escape, true);
        return () => document.removeEventListener("keydown", escape, true);
    }, [exit, pageFullscreen]);

    return {fullscreen: nativeFullscreen || pageFullscreen, pageFullscreen, pending, toggle, exit};
};

export default useFullscreen;
