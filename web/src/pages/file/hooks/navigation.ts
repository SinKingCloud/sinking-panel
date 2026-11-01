import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {history, useLocation} from "umi";
import {getFileDisks} from "@/service/api/file";
import useFileList from "./list";
import {comparableFilePath, isFilePathWithin, isWindowsFilePath, normalizeFilePath} from "../utils";
import {readFileStorage, updateFileStorage} from "./file-storage";

const readLastPath = () => {
    try {
        const value = readFileStorage().lastPath;
        return value ? normalizeFilePath(value) : "";
    } catch {
        return "";
    }
};

const saveLastPath = (path: string) => {
    try {
        updateFileStorage((current) => ({
            ...current,
            lastPath: normalizeFilePath(path),
        }));
    } catch {
        // Storage can be disabled in restricted browser contexts.
    }
};

const clearLastPath = () => {
    try {
        updateFileStorage((current) => {
            const next = {...current};
            delete next.lastPath;
            return next;
        });
    } catch {
        // Storage can be disabled in restricted browser contexts.
    }
};

const getPathFromSearch = (search: string) => (
    normalizeFilePath(new URLSearchParams(search).get("path") || "/")
);

const useFileNavigation = ({onBeforeNavigate, onExternalPathChange}: any) => {
    const {message} = App.useApp();
    const location = useLocation();
    const search = new URLSearchParams(location.search);
    const initialPathRef = useRef(
        search.has("path") ? getPathFromSearch(location.search) : readLastPath(),
    );
    const [disks, setDisks] = useState<string[]>([]);
    const requestedPathRef = useRef(initialPathRef.current || "/");
    const locationPath = search.has("path") ? getPathFromSearch(location.search) : "";
    const currentPath = locationPath || initialPathRef.current;
    // Windows 仅浏览盘符或共享路径，其他路径显示磁盘入口。
    const showingDisks = search.get("view") === "disks" || !currentPath
        || (currentPath === "/" && !disks.includes("/"))
        || (disks.some(isWindowsFilePath) && !isWindowsFilePath(currentPath));
    const list = useFileList(initialPathRef.current || "/", !showingDisks
        && (!locationPath || locationPath === requestedPathRef.current));
    const navigationVersionRef = useRef(0);
    const diskRequestRef = useRef(0);
    const restoredPathRef = useRef("");
    const [disksLoading, setDisksLoading] = useState(true);
    const [disksError, setDisksError] = useState("");
    const [diskGeneration, setDiskGeneration] = useState(0);
    const [keyword, setKeyword] = useState(list.keyword);

    const updateLocation = useCallback((path: string, replace = false) => {
        const normalized = normalizeFilePath(path);
        const search = new URLSearchParams(location.search);
        search.set("path", normalized);
        search.delete("view");
        const next = {pathname: location.pathname, search: `?${search.toString()}`};
        if (replace) {
            history.replace(next);
        } else {
            history.push(next);
        }
    }, [location.pathname, location.search]);

    const navigate = useCallback((path: string, replace = false) => {
        navigationVersionRef.current += 1;
        onBeforeNavigate();
        setKeyword("");
        const normalized = normalizeFilePath(path);
        if (showingDisks) {
            // 先更新真实路径，避免启用列表时请求虚拟层保留的旧目录。
            requestedPathRef.current = normalized;
            list.navigate(normalized);
            updateLocation(normalized, replace);
            return;
        }
        if (normalized === list.requestedPath) {
            if (!new URLSearchParams(location.search).has("path")) {
                updateLocation(normalized, true);
            } else {
                list.reload();
            }
            return;
        }
        updateLocation(normalized, replace);
    }, [list.navigate, list.reload, list.requestedPath, location.search, onBeforeNavigate, showingDisks, updateLocation]);

    useEffect(() => {
        // Unix 类系统的挂载点属于同一目录树，直接从真实根目录进入。
        if (showingDisks && disks.includes("/")) {
            navigate("/", true);
        }
    }, [disks, navigate, showingDisks]);

    const showDisks = useCallback(() => {
        if (showingDisks) {
            return;
        }
        navigationVersionRef.current += 1;
        onBeforeNavigate();
        onExternalPathChange();
        setKeyword("");
        const search = new URLSearchParams(location.search);
        search.set("path", list.path);
        search.set("view", "disks");
        history.push({pathname: location.pathname, search: `?${search.toString()}`});
    }, [list.path, location.pathname, location.search, onBeforeNavigate, onExternalPathChange, showingDisks]);

    useEffect(() => {
        if (!showingDisks) {
            return;
        }
        navigationVersionRef.current += 1;
        onBeforeNavigate();
        onExternalPathChange();
        setKeyword("");
    }, [onBeforeNavigate, onExternalPathChange, showingDisks]);

    useEffect(() => {
        const search = new URLSearchParams(location.search);
        if (!showingDisks || search.get("view") === "disks") {
            return;
        }
        search.set("view", "disks");
        history.replace({pathname: location.pathname, search: `?${search.toString()}`});
    }, [location.pathname, location.search, showingDisks]);

    useEffect(() => {
        const search = new URLSearchParams(location.search);
        if (showingDisks || !search.has("path")) {
            return;
        }
        const nextPath = getPathFromSearch(location.search);
        if (nextPath !== list.requestedPath || nextPath !== requestedPathRef.current) {
            navigationVersionRef.current += 1;
            onBeforeNavigate();
            onExternalPathChange();
            setKeyword("");
            requestedPathRef.current = nextPath;
            list.navigate(nextPath);
        }
    }, [list.navigate, list.requestedPath, location.search, onBeforeNavigate, onExternalPathChange, showingDisks]);

    useEffect(() => {
        if (!showingDisks) {
            setKeyword(list.keyword);
        }
    }, [list.keyword, showingDisks]);

    useEffect(() => {
        if (showingDisks) {
            return;
        }
        const timer = window.setTimeout(() => list.search(keyword), 300);
        return () => window.clearTimeout(timer);
    }, [keyword, list.search, locationPath, showingDisks]);

    const reloadDisks = useCallback(() => {
        diskRequestRef.current += 1;
        setDiskGeneration((value) => value + 1);
    }, []);

    useEffect(() => {
        const requestId = ++diskRequestRef.current;
        let mounted = true;
        const current = () => mounted && diskRequestRef.current === requestId;
        setDisksLoading(true);
        setDisksError("");
        void getFileDisks().then((response) => {
            if (!current()) {
                return;
            }
            if (response?.code !== 200) {
                setDisks([]);
                setDisksError(response?.message || "获取磁盘列表失败");
                return;
            }
            setDisks(Array.from(new Set(
                (Array.isArray(response?.data) ? response?.data : [])
                    .filter((item) => typeof item === "string" && item !== "")
                    .map((item) => normalizeFilePath(item)),
            )));
        }).catch(() => {
            if (current()) {
                setDisks([]);
                setDisksError("获取磁盘列表失败");
            }
        }).finally(() => {
            if (current()) {
                setDisksLoading(false);
            }
        });
        return () => {
            mounted = false;
            if (diskRequestRef.current === requestId) {
                diskRequestRef.current += 1;
            }
        };
    }, [diskGeneration]);

    useEffect(() => {
        if (showingDisks || disks.length === 0 || new URLSearchParams(location.search).has("path")) {
            return;
        }
        const lastPath = readLastPath();
        const nextPath = lastPath && disks.some((disk) => isFilePathWithin(lastPath, disk))
            ? lastPath
            : disks[0];
        restoredPathRef.current = nextPath === lastPath ? comparableFilePath(lastPath) : "";
        navigate(nextPath, true);
    }, [disks, location.search, navigate, showingDisks]);

    useEffect(() => {
        if (showingDisks) {
            return;
        }
        const search = new URLSearchParams(location.search);
        const currentPath = search.has("path") ? getPathFromSearch(location.search) : "";
        if (currentPath === list.path && list.loaded && !list.loading && !list.navigating && !list.error) {
            restoredPathRef.current = "";
            saveLastPath(list.path);
            return;
        }
        if (list.initialError
            && disks.length > 0
            && restoredPathRef.current === comparableFilePath(list.requestedPath)) {
            restoredPathRef.current = "";
            clearLastPath();
            navigate(disks[0], true);
        }
    }, [
        disks,
        list.error,
        list.initialError,
        list.loaded,
        list.loading,
        list.navigating,
        list.path,
        list.requestedPath,
        location.search,
        navigate,
        showingDisks,
    ]);

    useEffect(() => {
        if (!showingDisks && list.error && list.loaded) {
            message.error("文件列表更新失败");
        }
    }, [list.error, list.loaded, message, showingDisks]);

    return {
        list, disks, disksLoading, disksError, showingDisks, showDisks, reloadDisks,
        keyword, setKeyword, navigate, navigationVersionRef,
    };
};

export default useFileNavigation;
