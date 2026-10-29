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

export type MonitorSeries = {
    key: Exclude<keyof MonitorSample, "process">;
    label: string;
    color: string;
};

export type MonitorProcessMetric = "cpu" | "memory" | "disk_io";

export type MonitorProcess = {
    pid: number;
    name: string;
    value: number;
};

export type MonitorProcessGroup = {
    time: number;
    items: MonitorProcess[];
    partial?: boolean;
};
