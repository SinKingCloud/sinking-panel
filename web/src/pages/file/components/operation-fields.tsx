import React, {memo, useRef} from "react";
import {Col, Form as AntForm, Input, Row, Select} from "antd";
import FilePicker from "@/pages/components/file-picker";
import {
    getRemoteFileName,
    validateFileName,
    validatePath,
    validateUrl,
} from "./operation-form.utils";

export interface FileOperationFormValues {
    dir?: string;
    name?: string;
    format?: "zip" | "tar" | "gz" | "tgz";
    url?: string;
    path?: string;
}

const formatOptions = [
    {label: "ZIP", value: "zip"},
    {label: "TAR", value: "tar"},
    {label: "GZIP", value: "gz"},
    {label: "TAR.GZ", value: "tgz"},
];

export const CompressOperationFields = memo(() => (
    <>
        <AntForm.Item
            name="dir"
            label="保存目录"
            rules={[
                {required: true, whitespace: true, message: "请输入保存目录"},
                {validator: validatePath},
            ]}>
            <Input placeholder="请输入压缩文件的保存目录"/>
        </AntForm.Item>
        <Row gutter={12}>
            <Col span={15}>
                <AntForm.Item
                    name="name"
                    label="压缩包名称"
                    rules={[
                        {required: true, whitespace: true, message: "请输入压缩包名称"},
                        {validator: validateFileName},
                    ]}>
                    <Input autoFocus maxLength={255} placeholder="无需输入扩展名"/>
                </AntForm.Item>
            </Col>
            <Col span={9}>
                <AntForm.Item
                    name="format"
                    label="压缩格式"
                    rules={[{required: true, message: "请选择压缩格式"}]}>
                    <Select options={formatOptions}/>
                </AntForm.Item>
            </Col>
        </Row>
    </>
));

CompressOperationFields.displayName = "CompressOperationFields";

export const ExtractOperationFields = memo(() => (
    <AntForm.Item
        name="dir"
        rules={[
            {required: true, whitespace: true, message: "请输入解压目标目录"},
            {validator: validatePath},
        ]}>
        <Input autoFocus aria-label="解压目标目录" placeholder="请输入解压目标目录"/>
    </AntForm.Item>
));

ExtractOperationFields.displayName = "ExtractOperationFields";

export const RemoteDownloadOperationFields = memo(() => {
    const form = AntForm.useFormInstance<FileOperationFormValues>();
    const nameEditedRef = useRef(false);

    return (
        <>
            <AntForm.Item
                name="url"
                label="文件地址"
                rules={[
                    {required: true, whitespace: true, message: "请输入文件地址"},
                    {validator: validateUrl},
                ]}>
                <Input
                    autoFocus
                    placeholder="https://example.com/file.zip"
                    onChange={(event) => {
                        if (!nameEditedRef.current) {
                            form.setFieldValue("name", getRemoteFileName(event.target.value));
                        }
                    }}/>
            </AntForm.Item>
            <AntForm.Item
                name="path"
                label="保存目录"
                rules={[
                    {required: true, whitespace: true, message: "请输入保存目录"},
                    {validator: validatePath},
                ]}>
                <FilePicker mode="directory" placeholder="请输入或选择文件保存目录"/>
            </AntForm.Item>
            <AntForm.Item
                name="name"
                label="保存名称"
                rules={[{validator: validateFileName}]}>
                <Input
                    maxLength={255}
                    placeholder="自动提取文件名称"
                    onChange={() => {
                        nameEditedRef.current = true;
                    }}/>
            </AntForm.Item>
        </>
    );
});

RemoteDownloadOperationFields.displayName = "RemoteDownloadOperationFields";
