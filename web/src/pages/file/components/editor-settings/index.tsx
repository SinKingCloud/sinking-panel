import React, {useCallback, useEffect, useId, useLayoutEffect, useRef, useState} from "react";
import {Button, InputNumber, Segmented, Select, Switch} from "antd";
import {Icon} from "sinking-antd";

export type FileEditorTheme =
    | "auto"
    | "github"
    | "github_dark"
    | "one_dark"
    | "dracula"
    | "monokai"
    | "tomorrow"
    | "tomorrow_night"
    | "solarized_light"
    | "solarized_dark";

export interface FileEditorPreferences {
    theme: FileEditorTheme;
    fontSize: number;
    tabSize: number;
    wrapEnabled: boolean;
    showLineNumbers: boolean;
}

interface FileEditorSettingsProps {
    value: FileEditorPreferences;
    popupClassName?: string;
    onChange: (value: FileEditorPreferences) => void;
}

const themeOptions: Array<{label: string; value: FileEditorTheme}> = [
    {label: "跟随界面", value: "auto"},
    {label: "GitHub Light", value: "github"},
    {label: "GitHub Dark", value: "github_dark"},
    {label: "One Dark", value: "one_dark"},
    {label: "Dracula", value: "dracula"},
    {label: "Monokai", value: "monokai"},
    {label: "Tomorrow", value: "tomorrow"},
    {label: "Tomorrow Night", value: "tomorrow_night"},
    {label: "Solarized Light", value: "solarized_light"},
    {label: "Solarized Dark", value: "solarized_dark"},
];

const FileEditorSettings = ({
    value,
    popupClassName,
    onChange,
}: FileEditorSettingsProps) => {
    const containerRef = useRef<HTMLDivElement | null>(null);
    const panelRef = useRef<HTMLDivElement | null>(null);
    const buttonRef = useRef<HTMLButtonElement | null>(null);
    const panelId = useId();
    const [open, setOpen] = useState(false);
    const [rendered, setRendered] = useState(false);
    const update = (nextValue: Partial<FileEditorPreferences>) => onChange({...value, ...nextValue});
    const closePanel = useCallback(() => setOpen(false), []);

    const togglePanel = useCallback(() => {
        if (open) {
            closePanel();
            return;
        }
        setRendered(true);
        setOpen(true);
    }, [closePanel, open]);

    const alignArrow = useCallback(() => {
        const panel = panelRef.current;
        const button = buttonRef.current;
        if (!panel || !button) {
            return;
        }
        const panelRect = panel.getBoundingClientRect();
        const buttonRect = button.getBoundingClientRect();
        const buttonCenter = buttonRect.left + buttonRect.width / 2;
        const arrowOffset = Math.min(
            panelRect.width - 14,
            Math.max(14, buttonCenter - panelRect.left),
        );
        panel.style.setProperty("--file-editor-settings-arrow-x", `${arrowOffset}px`);
    }, []);

    useLayoutEffect(() => {
        if (!rendered) {
            return;
        }
        alignArrow();
        window.addEventListener("resize", alignArrow);
        return () => window.removeEventListener("resize", alignArrow);
    }, [alignArrow, rendered]);

    useLayoutEffect(() => {
        if (panelRef.current) {
            panelRef.current.inert = !open;
        }
    }, [open, rendered]);

    useEffect(() => {
        if (!rendered || open) {
            return;
        }
        const timer = window.setTimeout(() => setRendered(false), 200);
        return () => window.clearTimeout(timer);
    }, [open, rendered]);

    useEffect(() => {
        if (!open) {
            return;
        }
        const closeOnPointerDown = (event: PointerEvent) => {
            if (!containerRef.current?.contains(event.target as Node)) {
                closePanel();
            }
        };
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key !== "Escape" || event.isComposing || event.keyCode === 229) {
                return;
            }
            event.preventDefault();
            event.stopPropagation();
            closePanel();
            window.requestAnimationFrame(() => buttonRef.current?.focus());
        };
        document.addEventListener("pointerdown", closeOnPointerDown, true);
        document.addEventListener("keydown", closeOnEscape, true);
        return () => {
            document.removeEventListener("pointerdown", closeOnPointerDown, true);
            document.removeEventListener("keydown", closeOnEscape, true);
        };
    }, [closePanel, open]);

    const content = (
        <div className="file-editor-settings" aria-label="编辑器设置">
            <div className="file-editor-setting-row">
                <span>主题</span>
                <Select
                    value={value.theme}
                    options={themeOptions}
                    aria-label="编辑器主题"
                    getPopupContainer={(triggerNode) => (
                        panelRef.current || triggerNode.parentElement || triggerNode
                    )}
                    onChange={(nextValue) => update({theme: nextValue as FileEditorTheme})}/>
            </div>
            <div className="file-editor-setting-row">
                <span>字号</span>
                <InputNumber
                    min={12}
                    max={24}
                    precision={0}
                    value={value.fontSize}
                    aria-label="编辑器字号"
                    onChange={(nextValue) => update({fontSize: Number(nextValue) || 12})}/>
            </div>
            <div className="file-editor-setting-row">
                <span>Tab 宽度</span>
                <Segmented
                    size="small"
                    value={value.tabSize}
                    options={[2, 4, 8]}
                    aria-label="Tab 宽度"
                    onChange={(nextValue) => update({tabSize: Number(nextValue)})}/>
            </div>
            <div className="file-editor-setting-row">
                <span>自动换行</span>
                <Switch
                    size="small"
                    checked={value.wrapEnabled}
                    aria-label="自动换行"
                    onChange={(checked) => update({wrapEnabled: checked})}/>
            </div>
            <div className="file-editor-setting-row">
                <span>显示行号</span>
                <Switch
                    size="small"
                    checked={value.showLineNumbers}
                    aria-label="显示行号"
                    onChange={(checked) => update({showLineNumbers: checked})}/>
            </div>
        </div>
    );

    return (
        <div ref={containerRef} className="file-editor-settings-trigger">
            <Button
                ref={buttonRef}
                type="text"
                aria-label="编辑器设置"
                aria-expanded={open}
                aria-controls={panelId}
                icon={<Icon type="SettingOutlined"/>}
                onClick={togglePanel}/>
            {rendered && (
                <div
                    ref={panelRef}
                    id={panelId}
                    className={`file-editor-settings-panel ${open ? "is-open" : ""} ${popupClassName || ""}`}
                    role="region"
                    aria-hidden={!open}
                    onTransitionEnd={(event) => {
                        if (!open && event.target === event.currentTarget && event.propertyName === "opacity") {
                            setRendered(false);
                        }
                    }}
                    aria-label="编辑器设置">
                    {content}
                </div>
            )}
        </div>
    );
};

export default React.memo(FileEditorSettings);
