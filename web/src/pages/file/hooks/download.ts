import {useCallback} from "react";
import {App} from "antd";
import defaultSettings from "@/../config/defaultSettings";
import {getFileSign} from "@/service/api/file";
import type {FileRecord} from "@/service/api/file";
import {joinFilePath} from "../utils";

const apiUrl = (path: string) => {
    const origin = window.location.origin;
    const gateway = new URL(defaultSettings?.gateway || "/", origin);
    const gatewayPath = gateway.pathname.replace(/\/+$/, "");
    gateway.pathname = `${gatewayPath}/${path.replace(/^\/+/, "")}`.replace(/\/{2,}/g, "/");
    gateway.search = "";
    gateway.hash = "";
    return gateway.toString();
};

const useFileDownload = (path: string) => {
    const {message} = App.useApp();

    return useCallback(async (record: FileRecord) => {
        const response = await getFileSign({
            body: {path: joinFilePath(path, record.name), download: true},
        });
        if (!response) {
            return;
        }
        if (response.code !== 200 || typeof response.data !== "string" || !response.data) {
            message.error(response.message || "获取文件下载地址失败");
            return;
        }
        try {
            const url = new URL(apiUrl("/preview"));
            url.searchParams.set("key", response.data);
            const anchor = document.createElement("a");
            anchor.href = url.toString();
            anchor.download = record.name;
            anchor.rel = "noopener noreferrer";
            anchor.referrerPolicy = "no-referrer";
            anchor.style.display = "none";
            document.body.appendChild(anchor);
            try {
                anchor.click();
            } finally {
                anchor.remove();
            }
        } catch {
            message.error("下载文件失败");
        }
    }, [message, path]);
};

export default useFileDownload;
