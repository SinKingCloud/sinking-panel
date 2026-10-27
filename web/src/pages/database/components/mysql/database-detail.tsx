import {useLayoutEffect, useRef, useState} from "react";
import {App, Modal, Spin, Tag} from "antd";
import {Title, useTheme} from "sinking-antd";
import {useMysqlClient} from "./context";
import useStyles from "./component-styles/client";
import useFormStyles from "./form-styles";

export default ({record, onCancel}: any) => {
    const {api} = useMysqlClient();
    const {message} = App.useApp();
    const {styles} = useStyles();
    const {styles: formStyles} = useFormStyles();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const lastRecord = useRef(record);
    const currentRecord = record || lastRecord.current;
    const [data, setData] = useState<any>();
    const [loaded, setLoaded] = useState("");
    const key = currentRecord ? JSON.stringify([currentRecord.id, currentRecord.database_name || currentRecord.name]) : "";
    useLayoutEffect(() => {
        if (!record?.id) return;
        lastRecord.current = record;
        let disposed = false;
        setData(undefined); setLoaded("");
        if (record?.id) {
            void api.database_info(record, {database_name: record.database_name || record.name}).then((info) => {
                if (disposed) return;
                if (!info) throw new Error("获取数据库详情失败");
                setData(info); setLoaded(key);
            }).catch((reason) => {
                if (disposed) return;
                disposed = true; onCancel();
                message.error(reason?.message || "获取数据库详情失败");
            });
        }
        return () => { disposed = true; };
    }, [Boolean(record), key]);
    const size = () => {
        let value = Number(data?.size || 0), index = 0;
        if (!Number.isFinite(value) || value <= 0) return "0 B";
        const units = ["B", "KB", "MB", "GB", "TB"];
        while (value >= 1024 && index < units.length - 1) { value /= 1024; index++; }
        return `${value.toFixed(index === 0 ? 0 : 2)} ${units[index]}`;
    };

    return <Modal open={Boolean(record)} title={<Title>数据库详情</Title>} footer={null}
        width={compact ? 500 : 540} zIndex={1080} rootClassName={styles.modal} onCancel={onCancel}
        style={{maxWidth: "calc(100vw - 24px)"}}
        destroyOnHidden focusable={{focusTriggerAfterClose: false}}>
        {!currentRecord || !data || loaded !== key ? <div className={styles.loading}><Spin/></div>
            : <div className={formStyles.detail}>
                <div className="db-detail-heading"><span>{data?.name || currentRecord.database_name || currentRecord.name}</span><Tag variant="filled">MySQL</Tag></div>
                <div className="db-detail-metrics">
                    <div className="db-detail-metric"><span>数据表</span><strong>{data?.table_count ?? "—"}</strong></div>
                    <div className="db-detail-metric"><span>占用空间</span><strong>{data?.size == null ? "—" : size()}</strong></div>
                </div>
                <section className="db-detail-group">
                    <h3>数据库属性</h3>
                    <dl className="db-detail-grid">
                        <div><dt>字符集</dt><dd>{data?.charset || "—"}</dd></div>
                        <div><dt>排序规则</dt><dd>{data?.collation || "—"}</dd></div>
                    </dl>
                </section>
            </div>}
    </Modal>;
};
