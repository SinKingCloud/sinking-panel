import {useCallback, useEffect, useRef, useState} from "react";
import {getFileList} from "@/service/api/file";
import {
    buildFileBreadcrumbs,
    comparableFilePath as comparablePath,
    isAbsoluteFilePath,
    isFilePathWithin,
    joinFilePath,
    normalizeFilePath,
} from "@/pages/file/utils";

type SelectionOptions = {
    type?: "file" | "directory";
    multiple: boolean;
    disabled?: boolean;
    onError: (message: string) => void;
};

// 单选、多选共用路径草稿；目录联动只在多选且允许目录时启用。
export default function useFileSelection({type, multiple, disabled, onError}: SelectionOptions) {
    const [paths, setPaths] = useState<string[]>([]);
    const [pending, setPending] = useState(false);
    const currentPaths = useRef<string[]>([]);
    const request = useRef(0);
    const busy = useRef(false);
    const activeDirectories = useRef(new Set<string>());
    const directories = useRef(new Map<string, {path: string; children: string[]}>());
    const errorHandler = useRef(onError);
    errorHandler.current = onError;
    const tree = multiple && type !== "file";

    const compact = useCallback((values: string[]) => {
        const unique = new Map(values.map((value) => [isAbsoluteFilePath(value) ? comparablePath(value) : value, value]));
        if (!multiple) return [...unique.values()].slice(0, 1);
        if (!tree) return [...unique.values()];
        // 只根据已完整读取的层级合并，不为计算勾选状态预读目录。
        for (const directory of [...directories.current.values()].sort((a, b) => b.path.length - a.path.length)) {
            if (!directory.children.length || !directory.children.every((child) => unique.has(comparablePath(child)))) continue;
            directory.children.forEach((child) => unique.delete(comparablePath(child)));
            unique.set(comparablePath(directory.path), directory.path);
        }
        return [...unique.values()].filter((value) => !isAbsoluteFilePath(value)
            || !buildFileBreadcrumbs(value).slice(0, -1).some((parent) => unique.has(comparablePath(parent.path))));
    }, [multiple, tree]);

    const commit = useCallback((values: string[]) => {
        currentPaths.current = values;
        setPaths(values);
    }, []);

    const cancel = useCallback(() => {
        request.current++;
        busy.current = false;
        activeDirectories.current.clear();
        directories.current.clear();
        setPending(false);
    }, []);

    const reset = useCallback((values: string[] = []) => {
        cancel();
        commit(compact(values));
    }, [cancel, commit, compact]);

    const replacePath = useCallback((path: string, target?: string, accepts?: (value: string) => boolean) => {
        const source = normalizeFilePath(path);
        reset(currentPaths.current.flatMap((value) => {
            if (!isAbsoluteFilePath(value) || !isFilePathWithin(value, source)) return [value];
            if (!target) return [];
            const next = target + normalizeFilePath(value).slice(source.length);
            return !accepts || accepts(next) ? [next] : [];
        }));
    }, [reset]);

    const remember = useCallback((path: string, items: {name: string; is_dir?: boolean}[]) => {
        if (!tree || !isAbsoluteFilePath(path)) return;
        const children = items.filter((item) => type !== "directory" || item.is_dir).map((item) => joinFilePath(path, item.name));
        directories.current.set(comparablePath(path), {path, children});
        if (!busy.current) commit(compact(currentPaths.current));
    }, [commit, compact, tree, type]);

    const forget = useCallback((path: string) => {
        const key = comparablePath(path);
        directories.current.delete(key);
        // 仅作废已访问变更目录的选择请求，其他目录的后台上传不打断当前选择。
        if (busy.current && activeDirectories.current.has(key)) {
            request.current++;
            busy.current = false;
            activeDirectories.current.clear();
            setPending(false);
        }
    }, []);

    const select = useCallback(async (targets: string[], checked: boolean) => {
        if (disabled || busy.current || !targets.length) return;
        if (checked) {
            commit(compact(multiple ? [...currentPaths.current, ...targets] : targets));
            return;
        }
        const removes = (path: string) => targets.some((target) => tree && isAbsoluteFilePath(path) && isAbsoluteFilePath(target)
            ? isFilePathWithin(path, target) : path === target || (isAbsoluteFilePath(path) && isAbsoluteFilePath(target) && comparablePath(path) === comparablePath(target)));
        const branches = tree ? currentPaths.current.filter((path) => isAbsoluteFilePath(path) && !removes(path)
            && targets.some((target) => isAbsoluteFilePath(target) && isFilePathWithin(target, path))) : [];
        const next = currentPaths.current.filter((path) => !removes(path) && !branches.includes(path));
        if (!branches.length) {
            commit(compact(next));
            return;
        }
        const id = ++request.current;
        busy.current = true;
        activeDirectories.current.clear();
        setPending(true);
        // 请求失败或弹窗关闭时不提交部分选择，也不缓存未完成的目录。
        const loaded = new Map<string, {path: string; children: string[]}>();
        try {
            for (let index = 0; index < branches.length; index++) {
                if (request.current !== id) return;
                const path = branches[index];
                const key = comparablePath(path);
                activeDirectories.current.add(key);
                let directory = loaded.get(key) || directories.current.get(key);
                if (!directory) {
                    const children: string[] = [];
                    let count = 0;
                    let total: number | undefined;
                    for (let page = 1; ; page++) {
                        const response = await getFileList({body: {path, page, page_size: 1000, order_by_field: "name", order_by_type: "asc"}});
                        if (request.current !== id) return;
                        const data = response?.data;
                        if (response?.code !== 200 || !Array.isArray(data?.list)
                            || !Number.isInteger(Number(data?.total)) || Number(data?.total) < 0) {
                            throw new Error(response?.message || "读取目录失败");
                        }
                        if (total !== undefined && total !== Number(data.total)) throw new Error("目录内容已变化，请刷新后重试");
                        total = Number(data.total);
                        count += data.list.length;
                        for (const item of data.list) {
                            if (typeof item?.name !== "string" || !item.name) throw new Error("目录数据不完整，请刷新后重试");
                            if (type !== "directory" || item.is_dir) children.push(joinFilePath(path, item.name));
                        }
                        if (count >= total) break;
                        if (!data.list.length) throw new Error("目录内容已变化，请刷新后重试");
                    }
                    directory = {path, children};
                    loaded.set(key, directory);
                }
                // 仅展开通向反选项的分支，其余兄弟项保留整项选择。
                for (const child of directory.children) {
                    if (removes(child)) continue;
                    if (targets.some((target) => isAbsoluteFilePath(target) && isFilePathWithin(target, child))) branches.push(child);
                    else next.push(child);
                }
            }
            if (request.current === id) {
                loaded.forEach((directory, key) => directories.current.set(key, directory));
                commit(compact(next));
            }
        } catch (reason: any) {
            if (request.current === id) errorHandler.current(reason?.message || "更新选择失败");
        } finally {
            if (request.current === id) {
                busy.current = false;
                activeDirectories.current.clear();
                setPending(false);
            }
        }
    }, [commit, compact, disabled, multiple, tree, type]);

    useEffect(() => {
        cancel();
        commit(compact(currentPaths.current));
        return () => {
            request.current++;
            busy.current = false;
            activeDirectories.current.clear();
            directories.current.clear();
        };
    }, [cancel, commit, compact, disabled, type]);

    return {paths, pending, select, reset, replacePath, cancel, compact, remember, forget};
}
