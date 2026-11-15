import {isAbsoluteFilePath} from "@/pages/file/utils";

export type BackupType = "site" | "file";

type BackupOptions = {type: "site"; options: {all_sites: boolean; site_ids: number[]}}
    | {type: "file"; options: {paths: string[]}};
type BackupConfig = BackupOptions & {
    storage: string;
    local_keep: number;
    oss_keep: number;
    oss_id: number;
    prefix: string;
};

// validateBackupPaths 保留路径中的空格，只去重完全相同的路径。
export const validateBackupPaths = (paths: unknown): string[] => {
    if (!Array.isArray(paths) || paths.length === 0
        || paths.some((path) => typeof path !== "string" || !path.trim())) {
        throw new Error("请至少选择一个文件或目录");
    }
    if (paths.length > 1000) {
        throw new Error("最多选择 1000 个文件或目录");
    }
    if (paths.some((path) => !isAbsoluteFilePath(path) || path.includes("\0"))) {
        throw new Error("请输入有效的文件或目录绝对路径");
    }
    return [...new Set<string>(paths)];
};

// parseBackup 按备份类型读取参数，只保留当前类型的内容。
export const parseBackup = (value?: any): BackupConfig => {
    const data = value === undefined
        ? {type: "site", options: {all_sites: true, site_ids: []}, storage: "local", local_keep: 7, oss_keep: 7, oss_id: 0, prefix: "backups"}
        : typeof value === "string" ? JSON.parse(value) : value;
    if (!data || typeof data !== "object" || Array.isArray(data)) {
        throw new Error("备份配置格式不正确");
    }
    if (!["site", "file"].includes(data.type)) {
        throw new Error("不支持的备份内容");
    }
    const options = data.options;
    if (!options || typeof options !== "object" || Array.isArray(options)) {
        throw new Error("备份参数格式不正确");
    }
    let target: BackupOptions;
    if (data.type === "site") {
        if (typeof options.all_sites !== "boolean" || !Array.isArray(options.site_ids)) {
            throw new Error("网站备份参数格式不正确");
        }
        target = {type: "site", options: {all_sites: options.all_sites, site_ids: options.site_ids.map(Number)}};
    } else {
        if (!Array.isArray(options.paths) || options.paths.some((path: unknown) => typeof path !== "string")) {
            throw new Error("路径备份参数格式不正确");
        }
        target = {type: "file", options: {paths: [...options.paths]}};
    }
    return {
        ...target,
        storage: String(data.storage || "local"),
        local_keep: Number(data.local_keep),
        oss_keep: Number(data.oss_keep),
        oss_id: Number(data.oss_id || 0),
        prefix: String(data.prefix ?? "").replace(/^\/+|\/+$/g, ""),
    };
};

// stringifyBackup 只序列化备份参数，周期及启停状态由计划任务统一保存。
export const stringifyBackup = (value: any): string => {
    const data = parseBackup(value);
    if (data.type === "site") {
        const options = data.options;
        options.site_ids = options.all_sites ? [] : [...new Set<number>(options.site_ids)];
        if (!options.all_sites && (options.site_ids.length === 0 || options.site_ids.some((id) => !Number.isInteger(id) || id <= 0))) {
            throw new Error("请选择要备份的网站");
        }
    } else {
        data.options.paths = validateBackupPaths(data.options.paths);
    }
    if (data.storage !== "local" && data.storage !== "oss") {
        throw new Error("请选择存储方式");
    }
    if (!Number.isInteger(data.local_keep) || data.local_keep < (data.storage === "oss" ? 0 : 1) || data.local_keep > 1000) {
        throw new Error(data.storage === "oss" ? "本地保留副本须为 0 到 1000 份" : "本地保留副本须为 1 到 1000 份");
    }
    if (data.storage === "oss" && (!Number.isInteger(data.oss_id) || data.oss_id <= 0)) {
        throw new Error("请选择存储实例");
    }
    if (data.storage === "oss" && (!Number.isInteger(data.oss_keep) || data.oss_keep < 1 || data.oss_keep > 1000)) {
        throw new Error("云端保留副本须为 1 到 1000 份");
    }
    if (data.prefix.length > 512 || /[\\\x00\r\n]/.test(data.prefix) || data.prefix.split("/").some((part) => part === "." || part === "..")) {
        throw new Error("备份目录格式不正确");
    }
    if (data.storage === "local") {
        data.oss_id = 0;
        data.oss_keep = 0;
        data.prefix = "";
    }
    return JSON.stringify(data);
};
