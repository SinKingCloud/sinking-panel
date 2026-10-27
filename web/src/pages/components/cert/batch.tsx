import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Button, ConfigProvider, Empty, Form, Grid, Progress, Spin, Switch, Tooltip} from "antd";
import {createStyles} from "antd-style";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import {getCertList, obtainCert, renewCert} from "@/service/api/cert";
import {getEnum as getSystemEnum} from "@/service/api/system";
import SecretManager from "@/pages/components/secret";
import type {SecretChange, SecretManagerRef} from "@/pages/components/secret";
import BatchVerification from "./batch-verification";
import type {BatchVerificationRef} from "./batch-verification";
import BatchForm from "./batch-form";

export interface BatchCertRef {
    open: (domains: string[], onObtained?: (cert: any) => void) => void;
    close: () => void;
}


const statusNames: Record<string, string> = {
    pending: "待申请", running: "申请中", waiting: "待验证", success: "成功", failed: "失败", skipped: "已跳过",
};

const isIPAddress = (value: string) => {
    const address = value.replace(/^\[([^\]]+)\]$/, "$1");
    if (/^(?:0|[1-9]\d{0,2})(?:\.(?:0|[1-9]\d{0,2})){3}$/.test(address)) {
        return address.split(".").every((part) => Number(part) <= 255);
    }
    if (!address.includes(":") || !/^[\da-f:.]+$/i.test(address)) return false;
    try {
        return new URL(`http://[${address}]`).hostname.startsWith("[");
    } catch {
        return false;
    }
};

