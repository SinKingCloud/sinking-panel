import {useCallback, useMemo, useRef, useState} from "react";

export interface FileOperationLock {
    paths: ReadonlySet<string>;
    begin: (path: string) => boolean;
    finish: (path: string) => void;
    beginMany: (paths: readonly string[]) => boolean;
    finishMany: (paths: readonly string[]) => void;
}

const useFileOperationLock = (): FileOperationLock => {
    const pathsRef = useRef(new Set<string>());
    const [paths, setPaths] = useState<ReadonlySet<string>>(() => new Set());

    const publish = useCallback(() => {
        setPaths(new Set(pathsRef.current));
    }, []);

    const begin = useCallback((path: string) => {
        if (pathsRef.current.has(path)) {
            return false;
        }
        pathsRef.current.add(path);
        publish();
        return true;
    }, [publish]);

    const finish = useCallback((path: string) => {
        if (pathsRef.current.delete(path)) {
            publish();
        }
    }, [publish]);

    const beginMany = useCallback((values: readonly string[]) => {
        const unique = Array.from(new Set(values));
        if (unique.some((path) => pathsRef.current.has(path))) {
            return false;
        }
        unique.forEach((path) => pathsRef.current.add(path));
        if (unique.length > 0) {
            publish();
        }
        return true;
    }, [publish]);

    const finishMany = useCallback((values: readonly string[]) => {
        let changed = false;
        values.forEach((path) => {
            changed = pathsRef.current.delete(path) || changed;
        });
        if (changed) {
            publish();
        }
    }, [publish]);

    return useMemo(() => ({
        paths,
        begin,
        finish,
        beginMany,
        finishMany,
    }), [begin, beginMany, finish, finishMany, paths]);
};

export default useFileOperationLock;
