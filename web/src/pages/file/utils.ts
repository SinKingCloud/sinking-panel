import dayjs from "dayjs";

// parsePath 保留盘符或共享卷作为根，避免上级导航越过 UNC 共享边界。
const parsePath = (value: string) => {
    const raw = String(value || "");
    const windows = /^[a-zA-Z]:(?:[\\/]|$)/.test(raw) || /^(?:\\\\|\/\/)[^\\/]/.test(raw);
    const source = windows ? raw.replace(/\\/g, "/") : raw;
    const drive = source.match(/^(\/\/\?\/)?([a-zA-Z]):(?:\/+|$)/);
    const share = !drive && source.match(/^(\/\/(?:\?\/UNC\/)?[^/]+\/[^/]+)(?:\/+|$)/i);
    const root = drive ? `${drive[1] || ""}${drive[2].toUpperCase()}:/` : share ? `${share[1]}/` : "/";
    const remainder = drive ? source.slice(drive[0].length)
        : share ? source.slice(share[0].length) : source.replace(/^\/+/, "");
    const parts: string[] = [];
    remainder.split(/\/+/).forEach((part) => {
        if (!part || part === ".") {
            return;
        }
        if (part === "..") {
            parts.pop();
            return;
        }
        parts.push(part);
    });
    return {root, parts};
};

// normalizeFilePath 统一目录分隔符，保留 Windows 盘符和 UNC 共享根。
export const normalizeFilePath = (value: string) => {
    const {root, parts} = parsePath(value);
    return parts.length > 0 ? `${root}${parts.join("/")}` : root;
};

export const parentFilePath = (value: string) => {
    const {root, parts} = parsePath(value);
    parts.pop();
    return `${root}${parts.join("/")}`;
};

export const buildFileBreadcrumbs = (value: string): any[] => {
    const {root, parts} = parsePath(value);
    const result: any[] = [{label: root === "/" ? root : root.slice(0, -1), path: root}];
    parts.forEach((part, index) => {
        result.push({
            label: part,
            path: `${root}${parts.slice(0, index + 1).join("/")}`,
        });
    });
    return result;
};

// getFileDiskName 返回磁盘目录名称，实际文件操作仍使用原始路径。
export const getFileDiskName = (value: string) => {
    const path = normalizeFilePath(value);
    const {root} = parsePath(path);
    if (path === "/") return "/";
    if (/^[A-Z]:\/$/.test(path)) return `磁盘(${path.slice(0, 2)})`;
    if (path === root) return root.slice(0, -1);
    return path.slice(path.lastIndexOf("/") + 1);
};

// isWindowsFilePath 识别盘符和网络共享，避免把 UNC 路径当作 Unix 根目录。
export const isWindowsFilePath = (value: string) => parsePath(value).root !== "/";

const categoryExtensions: Partial<Record<string, Set<string>>> = {
    archive: new Set(["7z", "bz", "bz2", "gz", "rar", "tar", "tgz", "xz", "zip"]),
    image: new Set(["avif", "bmp", "gif", "ico", "jpeg", "jpg", "png", "svg", "webp"]),
    video: new Set(["avi", "flv", "m4v", "mkv", "mov", "mp4", "mpeg", "mpg", "webm", "wmv"]),
    audio: new Set(["aac", "flac", "m4a", "mp3", "ogg", "wav", "wma"]),
    pdf: new Set(["pdf"]),
    document: new Set(["doc", "docx", "odt", "rtf"]),
    spreadsheet: new Set(["csv", "ods", "xls", "xlsx"]),
    presentation: new Set(["odp", "ppt", "pptx"]),
    code: new Set([
        "c", "cc", "conf", "cpp", "css", "env", "go", "h", "hpp", "html", "ini", "java", "js", "json",
        "jsx", "less", "lua", "php", "properties", "py", "rb", "rs", "scss", "sh", "sql", "toml", "ts",
        "tsx", "vue", "xml", "yaml", "yml",
    ]),
    text: new Set(["log", "md", "text", "txt"]),
};

const categoryByExtension = new Map<string, string>();
(Object.entries(categoryExtensions) as [string, Set<string>][]).forEach(([category, extensions]) => {
    extensions.forEach((extension) => categoryByExtension.set(extension, category));
});

