const storageKey = "database";
let history: Record<string, any[]> = Object.create(null);
let unsaved = false;
let lastSaved: string | undefined;

// 所有实例共用一个存储键；存储不可用时继续保留当前会话的历史。
export default (id: string, execution?: any | null) => {
    let storage: any;
    let records: any[] = [];
    let value: string | null = null;
    try {
        value = localStorage.getItem(storageKey);
        // 未变更的读取复用内存结果，避免打开历史时重复解析、排序并写回全部实例。
        if (execution === undefined && !unsaved && value === lastSaved) return history[id] || [];
        try {
            const saved = JSON.parse(value || "{}");
            storage = saved && typeof saved === "object" && !Array.isArray(saved) ? saved : {};
            if (!unsaved) {
                const savedHistory = storage.queryHistory;
                records = savedHistory && typeof savedHistory === "object" && !Array.isArray(savedHistory)
                    ? Object.entries(savedHistory).flatMap(([id, items]) => Array.isArray(items) ? items.map((item) => ({...item, id})) : [])
                    : [];
            }
        } catch { storage = {}; if (!unsaved) records = []; }
    } catch { /* 存储不可读时继续使用当前会话的历史。 */ }
    if (unsaved || !storage) records = Object.entries(history).flatMap(([id, items]) => items.map((item) => ({...item, id})));
    if (execution === null) records = records.filter((item) => String(item?.id) !== id);
    else if (execution) records = [{...execution, id}, ...records];

    const seen = new Set<string>();
    history = Object.create(null);
    // 按时间倒序整理，每个实例独立保留最新记录。
    for (const item of records.filter((item) => item && ["string", "number"].includes(typeof item.id)
        && typeof item.statement === "string" && item.statement.length <= 65536 && Number.isFinite(item.time))
        .sort((left, right) => right.time - left.time)) {
        const instanceId = String(item.id), key = `${instanceId}:${item.time}`;
        const record = {time: item.time, operation: item.operation === "query" ? "query" : "execute",
            statement: item.statement, success: Boolean(item.success), elapsed_ms: Number.isFinite(item.elapsed_ms) ? item.elapsed_ms : undefined};
        // 单实例最多 100 条；不保存参数或截断的 SQL。
        if (seen.has(key) || (history[instanceId]?.length || 0) >= 100) continue;
        (history[instanceId] ||= []).push(record);
        seen.add(key);
    }
    if (!storage) {
        unsaved = true;
        return history[id] || [];
    }
    try {
        // 历史只是数据库缓存的一部分，清空时也保留其他配置。
        const serialized = JSON.stringify({...storage, queryHistory: history});
        if (serialized !== value) localStorage.setItem(storageKey, serialized);
        lastSaved = serialized;
        unsaved = false;
    } catch { unsaved = true; }
    return history[id] || [];
};
