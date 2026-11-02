export interface SystemTaskRef {
    /** 打开系统任务弹窗，并立即刷新任务列表。 */
    open: () => void;
}

/** 通过 Umi Outlet 向页面提供布局内的公共组件。 */
export interface LayoutContext {
    task: SystemTaskRef | null;
}
