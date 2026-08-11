import React, {useEffect, useLayoutEffect} from "react";
import {Outlet, useModel, useSelectedRoutes} from "umi";
import {App, ConfigProvider, Spin} from "antd";
import {historyPush} from "@/utils/route";
import {Layout} from "@/layouts/components";
import Title from "./components/title";
import {deleteHeader, getHeaders} from "@/utils/auth";
import request from "@/utils/request";
import defaultSettings from "../../config/defaultSettings";
import {setIconfontUrl, Theme, useTheme} from "sinking-antd";
import {createStyles} from "antd-style";
import zhCN from 'antd/locale/zh_CN';

/**
 * 请求登录校验
 */
const check = async (ctx: any, next: any) => {
    ctx.req.options.headers = {
        ...(ctx.req.options.headers || {}),
        ...getHeaders(),
    };
    await next();
    if (ctx.res?.code == 403 || ctx.res?.code == 503) {
        deleteHeader();
        historyPush("login");
    }
}
request.use(check);

const useStyles = createStyles((): any => {
    return {
        load: {
            width: "100%",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            height: "80vh",
        },
    };
});

const ProLayout = () => {
    const web = useModel("web");
    const user = useModel("user");
    const routes = useSelectedRoutes();
    const {styles: {load}} = useStyles();
    const currentRoute = routes?.at(-1)?.route as any;
    const isPublic = currentRoute?.auth === false;

    // 初始化 iconfont 地址
    useEffect(() => {
        if (defaultSettings.iconfontUrl) {
            setIconfontUrl(defaultSettings.iconfontUrl);
        }
    }, []);

    const theme = useTheme();
    const ui = web?.info?.ui;

    // 仅受保护页面需要加载账户信息
    useEffect(() => {
        if (!isPublic && !user?.web) {
            user?.refreshWebUser();
        }
    }, [isPublic]);

    // 初始化主题
    useLayoutEffect(() => {
        if (ui) {
            if (ui.color) {
                theme?.setColor(ui.color);
            }
            if (ui.radius >= 0) {
                theme?.setRadius(ui.radius <= 15 ? ui.radius : 0);
            }
            if (ui.compact) {
                theme?.setCompactTheme();
            } else {
                theme?.setDefaultTheme();
            }
        }
    }, [ui]);

    if (!web.info || !ui) {
        return <Spin spinning={true} size="large" className={load}/>;
    }
    if (isPublic) {
        return <>
            <Title/>
            <App>
                <Outlet/>
            </App>
        </>;
    }
    if (!user?.web) {
        return <Spin spinning={true} size="large" className={load}/>;
    }
    return <Layout/>;
}

export default () => {
    return (
        <ConfigProvider locale={zhCN}>
            <Theme>
                <ProLayout/>
            </Theme>
        </ConfigProvider>
    );
}
