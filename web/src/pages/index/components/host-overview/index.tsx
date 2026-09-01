import React, {useState} from "react";
import {Button, Card, Col, Row, Tooltip, Typography} from "antd";
import {createStyles} from "antd-style";
import dayjs from "dayjs";
import {Icon} from "sinking-antd";
import {formatSize} from "@/utils/string";
import {formatUptime, gutter, numberValue} from "../../utils";

const systemVisuals: any = {
    windows: {icon: "WindowsOutlined", name: "Windows"},
    linux: {icon: "LinuxOutlined", name: "Linux"},
    darwin: {icon: "AppleOutlined", name: "macOS"},
    ios: {icon: "AppleOutlined", name: "iOS"},
    android: {icon: "AndroidOutlined", name: "Android"},
    freebsd: {icon: "CodeOutlined", name: "FreeBSD"},
    openbsd: {icon: "CodeOutlined", name: "OpenBSD"},
    netbsd: {icon: "CodeOutlined", name: "NetBSD"},
    dragonfly: {icon: "CodeOutlined", name: "DragonFly BSD"},
    solaris: {icon: "GlobalOutlined", name: "Solaris"},
    illumos: {icon: "GlobalOutlined", name: "illumos"},
    aix: {icon: "GlobalOutlined", name: "AIX"},
};

