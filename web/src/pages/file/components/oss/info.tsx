import {App, Button, Grid, Spin, Tooltip} from "antd";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {formatFileSize, formatFileTime, getFileIconType} from "../../utils";
import usePropertiesStyles from "../properties/styles";
import {copyTextToClipboard} from "../properties/utils";
import useStyles from "./styles";

interface FileInfoProps {
    info?: {
        key: string;
        size: number;
        content_type?: string;
        last_modified?: string;
        etag?: string;
        storage_class?: string;
        metadata?: Record<string, string>;
    };
    loading: boolean;
    active: boolean;
    onCancel: () => void;
}

// FileInfo 沿用本地文件属性布局，路径与复制内容保留云端原始值。
const FileInfo = ({info, loading, active, onCancel}: FileInfoProps) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles: propertiesStyles} = usePropertiesStyles({compact});
    const {styles} = useStyles({compact});
    const name = info?.key.slice(info.key.lastIndexOf("/") + 1) || "";
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

    return <ProModal title={<Title>文件信息</Title>} width="480px" onCancel={onCancel}
        modalProps={{open: active && Boolean(info), zIndex: 1030, footer: null, mask: {closable: true},
            rootClassName: `${propertiesStyles.root} ${styles.info}`,
            style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24}}}>
        {loading ? <div style={{height: 240, display: "flex", alignItems: "center", justifyContent: "center"}}
            role="status" aria-label="正在读取文件信息"><Spin/></div> : info && <section className="file-properties-panel">
            <div className="file-properties-overview">
                <span className="file-properties-icon"><Icon type={getFileIconType(name)}/></span>
                <div className="file-properties-identity">
                    <div className="file-properties-name-row">
                        <div className="file-properties-name" title={name}>{name}</div>
                        {copyButton(name, "名称")}
                    </div>
                </div>
            </div>
            <div className="file-properties-details">
                {[
                    {label: "大小", value: formatFileSize(info.size)},
                    {label: "修改时间", value: info.last_modified && !info.last_modified.startsWith("0001-") ? formatFileTime(info.last_modified) : "—"},
                    {label: "内容类型", value: info.content_type || "—"},
                    {label: "存储类型", value: info.storage_class || "—"},
                ].map(({label, value}) => <div className="file-properties-item" key={label}>
                    <div className="file-properties-label">{label}</div>
                    <div className="file-properties-value">{value}</div>
                </div>)}
                {[
                    {label: "文件路径", value: info.key},
                    {label: "ETag", value: info.etag},
                ].map(({label, value}) => <div className="file-properties-item is-wide" key={label}>
                    <div className="file-properties-label">{label}</div>
                    <div className="file-properties-value-row">
                        <div className="file-properties-value file-properties-path">{value || "—"}</div>
                        {value && copyButton(value, label)}
                    </div>
                </div>)}
                <div className="file-properties-item is-wide">
                    <div className="file-properties-label">元数据</div>
                    {metadata.length ? <dl className="file-info-metadata">
                        {metadata.map(([key, value]) => <div key={key}>
                            <dt>{key}</dt><dd>{value || "—"}</dd>
                        </div>)}
                    </dl> : <div className="file-info-empty">暂无元数据</div>}
                </div>
            </div>
        </section>}
    </ProModal>;
};

export default FileInfo;
