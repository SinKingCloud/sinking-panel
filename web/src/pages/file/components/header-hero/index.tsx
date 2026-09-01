import React, {useMemo} from "react";
import type {MenuProps} from "antd";
import {TableHero} from "@/pages/components/table";
import {normalizeFilePath} from "../../utils";

interface HeaderHeroProps {
    disks: string[];
    onNavigate: (path: string) => void;
}

const HeaderHero = ({disks, onNavigate}: HeaderHeroProps) => {
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

    return <TableHero
        eyebrow="FILE MANAGER"
        title="文件管理"
        action={disks.length > 1 ? {
            label: "切换磁盘",
            ariaLabel: "切换磁盘",
            icon: "DatabaseOutlined",
            suffixIcon: "SwapOutlined",
            menu: diskMenu,
        } : undefined}/>;
};

export default React.memo(HeaderHero);
