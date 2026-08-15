import React from "react";
import {Button, InputNumber, Popover, Segmented, Select, Switch} from "antd";
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
    getPopupContainer: () => HTMLElement;
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
    getPopupContainer,
    popupClassName,
    onChange,
}: FileEditorSettingsProps) => {
    const update = (nextValue: Partial<FileEditorPreferences>) => onChange({...value, ...nextValue});

    const content = (
        <div className="file-editor-settings" aria-label="编辑器设置">
            <div className="file-editor-setting-row">
                <span>主题</span>
                <Select
                    value={value.theme}
                    options={themeOptions}
                    aria-label="编辑器主题"
                    getPopupContainer={getPopupContainer}
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
        <Popover
            trigger="click"
            placement="bottomRight"
            content={content}
            rootClassName={popupClassName}
            getPopupContainer={getPopupContainer}>
            <Button
                type="text"
                aria-label="编辑器设置"
                icon={<Icon type="SettingOutlined"/>}/>
        </Popover>
    );
};

export default React.memo(FileEditorSettings);
