import React from "react";
import {Col, Row, Tooltip} from "antd";
import {createStyles} from "antd-style";
import {Icon} from "sinking-antd";
import {gutter, percentValue} from "../../utils";

const useStyles = createStyles(({css, token, isDarkMode}: any) => ({
    resource: css`
        --resource-color: ${token.colorPrimary};
        position: relative;
        min-width: 0;
        min-height: 138px;
        padding: 15px 16px;
        background:
            radial-gradient(circle at 108% 118%, color-mix(in srgb, var(--resource-color) ${isDarkMode ? 18 : 13}%, transparent) 0%, transparent 54%),
            radial-gradient(circle at 18% 45%, color-mix(in srgb, var(--resource-color) ${isDarkMode ? 7 : 5}%, transparent) 0%, transparent 58%),
            ${token.colorBgContainer};
        overflow: hidden;
        isolation: isolate;

        > .ant-row {
            position: relative;
            z-index: 1;
            min-height: 108px;
        }

        > .ant-row > .ant-col {
            min-width: 0;
        }

        &::before {
            content: "";
            position: absolute;
            z-index: 0;
            right: -62px;
            bottom: -82px;
            width: 176px;
            height: 176px;
            border-radius: 50%;
            background: color-mix(in srgb, var(--resource-color) ${isDarkMode ? 48 : 38}%, transparent);
            opacity: ${isDarkMode ? .28 : .22};
            pointer-events: none;
            filter: blur(30px) saturate(1.12);
            transform: translate3d(0, 0, 0) scale(1);
            transition: transform .34s cubic-bezier(.2, .8, .2, 1);
            will-change: transform;
        }

        &::after {
            content: "";
            position: absolute;
            z-index: 0;
            top: 12px;
            left: -38px;
            width: 118px;
            height: 118px;
            border-radius: 50%;
            background: color-mix(in srgb, var(--resource-color) ${isDarkMode ? 38 : 30}%, transparent);
            opacity: ${isDarkMode ? .08 : .055};
            pointer-events: none;
            filter: blur(25px) saturate(1.08);
            transform: translate3d(0, 0, 0) scale(.94);
            transition: transform .34s cubic-bezier(.2, .8, .2, 1);
            will-change: transform;
        }

        .resource-watermark {
            position: absolute;
            right: -25px;
            bottom: -29px;
            color: var(--resource-color);
            font-size: 138px;
            line-height: 1;
            opacity: ${isDarkMode ? .045 : .075};
            pointer-events: none;
            filter: blur(${isDarkMode ? 8 : 6}px);
            transform: translate3d(0, 0, 0) rotate(-24deg) scale(.98);
            transition: transform .34s cubic-bezier(.2, .8, .2, 1);
            will-change: transform;
            backface-visibility: hidden;
        }

        .resource-ring {
            position: relative;
            z-index: 1;
            width: 100px;
            height: 100px;
            flex: none;
            contain: layout paint;
            isolation: isolate;
            transform: translate3d(0, 0, 0);
            backface-visibility: hidden;
        }

        .resource-ring svg {
            width: 100%;
            height: 100%;
            display: block;
        }

        .ring-track,
        .ring-progress {
            fill: none;
            stroke-width: 8;
        }

        .ring-track {
            stroke: color-mix(in srgb, var(--resource-color) 8%, ${token.colorFillSecondary});
        }

        .ring-progress {
            stroke: var(--resource-color);
            stroke-linecap: round;
            transition: stroke-dashoffset .25s ease;
        }

        .ring-content {
            position: absolute;
            inset: 0;
            padding: 0 11px;
            display: flex;
            align-items: center;
            justify-content: center;
            box-sizing: border-box;
            color: ${token.colorTextHeading};
            font-size: 15px;
            font-weight: 700;
            line-height: 22px;
            letter-spacing: 0;
            text-align: center;
            white-space: nowrap;
        }

        .ring-content.long {
            font-size: 13px;
            line-height: 19px;
        }

        .resource-copy {
            position: relative;
            z-index: 1;
            min-width: 0;
            flex: 1;
        }

        .resource-title {
            min-width: 0;
            display: flex;
            align-items: center;
            flex-wrap: nowrap;
            gap: 7px;
            color: ${token.colorTextHeading};
            font-size: 13px;
            line-height: 20px;
        }

        .resource-title strong {
            min-width: 0;
            font-weight: 600;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .resource-icon {
            flex: none;
            color: var(--resource-color);
            font-size: 16px;
        }

        .resource-detail {
            margin-top: 5px;
            color: ${token.colorTextTertiary};
            font-size: 12px;
            line-height: 18px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .resource-meta {
            margin-top: 6px;
            color: ${token.colorTextQuaternary};
            font-size: 11px;
            line-height: 17px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        @media (hover: hover) {
            &:hover::before {
                transform: translate3d(-8px, -6px, 0) scale(1.1);
            }

            &:hover::after {
                transform: translate3d(7px, 3px, 0) scale(1.08);
            }

            &:hover .resource-watermark {
                transform: translate3d(-3px, -2px, 0) rotate(-21deg) scale(1.02);
            }
        }

        @media (prefers-reduced-motion: reduce) {
            &::before,
            &::after,
            .resource-watermark,
            .ring-progress {
                transition: none;
            }
        }

        @media (max-width: 576px) {
            min-height: 124px;

            > .ant-row {
                min-height: 94px;
            }

            .resource-ring {
                width: 92px;
                height: 92px;
            }
        }
    `,
}));

const ResourceMetric = React.memo(({
    color,
    icon,
    title,
    value,
    detail,
    meta,
    percent,
    detailTooltip = true,
}: any) => {
    const {styles} = useStyles();

    return (
        <div className={styles.resource} style={{"--resource-color": color} as React.CSSProperties}>
            <span className="resource-watermark"><Icon type={icon}/></span>
            <Row gutter={gutter} align="middle" wrap={false}>
                <Col flex="none">
                    <div className="resource-ring">
                        <svg viewBox="0 0 72 72" aria-hidden="true">
                            <circle className="ring-track" cx="36" cy="36" r="30"/>
                            <circle
                                className="ring-progress"
                                cx="36"
                                cy="36"
                                r="30"
                                pathLength="100"
                                strokeDasharray="100"
                                strokeDashoffset={100 - percentValue(percent)}
                                transform="rotate(-90 36 36)"
                            />
                        </svg>
                        <span className={`ring-content ${value.length > 5 ? "long" : ""}`}>{value}</span>
                    </div>
                </Col>
                <Col flex="auto">
                    <div className="resource-copy">
                        <div className="resource-title">
                            <span className="resource-icon"><Icon type={icon}/></span>
                            <strong>{title}</strong>
                        </div>
                        {detailTooltip ? (
                            <Tooltip title={detail}>
                                <div className="resource-detail">{detail}</div>
                            </Tooltip>
                        ) : <div className="resource-detail">{detail}</div>}
                        <div className="resource-meta">{meta}</div>
                    </div>
                </Col>
            </Row>
        </div>
    );
});

export default ResourceMetric;
