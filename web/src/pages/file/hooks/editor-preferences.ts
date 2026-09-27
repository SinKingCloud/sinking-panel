import {useCallback, useEffect, useState} from "react";
import {readFileStorage, updateFileStorage} from "./file-storage";

const preferencesVersion = 2;
const availableThemes = new Set<string>([
    "auto",
    "github",
    "github_dark",
    "one_dark",
    "dracula",
    "monokai",
    "tomorrow",
    "tomorrow_night",
    "solarized_light",
    "solarized_dark",
]);

const defaultPreferences: any = {
    theme: "auto",
    fontSize: 12,
    tabSize: 4,
    wrapEnabled: false,
    showLineNumbers: true,
};

const clampInteger = (value: unknown, minimum: number, maximum: number, fallback: number) => {
    const numeric = Math.trunc(Number(value));
    return Number.isFinite(numeric) ? Math.min(maximum, Math.max(minimum, numeric)) : fallback;
};

const loadPreferences = (): any => {
    if (typeof window === "undefined") {
        return defaultPreferences;
    }
    try {
        const stored = readFileStorage().editor?.preferences as any;
        if (!stored || typeof stored !== "object") {
            return defaultPreferences;
        }
        const theme = availableThemes.has(stored.theme as string)
            ? stored.theme as string
            : defaultPreferences.theme;
        const tabSize = [2, 4, 8].includes(Number(stored.tabSize))
            ? Number(stored.tabSize)
            : defaultPreferences.tabSize;
        const usesLegacyDefaults = stored.version !== preferencesVersion
            && Number(stored.fontSize) === 13
            && (stored.theme === undefined || stored.theme === "auto")
            && (stored.tabSize === undefined || Number(stored.tabSize) === 4)
            && (stored.wrapEnabled === undefined || stored.wrapEnabled)
            && (stored.showLineNumbers === undefined || stored.showLineNumbers);
        return {
            theme,
            fontSize: clampInteger(usesLegacyDefaults ? 12 : stored.fontSize, 12, 24, defaultPreferences.fontSize),
            tabSize,
            wrapEnabled: typeof stored.wrapEnabled === "boolean"
                ? stored.wrapEnabled
                : defaultPreferences.wrapEnabled,
            showLineNumbers: typeof stored.showLineNumbers === "boolean"
                ? stored.showLineNumbers
                : defaultPreferences.showLineNumbers,
        };
    } catch {
        return defaultPreferences;
    }
};

const useFileEditorPreferences = () => {
    const [preferences, setPreferences] = useState<any>(loadPreferences);

    useEffect(() => {
        updateFileStorage((current) => ({
            ...current,
            editor: {
                ...current.editor,
                preferences: {
                    ...preferences,
                    version: preferencesVersion,
                },
            },
        }));
    }, [preferences]);

    const resolveTheme = useCallback((dark: boolean) => (
        preferences.theme === "auto"
            ? dark ? "one_dark" : "github"
            : preferences.theme
    ), [preferences.theme]);

    return {preferences, setPreferences, resolveTheme};
};

export default useFileEditorPreferences;
