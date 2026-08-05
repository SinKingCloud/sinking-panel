import React, {useMemo, useState} from "react";
import {Layout, Icon, useTheme} from "sinking-antd";
import {useModel, useSelectedRoutes, useLocation, history, Outlet} from "umi";
import {deleteHeader} from "@/utils/auth";
import {getAllMenuItems, getFirstMenuWithoutChildren, getParentList, historyPush} from "@/utils/route";
import {App, Popover, Tooltip} from "antd";
import {createStyles} from "antd-style";
import Settings from "@/../config/defaultSettings";
import {logout} from "@/service/auth/login";
import defaultSettings from "@/../config/defaultSettings";

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
            .ant-popover-container {
                padding: 0 !important;
                overflow: hidden;
                border-radius: ${token?.borderRadiusLG}px;
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
            borderRadius: "5px",
            transition: "background-color 0.3s ease",
            ":hover": {
                backgroundColor: "rgba(0, 0, 0, 0.1)",
            },
            color: isDarkMode ? token.colorTextSecondary : "rgb(150,150,150)"
        },
    };
});

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
            okText: "确 定",
            okButtonProps: {danger: true, type: "default"},
            cancelText: "取 消",
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
        }
    } = useRightTopStyles();
    return <>
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
            maxWidth: "1440px",
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
    );
}
export default SKLayout;
