import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Layout, Icon, ProModal, Title as ModalTitle, useTheme} from "sinking-antd";
import {useModel, useSelectedRoutes, useLocation, history, Outlet} from "umi";
import {deleteHeader} from "@/utils/auth";
import {getAllMenuItems, getFirstMenuWithoutChildren, getParentList, historyPush} from "@/utils/route";
import {App, Button, Empty, Popover, Progress, Spin, Tooltip} from "antd";
import {createStyles} from "antd-style";
import Settings from "@/../config/defaultSettings";
import {logout} from "@/service/auth/login";
import {cancelSystemTask, getSystemTask, getSystemTaskList} from "@/service/api/system";
import TaskLog, {LogRef as TaskLogRef} from "@/pages/task/components/log";
import defaultSettings from "@/../config/defaultSettings";
import Title from "../title";

/**
 * 样式
 */
const useRightTopStyles = createStyles(({css, token, isDarkMode}: any): any => {
    return {
        nickname: {
            fontSize: "13px",
            color: isDarkMode ? token.colorTextSecondary : "rgb(150,150,150)",
            fontWeight: "bold",
            marginLeft: "3px",
        },
        bottomIconDark: {
            color: isDarkMode ? token.colorTextSecondary : "rgb(150,150,150)",
        },
        pop: css`
            margin-left: 10px;
            margin-right: 20px;
            display: initial;
            padding: 11px 5px;
            border-radius: 10px;
            transition: background-color 0.3s ease;
            cursor: pointer;

            .anticon {
                margin-left: 2px;
                font-size: 10px;
            }
        `,
        box: css`
            &.ant-popover {
                filter: none !important;
            }

            .ant-popover-container {
                padding: 0 !important;
                overflow: hidden;
                border-radius: ${token?.borderRadiusLG}px;
                box-shadow: 0 0 5px 0 rgba(0, 0, 0, 0.15) !important;
                width: 210px;
            }

            .ant-popover-arrow:before {
                background-color: ${token?.colorPrimary} !important;
            }
        `,
        content_top: css`
            height: 64px;
            box-sizing: border-box;
            background-color: ${token?.colorPrimary};
            overflow: hidden;
            background-image: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAIgAAACGBAMAAAD0nt8RAAAAD1BMVEVHcEz///////////////8T4DEaAAAABXRSTlMADAYJA8T7L0gAAALASURBVGjezVpbcoMwDCS2DxBhDgC0BwhNDoDb3v9MfaQQMLb1cqfVVz6YjbS7ks2IptnFCOD7RhcDfIafVRgGvqPVJ/IZmoLcD4YqFbuAgAIkrCCKeqYaICsGnP8WxP0bkE05F71hK4EoWvC0YHh9EwN0v93Fr9frs3aefP/PjdA9Pfo3F8xuxRk74LlaTBpLaPOATaRA0A+lBPAB6rBUNy2KVXwmDOF8YwSs3g1Ij8zYVgPi0PYjlGOppJVATmi9BgcZUJDH1PKokzIamwdGVkGDPGEeB2S+jU9lT2/4KFASip4edxgtfpxDidJiIpvmOsoTYfQEI8UPmR3GOFJuOLHGe0o97YYT8fYKNE4jSnYP7mUpH2x292SW0vKtIyQlNAeM4pEzpTQ0E3BAXOpJA4mYqZTcCx+BCRKOOg5JDE+m5AskjVECcfGzJoNRsokFanRkSgpxJlNSiAudEgmIoYPM+AVWAzKQMTzlzUAOwqCkq0FJV8FqBa9NFWzCsBrt2BLbhEEJ1KDE16DE16Ckq0FJV4GSrNc4lGS9xqEk6zUOJVmvcSgB/UAq2GSqAWJq2IQxpQsgpsZcY4h89Jpbf0xSr5kngBeuaSOv3f/9xuQ2LUnP4tanFWlZqfiMID1nHnQZry/Kv/FB3PF9giTz9foyju/vc6Spl7TQHW5IqDaAPBphKhn/hBogbqoAwpv7WRNKue2kkzI/ZoSpzPIjtTWZ1+03BkgfzRPJKdQupmjlo98v1hoVp9AyiGbFKdRSN0Wy1x56C6FLJKsuhiYzYZtleL0iu2zQtoROXQwqM3nlMmmLQWRm7BkHfSJ5mXkL6aAuJiszd5FsBc1Lkpm/ATbqYpIyS1bRcSo35U5dseLfT0rpXt3qE6GtUTjcKj4SWFNRfcPh9Kvs9bLRN7qY1B+k/HCrTeSL24saozF4MR+gwScpYNNOxQAAAABJRU5ErkJggg==);
            background-repeat: no-repeat;
            background-position: right;
            display: flex;
            align-items: center;
            padding: 0 15px;
        `,
        top_text: {
            color: "#fff",
            width: "100%",
            minWidth: 0,
        },
        userName: {
            fontSize: "14px",
            fontWeight: 600,
            lineHeight: "20px",
            whiteSpace: "nowrap",
            overflow: "hidden",
            textOverflow: "ellipsis",
        },
        userIp: {
            marginTop: "3px",
            fontSize: "12px",
            lineHeight: "18px",
            opacity: 0.8,
            whiteSpace: "nowrap",
            overflow: "hidden",
            textOverflow: "ellipsis",
        },
        menuItemLabel: {
            display: "flex",
        },
        menuItemLeadIcon: {
            fontSize: "11px",
        },
        menu: {
            listStyle: "none",
            padding: 0,
            margin: 0,
            userSelect: "none",
            "li:last-of-type": {
                borderTop: "0.5px solid rgba(189, 189, 189, 0.2)",
                borderRadius: "0px 0px " + token.borderRadius + "px " + token.borderRadius + "px",
                height: "45px",
                lineHeight: "45px",
            }
        },
        menuItem: {
            cursor: "pointer",
            letterSpacing: "1px",
            height: "40px",
            lineHeight: "40px",
            fontSize: "12px",
            padding: "0px 15px",
            transition: "background-color 0.3s ease",
            color: isDarkMode ? token.colorTextSecondary : "rgba(0,0,0,0.65)",
            display: "flex",
            justifyContent: "space-between",
            ":hover": {
                backgroundColor: "rgba(0, 0, 0, 0.03)",
            },
            ".anticon": {
                fontSize: "11px",
            },
            "div>.anticon": {
                fontSize: "12.5px",
                marginRight: "7px"
            }
        },
        icon: {
            fontSize: "17px",
            padding: "7px",
            marginRight: "5px",
            cursor: "pointer",
            verticalAlign: "middle",
            borderRadius: "5px",
            transition: "background-color 0.3s ease",
            ":hover": {
                backgroundColor: "rgba(0, 0, 0, 0.1)",
            },
            color: isDarkMode ? token.colorTextSecondary : "rgb(150,150,150)"
        },
        taskBadge: css`
            position: relative;
            display: inline-flex;
            align-items: center;
            line-height: 1;
            vertical-align: middle;
        `,
        taskBadgeCount: css`
            position: absolute;
            z-index: 1;
            top: 3px;
            inset-inline-end: 3px;
            min-width: 15px;
            height: 15px;
            padding: 0 4px;
            box-sizing: border-box;
            border: 1px solid ${token.colorBgContainer};
            border-radius: 8px;
            color: ${token.colorTextLightSolid};
            background: ${token.colorError};
            font-size: 10px;
            font-weight: 600;
            line-height: 13px;
            text-align: center;
            white-space: nowrap;
            pointer-events: none;
        `,
        taskList: css`
            max-height: min(520px, 58dvh);
            overflow-y: auto;
            padding: 6px 0 0;
        `,
        taskItem: css`
            min-width: 0;
            padding: 8px 9px;
            border-radius: ${token.borderRadius}px;
            background: ${isDarkMode ? token.colorFillQuaternary : token.colorFillSecondary};

            & + & {
                margin-top: 4px;
            }
        `,
        taskBody: css`
            min-width: 0;
        `,
        taskHeader: css`
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 8px;
        `,
        taskStatusDot: css`
            width: 7px;
            height: 7px;
            flex: none;
            border-radius: 50%;
            background: ${token.colorPrimary};

            &.is-completed {
                background: ${token.colorSuccess};
            }

            &.is-failed {
                background: ${token.colorError};
            }

            &.is-canceled {
                background: ${token.colorTextQuaternary};
            }
        `,
        taskName: css`
            min-width: 0;
            flex: 1;
            overflow: hidden;
            cursor: pointer;
            color: ${token.colorText};
            font-size: 13px;
            font-weight: 500;
            line-height: 22px;
            text-overflow: ellipsis;
            white-space: nowrap;
        `,
        taskStatus: css`
            flex: none;
            color: ${token.colorPrimary};
            font-size: 11px;
            font-variant-numeric: tabular-nums;
            line-height: 22px;

            &.is-completed {
                color: ${token.colorSuccess};
            }

            &.is-failed {
                color: ${token.colorError};
            }

            &.is-canceled {
                color: ${token.colorTextTertiary};
            }
        `,
        taskLogAction: css`
            flex: none;
            width: 24px;
            height: 24px;
            padding: 0;
            border-radius: ${token.borderRadiusSM}px;
            color: ${token.colorTextTertiary};

            &:hover {
                color: ${token.colorPrimary};
                background: ${token.colorFillQuaternary};
            }

            .anticon {
                font-size: 10px !important;
            }
        `,
        taskProgressRow: css`
            display: flex;
            align-items: center;
            gap: 8px;
            margin-top: 5px;

            .ant-progress {
                min-width: 0;
                flex: 1;
                margin: 0;
            }
        `,
        taskProgressValue: css`
            flex: none;
            min-width: 34px;
            color: ${token.colorTextSecondary};
            font-size: 11px;
            font-variant-numeric: tabular-nums;
            line-height: 16px;
            text-align: end;
        `,
        taskMessage: css`
            min-width: 0;
            flex: 1;
            overflow: hidden;
            color: ${token.colorTextTertiary};
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: 10px;
            line-height: 15px;
            text-overflow: ellipsis;
            white-space: nowrap;
        `,
        taskMessageRow: css`
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 6px;
            margin-top: 4px;
        `,
        taskEmpty: css`
            min-height: 150px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: ${token.colorTextTertiary};
        `,
    };
});

