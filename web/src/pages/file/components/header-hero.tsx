import React, {useMemo} from "react";
import type {MenuProps} from "antd";
import {Icon} from "sinking-antd";
import Dropdown from "@/pages/components/stable-dropdown";
import HeroGraphic from "@/pages/components/hero-graphic";
import {normalizeFilePath} from "../utils";
import type {HeaderStyles} from "./header.types";

interface HeaderHeroProps {
    disks: string[];
    onNavigate: (path: string) => void;
    styles: Pick<HeaderStyles, "hero" | "toolbarDropdown">;
}

const HeaderHero = ({disks, onNavigate, styles}: HeaderHeroProps) => {
    const diskMenu = useMemo<MenuProps>(() => ({
        items: disks.map((item, index) => ({
            key: String(index),
            label: item,
        })),
        onClick: ({key}) => {
            const target = disks[Number(key)];
            if (target) onNavigate(normalizeFilePath(target));
        },
    }), [disks, onNavigate]);

    return (
        <section className={styles.hero}>
            <div className="hero-copy">
                <div className="eyebrow"><span className="status-dot"/>FILE MANAGER</div>
                <h1>文件管理</h1>
            </div>
            <div className="hero-visual"><HeroGraphic variant="task"/></div>
            {disks.length > 1 && (
                <Dropdown
                    trigger={["click"]}
                    placement="bottomRight"
                    classNames={{root: styles.toolbarDropdown}}
                    menu={diskMenu}>
                    <button className="create-button" type="button" aria-label="切换磁盘">
                        <Icon type="DatabaseOutlined"/>
                        <span>切换磁盘</span>
                        <Icon type="SwapOutlined" className="switch-arrow"/>
                    </button>
                </Dropdown>
            )}
        </section>
    );
};

export default React.memo(HeaderHero);
