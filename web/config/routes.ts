/**
 * 系统路由
 */
export default [
    {
        path: "/",
        title: "首页",
        name: "index",
        icon: 'icon-home',
        hideInMenu: false,
        hideBreadCrumb: true,
        component: "@/pages/index",
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
