import React, {useMemo} from "react";
import {Card, Col, Popover, Row, theme as antdTheme} from "antd";
import {createStyles} from "antd-style";
import {Title} from "sinking-antd";
import {formatSize} from "@/utils/string";
import DiskPopover from "../disk-popover";
import {formatPercent, gutter, numberValue, percentValue} from "../../utils";
import ResourceMetric from "../resource-metric";

const useStyles = createStyles(({css, token}: any) => ({
    metricCard: css`
        height: 100%;
        overflow: hidden;
        contain: paint;
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorBgContainer};

        .ant-card-body {
            height: 100%;
            padding: 0;
        }
    `,
}));

const ResourceOverview = React.memo(({info}: any) => {
    const {styles} = useStyles();
    const {token} = antdTheme.useToken();
    const disks: any[] = info?.disks || [];
    const cpu = info?.cpu;
    const memory = info?.memory;
    const load = info?.load;
    const summaryDisks = useMemo<any[]>(() => Array.from(new Map(disks.map((item: any) => {
        const key = `${item.filesystem || item.path}-${numberValue(item?.size?.total)}`;
        return [key, item] as const;
    })).values()), [disks]);
    const diskSummary = useMemo(() => summaryDisks.reduce((summary, item: any) => {
        const total = numberValue(item?.size?.total);
        const used = numberValue(item?.size?.used);
        summary.total += total;
        summary.used += used;
        summary.free += item?.size?.un_used === undefined ? Math.max(0, total - used) : numberValue(item.size.un_used);
        return summary;
    }, {total: 0, used: 0, free: 0}), [summaryDisks]);
    const diskPercent = diskSummary.total > 0 ? diskSummary.used / diskSummary.total * 100 : 0;
    const diskFree = diskSummary.free;
    const loadPercent = numberValue(load?.load1) / Math.max(1, numberValue(cpu?.processor)) * 100;
    const metricColor = (value: any, color: string) => {
        const percent = numberValue(value);
        if (percent >= 95) {
            return token.colorError;
        }
        if (percent >= 80) {
            return token.colorWarning;
        }
        return color;
    };
    let loadDescription = "负载较低，系统资源充足";
    if (loadPercent >= 100) {
        loadDescription = "负载过高，建议检查运行进程";
    } else if (loadPercent >= 75) {
        loadDescription = "负载偏高，建议持续关注";
    } else if (loadPercent >= 40) {
        loadDescription = "负载正常，系统运行平稳";
    }

    return (
        <Col span={24}>
            <Row gutter={[gutter, gutter]}>
                <Col xs={24} md={12} xl={6}>
                    <Card className={styles.metricCard} variant="borderless">
                        <ResourceMetric
                            color={metricColor(loadPercent, token.colorPrimary)}
                            icon="LineChartOutlined"
                            title="系统负载"
                            value={formatPercent(loadPercent)}
                            detail={`1m ${numberValue(load?.load1).toFixed(2)} · 5m ${numberValue(load?.load5).toFixed(2)} · 15m ${numberValue(load?.load15).toFixed(2)}`}
                            meta={loadDescription}
                            percent={percentValue(loadPercent)}
                        />
                    </Card>
                </Col>
                <Col xs={24} md={12} xl={6}>
                    <Card className={styles.metricCard} variant="borderless">
                        <ResourceMetric
                            color={metricColor(cpu?.usage, "#13c2c2")}
                            icon="DashboardOutlined"
                            title="CPU 使用率"
                            value={formatPercent(cpu?.usage)}
                            detail={cpu?.model || "未知处理器"}
                            meta={`${numberValue(cpu?.processor)} 个逻辑处理器`}
                            percent={percentValue(cpu?.usage)}
                        />
                    </Card>
                </Col>
                <Col xs={24} md={12} xl={6}>
                    <Card className={styles.metricCard} variant="borderless">
                        <ResourceMetric
                            color={metricColor(memory?.used_percent, "#52c41a")}
                            icon="DatabaseOutlined"
                            title="内存使用率"
                            value={formatPercent(memory?.used_percent)}
                            detail={`${formatSize(memory?.used)} / ${formatSize(memory?.total)}`}
                            meta={`${formatSize(memory?.free)} 空闲`}
                            percent={percentValue(memory?.used_percent)}
                        />
                    </Card>
                </Col>
                <Col xs={24} md={12} xl={6}>
                    <Popover
                        placement="bottom"
                        trigger="hover"
                        title={<Title size="small">磁盘空间</Title>}
                        content={(
                            <DiskPopover
                                diskFree={diskFree}
                                diskPercent={diskPercent}
                                diskSummary={diskSummary}
                                summaryDisks={summaryDisks}
                            />
                        )}
                    >
                        <Card className={styles.metricCard} variant="borderless">
                            <ResourceMetric
                                color={metricColor(diskPercent, "#5b8ff9")}
                                icon="PieChartOutlined"
                                title="磁盘使用率"
                                value={formatPercent(diskPercent)}
                                detail={`${formatSize(diskSummary.used)} / ${formatSize(diskSummary.total)}`}
                                meta={`${summaryDisks.length} 个挂载点`}
                                percent={percentValue(diskPercent)}
                                detailTooltip={false}
                            />
                        </Card>
                    </Popover>
                </Col>
            </Row>
        </Col>
    );
});

export default ResourceOverview;
