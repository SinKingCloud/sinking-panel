import React, {forwardRef, useCallback, useImperativeHandle, useRef, useState} from "react";
import {App, Button, Grid, Tooltip} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {getFileInfo} from "@/service/api/file";
import {formatFileMode, formatFileSize, formatFileTime, getFileIconType} from "../../utils";
import useStyles from "./styles";
import {copyTextToClipboard} from "./utils";

interface PropertiesState {
    path: string;
    record: any;
}

export interface FilePropertiesRef {
    open: (path: string, record: any) => void;
    close: () => void;
    updateRename: (sourcePath: string, targetPath: string, name: string) => void;
    updatePermissions: (path: string, permissions: string) => void;
}

interface FilePropertiesProps {
    onRename: (path: string, record: any) => void;
    onPermissions: (path: string, record: any) => void;
}

const FileProperties = forwardRef<FilePropertiesRef, FilePropertiesProps>(({onRename, onPermissions}, ref) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const {message} = App.useApp();
    const requestRef = useRef(0);
    const generationRef = useRef(0);
    const stateRef = useRef<PropertiesState | undefined>(undefined);
    const [state, setState] = useState<PropertiesState>();
    const [info, setInfo] = useState<any>();

    const close = useCallback(() => {
        generationRef.current += 1;
        requestRef.current += 1;
        stateRef.current = undefined;
        setState(undefined);
        setInfo(undefined);
    }, []);

    const updateRename = useCallback((sourcePath: string, targetPath: string, name: string) => {
        const currentState = stateRef.current;
        if (!currentState || currentState.path !== sourcePath) {
            return;
        }
        requestRef.current += 1;
        const nextState = {
            ...currentState,
            path: targetPath,
            record: {...currentState.record, name},
        };
        stateRef.current = nextState;
        setState(nextState);
        setInfo((currentInfo) => currentInfo ? {...currentInfo, name} : currentInfo);
    }, []);

    const updatePermissions = useCallback((path: string, permissions: string) => {
        const currentState = stateRef.current;
        const mode = Number.parseInt(permissions, 8);
        if (!currentState || currentState.path !== path || !Number.isFinite(mode)) {
            return;
        }
        requestRef.current += 1;
        const nextState = {
            ...currentState,
            record: {...currentState.record, mode},
        };
        stateRef.current = nextState;
        setState(nextState);
        setInfo((currentInfo) => currentInfo ? {...currentInfo, mode} : currentInfo);
    }, []);

    useImperativeHandle(ref, () => ({
        open: (path, record) => {
            const generation = ++generationRef.current;
            const requestId = ++requestRef.current;
            const nextState = {path, record};
            stateRef.current = nextState;
            setState(nextState);
            setInfo(undefined);
            void getFileInfo({body: {path}}).then((response) => {
                if (requestRef.current !== requestId || generationRef.current !== generation) {
                    return;
                }
                if (response?.code === 200 && response.data) {
                    setInfo(response.data);
                } else {
                    message.error(response?.message || "获取文件属性失败");
                }
            });
        },
        close,
        updateRename,
        updatePermissions,
    }), [close, message, updatePermissions, updateRename]);

    const current = info || state?.record;

    const copyValue = async (value: string, label: "名称" | "路径") => {
        if (await copyTextToClipboard(value)) {
            message.success(`${label}已复制`);
        } else {
            message.error(`复制${label}失败`);
        }
    };

    const openEditor = (type: "rename" | "permissions") => {
        if (!state || !current) {
            return;
        }
        const record = {
            ...state.record,
            name: current.name,
            size: Number(current.size) || 0,
            mode: current.mode,
            is_dir: current.is_dir,
        };
        if (type === "rename") {
            onRename(state.path, record);
        } else {
            onPermissions(state.path, record);
        }
    };

    return (
        <ProModal
            title={<Title>文件属性</Title>}
            width="480px"
            onCancel={close}
            modalProps={{
                open: Boolean(state),
                rootClassName: styles.root,
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                footer: null,
                focusable: {focusTriggerAfterClose: false},
                mask: {closable: true},
            }}>
            {current && (
                <section className="file-properties-panel">
                    <div className="file-properties-overview">
                        <span className={`file-properties-icon${current.is_dir ? " is-directory" : ""}`}>
                            <Icon type={getFileIconType(current.name, current.is_dir)}/>
                        </span>
                        <div className="file-properties-identity">
                            <div className="file-properties-name-row">
                                <div className="file-properties-name" title={current.name}>{current.name}</div>
                                <Tooltip title="复制名称">
                                    <Button
                                        color="default"
                                        variant="text"
                                        size="small"
                                        className="file-properties-action"
                                        aria-label="复制名称"
                                        icon={<Icon type="CopyOutlined"/>}
                                        onClick={() => void copyValue(current.name, "名称")}/>
                                </Tooltip>
                                <Tooltip title="重命名">
                                    <Button
                                        color="default"
                                        variant="text"
                                        size="small"
                                        className="file-properties-action"
                                        aria-label="重命名"
                                        icon={<Icon type="EditOutlined"/>}
                                        onClick={() => openEditor("rename")}/>
                                </Tooltip>
                            </div>
                        </div>
                    </div>
                    <div className="file-properties-details">
                        <div className="file-properties-item">
                            <div className="file-properties-label">类型</div>
                            <div className="file-properties-value">{current.is_dir ? "文件夹" : "文件"}</div>
                        </div>
                        {!current.is_dir && (
                            <div className="file-properties-item">
                                <div className="file-properties-label">大小</div>
                                <div className="file-properties-value">{formatFileSize(current.size)}</div>
                            </div>
                        )}
                        <div className="file-properties-item">
                            <div className="file-properties-label">权限</div>
                            <div className="file-properties-value-row">
                                <div className="file-properties-value is-permission">{formatFileMode(current.mode)}</div>
                                <Tooltip title="修改权限">
                                    <Button
                                        color="default"
                                        variant="text"
                                        size="small"
                                        className="file-properties-action"
                                        aria-label="修改权限"
                                        icon={<Icon type="EditOutlined"/>}
                                        onClick={() => openEditor("permissions")}/>
                                </Tooltip>
                            </div>
                        </div>
                        <div className={`file-properties-item${current.is_dir ? " is-wide" : ""}`}>
                            <div className="file-properties-label">修改时间</div>
                            <div className="file-properties-value is-time" title={String(current.update_time)}>
                                {formatFileTime(current.update_time)}
                            </div>
                        </div>
                        <div className="file-properties-item is-wide">
                            <div className="file-properties-label">完整路径</div>
                            <div className="file-properties-value-row">
                                <div className="file-properties-value file-properties-path" title={state?.path}>
                                    {state?.path}
                                </div>
                                <Tooltip title="复制完整路径">
                                    <Button
                                        color="default"
                                        variant="text"
                                        size="small"
                                        className="file-properties-action"
                                        aria-label="复制完整路径"
                                        icon={<Icon type="CopyOutlined"/>}
                                        onClick={() => void copyValue(state?.path || "", "路径")}/>
                                </Tooltip>
                            </div>
                        </div>
                    </div>
                </section>
            )}
        </ProModal>
    );
});

FileProperties.displayName = "FileProperties";

export default React.memo(FileProperties);
