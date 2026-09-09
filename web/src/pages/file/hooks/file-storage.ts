export interface FileStorageTask {
    id: string;
    title: string;
}

export interface FileStorageData {
    editor?: {
        preferences?: unknown;
        treeCollapsed?: boolean;
        treeWidth?: number;
    };
    lastPath?: string;
    tasks?: FileStorageTask[];
}

const storageKey = "file";

const parseObject = (value: string | null): Record<string, unknown> | undefined => {
    if (!value) {
        return undefined;
    }
    try {
        const parsed = JSON.parse(value);
        return parsed && typeof parsed === "object" && !Array.isArray(parsed)
            ? parsed as Record<string, unknown>
            : undefined;
    } catch {
        return undefined;
    }
};

const readRaw = (): FileStorageData => {
    if (typeof window === "undefined") {
        return {};
    }
    const stored = parseObject(window.localStorage.getItem(storageKey));
    return stored ? stored as FileStorageData : {};
};

const writeRaw = (value: FileStorageData) => {
    if (typeof window === "undefined") {
        return;
    }
    try {
        window.localStorage.setItem(storageKey, JSON.stringify(value));
    } catch {
        // Storage can be unavailable in restricted browser contexts.
    }
};

export const readFileStorage = (): FileStorageData => {
    if (typeof window === "undefined") {
        return {};
    }
    return readRaw();
};

export const updateFileStorage = (updater: (current: FileStorageData) => FileStorageData) => {
    writeRaw(updater(readFileStorage()));
};