const useStyles = createStyles(({css, token}: any, props: {compact: boolean}) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); outline: none; }
        .ant-modal-body { padding: 0 !important; overflow: hidden; }
        .ant-modal-footer { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
        .ant-modal-footer > .ant-btn { margin-inline-start: 0 !important; }
    `,
    shell: css`
        min-height: min(${props.compact ? 380 : 410}px, calc(100dvh - 290px));
        max-height: calc(100dvh - 290px);
        padding: 4px;
        overflow-x: hidden;
        overflow-y: auto;
        scrollbar-gutter: stable;
        overscroll-behavior: contain;
        @supports not (max-height: 100dvh) { max-height: calc(100vh - 290px); }
        @media (max-width: 760px) {
            min-height: min(${props.compact ? 380 : 410}px, calc(100dvh - 210px));
            max-height: calc(100dvh - 210px);
            @supports not (max-height: 100dvh) { max-height: calc(100vh - 210px); }
        }
    `,
    loading: css`
        min-height: 220px;
        display: flex;
        align-items: center;
        justify-content: center;
    `,
    progress: css`
        margin-bottom: ${props.compact ? 14 : 18}px;
        .progress-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 6px 12px; margin-bottom: 8px; }
        .progress-title { color: ${token.colorText}; font-weight: 500; }
        .progress-count, .progress-note { color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .progress-stats { display: flex; flex-wrap: wrap; gap: 6px 16px; margin-top: 8px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .progress-success { color: ${token.colorSuccess}; }
        .progress-failed { color: ${token.colorError}; }
        .progress-note { margin-top: 8px; overflow-wrap: anywhere; }
    `,
    taskHeader: css`
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        margin-bottom: 10px;
        color: ${token.colorTextSecondary};
        font-size: ${token.fontSizeSM}px;
        .ant-btn { flex: none; font-size: ${token.fontSizeSM}px; }
    `,
    taskList: css`
        margin: 0;
        padding: 0;
        list-style: none;
        > li {
            padding: 10px;
        }
        > li + li { margin-top: 4px; }
        .task-top { display: flex; align-items: center; gap: 10px; }
        .task-identity { flex: 1; min-width: 0; }
        .task-domain { display: block; font-size: ${props.compact ? 12 : 13}px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .task-method { display: flex; flex-wrap: wrap; gap: 2px 8px; margin-top: 2px; color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; overflow-wrap: anywhere; }
        .task-status { flex: none; font-size: 12px; white-space: nowrap; color: ${token.colorTextTertiary}; }
        .task-status.success { color: ${token.colorSuccess}; }
        .task-status.failed { color: ${token.colorError}; }
        .task-status.running, .task-status.waiting { color: ${token.colorPrimary}; }
        .task-error { margin-top: 5px; color: ${token.colorError}; font-size: 12px; overflow-wrap: anywhere; }
        .task-actions { display: flex; justify-content: flex-end; flex: none; width: 48px; }
        .task-actions .ant-btn { font-size: ${token.fontSizeSM}px; }
        .task-status.running { display: inline-flex; align-items: center; gap: 5px; }
    `,
}));

const BatchCert = forwardRef<BatchCertRef, {onChange?: () => void; secretVersion?: number}>(({onChange, secretVersion = 0}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const [form] = Form.useForm();
    const listRef = useRef<HTMLDivElement>(null);
    const modalRef = useRef<any>(null);
    const verificationRef = useRef<BatchVerificationRef | null>(null);
    const secretManagerRef = useRef<SecretManagerRef | null>(null);
    const openRef = useRef(false);
    const certificatesReadyRef = useRef(false);
    const versionRef = useRef(0);
    const busyRef = useRef(false);
    const pauseRef = useRef(false);
    const closePendingRef = useRef(false);
    const reopenRef = useRef<(() => void) | null>(null);
    const tasksRef = useRef<any[]>([]);
    const requestedDomainsRef = useRef<string[]>([]);
    const configurationsRef = useRef(new Map<string, {id?: string | number; values: Record<string, any>}>());
    const queueSelectionRef = useRef<number[] | null>(null);
    const manualRef = useRef<any>(null);
    const onObtainedRef = useRef<((cert: any) => void) | undefined>(undefined);
    const onChangeRef = useRef(onChange);
    onChangeRef.current = onChange;
    const secretRequestRef = useRef(0);
    const secretVisitRef = useRef<{version: number; index: number} | null>(null);
    const createdSecretRef = useRef<number | null>(null);
    const [tasks, setTasks] = useState<any[]>([]);
    const [mergeDomains, setMergeDomains] = useState(true);
    const [checkingCertificates, setCheckingCertificates] = useState(false);
    const [certificateError, setCertificateError] = useState("");
    const [busy, setBusy] = useState(false);
    const [paused, setPaused] = useState(false);
    const [closing, setClosing] = useState(false);
    const [started, setStarted] = useState(false);
    const [editingConfig, setEditingConfig] = useState(true);
    const [expandedIndex, setExpandedIndex] = useState<number | null>(null);
    const [scrollRequest, setScrollRequest] = useState<{field: (string | number)[]; version: number} | null>(null);
    const [manual, setManual] = useState<any>(null);
    const [secretOptions, setSecretOptions] = useState<{value: number; label: string}[]>([]);
    const [secretLoading, setSecretLoading] = useState(false);
    const [secretError, setSecretError] = useState(false);
    const counts = useMemo(() => {
        const result: Record<string, number> = {pending: 0, running: 0, waiting: 0, failed: 0, success: 0, skipped: 0, renewal: 0};
        for (const task of tasks) {
            result[task.status]++;
            if (task.certificate_id) result.renewal++;
        }
        return result;
    }, [tasks]);
    const unfinishedCount = tasks.length - counts.success - counts.skipped;
    const finished = started && !busy && !manual && counts.pending === 0;
    const configLocked = busy || Boolean(manual) || checkingCertificates || Boolean(certificateError);
    const actionLabel = counts.renewal === tasks.length && tasks.length > 0 ? "续签" : counts.renewal > 0 ? "处理" : "申请";
    const totalCount = tasks.length - counts.skipped;
    const completedCount = counts.success + counts.failed;
    const manualTask = manual ? tasks[manual.index] : undefined;
    const verificationSession = useMemo(() => manual ? {domain: manualTask?.domain || "", data: manual.data, error: manualTask?.error} : null, [manual, manualTask?.domain, manualTask?.error]);

    useEffect(() => {
        if (!scrollRequest) return;
        let frame = 0;
        const scroll = () => {
            if (!openRef.current || scrollRequest.version !== versionRef.current) return;
            // 等列表中的展开、收起动画都结束，再定位错误字段。
            if (listRef.current?.querySelector(".ant-motion-collapse")) {
                frame = window.requestAnimationFrame(scroll);
                return;
            }
            form.scrollToField(scrollRequest.field, {block: "center", focus: true});
            setScrollRequest(null);
        };
        frame = window.requestAnimationFrame(scroll);
        return () => window.cancelAnimationFrame(frame);
    }, [form, scrollRequest]);

    const updateTask = useCallback((index: number, changes: Partial<any>) => {
        tasksRef.current = tasksRef.current.map((task, position) => position === index ? {...task, ...changes} : task);
        setTasks(tasksRef.current);
    }, []);

    const cancelManual = useCallback(() => {
        const session = manualRef.current;
        manualRef.current = null;
        if (session) void (session.method === "renew" ? renewCert : obtainCert)({body: {type: 1, action: "cancel", session_id: session?.data?.session_id, id: session.id}}).catch(() => undefined);
    }, []);

    const loadSecrets = useCallback(async () => {
        const version = versionRef.current;
        const request = ++secretRequestRef.current;
        setSecretLoading(true);
        setSecretError(false);
        try {
            const response = await getSystemEnum({body: {name: "secret"}});
            if (!openRef.current || version !== versionRef.current || request !== secretRequestRef.current) return;
            if (response?.code !== 200 || !response?.data?.name) throw new Error("获取密钥失败");
            const options = Object.entries(response?.data?.name || {}).map(([value, label]) => ({value: Number(value), label: String(label)}));
            setSecretOptions(options);
            const secretIds = new Set(options.map((option) => option.value));
            const configurations = form.getFieldValue("items") || [];
            configurations.forEach((values: any, index: number) => {
                const selected = Number(values?.secret_id);
                if (selected && !secretIds.has(selected)) form.setFieldValue(["items", index, "secret_id"], undefined);
            });
        } catch {
            if (openRef.current && version === versionRef.current && request === secretRequestRef.current) setSecretError(true);
        } finally {
            if (openRef.current && version === versionRef.current && request === secretRequestRef.current) setSecretLoading(false);
        }
    }, [form]);

    const close = useCallback(() => {
        setScrollRequest(null);
        reopenRef.current = null;
        if (busyRef.current) {
            closePendingRef.current = true;
            pauseRef.current = true;
            setPaused(true);
            setClosing(true);
            return;
        }
        versionRef.current += 1;
        secretRequestRef.current += 1;
        openRef.current = false;
        certificatesReadyRef.current = false;
        busyRef.current = false;
        pauseRef.current = false;
        closePendingRef.current = false;
        secretVisitRef.current = null;
        createdSecretRef.current = null;
        onObtainedRef.current = undefined;
        requestedDomainsRef.current = [];
        configurationsRef.current.clear();
        cancelManual();
        secretManagerRef.current?.close();
        verificationRef.current?.hide();
        modalRef.current?.hide();
        setBusy(false);
        setPaused(false);
        setClosing(false);
        setManual(null);
        setSecretLoading(false);
        setCheckingCertificates(false);
        setCertificateError("");
    }, [cancelManual]);

    const loadTasks = useCallback((merge: boolean) => {
        if (!openRef.current || busyRef.current || manualRef.current) return;
        if (certificatesReadyRef.current) {
            tasksRef.current.forEach((task, index) => {
                configurationsRef.current.set(task.domain, {id: task.certificate_id, values: {...form.getFieldValue(["items", index])}});
            });
        }
        const version = ++versionRef.current;
        secretRequestRef.current += 1;
        secretVisitRef.current = null;
        createdSecretRef.current = null;
        certificatesReadyRef.current = false;
        setMergeDomains(merge);
        setCertificateError("");
        setExpandedIndex(null);
        setScrollRequest(null);
        const requested = requestedDomainsRef.current;
        let targets = requested;
        if (merge) {
            const wildcards = new Map<string, string>();
            const groups = new Map<string, number>();
            requested.forEach((domain) => {
                if (domain.startsWith("*.") || domain.includes("..") || isIPAddress(domain)) return;
                const dot = domain.indexOf(".");
                if (dot <= 0) return;
                const parent = domain.slice(dot + 1);
                if (!parent.includes(".")) return;
                // 同一直接父域下的多个子域名合并为一张泛域名证书。
                const wildcard = "*." + parent;
                wildcards.set(domain, wildcard);
                groups.set(wildcard, (groups.get(wildcard) || 0) + 1);
            });
            const merged = new Set(requested.map((domain) => {
                const wildcard = wildcards.get(domain);
                return wildcard && (groups.get(wildcard) || 0) > 1 ? wildcard : domain;
            }));
            targets = [...merged].filter((domain) => {
                if (domain.startsWith("*.") || isIPAddress(domain)) return true;
                const dot = domain.indexOf(".");
                return dot < 0 || !merged.has("*." + domain.slice(dot + 1));
            });
        }
        if (!targets.length) return;
        tasksRef.current = targets.map((domain) => ({domain, status: "pending"}));
        queueSelectionRef.current = null;
        setTasks(tasksRef.current);
        setSecretError(false);
        setSecretLoading(false);
        setSecretOptions([]);
        setCheckingCertificates(true);
        form.resetFields();
        void (async () => {
            let failed = false;
            try {
                const certificates: any[] = new Array(targets.length);
                let nextIndex = 0;
                // 限制并行查询数量；包含已过期证书，按最新记录优先续签。
                await Promise.all(Array.from({length: Math.min(4, targets.length)}, async () => {
                    while (!failed && openRef.current && version === versionRef.current) {
                        const index = nextIndex++;
                        if (index >= targets.length) return;
                        const domain = targets[index].replace(/^\[([^\]]+)\]$/, "$1");
                        for (let page = 1; !failed && openRef.current && version === versionRef.current; page++) {
                            const response = await getCertList({body: {domain, page, page_size: 100, order_by_field: "id", order_by_type: "desc"}});
                            if (failed || !openRef.current || version !== versionRef.current) return;
                            if (response?.code !== 200 || !Array.isArray(response?.data?.list)) throw new Error(response?.message || "获取已有证书失败");
                            const list = response?.data?.list;
                            certificates[index] = list?.find((certificate: any) => {
                                if (!certificate?.id || ![1, 2].includes(Number(certificate?.type))) return false;
                                let names = certificate?.domains;
                                if (typeof names === "string") {
                                    try { names = JSON.parse(names); } catch { return false; }
                                }
                                if (!Array.isArray(names)) return false;
                                const domains = [...new Set(names.map((name: any) => String(name).trim().toLowerCase().replace(/\.+$/, "").replace(/^\[([^\]]+)\]$/, "$1")))];
                                // 当前续签接口签发单个域名，避免覆盖多域名证书的其他 SAN。
                                return domains.length === 1 && domains[0] === domain;
                            });
                            if (certificates[index] || page * 100 >= Number(response?.data?.total || 0)) break;
                            if (!list?.length) throw new Error("获取已有证书失败，请重试");
                        }
                    }
                }));
                if (!openRef.current || version !== versionRef.current) return;
                tasksRef.current = targets.map((domain, index) => ({domain, status: "pending", certificate_id: certificates[index]?.id}));
                setTasks(tasksRef.current);
                form.setFieldsValue({items: targets.map((domain, index) => {
                    const certificate = certificates[index];
                    const saved = configurationsRef.current.get(domain);
                    const type = Number(certificate?.type) === 1 ? 1 : 2;
                    const challenge = isIPAddress(domain) ? "http" : domain.startsWith("*.") ? "dns" : certificate?.challenge === "dns" ? "dns" : "http";
                    const values = {
                        challenge, type,
                        secret_id: type === 2 && challenge === "dns" ? Number(certificate?.secret_id || 0) || undefined : undefined,
                        auto_renew: type === 1 ? 0 : Number(certificate?.auto_renew ?? 1),
                        ca: certificate ? undefined : "production",
                        ...saved?.values,
                    };
                    if (saved && String(saved.id || "") !== String(certificate?.id || "")) values.ca = certificate ? undefined : "production";
                    if (isIPAddress(domain)) values.challenge = "http";
                    else if (domain.startsWith("*.")) values.challenge = "dns";
                    if (Number(values.type) === 1) values.auto_renew = 0;
                    if (Number(values.type) === 1 || values.challenge !== "dns") values.secret_id = undefined;
                    if (!certificate && !values.ca) values.ca = "production";
                    return values;
                })});
                certificatesReadyRef.current = true;
                void loadSecrets();
            } catch (error) {
                failed = true;
                if (openRef.current && version === versionRef.current) setCertificateError(error instanceof Error ? error.message : "获取已有证书失败");
            } finally {
                if (openRef.current && version === versionRef.current) setCheckingCertificates(false);
            }
        })();
    }, [form, loadSecrets]);

    const open = useCallback((domains: string[], onObtained?: (cert: any) => void) => {
        if (busyRef.current) {
            close();
            reopenRef.current = () => open(domains, onObtained);
            return;
        }
        close();
        requestedDomainsRef.current = [...new Set(domains.map((domain) => domain.trim().toLowerCase().replace(/\.+$/, "")).filter(Boolean))];
        if (!requestedDomainsRef.current.length) return;
        openRef.current = true;
        onObtainedRef.current = onObtained;
        setStarted(false);
        setEditingConfig(true);
        loadTasks(true);
        modalRef.current?.show();
    }, [close, loadTasks]);

    useImperativeHandle(ref, () => ({open, close}), [open, close]);

    useEffect(() => {
        if (openRef.current) void loadSecrets();
    }, [loadSecrets, secretVersion]);

    useEffect(() => () => {
        openRef.current = false;
        certificatesReadyRef.current = false;
        versionRef.current += 1;
        secretRequestRef.current += 1;
        cancelManual();
    }, [cancelManual]);

    const completeTask = useCallback((index: number, certificate: any, version: number) => {
        verificationRef.current?.hide();
        manualRef.current = null;
        setManual(null);
        updateTask(index, {status: "success", error: undefined});
        // 回填或刷新可能让调用方重渲染，队列与已完成项始终由当前会话持有。
        const obtained = onObtainedRef.current;
        const changed = onChangeRef.current;
        for (const callback of [() => obtained?.(certificate), () => changed?.()]) {
            if (!openRef.current || version !== versionRef.current) break;
            try { void Promise.resolve(callback()).catch(() => undefined); } catch {}
        }
    }, [updateTask]);

    const runQueue = useCallback(async (retryFailures: boolean | number = false, submitManual = false, preserveSelection = false) => {
        if (!openRef.current || !certificatesReadyRef.current || busyRef.current || (manualRef.current && !submitManual)) return;
        const version = versionRef.current;
        const current = () => openRef.current && version === versionRef.current;
        busyRef.current = true;
        setBusy(true);
        pauseRef.current = false;
        setPaused(false);
        try {
            const targets = tasksRef.current.flatMap((task, index) => (retryFailures !== false
                ? task.status === "failed" && (retryFailures === true || retryFailures === index)
                : task.status === "pending" && (!preserveSelection || !queueSelectionRef.current || queueSelectionRef.current.includes(index))) ? [index] : []);
            if (!submitManual) {
                await form.validateFields(targets.flatMap((index) => {
                    const values = form.getFieldValue(["items", index]) || {};
                    const task = tasksRef.current[index];
                    const automatic = Number(values.type) !== 1;
                    const needsDNS = automatic && !isIPAddress(task.domain) && (task.domain.startsWith("*.") || values.challenge === "dns");
                    return ["challenge", "type", ...(!task.certificate_id ? ["ca"] : []), ...(automatic ? ["auto_renew"] : []), ...(needsDNS ? ["secret_id"] : [])].map((field) => ["items", index, field]);
                }));
            }
            if (!current()) return;
            const configured = form.getFieldsValue(true).items || [];
            if (retryFailures !== false) {
                queueSelectionRef.current = tasksRef.current.flatMap((task, index) => task.status === "failed" && (retryFailures === true || retryFailures === index) ? [index] : []);
                tasksRef.current = tasksRef.current.map((task, index) => queueSelectionRef.current?.includes(index) ? {...task, status: "pending", error: undefined} : task);
                setTasks(tasksRef.current);
            } else if (!submitManual && !preserveSelection) {
                queueSelectionRef.current = null;
            }
            setStarted(true);
            setEditingConfig(false);
            let activeManual = submitManual ? manualRef.current : null;
            while (current()) {
                if (pauseRef.current && !activeManual) break;
                const index = activeManual?.index ?? tasksRef.current.findIndex((task, index) => task.status === "pending" && (!queueSelectionRef.current || queueSelectionRef.current.includes(index)));
                if (index < 0) break;
                const task = tasksRef.current[index];
                const values = configured[index] || {};
                const isIP = isIPAddress(task.domain);
                const verification = isIP ? "http" : task.domain.startsWith("*.") ? "dns" : values.challenge;
                const isManual = Number(values.type) === 1;
                const method = task.certificate_id ? "renew" : "obtain";
                const body = {
                    ...(task.certificate_id ? {id: task.certificate_id} : {name: task.domain.slice(0, 100)}),
                    action: "start",
                    type: isManual ? 1 : 2,
                    domain: isIP ? task.domain.replace(/^\[([^\]]+)\]$/, "$1") : task.domain,
                    ...(values.ca ? {ca: values.ca} : {}),
                    challenge: verification,
                    secret_id: !isManual && verification === "dns" ? Number(values.secret_id || 0) : 0,
                    auto_renew: isManual ? 0 : Number(values.auto_renew || 0),
                };
                updateTask(index, {
                    status: "running", error: undefined, challenge: verification,
                    type: isManual ? 1 : 2, ca: values.ca,
                    secret_id: body.secret_id,
                    auto_renew: body.auto_renew,
                });
                try {
                    let response: any;
                    if (isManual) {
                        let session = activeManual;
                        if (!session) {
                            const started = await (method === "renew" ? renewCert : obtainCert)({body});
                            const data = started?.data;
                            if (!current()) {
                                if (data?.session_id) void (method === "renew" ? renewCert : obtainCert)({body: {type: 1, action: "cancel", session_id: data?.session_id, id: task.certificate_id}}).catch(() => undefined);
                                return;
                            }
                            if (started?.code !== 200 || !data?.session_id || !Array.isArray(data?.challenges) || !Number.isFinite(data?.expires_at)) {
                                if (data?.session_id) void (method === "renew" ? renewCert : obtainCert)({body: {type: 1, action: "cancel", session_id: data?.session_id, id: task.certificate_id}}).catch(() => undefined);
                                throw new Error(started?.code !== 200 ? started?.message || "获取验证信息失败，请重新申请" : "验证信息无效，请重新申请");
                            }
                            session = {index, data, method, id: task.certificate_id};
                            manualRef.current = session;
                            setManual(session);
                            if (!closePendingRef.current) verificationRef.current?.show();
                            if (data?.challenges?.length > 0 || pauseRef.current) {
                                updateTask(index, {status: "waiting"});
                                return;
                            }
                        }
                        if (session?.data?.expires_at <= Math.floor(Date.now() / 1000)) {
                            cancelManual();
                            verificationRef.current?.hide();
                            setManual(null);
                            throw new Error("验证信息已过期，请重试获取新的验证信息");
                        }
                        response = await (session.method === "renew" ? renewCert : obtainCert)({body: {type: 1, action: "submit", session_id: session?.data?.session_id, id: session.id}});
                    } else {
                        response = await (task.certificate_id ? renewCert : obtainCert)({body});
                    }
                    if (!current()) return;
                    if (response?.code !== 200) throw new Error(response?.message || (task.certificate_id ? "续签失败，请重试" : "申请失败，请重试"));
                    if (!response?.data?.id) throw new Error("未收到完整证书信息，请先在证书管理中检查结果");
                    completeTask(index, response?.data, version);
                    activeManual = null;
                } catch (error) {
                    if (!current()) return;
                    const errorText = error instanceof Error ? error.message : task.certificate_id ? "续签失败，请重试" : "申请失败，请重试";
                    if (manualRef.current?.index === index) {
                        updateTask(index, {status: "waiting", error: errorText});
                        return;
                    }
                    updateTask(index, {status: "failed", error: errorText});
                    activeManual = null;
                }
            }
        } catch (error) {
            if (current()) {
                setEditingConfig(true);
                if (error && typeof error === "object" && "errorFields" in error) {
                    const field = (error as {errorFields?: {name: (string | number)[]}[]}).errorFields?.[0]?.name;
                    if (field?.[0] === "items" && typeof field[1] === "number") setExpandedIndex(field[1]);
                    if (field) setScrollRequest({field, version: versionRef.current});
                } else message.error("请检查申请配置后重试");
            }
        } finally {
            if (current()) {
                busyRef.current = false;
                setBusy(false);
                if (closePendingRef.current) {
                    const reopen = reopenRef.current;
                    close();
                    reopen?.();
                }
            }
        }
    }, [cancelManual, close, completeTask, form, message, updateTask]);

    const openSecrets = useCallback((index: number, create = false) => {
        if (busyRef.current || manualRef.current) return;
        secretVisitRef.current = {version: versionRef.current, index};
        createdSecretRef.current = null;
        if (create) secretManagerRef.current?.openCreate();
        else secretManagerRef.current?.open();
    }, []);

    const toggleSettings = useCallback((index: number) => {
        setScrollRequest(null);
        setExpandedIndex((current) => current === index ? null : index);
    }, []);

    const onSecretChange = useCallback((change: SecretChange) => {
        try { void Promise.resolve(onChangeRef.current?.()).catch(() => undefined); } catch {}
        if (!openRef.current) return;
        const visit = secretVisitRef.current;
        if (visit?.version === versionRef.current && change.action === "create") createdSecretRef.current = change.id;
        if (visit?.version === versionRef.current && change.action === "update" && createdSecretRef.current === change.id && !busyRef.current && !manualRef.current) {
            const values = form.getFieldValue(["items", visit.index]);
            if (Number(values?.type) === 2 && values?.challenge === "dns" && ["pending", "failed"].includes(tasksRef.current[visit.index]?.status)) {
                form.setFieldValue(["items", visit.index, "secret_id"], change.id);
            }
            createdSecretRef.current = null;
        }
        if (change.action === "delete" && createdSecretRef.current === change.id) createdSecretRef.current = null;
        void loadSecrets();
    }, [form, loadSecrets]);

    const releaseManual = (skip = false) => {
        if (busyRef.current || !manualRef.current) return;
        const index = manualRef.current.index;
        verificationRef.current?.hide();
        cancelManual();
        setManual(null);
        updateTask(index, {status: skip ? "skipped" : "pending", error: undefined});
        void runQueue(false, false, true);
    };

    return <>
        <ProModal ref={modalRef} title={<Title>申请证书</Title>} width={compact ? 560 : 600} onCancel={close}
            modalProps={{
                rootClassName: styles.modal, zIndex: 1030, forceRender: true, closable: !closing, keyboard: !closing,
                mask: {closable: !busy && !manual},
                style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                focusable: {focusTriggerAfterClose: false},
                footer: editingConfig ? [
                    <Button key="back" disabled={busy} onClick={() => started ? setEditingConfig(false) : close()}>{started ? "返回" : "取消"}</Button>,
                    <Button key="apply" type="primary" loading={busy} disabled={checkingCertificates || Boolean(certificateError) || unfinishedCount === 0} onClick={form.submit}>
                        {started ? counts.pending > 0 ? `继续${actionLabel}` : "重试" : actionLabel === "处理" ? "开始" : actionLabel}
                    </Button>,
                ] : [
                    <Button key="close" loading={closing} onClick={close}>{finished && !counts.failed ? "完成" : "关闭"}</Button>,
                    busy ? <Button key="pause" type="primary" disabled={paused} onClick={() => {pauseRef.current = true; setPaused(true);}}>{paused ? "暂停中" : "暂停"}</Button>
                        : manual ? <Button key="verify" type="primary" onClick={() => verificationRef.current?.show()}>验证</Button>
                        : counts.pending > 0 || counts.failed > 0 ? <Button key="continue" type="primary" onClick={() => void runQueue(counts.pending === 0 && counts.failed > 0)}>{counts.pending > 0 ? `继续${actionLabel}` : "重试"}</Button>
                        : null,
                ],
            } as any}>
            <div className={styles.shell} style={editingConfig ? {minHeight: 0} : undefined}>
                {editingConfig && <div className={styles.taskHeader}>
                    <span>合并域名</span>
                    <Switch size="small" checked={mergeDomains} disabled={started || busy || Boolean(manual)} aria-label="合并域名"
                        onChange={(value) => {if (!started) loadTasks(value);}}/>
                </div>}
                {checkingCertificates ? <div className={styles.loading}><Spin/></div> : certificateError ? <div className={styles.loading}>
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={certificateError}>
                        <Button color="default" variant="filled" size="small" onClick={() => loadTasks(mergeDomains)}>重试</Button>
                    </Empty>
                </div> : null}
                <ConfigProvider componentSize="small" theme={{
                    token: {fontSize: compact ? 12 : 13, fontSizeSM: 12, controlHeightSM: compact ? 24 : 28},
                    components: {Form: {labelFontSize: compact ? 12 : 13}, Select: {optionFontSize: compact ? 12 : 13}, Button: {contentFontSizeSM: compact ? 12 : 13}},
                }}>
                    <Form form={form} layout="horizontal" colon={false} requiredMark={false} variant="filled" disabled={configLocked} style={editingConfig && !checkingCertificates && !certificateError ? undefined : {display: "none"}}
                        onFinishFailed={({errorFields, outOfDate}) => {
                            if (!openRef.current || outOfDate) return;
                            const field = errorFields[0]?.name;
                            if (!field) return;
                            if (field[0] === "items" && typeof field[1] === "number") setExpandedIndex(field[1]);
                            setScrollRequest({field, version: versionRef.current});
                        }}
                        onFinish={() => {
                            if (!openRef.current || !certificatesReadyRef.current || busyRef.current || manualRef.current) return;
                            tasksRef.current = tasksRef.current.map((task) => task.status === "failed" ? {...task, status: "pending", error: undefined} : task);
                            setTasks(tasksRef.current);
                            void runQueue();
                        }}>
                        <div ref={listRef} role="list" aria-label="域名申请设置">
                            {tasks.map((task, index) => !["success", "skipped"].includes(task.status) && <BatchForm
                                key={`${index}:${task.domain}`}
                                index={index}
                                domain={task.domain}
                                renew={Boolean(task.certificate_id)}
                                form={form}
                                disabled={configLocked}
                                expanded={expandedIndex === index}
                                onToggle={toggleSettings}
                                secretOptions={secretOptions}
                                secretLoading={secretLoading}
                                secretError={secretError}
                                onReloadSecrets={loadSecrets}
                                onManageSecrets={openSecrets}
                                error={task.error}/>)}
                        </div>
                    </Form>
                </ConfigProvider>
                {!editingConfig && <>
                    <div className={styles.progress} aria-live="polite">
                        <div className="progress-heading">
                            <span className="progress-title">{closing ? "正在结束任务" : busy ? paused ? "正在暂停" : manual ? "正在验证" : `正在${actionLabel}` : manual ? "等待手动验证" : counts.pending > 0 ? `等待继续${actionLabel}` : counts.failed > 0 ? counts.success > 0 ? `部分证书${actionLabel}失败` : `${actionLabel}失败` : counts.success > 0 ? `${actionLabel}完成` : "未处理证书"}</span>
                            <span className="progress-count">已处理 {completedCount} / {totalCount}</span>
                        </div>
                        <Progress percent={totalCount > 0 ? Math.round(completedCount / totalCount * 100) : 0} success={{percent: totalCount > 0 ? counts.success / totalCount * 100 : 0}} status={busy ? "active" : counts.failed > 0 ? "exception" : "normal"} showInfo={false} size="small"/>
                        <div className="progress-stats">
                            <span className="progress-success">成功 {counts.success}</span>
                            <span className={counts.failed ? "progress-failed" : undefined}>失败 {counts.failed}</span>
                            <span>待处理 {Math.max(0, totalCount - completedCount)}</span>
                            {counts.skipped > 0 && <span>跳过 {counts.skipped}</span>}
                        </div>
                        {(closing || paused || manual) && <div className="progress-note">
                            {closing ? "当前域名处理结束后关闭，后续域名不再处理。" : busy && paused ? "当前域名处理结束后暂停。" : manual ? `请完成 ${tasks[manual.index]?.domain} 的验证，验证成功后继续下一项。` : "可以调整设置后继续，已成功的证书不会重复处理。"}
                        </div>}
                    </div>
                    <div className={styles.taskHeader}>
                        <span>域名与验证设置</span>
                        {!configLocked && unfinishedCount > 0 && <Button size="small" color="default" variant="filled" onClick={() => setEditingConfig(true)}>修改设置</Button>}
                    </div>
                    <ul className={styles.taskList} aria-label="申请进度">
                        {tasks.map((task, index) => {
                            const values = task.status === "success" ? task : form.getFieldValue(["items", index]) || task;
                            const verification = isIPAddress(task.domain) ? "http" : task.domain.startsWith("*.") ? "dns" : values.challenge || "http";
                            return <li key={task.domain}>
                                <div className="task-top">
                                    <div className="task-identity">
                                        <Tooltip title={task.domain}><span className="task-domain">{task.domain}</span></Tooltip>
                                        <div className="task-method">
                                            <span>{task.certificate_id ? "续签" : "申请"}</span>
                                            <span>{verification.toUpperCase()} · {Number(values.type) === 1 ? "手动" : "自动"}</span>
                                            <span>{task.certificate_id && !values.ca ? "沿用原环境" : values.ca === "staging" ? "测试环境" : "正式环境"}</span>
                                            {Number(values.type) !== 1 && verification === "dns" && <span>{secretOptions.find((option) => option.value === Number(values.secret_id))?.label || (values.secret_id ? `密钥 #${values.secret_id}` : "未选择密钥")}</span>}
                                        </div>
                                    </div>
                                    <span className={`task-status ${task.status}`}>{task.status === "running" && <Icon type="LoadingOutlined" spin/>}{task.status === "running" && manual?.index === index ? "验证中" : task.certificate_id && task.status === "pending" ? "待续签" : task.certificate_id && task.status === "running" ? "续签中" : statusNames[task.status]}</span>
                                    <span className="task-actions">
                                        {!busy && manual?.index === index && <Button size="small" color="default" variant="filled" onClick={() => verificationRef.current?.show()}>验证</Button>}
                                        {!configLocked && task.status === "failed" && <Button size="small" color="default" variant="filled" onClick={() => void runQueue(index)} aria-label={`重试 ${task.domain}`}>重试</Button>}
                                        {!configLocked && ["pending", "skipped"].includes(task.status) && <Button size="small" type="text" onClick={() => updateTask(index, {status: task.status === "skipped" ? "pending" : "skipped", error: undefined})}>{task.status === "skipped" ? "恢复" : "跳过"}</Button>}
                                    </span>
                                </div>
                                {task.error && <div className="task-error" role="alert">{task.error}</div>}
                            </li>;
                        })}
                    </ul>
                </>}
            </div>
        </ProModal>
        <BatchVerification ref={verificationRef}
            session={verificationSession}
            busy={busy}
            onVerify={() => void runQueue(false, true)}
            onRefresh={() => releaseManual()}
            onSkip={() => releaseManual(true)}/>
        <SecretManager ref={secretManagerRef} onChange={onSecretChange}/>
    </>;
});

BatchCert.displayName = "BatchCert";

export default memo(BatchCert);
