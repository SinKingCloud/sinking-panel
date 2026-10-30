import React, {useCallback, useEffect, useRef, useState} from "react";
import {Col, Row} from "antd";
import {createStyles} from "antd-style";
import {Body} from "sinking-antd";
import {useModel} from "umi";
import {getSystemStatis, getSystemIO} from "@/service/api/system";
import {formatUptime, gutter, numberValue} from "./utils";
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
    const requestRef = useRef(0);
    const systemInfo = useModel("system", (model) => model.info);
    const infoRef = useRef(systemInfo);
    infoRef.current = systemInfo;
    const snapshotRef = useRef({statis: "", disks: "", info: "", io: "", sampleId: 0, sampleTime: 0});
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
        const request = ++requestRef.current;
        if (initial) {
            setLoading(true);
        }

        try {
            // 主机信息复用全局数据；实时资源与 I/O 仍每秒获取。
            const [statisResult, ioResult] = await Promise.allSettled([
                getSystemStatis(),
                getSystemIO({body: {after: snapshotRef.current.sampleId}}),
            ]);
            if (!mountedRef.current || request !== requestRef.current || document.visibilityState === "hidden") {
                return;
            }
            const statisResponse = statisResult.status === "fulfilled" ? statisResult.value : undefined;
            const ioResponse = ioResult.status === "fulfilled" ? ioResult.value : undefined;
            const statisData = statisResponse?.code === 200 && statisResponse?.data
                ? {cpu: statisResponse.data.cpu, memory: statisResponse.data.memory, load: statisResponse.data.load, disks: statisResponse.data.disks}
                : undefined;
            const ioData = ioResponse?.code === 200 && ioResponse?.data
                ? ioResponse?.data
                : undefined;
            const baseInfo = infoRef.current;
            const sampleTime = numberValue(ioData?.sample_time) || snapshotRef.current.sampleTime;
            const infoData = baseInfo ? {
                ...baseInfo,
                // 使用服务器采样时间，避免浏览器与服务器的时钟差影响运行时长。
                uptime: sampleTime > 0 && baseInfo.boot_time > 0
                    ? Math.max(numberValue(baseInfo.uptime), Math.floor(sampleTime / 1000) - baseInfo.boot_time)
                    : numberValue(baseInfo.uptime),
            } : undefined;
            const currentIOData = ioData ? {...ioData, history: undefined} : undefined;
            const nextStatisSnapshot = statisData ? JSON.stringify([statisData.cpu, statisData.memory, statisData.load]) : snapshotRef.current.statis;
            const nextDisksSnapshot = statisData ? JSON.stringify(statisData.disks || []) : snapshotRef.current.disks;
            const nextInfoSnapshot = infoData ? JSON.stringify([
                infoData.hostname, infoData.os, infoData.platform, infoData.platform_version,
                infoData.kernel_version, infoData.kernel_arch, infoData.boot_time, formatUptime(infoData.uptime),
            ]) : snapshotRef.current.info;
            const nextIOSnapshot = currentIOData ? JSON.stringify([currentIOData.network, currentIOData.disk]) : snapshotRef.current.io;
            const disksChanged = nextDisksSnapshot !== snapshotRef.current.disks;
            const statisChanged = nextStatisSnapshot !== snapshotRef.current.statis || disksChanged;
            const infoChanged = nextInfoSnapshot !== snapshotRef.current.info;
            const ioChanged = nextIOSnapshot !== snapshotRef.current.io;
            const responseSampleId = numberValue(ioData?.sample_id);
            const sampleReset = responseSampleId > 0 && responseSampleId < snapshotRef.current.sampleId;
            const sampleCursor = sampleReset ? 0 : snapshotRef.current.sampleId;
            const sampleTimeCursor = sampleReset ? 0 : snapshotRef.current.sampleTime;
            const responseHistory = Array.isArray(ioData?.history) ? ioData?.history : [];
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
            const samples = ioSamples.map((item: any) => {
                const networkTotal = {first: 0, second: 0};
                const diskTotal = {first: 0, second: 0};
                const network = Object.fromEntries((Object.entries(item?.network || {}) as [string, any][]).map(([name, value]) => {
                    const receiveRate = numberValue(value?.recv_rate);
                    const sendRate = numberValue(value?.send_rate);
                    networkTotal.first += receiveRate;
                    networkTotal.second += sendRate;
                    return [name, {receiveRate, sendRate}];
                }));
                const disk = Object.fromEntries((Object.entries(item?.disk || {}) as [string, any][]).map(([name, value]) => {
                    const readRate = numberValue(value?.read_bytes_rate);
                    const writeRate = numberValue(value?.write_bytes_rate);
                    diskTotal.first += readRate;
                    diskTotal.second += writeRate;
                    return [name, {readRate, writeRate}];
                }));
                return {sampleId: numberValue(item?.sample_id), time: numberValue(item?.sample_time), network, disk, networkTotal, diskTotal};
            });
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
                    disks: nextDisksSnapshot,
                    info: nextInfoSnapshot,
                    io: nextIOSnapshot,
                    sampleId: latestSampleId,
                    sampleTime: latestSampleTime,
                };
                const commit = () => setDashboard((current: any) => {
                    let throughputHistory = current.throughputHistory;
                    if (samples.length > 0) {
                        const previous = sampleReset ? [] : current.throughputHistory.filter((item: any) => !item.placeholder);
                        const ordered = samples.every((item: any, index: number) => index === 0 || item.time > samples[index - 1].time);
                        // 正常增量直接拼接；只有重叠/乱序批次才需要按时间去重排序。
                        const appendOnly = ordered && (!previous.length || samples[0].time > previous[previous.length - 1].time);
                        const realHistory = (appendOnly ? [...previous, ...samples]
                            : Array.from(new Map([...previous, ...samples].map((item: any) => [item.time, item])).values())
                                .sort((first: any, second: any) => first.time - second.time))
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
                        statis: statisChanged && statisData ? {...statisData, disks: disksChanged ? statisData.disks : current.statis?.disks} : current.statis,
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
            if (request === requestRef.current) requestingRef.current = false;
            if (mountedRef.current && request === requestRef.current) {
                if (initial) {
                    setLoading(false);
                }
            }
        }
    }, []);

    useEffect(() => {
        mountedRef.current = true;
        let stopped = false;
        let polling = false;
        let initial = true;
        let timer = 0;
        let nextPollAt = Date.now();
        const poll = async () => {
            if (stopped || polling || document.visibilityState === "hidden") return;
            polling = true;
            await loadData(initial);
            initial = false;
            polling = false;
            if (stopped || document.hidden) return;
            nextPollAt = Math.max(nextPollAt + 1000, Date.now());
            timer = window.setTimeout(poll, Math.max(0, nextPollAt - Date.now()));
        };
        void poll();
        const handleVisibilityChange = () => {
            window.clearTimeout(timer);
            if (document.visibilityState === "visible") {
                nextPollAt = Date.now();
                void poll();
            }
        };
        document.addEventListener("visibilitychange", handleVisibilityChange);
        return () => {
            stopped = true;
            mountedRef.current = false;
            requestRef.current += 1;
            requestingRef.current = false;
            window.clearTimeout(timer);
            document.removeEventListener("visibilitychange", handleVisibilityChange);
        };
    }, [loadData]);

    return (
        <Body loading={loading}>
            <Row className={styles.page} gutter={[0, gutter]}>
                <Col span={24}>
                    <HostOverview
                        info={info}
                        cpuCount={statis?.cpu?.processor}
                        memoryTotal={statis?.memory?.total}
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
