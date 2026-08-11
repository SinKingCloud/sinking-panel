import type {CSSProperties} from "react";
import {useEffect, useRef, useState} from "react";

const useHeight = (dependency?: unknown) => {
    const pageRef = useRef<HTMLDivElement | null>(null);
    const [height, setHeight] = useState<number>();

    useEffect(() => {
        let frame = 0;
        const update = () => {
            cancelAnimationFrame(frame);
            frame = requestAnimationFrame(() => {
                const page = pageRef.current;
                if (!page) return;
                const top = page.getBoundingClientRect().top;
                const next = Math.max(320, Math.floor(window.innerHeight - top));
                setHeight((current) => current === next ? current : next);
            });
        };
        update();
        window.addEventListener("resize", update);
        window.addEventListener("orientationchange", update);
        return () => {
            cancelAnimationFrame(frame);
            window.removeEventListener("resize", update);
            window.removeEventListener("orientationchange", update);
        };
    }, [dependency]);

    const pageStyle = height
        ? {"--terminal-page-height": `${height}px`} as CSSProperties
        : undefined;

    return {pageRef, pageStyle};
};

export default useHeight;
