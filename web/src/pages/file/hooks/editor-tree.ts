import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App} from "antd";
import {getFileList} from "@/service/api/file";
import {
    createEditorEntryNode,
    createEditorMoreNode,
    createEditorRootNode,
    editorTreeKey,
    editorTreePageSize,
    getEditorAncestorPaths,
    resolveEditorRoot,
    setEditorTreeChildren,
} from "../components/editor.utils";
import type {FileEditorTreeNode} from "../components/editor.utils";
import {normalizeFilePath, parentFilePath} from "../utils";

export interface UseFileEditorTreeOptions {
    roots: readonly string[];
    message: ReturnType<typeof App.useApp>["message"];
}

type DirectoryLoadResult = {
    status: "success";
    entries: FileEditorTreeNode[];
    hasMore: boolean;
} | {
    status: "failed" | "stale";
};

interface PendingDirectoryLoad {
    page: number;
    append: boolean;
    generation: number;
    requestId: number;
    promise: Promise<DirectoryLoadResult>;
}

const useFileEditorTree = ({roots, message}: UseFileEditorTreeOptions) => {
    const generationRef = useRef(0);
    const selectionEpochRef = useRef(0);
    const expansionEpochRef = useRef(0);
    const requestSequenceRef = useRef(0);
    const directoryRequestsRef = useRef(new Map<string, number>());
    const pendingLoadsRef = useRef(new Map<string, PendingDirectoryLoad>());
    const [treeData, setTreeData] = useState<FileEditorTreeNode[]>([]);
    const [expandedKeys, setExpandedKeys] = useState<string[]>([]);
    const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
    const [targetDirectory, setTargetDirectory] = useState("");
    const [loadingPaths, setLoadingPaths] = useState<ReadonlySet<string>>(new Set());
    const [initializing, setInitializing] = useState(false);

    const normalizedRoots = useMemo(() => roots
        .map(normalizeFilePath)
        .filter((root, index, values) => values.indexOf(root) === index), [roots]);
    const rootKeys = useMemo(() => new Set(normalizedRoots.map(editorTreeKey)), [normalizedRoots]);
    const rootKeysRef = useRef(rootKeys);
    rootKeysRef.current = rootKeys;

    const loadDirectory = useCallback((
        path: string,
        page = 1,
        append = false,
        generation = generationRef.current,
        force = false,
    ): Promise<DirectoryLoadResult> => {
        const normalizedPath = normalizeFilePath(path);
        const parentKey = editorTreeKey(normalizedPath);
        const pending = pendingLoadsRef.current.get(parentKey);
        if (
            !force &&
            pending?.page === page &&
            pending.append === append &&
            pending.generation === generation
        ) {
            return pending.promise;
        }
        const requestId = ++requestSequenceRef.current;
        directoryRequestsRef.current.set(parentKey, requestId);
        setLoadingPaths((current) => new Set(current).add(parentKey));
        const promise = (async (): Promise<DirectoryLoadResult> => {
            try {
                const response = await getFileList({body: {
                    path: normalizedPath,
                    page,
                    page_size: editorTreePageSize,
                    order_by_field: "name",
                    order_by_type: "asc",
                }});
                if (
                    generationRef.current !== generation ||
                    directoryRequestsRef.current.get(parentKey) !== requestId
                ) {
                    return {status: "stale"};
                }
                if (response?.code !== 200 || !response.data || !Array.isArray(response.data.list)) {
                    message.error(response?.message || `读取目录失败：${normalizedPath}`);
                    return {status: "failed"};
                }
                const data = response.data;
                const children = data.list
                    .map((record) => createEditorEntryNode(normalizedPath, record))
                    .filter((node) => !rootKeysRef.current.has(node.key));
                const loaded = page * editorTreePageSize;
                const hasMore = loaded < Number(data.total || 0);
                if (hasMore) {
                    children.push(createEditorMoreNode(
                        normalizedPath,
                        page + 1,
                        Math.max(0, Number(data.total || 0) - loaded),
                    ));
                }
                setTreeData((current) => setEditorTreeChildren(
                    current,
                    parentKey,
                    children,
                    append,
                ));
                return {
                    status: "success",
                    entries: children.filter((node) => node.kind === "entry"),
                    hasMore,
                };
            } catch (reason: unknown) {
                if (
                    generationRef.current !== generation ||
                    directoryRequestsRef.current.get(parentKey) !== requestId
                ) {
                    return {status: "stale"};
                }
                message.error(reason instanceof Error && reason.message
                    ? reason.message
                    : `读取目录失败：${normalizedPath}`);
                return {status: "failed"};
            } finally {
                if (directoryRequestsRef.current.get(parentKey) === requestId) {
                    directoryRequestsRef.current.delete(parentKey);
                    setLoadingPaths((current) => {
                        const next = new Set(current);
                        next.delete(parentKey);
                        return next;
                    });
                }
                if (pendingLoadsRef.current.get(parentKey)?.requestId === requestId) {
                    pendingLoadsRef.current.delete(parentKey);
                }
            }
        })();
        pendingLoadsRef.current.set(parentKey, {page, append, generation, requestId, promise});
        return promise;
    }, [message]);

    useEffect(() => () => {
        generationRef.current += 1;
        directoryRequestsRef.current.clear();
        pendingLoadsRef.current.clear();
    }, []);

    useEffect(() => {
        if (normalizedRoots.length === 0) {
            return;
        }
        setTreeData((current) => {
            if (current.length === 0) {
                return current;
            }
            const findNode = (nodes: FileEditorTreeNode[], key: string): FileEditorTreeNode | undefined => {
                for (const node of nodes) {
                    if (node.key === key) {
                        return node;
                    }
                    const child = node.children ? findNode(node.children, key) : undefined;
                    if (child) {
                        return child;
                    }
                }
                return undefined;
            };
            const nextRoots = normalizedRoots.map((path) => {
                const root = createEditorRootNode(path);
                const previous = findNode(current, root.key);
                return previous?.children ? {...root, children: previous.children} : root;
            });
            const nextRootKeys = new Set(nextRoots.map((node) => node.key));
            current.forEach((node) => {
                if (!nextRootKeys.has(node.key)) {
                    nextRootKeys.add(node.key);
                    nextRoots.push(node);
                }
            });
            const pruneNestedRoots = (node: FileEditorTreeNode): FileEditorTreeNode => {
                if (!node.children) {
                    return node;
                }
                const children = node.children
                    .filter((child) => !nextRootKeys.has(child.key))
                    .map(pruneNestedRoots);
                return children.length === node.children.length && children.every((child, index) => child === node.children?.[index])
                    ? node
                    : {...node, children};
            };
            return nextRoots.map(pruneNestedRoots);
        });
    }, [normalizedRoots]);

    const loadedKeys = useMemo(() => {
        const keys: string[] = [];
        const collect = (nodes: FileEditorTreeNode[]) => {
            nodes.forEach((node) => {
                if (node.children !== undefined) {
                    keys.push(node.key);
                    collect(node.children);
                }
            });
        };
        collect(treeData);
        return keys;
    }, [treeData]);

    const initialize = useCallback((directoryPath: string, filePath?: string) => {
        const generation = ++generationRef.current;
        selectionEpochRef.current += 1;
        const expansionEpoch = ++expansionEpochRef.current;
        directoryRequestsRef.current.clear();
        pendingLoadsRef.current.clear();
        const directory = normalizeFilePath(directoryPath);
        let availableRoots = normalizedRoots.length > 0
            ? normalizedRoots
            : [resolveEditorRoot(directory, [])];
        const root = resolveEditorRoot(directory, availableRoots);
        if (!availableRoots.some((item) => editorTreeKey(item) === editorTreeKey(root))) {
            availableRoots = [...availableRoots, root];
        }
        const ancestors = getEditorAncestorPaths(directory, root);
        setTreeData(availableRoots.map(createEditorRootNode));
        setExpandedKeys([]);
        setSelectedKeys([editorTreeKey(filePath || directory)]);
        setTargetDirectory(directory);
        setLoadingPaths(new Set());
        setInitializing(true);

        void (async () => {
            try {
                for (let index = 0; index < ancestors.length; index += 1) {
                    if (generationRef.current !== generation) {
                        return;
                    }
                    const ancestor = ancestors[index];
                    const target = ancestors[index + 1] || filePath;
                    let page = 1;
                    while (true) {
                        const result = await loadDirectory(ancestor, page, page > 1, generation);
                        if (result.status !== "success" || generationRef.current !== generation) {
                            return;
                        }
                        if (!target || result.entries.some((node) => node.key === editorTreeKey(target)) || !result.hasMore) {
                            break;
                        }
                        page += 1;
                    }
                }
                if (
                    generationRef.current !== generation ||
                    expansionEpochRef.current !== expansionEpoch
                ) {
                    return;
                }
                setExpandedKeys((current) => Array.from(new Set([
                    ...current,
                    ...ancestors.map(editorTreeKey),
                ])));
            } finally {
                if (generationRef.current === generation) {
                    setInitializing(false);
                }
            }
        })();
    }, [loadDirectory, normalizedRoots]);

    const close = useCallback(() => {
        generationRef.current += 1;
        selectionEpochRef.current += 1;
        expansionEpochRef.current += 1;
        directoryRequestsRef.current.clear();
        pendingLoadsRef.current.clear();
        setTreeData([]);
        setExpandedKeys([]);
        setSelectedKeys([]);
        setTargetDirectory("");
        setLoadingPaths(new Set());
        setInitializing(false);
    }, []);

    const loadData = useCallback(async (node: FileEditorTreeNode) => {
        if (node.kind !== "entry" || !node.isDirectory || node.children) {
            return;
        }
        const result = await loadDirectory(node.path);
        if (result.status === "failed") {
            expansionEpochRef.current += 1;
            setExpandedKeys((current) => current.filter((key) => key !== node.key));
        }
    }, [loadDirectory]);

    const selectDirectory = useCallback((node: FileEditorTreeNode) => {
        selectionEpochRef.current += 1;
        setSelectedKeys([node.key]);
        setTargetDirectory(node.path);
    }, []);

    const selectFile = useCallback((node: FileEditorTreeNode) => {
        selectionEpochRef.current += 1;
        setSelectedKeys([node.key]);
        setTargetDirectory(node.parentPath);
    }, []);

    const selectPath = useCallback((path: string, isDirectory = false) => {
        const normalizedPath = normalizeFilePath(path);
        selectionEpochRef.current += 1;
        setSelectedKeys([editorTreeKey(normalizedPath)]);
        setTargetDirectory(isDirectory ? normalizedPath : parentFilePath(normalizedPath));
    }, []);

    const loadMore = useCallback((node: FileEditorTreeNode) => {
        if (node.kind !== "more" || !node.nextPage) {
            return;
        }
        const parentKey = editorTreeKey(node.parentPath);
        if (loadingPaths.has(parentKey)) {
            return;
        }
        void loadDirectory(node.parentPath, node.nextPage, true);
    }, [loadDirectory, loadingPaths]);

    const refresh = useCallback((path = targetDirectory) => {
        if (!path) {
            return Promise.resolve(false);
        }
        return loadDirectory(path, 1, false).then((result) => result.status === "success");
    }, [loadDirectory, targetDirectory]);

    const revealCreated = useCallback(async (
        parentPath: string,
        targetPath: string,
        isDirectory: boolean,
        selectTarget = true,
    ) => {
        const selectionEpoch = selectionEpochRef.current;
        const expansionEpoch = expansionEpochRef.current;
        let page = 1;
        while (true) {
            const result = await loadDirectory(parentPath, page, page > 1, generationRef.current, true);
            if (result.status !== "success") {
                return;
            }
            if (result.entries.some((node) => node.key === editorTreeKey(targetPath)) || !result.hasMore) {
                break;
            }
            page += 1;
        }
        if (expansionEpochRef.current !== expansionEpoch) {
            return;
        }
        setExpandedKeys((current) => Array.from(new Set([...current, editorTreeKey(parentPath)])));
        if (!selectTarget || selectionEpochRef.current !== selectionEpoch) {
            return;
        }
        const targetKey = editorTreeKey(targetPath);
        setSelectedKeys([targetKey]);
        setTargetDirectory(isDirectory ? targetPath : parentPath);
    }, [loadDirectory]);

    const changeExpandedKeys = useCallback((keys: string[]) => {
        expansionEpochRef.current += 1;
        setExpandedKeys(keys);
    }, []);

    return {
        treeData,
        expandedKeys,
        selectedKeys,
        loadedKeys,
        loadingPaths,
        initializing,
        targetDirectory,
        setExpandedKeys: changeExpandedKeys,
        initialize,
        close,
        loadData,
        selectDirectory,
        selectFile,
        selectPath,
        loadMore,
        refresh,
        revealCreated,
    };
};

export default useFileEditorTree;
