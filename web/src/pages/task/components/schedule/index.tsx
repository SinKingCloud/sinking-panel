import {forwardRef, memo, useEffect, useMemo, useState} from "react";
import {Input, InputNumber, Select, TimePicker, Typography} from "antd";
import {createStyles} from "antd-style";
import {Icon, useTheme} from "sinking-antd";
import dayjs from "dayjs";
import {describeSchedule, parseSchedule} from "../../utils";

const scheduleOptions = [
    {label: "每秒", value: "second"},
    {label: "每分", value: "minute"},
    {label: "每小时", value: "hour"},
    {label: "每天", value: "day"},
    {label: "每周", value: "week"},
    {label: "每月", value: "month"},
    {label: "自定义", value: "custom"},
];

const weekdayOptions = [
    {label: "周一", value: 1},
    {label: "周二", value: 2},
    {label: "周三", value: 3},
    {label: "周四", value: 4},
    {label: "周五", value: 5},
    {label: "周六", value: 6},
    {label: "周日", value: 0},
];

const defaultExpressions: any = {
    second: "*/5 * * * * *",
    minute: "0 */5 * * * *",
    hour: "0 0 */1 * * *",
    day: "0 0 2 * * *",
    week: "0 0 2 * * 1",
    month: "0 0 2 1 * *",
};

const useStyles = createStyles(({css, token}: any, props: any = {}) => {
    const compact = Boolean(props?.isCompactMode);

    return {
    schedule: css`
        min-width: 0;
        container: task-schedule / inline-size;
    `,
    editor: css`
        --schedule-number-width: 96px;
        --schedule-minute-width: 88px;
        --schedule-time-width: 120px;
        --schedule-weekday-width: 104px;
        min-width: 0;
        display: grid;
        grid-template-columns: 132px minmax(0, 1fr);
        align-items: center;
        gap: 8px;

        .cycle-type.ant-select {
            width: 100%;
        }

        .cycle-value {
            min-width: 0;
        }

        .schedule-number {
            width: var(--schedule-number-width);
            flex: none;
        }

        .schedule-minute {
            width: var(--schedule-minute-width);
            flex: none;
        }

        .schedule-time {
            width: var(--schedule-time-width);
            flex: none;
        }

        .schedule-weekday.ant-select {
            width: var(--schedule-weekday-width);
            flex: none;
        }

        @container task-schedule (max-width: 480px) {
            --schedule-number-width: 64px;
            --schedule-minute-width: 64px;
            --schedule-time-width: 112px;
            --schedule-weekday-width: 86px;
            grid-template-columns: 108px minmax(0, 1fr);
        }

        @container task-schedule (max-width: 360px) {
            grid-template-columns: minmax(0, 1fr);
            align-items: stretch;
        }
    `,
    inlineControl: css`
        min-width: 0;
        min-height: ${token.controlHeight}px;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 8px;
        white-space: nowrap;

        > span {
            flex: none;
        }

        @container task-schedule (max-width: 480px) {
            gap: 6px;
        }
    `,
    preview: css`
        min-width: 0;
        min-height: ${compact ? 32 : 38}px;
        margin-top: ${compact ? 6 : 8}px;
        padding: ${compact ? "6px 9px" : "8px 11px"};
        box-sizing: border-box;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: ${compact ? 10 : 12}px;
        border-radius: ${token.borderRadius}px;
        background: ${token.colorFillQuaternary};

        .schedule-description.ant-typography {
            min-width: 0;
            flex: 1 1 auto;
            margin: 0;
            display: inline-flex;
            align-items: center;
            gap: ${compact ? 6 : 7}px;
            overflow: hidden;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
            line-height: 20px;
            white-space: nowrap;
            text-overflow: ellipsis;
        }

        .schedule-description .anticon {
            flex: none;
            color: ${token.colorTextTertiary};
            font-size: ${compact ? 11 : 12}px;
        }

        .schedule-expression.ant-typography {
            min-width: 0;
            flex: 0 1 auto;
            margin: 0;
            padding-inline-start: ${compact ? 10 : 12}px;
            display: inline-flex;
            align-items: center;
            gap: ${compact ? 5 : 6}px;
            overflow: hidden;
            border-inline-start: 1px solid ${token.colorSplit};
            color: ${token.colorTextTertiary};
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: ${compact ? 11 : 12}px;
            line-height: 20px;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .schedule-expression .expression-value {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .schedule-expression .ant-typography-copy {
            flex: none;
            margin-inline-start: 0;
            color: ${token.colorTextTertiary};
            font-family: ${token.fontFamily};
            transition: color .16s ease;
        }

        .schedule-expression .ant-typography-copy:hover {
            color: ${token.colorPrimary};
        }

        @container task-schedule (max-width: 360px) {
            flex-direction: column;
            align-items: stretch;
            gap: 2px;

            .schedule-expression.ant-typography {
                padding-inline-start: 0;
                border-inline-start: 0;
            }
        }

    `,
    };
});