const useStyles = createStyles(({css, token, isDarkMode}: any) => ({
    hostCard: css`
        overflow: hidden;
        contain: paint;
        border-radius: ${token.borderRadiusLG}px;
        background: ${token.colorBgContainer};

        .ant-card-body {
            padding: 0;
        }
    `,
    host: css`
        position: relative;
        min-height: 108px;
        padding: 14px 46px 14px 16px;
        overflow: hidden;
        isolation: isolate;
        background:
            radial-gradient(ellipse at 2% -35%, color-mix(in srgb, ${token.colorPrimary} ${isDarkMode ? 18 : 12}%, transparent) 0%, transparent 42%),
            radial-gradient(ellipse at 50% 135%, color-mix(in srgb, ${token.colorInfo} ${isDarkMode ? 13 : 9}%, transparent) 0%, transparent 38%),
            radial-gradient(ellipse at 98% -35%, color-mix(in srgb, ${token.colorSuccess} ${isDarkMode ? 10 : 7}%, transparent) 0%, transparent 40%),
            ${token.colorBgContainer};

        &::before,
        &::after {
            content: "";
            position: absolute;
            z-index: 0;
            width: 270px;
            height: 120px;
            border-radius: 999px;
            pointer-events: none;
            filter: blur(34px) saturate(1.12);
            transform: translate3d(0, 0, 0);
            transition: opacity .28s ease, transform .36s cubic-bezier(.2, .8, .2, 1);
            will-change: opacity, transform;
        }

        &::before {
            top: -78px;
            left: -58px;
            background: color-mix(in srgb, ${token.colorPrimary} ${isDarkMode ? 50 : 38}%, transparent);
            opacity: ${isDarkMode ? .34 : .3};
        }

        &::after {
            right: -48px;
            bottom: -82px;
            background: color-mix(in srgb, ${token.colorSuccess} ${isDarkMode ? 42 : 32}%, transparent);
            opacity: ${isDarkMode ? .26 : .22};
        }

        > .ant-row {
            position: relative;
            z-index: 2;
            min-width: 0;
        }

        @media (hover: hover) {
            &:hover::before {
                opacity: ${isDarkMode ? .42 : .36};
                transform: translate3d(10px, 5px, 0) scale(1.06);
            }

            &:hover::after {
                opacity: ${isDarkMode ? .34 : .28};
                transform: translate3d(-8px, -4px, 0) scale(1.08);
            }
        }

        @media (prefers-reduced-motion: reduce) {
            &::before,
            &::after {
                transition: none;
            }
        }

        @media (max-width: 991px) {
            min-height: auto;
            padding: 14px 44px 14px 16px;
        }

        @media (max-width: 575px) {
            padding: 12px 14px;
        }
    `,
    hostIdentity: css`
        position: relative;
        min-width: 0;
        min-height: 78px;
        align-items: center;

        > .ant-col {
            min-width: 0;
        }

        @media (max-width: 991px) {
            min-height: auto;
        }

        @media (max-width: 575px) {
            min-height: 52px;
            padding-right: 30px;
        }
    `,
    hostWatermark: css`
        position: absolute;
        z-index: 1;
        top: 50%;
        right: 6%;
        width: 230px;
        height: 230px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: ${token.colorPrimary};
        font-size: 208px;
        line-height: 1;
        opacity: ${isDarkMode ? .045 : .04};
        pointer-events: none;
        filter: blur(${isDarkMode ? 11 : 9}px);
        transform: translateY(-48%) rotate(-18deg);

        > span {
            font-size: inherit;
            line-height: inherit;
        }

        @media (max-width: 991px) {
            right: -10px;
            width: 204px;
            height: 204px;
            font-size: 184px;
            opacity: ${isDarkMode ? .04 : .035};
        }

        @media (max-width: 575px) {
            top: 42%;
            right: -46px;
            width: 148px;
            height: 148px;
            font-size: 132px;
            opacity: ${isDarkMode ? .028 : .022};
        }
    `,
    systemMark: css`
        position: relative;
        width: 72px;
        height: 72px;
        flex: none;
        color: ${token.colorPrimary};

        .system-visual {
            width: 100%;
            height: 100%;
            display: block;
            overflow: visible;
        }

        .system-core {
            fill: color-mix(in srgb, ${token.colorPrimaryBg} 78%, ${token.colorBgContainer});
            stroke: color-mix(in srgb, ${token.colorPrimary} 34%, ${token.colorBorderSecondary});
            stroke-width: 1.2;
        }

        .system-ring,
        .system-arc {
            fill: none;
            stroke: currentColor;
            vector-effect: non-scaling-stroke;
        }

        .system-ring {
            stroke-width: 1;
            stroke-dasharray: 3 7;
            opacity: ${isDarkMode ? .34 : .26};
        }

        .system-arc {
            stroke-width: 1.8;
            stroke-linecap: round;
            stroke-dasharray: 42 156;
            opacity: ${isDarkMode ? .7 : .6};
        }

        .system-orbit {
            transform-box: view-box;
            transform-origin: 48px 48px;
            animation: systemOrbit 14s linear infinite;
        }

        .system-node {
            fill: ${token.colorBgContainer};
            stroke: currentColor;
            stroke-width: 1.5;
            vector-effect: non-scaling-stroke;
        }

        .system-node-active {
            fill: ${token.colorPrimary};
            stroke: ${token.colorBgContainer};
            stroke-width: 2.5;
        }

        .system-icon {
            position: absolute;
            inset: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 26px;
            line-height: 1;
        }

        @keyframes systemOrbit {
            to {
                transform: rotate(360deg);
            }
        }

        @media (prefers-reduced-motion: reduce) {
            .system-orbit {
                animation: none;
            }
        }

        @media (max-width: 575px) {
            width: 52px;
            height: 52px;

            .system-icon { font-size: 20px; }
        }
    `,
    hostInfo: css`
        min-width: 0;
        flex: 1;
    `,
    systemLabel: css`
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 8px;
        color: ${token.colorPrimary};
        font-size: 10px;
        font-weight: 600;
        line-height: 16px;

        &::before {
            content: "";
            width: 18px;
            height: 2px;
            flex: none;
            border-radius: 1px;
            background: currentColor;
            opacity: .72;
        }

        @media (max-width: 575px) {
            gap: 6px;
            font-size: 9px;
            line-height: 14px;

            &::before {
                width: 12px;
            }
        }
    `,
    hostName: css`
        margin: 3px 0 0;
        color: ${token.colorTextHeading};
        font-size: 21px;
        font-weight: 700;
        line-height: 28px;
        letter-spacing: 0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;

        @media (max-width: 575px) {
            margin-top: 2px;
            font-size: 17px;
            line-height: 23px;
        }
    `,
    hostSummary: css`
        min-width: 0;
        height: 100%;
        border-left: 1px solid color-mix(in srgb, ${token.colorBorderSecondary} 72%, transparent);

        > .ant-row {
            height: 100%;
        }

        > .ant-row > .ant-col {
            min-width: 0;
            display: flex;
            align-items: center;
        }

        > .ant-row > .ant-col + .ant-col {
            border-left: 1px solid color-mix(in srgb, ${token.colorBorderSecondary} 72%, transparent);
        }

        .host-summary-item {
            width: 100%;
            min-width: 0;
            padding: 3px 14px;
        }

        .summary-label,
        .summary-value,
        .summary-meta {
            display: block;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .summary-label {
            display: flex;
            align-items: center;
            gap: 6px;
            color: ${token.colorTextTertiary};
            font-size: 10px;
            line-height: 16px;
        }

        .summary-label > span {
            flex: none;
            color: ${token.colorPrimary};
            font-size: 12px;
        }

        .summary-value {
            margin-top: 3px;
            color: ${token.colorTextHeading};
            font-size: 15px;
            font-weight: 700;
            line-height: 22px;
        }

        .summary-meta {
            margin-top: 1px;
            color: ${token.colorTextQuaternary};
            font-size: 10px;
            line-height: 15px;
        }

        @media (max-width: 991px) {
            height: auto;
            border-top: 1px solid color-mix(in srgb, ${token.colorBorderSecondary} 72%, transparent);
            border-left: 0;

            .host-summary-item {
                padding: 3px 12px;
            }
        }

        @media (max-width: 575px) {
            > .ant-row > .ant-col + .ant-col {
                border-top: 0;
                border-left: 1px solid color-mix(in srgb, ${token.colorBorderSecondary} 72%, transparent);
            }

            .host-summary-item {
                padding: 8px 6px 7px;
                text-align: center;
            }

            .summary-label {
                justify-content: center;
                gap: 4px;
                font-size: 9px;
                line-height: 14px;
            }

            .summary-label > span {
                font-size: 11px;
            }

            .summary-value {
                margin-top: 2px;
                font-size: 12px;
                line-height: 18px;
            }

            .summary-meta {
                display: none;
            }
        }
    `,
    refreshButton: css`
        position: absolute;
        top: 10px;
        right: 10px;

        &.ant-btn {
            width: 30px;
            height: 30px;
            padding: 0;
            color: ${token.colorTextSecondary};
            background: transparent;
        }

        &.ant-btn:hover {
            color: ${token.colorPrimary} !important;
            background: ${token.colorFillTertiary} !important;
        }
    `,
}));

