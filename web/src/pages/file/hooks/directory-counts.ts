import {useCallback, useEffect, useRef, useState} from "react";
import {countFile} from "@/service/api/file";
import type {FileCountData, FileRecord} from "@/service/api/file";
import {joinFilePath} from "../utils";

const maxConcurrentCounts = 2;

export type DirectoryCountState =
    | {status: "loading"}
    | {status: "success"; data: FileCountData}
    | {status: "error"; message: string};

export type DirectoryCountMap = ReadonlyMap<string, DirectoryCountState>;

export interface UseDirectoryCountsOptions {
    path: string;
    items: readonly FileRecord[];
    enabled?: boolean;
}

interface CountJob {
    generation: number;
    key: string;
    path: string;
}

export interface UseDirectoryCountsResult {
    counts: DirectoryCountMap;
    countDirectory: (path: string) => void;
}

export const useDirectoryCounts = ({
    path,
    items,
    enabled = true,
}: UseDirectoryCountsOptions): UseDirectoryCountsResult => {
    const mountedRef = useRef(true);
    const generationRef = useRef(0);
    const activeRef = useRef(0);
    const pendingRef = useRef(new Set<string>());
    const visiblePathsRef = useRef<ReadonlySet<string>>(new Set());
    const queueRef = useRef<CountJob[]>([]);
    const pumpRef = useRef<() => void>(() => undefined);
    const [counts, setCounts] = useState<DirectoryCountMap>(() => new Map());

    const publish = (job: CountJob, state: DirectoryCountState) => {
        if (!mountedRef.current || generationRef.current !== job.generation) {
            return;
        }
        setCounts((current) => {
            if (!current.has(job.path)) {
                return current;
            }
            const next = new Map(current);
            next.set(job.path, state);
            return next;
        });
    };

    pumpRef.current = () => {
        if (!mountedRef.current) {
            return;
        }
        while (activeRef.current < maxConcurrentCounts && queueRef.current.length > 0) {
            const job = queueRef.current.shift();
            if (!job) {
                return;
            }
            activeRef.current += 1;
            void countFile({body: {path: job.path}}).then((response) => {
                const data = response?.data;
                if (
                    response?.code === 200
                    && data
                    && Number.isFinite(Number(data.size))
                    && Number.isFinite(Number(data.file))
                    && Number.isFinite(Number(data.dir))
                ) {
                    publish(job, {
                        status: "success",
                        data: {
                            size: Number(data.size),
                            file: Number(data.file),
                            dir: Number(data.dir),
                        },
                    });
                    return;
                }
                publish(job, {status: "error", message: response?.message || "统计失败"});
            }).catch(() => {
                publish(job, {status: "error", message: "统计失败"});
            }).finally(() => {
                pendingRef.current.delete(job.key);
                activeRef.current = Math.max(0, activeRef.current - 1);
                pumpRef.current();
            });
        }
    };

    useEffect(() => {
        mountedRef.current = true;
        return () => {
            mountedRef.current = false;
            generationRef.current += 1;
            queueRef.current = [];
            pendingRef.current.clear();
            visiblePathsRef.current = new Set();
        };
    }, []);

    useEffect(() => {
        const generation = ++generationRef.current;
        queueRef.current.forEach((job) => pendingRef.current.delete(job.key));
        queueRef.current = [];
        const directoryPaths = enabled ? Array.from(new Set(
            items
                .filter((item) => item.is_dir)
                .map((item) => joinFilePath(path, item.name)),
        )) : [];
        visiblePathsRef.current = new Set(directoryPaths);
        setCounts(new Map());

        return () => {
            const queued = queueRef.current.filter((job) => job.generation === generation);
            queued.forEach((job) => pendingRef.current.delete(job.key));
            queueRef.current = queueRef.current.filter((job) => job.generation !== generation);
        };
    }, [enabled, items, path]);

    const countDirectory = useCallback((directoryPath: string) => {
        if (!mountedRef.current || !visiblePathsRef.current.has(directoryPath)) {
            return;
        }
        const generation = generationRef.current;
        const key = `${generation}\u0000${directoryPath}`;
        if (pendingRef.current.has(key)) {
            return;
        }
        pendingRef.current.add(key);
        setCounts((current) => {
            const next = new Map(current);
            next.set(directoryPath, {status: "loading"});
            return next;
        });
        queueRef.current.push({generation, key, path: directoryPath});
        pumpRef.current();
    }, []);

    return {counts, countDirectory};
};

export default useDirectoryCounts;