const taskStatusValues = {
    pending: 0,
    running: 1,
    completed: 2,
    failed: 3,
    canceled: 4,
} as const;

const taskTerminalStatuses = new Set([
    taskStatusValues.completed,
    taskStatusValues.failed,
    taskStatusValues.canceled,
]);
const taskStatusMeta: Record<number, {label: string; className: string}> = {
    [taskStatusValues.pending]: {label: "排队中", className: "running"},
    [taskStatusValues.running]: {label: "处理中", className: "running"},
    [taskStatusValues.completed]: {label: "已完成", className: "completed"},
    [taskStatusValues.failed]: {label: "失败", className: "failed"},
    [taskStatusValues.canceled]: {label: "已取消", className: "canceled"},
};

/**
 * 右侧部分组件
 * @constructor
 */
const RightTop: React.FC = () => {
    /**
     * 全局数据
     */
    const user = useModel("user");//用户信息
    const theme = useTheme();//主题信息
    const {message, modal} = App.useApp();
    const [userOpen, setUserOpen] = useState(false);
    const [taskOpen, setTaskOpen] = useState(false);
    const [taskLoading, setTaskLoading] = useState(false);
    const [tasks, setTasks] = useState<any[]>([]);
    const taskRequestRef = useRef(0);
    const taskLoadedRef = useRef(false);
    const mountedRef = useRef(true);
    const taskLogRef = useRef<TaskLogRef>({} as TaskLogRef);

    const refreshTasks = useCallback(async () => {
        const requestId = ++taskRequestRef.current;
        if (!taskLoadedRef.current) {
            setTaskLoading(true);
        }
        try {
            const response = await getSystemTaskList();
            if (!mountedRef.current || requestId !== taskRequestRef.current) {
                return;
            }
            if (response?.code === 200 && Array.isArray(response.data)) {
                setTasks(response.data);
            }
        } finally {
            if (mountedRef.current && requestId === taskRequestRef.current) {
                taskLoadedRef.current = true;
                setTaskLoading(false);
            }
        }
    }, []);

    useEffect(() => {
        mountedRef.current = true;
        let stopped = false;
        let timer = 0;
        const poll = async () => {
            try {
                await refreshTasks();
            } catch {
                // A transient task-list failure should not stop future polling.
            } finally {
                if (!stopped) {
                    timer = window.setTimeout(poll, 3000);
                }
            }
        };
        void poll();
        return () => {
            stopped = true;
            mountedRef.current = false;
            window.clearTimeout(timer);
        };
    }, [refreshTasks]);

    const cancelTask = useCallback(async (id: string) => {
        const response = await cancelSystemTask({body: {id}});
        if (response?.code !== 200) {
            message.error(response?.message || "取消任务失败");
            return;
        }
        await refreshTasks();
    }, [message, refreshTasks]);

    const deleteTask = useCallback(async (id: string) => {
        const response = await getSystemTask({body: {id, action: "delete"}});
        if (response?.code !== 200) {
            message.error(response?.message || "删除任务失败");
            return;
        }
        await refreshTasks();
    }, [message, refreshTasks]);

    const openTaskLog = useCallback((task: any) => {
        taskLogRef.current?.open(task);
    }, []);

    const requestSystemTaskLog = useCallback((params: API.RequestParams = {}) => getSystemTask({
        ...params,
        body: {...(params.body || {}), action: "log"},
    }), []);

    /**
     * 退出登录
     */
    const outLogin = async () => {
        message?.loading({content: "正在退出登录", duration: 600000, key: "outLogin"});
        await logout({
            onSuccess: (r) => {
                message?.success(r?.message || "退出登录成功");
                deleteHeader();
                user?.setWeb(undefined);
                historyPush("login");
            },
            onFail: (r) => {
                message?.error(r?.message || "退出登录失败");
            },
            onFinally: () => {
                message?.destroy("outLogin");
            },
        });
    };

    /**
     * 退出登录确认
     */
    const confirmOutLogin = () => {
        modal.confirm({
            title: "退出登录",
            content: "确定退出当前账号吗？",
            okText: "确定",
            okButtonProps: {danger: true, type: "default"},
            cancelText: "取消",
            mask: {
                closable: true,
            },
            onOk: outLogin,
        } as any);
    };

    const {
        styles: {
            nickname,
            bottomIconDark,
            pop,
            content_top,
            top_text,
            userName,
            userIp,
            box,
            menu,
            menuItem,
            icon,
            menuItemLabel,
            menuItemLeadIcon,
            taskList,
            taskItem,
            taskBody,
            taskHeader,
            taskStatusDot,
            taskName,
            taskStatus,
            taskLogAction,
            taskProgressRow,
            taskProgressValue,
            taskMessage,
            taskMessageRow,
            taskEmpty,
            taskBadge,
            taskBadgeCount,
        }
    } = useRightTopStyles();

    const pendingTaskCount = tasks.filter((task) => !taskTerminalStatuses.has(Number(task.status))).length;
    const displayTasks = useMemo(() => tasks
        .map((task, index) => ({task, index}))
        .sort((first, second) => {
            const firstStatus = Number(first.task.status);
            const secondStatus = Number(second.task.status);
            const firstPriority = firstStatus === taskStatusValues.running
                ? 0
                : firstStatus === taskStatusValues.pending ? 1 : 2;
            const secondPriority = secondStatus === taskStatusValues.running
                ? 0
                : secondStatus === taskStatusValues.pending ? 1 : 2;
            return firstPriority - secondPriority || first.index - second.index;
        })
        .map(({task}) => task), [tasks]);
    const taskContent = (
        <>
            {taskLoading && tasks.length === 0 ? (
                <div className={taskEmpty}>
                    <Spin size="small"/>
                </div>
            ) : tasks.length === 0 ? (
                <div className={taskEmpty}>
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据"/>
                </div>
            ) : (
                <div className={taskList}>
                    {displayTasks.map((task) => {
                        const status = Number(task.status);
                        const meta = taskStatusMeta[status] || taskStatusMeta[taskStatusValues.pending];
                        const rawProgress = Number(task.progress);
                        const progress = status === taskStatusValues.completed
                            ? 100
                            : Number.isFinite(rawProgress)
                            ? Math.max(0, Math.min(100, rawProgress))
                            : 0;
                        const name = task.name || "未命名任务";
                        return (
                            <div className={taskItem} key={task.id}>
                                <div className={taskBody}>
                                    <div className={taskHeader}>
                                        <span className={`${taskStatusDot} is-${meta.className}`} aria-hidden="true"/>
                                        <span
                                            className={taskName}
                                            title={`${name}，点击查看日志`}
                                            role="button"
                                            tabIndex={0}
                                            onClick={() => openTaskLog(task)}
                                            onKeyDown={(event) => {
                                                if (event.key === "Enter" || event.key === " ") {
                                                    event.preventDefault();
                                                    openTaskLog(task);
                                                }
                                            }}>{name}</span>
                                        <span className={`${taskStatus} is-${meta.className}`}>
                                            {meta.label}
                                        </span>
                                        {(status === taskStatusValues.pending || status === taskStatusValues.running || taskTerminalStatuses.has(status)) && (
                                            <Tooltip title={status === taskStatusValues.pending || status === taskStatusValues.running ? "取消任务" : "删除任务"}>
                                                <Button
                                                    type="text"
                                                    danger
                                                    size="small"
                                                    icon={<Icon type={status === taskStatusValues.pending || status === taskStatusValues.running ? "CloseOutlined" : "DeleteOutlined"}/>}
                                                    aria-label={`${status === taskStatusValues.pending || status === taskStatusValues.running ? "取消" : "删除"}${name}`}
                                                    title={status === taskStatusValues.pending || status === taskStatusValues.running ? "取消任务" : "删除任务"}
                                                    onClick={(event) => {
                                                        event.stopPropagation();
                                                        void (status === taskStatusValues.pending || status === taskStatusValues.running
                                                            ? cancelTask(task.id)
                                                            : deleteTask(task.id));
                                                    }}/>
                                                </Tooltip>
                                        )}
                                    </div>
                                    {status === taskStatusValues.running && (
                                        <div className={taskProgressRow}>
                                            <Progress
                                                percent={progress}
                                                size="small"
                                                showInfo={false}/>
                                            <span className={taskProgressValue}>{Math.floor(progress)}%</span>
                                        </div>
                                    )}
                                    <div className={taskMessageRow}>
                                        <div className={taskMessage} title={task.message || undefined}>
                                            {task.message || "等待处理"}
                                        </div>
                                        <Tooltip title="查看详细日志">
                                            <Button
                                                type="text"
                                                size="small"
                                                className={taskLogAction}
                                                icon={<Icon type="FileTextOutlined" style={{fontSize: 10}}/>}
                                                aria-label={`查看${name}的详细日志`}
                                                onClick={(event) => {
                                                    event.stopPropagation();
                                                    openTaskLog(task);
                                                }}/>
                                        </Tooltip>
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </>
    );

    return <>
        <Tooltip title="系统任务">
            <span
                className={taskBadge}
                style={{
                    position: "relative",
                    display: "inline-flex",
                    alignItems: "center",
                    lineHeight: 1,
                    verticalAlign: "middle",
                }}>
                <Icon
                    type="CloudServerOutlined"
                    className={icon}
                    aria-label="系统任务"
                    onClick={() => {
                        setTaskOpen(true);
                        void refreshTasks();
                    }}/>
                {pendingTaskCount > 0 && (
                    <span
                        className={taskBadgeCount}
                        style={{
                            position: "absolute",
                            top: 3,
                            insetInlineEnd: 3,
                        }}
                        aria-label={`${pendingTaskCount} 个进行中的任务`}>
                        {pendingTaskCount > 99 ? "99+" : pendingTaskCount}
                    </span>
                )}
            </span>
        </Tooltip>
        <Tooltip title={theme?.getModeName(theme?.mode as any)}>
            <Icon type={theme?.isDarkMode() ? "icon-dark" : (theme?.isAutoMode() ? "icon-auto" : "icon-light")}
                  className={icon}
                  onClick={() => {
                      theme?.toggle?.();
                  }}/>
        </Tooltip>
        <Popover rootClassName={box} autoAdjustOverflow={false}
                 open={userOpen}
                 onOpenChange={setUserOpen}
                 placement="bottomRight"
                 content={<>
                     <div className={content_top}>
                         <div className={top_text}>
                             <div className={userName}>{user?.web?.account || "未登录"}</div>
                             <div className={userIp}>{user?.web?.login_ip || "未知IP"}</div>
                         </div>
                     </div>
                     <ul className={menu}>
                         <li className={menuItem} onClick={() => {
                             historyPush("log");
                         }}>
                             <div className={menuItemLabel}>
                                 <Icon type="FileTextOutlined" className={menuItemLeadIcon}/>操作日志
                             </div>
                             <Icon type="RightOutlined"/>
                         </li>
                         <li className={menuItem} onClick={() => {
                             historyPush("setting");
                         }}>
                             <div className={menuItemLabel}>
                                 <Icon type="SettingOutlined" className={menuItemLeadIcon}/>系统设置
                             </div>
                             <Icon type="RightOutlined"/>
                         </li>
                         <li className={menuItem} onClick={() => {
                             setUserOpen(false);
                             confirmOutLogin();
                         }}>
                             <div className={menuItemLabel}>
                                 <Icon type="LogoutOutlined" className={menuItemLeadIcon}/>退出登录
                             </div>
                             <Icon type="RightOutlined"/>
                         </li>
                     </ul>
                 </>}>
            <div className={pop}>
                <span className={nickname}>{user?.web?.account || "未登录"}</span>
                <Icon className={theme?.isDarkMode() ? bottomIconDark : ""} type="DownOutlined"/>
            </div>
        </Popover>
        <ProModal
            title={<ModalTitle>系统任务</ModalTitle>}
            width="min(520px, calc(100vw - 24px))"
            onCancel={() => setTaskOpen(false)}
            modalProps={{
                open: taskOpen,
                footer: null,
                destroyOnHidden: false,
                mask: {closable: true},
                style: {top: 100, paddingBottom: 100},
                styles: {body: {padding: 0}},
            }}>
            {taskContent}
        </ProModal>
        <TaskLog ref={taskLogRef} request={requestSystemTaskLog} showClear={false}/>
    </>
}

/**
 * 样式信息
 */
const useSKLayoutStyles = createStyles((): any => {
    return {
        collapsedLogo: {
            fontSize: "27px",
        },
        content: {
            width: "100%",
            maxWidth: "1450px",
            minWidth: 0,
            margin: "0 auto",
        },
        unCollapsed: {
            overflow: "hidden",
            position: "absolute",
            display: "inline-flex",
            ">span": {
                fontSize: "27px",
            },
            ">div": {
                fontSize: "25px",
                marginLeft: "5px",
                fontWeight: "bolder",
                float: "left",
                lineHeight: "30px",
                whiteSpace: "nowrap",
            }
        },
    };
});

/**
 * 用户系统
 */
const SKLayout: React.FC = () => {
    /**
     * 全局信息
     */
    const user = useModel("user");
    const web = useModel("web");
    const location = useLocation();
    const match = useSelectedRoutes();
    const currentRoute = match?.at(-1)?.route as any;
    const menus = useMemo(() => getAllMenuItems(true), []);
    const menusWithHidden = useMemo(() => getAllMenuItems(false), []);
    const isTopLayout = web?.info?.ui?.layout != "left";
    const {styles: {collapsedLogo, content, unCollapsed}} = useSKLayoutStyles();

    /**
     * 计算面包屑数据
     */
    const breadCrumbItems = useMemo(() => {
        if (!location?.pathname) return [];

        const items = getParentList(menusWithHidden, currentRoute?.name);
        const temp = [{
            title: "系统概览",
            onClick: () => {
                historyPush("index");
            },
        }];

        const handleItemClick = (x: any) => {
            if (x?.children && x?.children?.length > 0) {
                historyPush(getFirstMenuWithoutChildren(x?.children)?.name || "");
            } else {
                historyPush(x?.name);
            }
        };

        items.forEach((x) => {
            temp.push({
                title: x?.label,
                onClick: () => handleItemClick(x),
            });
        });

        return temp;
    }, [location?.pathname, match, menusWithHidden]);
    return (
        <>
            <Title/>
            <Layout
                pathname={location?.pathname}
                matchedRoutes={match || []}
                onNavigate={(path) => history.push(path)}
                breadCrumbItems={breadCrumbItems}
                hideBreadCrumb={currentRoute?.hideBreadCrumb}
                waterMark={web?.info?.ui?.watermark ? [web?.info?.name, user?.web?.account] : ""}
                menus={menus}
                layout={isTopLayout ? "horizontal" : "inline"}
                flowLayout={isTopLayout}
                menuTheme={web?.info?.ui?.theme == "dark" ? "dark" : "light"}
                footer={<>©{new Date().getFullYear()} {web?.info?.name || Settings?.title}</>}
                headerHidden={false}
                headerFixed={false}
                headerRight={<RightTop/>}
                menuCollapsedWidth={60}
                menuUnCollapsedWidth={210}
                collapsedLogo={() => {
                    return <Icon type={"icon-logo"} style={{color: web?.info?.ui?.color || defaultSettings?.color}}
                                 className={collapsedLogo}/>;
                }}
                unCollapsedLogo={() => {
                    return (
                        <div className={unCollapsed}>
                            <Icon type={"icon-logo"}
                                  style={{color: web?.info?.ui?.color || defaultSettings?.color}}/>
                            <div style={{color: web?.info?.ui?.color || defaultSettings?.color}}>
                                {web?.info?.name || Settings?.title}
                            </div>
                        </div>)
                }}>
                <div className={content}>
                    <Outlet/>
                </div>
            </Layout>
        </>
    );
}
export default SKLayout;
