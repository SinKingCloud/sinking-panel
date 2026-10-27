import React, {lazy, Suspense} from "react";
import * as jsxRuntime from "react/jsx-runtime";
import * as jsxDevRuntime from "react/jsx-dev-runtime";
import * as antd from "antd";
import * as antdStyle from "antd-style";
import * as sinkingAntd from "sinking-antd";
import dayjs from "dayjs";
import ActionDropdown from "@/pages/components/action-dropdown";
import Log from "@/pages/components/task-log";
import {formatSize} from "@/utils/string";
import FormField from "./fields";

export type PluginModalProps = Omit<antd.ModalProps,
    "open" | "onCancel" | "afterClose" | "afterOpenChange" | "destroyOnHidden" |
    "destroyOnClose" | "forceRender" | "getContainer" | "children" | "loading">;

export interface PluginModalContext {
    software: any;
    compact: boolean;
    mobile: boolean;
    token: antd.GlobalToken;
    onClose: () => void;
}

export type PluginModal = PluginModalProps | ((context: PluginModalContext) => PluginModalProps);

const Terminal = lazy(() => import("@/pages/components/container-terminal"));

function ContainerTerminal(props: any) {
    return <Suspense fallback={<antd.Spin/>}><Terminal {...props}/></Suspense>;
}

// import 直接复用面板依赖，保持同一份 React 和主题上下文。
export default {
    react: React,
    "react/jsx-runtime": jsxRuntime,
    "react/jsx-dev-runtime": jsxDevRuntime,
    antd,
    "sinking-antd": sinkingAntd,
    "antd-style": antdStyle,
    dayjs,
    sdk: {ActionDropdown, Log, ContainerTerminal, FormField, formatSize},
};
