/**
 * 系统路由
 */
export default [
    {
        path: "/",
        title: "系统概览",
        name: "index",
        icon: 'icon-home',
        hideInMenu: false,
        hideBreadCrumb: false,
        component: "@/pages/index",
    },
    {
        path: "file",
        title: "文件管理",
        name: "file",
        icon: "FolderOutlined",
        hideInMenu: false,
        component: "@/pages/test",
    },
    {
        path: "task",
        title: "计划任务",
        name: "task",
        icon: "ScheduleOutlined",
        hideInMenu: false,
        component: "@/pages/test",
    },
    {
        path: "terminal",
        title: "终端管理",
        name: "terminal",
        icon: "CodeOutlined",
        hideInMenu: false,
        component: "@/pages/test",
    },
    {
        path: "setting",
        title: "系统设置",
        name: "setting",
        icon: "SettingOutlined",
        hideInMenu: false,
        component: "@/pages/test",
    },
    {
        path: 'login',
        component: '@/pages/login',
        name: "login",
        title: "帐号登录",
        auth: false,
        hideInMenu: true,
    },
    {
        path: '500',
        component: '@/pages/500',
        name: "error",
        title: "服务器错误",
        auth: false,
        hideInMenu: true,
    },
    {
        path: '403',
        component: '@/pages/403',
        name: "notAllowed",
        title: "无权限",
        auth: false,
        hideInMenu: true,
    },
    {
        path: '*',
        component: '@/pages/404',
        name: "notFound",
        title: "页面不存在",
        auth: false,
        hideInMenu: true,
    }
];
