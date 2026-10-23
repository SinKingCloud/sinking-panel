import React, {useCallback, useEffect, useRef, useState} from "react";
import {Col, Row} from "antd";
import {createStyles} from "antd-style";
import {Body} from "sinking-antd";
import {getSystemStatis, getSystemIO, getSystemInfo} from "@/service/api/system";
import {gutter, numberValue} from "./utils";
import HostOverview from "./components/host-overview";
import MonitorOverview from "./components/monitor-overview";
import ResourceOverview from "./components/resource-overview";

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
    const snapshotRef = useRef({statis: "", info: "", io: "", sampleId: 0, sampleTime: 0});
    const [dashboard, setDashboard] = useState<any>(() => {
        const initialTime = Date.now();
        return {
            statis: undefined,
            info: undefined,
            io: undefined,
            throughputHistory: Array.from({length: throughputInitialPointCount}, (_, index) => ({
                time: initialTime - (throughputInitialPointCount - index) * 1000,
                network: {},
                disk: {},
                placeholder: true,
            })),
        };
    });
    const [loading, setLoading] = useState(true);
    const {statis, info, io, throughputHistory} = dashboard;

    const loadData = useCallback(async (initial = false) => {
        if (requestingRef.current) {
            return;
        }
        requestingRef.current = true;
        if (initial) {
            setLoading(true);
        }

        try {
            const [statisResponse, infoResponse, ioResponse] = await Promise.all([
                getSystemStatis(),
                getSystemInfo(),
                getSystemIO({body: {after: snapshotRef.current.sampleId}}),
            ]);
            if (!mountedRef.current) {
                return;
            }

            const statisData = statisResponse?.code === 200 && statisResponse?.data
                ? statisResponse.data
                : undefined;
            const infoData = infoResponse?.code === 200 && infoResponse?.data
                ? infoResponse.data
                : undefined;
            const ioData = ioResponse?.code === 200 && ioResponse?.data
                ? ioResponse.data
                : undefined;
            const currentIOData = ioData ? {...ioData, history: undefined} : undefined;
            const nextStatisSnapshot = JSON.stringify(statisData || null);
            const nextInfoSnapshot = JSON.stringify(infoData || null);
            const nextIOSnapshot = JSON.stringify(currentIOData || null);
            const statisChanged = nextStatisSnapshot !== snapshotRef.current.statis;
            const infoChanged = nextInfoSnapshot !== snapshotRef.current.info;
            const ioChanged = nextIOSnapshot !== snapshotRef.current.io;
            const responseSampleId = numberValue(ioData?.sample_id);
            const sampleReset = responseSampleId > 0 && responseSampleId < snapshotRef.current.sampleId;
            const sampleCursor = sampleReset ? 0 : snapshotRef.current.sampleId;
            const sampleTimeCursor = sampleReset ? 0 : snapshotRef.current.sampleTime;
            const responseHistory = Array.isArray(ioData?.history) ? ioData.history : [];
            const responseSampleTime = numberValue(ioData?.sample_time);
            const fallbackHistory = responseHistory.length === 0 && responseSampleTime > 0 && (sampleCursor === 0
                || responseSampleTime > sampleTimeCursor)
                ? [ioData]
                : [];
            const ioSamples = [...responseHistory, ...fallbackHistory]
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
            const samples = ioSamples.map((item: any) => ({
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
                numberValue(ioData?.sample_time),
                ...samples.map((item: any) => item.time),
            );
            const ioSampleChanged = samples.length > 0;
            if (statisChanged || infoChanged || ioChanged || ioSampleChanged) {
                snapshotRef.current = {
                    statis: nextStatisSnapshot,
                    info: nextInfoSnapshot,
                    io: nextIOSnapshot,
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
                        statis: statisChanged && statisData ? statisData : current.statis,
                        info: infoChanged && infoData ? infoData : current.info,
                        io: ioChanged && currentIOData ? currentIOData : current.io,
                        throughputHistory,
                    };
                });

                if (!initial) {
                    React.startTransition(commit);
                } else {
                    commit();
                }
            }
        } catch {
            // 请求失败时保留已有数据，下一次轮询继续重试。
        } finally {
            requestingRef.current = false;
            if (mountedRef.current) {
                if (initial) {
                    setLoading(false);
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
                    await loadData();
                }
                nextPollAt += 1000;
                if (nextPollAt < Date.now() - 1000) {
                    nextPollAt = Date.now();
                }
                schedulePoll();
            }, delay);
        };
        void loadData(true).finally(() => {
            if (!stopped) {
                nextPollAt = Date.now() + 1000;
                schedulePoll();
            }
        });
        const handleVisibilityChange = () => {
            if (document.visibilityState === "visible") {
                nextPollAt = Date.now() + 1000;
                void loadData();
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
        <Body loading={loading}>
            <Row className={styles.page} gutter={[0, gutter]}>
                <Col span={24}>
                    <HostOverview
                        statis={statis}
                        info={info}
                    />
                </Col>
                <ResourceOverview statis={statis} diskSizeBase={info?.os === "darwin" ? 1000 : 1024}/>
                <Col span={24}>
                    <MonitorOverview io={io} throughputHistory={throughputHistory}/>
                </Col>
            </Row>
        </Body>
    );
};
