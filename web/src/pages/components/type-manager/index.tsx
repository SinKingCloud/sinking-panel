import {
    forwardRef,
    memo,
    useCallback,
    useEffect,
    useImperativeHandle,
    useMemo,
    useRef,
    useState,
} from "react";
import {flushSync} from "react-dom";
import type {
    KeyboardEvent as ReactKeyboardEvent,
    PointerEvent as ReactPointerEvent,
} from "react";
import {App, Button, Dropdown, Empty, Form as AntForm, Input, Modal as AntModal, Spin} from "antd";
import {createStyles} from "antd-style";
import {Icon, ProModal, ProModalRef, Title, useTheme} from "sinking-antd";
import {
    createType,
    deleteType,
    getAllTypes,
    updateType,
} from "@/service/api/type";
import type {TypeModule, TypeRecord} from "@/service/api/type";
import {clearEnumCache} from "@/utils/enum";

export type {TypeModule, TypeRecord};

export interface TypeManagerRef {
    open: () => void;
    refresh: () => void;
}

export interface TypeManagerProps {
    module: TypeModule;
    onChange?: (items: TypeRecord[]) => void;
    onMutation?: () => void;
}

interface DraggableRowProps {
    actionDisabled: boolean;
    actionButtonClassName: string;
    actionsClassName: string;
    disabled: boolean;
    handleClassName: string;
    itemClassName: string;
    nameClassName: string;
    operating: boolean;
    record: TypeRecord;
    onEdit: (record: TypeRecord) => void;
    onRemove: (record: TypeRecord) => void;
    onPointerDown: (id: number, event: ReactPointerEvent<HTMLButtonElement>) => void;
    onCancelDrag: () => void;
    onKeyboardMove: (id: number, direction: -1 | 1) => void;
}

interface DragClasses {
    dragging: string;
    dropAfter: string;
    dropBefore: string;
    shifted: string;
}

interface DragRow {
    center: number;
    element: HTMLElement;
    height: number;
    id: number;
}

interface DragState {
    activeId: number;
    activeIndex: number;
    activated: boolean;
    classes: DragClasses;
    clientX: number;
    clientY: number;
    ended: boolean;
    frameId: number;
    handle: HTMLButtonElement;
    lastFrameTime: number;
    list: HTMLDivElement | null;
    listRect: DOMRect | null;
    originItems: TypeRecord[];
    overIndex: number;
    pointerId: number;
    rowGap: number;
    rows: DragRow[];
    startScrollTop: number;
    startX: number;
    startY: number;
}

interface TypeFormValues {
    name: string;
}

const moveItem = <T, >(items: T[], from: number, to: number): T[] => {
    const next = items.slice();
    const [item] = next.splice(from, 1);
    next.splice(to, 0, item);
    return next;
};

const dragAnimationDuration = 180;
const dragAnimationEasing = "cubic-bezier(.2, .8, .2, 1)";
const managedRowAnimations = new Map<HTMLElement, Animation>();

const rowElements = (list: HTMLElement): HTMLElement[] => (
    Array.from(list.querySelectorAll<HTMLElement>("[data-type-id]"))
);

const captureRowRects = (
    list: HTMLElement,
    ids?: ReadonlySet<number>,
): Map<number, DOMRect> => {
    const rects = new Map<number, DOMRect>();
    rowElements(list).forEach((element) => {
        const id = Number(element.dataset.typeId);
        if (Number.isFinite(id) && (!ids || ids.has(id))) {
            rects.set(id, element.getBoundingClientRect());
        }
    });
    return rects;
};

const cancelRowAnimations = (list: HTMLElement) => {
    Array.from(managedRowAnimations.entries()).forEach(([element, animation]) => {
        if (list.contains(element)) {
            animation.cancel();
            managedRowAnimations.delete(element);
        }
    });
};

const pauseRowTransitions = (list: HTMLElement, ids: ReadonlySet<number>): (() => void) => {
    const elements = rowElements(list).filter((element) => ids.has(Number(element.dataset.typeId)));
    elements.forEach((element) => element.style.setProperty("transition", "none"));
    return () => elements.forEach((element) => element.style.removeProperty("transition"));
};

const animateRowsFrom = (list: HTMLElement, firstRects: Map<number, DOMRect>) => {
    if (window.matchMedia?.("(prefers-reduced-motion: reduce)").matches) {
        return;
    }
    const elements = new Map<number, HTMLElement>();
    rowElements(list).forEach((element) => {
        const id = Number(element.dataset.typeId);
        if (firstRects.has(id)) {
            elements.set(id, element);
        }
    });
    const movements: Array<{deltaX: number; deltaY: number; element: HTMLElement}> = [];
    firstRects.forEach((first, id) => {
        const element = elements.get(id);
        if (!element) {
            return;
        }
        const last = element.getBoundingClientRect();
        const deltaX = first.left - last.left;
        const deltaY = first.top - last.top;
        if (Math.abs(deltaX) >= .5 || Math.abs(deltaY) >= .5) {
            movements.push({deltaX, deltaY, element});
        }
    });
    movements.forEach(({deltaX, deltaY, element}) => {
        managedRowAnimations.get(element)?.cancel();
        const animation = element.animate([
            {transform: `translate3d(${deltaX}px, ${deltaY}px, 0)`},
            {transform: "translate3d(0, 0, 0)"},
        ], {
            duration: dragAnimationDuration,
            easing: dragAnimationEasing,
        });
        managedRowAnimations.set(element, animation);
        const clear = () => {
            if (managedRowAnimations.get(element) === animation) {
                managedRowAnimations.delete(element);
            }
        };
        animation.addEventListener("finish", clear, {once: true});
        animation.addEventListener("cancel", clear, {once: true});
    });
};

