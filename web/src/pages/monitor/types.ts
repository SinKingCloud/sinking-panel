export type MonitorSettings = {
    enabled: boolean;
    retention_days: number;
    interval: number;
    storage_bytes: number;
    last_error?: string;
};

export type MonitorSample = {
    time: number;
    samples: number;
    cpu: number;
    cpu_max: number;
    memory: number;
    memory_used: number;
    memory_total: number;
    disk_usage: number;
    disk_used: number;
    disk_total: number;
    disk_read: number;
    disk_write: number;
    disk_read_iops: number;
    disk_write_iops: number;
    net_sent: number;
    net_recv: number;
    load1: number;
    load5: number;
    load15: number;
    process?: Partial<Record<MonitorProcessMetric, MonitorProcessGroup>>;
};

export type MonitorHistory = {
    data: MonitorSample[];
    interval: number;
    start: number;
    end: number;
    process_error?: string;
};

export type MonitorUnit = "percent" | "bytes" | "load" | "iops";

export type MonitorSeries = {
    key: Exclude<keyof MonitorSample, "process">;
    label: string;
    color: string;
    unit?: MonitorUnit;
};

export type MonitorProcessMetric = "cpu" | "memory" | "disk_io";

export type MonitorProcess = {
    pid: number;
    name: string;
    value: number;
};

export type MonitorProcessGroup = {
    io_scope?: "process";
    time: number;
    items: MonitorProcess[];
    partial?: boolean;
    // 采样窗口的整机高点：CPU 和 I/O 为窗口均值，内存为采样时刻的使用率。
    peak?: number;
};
