import {useCallback, useEffect, useRef, useState} from "react";
import {App} from "antd";
import {history, useLocation} from "umi";
import {getFileDisks} from "@/service/api/file";
import useFileList from "./list";
import {comparableFilePath, isFilePathWithin, normalizeFilePath} from "../utils";

const lastPathStorageKey = "file.last.path";

const readLastPath = () => {
    try {
        window.localStorage.removeItem("file.selected.disk");
        const value = window.localStorage.getItem(lastPathStorageKey);
        return value ? normalizeFilePath(value) : "";
    } catch {
        return "";
    }
};

const saveLastPath = (path: string) => {
    try {
        window.localStorage.setItem(lastPathStorageKey, normalizeFilePath(path));
    } catch {
        // Storage can be disabled in restricted browser contexts.
    }
};

const clearLastPath = () => {
    try {
        window.localStorage.removeItem(lastPathStorageKey);
    } catch {
        // Storage can be disabled in restricted browser contexts.
    }
};

const getPathFromSearch = (search: string) => (
    normalizeFilePath(new URLSearchParams(search).get("path") || "/")
);

export interface UseFileNavigationOptions {
    onBeforeNavigate: () => void;
    onExternalPathChange: () => void;
}

const useFileNavigation = ({onBeforeNavigate, onExternalPathChange}: UseFileNavigationOptions) => {
    const {message} = App.useApp();
    const location = useLocation();
    const initialPathRef = useRef(
        new URLSearchParams(location.search).has("path") ? getPathFromSearch(location.search) : readLastPath() || "/",
    );
    const list = useFileList(initialPathRef.current);
    const navigationVersionRef = useRef(0);
    const diskRequestRef = useRef(0);
    const restoredPathRef = useRef("");
    const [disks, setDisks] = useState<string[]>([]);
    const [keyword, setKeyword] = useState(list.keyword);

    const updateLocation = useCallback((path: string, replace = false) => {
        const normalized = normalizeFilePath(path);
        const search = new URLSearchParams(location.search);
        search.set("path", normalized);
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
        if (normalized === list.requestedPath) {
            if (!new URLSearchParams(location.search).has("path")) {
                updateLocation(normalized, true);
            } else {
                list.reload();
            }
            return;
        }
        updateLocation(normalized, replace);
    }, [list.reload, list.requestedPath, location.search, onBeforeNavigate, updateLocation]);

    useEffect(() => {
        const search = new URLSearchParams(location.search);
        if (!search.has("path")) {
            return;
        }
        const nextPath = getPathFromSearch(location.search);
        if (nextPath !== list.requestedPath) {
            navigationVersionRef.current += 1;
            onBeforeNavigate();
            onExternalPathChange();
            list.navigate(nextPath);
        }
    }, [list.navigate, list.requestedPath, location.search, onBeforeNavigate, onExternalPathChange]);

    useEffect(() => {
        setKeyword(list.keyword);
    }, [list.keyword]);

    useEffect(() => {
        const timer = window.setTimeout(() => list.search(keyword), 300);
        return () => window.clearTimeout(timer);
    }, [keyword, list.search]);

    useEffect(() => {
        const requestId = ++diskRequestRef.current;
        let active = true;
        void getFileDisks().then((response) => {
            if (!active || diskRequestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200) {
                message.warning(response?.message || "获取磁盘列表失败");
                return;
            }
            setDisks(Array.from(new Set(
                (Array.isArray(response.data) ? response.data : [])
                    .map((item) => normalizeFilePath(item))
                    .filter(Boolean),
            )));
        });
        return () => {
            active = false;
            if (diskRequestRef.current === requestId) {
                diskRequestRef.current += 1;
            }
        };
    }, [message]);

    useEffect(() => {
        if (disks.length === 0 || new URLSearchParams(location.search).has("path")) {
            return;
        }
        const lastPath = readLastPath();
        const nextPath = lastPath && disks.some((disk) => isFilePathWithin(lastPath, disk))
            ? lastPath
            : disks[0];
        restoredPathRef.current = nextPath === lastPath ? comparableFilePath(lastPath) : "";
        navigate(nextPath, true);
    }, [disks, location.search, navigate]);

    useEffect(() => {
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
    ]);

    useEffect(() => {
        if (list.error && list.loaded) {
            message.error("文件列表更新失败");
        }
    }, [list.error, list.loaded, message]);

    return {list, disks, keyword, setKeyword, navigate, navigationVersionRef};
};

export default useFileNavigation;
