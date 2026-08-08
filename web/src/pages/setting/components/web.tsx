import React, {useMemo} from "react";
import {Form, Input} from "antd";
import useConfig from "../hooks/config";
import Actions from "./actions";
import Error from "./error";
import Loading from "./loading";

const fieldStyle = {width: "100%", maxWidth: 350};

export default ({styles, info}: any): React.ReactNode => {
    const defaults = useMemo(() => ({
        name: info?.name || "Sinking Panel",
        title: info?.title || "Sinking Panel",
    }), [info?.name, info?.title]);
    const config = useConfig("web", defaults);

    if (config.error && !config.loading) {
        return <Error styles={styles} message={config.error} onRetry={config.load}/>;
    }
    if (config.loading) {
        return <Loading styles={styles}/>;
    }

    return (
        <Form form={config.form} layout="vertical" onFinish={config.save}>
            <Form.Item
                name="name"
                label="网站名称"
                tooltip="网站名称"
                style={fieldStyle}
                rules={[
                    {required: true, whitespace: true, message: "请输入网站名称"},
                    {max: 50, message: "网站名称不能超过50个字符"},
                ]}>
                <Input placeholder="请输入网站名称" maxLength={50}/>
            </Form.Item>
            <Form.Item
                name="title"
                label="网站标题"
                tooltip="网站标题"
                style={fieldStyle}
                rules={[
                    {required: true, whitespace: true, message: "请输入网站标题"},
                    {max: 100, message: "网站标题不能超过100个字符"},
                ]}>
                <Input placeholder="请输入网站标题" maxLength={100}/>
            </Form.Item>
            <Actions styles={styles} saving={config.saving} onReset={config.reset}/>
        </Form>
    );
};
