import {forwardRef, lazy, memo, Suspense, useCallback, useEffect, useImperativeHandle, useRef, useState} from "react";
import {Alert, Button, Descriptions, Grid, Segmented, Spin, Tag, Typography} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import dayjs from "dayjs";
import Log from "@/pages/components/task-log";
import type {LogRef} from "@/pages/components/task-log";

import {getContainerInfo, getContainerLog} from "@/service/api/container";

import {formatSize} from "@/utils/string";
import useEnum from "@/utils/enum";
import useStyles from "./styles";

const ContainerResources = lazy(() => import("./resources"));

const dateText = (value?: number) => value ? dayjs.unix(value).format("YYYY-MM-DD HH:mm:ss") : "—";

const DetailContent = ({record, active, initialTab}: {
    record: any; active: boolean; initialTab: string;
}) => {
    const {styles} = useStyles();
    const [enumData] = useEnum("container");
    const screens = Grid.useBreakpoint();
    const [tab, setTab] = useState(initialTab);
    const [sampling, setSampling] = useState(initialTab === "resources");
    const [info, setInfo] = useState<any>();
    const [status, setStatus] = useState<any>();
    const [infoError, setInfoError] = useState("");
    const [loading, setLoading] = useState(true);
    const [refresh, setRefresh] = useState(0);
    const updateStatus = useCallback((next: any) => {
        // 资源数据每秒采样；只有概览实际展示的字段变化时才更新详情。
        setStatus((current: any) => current && ["status", "started_at", "ended_at", "pid", "manual_stop", "restart_pending", "exit_code", "error"]
            .every((key) => current[key] === next?.[key]) ? current : next);
    }, []);

    useEffect(() => {
        if (!active) return;
        let disposed = false;
        setLoading(true);
        void getContainerInfo({body: {id: record.id}}).then((response) => {
            if (disposed) return;
            if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "获取容器详情失败");
            setInfo({
                ...response?.data,
                command: Array.isArray(response?.data?.command) ? response?.data?.command : [],
                mounts: Array.isArray(response?.data?.mounts) ? response?.data?.mounts.filter(Boolean) : [],
                env: Array.isArray(response?.data?.env) ? response?.data?.env.filter((value: unknown) => typeof value === "string") : [],
            });
            setInfoError("");
        }).catch((error) => {
            if (!disposed) setInfoError(error?.message || "获取容器详情失败");
        }).finally(() => { if (!disposed) setLoading(false); });
        return () => { disposed = true; };
    }, [active, record.id, refresh]);

    const descriptions = {
        size: "small" as const, column: screens.sm ? 2 : 1,
        layout: screens.sm ? "vertical" as const : "horizontal" as const, colon: false,
        styles: screens.sm ? undefined : {label: {flex: "0 0 84px"}, content: {minWidth: 0}},
    };
    const runningState = status?.status || info?.status || record.status;
    const overview = tab === "overview" && <dl className="detail-overview">{[
        {key: "name", label: "容器名称", children: info?.name || record.name || "—"},
        {key: "id", label: "容器 ID", children: <Typography.Text copyable={{text: record.id}}>{record.id}</Typography.Text>},
        {key: "image", label: "使用镜像", children: record.image_name || record.image_id},
        {key: "state", label: "运行状态", children: <Tag variant="filled" color={runningState === "running" ? "success" : runningState === "failed" ? "error" : "default"}>{enumData?.status?.[runningState] || "未知状态"}</Tag>},
        {key: "user", label: "运行用户", children: info ? info?.user || "沿用镜像" : "—"},
        {key: "network", label: "网络模式", children: "宿主机网络"},
        {key: "created", label: "创建时间", children: info ? dateText(info?.created_at) : record.create_time || "—"},
        {key: "started", label: "启动时间", children: dateText(status?.started_at ?? info?.started_at)},
        {key: "ended", label: "停止时间", children: dateText(status?.ended_at ?? info?.ended_at)},
        {key: "pid", label: "进程 PID", children: (status?.pid ?? info?.pid) || "—"},
        {key: "auto", label: "自动重启", children: (info?.auto_restart ?? record.auto_restart) ? (status?.manual_stop ?? info?.manual_stop ?? record.manual_stop) ? "已开启，手动停止中" : status?.restart_pending ? "等待自动重启" : "已开启" : "已关闭"},
        {key: "exit", label: "退出码", children: ["stopped", "failed"].includes(runningState) ? status?.exit_code ?? info?.exit_code ?? "—" : "—"},
        {key: "rootfs", label: "容器根目录", wide: true, children: <Typography.Text copyable={info?.rootfs ? {text: info?.rootfs} : false}>{info?.rootfs || "—"}</Typography.Text>},
        {key: "log", label: "日志路径", wide: true, children: <Typography.Text copyable={info?.log_path ? {text: info?.log_path} : false}>{info?.log_path || "—"}</Typography.Text>},
    ].map(({key, label, children, wide}) => <div key={key} className={wide ? "detail-overview-wide" : undefined}>
        <dt>{label}</dt><dd>{children}</dd>
    </div>)}</dl>;

    const config = tab === "config" && <div className="detail-config">
        <Descriptions {...descriptions} items={[
            {key: "dir", label: "工作目录", children: info ? info?.working_dir || "沿用镜像" : "—"},
            {key: "readonly", label: "根文件系统", children: info ? info?.writable_layer ? "可写" : "只读" : "—"},
            {key: "memory", label: "内存上限", children: info ? info?.resources?.memory ? formatSize(info?.resources?.memory) : "不限制" : "—"},
            {key: "cpu", label: "CPU 上限", children: info ? info?.resources?.cpu_quota ? `${Number((info?.resources?.cpu_quota / (info?.resources?.cpu_period || 100000)).toFixed(2))} 核` : "不限制" : "—"},
            {key: "pids", label: "进程数上限", children: info ? info?.resources?.pids_limit || "使用默认值" : "—"},
            {key: "uid", label: "UID 映射起点", children: info?.uid_map_start ?? "—"},
            {key: "gid", label: "GID 映射起点", children: info?.gid_map_start ?? "—"},
        ]}/>
        <section className="detail-field detail-command">
            <span className="detail-label">启动参数</span>
            <div className="detail-code"><Typography.Text copyable={info?.command?.length ? {text: info?.command?.join("\n")} : false}>
                {info ? info?.command?.join("\n") || "沿用镜像" : "—"}
            </Typography.Text></div>
        </section>
        <section className="detail-field">
            <span className="detail-label">数据挂载</span>
            {info?.mounts?.length ? <div className="detail-mounts">{info?.mounts?.map((mount: any, index: number) => <dl className="detail-mount" key={index}>
                <div className="detail-field">
                    <dt className="detail-label">宿主目录</dt>
                    <dd><Typography.Text copyable={{text: mount.source}}>{mount.source}</Typography.Text></dd>
                </div>
                <div className="detail-field">
                    <dt className="detail-field-heading"><span className="detail-label">容器目录</span><span className="detail-muted">{mount.read_only ? "只读" : "读写"}</span></dt>
                    <dd><Typography.Text copyable={{text: mount.destination}}>{mount.destination}</Typography.Text></dd>
                </div>
            </dl>)}</div> : <div className="detail-code detail-muted">{info ? "未配置挂载，数据保存在容器写层，删除容器时会清除。" : "—"}</div>}
        </section>
        <section className="detail-field">
            <span className="detail-label">环境变量</span>
            {info?.env?.length ? <dl className="detail-config">{info?.env?.map((line: string, index: number) => {
                const separator = line.indexOf("=");
                const name = separator < 0 ? line : line.slice(0, separator);
                const value = separator < 0 ? "" : line.slice(separator + 1);
                return <div className="detail-field detail-code" key={index}>
                    <dt className="detail-label">{name}</dt>
                    <dd style={{margin: 0, minWidth: 0}}><Typography.Text copyable={{text: line, tooltips: [`复制 ${name || "环境变量"}`, "已复制"]}}>
                        {value || <span className="detail-label">空值</span>}
                    </Typography.Text></dd>
                </div>;
            })}</dl> : <div className="detail-code detail-muted">{info ? "沿用镜像环境变量" : "—"}</div>}
        </section>
        <section className="detail-field">
            <span className="detail-label">存储路径</span>
            <Descriptions {...descriptions} column={1} items={[
                {key: "upper", label: "容器写层", children: <Typography.Text copyable={info?.upper_dir ? {text: info?.upper_dir} : false}>{info?.upper_dir || "—"}</Typography.Text>},
                {key: "work", label: "工作路径", children: <Typography.Text copyable={info?.work_dir ? {text: info?.work_dir} : false}>{info?.work_dir || "—"}</Typography.Text>},
            ]}/>
        </section>
    </div>;

    return <div className={styles.detail}>
        {infoError && <Alert type="error" showIcon title={infoError}
            action={<Button size="small" color="default" variant="filled" loading={loading} onClick={() => setRefresh((value) => value + 1)}>重试</Button>}/>}
        {(status?.error || info?.error) && <Alert type="warning" showIcon title={status?.error || info?.error}/>}
        <Segmented className="detail-switch" aria-label="容器详情内容" size="middle" value={tab} onChange={(value) => {
            setTab(value);
            if (value === "resources") setSampling(true);
        }}
            options={[
                {value: "overview", label: "基本信息"},
                {value: "config", label: "运行配置"},
                {value: "resources", label: "资源状态"},
            ]}/>
        <div className="detail-content">
            {sampling && <Suspense fallback={tab === "resources" ? <div className="detail-loading"><Spin/></div> : null}>
                <ContainerResources id={record.id} active={active} visible={tab === "resources"} onStatus={updateStatus}/>
            </Suspense>}
            {tab !== "resources" && <Spin spinning={loading} classNames={{root: "detail-data", container: "detail-data"}}>
                {tab === "overview" ? overview : config}
            </Spin>}
        </div>
    </div>;
};

