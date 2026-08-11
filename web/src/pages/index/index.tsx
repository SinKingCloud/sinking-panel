import React, {useCallback, useEffect, useRef, useState} from "react";
import {Alert, Col, Row} from "antd";
import {createStyles} from "antd-style";
import {Body} from "sinking-antd";
import {getSystemInfo, getSystemStatus} from "@/service/api/system";
import {gutter, numberValue} from "./components/helper";
import HostOverview from "./components/host-overview";
import MonitorOverview from "./components/monitor-overview";
import ResourceOverview from "./components/resource-overview";

type RefreshMode = "initial" | "manual" | "silent";

const throughputPointCount = 120;
const throughputInitialPointCount = 30;

const useStyles = createStyles(({css}: any) => ({
    page: css`
        width: 100%;
        max-width: 1450px;
        margin: 0 auto;
        font-variant-numeric: tabular-nums;

        > .ant-col {
            min-width: 0;
        }
    `,
}));

export default (): React.ReactNode => {
    const {styles} = useStyles();
    const mountedRef = useRef(true);
    const requestingRef = useRef(false);
    const snapshotRef = useRef({info: "", status: "", error: "", sampleId: 0, sampleTime: 0});
    const [dashboard, setDashboard] = useState<any>(() => {
        const initialTime = Date.now();
        return {
            info: undefined,
            status: undefined,
            throughputHistory: Array.from({length: throughputInitialPointCount}, (_, index) => ({
                time: initialTime - (throughputInitialPointCount - index) * 1000,
                network: {},
                disk: {},
                placeholder: true,
            })),
            error: "",
        };
    });
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const {info, status, throughputHistory, error} = dashboard;

    const loadData = useCallback(async (mode: RefreshMode = "silent") => {
        if (requestingRef.current) {
            return;
        }
        requestingRef.current = true;
        if (mode === "initial") {
            setLoading(true);
        }
        if (mode === "manual") {
            setRefreshing(true);
        }

        try {
            const [infoResponse, statusResponse] = await Promise.all([
                getSystemInfo(),
                getSystemStatus({body: {after: snapshotRef.current.sampleId}}),
            ]);
            if (!mountedRef.current) {
                return;
            }

            const errors: string[] = [];
            const infoData = infoResponse?.code === 200 && infoResponse?.data
                ? infoResponse.data
                : undefined;
            const statusData = statusResponse?.code === 200 && statusResponse?.data
                ? statusResponse.data
                : undefined;
            if (!infoData) {
                errors.push(infoResponse?.message || "系统信息获取失败");
            }
            if (!statusData) {
                errors.push(statusResponse?.message || "系统状态获取失败");
            }

            const nextError = errors.join("；");
            const currentStatusData = statusData ? {...statusData, history: undefined} : undefined;
            const nextInfoSnapshot = JSON.stringify(infoData || null);
            const nextStatusSnapshot = JSON.stringify(currentStatusData || null);
            const infoChanged = nextInfoSnapshot !== snapshotRef.current.info;
            const statusChanged = nextStatusSnapshot !== snapshotRef.current.status;
            const errorChanged = nextError !== snapshotRef.current.error;
            const responseSampleId = numberValue(statusData?.sample_id);
            const sampleReset = responseSampleId > 0 && responseSampleId < snapshotRef.current.sampleId;
            const sampleCursor = sampleReset ? 0 : snapshotRef.current.sampleId;
            const sampleTimeCursor = sampleReset ? 0 : snapshotRef.current.sampleTime;
            const responseHistory = Array.isArray(statusData?.history) ? statusData.history : [];
            const responseSampleTime = numberValue(statusData?.sample_time);
            const fallbackHistory = responseHistory.length === 0 && responseSampleTime > 0 && (sampleCursor === 0
                || responseSampleTime > sampleTimeCursor)
                ? [statusData]
                : [];
            const statusSamples = [...responseHistory, ...fallbackHistory]
                .filter((item: any) => {
                    const sampleId = numberValue(item?.sample_id);
                    return sampleId > 0
                        ? sampleId > sampleCursor || (sampleId === sampleCursor
                            && numberValue(item?.sample_time) > sampleTimeCursor)
                        : numberValue(item?.sample_time) > sampleTimeCursor;
                })
                .sort((first: any, second: any) => {
                    const idDiff = numberValue(first?.sample_id) - numberValue(second?.sample_id);
                    return idDiff || numberValue(first?.sample_time) - numberValue(second?.sample_time);
                });
            const samples = statusSamples.map((item: any) => ({
                sampleId: numberValue(item.sample_id),
                time: numberValue(item.sample_time),
                network: Object.fromEntries((Object.entries(item.network || {}) as [string, any][]).map(([name, value]) => [name, {
                    receiveRate: numberValue(value.recv_rate),
                    sendRate: numberValue(value.send_rate),
                }])),
                disk: Object.fromEntries((Object.entries(item.disk || {}) as [string, any][]).map(([name, value]) => [name, {
                    readRate: numberValue(value.read_bytes_rate),
                    writeRate: numberValue(value.write_bytes_rate),
                }])),
            }));
            const latestSampleId = Math.max(
                sampleCursor,
                responseSampleId,
                ...samples.map((item: any) => item.sampleId),
            );
            const latestSampleTime = Math.max(
                sampleTimeCursor,
                numberValue(statusData?.sample_time),
                ...samples.map((item: any) => item.time),
            );
            const statusSampleChanged = samples.length > 0;
            if (infoChanged || statusChanged || errorChanged || statusSampleChanged) {
                snapshotRef.current = {
                    info: nextInfoSnapshot,
                    status: nextStatusSnapshot,
                    error: nextError,
                    sampleId: latestSampleId,
                    sampleTime: latestSampleTime,
                };
                const commit = () => setDashboard((current: any) => {
                    let throughputHistory = current.throughputHistory;
                    if (samples.length > 0) {
                        const history = new Map<number, any>();
                        if (!sampleReset) {
                            current.throughputHistory.forEach((item: any) => {
                                if (!item.placeholder) {
                                    history.set(item.time, item);
                                }
                            });
                        }
                        samples.forEach((item: any) => history.set(item.time, item));
                        const realHistory = Array.from(history.values())
                            .sort((first: any, second: any) => first.time - second.time)
                            .slice(-throughputPointCount);
                        const placeholderCount = Math.max(0, throughputInitialPointCount - realHistory.length);
                        const firstTime = realHistory[0]?.time || Date.now();
                        const placeholders = Array.from({length: placeholderCount}, (_, index) => ({
                            time: firstTime - (placeholderCount - index) * 1000,
                            network: {},
                            disk: {},
                            placeholder: true,
                        }));
                        throughputHistory = [...placeholders, ...realHistory].slice(-throughputPointCount);
                    }

                    return {
                        info: infoChanged && infoData ? infoData : current.info,
                        status: statusChanged && currentStatusData ? currentStatusData : current.status,
                        throughputHistory,
                        error: nextError,
                    };
                });

                if (mode === "silent") {
                    React.startTransition(commit);
                } else {
                    commit();
                }
            }
        } catch (requestError: any) {
            if (mountedRef.current) {
                const nextError = requestError?.message || "系统状态获取失败";
                if (nextError !== snapshotRef.current.error) {
                    snapshotRef.current = {...snapshotRef.current, error: nextError};
                    const commit = () => setDashboard((current: any) => ({
                        ...current,
                        error: nextError,
                    }));
                    if (mode === "silent") {
                        React.startTransition(commit);
                    } else {
                        commit();
                    }
                }
            }
        } finally {
            requestingRef.current = false;
            if (mountedRef.current) {
                if (mode === "initial") {
                    setLoading(false);
                }
                if (mode === "manual") {
                    setRefreshing(false);
                }
            }
        }
    }, []);

    useEffect(() => {
        mountedRef.current = true;
        let stopped = false;
        let timer = 0;
        let nextPollAt = Date.now() + 1000;
        const schedulePoll = () => {
            const delay = Math.max(0, nextPollAt - Date.now());
            timer = window.setTimeout(async () => {
                if (stopped) {
                    return;
                }
                if (document.visibilityState === "visible") {
                    await loadData("silent");
                }
                nextPollAt += 1000;
                if (nextPollAt < Date.now() - 1000) {
                    nextPollAt = Date.now();
                }
                schedulePoll();
            }, delay);
        };
        void loadData("initial").finally(() => {
            if (!stopped) {
                nextPollAt = Date.now() + 1000;
                schedulePoll();
            }
        });
        const handleVisibilityChange = () => {
            if (document.visibilityState === "visible") {
                nextPollAt = Date.now() + 1000;
                void loadData("silent");
            }
        };
        document.addEventListener("visibilitychange", handleVisibilityChange);
        return () => {
            stopped = true;
            mountedRef.current = false;
            window.clearTimeout(timer);
            document.removeEventListener("visibilitychange", handleVisibilityChange);
        };
    }, [loadData]);

    return (
        <Body loading={loading} space={false}>
            <Row className={styles.page} gutter={[0, gutter]}>
                {error && (
                    <Col span={24}>
                        <Alert type="warning" showIcon message={error}/>
                    </Col>
                )}
                <Col span={24}>
                    <HostOverview
                        info={info}
                        refreshing={refreshing}
                        onRefresh={() => loadData("manual")}
                    />
                </Col>
                <ResourceOverview info={info}/>
                <Col span={24}>
                    <MonitorOverview status={status} throughputHistory={throughputHistory}/>
                </Col>
            </Row>
        </Body>
    );
};
