import React, {useEffect, useMemo, useState} from "react";
import {Button, ColorPicker, Flex, Form, Image, Input, Select, Spin, Tooltip} from "antd";
import {Icon} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import FileSelect from "@/pages/components/file-select";
import {getFileCategory, imageExtensions} from "@/pages/file/utils";
import {filePreviewUrl, getFileSign} from "@/service/api/file";
import useConfig from "../../hooks/config";
import Actions from "../actions";
import ConfigError from "../error";
import Loading from "../loading";

const toBoolean = (value: any) => value === true || value === 1 || value === "1" || value === "true";
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
    logo: values.logo || "",
    layout: values.layout === "top" ? "top" : "left",
    watermark: toBoolean(values.watermark) ? "1" : "0",
    theme: values.theme === "light" ? "light" : "dark",
    compact: toBoolean(values.compact) ? "1" : "0",
    color: values.color === undefined || values.color === null ? defaultSettings.color : String(values.color),
    radius: String(Math.min(15, Math.max(0, Number(values.radius) || 0))),
});

const ThemeColorField = ({id, value, onChange, className}: {
    id?: string; value?: string; onChange?: (color: string) => void; className: string;
}) => (
    <Flex gap={8} align="center">
        <ColorPicker value={value || defaultSettings.color} format="rgb" onChange={(_, color) => onChange?.(color)}>
            <span className={className}>
                <Input id={id} type="button" variant="filled" aria-label="选择主题颜色" aria-haspopup="dialog"
                       value={value || "默认颜色"} prefix={<span className="color-swatch" aria-hidden
                       style={{backgroundColor: value || defaultSettings.color}}/>}/>
            </span>
        </ColorPicker>
        <Tooltip title="清除主题颜色">
            <Button htmlType="button" color="default" variant="filled" aria-label="清除主题颜色"
                    icon={<Icon type="ClearOutlined"/>} onClick={() => onChange?.("")}/>
        </Tooltip>
    </Flex>
);

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
    const logoPath: string = Form.useWatch("logo", config.form) || "";
    const [preview, setPreview] = useState<{path: string; url: string}>();
    const [previewOpen, setPreviewOpen] = useState(false);
    const isImage = getFileCategory(logoPath) === "image";
    const previewUrl = isImage && preview?.path === logoPath ? preview.url : "";
    useEffect(() => {
        if (!isImage) return;
        let active = true;
        getFileSign({body: {path: logoPath}}).then(response => {
            if (active) setPreview({
                path: logoPath,
                url: response?.code === 200 && response?.data ? filePreviewUrl(response.data) : "",
            });
        }).catch(() => {
            if (active) setPreview({path: logoPath, url: ""});
        });
        return () => { active = false; };
    }, [logoPath, isImage]);

    return (
        <div className={styles.uiWrapper}>
            <Form form={config.form} layout="vertical" variant="filled" onFinish={config.save}>
                {config.loading ? <Loading styles={styles}/> : config.error ? (
                    <ConfigError styles={styles} message={config.error} onRetry={config.load}/>
                ) : <>
                    <Form.Item name="compact" label="紧凑模式" tooltip="控制界面元素间距，开启后界面更紧凑"
                               className={styles.formField} rules={[{required: true, message: "请选择紧凑模式"}]}>
                        <Select placeholder="请选择紧凑模式是否开启" options={switchOptions}/>
                    </Form.Item>
                    <Form.Item name="layout" label="界面布局" tooltip="选择界面的整体布局方式，上下布局或左右布局"
                               className={styles.formField} rules={[{required: true, message: "请选择界面布局"}]}>
                        <Select placeholder="请选择界面布局方式" options={layoutOptions}/>
                    </Form.Item>
                    <Form.Item name="theme" label="菜单主题" tooltip="选择菜单的主题模式，亮色或暗色"
                               className={styles.formField} rules={[{required: true, message: "请选择菜单主题"}]}>
                        <Select placeholder="请选择菜单主题模式" options={themeOptions}/>
                    </Form.Item>
                    <Form.Item name="watermark" label="界面水印" tooltip="控制界面是否显示账户水印"
                               className={styles.formField} rules={[{required: true, message: "请选择界面水印"}]}>
                        <Select placeholder="请选择是否显示界面水印" options={switchOptions}/>
                    </Form.Item>
                    <Form.Item name="radius" label="主题圆角" tooltip="调整界面元素的圆角大小"
                               className={styles.formField} rules={[{required: true, message: "请选择主题圆角"}]}>
                        <Select placeholder="请选择主题圆角大小" options={radiusOptions}/>
                    </Form.Item>
                    <Form.Item name="color" label="主题颜色" tooltip="选择界面的主题颜色" className={styles.formField}>
                        <ThemeColorField className={styles.colorTrigger}/>
                    </Form.Item>
                    <Form.Item name="logo" label="网站 Logo" className={styles.formField}
                               rules={[{validator: async (_, path) => {
                                   if (path && getFileCategory(path) !== "image") throw new Error("请选择图片文件");
                               }}]}
                               getValueProps={(path) => ({value: path ? [path] : []})}
                               getValueFromEvent={(paths: string[]) => paths[0] || ""}>
                        <FileSelect type="file" extensions={imageExtensions} renderTrigger={open => (
                            <Flex align="center" gap="small">
                                {logoPath && <div className={styles.logoTile} data-preview>
                                    {previewUrl ? <Image src={previewUrl} alt="网站 Logo" width="100%" height="100%"
                                        style={{display: "block", objectFit: "cover"}} preview={{open: previewOpen, onOpenChange: setPreviewOpen, cover: false}}
                                        onError={() => setPreview(current => current === preview ? {path: logoPath, url: ""} : current)}/> :
                                        isImage && preview?.path !== logoPath ? <Spin size="small"/> : "无法预览"}
                                    <div className="logo-actions">
                                        {previewUrl && <button type="button" aria-label="预览图片" title="预览图片" onClick={() => setPreviewOpen(true)}>
                                            <Icon type="EyeOutlined"/>
                                        </button>}
                                        <button type="button" aria-label="移除图片" title="移除图片" onClick={() => config.form.setFieldValue("logo", "")}>
                                            <Icon type="DeleteOutlined"/>
                                        </button>
                                    </div>
                                </div>}
                                <Button type="dashed" className={styles.logoTile} aria-label={logoPath ? "更换 Logo 图片" : "选择 Logo 图片"} onClick={open}>
                                    <Icon type="PlusOutlined"/>{logoPath ? "更换" : "选择"}
                                </Button>
                            </Flex>
                        )}/>
                    </Form.Item>
                    <Actions styles={styles} saving={config.saving} onReset={config.reset}/>
                </>}
            </Form>
        </div>
    );
};
