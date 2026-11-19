import {useEffect, useState} from "react";
import {App, Button, Grid, Spin, Tooltip} from "antd";
import {ossOperate} from "@/service/api/oss";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {formatFileSize, formatFileTime, getFileIconType} from "../../utils";
import usePropertiesStyles from "../properties/styles";
import {copyTextToClipboard} from "../properties/utils";
import useStyles from "./styles";

interface FileInfoProps {
    id: number;
    item?: {
        key: string;
        directory?: boolean;
        size?: number;
        content_type?: string;
        last_modified?: string;
        etag?: string;
        storage_class?: string;
        metadata?: Record<string, string>;
    };
    active: boolean;
    zIndex?: number;
    onCancel: () => void;
}

// FileInfo 沿用本地文件属性布局，路径与复制内容保留云端原始值。
const FileInfo = ({id, item, active, zIndex = 1030, onCancel}: FileInfoProps) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles: propertiesStyles} = usePropertiesStyles({compact});
    const {styles} = useStyles({compact});
    const [detail, setDetail] = useState<FileInfoProps["item"]>();
    const key = item?.key;
    const directory = Boolean(item?.directory);
    useEffect(() => {
        setDetail(undefined);
        if (!active || !key || directory) return;
        let current = true;
        void ossOperate({body: {id, action: "info", key}}).then((response) => {
            if (!current) return;
            if (response?.code !== 200) throw new Error(response?.message || "获取文件信息失败");
            if (response.data?.key !== key) throw new Error("获取文件信息失败");
            setDetail(response.data);
        }).catch((err) => {
            if (current) { onCancel(); message.error(err?.message || "获取文件信息失败"); }
        });
        return () => { current = false; };
    }, [active, id, key, directory, message, onCancel]);
    const info = directory ? item : detail?.key === key ? detail : undefined;
    const name = key?.replace(/\/$/, "").split("/").pop() || "";
    const metadata = Object.entries(info?.metadata || {});

    const copy = async (value: string, label: string) => {
        if (await copyTextToClipboard(value)) message.success(`${label}已复制`);
        else message.error(`复制${label}失败`);
    };

    const copyButton = (value: string, label: string) => (
        <Tooltip title={`复制${label}`}>
            <Button color="default" variant="text" size="small" className="file-properties-action"
                aria-label={`复制${label}`} icon={<Icon type="CopyOutlined"/>}
                onClick={() => void copy(value, label)}/>
        </Tooltip>
    );

    return <ProModal title={<Title>{directory ? "目录信息" : "文件信息"}</Title>} width="480px" onCancel={onCancel}
        modalProps={{open: active && Boolean(key), zIndex, footer: null, mask: {closable: true},
            rootClassName: `${propertiesStyles.root} ${styles.info}`,
            style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24}}}>
        {!info ? <div style={{height: 240, display: "flex", alignItems: "center", justifyContent: "center"}}
            role="status" aria-label="正在读取文件信息"><Spin/></div> : <section className="file-properties-panel">
            <div className="file-properties-overview">
                <span className="file-properties-icon"><Icon type={getFileIconType(name, directory)}/></span>
                <div className="file-properties-identity">
                    <div className="file-properties-name-row">
                        <div className="file-properties-name" title={name}>{name}</div>
                        {copyButton(name, "名称")}
                    </div>
                </div>
            </div>
            <div className="file-properties-details">
                {!directory && [
                    {label: "大小", value: formatFileSize(info.size ?? 0)},
                    {label: "修改时间", value: info.last_modified && !info.last_modified.startsWith("0001-") ? formatFileTime(info.last_modified) : "—"},
                    {label: "内容类型", value: info.content_type || "—"},
                    {label: "存储类型", value: info.storage_class || "—"},
                ].map(({label, value}) => <div className="file-properties-item" key={label}>
                    <div className="file-properties-label">{label}</div>
                    <div className="file-properties-value">{value}</div>
                </div>)}
                {[
                    {label: directory ? "目录路径" : "文件路径", value: info.key},
                    ...(!directory ? [{label: "ETag", value: info.etag}] : []),
                ].map(({label, value}) => <div className="file-properties-item is-wide" key={label}>
                    <div className="file-properties-label">{label}</div>
                    <div className="file-properties-value-row">
                        <div className="file-properties-value file-properties-path">{value || "—"}</div>
                        {value && copyButton(value, label)}
                    </div>
                </div>)}
                {!directory && <div className="file-properties-item is-wide">
                    <div className="file-properties-label">元数据</div>
                    {metadata.length ? <dl className="file-info-metadata">
                        {metadata.map(([key, value]) => <div key={key}>
                            <dt>{key}</dt><dd>{value || "—"}</dd>
                        </div>)}
                    </dl> : <div className="file-info-empty">暂无元数据</div>}
                </div>}
            </div>
        </section>}
    </ProModal>;
};

export default FileInfo;