const ContainerDetail = forwardRef<any>((_, ref) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const logRef = useRef<LogRef>(null);
    const [open, setOpen] = useState(false);
    const [target, setTarget] = useState<{record: any; tab: string; session: number}>();
    const openRef = useRef(false);
    const sessionRef = useRef(0);
    const showLog = useCallback((record: any) => logRef.current?.open({id: record.id, name: record.name || record.id}), []);
    useImperativeHandle(ref, () => ({open: (record: any, tab = "overview") => {
        if (tab === "log") { showLog(record); return; }
        openRef.current = true;
        setTarget({record, tab, session: ++sessionRef.current});
        setOpen(true);
    }}), [showLog]);

    return <>
        <ProModal title={<Title>容器详情</Title>}
                  width={compact ? 600 : 660} onCancel={() => { openRef.current = false; setOpen(false); }}
                  modalProps={{open, zIndex: 1050, footer: null, destroyOnHidden: true, afterClose: () => { if (!openRef.current) setTarget(undefined); },
                      style: {top: screens.md ? 80 : 20, paddingBottom: 40, maxWidth: "calc(100vw - 24px)"}, mask: {closable: true}} as any}>
            {target && <DetailContent key={target.session} record={target.record} active={open} initialTab={target.tab}/>}
        </ProModal>
        <Log ref={logRef} title="容器日志" request={getContainerLog} body={{action: "read"}}/>
    </>;
});

ContainerDetail.displayName = "ContainerDetail";
export default memo(ContainerDetail);
