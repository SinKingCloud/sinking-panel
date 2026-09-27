import React, {useEffect, useMemo, useRef, useState} from "react";
import {App, Form, InputNumber, Tooltip, Typography} from "antd";
import {Body, PageTable, ProModal, Title} from "sinking-antd";
import type {PageTableProps, PageTableRef, ProModalRef} from "sinking-antd";
import {getLog} from "@/service/api/system";
import useEnum from "@/utils/enum";
import {getParams} from "@/utils/page";
import useStyles from "./styles";

const emptyEnum = {};

export default (): React.ReactNode => {
    const {styles} = useStyles();
    const {message} = App.useApp();
    const [enumData, enumLoading]: any = useEnum("log");
    const tableRef = useRef<PageTableRef<any> | null>(null);
    const clearRef = useRef<ProModalRef | null>(null);
    const requestRef = useRef(0);
    const submittingRef = useRef(false);
    const [form] = Form.useForm<{ day: number }>();
    const [submitting, setSubmitting] = useState(false);
    const typeData: Record<string, string> = enumData?.type || emptyEnum;

    useEffect(() => () => {
        requestRef.current += 1;
    }, []);

    const columns = useMemo<PageTableProps<any>["columns"]>(() => [
        {
            title: "操作IP",
            dataIndex: "ip",
            key: "ip",
            width: 150,
            render: (value: string) => (
                <Typography.Text className="log-copy" copyable={value ? {text: value} : false}>
                    {value || "-"}
                </Typography.Text>
            ),
        },
        {
            title: "IP归属地",
            dataIndex: "location",
            key: "location",
            width: 220,
            render: (value: string) => (
                <Typography.Text className="log-copy" copyable={value ? {text: value} : false}>
                    {value || "-"}
                </Typography.Text>
            ),
        },
        {
            title: "操作类型",
            dataIndex: "type",
            key: "type",
            width: 100,
            render: (value: number) => (
                <span className={`log-type is-${value}`}>
                    <i/>
                    {typeData[String(value)] || "未知类型"}
                </span>
            ),
        },
        {
            title: "操作标题",
            dataIndex: "title",
            key: "title",
            width: 150,
            render: (value: string) => (
                <Tooltip title={value || "-"}><span className="log-text">{value || "-"}</span></Tooltip>
            ),
        },
        {
            title: "操作内容",
            dataIndex: "content",
            key: "content",
            width: 260,
            render: (value: string) => (
                <Tooltip title={value || "-"}><span className="log-text">{value || "-"}</span></Tooltip>
            ),
        },
        {
            title: "操作时间",
            dataIndex: "create_time",
            key: "create_time",
            width: 170,
            sorter: true,
            render: (value: string) => (
                <Tooltip title={value || "-"}><span className="log-time">{value || "-"}</span></Tooltip>
            ),
        },
    ], [typeData]);

    const submit = async (values: { day: number }) => {
        if (submittingRef.current) return;
        const requestId = ++requestRef.current;
        submittingRef.current = true;
        setSubmitting(true);
        try {
            const response = await getLog({body: {action: "clear", day: String(values.day)}});
            if (requestRef.current !== requestId) return;
            if (response?.code !== 200) {
                message.error(response?.message || "清理失败");
                return;
            }
            clearRef.current?.hide();
            tableRef.current?.reload();
            message.success(response.message || "清理成功");
        } catch {
            if (requestRef.current === requestId) message.error("清理失败");
        } finally {
            if (requestRef.current === requestId) {
                submittingRef.current = false;
                setSubmitting(false);
            }
        }
    };

    return (
        <Body loading={enumLoading}>
            <PageTable<any>
                ref={tableRef}
                ariaLabel="操作日志"
                className={styles.logTable}
                columns={columns}
                rowKey="id"
                request={async (params, sort) => {
                    const response = await getLog({body: getParams(params, sort)});
                    if (response?.code !== 200) {
                        throw new Error(response?.message || "获取列表失败");
                    }
                    return {data: response.data?.list, total: response.data?.total ?? 0, success: true};
                }}
                defaultPageSize={20}
                defaultSort={{id: "descend"}}
                paginationAffix={{offsetBottom: 15}}
                hero={{
                    eyebrow: "OPERATION AUDIT",
                    title: "操作日志",
                    action: {
                        label: "清理日志",
                        onClick: () => {
                            requestRef.current += 1;
                            submittingRef.current = false;
                            setSubmitting(false);
                            form.resetFields();
                            form.setFieldsValue({day: 30});
                            clearRef.current?.show();
                        },
                    },
                }}
                search={{
                    placeholder: "搜索日志",
                    maxLength: 200,
                    fields: [
                        {value: "keyword", label: "全部", placeholder: "搜索 IP、归属地、标题或内容"},
                        {value: "title", label: "标题", placeholder: "搜索标题"},
                        {value: "content", label: "内容", placeholder: "搜索内容"},
                        {value: "location", label: "归属地", placeholder: "搜索归属地", maxLength: 100},
                        {value: "ip", label: "IP", placeholder: "搜索 IP"},
                    ],
                }}
                filters={[
                    {
                        type: "select",
                        name: "type",
                        label: "日志类型筛选",
                        icon: "TagsOutlined",
                        allLabel: "全部类型",
                        valueEnum: typeData
                    },
                    {type: "date", name: "create_time", label: "操作时间筛选"},
                ]}
                toolbar={{refresh: {ariaLabel: "刷新操作日志列表"}}}
            />
            <ProModal
                ref={clearRef}
                title={<Title>清理日志</Title>}
                width={320}
                onOk={form.submit}
                onCancel={() => {
                    requestRef.current += 1;
                    clearRef.current?.hide();
                }}
                modalProps={{
                    confirmLoading: submitting,
                    okText: "提交",
                    cancelText: "取消",
                    cancelButtonProps: {disabled: false},
                    closable: true,
                    keyboard: true,
                    forceRender: true,
                    style: {top: 100, paddingBottom: 100},
                    mask: {closable: true},
                    afterClose: () => {
                        requestRef.current += 1;
                        submittingRef.current = false;
                        form.resetFields();
                        setSubmitting(false);
                    },
                } as any}
            >
                <Form form={form} layout="vertical" onFinish={submit}>
                    <Form.Item
                        name="day"
                        label="保留天数"
                        rules={[{required: true, message: "请输入保留天数"}]}
                    >
                        <InputNumber min={1} precision={0} step={1} style={{width: "100%"}}
                                     placeholder="请输入保留天数"/>
                    </Form.Item>
                </Form>
            </ProModal>
        </Body>
    );
};
