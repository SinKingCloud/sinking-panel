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
};

export type MonitorHistory = {
    points: MonitorSample[];
    interval: number;
    start: number;
    end: number;
};

export type MonitorSeries = {
    key: keyof MonitorSample;
    label: string;
    color: string;
};