const dragRangeIds = (state: DragState): Set<number> => {
    const targetIndex = state.overIndex < 0 ? state.activeIndex : state.overIndex;
    const start = Math.min(state.activeIndex, targetIndex);
    const end = Math.max(state.activeIndex, targetIndex);
    return new Set(state.rows.slice(start, end + 1).map((row) => row.id));
};

const moveRangeIds = (items: TypeRecord[], from: number, to: number): Set<number> => {
    if (from < 0 || to < 0) {
        return new Set();
    }
    const start = Math.min(from, to);
    const end = Math.max(from, to);
    return new Set(items.slice(start, end + 1).map((item) => item.id));
};

const changedOrderIds = (current: TypeRecord[], next: TypeRecord[]): Set<number> => {
    const ids = new Set<number>();
    const length = Math.max(current.length, next.length);
    for (let index = 0; index < length; index += 1) {
        const currentId = current[index]?.id;
        const nextId = next[index]?.id;
        if (currentId === nextId) {
            continue;
        }
        if (currentId !== undefined) {
            ids.add(currentId);
        }
        if (nextId !== undefined) {
            ids.add(nextId);
        }
    }
    return ids;
};

const closestRowIndex = (rows: DragRow[], center: number): number => {
    let left = 0;
    let right = rows.length - 1;
    while (left < right) {
        const middle = Math.floor((left + right) / 2);
        if (rows[middle].center < center) {
            left = middle + 1;
        } else {
            right = middle;
        }
    }
    if (left > 0) {
        const previous = left - 1;
        if (Math.abs(rows[previous].center - center) <= Math.abs(rows[left].center - center)) {
            return previous;
        }
    }
    return left;
};

const clearDragVisuals = (state: DragState) => {
    const targetIndex = state.overIndex < 0 ? state.activeIndex : state.overIndex;
    const start = Math.min(state.activeIndex, targetIndex);
    const end = Math.max(state.activeIndex, targetIndex);
    for (let index = start; index <= end; index += 1) {
        const element = state.rows[index]?.element;
        if (!element) {
            continue;
        }
        element.style.removeProperty("transform");
        element.classList.remove(
            state.classes.dragging,
            state.classes.shifted,
            state.classes.dropBefore,
            state.classes.dropAfter,
        );
    }
};

const releasePointerCapture = (state: DragState) => {
    const {handle, pointerId} = state;
    if (typeof handle.hasPointerCapture !== "function"
        || typeof handle.releasePointerCapture !== "function") {
        return;
    }
    try {
        if (handle.hasPointerCapture(pointerId)) {
            handle.releasePointerCapture(pointerId);
        }
    } catch {
        // Window-level pointer listeners keep drag cleanup reliable without capture.
    }
};

const updateDragPreview = (
    state: DragState,
    nextOverIndex: number,
) => {
    const activeRow = state.rows[state.activeIndex];
    const span = activeRow.height + state.rowGap;
    const previousOverIndex = state.overIndex;
    const previousTarget = previousOverIndex < 0 ? state.activeIndex : previousOverIndex;
    const nextTarget = nextOverIndex < 0 ? state.activeIndex : nextOverIndex;
    const start = Math.min(previousTarget, nextTarget);
    const end = Math.max(previousTarget, nextTarget);
    for (let index = start; index <= end; index += 1) {
        if (index === state.activeIndex) {
            continue;
        }
        const previousOffset = previousOverIndex > state.activeIndex
            && index > state.activeIndex && index <= previousOverIndex
            ? -span
            : previousOverIndex >= 0 && previousOverIndex < state.activeIndex
                && index >= previousOverIndex && index < state.activeIndex
                ? span
                : 0;
        const nextOffset = nextOverIndex > state.activeIndex
            && index > state.activeIndex && index <= nextOverIndex
            ? -span
            : nextOverIndex >= 0 && nextOverIndex < state.activeIndex
                && index >= nextOverIndex && index < state.activeIndex
                ? span
                : 0;
        if (previousOffset === nextOffset) {
            continue;
        }
        const element = state.rows[index].element;
        element.style.transform = nextOffset === 0 ? "" : `translate3d(0, ${nextOffset}px, 0)`;
        element.classList.toggle(state.classes.shifted, nextOffset !== 0);
    }
    if (previousOverIndex >= 0 && previousOverIndex !== state.activeIndex) {
        state.rows[previousOverIndex].element.classList.remove(
            state.classes.dropBefore,
            state.classes.dropAfter,
        );
    }
    if (nextOverIndex >= 0 && nextOverIndex !== state.activeIndex) {
        state.rows[nextOverIndex].element.classList.add(
            nextOverIndex < state.activeIndex ? state.classes.dropBefore : state.classes.dropAfter,
        );
    }
    state.overIndex = nextOverIndex;
};

