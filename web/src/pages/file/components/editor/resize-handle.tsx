import React, {useCallback, useEffect, useLayoutEffect, useRef, useState} from "react";
import {readFileStorage, updateFileStorage} from "../../hooks/file-storage";

interface FileEditorResizeHandleProps {
    workspaceRef: React.RefObject<HTMLDivElement | null>;
    minWidth: number;
    disabled: boolean;
}

const FileEditorResizeHandle = ({workspaceRef, minWidth, disabled}: FileEditorResizeHandleProps) => {
    const handleRef = useRef<HTMLDivElement>(null);
    const [width, setWidth] = useState(() => {
        try {
            const stored = readFileStorage().editor?.treeWidth;
            return typeof stored === "number" && Number.isFinite(stored) ? Math.round(stored) : minWidth;
        } catch {
            return minWidth;
        }
    });
    const [drag, setDrag] = useState<{pointerId: number; x: number; width: number}>();
    const maxWidth = minWidth * 2;
    const currentWidth = Math.max(minWidth, Math.min(maxWidth, width));
    const widthRef = useRef(currentWidth);
    widthRef.current = currentWidth;
    const savedWidthRef = useRef<number | undefined>(undefined);
    const updateWidth = (value: number) => setWidth(Math.max(minWidth, Math.min(maxWidth, Math.round(value))));

    const saveWidth = useCallback(() => {
        const treeWidth = widthRef.current;
        if (savedWidthRef.current === treeWidth) return;
        try {
            updateFileStorage((current) => ({
                ...current,
                editor: {...current.editor, treeWidth},
            }));
            savedWidthRef.current = treeWidth;
        } catch {
            // Keep resizing usable when browser storage is unavailable.
        }
    }, []);

    useLayoutEffect(() => {
        // The parent ref may not be attached when this child first mounts.
        const workspace = handleRef.current?.parentElement || workspaceRef.current;
        workspace?.style.setProperty("--file-editor-tree-width", `${currentWidth}px`);
    }, [currentWidth, workspaceRef]);

    useEffect(() => {
        if (!drag) saveWidth();
    }, [currentWidth, drag, saveWidth]);

    useLayoutEffect(() => {
        const workspace = workspaceRef.current;
        workspace?.classList.toggle("tree-resizing", Boolean(drag) && !disabled);
        return () => workspace?.classList.remove("tree-resizing");
    }, [disabled, drag, workspaceRef]);

    useEffect(() => {
        setDrag(undefined);
    }, [disabled, minWidth]);

    useEffect(() => {
        const workspace = workspaceRef.current;
        const mobile = window.matchMedia("(max-width: 720px)");
        const stopDragging = () => setDrag(undefined);
        mobile.addEventListener("change", stopDragging);
        window.addEventListener("blur", stopDragging);
        return () => {
            saveWidth();
            mobile.removeEventListener("change", stopDragging);
            window.removeEventListener("blur", stopDragging);
            workspace?.style.removeProperty("--file-editor-tree-width");
        };
    }, [saveWidth, workspaceRef]);

    const resetPointerDrag = () => {
        setDrag(undefined);
        handleRef.current?.blur();
    };

    const endDrag = (event: React.PointerEvent<HTMLDivElement>) => {
        if (drag?.pointerId !== event.pointerId) return;
        if (event.currentTarget.hasPointerCapture(event.pointerId)) {
            event.currentTarget.releasePointerCapture(event.pointerId);
        }
        resetPointerDrag();
    };

    return (
        <div
            ref={handleRef}
            hidden={disabled}
            className={`file-editor-tree-resize-handle${drag ? " is-dragging" : ""}`}
            role="separator"
            tabIndex={0}
            aria-label="目录宽度"
            aria-orientation="vertical"
            aria-valuemin={minWidth}
            aria-valuemax={maxWidth}
            aria-valuenow={currentWidth}
            title="拖动调整目录宽度，双击恢复默认宽度"
            onPointerDown={(event) => {
                if (event.button !== 0 || !event.isPrimary || window.matchMedia("(max-width: 720px)").matches) return;
                event.preventDefault();
                event.currentTarget.setPointerCapture(event.pointerId);
                setDrag({pointerId: event.pointerId, x: event.clientX, width: currentWidth});
            }}
            onPointerMove={(event) => {
                if (drag?.pointerId === event.pointerId) {
                    updateWidth(drag.width + event.clientX - drag.x);
                }
            }}
            onPointerUp={endDrag}
            onPointerCancel={endDrag}
            onLostPointerCapture={resetPointerDrag}
            onDoubleClick={() => updateWidth(minWidth)}
            onKeyDown={(event) => {
                const next = event.key === "ArrowLeft" ? currentWidth - 10
                    : event.key === "ArrowRight" ? currentWidth + 10
                        : event.key === "Home" ? minWidth
                            : event.key === "End" ? maxWidth : undefined;
                if (next !== undefined) {
                    event.preventDefault();
                    updateWidth(next);
                }
            }}/>
    );
};

export default React.memo(FileEditorResizeHandle);
