import React, {useCallback, useEffect, useRef} from "react";
import {Button, Tooltip} from "antd";
import {Icon} from "sinking-antd";
import type {FileEditorTab} from "../hooks/editor-document";
import {getFileIconType} from "../utils";

interface FileEditorTabsProps {
    tabs: FileEditorTab[];
    activeKey?: string;
    tooltipsDisabled: boolean;
    getPopupContainer: () => HTMLElement;
    onActivate: (key: string) => void;
    onClose: (key: string) => void;
}

const FileEditorTabs = ({
    tabs,
    activeKey,
    tooltipsDisabled,
    getPopupContainer,
    onActivate,
    onClose,
}: FileEditorTabsProps) => {
    const activeTabRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        activeTabRef.current?.scrollIntoView({block: "nearest", inline: "nearest"});
    }, [activeKey]);

    const closeTab = useCallback((event: React.MouseEvent, key: string) => {
        event.stopPropagation();
        onClose(key);
    }, [onClose]);

    if (tabs.length === 0) {
        return <div className="file-editor-tabs-empty">未打开文件</div>;
    }

    return (
        <div className="file-editor-tabs" role="tablist" aria-label="已打开文件">
            {tabs.map((tab) => {
                const active = tab.key === activeKey;
                const stateClassName = tab.error
                    ? "has-error"
                    : tab.dirty ? "is-dirty" : "";
                return (
                    <div
                        key={tab.key}
                        ref={active ? activeTabRef : undefined}
                        className={`file-editor-tab ${active ? "is-active" : ""} ${stateClassName}`}
                        role="presentation">
                        <button
                            className="file-editor-tab-main"
                            type="button"
                            role="tab"
                            title={tooltipsDisabled ? undefined : tab.path}
                            aria-selected={active}
                            aria-controls="file-editor-canvas"
                            onClick={() => onActivate(tab.key)}
                            onAuxClick={(event) => {
                                if (event.button === 1 && !tab.saving) {
                                    event.preventDefault();
                                    onClose(tab.key);
                                }
                            }}>
                            {tab.loading || tab.saving ? (
                                <Icon className="anticon-spin" type="LoadingOutlined" aria-hidden/>
                            ) : (
                                <Icon type={getFileIconType(tab.name)} aria-hidden/>
                            )}
                            <span>{tab.name}</span>
                            {tab.dirty && <i aria-label="未保存"/>}
                        </button>
                        <Tooltip
                            title={tab.saving ? "正在保存" : "关闭"}
                            open={tooltipsDisabled ? false : undefined}
                            getPopupContainer={getPopupContainer}>
                            <Button
                                className="file-editor-tab-close"
                                type="text"
                                size="small"
                                tabIndex={active ? 0 : -1}
                                disabled={tab.saving}
                                aria-label={`关闭 ${tab.name}`}
                                icon={<Icon type="CloseOutlined"/>}
                                onClick={(event) => closeTab(event, tab.key)}/>
                        </Tooltip>
                    </div>
                );
            })}
        </div>
    );
};

export default React.memo(FileEditorTabs);