const renderDragFrame = (
    state: DragState,
    timestamp: number,
) => {
    state.frameId = 0;
    if (state.ended || !state.activated || !state.list || !state.listRect) {
        return;
    }

    const elapsed = state.lastFrameTime === 0 ? 16 : Math.min(timestamp - state.lastFrameTime, 32);
    state.lastFrameTime = timestamp;
    const edge = 40;
    const maxSpeed = 520;
    const {list, listRect} = state;
    const withinScrollBounds = state.clientX >= listRect.left - 20
        && state.clientX <= listRect.right + 20
        && state.clientY >= listRect.top - 20
        && state.clientY <= listRect.bottom + 20;
    let speed = 0;
    if (withinScrollBounds && state.clientY < listRect.top + edge) {
        speed = -maxSpeed * Math.min(1, (listRect.top + edge - state.clientY) / edge);
    } else if (withinScrollBounds && state.clientY > listRect.bottom - edge) {
        speed = maxSpeed * Math.min(1, (state.clientY - listRect.bottom + edge) / edge);
    }
    const previousScrollTop = list.scrollTop;
    if (speed !== 0) {
        list.scrollTop += speed * elapsed / 1000;
    }
    const canContinueScrolling = list.scrollTop !== previousScrollTop;
    const activeOffset = state.clientY - state.startY + list.scrollTop - state.startScrollTop;
    state.rows[state.activeIndex].element.style.transform = `translate3d(0, ${activeOffset}px, 0)`;

    const outside = state.clientX < listRect.left - 20
        || state.clientX > listRect.right + 20
        || state.clientY < listRect.top - 20
        || state.clientY > listRect.bottom + 20;
    const nextOverIndex = outside
        ? -1
        : closestRowIndex(state.rows, state.rows[state.activeIndex].center + activeOffset);
    if (nextOverIndex !== state.overIndex) {
        updateDragPreview(state, nextOverIndex);
    }

    if (canContinueScrolling && !state.ended) {
        state.frameId = window.requestAnimationFrame((time) => renderDragFrame(state, time));
    }
};

const scheduleDragFrame = (state: DragState) => {
    if (state.frameId === 0 && !state.ended) {
        state.frameId = window.requestAnimationFrame((time) => renderDragFrame(state, time));
    }
};

const useStyles = createStyles(({css, token}: any, props: {compact: boolean}) => ({
    modal: css`
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
    `,
    content: css`
        display: flex;
        min-width: 0;
        flex-direction: column;
        gap: ${props.compact ? 6 : 8}px;
    `,
    listShell: css`
        display: flex;
        min-width: 0;
        min-height: 0;
        flex-direction: column;
        gap: ${props.compact ? 4 : 6}px;
        padding-top: ${props.compact ? 4 : 6}px;
    `,
    list: css`
        display: flex;
        min-width: 0;
        max-height: clamp(
            ${props.compact ? 120 : 132}px,
            calc(100dvh - ${props.compact ? 176 : 192}px),
            ${props.compact ? 300 : 340}px
        );
        flex-direction: column;
        gap: ${props.compact ? 3 : 4}px;
        overflow-x: hidden;
        overflow-y: auto;
        padding-block: 2px;
        scrollbar-width: thin;
    `,
    listLocked: css`
        [data-drag-handle],
        [data-row-action] {
            pointer-events: none;
            opacity: .45;
        }
    `,
    loading: css`
        display: flex;
        min-height: ${props.compact ? 156 : 176}px;
        align-items: center;
        justify-content: center;
    `,
    empty: css`
        display: flex;
        min-height: ${props.compact ? 156 : 176}px;
        align-items: center;
        justify-content: center;

        .ant-empty {
            margin-block: 0;
        }
    `,
    item: css`
        position: relative;
        display: grid;
        min-width: 0;
        min-height: ${props.compact ? 38 : 44}px;
        grid-template-columns: ${props.compact ? 32 : 36}px minmax(0, 1fr) auto;
        align-items: center;
        padding: ${props.compact ? "3px 6px 3px 4px" : "4px 8px 4px 5px"};
        border-radius: ${token.borderRadiusSM}px;
        background: ${token.colorFillQuaternary};
        transition:
            transform ${dragAnimationDuration}ms ${dragAnimationEasing},
            background-color 120ms ease,
            box-shadow 120ms ease;

        &:hover,
        &:focus-within {
            background: ${token.colorFillTertiary};
        }

        @media (pointer: coarse) {
            min-height: 44px;
        }
    `,
    dragging: css`
        z-index: 2;
        background: ${token.colorPrimaryBg};
        box-shadow: ${token.boxShadowTertiary};
        transition: background-color 120ms ease, box-shadow 120ms ease;
        will-change: transform;
    `,
    shifted: css`
        will-change: transform;
    `,
    dropBefore: css`
        &::before {
            position: absolute;
            z-index: 3;
            top: 0;
            right: 8px;
            left: 8px;
            height: 2px;
            border-radius: 1px;
            background: ${token.colorPrimary};
            content: "";
            pointer-events: none;
        }
    `,
    dropAfter: css`
        &::after {
            position: absolute;
            z-index: 3;
            right: 8px;
            bottom: 0;
            left: 8px;
            height: 2px;
            border-radius: 1px;
            background: ${token.colorPrimary};
            content: "";
            pointer-events: none;
        }
    `,
    handle: css`
        display: inline-flex;
        width: ${props.compact ? 28 : 32}px;
        height: ${props.compact ? 28 : 32}px;
        align-items: center;
        justify-content: center;
        padding: 0;
        border: 0;
        border-radius: ${token.borderRadiusSM}px;
        outline: none;
        background: transparent;
        color: ${token.colorTextQuaternary};
        cursor: grab;
        touch-action: none;
        user-select: none;

        &:hover,
        &:focus-visible {
            background: ${token.colorFillTertiary};
            color: ${token.colorTextSecondary};
        }

        &:active {
            cursor: grabbing;
        }

        &:disabled {
            background: transparent;
            color: ${token.colorTextQuaternary};
            cursor: default;
            opacity: .45;
        }

        @media (pointer: coarse) {
            width: 34px;
            height: 34px;
        }
    `,
    name: css`
        display: block;
        min-width: 0;
        overflow: hidden;
        color: ${token.colorText};
        font-size: ${props.compact ? 12 : 13}px;
        font-weight: 400;
        text-overflow: ellipsis;
        white-space: nowrap;
    `,
    actions: css`
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 2px;
    `,
    actionButton: css`
        width: ${props.compact ? 27 : 30}px;
        min-width: ${props.compact ? 27 : 30}px;
        height: ${props.compact ? 27 : 30}px;
        padding: 0;

        @media (pointer: coarse) {
            width: 34px;
            min-width: 34px;
            height: 34px;
        }
    `,
}));

