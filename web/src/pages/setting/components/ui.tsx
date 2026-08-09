import React, {useMemo} from "react";
import {Button, ColorPicker, Flex, Form, Select} from "antd";
import defaultSettings from "@/../config/defaultSettings";
import useConfig from "../hooks/config";
import Actions from "./actions";
import Error from "./error";
import Loading from "./loading";

const toBoolean = (value: any) => value === true || value === 1 || value === "1" || value === "true";
const fieldStyle = {width: "100%", maxWidth: 350};
const switchOptions = [
    {value: "1", label: "开启"},
    {value: "0", label: "关闭"},
];
const layoutOptions = [
    {value: "top", label: "上下布局"},
    {value: "left", label: "左右布局"},
];
const themeOptions = [
    {value: "light", label: "亮色模式"},
    {value: "dark", label: "暗色模式"},
];
const radiusOptions = Array.from({length: 16}, (_, index) => ({
    value: String(index),
    label: `${index}px${index === 0 ? " (无圆角)" : ""}`,
}));

const normalize = (values: any) => ({
    layout: values.layout === "top" ? "top" : "left",
    watermark: toBoolean(values.watermark) ? "1" : "0",
    theme: values.theme === "light" ? "light" : "dark",
    compact: toBoolean(values.compact) ? "1" : "0",
    color: values.color === undefined || values.color === null ? defaultSettings.color : String(values.color),
    radius: String(Math.min(15, Math.max(0, Number(values.radius) || 0))),
});

export default ({styles, info}: any): React.ReactNode => {
    const ui = info?.ui;
    const defaults = useMemo(() => normalize({
        layout: ui?.layout,
        watermark: ui?.watermark,
        theme: ui?.theme,
        compact: ui?.compact,
        color: ui?.color,
        radius: ui?.radius,
    }), [ui?.color, ui?.compact, ui?.layout, ui?.radius, ui?.theme, ui?.watermark]);
    const config = useConfig("ui", defaults, normalize);

    return (
        <div className={styles.uiWrapper}>
            <Form form={config.form} layout="vertical" onFinish={config.save}>
                {config.loading ? <Loading styles={styles}/> : config.error ? (
                    <Error styles={styles} message={config.error} onRetry={config.load}/>
                ) : <>
                    <Form.Item name="compact" label="紧凑模式" tooltip="控制界面元素间距，开启后界面更紧凑"
                               style={fieldStyle} rules={[{required: true, message: "请选择紧凑模式"}]}>
                        <Select placeholder="请选择紧凑模式是否开启" options={switchOptions}/>
                    </Form.Item>
                    <Form.Item name="layout" label="界面布局" tooltip="选择界面的整体布局方式，上下布局或左右布局"
                               style={fieldStyle} rules={[{required: true, message: "请选择界面布局"}]}>
                        <Select placeholder="请选择界面布局方式" options={layoutOptions}/>
                    </Form.Item>
                    <Form.Item name="theme" label="菜单主题" tooltip="选择菜单的主题模式，亮色或暗色"
                               style={fieldStyle} rules={[{required: true, message: "请选择菜单主题"}]}>
                        <Select placeholder="请选择菜单主题模式" options={themeOptions}/>
                    </Form.Item>
                    <Form.Item name="watermark" label="界面水印" tooltip="控制界面是否显示账户水印"
                               style={fieldStyle} rules={[{required: true, message: "请选择界面水印"}]}>
                        <Select placeholder="请选择是否显示界面水印" options={switchOptions}/>
                    </Form.Item>
                    <Form.Item name="radius" label="主题圆角" tooltip="调整界面元素的圆角大小"
                               style={fieldStyle} rules={[{required: true, message: "请选择主题圆角"}]}>
                        <Select placeholder="请选择主题圆角大小" options={radiusOptions}/>
                    </Form.Item>
                    <Form.Item label="主题颜色" tooltip="选择界面的主题颜色" style={fieldStyle}>
                        <Flex gap={8} align="center">
                            <Form.Item name="color" noStyle
                                       getValueFromEvent={(_: any, value: string) => value}>
                                <ColorPicker showText format="rgb" defaultFormat="rgb"/>
                            </Form.Item>
                            <Form.Item shouldUpdate noStyle>
                                {({setFieldValue}) => (
                                    <Button onClick={() => setFieldValue("color", "")}>清空</Button>
                                )}
                            </Form.Item>
                        </Flex>
                    </Form.Item>
                    <Actions styles={styles} saving={config.saving} onReset={config.reset}/>
                </>}
            </Form>
        </div>
    );
};
