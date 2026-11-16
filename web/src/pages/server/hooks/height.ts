import {useLayoutEffect, useRef} from "react";

const useHeight = (dependency?: unknown) => {
    const pageRef = useRef<HTMLDivElement | null>(null);
    useLayoutEffect(() => {
        const page = pageRef.current;
        if (!page) return;
        const body = page.closest<HTMLElement>(".ant-layout-body");
        let frame = 0;
        const update = () => {
            // offsetTop 不包含 Body 入场动画的 transform，也不受页面滚动影响。
            let top = 0;
            let element: HTMLElement | null = page;
            while (element) {
                top += element.offsetTop;
                element = element.offsetParent as HTMLElement | null;
                if (element) top += element.clientTop;
            }
            const bottom = body ? Number.parseFloat(getComputedStyle(body).paddingBottom) || 0 : 10;
            const offset = `${Math.ceil(top + bottom)}px`;
            if (page.style.getPropertyValue("--terminal-page-offset") !== offset) {
                page.style.setProperty("--terminal-page-offset", offset);
            }
        };
        const scheduleUpdate = () => {
            if (frame) return;
            frame = requestAnimationFrame(() => {
                frame = 0;
                update();
            });
        };
        update();
        const observer = new ResizeObserver(scheduleUpdate);
        observer.observe(body || page);
        window.addEventListener("resize", scheduleUpdate);
        return () => {
            cancelAnimationFrame(frame);
            observer.disconnect();
            window.removeEventListener("resize", scheduleUpdate);
        };
    }, [dependency]);

    return pageRef;
};

export default useHeight;
