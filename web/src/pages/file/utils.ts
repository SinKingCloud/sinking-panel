import dayjs from "dayjs";

export interface FileBreadcrumbItem {
    label: string;
    path: string;
}

export type FileCategory =
    | "folder"
    | "archive"
    | "image"
    | "video"
    | "audio"
    | "pdf"
    | "document"
    | "spreadsheet"
    | "presentation"
    | "code"
    | "text"
    | "file";

const categoryExtensions: Partial<Record<FileCategory, Set<string>>> = {
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

const categoryByExtension = new Map<string, FileCategory>();
(Object.entries(categoryExtensions) as [FileCategory, Set<string>][]).forEach(([category, extensions]) => {
    extensions.forEach((extension) => categoryByExtension.set(extension, category));
});

const previewableExtensions = new Set([
    "bmp", "gif", "jpeg", "jpg", "mp3", "mp4", "pdf", "png", "svg", "webm", "webp",
]);

const categoryIcons: Record<FileCategory, string> = {
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

const parsePath = (value: string) => {
    const raw = String(value || "");
    const driveMatch = raw.match(/^([a-zA-Z]):(?:[\\/]+|$)/);
    const source = driveMatch ? raw.replace(/\\/g, "/") : raw;
    const drive = driveMatch ? `${driveMatch[1].toUpperCase()}:` : "";
    const normalizedDriveMatch = drive ? source.match(/^([a-zA-Z]):(?:\/+|$)/) : null;
    const remainder = drive ? source.slice(normalizedDriveMatch?.[0].length || 0) : source.replace(/^\/+/, "");
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
    return {drive, parts};
};

export const normalizeFilePath = (value: string) => {
    const {drive, parts} = parsePath(value);
    const root = drive ? `${drive}/` : "/";
    return parts.length > 0 ? `${root}${parts.join("/")}` : root;
};

export const comparableFilePath = (value: string) => {
    const normalized = normalizeFilePath(value);
    const withoutTrailingSlash = normalized === "/" || /^[a-z]:\/$/i.test(normalized)
        ? normalized
        : normalized.replace(/\/$/, "");
    return /^[a-z]:(?:\/|$)/i.test(withoutTrailingSlash)
        ? withoutTrailingSlash.toLowerCase()
        : withoutTrailingSlash;
};

export const isFilePathWithin = (path: string, root: string) => {
    const current = comparableFilePath(path);
    const parent = comparableFilePath(root);
    return current === parent || current.startsWith(parent.endsWith("/") ? parent : `${parent}/`);
};

export const joinFilePath = (base: string, name: string) => {
    const parent = normalizeFilePath(base);
    const child = String(name ?? "");
    return `${parent}${parent.endsWith("/") ? "" : "/"}${child}`;
};

export const parentFilePath = (value: string) => {
    const {drive, parts} = parsePath(value);
    parts.pop();
    return normalizeFilePath(`${drive ? `${drive}/` : "/"}${parts.join("/")}`);
};

export const buildFileBreadcrumbs = (value: string): FileBreadcrumbItem[] => {
    const {drive, parts} = parsePath(value);
    const root = drive ? `${drive}/` : "/";
    const result: FileBreadcrumbItem[] = [{label: drive || "/", path: root}];
    parts.forEach((part, index) => {
        result.push({
            label: part,
            path: normalizeFilePath(`${root}${parts.slice(0, index + 1).join("/")}`),
        });
    });
    return result;
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
    return path.startsWith("/") || /^[a-zA-Z]:[\\/]/.test(path);
};

export const getFileCategory = (name: string, isDir = false): FileCategory => {
    if (isDir) {
        return "folder";
    }
    return categoryByExtension.get(getFileExtension(name)) || "file";
};

export const getFileIconType = (name: string, isDir = false) => {
    return categoryIcons[getFileCategory(name, isDir)];
};

export const isTextEditableFile = (name: string) => {
    const category = getFileCategory(name);
    return category === "code" || category === "text" || !getFileExtension(name);
};

export const isPreviewableFile = (name: string) => {
    return previewableExtensions.has(getFileExtension(name));
};

export const isExtractableFile = (name: string) => {
    const value = String(name || "").toLowerCase();
    return value.endsWith(".zip") || value.endsWith(".tar") || value.endsWith(".gz") ||
        value.endsWith(".gzip") || value.endsWith(".tgz") || value.endsWith(".tar.gz");
};
