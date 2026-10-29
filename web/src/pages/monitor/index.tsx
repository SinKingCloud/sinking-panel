import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {App, DatePicker, Grid, Segmented, theme as antdTheme} from "antd";
import dayjs from "dayjs";
import type {Dayjs} from "dayjs";
import {Body, Icon, PageTable} from "sinking-antd";
import {systemMonitor} from "@/service/api/system";
import {formatSize} from "@/utils/string";
import MonitorChart from "./components/chart";
import MonitorSettingsModal from "./components/settings";
import type {MonitorSettingsRef} from "./components/settings";
import type {MonitorHistory, MonitorSample, MonitorSeries, MonitorSettings} from "./types";
import useStyles from "./styles";

const periods = [
    {value: "1", label: "1 小时"},
    {value: "6", label: "6 小时"},
    {value: "24", label: "24 小时"},
    {value: "168", label: "7 天"},
];

export default (): React.ReactNode => {
    const {styles, cx} = useStyles();
    const {token} = antdTheme.useToken();
    const screens = Grid.useBreakpoint();
    const isMobile = screens.md === false;
    const pointLimit = screens.md === undefined ? undefined : isMobile ? 240 : 720;
    const {message} = App.useApp();
    const settingsRef = useRef<MonitorSettingsRef>(null);
    const historyRequestRef = useRef(0);
    const historyLoadingRef = useRef(false);
    const [rangeKey, setRangeKey] = useState("");
    const [settings, setSettings] = useState<MonitorSettings>();
    const [historyError, setHistoryError] = useState("");
    const [loading, setLoading] = useState(true);
    const [initialLoading, setInitialLoading] = useState(true);
    const [period, setPeriod] = useState("1");
    const [customRange, setCustomRange] = useState<[Dayjs, Dayjs] | null>(null);
    const [datePickerOpen, setDatePickerOpen] = useState(false);
    const [data, setData] = useState<MonitorHistory>();

    const loadHistory = useCallback(async (background = false) => {
        if (pointLimit === undefined) return;
        if (background && historyLoadingRef.current) return;
        const request = ++historyRequestRef.current;
        historyLoadingRef.current = true;
        const end = customRange ? Math.min(customRange[1].endOf("day").valueOf(), Date.now()) : Date.now();
        const start = customRange ? customRange[0].startOf("day").valueOf() : end - Number(period) * 3600000;
        const span = end - start;
        // 一天内的趋势不需要铺满像素，24 小时在桌面按 5 分钟、移动端按 15 分钟聚合。
        const maxPoints = span > 6 * 3600000 && span <= 24 * 3600000 ? (isMobile ? 120 : 300) : pointLimit;
        if (!background) setLoading(true);
        try {
            const response = await systemMonitor({body: {action: "history", start, end, max_points: maxPoints}});
            if (request !== historyRequestRef.current) return;
            if (response?.code !== 200 || !response.data) throw new Error(response?.message || "读取监控历史失败");
            setRangeKey(customRange ? customRange.map((date) => date.format("YYYY-MM-DD")).join("/") : period);
            setData({...response.data, points: Array.isArray(response.data.points) ? response.data.points : []});
            setHistoryError("");
        } catch (error) {
            if (request === historyRequestRef.current) {
                setHistoryError(error instanceof Error ? error.message : "读取监控历史失败");
            }
        } finally {
            if (request === historyRequestRef.current) {
                historyLoadingRef.current = false;
                setLoading(false);
                setInitialLoading(false);
            }
        }
    }, [period, customRange, pointLimit, isMobile]);
    useEffect(() => {
        void loadHistory();
        return () => {
            historyRequestRef.current += 1;
            historyLoadingRef.current = false;
        };
    }, [loadHistory]);

    useEffect(() => {
        if ((customRange && !customRange[1].isSame(dayjs(), "day")) || settings?.enabled === false) return;
        const refresh = () => {
            if (document.visibilityState === "visible" && (!customRange || customRange[1].isSame(dayjs(), "day"))) {
                void loadHistory(true);
            }
        };
        const timer = window.setInterval(refresh, 60000);
        document.addEventListener("visibilitychange", refresh);
        return () => {
            window.clearInterval(timer);
            document.removeEventListener("visibilitychange", refresh);
        };
    }, [loadHistory, customRange, settings?.enabled]);

    const latest = data?.points[data.points.length - 1];
    const charts = useMemo<{title: string; capacity?: [keyof MonitorSample, keyof MonitorSample]; unit: "percent" | "bytes" | "load"; series: MonitorSeries[]}[]>(() => [
        {title: "系统负载", unit: "load", series: [
            {key: "load1", label: "1 分钟", color: token.colorPrimary},
            {key: "load5", label: "5 分钟", color: "#50b6b2"},
            {key: "load15", label: "15 分钟", color: "#d8a55b"},
        ]},
        {title: "CPU 使用率", unit: "percent", series: [
            {key: "cpu", label: "平均使用率", color: token.colorPrimary},
            {key: "cpu_max", label: "峰值使用率", color: "#50b6b2"},
        ]},
        {title: "内存使用率", capacity: ["memory_used", "memory_total"], unit: "percent", series: [
            {key: "memory", label: "内存使用率", color: token.colorPrimary},
        ]},
        {title: "磁盘使用率", capacity: ["disk_used", "disk_total"], unit: "percent", series: [
            {key: "disk_usage", label: "磁盘使用率", color: token.colorPrimary},
        ]},
        {title: "磁盘 I/O", unit: "bytes", series: [
            {key: "disk_read", label: "读取", color: token.colorPrimary},
            {key: "disk_write", label: "写入", color: "#50b6b2"},
        ]},
        {title: "网络流量", unit: "bytes", series: [
            {key: "net_recv", label: "接收", color: token.colorPrimary},
            {key: "net_sent", label: "发送", color: "#50b6b2"},
        ]},
    ], [token.colorPrimary]);

    const problem = [historyError, settings?.last_error].filter(Boolean).join("；");
    const selectedPeriod = datePickerOpen || customRange ? "custom" : period;
    const periodOptions = [...periods, {
        value: "custom", label: "自定义", className: "monitor-custom-period",
        title: customRange ? `${customRange[0].format("YYYY-MM-DD")} 至 ${customRange[1].format("YYYY-MM-DD")}` : "自定义日期",
    }];

    const changePeriod = (value: string) => {
        if (value === "custom") {
            setDatePickerOpen(true);
            return;
        }
        setDatePickerOpen(false);
        setPeriod(value);
        setCustomRange(null);
    };

    const changeDateRange = (value: [Dayjs, Dayjs] | null) => {
        if (value) {
            const now = dayjs();
            if (value.some((date) => date.isAfter(now, "day"))) {
                message.warning("监控日期不能晚于今天");
                return;
            }
            const start = value[0].startOf("day");
            const end = value[1].endOf("day");
            if (end.isBefore(start) || end.valueOf() - start.valueOf() > 365 * 86400000) {
                message.warning("请选择起始日期不晚于结束日期、跨度不超过 365 天的范围");
                return;
            }
        }
        setCustomRange(value);
        setDatePickerOpen(false);
    };

    const datePicker = <DatePicker.RangePicker className={styles.dateAnchor}
        classNames={{popup: {root: styles.datePopup}}}
        value={customRange} open={datePickerOpen} onOpenChange={setDatePickerOpen}
        tabIndex={-1} allowClear={false} inputReadOnly placement="bottomRight"
        popupAlign={{overflow: {adjustX: true, adjustY: true, shiftX: true}}}
        format="YYYY-MM-DD" aria-label="自定义监控日期" placeholder={["开始日期", "结束日期"]}
        disabledDate={(value) => value.isAfter(dayjs(), "day")}
        onChange={(value) => {
            if (value?.[0] && value[1]) changeDateRange([value[0], value[1]]);
        }}/>;

    return (
        <Body loading={initialLoading}>
            <PageTable
                rootClassName={cx(styles.page, isMobile && styles.mobile)}
                ariaLabel="监控管理"
                search={false}
                pagination={false}
                hero={{
                    eyebrow: "SYSTEM MONITOR",
                    title: "监控管理",
                }}
                filters={isMobile ? [{
                    type: "select",
                    name: "period",
                    label: "监控时间范围",
                    icon: "ClockCircleOutlined",
                    allLabel: false,
                    value: selectedPeriod,
                    options: periodOptions,
                    onChange: changePeriod,
                }] : undefined}
                toolbar={{
                    left: isMobile ? undefined : <div className={styles.timeFilter}>
                        <Segmented<string>
                            aria-label="监控时间范围"
                            value={selectedPeriod}
                            options={periodOptions}
                            onChange={changePeriod}
                            onClick={(event) => {
                                if ((event.target as HTMLElement).closest(".monitor-custom-period")) setDatePickerOpen(true);
                            }}
                            onKeyDownCapture={(event) => {
                                if ((event.key === "Enter" || event.key === " ") && (event.target as HTMLElement).closest(".monitor-custom-period")) {
                                    event.preventDefault();
                                    setDatePickerOpen(true);
                                }
                            }}/>
                        {datePicker}
                    </div>,
                    actions: [{
                        key: "custom-range",
                        type: "custom",
                        hidden: !isMobile,
                        render: datePicker,
                    }, {
                        key: "settings",
                        type: "button",
                        label: <>监控设置{problem && <Icon type="ExclamationCircleOutlined" className={styles.problem}/>}</>,
                        icon: "SettingOutlined",
                        tooltip: problem || undefined,
                        onClick: () => settingsRef.current?.open(),
                    }],
                    refresh: {
                        ariaLabel: "刷新监控数据",
                        loading,
                        onClick: () => { void loadHistory(); },
                    },
                }}
                tableRender={() => <div className={styles.charts}>
                    {charts.map(({capacity, ...chart}) => <MonitorChart key={chart.title} {...chart}
                        subtitle={capacity ? latest ? `${formatSize(latest[capacity[0]])} / ${formatSize(latest[capacity[1]])}` : "-- / --" : ""}
                        history={data} loading={loading} rangeKey={rangeKey}/>)}
                </div>}
            />
            <MonitorSettingsModal
                ref={settingsRef}
                onLoaded={setSettings}
                onSuccess={(value) => {
                    setSettings(value);
                    void loadHistory();
                }}
            />
        </Body>
    );
};
