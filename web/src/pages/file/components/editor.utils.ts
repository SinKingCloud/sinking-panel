import type {FileRecord} from "@/service/api/file";
import {
    buildFileBreadcrumbs,
    comparableFilePath,
    isFilePathWithin,
    joinFilePath,
    normalizeFilePath,
} from "../utils";

export const editorReadPageSize = 100000;
export const editorTreePageSize = 300;
export const maxEditorContentSize = 8 * 1024 * 1024;
const textEncoder = new TextEncoder();

export interface FileEditorTreeNode {
    key: string;
    title: string;
    name: string;
    path: string;
    parentPath: string;
    kind: "entry" | "more";
    isDirectory: boolean;
    isLeaf: boolean;
    record?: FileRecord;
    nextPage?: number;
    remaining?: number;
    children?: FileEditorTreeNode[];
}

const extensionModes: Record<string, string> = {
    bash: "sh",
    bat: "batchfile",
    c: "c_cpp",
    cc: "c_cpp",
    cfg: "ini",
    conf: "ini",
    cpp: "c_cpp",
    cs: "csharp",
    css: "css",
    cts: "typescript",
    cxx: "c_cpp",
    dart: "dart",
    go: "golang",
    graphql: "graphqlschema",
    groovy: "groovy",
    h: "c_cpp",
    hpp: "c_cpp",
    htm: "html",
    html: "html",
    ini: "ini",
    java: "java",
    js: "javascript",
    json: "json",
    jsonc: "json",
    jsx: "jsx",
    kt: "kotlin",
    kts: "kotlin",
    less: "less",
    lua: "lua",
    m: "objectivec",
    markdown: "markdown",
    md: "markdown",
    mjs: "javascript",
    mts: "typescript",
    nginx: "nginx",
    php: "php",
    pl: "perl",
    properties: "properties",
    ps1: "powershell",
    py: "python",
    r: "r",
    rb: "ruby",
    rs: "rust",
    sass: "sass",
    scala: "scala",
    scss: "scss",
    sh: "sh",
    sql: "sql",
    svg: "svg",
    swift: "swift",
    toml: "toml",
    ts: "typescript",
    tsx: "tsx",
    vue: "vue",
    xml: "xml",
    yaml: "yaml",
    yml: "yaml",
    zsh: "sh",
};

export const getEditorMode = (name: string) => {
    const normalized = String(name || "").toLowerCase();
    if (normalized === "dockerfile" || normalized.startsWith("dockerfile.")) {
        return "dockerfile";
    }
    if (normalized === "makefile" || normalized.startsWith("makefile.")) {
        return "makefile";
    }
    if (normalized === ".gitignore") {
        return "gitignore";
    }
    const extension = normalized.includes(".") ? normalized.split(".").pop() || "" : "";
    return extensionModes[extension] || "text";
};

export const contentByteSize = (value: string) => textEncoder.encode(value).byteLength;

export const editorTreeKey = (path: string) => comparableFilePath(path);

export const createEditorRootNode = (path: string): FileEditorTreeNode => {
    const normalized = normalizeFilePath(path);
    const breadcrumbs = buildFileBreadcrumbs(normalized);
    const name = breadcrumbs[breadcrumbs.length - 1]?.label || normalized;
    return {
        key: editorTreeKey(normalized),
        title: name,
        name,
        path: normalized,
        parentPath: normalized,
        kind: "entry",
        isDirectory: true,
        isLeaf: false,
    };
};

export const createEditorEntryNode = (
    parentPath: string,
    record: FileRecord,
): FileEditorTreeNode => {
    const path = joinFilePath(parentPath, record.name);
    return {
        key: editorTreeKey(path),
        title: record.name,
        name: record.name,
        path,
        parentPath,
        kind: "entry",
        isDirectory: record.is_dir,
        isLeaf: !record.is_dir,
        record,
    };
};

export const createEditorMoreNode = (
    parentPath: string,
    nextPage: number,
    remaining: number,
): FileEditorTreeNode => ({
    key: `editor-more:${encodeURIComponent(editorTreeKey(parentPath))}:${nextPage}`,
    title: `加载更多（${remaining}）`,
    name: "加载更多",
    path: parentPath,
    parentPath,
    kind: "more",
    isDirectory: false,
    isLeaf: true,
    nextPage,
    remaining,
});

const preserveLoadedChildren = (
    next: FileEditorTreeNode,
    previous?: FileEditorTreeNode,
) => previous?.children && next.isDirectory
    ? {...next, children: previous.children}
    : next;

const mergeChildren = (
    current: FileEditorTreeNode[] | undefined,
    incoming: FileEditorTreeNode[],
    append: boolean,
) => {
    const existing = (current || []).filter((node) => node.kind !== "more");
    const existingByKey = new Map(existing.map((node) => [node.key, node]));
    const incomingEntries = incoming
        .filter((node) => node.kind !== "more")
        .map((node) => preserveLoadedChildren(node, existingByKey.get(node.key)));
    const more = incoming.find((node) => node.kind === "more");
    if (!append) {
        return more ? [...incomingEntries, more] : incomingEntries;
    }
    const next = [...existing];
    const indexes = new Map(next.map((node, index) => [node.key, index]));
    incomingEntries.forEach((node) => {
        const index = indexes.get(node.key);
        if (index === undefined) {
            indexes.set(node.key, next.length);
            next.push(node);
        } else {
            next[index] = node;
        }
    });
    return more ? [...next, more] : next;
};

export const setEditorTreeChildren = (
    nodes: FileEditorTreeNode[],
    parentKey: string,
    children: FileEditorTreeNode[],
    append = false,
): FileEditorTreeNode[] => {
    let changed = false;
    const next = nodes.map((node) => {
        if (node.key === parentKey) {
            changed = true;
            return {...node, children: mergeChildren(node.children, children, append)};
        }
        if (!node.children) {
            return node;
        }
        const nextChildren = setEditorTreeChildren(node.children, parentKey, children, append);
        if (nextChildren === node.children) {
            return node;
        }
        changed = true;
        return {...node, children: nextChildren};
    });
    return changed ? next : nodes;
};

export const resolveEditorRoot = (path: string, roots: readonly string[]) => {
    const normalized = normalizeFilePath(path);
    const matches = roots
        .map(normalizeFilePath)
        .filter((root) => isFilePathWithin(normalized, root))
        .sort((left, right) => right.length - left.length);
    return matches[0] || buildFileBreadcrumbs(normalized)[0]?.path || "/";
};

export const getEditorAncestorPaths = (path: string, root: string) => {
    const normalizedRoot = normalizeFilePath(root);
    return buildFileBreadcrumbs(path)
        .map((item) => item.path)
        .filter((item) => isFilePathWithin(item, normalizedRoot));
};