const categoryIcons: Record<string, string> = {
    folder: "FolderOutlined",
    archive: "FileZipOutlined",
    image: "FileImageOutlined",
    video: "VideoCameraOutlined",
    audio: "AudioOutlined",
    pdf: "FilePdfOutlined",
    document: "FileWordOutlined",
    spreadsheet: "FileExcelOutlined",
    presentation: "FilePptOutlined",
    code: "CodeOutlined",
    text: "FileTextOutlined",
    file: "FileOutlined",
};

export const comparableFilePath = (value: string) => {
    const normalized = normalizeFilePath(value);
    return isWindowsFilePath(normalized) ? normalized.toLowerCase() : normalized;
};

export const isFilePathWithin = (path: string, root: string) => {
    const current = comparableFilePath(path);
    const parent = comparableFilePath(root);
    if (parsePath(current).root !== parsePath(parent).root) return false;
    return current === parent || current.startsWith(parent.endsWith("/") ? parent : `${parent}/`);
};

export const joinFilePath = (base: string, name: string) => {
    const parent = normalizeFilePath(base);
    const child = String(name ?? "");
    return `${parent}${parent.endsWith("/") ? "" : "/"}${child}`;
};

// getFileNavigation 统一文件管理和选择器的目录层级，空路径表示 Windows 磁盘列表。
export const getFileNavigation = (path: string, roots: readonly string[]) => {
    const disks = {label: "/", path: ""};
    if (!path) return {breadcrumbs: [disks], parentPath: undefined};
    const current = normalizeFilePath(path);
    const all = buildFileBreadcrumbs(current);
    if (!isWindowsFilePath(current)) {
        return {breadcrumbs: all, parentPath: current === "/" ? undefined : parentFilePath(current)};
    }
    const root = roots.map(normalizeFilePath)
        .filter((item) => isFilePathWithin(current, item))
        .sort((left, right) => right.length - left.length)[0] || all[0].path;
    const start = all.findIndex((item) => comparableFilePath(item.path) === comparableFilePath(root));
    return {
        breadcrumbs: [disks, {label: getFileDiskName(root), path: root}, ...all.slice(start + 1)],
        parentPath: comparableFilePath(current) === comparableFilePath(root) ? "" : parentFilePath(current),
    };
};

export const formatFileSize = (size: unknown) => {
    const value = Number(size || 0);
    if (!Number.isFinite(value) || value <= 0) {
        return "0 B";
    }
    const units = ["B", "KB", "MB", "GB", "TB", "PB"];
    let amount = value;
    let index = 0;
    while (amount >= 1024 && index < units.length - 1) {
        amount /= 1024;
        index += 1;
    }
    return `${amount.toLocaleString("zh-CN", {
        minimumFractionDigits: 0,
        maximumFractionDigits: index === 0 ? 0 : 2,
    })} ${units[index]}`;
};

export const formatFileMode = (mode: unknown) => {
    const value = Number(mode);
    if (!Number.isFinite(value) || value < 0) {
        return "-";
    }
    return (Math.trunc(value) & 0o777).toString(8).padStart(4, "0");
};

export const formatFileTime = (value: unknown) => {
    if (value === undefined || value === null || value === "") {
        return "-";
    }
    const numeric = Number(value);
    const time = Number.isFinite(numeric)
        ? (numeric > 1_000_000_000_000 ? dayjs(numeric) : dayjs.unix(numeric))
        : dayjs(String(value));
    return time.isValid() ? time.format("YYYY-MM-DD HH:mm:ss") : "-";
};

export const getFileExtension = (name: string) => {
    const value = String(name || "").toLocaleLowerCase();
    const index = value.lastIndexOf(".");
    return index > 0 && index < value.length - 1 ? value.slice(index + 1) : "";
};

export const isAbsoluteFilePath = (value: string) => {
    const path = String(value || "");
    return path.startsWith("/") || /^[a-zA-Z]:[\\/]/.test(path)
        || (path.startsWith("\\\\") && isWindowsFilePath(path));
};

export const getFileCategory = (name: string, isDir = false): string => {
    if (isDir) {
        return "folder";
    }
    return categoryByExtension.get(getFileExtension(name)) || "file";
};

export const getFileIconType = (name: string, isDir = false) => {
    return categoryIcons[getFileCategory(name, isDir)];
};

export const isExtractableFile = (name: string) => {
    const value = String(name || "").toLowerCase();
    return value.endsWith(".zip") || value.endsWith(".tar") || value.endsWith(".gz") ||
        value.endsWith(".gzip") || value.endsWith(".tgz") || value.endsWith(".tar.gz");
};