const Schedule = forwardRef<HTMLDivElement, any>(({value, onChange}, ref):any => {
    const theme = useTheme();
    const isCompactMode = theme?.isCompactTheme?.() || false;
    const {styles} = useStyles({isCompactMode});
    const [mode, setMode] = useState(() => parseSchedule(value).mode);
    const schedule = useMemo(() => parseSchedule(value), [value]);
    const description = useMemo(() => describeSchedule(schedule), [schedule]);
    const timeValue = useMemo(() => dayjs()
        .startOf("day")
        .hour(schedule.hour || 0)
        .minute(schedule.minute || 0), [schedule.hour, schedule.minute]);

    useEffect(() => {
        setMode(schedule.mode);
    }, [schedule.mode]);

    const emit = (expression: string) => onChange?.(expression);

    const changeMode = (nextMode: string) => {
        setMode(nextMode);
        if (defaultExpressions[nextMode]) {
            emit(defaultExpressions[nextMode]);
        }
    };

    const changeTime = (time: any, targetMode: "day" | "week" | "month") => {
        if (!time) {
            return;
        }
        const minute = time.minute();
        const hour = time.hour();
        if (targetMode === "day") {
            emit(`0 ${minute} ${hour} * * *`);
        } else if (targetMode === "week") {
            emit(`0 ${minute} ${hour} * * ${schedule.weekday ?? 1}`);
        } else {
            emit(`0 ${minute} ${hour} ${schedule.day || 1} * *`);
        }
    };

    return (
        <div ref={ref} className={styles.schedule}>
            <div className={styles.editor}>
                <Select
                    className="cycle-type"
                    value={mode}
                    options={scheduleOptions}
                    onChange={(nextMode) => changeMode(String(nextMode))}
                />
                <div className="cycle-value">
                    {mode === "second" && (
                        <div className={styles.inlineControl}>
                            <span>每</span>
                            <InputNumber
                                className="schedule-number"
                                min={1}
                                max={59}
                                value={schedule.interval || 5}
                                onChange={(interval) => emit(`*/${interval || 1} * * * * *`)}
                            />
                            <span>秒</span>
                        </div>
                    )}
                    {mode === "minute" && (
                        <div className={styles.inlineControl}>
                            <span>每</span>
                            <InputNumber
                                className="schedule-number"
                                min={1}
                                max={59}
                                value={schedule.interval || 5}
                                onChange={(interval) => emit(`0 */${interval || 1} * * * *`)}
                            />
                            <span>分钟</span>
                        </div>
                    )}
                    {mode === "hour" && (
                        <div className={styles.inlineControl}>
                            <span>每</span>
                            <InputNumber
                                className="schedule-number"
                                min={1}
                                max={23}
                                value={schedule.interval || 1}
                                onChange={(interval) => emit(`0 ${schedule.minute || 0} */${interval || 1} * * *`)}
                            />
                            <span>小时</span>
                            <InputNumber
                                className="schedule-minute"
                                min={0}
                                max={59}
                                value={schedule.minute || 0}
                                onChange={(minute) => emit(`0 ${minute || 0} */${schedule.interval || 1} * * *`)}
                            />
                            <span>分钟</span>
                        </div>
                    )}
                    {mode === "day" && (
                        <div className={styles.inlineControl}>
                            <span>每天</span>
                            <TimePicker
                                className="schedule-time"
                                value={timeValue}
                                format="HH:mm"
                                minuteStep={1}
                                allowClear={false}
                                onChange={(time) => changeTime(time, "day")}
                            />
                        </div>
                    )}
                    {mode === "week" && (
                        <div className={styles.inlineControl}>
                            <span>每周</span>
                            <Select
                                className="schedule-weekday"
                                value={schedule.weekday ?? 1}
                                options={weekdayOptions}
                                onChange={(weekday) => emit(`0 ${schedule.minute || 0} ${schedule.hour || 0} * * ${weekday}`)}
                            />
                            <TimePicker
                                className="schedule-time"
                                value={timeValue}
                                format="HH:mm"
                                minuteStep={1}
                                allowClear={false}
                                onChange={(time) => changeTime(time, "week")}
                            />
                        </div>
                    )}
                    {mode === "month" && (
                        <div className={styles.inlineControl}>
                            <span>每月</span>
                            <InputNumber
                                className="schedule-number"
                                min={1}
                                max={31}
                                value={schedule.day || 1}
                                onChange={(day) => emit(`0 ${schedule.minute || 0} ${schedule.hour || 0} ${day || 1} * *`)}
                            />
                            <span>日</span>
                            <TimePicker
                                className="schedule-time"
                                value={timeValue}
                                format="HH:mm"
                                minuteStep={1}
                                allowClear={false}
                                onChange={(time) => changeTime(time, "month")}
                            />
                        </div>
                    )}
                    {mode === "custom" && (
                        <Input
                            value={String(value || "")}
                            placeholder="秒 分 时 日 月 星期"
                            onChange={(event) => emit(event.target.value)}
                        />
                    )}
                </div>
            </div>
            <div className={styles.preview}>
                <Typography.Text type="secondary" className="schedule-description">
                    <Icon type="ClockCircleOutlined"/>
                    {description}
                </Typography.Text>
                <Typography.Text type="secondary" className="schedule-expression" copyable={{text: String(value || "")}}>
                    <span className="expression-value">{String(value || "-")}</span>
                </Typography.Text>
            </div>
        </div>
    );
});

Schedule.displayName = "Schedule";

export default memo(Schedule);
