import React, {Component, useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Button, Empty, Grid, Spin, theme as antdTheme} from "antd";
import type {ModalProps} from "antd";
import {ProModal, Title, useTheme} from "sinking-antd";
import {callSoftware, getSoftwarePage, getSoftwareInfo} from "@/service/api/software";
import createSoftwareModules from "./modules";
import type {PluginModal} from "./sdk";
import useStyles from "./styles";

class SoftwareBoundary extends Component<any, {error: string}> {
    state = {error: ""};
    static getDerivedStateFromError(error: Error) { return {error: error?.message || "软件页面加载失败"}; }
    render() { return this.state.error ? this.props.fallback(this.state.error) : this.props.children; }
}

export default function SoftwareWorkspace({open, software, onCancel}: any) {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const screens = Grid.useBreakpoint();
    const {token} = antdTheme.useToken();
    const {styles} = useStyles();
    const [page, setPage] = useState<{id: number; component: any; record: any; modal?: PluginModal}>();
    const [error, setError] = useState("");
    const [refresh, setRefresh] = useState(0);
    const modulesRef = useRef<ReturnType<typeof createSoftwareModules> | undefined>(undefined);
    const openRef = useRef(open);
    openRef.current = open;
    useEffect(() => {
        if (!open || !software?.id) return;
        let disposed = false;
        let loaded = false;
        let modules: ReturnType<typeof createSoftwareModules> | undefined;
        modulesRef.current?.dispose();
        modulesRef.current = undefined;
        setPage(undefined); setError("");
        void (async () => {
            try {
                const response = await getSoftwareInfo({body: {id: software.id}});
                if (disposed) return;
                if (response?.code !== 200 || !response?.data) throw new Error(response?.message || "获取软件失败");
                const item = response?.data;
                const code = item?.code;
                if (!code) throw new Error("软件标识不存在");
                const frontend = item?.frontend;
                if (!frontend) throw new Error("软件未配置页面入口");
                const {default: shared} = await import("./sdk");
                if (disposed) return;
                modules = createSoftwareModules(frontend,
                    (path) => getSoftwarePage(code, path === frontend ? undefined : path.replace(/^web\//, "")), shared, code);
                modulesRef.current = modules;
                const module = await modules.load();
                if (disposed) return;
                const component = module?.default;
                if (typeof component !== "function" && !component?.$$typeof) throw new Error("软件页面需要默认导出 React 组件");
                loaded = true;
                setPage({id: software.id, record: item, component, modal: module?.modal});
            } catch (reason: any) {
                modules?.dispose();
                if (!disposed) setError(reason?.message || "加载软件失败");
            }
        })();
        // 已显示的页面保留到关闭动画结束，避免样式跳回默认值、动态模块提前失效。
        return () => { disposed = true; if (!loaded) modules?.dispose(); };
    }, [open, software?.id, refresh]);
    useEffect(() => () => modulesRef.current?.dispose(), []);
    const current = page?.id === software?.id ? page : undefined;
    const record = current?.record;
    const request = useCallback(async (method: string, params: any = {}) => {
        const code = record?.code;
        if (!code) throw new Error("软件标识不存在");
        const response = await callSoftware(code, method, {body: params});
        if (response?.code !== 200) throw new Error(response?.message || "软件操作失败");
        return response?.data;
    }, [record?.code]);
    const configuration = useMemo<{props: ModalProps; error: string}>(() => {
        try {
            const value = typeof current?.modal === "function"
                ? current.modal({software: record, compact, mobile: !screens.md, token, onClose: onCancel})
                : current?.modal;
            if (value != null && (typeof value !== "object" || Array.isArray(value) || typeof (value as any).then === "function")) {
                throw new Error("modal 必须是属性对象或同步返回属性对象的函数");
            }
            const props = {...value} as ModalProps;
            // 外观使用原生 Modal 属性，生命周期仍由宿主管理，避免绕过关闭清理。
            for (const key of ["open", "onCancel", "afterClose", "afterOpenChange", "destroyOnHidden", "destroyOnClose",
                "forceRender", "getContainer", "children", "loading"] as const) delete props[key];
            if (props.closable && typeof props.closable === "object") {
                props.closable = {...props.closable};
                delete props.closable.onClose;
                delete props.closable.afterClose;
            }
            return {props, error: ""};
        } catch (reason: any) {
            return {props: {}, error: reason?.message || "软件弹窗配置失败"};
        }
    }, [current?.modal, record, compact, screens.md, token, onCancel]);
    const activeModules = modulesRef.current;
    const afterClose = () => {
        if (openRef.current || modulesRef.current !== activeModules) return;
        activeModules?.dispose();
        modulesRef.current = undefined;
        setPage(undefined); setError(""); setRefresh((value) => value + 1);
    };
    const defaults: ModalProps = {width: compact ? 960 : 1080, zIndex: 1050, footer: null, mask: {closable: true},
        styles: {body: {minWidth: 0}},
        style: {top: screens.md ? 70 : 16, paddingBottom: screens.md ? 70 : 16, maxWidth: "calc(100vw - 24px)"}};
    const custom = configuration.props;
    const customStyles = custom.styles;
    const semanticStyles = (value: any) => ({...value, body: {minWidth: 0, ...value?.body}});
    const modalProps: ModalProps = {...defaults, ...custom,
        style: {...defaults.style, ...(custom.centered ? {top: 0, paddingBottom: 0} : {}), ...custom.style},
        styles: typeof customStyles === "function"
            ? (info) => semanticStyles(customStyles(info))
            : semanticStyles(customStyles)};
    const renderModal = (props: ModalProps, children: React.ReactNode) => <ProModal
        title={<Title>{software?.name || "软件"}管理</Title>}
        modalProps={{...props, open, onCancel, afterClose, destroyOnHidden: true} as any}>{children}</ProModal>;
    const renderError = (text: string) => <div className={styles.loading}>
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={text}>
            <Button color="default" variant="filled" onClick={() => setRefresh((value) => value + 1)}>重试</Button>
        </Empty>
    </div>;
    return <SoftwareBoundary key={`${software?.id}:${refresh}`} fallback={(text: string) => renderModal(defaults, renderError(text))}>
        {renderModal(modalProps, error || configuration.error ? renderError(error || configuration.error)
            : !current ? <div className={styles.loading}><Spin/></div>
                : <SoftwareBoundary fallback={renderError}>{React.createElement(current.component, {software: record, request, onClose: onCancel})}</SoftwareBoundary>)}
    </SoftwareBoundary>;
}
