import {useCallback, useEffect, useMemo, useState} from "react";
import type {FileRecord} from "@/service/api/file";
import {joinFilePath} from "../utils";

export interface UseFileSelectionOptions {
    path: string;
    items: readonly FileRecord[];
    operatingPaths: ReadonlySet<string>;
}

const useFileSelection = ({path, items, operatingPaths}: UseFileSelectionOptions) => {
    const [selectedPaths, setSelectedPaths] = useState<ReadonlySet<string>>(() => new Set());

    const selectedRecords = useMemo(() => items.filter((record) => (
        selectedPaths.has(joinFilePath(path, record.name))
    )), [items, path, selectedPaths]);

    const hasBusySelection = useMemo(() => selectedRecords.some((record) => (
        operatingPaths.has(joinFilePath(path, record.name))
    )), [operatingPaths, path, selectedRecords]);

    useEffect(() => {
        const visiblePaths = new Set(items.map((record) => joinFilePath(path, record.name)));
        setSelectedPaths((current) => {
            const next = new Set(Array.from(current).filter((value) => visiblePaths.has(value)));
            return next.size === current.size ? current : next;
        });
    }, [items, path]);

    const change = useCallback((values: string[]) => {
        setSelectedPaths(new Set(values));
    }, []);

    const clear = useCallback(() => {
        setSelectedPaths(new Set());
    }, []);

    const remove = useCallback((values: Iterable<string>) => {
        const removed = new Set(values);
        if (removed.size === 0) {
            return;
        }
        setSelectedPaths((current) => {
            const next = new Set(Array.from(current).filter((value) => !removed.has(value)));
            return next.size === current.size ? current : next;
        });
    }, []);

    return {
        selectedPaths,
        selectedRecords,
        hasBusySelection,
        change,
        clear,
        remove,
    };
};

export default useFileSelection;
