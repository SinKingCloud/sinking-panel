import React, {useEffect, useMemo, useState} from "react";
import {Col, Row, theme as antdTheme} from "antd";
import {createStyles} from "antd-style";
import {useTheme} from "sinking-antd";
import {formatSize} from "@/utils/string";
import {formatCount, gutter, numberValue} from "./helper";
import MonitorCard from "./monitor-card";

const allDevices = "__all__";
const diskWriteColor = "#13c2c2";

const useStyles = createStyles(({css}: any) => ({
    monitorColumn: css`
        min-width: 0;
        display: flex;

        > .ant-card {
            width: 100%;
        }
    `,
}));

const MonitorOverview = React.memo(({status, throughputHistory}: any) => {
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const theme = useTheme();
    const [networkDevice, setNetworkDevice] = useState(allDevices);
    const [diskDevice, setDiskDevice] = useState(allDevices);

    const networkList = useMemo<any[]>(() => {
        return (Object.entries(status?.network || {}) as [string, any][]).map(([name, item]) => ({
            ...item,
            key: name,
            name,
        })).sort((left: any, right: any) => {
            const leftRate = numberValue(left.recv_rate) + numberValue(left.send_rate);
            const rightRate = numberValue(right.recv_rate) + numberValue(right.send_rate);
            return rightRate - leftRate;
        });
    }, [status?.network]);

    const diskStatusList = useMemo<any[]>(() => (Object.entries(status?.disk || {}) as [string, any][]).map(([name, item]) => ({
        ...item,
        key: name,
        name: item?.name || name,
    })).sort((left: any, right: any) => {
        const leftRate = numberValue(left.read_bytes_rate) + numberValue(left.write_bytes_rate);
        const rightRate = numberValue(right.read_bytes_rate) + numberValue(right.write_bytes_rate);
        return rightRate - leftRate;
    }), [status?.disk]);

    const networkOptions = useMemo(() => [
        {label: "全部接口", value: allDevices},
        ...networkList.map((item: any) => ({label: item.name, value: item.key})),
    ], [networkList]);
    const diskOptions = useMemo(() => [
        {label: "全部磁盘", value: allDevices},
        ...diskStatusList.map((item: any) => ({label: item.name, value: item.key})),
    ], [diskStatusList]);

    useEffect(() => {
        if (networkDevice !== allDevices && !networkList.some((item: any) => item.key === networkDevice)) {
            setNetworkDevice(allDevices);
        }
    }, [networkDevice, networkList]);

    useEffect(() => {
        if (diskDevice !== allDevices && !diskStatusList.some((item: any) => item.key === diskDevice)) {
            setDiskDevice(allDevices);
        }
    }, [diskDevice, diskStatusList]);

    const selectedNetworkList = useMemo(() => networkDevice === allDevices
        ? networkList
        : networkList.filter((item: any) => item.key === networkDevice), [networkDevice, networkList]);
    const selectedDiskList = useMemo(() => diskDevice === allDevices
        ? diskStatusList
        : diskStatusList.filter((item: any) => item.key === diskDevice), [diskDevice, diskStatusList]);

    const networkThroughput = useMemo(() => selectedNetworkList.reduce((total, item: any) => ({
        receiveRate: total.receiveRate + numberValue(item.recv_rate),
        sendRate: total.sendRate + numberValue(item.send_rate),
        receiveTotal: total.receiveTotal + numberValue(item.bytes_recv),
        sendTotal: total.sendTotal + numberValue(item.bytes_sent),
    }), {receiveRate: 0, sendRate: 0, receiveTotal: 0, sendTotal: 0}), [selectedNetworkList]);

    const diskThroughput = useMemo(() => selectedDiskList.reduce((total, item: any) => ({
        readRate: total.readRate + numberValue(item.read_bytes_rate),
        writeRate: total.writeRate + numberValue(item.write_bytes_rate),
        readCount: total.readCount + numberValue(item.read_count_rate),
        writeCount: total.writeCount + numberValue(item.write_count_rate),
    }), {readRate: 0, writeRate: 0, readCount: 0, writeCount: 0}), [selectedDiskList]);

    const networkMetrics = useMemo(() => [
        {
            label: "下载速率",
            value: formatSize(networkThroughput.receiveRate),
            suffix: "/s",
            color: token.colorSuccess,
        },
        {
            label: "上传速率",
            value: formatSize(networkThroughput.sendRate),
            suffix: "/s",
            color: token.colorPrimary,
        },
        {
            label: "累计接收",
            value: formatSize(networkThroughput.receiveTotal),
            suffix: "",
            color: token.colorSuccess,
        },
        {
            label: "累计发送",
            value: formatSize(networkThroughput.sendTotal),
            suffix: "",
            color: token.colorPrimary,
        },
    ], [networkThroughput, token.colorPrimary, token.colorSuccess]);
    const diskMetrics = useMemo(() => [
        {
            label: "读取速率",
            value: formatSize(diskThroughput.readRate),
            suffix: "/s",
            color: token.colorInfo,
        },
        {
            label: "写入速率",
            value: formatSize(diskThroughput.writeRate),
            suffix: "/s",
            color: diskWriteColor,
        },
        {
            label: "读取 IOPS",
            value: formatCount(diskThroughput.readCount),
            suffix: "次/s",
            color: token.colorInfo,
        },
        {
            label: "写入 IOPS",
            value: formatCount(diskThroughput.writeCount),
            suffix: "次/s",
            color: diskWriteColor,
        },
    ], [diskThroughput, token.colorInfo]);

    const networkChartData = useMemo(() => throughputHistory.map((sample: any) => {
        const data = sample.network || {};
        const items: any[] = networkDevice === allDevices
            ? Object.values(data)
            : (data[networkDevice] ? [data[networkDevice]] : []);
        return {
            time: sample.time,
            first: items.reduce((total, item) => total + numberValue(item.receiveRate), 0),
            second: items.reduce((total, item) => total + numberValue(item.sendRate), 0),
        };
    }), [networkDevice, throughputHistory]);
    const diskChartData = useMemo(() => throughputHistory.map((sample: any) => {
        const data = sample.disk || {};
        const items: any[] = diskDevice === allDevices
            ? Object.values(data)
            : (data[diskDevice] ? [data[diskDevice]] : []);
        return {
            time: sample.time,
            first: items.reduce((total, item) => total + numberValue(item.readRate), 0),
            second: items.reduce((total, item) => total + numberValue(item.writeRate), 0),
        };
    }), [diskDevice, throughputHistory]);

    return (
        <Row gutter={[gutter, gutter]} align="stretch">
            <Col className={styles.monitorColumn} xs={24} xl={12}>
                <MonitorCard
                    title="网络吞吐"
                    device={networkDevice}
                    options={networkOptions}
                    onDeviceChange={setNetworkDevice}
                    metrics={networkMetrics}
                    data={networkChartData}
                    first={{label: "下载", color: token.colorSuccess}}
                    second={{label: "上传", color: token.colorPrimary}}
                    isDarkMode={theme?.isDarkMode?.()}
                />
            </Col>
            <Col className={styles.monitorColumn} xs={24} xl={12}>
                <MonitorCard
                    title="磁盘 IO"
                    device={diskDevice}
                    options={diskOptions}
                    onDeviceChange={setDiskDevice}
                    metrics={diskMetrics}
                    data={diskChartData}
                    first={{label: "读取", color: token.colorInfo}}
                    second={{label: "写入", color: diskWriteColor}}
                    isDarkMode={theme?.isDarkMode?.()}
                />
            </Col>
        </Row>
    );
});

MonitorOverview.displayName = "MonitorOverview";

export default MonitorOverview;