const DraggableRow = memo(({
    actionDisabled,
    actionButtonClassName,
    actionsClassName,
    disabled,
    handleClassName,
    itemClassName,
    nameClassName,
    operating,
    record,
    onEdit,
    onRemove,
    onPointerDown,
    onCancelDrag,
    onKeyboardMove,
}: DraggableRowProps) => {
    return (
        <div
            data-type-id={String(record.id)}
            className={itemClassName}>
            <button
                type="button"
                className={handleClassName}
                aria-label="拖动分类"
                data-drag-handle
                disabled={disabled}
                onPointerDown={(event) => onPointerDown(record.id, event)}
                onKeyDown={(event: ReactKeyboardEvent<HTMLButtonElement>) => {
                    if (event.key === "Escape") {
                        onCancelDrag();
                        return;
                    }
                    if (event.key !== "ArrowUp" && event.key !== "ArrowDown") {
                        return;
                    }
                    event.preventDefault();
                    onKeyboardMove(record.id, event.key === "ArrowUp" ? -1 : 1);
                }}>
                <Icon type="HolderOutlined"/>
            </button>
            <span className={nameClassName} title={record.name}>{record.name}</span>
            <div className={actionsClassName} data-row-action>
                <Dropdown
                    trigger={["click"]}
                    placement="bottom"
                    arrow={{pointAtCenter: true}}
                    menu={{
                        items: [
                            {key: "edit", label: "编辑"},
                            {type: "divider"},
                            {key: "delete", label: "删除", danger: true},
                        ],
                        onClick: ({key}) => {
                            if (key === "edit") {
                                onEdit(record);
                            } else if (key === "delete") {
                                onRemove(record);
                            }
                        },
                    }}>
                    <Button
                        className={actionButtonClassName}
                        type="text"
                        aria-label={`管理分类 ${record.name}`}
                        icon={<Icon type="MoreOutlined"/>}
                        loading={operating}
                        disabled={actionDisabled}/>
                </Dropdown>
            </div>
        </div>
    );
});

DraggableRow.displayName = "DraggableRow";

