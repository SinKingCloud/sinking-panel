import React from "react";
import {Col, Menu, Row} from "antd";
import {Body, useTheme} from "sinking-antd";
import {history, useLocation, useModel} from "umi";
import Account from "./components/account";
import Ui from "./components/ui";
import Web from "./components/web";
import HeroGraphic from "@/pages/components/hero-graphic";
import useStyles from "./styles";

const items = [
    {key: "web", title: "网站设置", component: Web},
    {key: "ui", title: "界面设置", component: Ui},
    {key: "account", title: "登录设置", component: Account},
];

export default (): React.ReactNode => {
    const theme = useTheme();
    const web = useModel("web");
    const location = useLocation();
    const isCompactMode = theme?.isCompactTheme?.() || false;
    const isDarkMode = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles({isCompactMode, isDarkMode});
    const selected = new URLSearchParams(location.search).get("tab") || "web";
    const activeItem = items.find((item) => item.key === selected) || items[0];
    const active = activeItem.key;
    const Component = activeItem.component;
    const changeActive = (key: string) => {
        const search = new URLSearchParams(location.search);
        search.set("tab", key);
        history.replace({pathname: location.pathname, search: `?${search.toString()}`});
    };
    return (
        <Body>
            <Row className={styles.page} gutter={[0, isCompactMode ? 10 : 12]}>
                <Col span={24}>
                    <section className={styles.workspace}>
                        <section className={styles.hero}>
                            <div className="hero-copy">
                                <div className="eyebrow"><span className="status-dot"/>SYSTEM SETTINGS</div>
                                <h1>系统设置</h1>
                            </div>
                            <div className="hero-visual"><HeroGraphic variant="setting"/></div>
                        </section>
                        <div className={styles.settingsMain}>
                            <div className={styles.leftMenu}>
                                <div className={styles.menuScroll}>
                                    <Menu
                                        className={styles.menu}
                                        mode="inline"
                                        selectedKeys={[active]}
                                        onClick={({key}) => changeActive(key)}
                                        items={items.map((item) => ({key: item.key, label: item.title}))}
                                    />
                                </div>
                            </div>
                            <section className={styles.right}>
                                <div className={styles.title}>{activeItem.title}</div>
                                <Component styles={styles} info={web?.info}/>
                            </section>
                        </div>
                    </section>
                </Col>
            </Row>
        </Body>
    );
};
