export interface SystemTaskRef {
    /** 单个任务打开日志；多个任务或不传 ID 时打开系统任务列表。 */
    open: (ids?: string | string[]) => void;
}

/** 通过 Umi Outlet 向页面提供布局内的公共组件。 */
export interface LayoutContext {
    task: SystemTaskRef | null;
}