const HostOverview = React.memo(({info, refreshing, onRefresh}: any) => {
    const {styles} = useStyles();
    const [operatingSystemEllipsis, setOperatingSystemEllipsis] = useState(false);
    const system = info?.system;
    const cpu = info?.cpu;
    const memory = info?.memory;
    const systemVisual = systemVisuals[String(system?.os || "").toLowerCase()] || {
        icon: "DesktopOutlined",
        name: system?.os || "未知系统",
    };
    const operatingSystem = [system?.platform, system?.platform_version].filter(Boolean).join(" ") || systemVisual.name;

    return (
        <Card className={styles.hostCard} variant="borderless">
            <section className={styles.host}>
                <span className={styles.hostWatermark}><Icon type={systemVisual.icon}/></span>
                <Row gutter={[gutter, gutter]} align="middle">
                    <Col xs={24} lg={10}>
                        <Row className={styles.hostIdentity} gutter={gutter} wrap={false}>
                            <Col flex="none">
                                <div className={styles.systemMark}>
                                    <svg className="system-visual" viewBox="0 0 96 96" aria-hidden="true">
                                        <circle className="system-core" cx="48" cy="48" r="25"/>
                                        <circle className="system-ring" cx="48" cy="48" r="39"/>
                                        <g className="system-orbit">
                                            <circle className="system-arc" cx="48" cy="48" r="33"/>
                                            <circle className="system-node system-node-active" cx="48" cy="9" r="3"/>
                                            <circle className="system-node" cx="87" cy="48" r="2.5"/>
                                            <circle className="system-node" cx="48" cy="87" r="2.5"/>
                                        </g>
                                    </svg>
                                    <span className="system-icon"><Icon type={systemVisual.icon}/></span>
                                </div>
                            </Col>
                            <Col flex="auto">
                                <div className={styles.hostInfo}>
                                    <div className={styles.systemLabel}>{systemVisual.name}</div>
                                    <h1 className={styles.hostName}>{system?.hostname || "未知主机"}</h1>
                                </div>
                            </Col>
                        </Row>
                    </Col>
                    <Col xs={24} lg={14}>
                        <div className={styles.hostSummary}>
                            <Row gutter={[0, gutter]} align="middle">
                                <Col span={8}>
                                    <div className="host-summary-item">
                                        <span className="summary-label"><Icon type="ClockCircleOutlined"/>连续运行</span>
                                        <strong className="summary-value">{formatUptime(system?.uptime)}</strong>
                                        <small className="summary-meta">
                                            启动于 {system?.boot_time ? dayjs.unix(numberValue(system.boot_time)).format("YYYY-MM-DD HH:mm") : "--"}
                                        </small>
                                    </div>
                                </Col>
                                <Col span={8}>
                                    <div className="host-summary-item">
                                        <span className="summary-label"><Icon type="DesktopOutlined"/>操作系统</span>
                                        <Tooltip
                                            title={operatingSystemEllipsis ? operatingSystem : undefined}
                                            mouseLeaveDelay={0.3}
                                            styles={{
                                                root: {pointerEvents: "auto"},
                                                container: {cursor: "text", userSelect: "text"},
                                            }}
                                        >
                                            <Typography.Text
                                                className="summary-value"
                                                ellipsis={{onEllipsis: setOperatingSystemEllipsis}}
                                            >
                                                {operatingSystem}
                                            </Typography.Text>
                                        </Tooltip>
                                        <small className="summary-meta">内核 {system?.kernel_version || "--"}</small>
                                    </div>
                                </Col>
                                <Col span={8}>
                                    <div className="host-summary-item">
                                        <span className="summary-label"><Icon type="ClusterOutlined"/>主机架构</span>
                                        <strong className="summary-value">{system?.kernel_arch || "--"}</strong>
                                        <small className="summary-meta">{numberValue(cpu?.processor)} 个逻辑处理器 · {formatSize(memory?.total)} 内存</small>
                                    </div>
                                </Col>
                            </Row>
                        </div>
                    </Col>
                </Row>
                <Button
                    className={styles.refreshButton}
                    type="text"
                    size="small"
                    loading={refreshing}
                    icon={<Icon type="ReloadOutlined"/>}
                    onClick={() => void onRefresh()}
                />
            </section>
        </Card>
    );
});

export default HostOverview;
