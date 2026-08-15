import type {FileRecord} from "@/service/api/file";
import {isAbsoluteFilePath} from "../utils";

export const validatePath = async (_: unknown, value: unknown) => {
    const path = String(value ?? "").trim();
    if (!path) {
        return;
    }
    if (!isAbsoluteFilePath(path)) {
        throw new Error("请输入绝对路径");
    }
};

export const validateFileName = async (_: unknown, value: unknown) => {
    const name = String(value ?? "").trim();
    if (!name) {
        return;
    }
    if (name === "." || name === ".." || /[\\/\0]/.test(name)) {
        throw new Error("文件名称不能包含路径分隔符");
    }
    if (name.length > 255) {
        throw new Error("文件名称不能超过255个字符");
    }
};

export const validateUrl = async (_: unknown, value: unknown) => {
    const input = String(value ?? "").trim();
    if (!input) {
        return;
    }
    try {
        const url = new URL(input);
        if (url.protocol !== "http:" && url.protocol !== "https:") {
            throw new Error();
        }
    } catch {
        throw new Error("请输入有效的 HTTP 或 HTTPS 文件地址");
    }
};

export const getRemoteFileName = (value: unknown) => {
    const input = String(value ?? "").trim();
    if (!input) {
        return "";
    }
    try {
        const url = new URL(input);
        if (url.protocol !== "http:" && url.protocol !== "https:") {
            return "";
        }
        const segment = url.pathname.split("/").filter(Boolean).pop();
        if (!segment) {
            return "";
        }
        const name = decodeURIComponent(segment).trim();
        if (!name || name === "." || name === ".." || /[\\/\0]/.test(name) || name.length > 255) {
            return "";
        }
        return name;
    } catch {
        return "";
    }
};

export const defaultArchiveName = (records: FileRecord[]) => {
    if (records.length !== 1) {
        return "archive";
    }
    const name = records[0].name.replace(/\.(?:tar\.gz|tgz|zip|tar|gz)$/i, "");
    return name || "archive";
};
