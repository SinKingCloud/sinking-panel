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

const readRaw = (): any => {
    if (typeof window === "undefined") {
        return {};
    }
    const stored = parseObject(window.localStorage.getItem(storageKey));
    return stored ? stored as any : {};
};

const writeRaw = (value: any) => {
    if (typeof window === "undefined") {
        return;
    }
    try {
        window.localStorage.setItem(storageKey, JSON.stringify(value));
    } catch {
        // Storage can be unavailable in restricted browser contexts.
    }
};

export const readFileStorage = (): any => {
    if (typeof window === "undefined") {
        return {};
    }
    return readRaw();
};

export const updateFileStorage = (updater: (current: any) => any) => {
    writeRaw(updater(readFileStorage()));
};
