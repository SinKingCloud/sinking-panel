const history: Record<string, any[]> = Object.create(null);
const pending = new Set<string>();

// 按实例保留最近 100 条命令，与其他数据库管理器共用 database 缓存。
export default (id: string, execution?: any) => {
    if (!id) return [];
    let storage: any;
    try {
        const saved = JSON.parse(localStorage.getItem("database") || "{}");
        if (saved && typeof saved === "object" && !Array.isArray(saved)) storage = saved;
    } catch { /* 禁止访问本地存储时保留当前会话的记录。 */ }
    let records = storage && !pending.has(id)
        ? (Array.isArray(storage.commandHistory?.[id]) ? storage.commandHistory[id] : [])
        : history[id] || [];
    const valid = (item: any) => item && !item.sensitive && Number.isFinite(item.time)
        && typeof item.command === "string" && item.command.trim() && item.command.length <= 65536
        && typeof item.database === "string";
    const added = execution && valid(execution);
    if (added) {
        const record = {...execution};
        while (records.some((item: any) => item?.time === record.time)) record.time++;
        records = [record, ...records];
    }
    // 不保存响应和过大的命令；认证、改密命令由调用处标记并排除。
    records = records.filter(valid).sort((left: any, right: any) => right.time - left.time).slice(0, 100)
        .map((item: any) => ({time: item.time, command: item.command, database: item.database,
            success: Boolean(item.success), elapsed_ms: Number.isFinite(item.elapsed_ms) ? item.elapsed_ms : undefined}));
    history[id] = records;
    if (added) pending.add(id);
    if (storage && pending.size) {
        const commandHistory = Object.assign(Object.create(null), storage.commandHistory && typeof storage.commandHistory === "object"
            && !Array.isArray(storage.commandHistory) ? storage.commandHistory : {});
        for (const instanceId of pending) commandHistory[instanceId] = history[instanceId];
        try {
            localStorage.setItem("database", JSON.stringify({...storage, commandHistory}));
            pending.clear();
        } catch { /* 存储已满时仍可查看本次会话的历史。 */ }
    }
    return records;
};
