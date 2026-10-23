import {useLayoutEffect, useRef, useState} from "react";
import {App, Modal, Spin, Tag} from "antd";
import {Title, useTheme} from "sinking-antd";
import useStyles from "./styles/client";
import useFormStyles from "./manager-form-styles";
import type {DatabasePanelRecord, DatabasePanelRequest} from "./entities";

const labels: Record<string, string> = {
    charset: "字符集", collation: "排序规则", owner: "所有者", encoding: "编码", status: "状态",
    table_count: "数据表", size: "占用空间", collections: "集合", objects: "文档",
    views: "视图", indexes: "索引", avg_object_size: "平均文档大小",
    data_size: "数据大小", storage_size: "存储大小", index_size: "索引大小",
    keys: "键数量", expires: "过期键", avg_ttl: "平均有效期",
    page_size: "页大小", page_count: "页数量", free_pages: "空闲页", tables: "数据表", journal_mode: "日志模式",
    exists: "实际存在", allow_connections: "允许连接", read_only: "只读模式",
    recovery_model: "恢复模式", compatibility_level: "兼容级别",
    documents: "文档数量", size_bytes: "占用空间", number_of_shards: "主分片", replicas: "副本数量", aliases: "索引别名", creation_time: "创建时间", uuid: "索引 UUID",
};

export default ({record, type, engineName, request, onCancel}: {
    record?: DatabasePanelRecord; type: number; engineName: string; request: DatabasePanelRequest; onCancel: () => void;
}) => {
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {styles: formStyles} = useFormStyles();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const lastRecord = useRef(record);
    const currentRecord = record || lastRecord.current;
    const [data, setData] = useState<any>();
    const [loaded, setLoaded] = useState("");
    const key = currentRecord ? JSON.stringify([currentRecord.id, type, currentRecord.database_name || currentRecord.name]) : "";
    useLayoutEffect(() => {
        if (!record?.id) return;
        lastRecord.current = record;
        let disposed = false;
        setData(undefined);
        setLoaded("");
        if (record?.id) {
            void request(type === 7 ? "info" : "database_info", {id: record.id, database_name: record.database_name || record.name}).then((response) => {
                if (disposed) return;
                if (response?.code !== 200 || !response.data) throw new Error(response?.message || "获取数据库详情失败");
                setData(response.data);
                setLoaded(key);
            }).catch((reason) => {
                if (disposed) return;
                disposed = true; onCancel(); message.error(reason.message || "获取数据库详情失败");
            });
        }
        return () => { disposed = true; };
    }, [Boolean(record), key]);
    const indexSettings = data?.settings?.index || data?.settings || {};
    const details = type === 7 ? {
        documents: data?.documents, size_bytes: data?.size_bytes, number_of_shards: indexSettings.number_of_shards ?? indexSettings["index.number_of_shards"],
        replicas: indexSettings.number_of_replicas ?? indexSettings["index.number_of_replicas"], aliases: Object.keys(data?.aliases || {}).join(", ") || "—",
        creation_time: indexSettings.creation_date || indexSettings["index.creation_date"] ? new Date(Number(indexSettings.creation_date || indexSettings["index.creation_date"])).toLocaleString() : "—",
        uuid: indexSettings.uuid || indexSettings["index.uuid"],
    } : data;
    const metrics = ["table_count", "tables", "collections", "keys", "objects", "size", "storage_size", "documents", "size_bytes"]
        .filter((key) => details?.[key] != null).slice(0, 3);
    const formatSize = (input: any) => {
        let value = Number(input), position = 0;
        if (!Number.isFinite(value) || value <= 0) return "0 B";
        const units = ["B", "KB", "MB", "GB", "TB"];
        while (value >= 1024 && position < units.length - 1) { value /= 1024; position++; }
        return `${value.toFixed(position ? 2 : 0)} ${units[position]}`;
    };
    const displayValue = (key: string, value: any) => value == null ? "—" : typeof value === "boolean" ? value ? "是" : "否"
        : ["size", "data_size", "storage_size", "index_size", "avg_object_size", "page_size", "size_bytes"].includes(key)
            ? formatSize(value) : key === "avg_ttl" ? `${Math.round(Number(value) / 1000)} 秒`
                : typeof value === "object" ? JSON.stringify(value) : String(value);

    return <Modal open={Boolean(record)} title={<Title>{type === 7 ? "索引详情" : "数据库详情"}</Title>} footer={null}
        width={compact ? 500 : 540} zIndex={1080} rootClassName={styles.modal} onCancel={onCancel}
        style={{maxWidth: "calc(100vw - 24px)"}}
        destroyOnHidden focusable={{focusTriggerAfterClose: false}}>
        {!currentRecord || !data || loaded !== key ? <div className={styles.loading}><Spin/></div>
            : <div className={formStyles.detail}>
                <div className="db-detail-heading"><span>{currentRecord.name}</span><Tag variant="filled">{engineName}</Tag></div>
                {metrics.length > 0 && <div className="db-detail-metrics">{metrics.map((key) => <div className="db-detail-metric" key={key}>
                    <span>{labels[key]}</span><strong>{displayValue(key, details[key])}</strong>
                </div>)}</div>}
                <section className="db-detail-group">
                    <h3>{type === 7 ? "索引属性" : "数据库属性"}</h3>
                    <dl className="db-detail-grid">{Object.entries(details || {}).filter(([key]) => labels[key] && !metrics.includes(key))
                        .map(([key, value]) => <div key={key}>
                            <dt>{labels[key] || key}</dt><dd>{displayValue(key, value)}</dd>
                        </div>)}</dl>
                </section>

            </div>
        }
    </Modal>;
};
