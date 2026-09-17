import React from "react";
import {Icon} from "sinking-antd";
import type {TableProps} from "@/components/table";

const useSiteHeader = ({
    keyword,
    searchField,
    typeId,
    type,
    status,
    typeItems,
    typeData,
    statusData,
    refreshing,
    onKeywordChange,
    onSearchFieldChange,
    onSearch,
    onTypeIdChange,
    onManageTypes,
    onTypeChange,
    onStatusChange,
    onRefresh,
    onCreate,
    onHTTPSettings,
    onManageCert,
    onManageSecret,
}: any): Pick<TableProps, "hero" | "toolbar"> => {
    const categoryOptions = React.useMemo(() => [
        {label: "全部分类", value: "0"},
        ...(typeItems || []).map((item: any) => ({value: String(item.id), label: String(item.name)})),
    ], [typeItems]);
    const typeOptions = React.useMemo(() => [
        {label: "全部类型", value: "all"},
        ...Object.entries(typeData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [typeData]);
    const statusOptions = React.useMemo(() => [
        {label: "全部状态", value: "all"},
        ...Object.entries(statusData || {}).map(([value, label]) => ({value, label: String(label)})),
    ], [statusData]);

    return {
        hero: {
            eyebrow: "WEBSITE MANAGER",
            title: "网站管理",
            action: {label: "添加网站", onClick: onCreate},
        },
        toolbar: {
            search: {
                value: keyword,
                field: searchField,
                maxLength: 100,
                fields: [
                    {value: "keyword", label: "全部", placeholder: "搜索名称或目录"},
                    {value: "name", label: "名称", placeholder: "搜索网站名称"},
                    {value: "root", label: "目录", placeholder: "搜索网站目录"},
                ],
                onChange: onKeywordChange,
                onFieldChange: onSearchFieldChange,
                onSearch,
            },
            refresh: {loading: refreshing, ariaLabel: "刷新网站列表", onClick: onRefresh},
            actions: [
                {
                    key: "category",
                    type: "filter",
                    value: typeId || "0",
                    options: categoryOptions,
                    icon: "FolderOutlined",
                    ariaLabel: "网站分类筛选",
                    matchWidth: true,
                    command: {label: "分类管理", icon: "SettingOutlined", onClick: onManageTypes},
                    onChange: onTypeIdChange,
                },
                {
                    key: "type",
                    type: "filter",
                    value: type || "all",
                    options: typeOptions,
                    icon: "AppstoreOutlined",
                    ariaLabel: "网站类型筛选",
                    matchWidth: true,
                    onChange: (value) => onTypeChange(value === "all" ? "" : value),
                },
                {
                    key: "status",
                    type: "filter",
                    value: status || "all",
                    options: statusOptions,
                    icon: "FlagOutlined",
                    ariaLabel: "网站状态筛选",
                    matchWidth: true,
                    onChange: (value) => onStatusChange(value === "all" ? "" : value),
                },
                {
                    key: "settings",
                    type: "button",
                    label: "服务管理",
                    icon: "SettingOutlined",
                    suffixIcon: "DownOutlined",
                    menu: {
                        items: [
                            {key: "cert", label: "证书管理", icon: <Icon type="SafetyCertificateOutlined"/>, onClick: onManageCert},
                            {key: "secret", label: "密钥管理", icon: <Icon type="KeyOutlined"/>, onClick: onManageSecret},
                            ...(onHTTPSettings ? [{key: "http", label: "服务设置", icon: <Icon type="SettingOutlined"/>, onClick: onHTTPSettings}] : []),
                        ],
                    },
                },
            ],
        },
    };
};

export default useSiteHeader;
