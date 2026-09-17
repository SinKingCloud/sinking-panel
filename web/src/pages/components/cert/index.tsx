import {forwardRef, memo, useCallback, useEffect, useImperativeHandle, useLayoutEffect, useMemo, useRef, useState} from "react";
import type {CSSProperties} from "react";
import {
    App,
    Button,
    Col,
    Descriptions,
    Empty,
    Flex,
    Form,
    Grid,
    Input,
    Row,
    Select,
    Spin,
    Tooltip,
    Typography,
} from "antd";
import {createStyles} from "antd-style";
import dayjs from "dayjs";
import {Icon, ProModal, Title, useTheme} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import ModalTable from "@/components/modal-table";
import type {ModalTableRef} from "@/components/modal-table";
import {
    createCert,
    deleteCert,
    getCertInfo,
    getCertList,
    getCertLog,
    obtainCert,
    requestCertManual,
    renewCert,
    updateCert,
} from "@/service/api/cert";
import type {ManualCertSession} from "@/service/api/cert";
import {getEnum as getSystemEnum} from "@/service/api/system";
import SecretManager, {SecretManagerRef} from "@/pages/components/secret";
import type {SecretChange} from "@/pages/components/secret";
import Log, {LogRef} from "@/pages/components/task-log";
import BatchCert, {BatchCertRef} from "./batch";

export interface CertManagerRef {
    open: () => void;
    close: () => void;
    openObtain: (domain: string | string[], onObtained?: (cert: any) => void) => void;
    openDeploy: (domain: string, onSelected: (cert: any) => void) => void;
    openRenew: (certificate: any, onRenewed?: (cert: any) => void, domain?: string) => void;
    refresh: () => Promise<void>;
}

export interface CertManagerProps {
    onChange?: () => void;
    secretVersion?: number;
}

const isIPAddress = (value: any) => {
    const address = String(value || "").trim().replace(/^\[([^\]]+)\]$/, "$1");
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

const parseCertificateTime = (value: any) => dayjs(String(value || ""));

const parseCertificate = (certificate: any) => {
    const type = Number(certificate?.type || 0);
    let domains: string[] = [];
    try {
        const values = Array.isArray(certificate?.domains) ? certificate.domains : JSON.parse(certificate?.domains || "[]");
        if (Array.isArray(values)) {
            domains = values.filter((value: any) => typeof value === "string" && value.trim());
        }
    } catch {}
    return {
        ...certificate,
        domains,
        type,
        challenge: type === 1 || type === 2 ? certificate?.challenge || "http" : "",
        secret_id: Number(certificate?.secret_id || 0),
        auto_renew: Number(certificate?.auto_renew || 0),
    };
};

const useStyles = createStyles(({css, token}: any, props: any = {}) => ({
    modal: css`
        .ant-modal { max-width: calc(100vw - 24px); outline: none; }
        .ant-modal-body { padding: 0 !important; overflow: hidden; }
    `,
    certificateTable: css`
        .record-name,
        .record-domains,
        .record-type,
        .record-secret {
            display: block;
            min-width: 0;
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            line-height: 20px;
        }

        .record-name { padding: 0; border: 0; background: transparent; color: inherit; font: inherit; text-align: left; }
        button.record-name { cursor: pointer; }
        button.record-name:hover, button.record-name:focus-visible { color: ${token.colorPrimary}; }
        .record-secret { color: ${token.colorTextSecondary}; }
        .record-issuance { display: flex; align-items: center; gap: 8px; min-width: 0; height: 20px; }
        .record-issuance .record-type { flex: none; }
        .record-issuance .record-secret { flex: 0 1 auto; }
        .record-renewal { flex: none; color: ${token.colorTextSecondary}; white-space: nowrap; }
        .record-domains { color: ${token.colorTextSecondary}; }
        .record-state { display: inline-flex; align-items: center; gap: 6px; line-height: 20px; white-space: nowrap; font-variant-numeric: tabular-nums; }
        .record-state i { width: 6px; height: 6px; flex: none; border-radius: 50%; background: currentColor; }
        .record-state.valid { color: ${token.colorSuccess}; }
        .record-state.expiring { color: ${token.colorWarning}; }
        .record-state.expired { color: ${token.colorError}; }
        .record-state.unknown { color: ${token.colorTextTertiary}; }
    `,
    loading: css`
        display: flex;
        flex: 1;
        min-height: var(--certificate-form-min-height, 180px);
        align-items: center;
        justify-content: center;
    `,
    formShell: css`
        box-sizing: border-box;
        width: 100%;
        min-height: var(--certificate-form-min-height, 0px);
        max-height: calc(100dvh - 290px);
        flex-shrink: 0;
        margin: 0 auto;
        padding: 4px 4px;
        scrollbar-gutter: stable;
        overflow-x: hidden;
        overflow-y: auto;
        overscroll-behavior: contain;

        @supports not (max-height: 100dvh) {
            max-height: calc(100vh - 320px);
        }

        @media (max-width: 767px) {
            max-height: calc(100dvh - 174px);

            @supports not (max-height: 100dvh) {
                max-height: calc(100vh - 174px);
            }
        }

    `,
    formSection: css`
        & + & {
            margin-top: ${props?.compact ? 15 : 18}px;
            padding-top: 0;
        }
        .ant-form-item:last-child { margin-bottom: 0; }
    `,
    formHints: css`
        .ant-form-item-extra,
        .ant-form-item-extra .ant-btn {
            font-size: ${token.fontSizeSM}px;
        }
    `,
    pemGrid: css`
        .ant-form-item { margin-bottom: 0; }
        @media (max-width: 767px) { .ant-col + .ant-col { margin-top: 14px; } }
    `,
    editor: css`
        &.ant-input {
            height: ${props?.compact ? 180 : 200}px;
            overflow-y: auto;
            resize: none;
        }
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
        font-size: 12px !important;
        line-height: 1.55;
        tab-size: 2;
    `,
    manualShell: css`
        width: 100%;
        overflow: hidden;
    `,
    manualContent: css`
        box-sizing: border-box;
        width: 100%;
        padding: 4px;
    `,
    verificationRecord: css`
        min-width: 0;
        & + & { margin-top: ${props?.compact ? 18 : 22}px; padding-top: ${props?.compact ? 18 : 22}px; border-top: 1px solid ${token.colorBorderSecondary}; }
        .verification-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: ${props?.compact ? 12 : 16}px; }
        .verification-heading-text { display: flex; align-items: center; flex-wrap: wrap; gap: 4px 12px; min-width: 0; line-height: 22px; }
        .verification-title { color: ${token.colorText}; font-size: ${token.fontSize}px; font-weight: 500; }
        .verification-expiry { color: ${token.colorTextSecondary}; font-size: 12px; white-space: nowrap; }
        .verification-heading > .ant-btn { flex: none; height: 22px; padding-inline: 6px; gap: 4px; font-size: 12px; }
        .verification-fields .ant-form-item { min-width: 0; margin-bottom: ${props?.compact ? 12 : 16}px; }
        .verification-fields .ant-form-item:last-child { margin-bottom: 0; }
        .verification-fields .ant-form-item-label { padding-bottom: ${props?.compact ? 4 : 6}px; }
        .verification-fields .ant-form-item-label > label { width: 100%; }
        .verification-field-label { display: flex; align-items: baseline; justify-content: space-between; flex-wrap: wrap; gap: 2px 12px; width: 100%; }
        .verification-field-label .verification-expiry { color: ${token.colorTextTertiary}; font-weight: 400; }
        .verification-type { margin-inline-start: 6px; color: ${token.colorTextSecondary}; font-size: 12px; font-weight: 400; }
        .verification-field-content { position: relative; min-width: 0; }
        .verification-fields textarea.ant-input { padding-inline-end: 32px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12px; line-height: 1.65; overflow-wrap: anywhere; resize: none; }
        .verification-copy { position: absolute; top: ${props?.compact ? 2 : 4}px; inset-inline-end: ${props?.compact ? 4 : 6}px; font-size: 12px; line-height: 22px; }
        .verification-copy .ant-typography-copy { display: inline-flex; align-items: center; justify-content: center; width: 22px; height: 22px; margin-inline-start: 0; }
    `,
    details: css`
        margin-bottom: ${props?.compact ? 12 : 16}px;
        padding: ${props?.compact ? "8px 10px" : "10px 12px"};
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};
        font-size: ${token.fontSizeSM}px;
    `,
    fieldLabel: css`
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        .ant-btn { height: 22px; padding: 0 4px; font-size: 12px; }
    `,
    pemField: css`
        .ant-form-item-label > label { width: 100%; }
    `,
}));

