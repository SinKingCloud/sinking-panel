const groups: Record<string, string> = {
    health: "集群健康", version: "服务与版本", indices: "集群索引统计", nodes: "节点资源统计",
    cluster: "其他集群统计", index_stats: "当前索引统计", settings: "索引配置", mappings: "字段映射", aliases: "索引别名", index: "索引信息",
};

const labels: Record<string, string> = {
    cluster_name: "集群名称", cluster_uuid: "集群 UUID", status: "状态", timed_out: "检查超时",
    number_of_nodes: "节点数量", number_of_data_nodes: "数据节点", active_primary_shards: "活动主分片",
    active_shards: "活动分片", relocating_shards: "迁移中分片", initializing_shards: "初始化分片",
    unassigned_shards: "未分配分片", delayed_unassigned_shards: "延迟分配分片", unassigned_primary_shards: "未分配主分片",
    number_of_pending_tasks: "等待任务", number_of_in_flight_fetch: "进行中获取任务",
    task_max_waiting_in_queue_millis: "最长任务等待", active_shards_percent_as_number: "分片活跃率",
    name: "名称", number: "版本号", lucene_version: "Lucene 版本", build_flavor: "构建类型", build_type: "构建方式",
    build_hash: "构建提交", build_date: "构建时间", build_snapshot: "快照版本", tagline: "产品说明",
    minimum_wire_compatibility_version: "最低通信兼容版本", minimum_index_compatibility_version: "最低索引兼容版本",
    timestamp: "采集时间", _nodes: "节点请求", total: "合计", successful: "成功数量", failed: "失败数量",
    count: "数量", deleted: "已删除数量", primaries: "主分片", replication: "副本比例", min: "最小值", max: "最大值", avg: "平均值",
    versions: "节点版本", master: "主节点", data: "数据节点", ingest: "摄取节点", coordinating_only: "协调节点",
    data_hot: "热数据节点", data_warm: "温数据节点", data_cold: "冷数据节点", data_frozen: "冻结数据节点", data_content: "内容数据节点",
    available_processors: "可用处理器", allocated_processors: "分配处理器", percent: "使用率", threads: "线程数量",
    heap_used_in_bytes: "已用堆内存", heap_max_in_bytes: "最大堆内存", non_heap_used_in_bytes: "已用非堆内存",
    non_heap_committed_in_bytes: "已提交非堆内存", heap_committed_in_bytes: "已提交堆内存", max_uptime_in_millis: "最长运行时间",
    total_in_bytes: "总量", free_in_bytes: "空闲量", available_in_bytes: "可用量", used_in_bytes: "已用量",
    adjusted_total_in_bytes: "调整后总量", total_virtual_in_bytes: "虚拟内存总量", used_percent: "已用比例", free_percent: "空闲比例",
    open_file_descriptors: "打开文件数", max_file_descriptors: "文件数上限", size_in_bytes: "存储大小",
    total_data_set_size_in_bytes: "数据集总大小", reserved_in_bytes: "预留空间", memory_in_bytes: "内存占用",
    documents: "文档数量", size_bytes: "存储大小", index_total: "写入次数", index_time_in_millis: "写入耗时", index_current: "正在写入",
    index_failed: "写入失败", delete_total: "删除次数", delete_time_in_millis: "删除耗时", delete_current: "正在删除",
    query_total: "查询次数", query_time_in_millis: "查询耗时", query_current: "正在查询",
    fetch_total: "获取次数", fetch_time_in_millis: "获取耗时", fetch_current: "正在获取", open_contexts: "打开的查询上下文",
    scroll_total: "滚动查询次数", scroll_time_in_millis: "滚动查询耗时", scroll_current: "正在滚动查询",
    total_time_in_millis: "总耗时", total_size_in_bytes: "总大小", current: "当前数量", current_docs: "当前文档数",
    current_size_in_bytes: "当前大小", total_docs: "文档总数", total_auto_throttle_in_bytes: "自动限速总量",
    total_throttled_time_in_millis: "限流耗时", total_stopped_time_in_millis: "暂停耗时", is_throttled: "正在限流",
    memory_size_in_bytes: "内存大小", cache_size: "缓存条目", cache_count: "累计缓存条目", hit_count: "缓存命中", miss_count: "缓存未命中",
    evictions: "缓存淘汰", operations: "操作次数", uncommitted_operations: "未提交操作", uncommitted_size_in_bytes: "未提交大小",
    number_of_shards: "主分片数量", number_of_replicas: "副本数量", auto_expand_replicas: "自动扩展副本",
    refresh_interval: "刷新间隔", max_result_window: "最大结果窗口", max_inner_result_window: "内部结果窗口",
    max_rescore_window: "重评分窗口", uuid: "UUID", creation_date: "创建时间", provided_name: "索引原名称",
    codec: "压缩编码", default_pipeline: "默认处理管道", final_pipeline: "最终处理管道", priority: "优先级",
    read_only: "只读", read_only_allow_delete: "只读且允许删除", write: "写入", read: "读取",
    durability: "持久化方式", sync_interval: "同步间隔", flush_threshold_size: "刷新阈值",
    enabled: "启用", dynamic: "动态映射", type: "类型", analyzer: "分析器", search_analyzer: "搜索分析器",
    format: "格式", ignore_above: "忽略长度阈值", doc_values: "列式存储", store: "独立存储", null_value: "空值替代",
    index: "建立索引", norms: "评分规范", index_options: "索引选项", copy_to: "复制到", eager_global_ordinals: "预加载全局序号",
    filter: "过滤条件", routing: "路由", index_routing: "写入路由", search_routing: "搜索路由", is_write_index: "写入索引",
    stats_error: "统计读取错误",
};

// 保留接口返回的全部叶子和空容器；路径段保留原始类型，避免字段名与数组下标混淆。
export default (data: any, info: any) => {
    const rows: Record<string, any[]> = Object.fromEntries(Object.keys(groups).map((group) => [group, []]));
    for (const [scope, source] of [["data", data], ["info", info]]) {
        if (source === undefined) continue;
        const stack: {value: any; path: (string | number)[]}[] = [{value: source, path: []}];
        while (stack.length) {
            const item = stack.pop()!;
            const entries = item.value !== null && typeof item.value === "object"
                ? Array.isArray(item.value) ? Array.from(item.value.entries()) : Object.entries(item.value) : [];
            if (entries.length) {
                for (let index = entries.length - 1; index >= 0; index--) {
                    stack.push({value: entries[index][1], path: [...item.path, entries[index][0]]});
                }
                continue;
            }
            const [first, second] = item.path;
            const group = scope === "data"
                ? first === "health" ? "health" : first === "version" ? "version"
                    : first === "stats" && second === "indices" ? "indices" : first === "stats" && second === "nodes" ? "nodes" : "cluster"
                : first === "stats" ? "index_stats" : first === "settings" ? "settings" : first === "mappings" ? "mappings" : first === "aliases" ? "aliases" : "index";
            const name = String(item.path[item.path.length - 1] ?? "");
            const path = item.path.map((part, index) => typeof part === "number" ? `[${part}]`
                : /^[A-Za-z_][A-Za-z0-9_]*$/.test(part) ? `${index ? "." : ""}${part}` : `[${JSON.stringify(part)}]`).join("") || "$";
            rows[group].push({key: JSON.stringify([scope, ...item.path]), group, title: groups[group], path, name,
                label: typeof labels[name] === "string" ? labels[name] : item.path.length ? name : "原始数据", value: item.value});
        }
    }
    return Object.values(rows).flat();
};