const TypeManager = forwardRef<TypeManagerRef, TypeManagerProps>(({module, onChange, onMutation}, ref) => {
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const styleProps = useMemo(() => ({compact}), [compact]);
    const {styles} = useStyles(styleProps);
    const dragClassesRef = useRef<DragClasses>({
        dragging: styles.dragging,
        dropAfter: styles.dropAfter,
        dropBefore: styles.dropBefore,
        shifted: styles.shifted,
    });
    dragClassesRef.current = {
        dragging: styles.dragging,
        dropAfter: styles.dropAfter,
        dropBefore: styles.dropBefore,
        shifted: styles.shifted,
    };
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const [form] = AntForm.useForm<TypeFormValues>();
    const requestRef = useRef(0);
    const actionRef = useRef(false);
    const sortingRef = useRef(false);
    const sortRequestRef = useRef(0);
    const sortDirtyRef = useRef(false);
    const mountedRef = useRef(true);
    const dragRef = useRef<DragState | null>(null);
    const listRef = useRef<HTMLDivElement | null>(null);
    const [items, setItems] = useState<TypeRecord[]>([]);
    const itemsRef = useRef(items);
    itemsRef.current = items;
    const [loading, setLoading] = useState(false);
    const [formActive, setFormActive] = useState(false);
    const [formOpen, setFormOpen] = useState(false);
    const [formRecord, setFormRecord] = useState<TypeRecord>();
    const [formSubmitting, setFormSubmitting] = useState(false);
    const [sorting, setSorting] = useState(false);
    const [operatingId, setOperatingId] = useState<number | null>(null);
    const [draggingId, setDraggingId] = useState<number | null>(null);
    const loadingRef = useRef(loading);
    const formActiveRef = useRef(formActive);
    loadingRef.current = loading;
    formActiveRef.current = formActive;
    const editing = formRecord !== undefined;
    const operationBusy = formSubmitting || sorting || operatingId !== null;
    const busy = operationBusy || draggingId !== null || formActive;
    const dragDisabled = operationBusy || loading || formActive;
    const rowBusy = formSubmitting || operatingId !== null || formActive;
    const dragDisabledRef = useRef(dragDisabled);
    dragDisabledRef.current = dragDisabled;

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            requestRef.current += 1;
            sortRequestRef.current += 1;
        };
    }, []);

    const clearPointerDrag = useCallback(() => {
        const state = dragRef.current;
        dragRef.current = null;
        if (state?.frameId) {
            window.cancelAnimationFrame(state.frameId);
        }
        if (state) {
            state.ended = true;
            clearDragVisuals(state);
        }
        if (state) {
            releasePointerCapture(state);
        }
        setDraggingId(null);
    }, []);

    const applyItemsWithAnimation = useCallback((
        nextItems: TypeRecord[],
        firstRects?: Map<number, DOMRect>,
        affectedIds?: ReadonlySet<number>,
    ) => {
        if (!firstRects && affectedIds?.size === 0) {
            itemsRef.current = nextItems;
            setItems(nextItems);
            return;
        }
        const list = listRef.current;
        const before = firstRects ?? (list ? captureRowRects(list, affectedIds) : undefined);
        if (list) {
            cancelRowAnimations(list);
        }
        flushSync(() => {
            itemsRef.current = nextItems;
            setDraggingId(null);
            setItems(nextItems);
        });
        if (list && before) {
            animateRowsFrom(list, before);
        }
    }, []);

    useEffect(() => {
        const cancel = () => clearPointerDrag();
        window.addEventListener("blur", cancel);
        return () => {
            window.removeEventListener("blur", cancel);
            const state = dragRef.current;
            dragRef.current = null;
            if (state?.frameId) {
                window.cancelAnimationFrame(state.frameId);
            }
            if (state) {
                state.ended = true;
                clearDragVisuals(state);
            }
            if (state) {
                releasePointerCapture(state);
            }
            if (listRef.current) {
                cancelRowAnimations(listRef.current);
            }
        };
    }, [clearPointerDrag]);

    const load = useCallback(async (
        notify = false,
        animate = false,
    ): Promise<TypeRecord[] | undefined> => {
        const requestId = ++requestRef.current;
        setLoading(true);
        try {
            const response = await getAllTypes(module);
            if (requestRef.current !== requestId || !response) {
                return undefined;
            }
            if (response.code !== 200) {
                message.error(response.message || "获取分类失败");
                return undefined;
            }
            const nextItems = Array.isArray(response.data?.list) ? response.data.list : [];
            if (animate) {
                applyItemsWithAnimation(nextItems, undefined, changedOrderIds(itemsRef.current, nextItems));
            } else {
                itemsRef.current = nextItems;
                setItems(nextItems);
            }
            sortDirtyRef.current = false;
            if (notify) {
                onChange?.(nextItems);
            }
            return nextItems;
        } finally {
            if (requestRef.current === requestId) {
                setLoading(false);
            }
        }
    }, [applyItemsWithAnimation, message, module, onChange]);

    const refresh = useCallback(() => {
        if (actionRef.current || sortingRef.current) {
            return;
        }
        clearPointerDrag();
        setFormActive(false);
        setFormOpen(false);
        form.resetFields();
        setFormRecord(undefined);
        clearEnumCache(module);
        void load(true);
    }, [clearPointerDrag, form, load, module]);

    const reset = useCallback(() => {
        const reconcile = sortDirtyRef.current;
        requestRef.current += 1;
        sortRequestRef.current += 1;
        actionRef.current = false;
        sortingRef.current = false;
        clearPointerDrag();
        form.resetFields();
        setItems([]);
        setLoading(false);
        setFormActive(false);
        setFormOpen(false);
        setFormRecord(undefined);
        setFormSubmitting(false);
        setSorting(false);
        setOperatingId(null);
        if (reconcile && mountedRef.current) {
            clearEnumCache(module);
            void load(true);
        }
    }, [clearPointerDrag, form, load, module]);

    const open = useCallback(() => {
        requestRef.current += 1;
        sortRequestRef.current += 1;
        actionRef.current = false;
        sortingRef.current = false;
        form.resetFields();
        setFormActive(false);
        setFormOpen(false);
        setFormRecord(undefined);
        modalRef.current?.show();
        clearEnumCache(module);
        void load(true);
    }, [form, load, module]);

    useImperativeHandle(ref, () => ({open, refresh}), [open, refresh]);

    const reloadAfterChange = useCallback(async () => {
        clearEnumCache(module);
        await load(true);
    }, [load, module]);

    const openCreateForm = useCallback(() => {
        clearPointerDrag();
        setFormRecord(undefined);
        form.resetFields();
        form.setFieldsValue({name: ""});
        setFormActive(true);
        setFormOpen(true);
    }, [clearPointerDrag, form]);

    const openEditForm = useCallback((record: TypeRecord) => {
        if (
            actionRef.current
            || sortingRef.current
            || dragRef.current
            || loadingRef.current
            || formActiveRef.current
        ) {
            return;
        }
        clearPointerDrag();
        setFormRecord(record);
        form.resetFields();
        form.setFieldsValue({name: record.name});
        setFormActive(true);
        setFormOpen(true);
    }, [clearPointerDrag, form]);

    const closeForm = useCallback(() => {
        setFormOpen(false);
    }, []);

    const resetForm = useCallback(() => {
        form.resetFields();
        setFormRecord(undefined);
        setFormActive(false);
    }, [form]);

    const submitForm = useCallback(async (values: TypeFormValues) => {
        if (actionRef.current) {
            return;
        }
        const name = values.name.trim();
        actionRef.current = true;
        setFormSubmitting(true);
        try {
            const response = await (formRecord
                ? updateType({body: {ids: [formRecord.id], name}})
                : createType({body: {module, name}}));
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || (editing ? "修改分类失败" : "添加分类失败"));
                return;
            }
            setFormOpen(false);
            message.success(response.message || (editing ? "修改成功" : "添加成功"));
            onMutation?.();
            await reloadAfterChange();
        } finally {
            actionRef.current = false;
            setFormSubmitting(false);
        }
    }, [editing, formRecord, message, module, onMutation, reloadAfterChange]);

    const remove = useCallback((record: TypeRecord) => {
        if (
            actionRef.current
            || sortingRef.current
            || dragRef.current
            || loadingRef.current
            || formActiveRef.current
        ) {
            return;
        }
        modal.confirm({
            title: "删除分类",
            content: `删除“${record.name}”后，关联项将恢复为全部分类。`,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true, type: "default"},
            mask: {closable: true},
            onOk: async () => {
                if (actionRef.current) {
                    return;
                }
                actionRef.current = true;
                setOperatingId(record.id);
                try {
                    const response = await deleteType({body: {ids: [record.id]}});
                    if (!response) {
                        return;
                    }
                    if (response.code !== 200) {
                        message.error(response.message || "删除分类失败");
                        return;
                    }
                    message.success(response.message || "删除成功");
                    onMutation?.();
                    await reloadAfterChange();
                } finally {
                    actionRef.current = false;
                    setOperatingId(null);
                }
            },
        } as any);
    }, [message, modal, onMutation, reloadAfterChange]);

    const persistMove = useCallback(async (
        originItems: TypeRecord[],
        nextItems: TypeRecord[],
        activeId: number,
    ) => {
        const oldIndex = originItems.findIndex((item) => item.id === activeId);
        const newIndex = nextItems.findIndex((item) => item.id === activeId);
        const affectedIds = moveRangeIds(originItems, oldIndex, newIndex);
        const rollback = () => applyItemsWithAnimation(originItems, undefined, affectedIds);
        if (actionRef.current || sortingRef.current) {
            rollback();
            return;
        }
        if (oldIndex < 0 || newIndex < 0) {
            rollback();
            return;
        }
        if (oldIndex === newIndex) {
            return;
        }
        const movedItem = originItems[oldIndex];
        const targetItem = originItems[newIndex];
        const nextSort = Number(targetItem.sort) + (newIndex < oldIndex ? -1 : 1);
        const optimisticItems = nextItems.map((item) => (
            item.id === movedItem.id ? {...item, sort: nextSort} : item
        ));
        itemsRef.current = optimisticItems;
        setItems(optimisticItems);
        const sortRequestId = ++sortRequestRef.current;
        sortingRef.current = true;
        setSorting(true);
        let committed = false;
        try {
            const response = await updateType({
                body: {
                    ids: [movedItem.id],
                    sort: String(nextSort),
                },
            });
            committed = response?.code === 200;
            if (committed) {
                sortDirtyRef.current = true;
            }
            if (sortRequestRef.current !== sortRequestId) {
                return;
            }
            if (!response || response.code !== 200) {
                rollback();
                if (response) {
                    message.error(response.message || "排序失败");
                }
                clearEnumCache(module);
                await load(true, true);
                return;
            }
            clearEnumCache(module);
            await load(true, true);
        } finally {
            if (sortRequestRef.current === sortRequestId) {
                sortingRef.current = false;
                setSorting(false);
                if (sortDirtyRef.current) {
                    clearEnumCache(module);
                    if (mountedRef.current) {
                        await load(true, true);
                    }
                }
            } else if (committed) {
                sortDirtyRef.current = true;
                clearEnumCache(module);
                if (sortingRef.current) {
                    sortDirtyRef.current = true;
                } else if (mountedRef.current) {
                    await load(true, true);
                }
            }
        }
    }, [applyItemsWithAnimation, load, message, module]);

    const startPointerDrag = useCallback((
        id: number,
        event: ReactPointerEvent<HTMLButtonElement>,
    ) => {
        if (
            dragRef.current ||
            !event.isPrimary ||
            dragDisabledRef.current ||
            actionRef.current ||
            sortingRef.current ||
            (event.pointerType === "mouse" && event.button !== 0)
        ) {
            return;
        }
        if (listRef.current) {
            cancelRowAnimations(listRef.current);
        }
        event.preventDefault();
        const handle = event.currentTarget;
        const snapshot = itemsRef.current.slice();
        dragRef.current = {
            activeId: id,
            activeIndex: -1,
            activated: false,
            classes: {...dragClassesRef.current},
            clientX: event.clientX,
            clientY: event.clientY,
            ended: false,
            frameId: 0,
            handle,
            lastFrameTime: 0,
            list: null,
            listRect: null,
            originItems: snapshot,
            overIndex: -1,
            pointerId: event.pointerId,
            rowGap: 0,
            rows: [],
            startScrollTop: 0,
            startX: event.clientX,
            startY: event.clientY,
        };
        try {
            handle.setPointerCapture(event.pointerId);
        } catch {
            // Window-level pointer listeners keep the drag active when capture is unavailable.
        }
    }, []);

    const movePointerDrag = useCallback((event: PointerEvent) => {
        const state = dragRef.current;
        if (!state || state.pointerId !== event.pointerId) {
            return;
        }
        state.clientX = event.clientX;
        state.clientY = event.clientY;
        if (!state.activated) {
            const distance = Math.hypot(event.clientX - state.startX, event.clientY - state.startY);
            if (distance < 6) {
                return;
            }
            const list = listRef.current;
            if (!list) {
                clearPointerDrag();
                return;
            }
            const listRect = list.getBoundingClientRect();
            const elements = rowElements(list);
            const activeIndex = elements.findIndex((element) => Number(element.dataset.typeId) === state.activeId);
            if (activeIndex < 0) {
                clearPointerDrag();
                return;
            }
            const rowGap = Number.parseFloat(window.getComputedStyle(list).rowGap) || 0;
            const activeRect = elements[activeIndex].getBoundingClientRect();
            const activeTop = activeRect.top - listRect.top + list.scrollTop;
            const span = activeRect.height + rowGap;
            const rows = elements.map((element, index): DragRow => ({
                center: activeTop + (index - activeIndex) * span + activeRect.height / 2,
                element,
                height: activeRect.height,
                id: Number(element.dataset.typeId),
            })).filter((row) => Number.isFinite(row.id));
            state.activated = true;
            state.activeIndex = activeIndex;
            state.list = list;
            state.listRect = listRect;
            state.overIndex = activeIndex;
            state.rowGap = rowGap;
            state.rows = rows;
            state.startScrollTop = list.scrollTop;
            rows[activeIndex].element.classList.add(state.classes.dragging);
            setDraggingId(state.activeId);
        }
        event.preventDefault();
        scheduleDragFrame(state);
    }, [clearPointerDrag]);

    const cancelActiveDrag = useCallback(() => {
        const state = dragRef.current;
        if (!state) {
            return;
        }
        const affectedIds = dragRangeIds(state);
        const firstRects = state.activated && state.list
            ? captureRowRects(state.list, affectedIds)
            : undefined;
        const resumeTransitions = state.activated && state.list
            ? pauseRowTransitions(state.list, affectedIds)
            : undefined;
        dragRef.current = null;
        state.ended = true;
        if (state.frameId) {
            window.cancelAnimationFrame(state.frameId);
        }
        releasePointerCapture(state);
        clearDragVisuals(state);
        if (firstRects) {
            applyItemsWithAnimation(state.originItems, firstRects);
        } else {
            setDraggingId(null);
        }
        resumeTransitions?.();
    }, [applyItemsWithAnimation]);

    const cancelPointerDrag = useCallback((event: PointerEvent) => {
        const state = dragRef.current;
        if (!state || state.pointerId !== event.pointerId) {
            return;
        }
        cancelActiveDrag();
    }, [cancelActiveDrag]);

    const finishPointerDrag = useCallback((event: PointerEvent) => {
        const state = dragRef.current;
        if (!state || state.pointerId !== event.pointerId) {
            return;
        }
        state.clientX = event.clientX;
        state.clientY = event.clientY;
        if (state.frameId) {
            window.cancelAnimationFrame(state.frameId);
            state.frameId = 0;
        }
        if (state.activated) {
            renderDragFrame(state, performance.now());
            if (state.frameId) {
                window.cancelAnimationFrame(state.frameId);
                state.frameId = 0;
            }
        }
        const affectedIds = dragRangeIds(state);
        const firstRects = state.activated && state.list
            ? captureRowRects(state.list, affectedIds)
            : undefined;
        const resumeTransitions = state.activated && state.list
            ? pauseRowTransitions(state.list, affectedIds)
            : undefined;
        const oldIndex = state.activeIndex;
        const newIndex = state.overIndex;
        dragRef.current = null;
        state.ended = true;
        releasePointerCapture(state);
        clearDragVisuals(state);
        if (!state.activated || oldIndex < 0 || newIndex < 0 || oldIndex === newIndex) {
            if (firstRects) {
                applyItemsWithAnimation(state.originItems, firstRects);
            } else {
                setDraggingId(null);
            }
            resumeTransitions?.();
            return;
        }
        const nextItems = moveItem(state.originItems, oldIndex, newIndex);
        applyItemsWithAnimation(nextItems, firstRects);
        resumeTransitions?.();
        void persistMove(state.originItems, nextItems, state.activeId);
    }, [applyItemsWithAnimation, persistMove]);

    useEffect(() => {
        window.addEventListener("pointermove", movePointerDrag, {passive: false});
        window.addEventListener("pointerup", finishPointerDrag);
        window.addEventListener("pointercancel", cancelPointerDrag);
        return () => {
            window.removeEventListener("pointermove", movePointerDrag);
            window.removeEventListener("pointerup", finishPointerDrag);
            window.removeEventListener("pointercancel", cancelPointerDrag);
        };
    }, [cancelPointerDrag, finishPointerDrag, movePointerDrag]);

    const moveByKeyboard = useCallback((id: number, direction: -1 | 1) => {
        if (dragDisabledRef.current || actionRef.current || sortingRef.current) {
            return;
        }
        const currentItems = itemsRef.current;
        const oldIndex = currentItems.findIndex((item) => item.id === id);
        const newIndex = oldIndex + direction;
        if (oldIndex < 0 || newIndex < 0 || newIndex >= currentItems.length) {
            return;
        }
        const nextItems = moveItem(currentItems, oldIndex, newIndex);
        const affectedIds = moveRangeIds(currentItems, oldIndex, newIndex);
        applyItemsWithAnimation(nextItems, undefined, affectedIds);
        void persistMove(currentItems, nextItems, id);
    }, [applyItemsWithAnimation, persistMove]);

    return (
        <ProModal
            ref={modalRef}
            title={<Title>分类管理</Title>}
            onOk={openCreateForm}
            width={480}
            modalProps={{
                rootClassName: styles.modal,
                okText: "添加",
                cancelText: "取消",
                closable: !busy,
                keyboard: !busy,
                forceRender: true,
                mask: {closable: !busy},
                afterClose: reset,
            } as any}>
            <>
                <div className={styles.content}>
                    <div className={styles.listShell}>
                        <div
                            ref={listRef}
                            className={[styles.list, dragDisabled ? styles.listLocked : ""]
                                .filter(Boolean).join(" ")}>
                            {loading && items.length === 0 ? (
                                <div className={styles.loading}><Spin size="small"/></div>
                            ) : items.length === 0 ? (
                                <div className={styles.empty}>
                                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无分类"/>
                                </div>
                            ) : (
                                <>
                                    {items.map((record) => (
                                        <DraggableRow
                                            key={record.id}
                                            record={record}
                                            disabled={rowBusy}
                                            actionDisabled={rowBusy}
                                            actionButtonClassName={styles.actionButton}
                                            actionsClassName={styles.actions}
                                            handleClassName={styles.handle}
                                            itemClassName={styles.item}
                                            nameClassName={styles.name}
                                            operating={operatingId === record.id}
                                            onEdit={openEditForm}
                                            onRemove={remove}
                                            onPointerDown={startPointerDrag}
                                            onCancelDrag={cancelActiveDrag}
                                            onKeyboardMove={moveByKeyboard}/>
                                    ))}
                                </>
                            )}
                        </div>
                    </div>
                </div>

                <AntModal
                    rootClassName={styles.modal}
                    title={<Title>{editing ? "编辑分类" : "添加分类"}</Title>}
                    width={360}
                    open={formOpen}
                    onOk={form.submit}
                    onCancel={closeForm}
                    okText={editing ? "保存" : "添加"}
                    cancelText="取消"
                    confirmLoading={formSubmitting}
                    closable={!formSubmitting}
                    keyboard={!formSubmitting}
                    mask={{closable: !formSubmitting}}
                    forceRender
                    focusable={{focusTriggerAfterClose: false}}
                    afterClose={resetForm}>
                    <AntForm<TypeFormValues>
                        form={form}
                        layout="vertical"
                        onFinish={submitForm}>
                        <AntForm.Item
                            name="name"
                            label="分类名称"
                            rules={[
                                {required: true, whitespace: true, message: "请输入分类名称"},
                                {max: 50, message: "分类名称不能超过 50 个字符"},
                            ]}>
                            <Input
                                maxLength={50}
                                placeholder="请输入分类名称"
                                autoComplete="off"
                                onPressEnter={form.submit}/>
                        </AntForm.Item>
                    </AntForm>
                </AntModal>
            </>
        </ProModal>
    );
});

TypeManager.displayName = "TypeManager";

export default memo(TypeManager);