const CertManager = forwardRef<CertManagerRef, CertManagerProps>(({onChange, secretVersion = 0}, ref) => {
    const {message, modal} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {styles} = useStyles({compact});
    const modalRef = useRef<ModalTableRef | null>(null);
    const formModalRef = useRef<any>(null);
    const manualModalRef = useRef<any>(null);
    const manualContentRef = useRef<HTMLDivElement | null>(null);
    const batchRef = useRef<BatchCertRef | null>(null);
    const secretManagerRef = useRef<SecretManagerRef | null>(null);
    const logRef = useRef<LogRef | null>(null);
    const secretFormRequestRef = useRef<number | null>(null);
    const createdSecretRef = useRef<number | null>(null);
    const secretVersionRef = useRef(secretVersion);
    const listRequestRef = useRef(0);
    const formRequestRef = useRef(0);
    const secretRequestRef = useRef(0);
    const formOpenRef = useRef(false);
    const submittingRef = useRef(false);
    const manualSessionRef = useRef<{data: ManualCertSession, method: "obtain" | "renew", id?: number} | null>(null);
    const obtainRef = useRef<{domain: string, onObtained?: (cert: any) => void} | null>(null);
    const deployRef = useRef<{domain: string, onSelected: (cert: any) => void} | null>(null);
    const suggestedNameRef = useRef("");
    const [form] = Form.useForm();
    const [active, setActive] = useState(false);
    const [loading, setLoading] = useState(false);
    const [listError, setListError] = useState(false);
    const [deployDomain, setDeployDomain] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [formLoading, setFormLoading] = useState(false);
    const [formOpen, setFormOpen] = useState(false);
    const [mode, setMode] = useState("import");
    const viewing = mode === "view";
    const certificateForm = viewing || mode === "import" || mode === "edit";
    const [obtainDomain, setObtainDomain] = useState("");
    const [challenge, setChallenge] = useState("http");
    const [manualSession, setManualSession] = useState<ManualCertSession | null>(null);
    const [secretOptions, setSecretOptions] = useState<{label: string, value: number}[]>([]);
    const [secretLoading, setSecretLoading] = useState(false);
    const [secretLoadError, setSecretLoadError] = useState(false);
    const [wildcard, setWildcard] = useState(false);
    const [ipAddress, setIPAddress] = useState(false);
    const [record, setRecord] = useState<any>();
    const importedCertificate = record?.type === 0;
    const formMinHeight = (mode === "import" || mode === "obtain" || (mode === "edit" && importedCertificate) ? 440 : viewing ? (importedCertificate ? 480 : 520) : 600) - (compact ? 40 : 0);
    const [items, setItems] = useState<any[]>([]);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [keyword, setKeyword] = useState("");
    const [searchField, setSearchField] = useState("keyword");
    const [appliedKeyword, setAppliedKeyword] = useState("");
    const [appliedField, setAppliedField] = useState("keyword");
    const selectedSecretId = Form.useWatch("secret_id", form);
    const certificateType = Number(Form.useWatch("type", form) ?? 2);
    const automaticDNS = challenge === "dns" && certificateType === 2;
    const renewalNeedsKey = automaticDNS && !selectedSecretId;
    const manualVerification = certificateType === 1;
    const secretNames = useMemo(() => Object.fromEntries(secretOptions.map(({value, label}) => [value, label])), [secretOptions]);

    useLayoutEffect(() => {
        const content = manualContentRef.current;
        const shell = content?.parentElement;
        if (!manualSession || !content || !shell) return;
        const resize = () => {
            const height = `${content.offsetHeight}px`;
            if (shell.style.height !== height) shell.style.height = height;
        };
        resize();
        const observer = new ResizeObserver(resize);
        observer.observe(content);
        return () => observer.disconnect();
    }, [manualSession]);

    const loadList = useCallback(async (nextPage = page, nextPageSize = pageSize, nextKeyword = appliedKeyword, nextField = appliedField) => {
        const requestId = ++listRequestRef.current;
        setAppliedKeyword(nextKeyword.trim());
        setAppliedField(nextField);
        setLoading(true);
        setListError(false);
        try {
            const response = await getCertList({
                body: {
                    page: nextPage,
                    page_size: nextPageSize,
                    ...(nextKeyword.trim() ? {[nextField]: nextKeyword.trim()} : {}),
                    order_by_field: "id",
                    order_by_type: "desc",
                    ...(deployRef.current ? {domain: deployRef.current.domain, usable: 1} : {}),
                },
            });
            if (listRequestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                setListError(true);
                if (deployRef.current) {
                    setItems([]);
                    setTotal(0);
                }
                message.error(response?.message || "获取证书列表失败");
                return;
            }
            const data: any = response?.data || {};
            const nextItems = Array.isArray(data.list) ? data.list.map(parseCertificate) : [];
            const nextTotal = Number(data.total || 0);
            setItems(nextItems);
            setTotal(nextTotal);
            setPage(Number(data.page || nextPage));
            setPageSize(Number(data.page_size || nextPageSize));
        } catch {
            if (listRequestRef.current === requestId) {
                setListError(true);
                if (deployRef.current) {
                    setItems([]);
                    setTotal(0);
                }
                message.error("获取证书列表失败");
            }
        } finally {
            if (listRequestRef.current === requestId) {
                setLoading(false);
            }
        }
    }, [appliedField, appliedKeyword, message, page, pageSize]);

    const loadSecrets = useCallback(async () => {
        const requestId = ++secretRequestRef.current;
        setSecretLoading(true);
        setSecretLoadError(false);
        try {
            const response = await getSystemEnum({body: {name: "secret"}});
            if (secretRequestRef.current !== requestId) return;
            if (response?.code !== 200 || !response.data?.name) {
                setSecretLoadError(true);
                return;
            }
            const options = Object.entries(response?.data?.name || {}).map(([value, label]) => ({
                value: Number(value),
                label: String(label),
            }));
            setSecretOptions(options);
            if (!formOpenRef.current) return;
            const selected = form.getFieldValue("secret_id");
            if (selected && !options.some((option) => option.value === selected)) {
                form.setFieldsValue({secret_id: undefined, auto_renew: 0});
            }
        } catch {
            if (secretRequestRef.current === requestId) {
                setSecretLoadError(true);
            }
        } finally {
            if (secretRequestRef.current === requestId) setSecretLoading(false);
        }
    }, [form]);

    useEffect(() => {
        if (secretVersionRef.current === secretVersion) return;
        secretVersionRef.current = secretVersion;
        if (active || (formOpenRef.current && mode !== "import")) void loadSecrets();
        if (active) void loadList();
    }, [active, loadList, loadSecrets, mode, secretVersion]);

    const cancelManual = useCallback(() => {
        const session = manualSessionRef.current;
        manualSessionRef.current = null;
        if (session) {
            void requestCertManual(session.method, {action: "cancel", session_id: session.data.session_id, id: session.id}).catch(() => undefined);
        }
    }, []);

    const closeForm = useCallback((backToForm = false) => {
        const returnToForm = backToForm && formOpenRef.current;
        formRequestRef.current += 1;
        cancelManual();
        submittingRef.current = false;
        setSubmitting(false);
        setFormLoading(false);
        manualModalRef.current?.hide();
        if (returnToForm) {
            formModalRef.current?.show();
            return;
        }
        secretFormRequestRef.current = null;
        createdSecretRef.current = null;
        secretRequestRef.current += 1;
        formOpenRef.current = false;
        obtainRef.current = null;
        suggestedNameRef.current = "";
        formModalRef.current?.hide();
        setFormOpen(false);
        setSecretLoading(false);
    }, [cancelManual]);

    const openSecretManager = useCallback((create = false) => {
        secretFormRequestRef.current = formOpenRef.current ? formRequestRef.current : null;
        createdSecretRef.current = null;
        if (create) secretManagerRef.current?.openCreate();
        else secretManagerRef.current?.open();
    }, []);

    const openForm = useCallback((nextMode: string, nextRecord?: any) => {
        batchRef.current?.close();
        formRequestRef.current += 1;
        cancelManual();
        manualModalRef.current?.hide();
        secretRequestRef.current += 1;
        formOpenRef.current = true;
        submittingRef.current = false;
        obtainRef.current = null;
        setMode(nextMode);
        setObtainDomain("");
        setRecord(nextRecord);
        setFormLoading(false);
        setSubmitting(false);
        form.resetFields();
        suggestedNameRef.current = "";
        setChallenge("http");
        setSecretLoading(false);
        setSecretLoadError(false);
        setWildcard(false);
        setIPAddress(false);
        if (nextMode === "obtain") {
            const domain = deployRef.current?.domain || "";
            const isWildcard = domain.startsWith("*.");
            const nextChallenge = isWildcard ? "dns" : "http";
            setObtainDomain(domain);
            setWildcard(isWildcard);
            setIPAddress(isIPAddress(domain));
            setChallenge(nextChallenge);
            suggestedNameRef.current = domain.slice(0, 100);
            form.setFieldsValue({domain, name: suggestedNameRef.current, ca: "production", challenge: nextChallenge, type: 2, auto_renew: 1});
        } else if (nextMode === "import" && deployRef.current) {
            form.setFieldValue("name", deployRef.current.domain.slice(0, 100));
        } else if (nextMode === "renew") {
            const domain = String(nextRecord?.domains?.[0] || "");
            const isWildcard = domain.trim().startsWith("*.");
            const isIP = isIPAddress(domain);
            const currentChallenge = String(nextRecord?.challenge || "http");
            const nextType = Number(nextRecord?.type ?? 2);
            const nextChallenge = isIP ? "http" : isWildcard ? "dns" : currentChallenge;
            const secretId = Number(nextRecord?.secret_id || 0);
            setChallenge(nextChallenge);
            setWildcard(isWildcard);
            setIPAddress(isIP);
            form.setFieldsValue({
                domain,
                challenge: nextChallenge,
                type: nextType,
                secret_id: nextType === 2 && nextChallenge === "dns" && secretId ? secretId : undefined,
                auto_renew: nextType !== 2 || (nextChallenge === "dns" && !secretId) ? 0 : Number(nextRecord?.auto_renew || 0),
            });
        }
        setFormOpen(true);
        formModalRef.current?.show();
        if (nextMode === "obtain" || nextMode === "renew") {
            void loadSecrets();
        }
        if (nextMode !== "edit" && nextMode !== "view") {
            return;
        }
        const requestId = ++formRequestRef.current;
        setFormLoading(true);
        void getCertInfo({body: {id: nextRecord?.id}}).then((response) => {
            if (formRequestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "获取证书详情失败");
                closeForm();
                return;
            }
            const data = parseCertificate(response?.data);
            setRecord(data);
            form.setFieldsValue({
                name: data.name,
                certificate: data.certificate,
                private_key: data.private_key,
                challenge: data.challenge,
                type: data.type,
                secret_id: data.type === 2 && data.challenge === "dns" ? data.secret_id || undefined : undefined,
                auto_renew: data.type !== 2 || (data.challenge === "dns" && !data.secret_id) ? 0 : data.auto_renew,
            });
            setChallenge(data.challenge);
            setWildcard(data.domains.some((domain: string) => domain.startsWith("*.")));
            setIPAddress(data.domains.some(isIPAddress));
            if (data.type === 1 || data.type === 2) void loadSecrets();
        }).catch(() => {
            if (formRequestRef.current === requestId) {
                message.error("获取证书详情失败");
                closeForm();
            }
        }).finally(() => {
            if (formRequestRef.current === requestId) {
                setFormLoading(false);
            }
        });
    }, [cancelManual, closeForm, form, loadSecrets, message]);

    const open = useCallback((selection?: {domain: string, onSelected: (cert: any) => void}) => {
        batchRef.current?.close();
        closeForm();
        deployRef.current = selection || null;
        setDeployDomain(selection?.domain || "");
        listRequestRef.current += 1;
        setActive(true);
        setLoading(true);
        setListError(false);
        setItems([]);
        setTotal(0);
        setPage(1);
        setKeyword("");
        setSearchField("keyword");
        setAppliedKeyword("");
        setAppliedField("keyword");
        modalRef.current?.show();
        void loadSecrets();
        if (active) void loadList(1, pageSize, "", "keyword");
    }, [active, closeForm, loadList, loadSecrets, pageSize]);

    const close = useCallback(() => {
        batchRef.current?.close();
        secretManagerRef.current?.close();
        closeForm();
        deployRef.current = null;
        setDeployDomain("");
        listRequestRef.current += 1;
        setActive(false);
        setLoading(false);
        setListError(false);
        setItems([]);
        setTotal(0);
        modalRef.current?.hide();
    }, [closeForm]);

    const openObtain = useCallback((domain: string | string[], onObtained?: (cert: any) => void) => {
        deployRef.current = null;
        setDeployDomain("");
        listRequestRef.current += 1;
        setActive(false);
        setLoading(false);
        modalRef.current?.hide();
        if (Array.isArray(domain)) {
            closeForm();
            batchRef.current?.open(domain, onObtained);
            return;
        }
        const nextDomain = domain.trim();
        if (!nextDomain) return;
        openForm("obtain");
        const name = nextDomain.slice(0, 100);
        const isWildcard = nextDomain.startsWith("*.");
        obtainRef.current = {domain: nextDomain, onObtained};
        suggestedNameRef.current = name;
        setObtainDomain(nextDomain);
        setWildcard(isWildcard);
        setIPAddress(isIPAddress(nextDomain));
        setChallenge(isWildcard ? "dns" : "http");
        form.setFieldsValue({domain: nextDomain, name, challenge: isWildcard ? "dns" : "http"});
    }, [closeForm, form, openForm]);

    const openDeploy = useCallback((domain: string, onSelected: (cert: any) => void) => {
        const target = domain.trim();
        if (!target) return;
        open({domain: target, onSelected});
    }, [open]);

    const openRenew = useCallback(async (certificate: any, onRenewed?: (cert: any) => void, domain?: string) => {
        close();
        let data = parseCertificate(certificate);
        if (!certificate?.domains) {
            openForm("renew");
            setFormLoading(true);
            const requestId = ++formRequestRef.current;
            try {
                const response = await getCertInfo({body: {id: certificate?.id}});
                if (formRequestRef.current !== requestId || !formOpenRef.current) return;
                if (response?.code !== 200) throw new Error(response?.message || "获取证书详情失败");
                data = parseCertificate(response.data);
            } catch (error) {
                if (formRequestRef.current === requestId) {
                    message.error(error instanceof Error ? error.message : "获取证书详情失败");
                    closeForm();
                }
                return;
            }
        }
        if (![1, 2].includes(data.type)) {
            message.info("导入证书无法直接续签，请申请或导入新证书。");
            openDeploy(domain || data.domains[0] || "", onRenewed || (() => undefined));
            return;
        }
        openForm("renew", data);
        obtainRef.current = {domain: "", onObtained: onRenewed};
    }, [close, closeForm, message, openDeploy, openForm]);

    useImperativeHandle(ref, () => ({open, close, openObtain, openDeploy, openRenew, refresh: loadList}), [close, loadList, open, openObtain, openDeploy, openRenew]);

    useEffect(() => () => {
        listRequestRef.current += 1;
        formRequestRef.current += 1;
        secretRequestRef.current += 1;
        formOpenRef.current = false;
        obtainRef.current = null;
        deployRef.current = null;
        cancelManual();
    }, [cancelManual]);

    const submit = useCallback(async (values: any) => {
        if (!formOpenRef.current || submittingRef.current || formLoading || mode === "view") {
            return;
        }
        const requestId = ++formRequestRef.current;
        const directObtain = obtainRef.current;
        const directDeploy = deployRef.current;
        submittingRef.current = true;
        setSubmitting(true);
        try {
            let response: any;
            if (mode === "import") {
                response = await createCert({body: values});
            } else if (mode === "edit") {
                const body: Record<string, any> = {id: record?.id};
                for (const field of ["name", "certificate", "private_key"]) {
                    if (values[field] !== record?.[field]) body[field] = values[field];
                }
                if (record?.type === 1 || record?.type === 2) {
                    const settings = {
                        challenge: values.challenge,
                        type: Number(values.type),
                        secret_id: Number(values.type) === 2 && values.challenge === "dns" ? Number(values.secret_id || 0) : 0,
                        auto_renew: Number(values.type) !== 2 || (values.challenge === "dns" && !values.secret_id) ? 0 : Number(values.auto_renew || 0),
                    };
                    for (const [field, value] of Object.entries(settings)) {
                        if (value !== record[field]) body[field] = String(value);
                    }
                }
                if (Object.keys(body).length === 1) {
                    closeForm();
                    message.info("没有需要保存的修改");
                    return;
                }
                response = await updateCert({body});
            } else {
                const target = String(directObtain?.domain || (mode === "obtain" ? directDeploy?.domain : "") || values.domain || "").trim();
                const isIP = isIPAddress(target);
                const manual = Number(values.type) === 1;
                const verification = isIP ? "http" : target.startsWith("*.") ? "dns" : values.challenge || "http";
                const body = {
                    name: values.name,
                    domain: isIP ? target.replace(/^\[([^\]]+)\]$/, "$1") : target,
                    ...(values.ca ? {ca: values.ca} : {}),
                    challenge: verification,
                    type: manual ? 1 : 2,
                    secret_id: !manual && verification === "dns" ? Number(values.secret_id) : 0,
                    auto_renew: manual ? 0 : Number(values.auto_renew || 0),
                };
                if (manual) {
                    let session = manualSessionRef.current;
                    if (!session) {
                        const method = mode === "renew" ? "renew" : "obtain";
                        const id = method === "renew" ? record?.id : undefined;
                        const started = await requestCertManual(method, {...body, action: "start", id});
                        const data = started?.data as ManualCertSession | undefined;
                        if (formRequestRef.current !== requestId) {
                            if (data?.session_id) void requestCertManual(method, {action: "cancel", session_id: data.session_id, id}).catch(() => undefined);
                            return;
                        }
                        if (started?.code !== 200) throw new Error(started?.message || "获取验证信息失败");
                        if (!data?.session_id || !Array.isArray(data.challenges) || !Number.isFinite(data.expires_at)) {
                            if (data?.session_id) void requestCertManual(method, {action: "cancel", session_id: data.session_id, id}).catch(() => undefined);
                            throw new Error("验证信息无效，请重新申请");
                        }
                        session = {data, method, id};
                        manualSessionRef.current = session;
                        setManualSession(data);
                        formModalRef.current?.hide();
                        manualModalRef.current?.show();
                        if (data.challenges.length > 0) return;
                    }
                    if (session.data.expires_at <= Math.floor(Date.now() / 1000)) {
                        cancelManual();
                        manualModalRef.current?.hide();
                        formModalRef.current?.show();
                        throw new Error("验证会话已过期，请重新申请");
                    }
                    response = await requestCertManual(session.method, {
                        action: "submit",
                        session_id: session.data.session_id,
                        id: session.id,
                    });
                } else {
                    response = mode === "obtain" ? await obtainCert({body}) : await renewCert({body: {...body, id: record?.id}});
                }
            }
            if (formRequestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.error(response?.message || "操作失败");
                return;
            }
            manualSessionRef.current = null;
            obtainRef.current = null;
            formOpenRef.current = false;
            setFormOpen(false);
            formModalRef.current?.hide();
            manualModalRef.current?.hide();
            // 通知调用方失败不会改变证书已申请成功的结果。
            try {
                void Promise.resolve(onChange?.()).catch(() => undefined);
            } catch {}
            if (formRequestRef.current !== requestId) return;
            if (directDeploy && deployRef.current === directDeploy && (mode === "import" || mode === "obtain")) {
                const certificate = parseCertificate(response?.data);
                const [domain, ...domains] = [directDeploy.domain, ...certificate.domains].map((value) => {
                    const host = String(value).trim().toLowerCase().replace(/\.$/, "").replace(/^\[([^\]]+)\]$/, "$1");
                    try {
                        return new URL(`http://${host.includes(":") ? `[${host}]` : host}`).hostname.replace(/^\[([^\]]+)\]$/, "$1");
                    } catch {
                        return host;
                    }
                });
                const matches = domains.some((value) => value === domain || (value.startsWith("*.") && !domain.startsWith("*.") && !isIPAddress(domain) && domain.split(".").length === value.split(".").length && domain.endsWith(value.slice(1))));
                const now = dayjs();
                const starts = parseCertificateTime(certificate.start_time);
                const expires = parseCertificateTime(certificate.expire_time);
                if (matches && starts.isValid() && expires.isValid() && !starts.isAfter(now) && expires.isAfter(now)) {
                    message.success(response?.message || "操作成功");
                    close();
                    try {
                        void Promise.resolve(directDeploy.onSelected(certificate)).catch(() => undefined);
                    } catch {}
                    return;
                }
                message.warning(matches ? "证书已保存，但不在有效期内，请选择其他证书。" : "证书已保存，但不适用于当前域名，请选择其他证书。");
            } else {
                message.success(response?.message || "操作成功");
            }
            if (active) {
                void loadList(page, pageSize, appliedKeyword);
            }
            try {
                if (formRequestRef.current === requestId) {
                    void Promise.resolve(directObtain?.onObtained?.(parseCertificate(response?.data))).catch(() => undefined);
                }
            } catch {}
        } catch (error) {
            if (formRequestRef.current === requestId) {
                message.error(error instanceof Error ? error.message : "操作失败");
            }
        } finally {
            if (formRequestRef.current === requestId) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    }, [active, appliedKeyword, cancelManual, close, closeForm, formLoading, loadList, message, mode, onChange, page, pageSize, record]);

    const remove = useCallback(async (item: any) => {
        try {
            const response = await deleteCert({body: {id: item.id}});
            if (response?.code !== 200) {
                message.error(response?.message || "删除证书失败");
                return;
            }
            message.success(response?.message || "删除成功");
            onChange?.();
            const nextPage = items.length === 1 && page > 1 ? page - 1 : page;
            void loadList(nextPage, pageSize, appliedKeyword);
        } catch {
            message.error("删除证书失败");
        }
    }, [appliedKeyword, items.length, loadList, message, onChange, page, pageSize]);

    const confirmRemove = useCallback((item: any) => {
        modal.confirm({
            title: "删除证书",
            content: `确定删除证书“${item?.name || "未命名证书"}”吗？被网站使用的证书无法删除。`,
            zIndex: 1040,
            okText: "删除",
            cancelText: "取消",
            okButtonProps: {danger: true},
            mask: {closable: true},
            onOk: () => remove(item),
        } as any);
    }, [modal, remove]);

    const renderActions = useCallback((item: any) => deployDomain ? (
            <Button size="small" autoInsertSpace={false} onClick={() => {
                const selection = deployRef.current;
                if (!selection) return;
                const now = dayjs();
                const starts = parseCertificateTime(item.start_time);
                const expires = parseCertificateTime(item.expire_time);
                if (!starts.isValid() || starts.isAfter(now) || !expires.isValid() || !expires.isAfter(now)) {
                    message.warning("证书不在有效期内，请选择其他证书。");
                    void loadList();
                    return;
                }
                close();
                try {
                    void Promise.resolve(selection.onSelected(item)).catch(() => undefined);
                } catch {}
            }}>部署</Button>
    ) : (
            <Dropdown
                trigger={["click"]}
                placement="bottom"
                arrow
                menu={{
                    items: [
                        {key: "view", label: "查看"},
                        {key: "edit", label: "编辑"},
                        {key: "renew", label: "续签", disabled: ![1, 2].includes(Number(item.type))},
                        {type: "divider" as const},
                        {key: "delete", label: "删除", danger: true},
                    ],
                    onClick: ({key}) => {
                        if (key === "view" || key === "edit") {
                            openForm(key, item);
                        } else if (key === "renew" && [1, 2].includes(Number(item.type))) {
                            openForm("renew", item);
                        } else if (key === "delete") {
                            confirmRemove(item);
                        }
                    },
                }}>
                <Button
                    size="small"
                    autoInsertSpace={false}
                    aria-label={`操作证书 ${item?.name || ""}`}>操作</Button>
            </Dropdown>
    ), [close, confirmRemove, deployDomain, loadList, message, openForm]);

    const columns = useMemo<any[]>(() => [
        {
            title: "证书名称",
            dataIndex: "name",
            key: "name",
            width: 160,
            render: (value: string, item: any) => deployDomain ? <span className="record-name" title={value}>{value || "未命名证书"}</span> : (
                <button type="button" className="record-name" title={value} onClick={() => openForm("view", item)}>
                    {value || "未命名证书"}
                </button>
            ),
        },
        {
            title: "适用域名",
            dataIndex: "domains",
            key: "domains",
            width: 190,
            render: (domains: string[]) => {
                const text = domains.join(", ") || "—";
                return <Tooltip title={text}><span className="record-domains">{text}</span></Tooltip>;
            },
        },
        {
            title: "签发信息",
            key: "issuance",
            width: 220,
            render: (_: any, item: any) => {
                const secretName = item.secret_id ? secretNames[item.secret_id] || `密钥 #${item.secret_id}` : "未关联密钥";
                return (
                    <div className="record-issuance">
                        <span className="record-type">{item.type === 2 ? "自动" : item.type === 1 ? "手动" : "导入证书"}</span>
                        {(item.type === 1 || item.type === 2) && <span className="record-type">{item.challenge === "dns" ? "DNS" : "HTTP"}</span>}
                        {item.type === 2 && item.challenge === "dns" && (
                            <Tooltip title={item.secret_id ? `关联密钥：${secretName}` : secretName}>
                                <span className="record-secret">{secretName}</span>
                            </Tooltip>
                        )}
                        {item.type === 2 && <span className="record-renewal">
                            {item.auto_renew === 1 ? "自动续签" : "手动续签"}
                        </span>}
                    </div>
                );
            },
        },
        {
            title: "有效期",
            key: "expiry",
            width: 145,
            render: (_: any, item: any) => {
                const now = dayjs();
                const expires = parseCertificateTime(item.expire_time);
                const starts = parseCertificateTime(item.start_time);
                const knownExpiry = expires.isValid();
                const expired = knownExpiry && !expires.isAfter(now);
                const pending = starts.isValid() && starts.isAfter(now);
                const days = knownExpiry ? expires.diff(now, "day") : 0;
                const status = !knownExpiry ? "unknown" : expired ? "expired" : pending || days <= 7 ? "expiring" : "valid";
                const text = !knownExpiry ? "有效期未知" : expired ? "已过期" : pending ? "尚未生效"
                    : days > 0 ? `剩余 ${days} 天` : "不足 1 天";
                return <Tooltip title={knownExpiry ? `${expires.format("YYYY-MM-DD HH:mm")} 到期` : undefined}>
                    <span className={`record-state ${status}`}><i/>{text}</span>
                </Tooltip>;
            },
        },
        {
            title: "操作",
            key: "action",
            width: 82,
            fixed: "right",
            className: "action-cell",
            render: (_: any, item: any) => renderActions(item),
        },
    ], [deployDomain, openForm, renderActions, secretNames]);

    const download = useCallback((field: string) => {
        const content = String(form.getFieldValue(field) || "");
        if (!content.trim()) {
            message.warning(field === "certificate" ? "证书内容为空" : "私钥内容为空");
            return;
        }
        const url = URL.createObjectURL(new Blob([content], {type: "application/x-pem-file"}));
        const anchor = document.createElement("a");
        anchor.href = url;
        anchor.download = field === "certificate" ? "fullchain.pem" : "private.key";
        anchor.click();
        window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    }, [form, message]);

    const onSecretChange = useCallback((change: SecretChange) => {
        onChange?.();
        if (active) void loadList();
        if (active || (formOpenRef.current && mode !== "import")) {
            if (secretFormRequestRef.current === formRequestRef.current && change.action === "create") {
                createdSecretRef.current = change.id;
            }
            const selectCreated = change.action === "update" && createdSecretRef.current === change.id && secretFormRequestRef.current === formRequestRef.current && form.getFieldValue("challenge") === "dns" && Number(form.getFieldValue("type")) === 2;
            if (selectCreated) form.setFieldValue("secret_id", change.id);
            if (selectCreated || (change.action === "delete" && createdSecretRef.current === change.id)) createdSecretRef.current = null;
            void loadSecrets();
        }
    }, [active, form, loadList, loadSecrets, mode, onChange]);

    const formTitle = {
        view: "查看证书",
        import: "导入证书",
        edit: "编辑证书",
        obtain: "申请证书",
        renew: "续签证书",
    }[mode] || "证书配置";

    const verificationField = (
        <Form.Item extra={mode === "edit" ? "用于下次续签验证，不改变当前证书。" : manualVerification ? challenge === "http" ? "自行将验证文件放到目标网站，公网需能通过 80 端口访问。" : "自行添加 TXT 解析记录，无需提供云密钥。" : ipAddress ? "公网 IP 的 80 端口需能访问本机验证服务。" : challenge === "dns" ? "自动添加 DNS 验证记录，网站无需上线。" : "域名需解析到本机，公网可访问验证端口（默认 80）。"}>
            <Flex gap={compact ? 12 : 16}>
                <Form.Item name="challenge" label="验证类型" style={{flex: 1, minWidth: 0, marginBottom: 0}} rules={[{required: true, message: "请选择验证类型"}]}>
                    <Select options={[
                        {label: "HTTP", value: "http", disabled: wildcard},
                        {label: "DNS", value: "dns", disabled: ipAddress},
                    ]}/>
                </Form.Item>
                <Form.Item name="type" label="验证模式" style={{flex: 1, minWidth: 0, marginBottom: 0}} rules={[{required: true, message: "请选择验证模式"}]}>
                    <Select options={[
                        {label: "自动", value: 2},
                        {label: "手动", value: 1},
                    ]}/>
                </Form.Item>
            </Flex>
        </Form.Item>
    );
    const secretField = (
        <Form.Item
            preserve={false}
            name="secret_id"
            label="关联密钥"
            extra={<>
                {secretLoadError ? <span>密钥列表加载失败，<Button type="link" size="small" htmlType="button" onClick={() => void loadSecrets()}>重试</Button></span> : !secretLoading && secretOptions.length === 0 ? <span>暂无密钥，请先<Button type="link" size="small" htmlType="button" onClick={() => openSecretManager(true)}>添加密钥</Button>并填写密钥信息。</span> : mode === "edit" ? "DNS 续签前需关联密钥。" : "请选择具有该域名 DNS 解析权限的密钥。"}
                <Button type="link" size="small" htmlType="button" onClick={() => openSecretManager()}>管理密钥</Button>
            </>}
            rules={mode === "edit" ? [] : [{required: true, message: "请选择云密钥"}]}>
            <Select
                allowClear={mode === "edit"}
                loading={secretLoading}
                disabled={secretLoading}
                showSearch={false}
                placeholder={mode === "edit" ? "暂不关联密钥" : "请选择云密钥"}
                options={secretOptions}
                notFoundContent={secretLoading ? <Spin size="small"/> : "暂无可用密钥"}/>
        </Form.Item>
    );
    const renewalField = (
        <Form.Item name="auto_renew" label="自动续签" rules={[{required: true, message: "请选择是否自动续签"}]}>
            <Select options={[
                {label: "关闭", value: 0},
                {label: "开启", value: 1, disabled: mode !== "obtain" && renewalNeedsKey},
            ]}/>
        </Form.Item>
    );

    return (
        <>
            <ModalTable
                ref={modalRef}
                title={<Title>{deployDomain ? "部署证书" : "证书管理"}</Title>}
                width={910}
                onCancel={close}
                className={styles.certificateTable}
                minHeight={compact ? 440 : 480}
                scrollMode="modal"
                columns={columns}
                dataSource={items}
                loading={loading}
                emptyContent={listError ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="获取证书列表失败">
                    <Button color="default" variant="filled" size="small" onClick={() => void loadList()}>重试</Button>
                </Empty> : undefined}
                onReload={loadList}
                rowKey={(item) => String(item.id)}
                scroll={{x: 907}}
                toolbar={{
                    search: {
                        value: keyword,
                        field: searchField,
                        maxLength: 100,
                        fields: [
                            {value: "keyword", label: "全部", placeholder: "搜索名称或域名"},
                            {value: "name", label: "名称", placeholder: "搜索证书名称"},
                            {value: "domains", label: "域名", placeholder: "搜索证书域名", maxLength: 253},
                        ],
                        ariaLabel: "搜索证书",
                        debounce: active && !formOpen ? 300 : undefined,
                        onChange: setKeyword,
                        onFieldChange: setSearchField,
                        onSearch: (value, field) => void loadList(1, pageSize, value, field),
                    },
                    refresh: {
                        loading,
                        disabled: loading,
                        ariaLabel: "刷新证书列表",
                        tooltip: "刷新",
                        onClick: () => void loadList(),
                    },
                    actions: [
                        {key: "import", type: "button", label: "导入", icon: "ImportOutlined", onClick: () => openForm("import")},
                        {key: "obtain", type: "button", label: "申请", icon: "PlusOutlined", onClick: () => openForm("obtain")},
                        {key: "secret", type: "button", label: "密钥", icon: "KeyOutlined", onClick: () => openSecretManager()},
                        {key: "log", type: "button", label: "日志", icon: "FileTextOutlined", hidden: Boolean(deployDomain), onClick: () => logRef.current?.open({id: "cert-auto-renew", name: "证书自动续签"})},
                    ],
                }}
                pagination={{
                    page,
                    pageSize,
                    total,
                    unit: "张",
                    disabled: loading,
                    onChange: (nextPage, nextPageSize) => void loadList(nextPageSize === pageSize ? nextPage : 1, nextPageSize, appliedKeyword),
                }}
                modalProps={{
                    zIndex: 1010,
                    footer: null,
                    closable: true,
                    keyboard: true,
                    mask: {closable: true},
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24, maxWidth: "calc(100vw - 24px)"},
                    focusable: {focusTriggerAfterClose: false},
                } as any}/>
            <ProModal
                ref={formModalRef}
                title={<Title>{formTitle}</Title>}
                width={mode === "import" || mode === "edit" || (viewing && importedCertificate) ? (compact ? 540 : 580) : viewing ? (compact ? 650 : 700) : (compact ? 460 : 500)}
                onOk={viewing ? () => closeForm() : form.submit}
                onCancel={() => closeForm()}
                modalProps={{
                    rootClassName: styles.modal,
                    zIndex: 1030,
                    styles: certificateForm || mode === "obtain" ? {
                        container: {minHeight: formMinHeight, display: "flex", flexDirection: "column"},
                        body: {flex: 1, display: "flex", flexDirection: "column"},
                    } : undefined,
                    footer: formLoading ? null : viewing ? [<Button key="close" onClick={() => closeForm()}>关闭</Button>] : [
                        <Button key="cancel" onClick={() => closeForm()}>取消</Button>,
                        <Button key="submit" type="primary" loading={submitting} disabled={(mode === "obtain" || mode === "renew") && automaticDNS && (secretLoading || secretLoadError)} onClick={form.submit}>
                            {mode === "import" ? "导入" : mode === "obtain" ? "申请" : mode === "renew" ? "续签" : "保存"}
                        </Button>,
                    ],
                    forceRender: true,
                    closable: true,
                    keyboard: true,
                    mask: {closable: true},
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                    focusable: {focusTriggerAfterClose: false},
                    afterClose: () => {
                        if (!formOpenRef.current) {
                            form.resetFields();
                            setRecord(undefined);
                        }
                    },
                } as any}>
                <Form
                    className={styles.formHints}
                    style={certificateForm || mode === "obtain" ? {
                        "--certificate-form-min-height": `${formMinHeight - 120}px`,
                        flex: 1,
                        display: "flex",
                        flexDirection: "column",
                    } as CSSProperties : undefined}
                    form={form}
                    layout="vertical"
                    variant="filled"
                    disabled={submitting}
                    onFinish={submit}
                    onValuesChange={(changed: any) => {
                        if (mode === "import" || viewing) return;
                        if (mode === "obtain" && Object.prototype.hasOwnProperty.call(changed, "name")) {
                            suggestedNameRef.current = "";
                        }
                        if (Object.prototype.hasOwnProperty.call(changed, "domain")) {
                            const domain = String(changed.domain || "").trim();
                            const isWildcard = domain.startsWith("*.");
                            const isIP = isIPAddress(domain);
                            const currentChallenge = String(form.getFieldValue("challenge") || "http");
                            const nextChallenge = isIP ? "http" : isWildcard ? "dns" : currentChallenge;
                            setWildcard(isWildcard);
                            setIPAddress(isIP);
                            setChallenge(nextChallenge);
                            form.setFieldValue("challenge", nextChallenge);
                            if (mode === "obtain") {
                                const currentName = String(form.getFieldValue("name") || "");
                                if (!currentName || currentName === suggestedNameRef.current) {
                                    suggestedNameRef.current = domain.slice(0, 100);
                                    form.setFieldValue("name", suggestedNameRef.current);
                                }
                            }
                        }
                        if (Object.prototype.hasOwnProperty.call(changed, "challenge")) {
                            const currentChallenge = String(changed.challenge || "http");
                            const nextChallenge = ipAddress ? "http" : wildcard ? "dns" : currentChallenge;
                            setChallenge(nextChallenge);
                            form.setFieldValue("challenge", nextChallenge);
                        }
                        if (["domain", "challenge", "type", "secret_id"].some((field) => Object.prototype.hasOwnProperty.call(changed, field))) {
                            const automatic = Number(form.getFieldValue("type")) === 2;
                            const nextChallenge = form.getFieldValue("challenge");
                            if (!automatic || nextChallenge !== "dns") form.setFieldValue("secret_id", undefined);
                            if (!automatic || (mode !== "obtain" && nextChallenge === "dns" && !form.getFieldValue("secret_id"))) {
                                form.setFieldValue("auto_renew", 0);
                            } else if (mode === "obtain" && Object.prototype.hasOwnProperty.call(changed, "type")) {
                                form.setFieldValue("auto_renew", 1);
                            }
                        }
                    }}>
                    {formLoading && <div className={styles.loading}><Spin size="medium"/></div>}
                    <div className={styles.formShell} style={formLoading ? {display: "none"} : undefined}>
                        {certificateForm && (
                            <>
                                {viewing && record?.certificate && (
                                    <Descriptions
                                        className={styles.details}
                                        size="small"
                                        column={{xs: 1, sm: 2}}
                                        colon={false}
                                        styles={{
                                            label: {fontSize: "inherit", flexShrink: 0, whiteSpace: "nowrap"},
                                            content: {fontSize: "inherit", minWidth: 0, overflowWrap: "anywhere", whiteSpace: "normal"},
                                        }}
                                        items={[
                                            {key: "domains", label: "适用域名", children: (record.domains || []).join("、") || "暂无域名"},
                                            {key: "type", label: "签发方式", children: record.type === 1 ? "手动申请" : record.type === 2 ? "自动申请" : "导入证书"},
                                            {key: "start_time", label: "生效时间", children: record.start_time || "未知"},
                                            {key: "expire_time", label: "到期时间", children: record.expire_time || "未知"},
                                            ...([1, 2].includes(record.type) ? [
                                                {key: "challenge", label: "验证类型", children: String(record.challenge || "http").toUpperCase()},
                                                {key: "auto_renew", label: "自动续签", children: record.auto_renew === 1 ? "开启" : "关闭"},
                                                ...(record.challenge === "dns" ? [{key: "secret_id", label: "关联密钥", children: record.secret_id ? secretNames[record.secret_id] || `密钥 #${record.secret_id}` : "未关联"}] : []),
                                            ] : []),
                                        ]}/>
                                )}
                                <section className={styles.formSection}>
                                    <Form.Item
                                        name="name"
                                        label="证书名称"
                                        rules={viewing ? [] : [
                                            {required: true, whitespace: true, message: "请输入证书名称"},
                                            {max: 100, message: "证书名称不能超过 100 个字符"},
                                        ]}>
                                        <Input readOnly={viewing} prefix={<Icon type="SafetyCertificateOutlined"/>} maxLength={100} placeholder="例如：example.com 正式证书" autoComplete="off"/>
                                    </Form.Item>
                                </section>
                                {mode === "edit" && (record?.type === 1 || record?.type === 2) && (
                                    <section className={styles.formSection}>
                                        {verificationField}
                                        {automaticDNS && secretField}
                                        {certificateType === 2 && renewalField}
                                    </section>
                                )}
                                <section className={styles.formSection}>
                                    <Row className={styles.pemGrid} gutter={[16, 0]}>
                                        <Col xs={24} md={12}>
                                            <Form.Item
                                                className={styles.pemField}
                                                name="certificate"
                                                label={<span className={styles.fieldLabel}><span>证书内容</span>{(viewing || mode === "edit") && <Button type="text" icon={<Icon type="DownloadOutlined"/>} onClick={(event) => { event.preventDefault(); download("certificate"); }}>下载</Button>}</span>}
                                                rules={viewing ? [] : [{required: true, whitespace: true, message: "请输入 PEM 证书内容"}]}>
                                                <Input.TextArea
                                                    readOnly={viewing}
                                                    className={styles.editor}
                                                    spellCheck={false}
                                                    placeholder={"粘贴完整证书链（PEM 格式）\n-----BEGIN CERTIFICATE-----"}/>
                                            </Form.Item>
                                        </Col>
                                        <Col xs={24} md={12}>
                                            <Form.Item
                                                className={styles.pemField}
                                                name="private_key"
                                                label={<span className={styles.fieldLabel}><span>私钥内容</span>{(viewing || mode === "edit") && <Button type="text" icon={<Icon type="DownloadOutlined"/>} onClick={(event) => { event.preventDefault(); download("private_key"); }}>下载</Button>}</span>}
                                                rules={viewing ? [] : [{required: true, whitespace: true, message: "请输入 PEM 私钥内容"}]}>
                                                <Input.TextArea
                                                    readOnly={viewing}
                                                    className={styles.editor}
                                                    spellCheck={false}
                                                    placeholder={"粘贴与证书匹配的私钥\n-----BEGIN PRIVATE KEY-----"}/>
                                            </Form.Item>
                                        </Col>
                                    </Row>
                                </section>
                            </>
                        )}
                        {(mode === "obtain" || mode === "renew") && (
                            <>
                                <Form.Item
                                    name="domain"
                                    label="域名或 IP"
                                    extra={ipAddress ? "公网 IP 使用 HTTP 验证，请留意证书有效期。" : wildcard ? "泛域名已自动切换为 DNS 验证。" : undefined}
                                    rules={[{required: true, whitespace: true, message: "请选择或输入域名或 IP"}, {validator: async (_, value) => {
                                        const target = String(value || "").trim();
                                        if (target && (/[\[\]:]/.test(target) || /^[\d.]+$/.test(target)) && !isIPAddress(target)) {
                                            throw new Error("请输入有效 IP 地址，域名无需填写端口");
                                        }
                                    }}]}>
                                    {mode === "renew" ? (
                                        <Select
                                            options={(record?.domains || []).map((domain: any) => ({label: domain, value: domain}))}
                                            placeholder="请选择续签域名或 IP"/>
                                    ) : (
                                        <Input readOnly={Boolean(obtainDomain)} placeholder="域名、公网 IPv4 或 IPv6" autoComplete="off"/>
                                    )}
                                </Form.Item>
                                {mode === "obtain" && (
                                    <Form.Item
                                        name="name"
                                        label="证书名称"
                                        rules={[
                                            {required: true, whitespace: true, message: "请输入证书名称"},
                                            {max: 100, message: "证书名称不能超过 100 个字符"},
                                        ]}>
                                        <Input maxLength={100} placeholder="默认使用证书域名" autoComplete="off"/>
                                    </Form.Item>
                                )}
                                {verificationField}
                                {automaticDNS && secretField}
                                {certificateType === 2 && renewalField}
                                <Form.Item name="ca" label="签发环境" extra={mode === "renew" ? "留空沿用原签发环境；测试证书不受浏览器信任。" : undefined} rules={[{required: mode === "obtain"}]}>
                                    <Select allowClear={mode === "renew"} placeholder={mode === "renew" ? "沿用原签发环境" : "请选择签发环境"} options={[
                                        {label: "正式环境", value: "production"},
                                        {label: "测试环境", value: "staging"},
                                    ]}/>
                                </Form.Item>
                                {!manualVerification && ipAddress && (
                                    <Typography.Paragraph type="secondary">
                                        私有 IP 请导入包含对应 IP 地址的证书。
                                    </Typography.Paragraph>
                                )}
                            </>
                        )}
                    </div>
                </Form>
            </ProModal>
            <ProModal
                ref={manualModalRef}
                title={<Title>验证信息</Title>}
                width={compact ? 380 : 400}
                onOk={form.submit}
                onCancel={() => closeForm(true)}
                modalProps={{
                    rootClassName: styles.modal,
                    zIndex: 1030,
                    footer: [
                        <Button key="cancel" onClick={() => closeForm(true)}>取消</Button>,
                        <Button key="submit" type="primary" loading={submitting} onClick={form.submit}>验证</Button>,
                    ],
                    forceRender: true,
                    closable: true,
                    keyboard: true,
                    mask: {closable: true},
                    style: {top: screens.md ? 100 : 24, paddingBottom: screens.md ? 100 : 24},
                    focusable: {focusTriggerAfterClose: false},
                    afterClose: () => {
                        if (!manualSessionRef.current) setManualSession(null);
                        if (!formOpenRef.current) {
                            form.resetFields();
                            setRecord(undefined);
                        }
                    },
                } as any}>
                {manualSession && (
                    <div className={styles.manualShell} aria-busy={submitting}>
                        <div className={styles.manualContent} ref={manualContentRef}>
                            {manualSession.challenges.length === 0 && (
                                <section className={styles.verificationRecord}>
                                    <div className="verification-heading">
                                        <div className="verification-heading-text">
                                            <span className="verification-title">提交验证</span>
                                            <span className="verification-expiry">有效至 {dayjs.unix(manualSession.expires_at).format("MM-DD HH:mm")}</span>
                                        </div>
                                    </div>
                                    <Typography.Text type="secondary">无需添加验证信息，点击“验证”继续。</Typography.Text>
                                </section>
                            )}
                            {manualSession.challenges.map((manualChallenge, index) => (
                                <section key={manualChallenge.id} className={styles.verificationRecord}>
                                    {manualChallenge.type === "HTTP" && (
                                        <div className="verification-heading">
                                            <div className="verification-heading-text">
                                                <span className="verification-title">部署验证文件{manualSession.challenges.length > 1 ? ` ${index + 1}` : ""}</span>
                                                {index === 0 && <span className="verification-expiry">有效至 {dayjs.unix(manualSession.expires_at).format("MM-DD HH:mm")}</span>}
                                            </div>
                                            <Button size="small" color="default" variant="filled" icon={<Icon type="DownloadOutlined"/>} onClick={() => {
                                                const url = URL.createObjectURL(new Blob([manualChallenge.value], {type: "application/octet-stream"}));
                                                const anchor = document.createElement("a");
                                                anchor.href = url;
                                                anchor.download = manualChallenge.name;
                                                anchor.click();
                                                window.setTimeout(() => URL.revokeObjectURL(url), 1000);
                                            }}>下载文件</Button>
                                        </div>
                                    )}
                                    <Form component="div" layout="vertical" variant="filled" className="verification-fields">
                                        {(manualChallenge.type === "HTTP" ? [
                                            {label: "文件名", value: manualChallenge.name},
                                            {label: "部署路径", value: manualChallenge.path || ""},
                                            {label: "文件内容", value: manualChallenge.value},
                                        ] : [
                                            {label: "记录名称", value: manualChallenge.name},
                                            {label: "记录值", value: manualChallenge.value},
                                        ]).map(({label, value}) => (
                                            <Form.Item key={label} label={manualChallenge.type !== "HTTP" && label === "记录名称" ? (
                                                <span className="verification-field-label">
                                                    <span>{label}<span className="verification-type">TXT{manualSession.challenges.length > 1 ? ` ${index + 1}` : ""}</span></span>
                                                    {index === 0 && <span className="verification-expiry">有效至 {dayjs.unix(manualSession.expires_at).format("MM-DD HH:mm")}</span>}
                                                </span>
                                            ) : label}>
                                                <div className="verification-field-content">
                                                    <Input.TextArea value={value} readOnly variant="filled" autoSize={{minRows: 1}} spellCheck={false} aria-label={label}/>
                                                    <Typography.Text className="verification-copy" copyable={{text: value, tooltips: [`复制${label}`, "已复制"]}}/>
                                                </div>
                                            </Form.Item>
                                        ))}
                                    </Form>
                                </section>
                            ))}
                        </div>
                    </div>
                )}
            </ProModal>
            <BatchCert ref={batchRef} onChange={onChange} secretVersion={secretVersion}/>
            <Log ref={logRef} request={getCertLog} title="续签日志"/>
            <SecretManager ref={secretManagerRef} onChange={onSecretChange}/>
        </>
    );
});

CertManager.displayName = "CertManager";

export default memo(CertManager);
