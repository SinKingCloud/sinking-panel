const history: Record<string, any[]> = Object.create(null);
const pending = new Set<string>();

// 请求历史与 SQL 历史共用 database 缓存；写入失败时保留当前会话记录。
export default (id: string, execution?: any) => {
    if (!id) return [];
    let storage: any;
    try {
        const saved = JSON.parse(localStorage.getItem("database") || "{}");
        if (saved && typeof saved === "object" && !Array.isArray(saved)) storage = saved;
    } catch { /* 存储不可读时使用当前会话的历史。 */ }
    let records = storage && !pending.has(id)
        ? (Array.isArray(storage.requestHistory?.[id]) ? storage.requestHistory[id] : [])
        : history[id] || [];
    const valid = (item: any) => item && Number.isFinite(item.time)
        && typeof item.method === "string" && typeof item.path === "string" && item.path.trim()
        && typeof item.body === "string" && item.method.length + item.path.length + item.body.length <= 65536;
    const added = execution && valid(execution);
    if (added) {
        const record = {...execution};
        while (records.some((item: any) => item?.time === record.time)) record.time++;
        records = [record, ...records];
    }
    // 不存响应，也不截断请求内容；过大的请求不写入历史。
    records = records.filter(valid).sort((left: any, right: any) => right.time - left.time).slice(0, 100)
        .map((item: any) => ({
            time: item.time, method: item.method.toUpperCase(), path: item.path, body: item.body,
            success: Boolean(item.success), elapsed_ms: Number.isFinite(item.elapsed_ms) ? item.elapsed_ms : undefined,
            status: Number.isFinite(item.status) ? item.status : undefined,
        }));
    history[id] = records;
    if (added) pending.add(id);
    if (storage && pending.size) {
        const requestHistory = Object.assign(Object.create(null), storage.requestHistory && typeof storage.requestHistory === "object"
            && !Array.isArray(storage.requestHistory) ? storage.requestHistory : {});
        for (const instanceId of pending) requestHistory[instanceId] = history[instanceId];
        try {
            localStorage.setItem("database", JSON.stringify({...storage, requestHistory}));
            pending.clear();
        } catch { /* 配额不足或禁止写入时继续使用当前会话的记录。 */ }
    }
    return records;
};
