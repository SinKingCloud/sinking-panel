import {useEffect, useMemo, useState} from "react";
import {getEnum as fetchEnum} from "@/service/api/system";

export type EnumValue = Record<string, string>;
export type EnumGroup = Record<string, EnumValue>;
export type EnumsData = Record<string, EnumGroup>;

const enumCache: Record<string, EnumGroup> = {};

function normalizeNames(name: string | string[], restNames: string[] = []): string[] {
    const names = Array.isArray(name) ? name : [name, ...restNames];
    return Array.from(new Set(names.map((item) => item?.trim()).filter(Boolean)) as any);
}

function getCachedEnums(names: string[]): EnumsData {
    return names.reduce((data: EnumsData, item) => {
        if (enumCache[item]) {
            data[item] = enumCache[item];
        }
        return data;
    }, {});
}

async function getEnumData(names: string[]): Promise<EnumsData> {
    if (names.length <= 0) {
        return {};
    }

    const cachedData = getCachedEnums(names);
    const needFetchNames = names.filter((item) => !cachedData[item]);
    if (needFetchNames.length <= 0) {
        return cachedData;
    }

    const responses = await Promise.all(needFetchNames.map(async (item) => {
        const res = await fetchEnum({
            body: {
                name: item,
            },
        });
        if (res?.code !== 200) {
            throw new Error(res?.message || "获取枚举数据失败");
        }
        return {
            name: item,
            data: res?.data || {},
        };
    }));

    responses.forEach((item) => {
        enumCache[item.name] = item.data;
    });

    return {
        ...cachedData,
        ...getCachedEnums(needFetchNames),
    };
}

export function clearEnumCache(name?: string) {
    if (name) {
        delete enumCache[name];
        return;
    }
    Object.keys(enumCache).forEach((key) => {
        delete enumCache[key];
    });
}

export async function getEnum(name: string): Promise<EnumGroup>;
export async function getEnum(name: string[]): Promise<EnumsData>;
export async function getEnum(...names: string[]): Promise<EnumsData>;
export async function getEnum(name: string | string[], ...restNames: string[]): Promise<EnumGroup | EnumsData> {
    const names = normalizeNames(name, restNames);
    if (names.length <= 0) {
        return {};
    }

    const data = await getEnumData(names);
    if (Array.isArray(name) || restNames.length > 0) {
        return data;
    }
    return data?.[names[0]] || {};
}

/**
 * 获取枚举 Hook
 * @param name 枚举标识、枚举标识数组或多个枚举标识
 */
export function useEnum(name: string): [EnumGroup, boolean, Error | null];
export function useEnum(name: string[]): [EnumsData, boolean, Error | null];
export function useEnum(...names: string[]): [EnumsData, boolean, Error | null];
export function useEnum(name: string | string[], ...restNames: string[]): [EnumGroup | EnumsData, boolean, Error | null] {
    const isMultiple = Array.isArray(name) || restNames.length > 0;
    const namesKey = useMemo(() => {
        return normalizeNames(name, restNames).join(",");
    }, [Array.isArray(name) ? name.join(",") : name, restNames.join(",")]);
    const [enumData, setEnumData] = useState<EnumGroup | EnumsData>({});
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<Error | null>(null);

    useEffect(() => {
        const names = normalizeNames(name, restNames);
        if (names.length <= 0) {
            setEnumData({});
            setLoading(false);
            setError(null);
            return;
        }

        let mounted = true;
        setLoading(true);
        setError(null);

        getEnum(isMultiple ? names : names[0] as any).then((data) => {
            if (mounted) {
                setEnumData(data);
            }
        }).catch((err) => {
            if (mounted) {
                setEnumData({});
                setError(err instanceof Error ? err : new Error("获取枚举数据失败"));
            }
        }).finally(() => {
            if (mounted) {
                setLoading(false);
            }
        });

        return () => {
            mounted = false;
        };
    }, [isMultiple, namesKey]);

    return [enumData, loading, error];
}

export default useEnum;
