import {forwardRef, memo, useCallback, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Col, Form as AntForm, Input, Row, Select, Spin} from "antd";
import {createStyles} from "antd-style";
import {ProModal, ProModalRef, Title, useTheme} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import {
    createScript,
    getScriptInfo,
    updateScript,
} from "@/service/api/script";
import type {ScriptRecord} from "@/service/api/script";
import type {TypeRecord} from "@/service/api/type";

const acePath = `${defaultSettings?.basePath || "/"}ace`;

const useStyles = createStyles(({css, token}: any) => ({
    modal: css`
        .ant-modal {
            max-width: calc(100vw - 24px);
            outline: none;
        }

        .ant-modal-body {
            max-height: calc(100dvh - 168px);
            overflow-y: auto;
            overscroll-behavior: contain;
        }

        @supports not (height: 100dvh) {
            .ant-modal-body {
                max-height: calc(100vh - 168px);
            }
        }
    `,
    editor: css`
        overflow: hidden;
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;

        .ace_editor {
            border-radius: ${token.borderRadius}px !important;
        }
    `,
    loading: css`
        height: min(320px, 46dvh);
        display: flex;
        align-items: center;
        justify-content: center;
    `,
}));

interface ScriptFormValues {
    name: string;
    type_id: number;
    script: string;
}

export interface ScriptFormRef {
    open: (record?: ScriptRecord, defaultTypeId?: number) => void;
}

interface ScriptFormProps {
    typeItems: TypeRecord[];
    onSuccess?: () => void;
}

const ScriptForm = forwardRef<ScriptFormRef, ScriptFormProps>(({typeItems, onSuccess}, ref) => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles();
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const requestRef = useRef(0);
    const [form] = AntForm.useForm<ScriptFormValues>();
    const [scriptId, setScriptId] = useState<number>();
    const [active, setActive] = useState(false);
    const [infoLoading, setInfoLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const editing = scriptId !== undefined;
    const typeOptions = useMemo(() => [
        {label: "全部分类", value: 0},
        ...typeItems.map((item) => ({label: item.name, value: Number(item.id)})),
    ], [typeItems]);

    const reset = useCallback(() => {
        requestRef.current += 1;
        form.resetFields();
        setScriptId(undefined);
        setActive(false);
        setInfoLoading(false);
        setSubmitting(false);
    }, [form]);

    const openCreate = useCallback((defaultTypeId = 0) => {
        requestRef.current += 1;
        setScriptId(undefined);
        setActive(true);
        setInfoLoading(false);
        form.resetFields();
        form.setFieldsValue({
            name: "",
            type_id: defaultTypeId,
            script: "",
        });
        modalRef.current?.show();
    }, [form]);

    const openEdit = useCallback((record: ScriptRecord) => {
        const requestId = ++requestRef.current;
        setScriptId(record.id);
        setActive(true);
        setInfoLoading(true);
        form.resetFields();
        modalRef.current?.show();

        void getScriptInfo({body: {id: record.id}}).then((response) => {
            if (requestRef.current !== requestId) {
                return;
            }
            if (response?.code !== 200 || !response.data) {
                modalRef.current?.hide();
                message.error(response?.message || "获取脚本详情失败");
                return;
            }
            form.setFieldsValue({
                name: String(response.data.name || ""),
                type_id: Number(response.data.type_id || 0),
                script: String(response.data.script || ""),
            });
        }).finally(() => {
            if (requestRef.current === requestId) {
                setInfoLoading(false);
            }
        });
    }, [form, message]);

    useImperativeHandle(ref, () => ({
        open: (record?: ScriptRecord, defaultTypeId = 0) => {
            if (record) {
                openEdit(record);
            } else {
                openCreate(defaultTypeId);
            }
        },
    }), [openCreate, openEdit]);

    const submit = useCallback(async (values: ScriptFormValues) => {
        const requestId = ++requestRef.current;
        const name = String(values.name || "").trim();
        const script = String(values.script || "");
        const typeId = Number(values.type_id || 0);
        setSubmitting(true);
        try {
            const response = editing
                ? await updateScript({body: {
                    ids: [scriptId as number],
                    type_id: String(typeId),
                    name,
                    script,
                }})
                : await createScript({body: {
                    type_id: typeId,
                    name,
                    script,
                }});
            const current = requestRef.current === requestId;
            if (response?.code !== 200) {
                if (current && response) {
                    message.error(response.message || (editing ? "修改脚本失败" : "添加脚本失败"));
                }
                return;
            }
            if (current) {
                modalRef.current?.hide();
            }
            message.success(response.message || (editing ? "修改成功" : "添加成功"));
            onSuccess?.();
        } finally {
            if (requestRef.current === requestId) {
                setSubmitting(false);
            }
        }
    }, [editing, message, onSuccess, scriptId]);

    return (
        <ProModal
            ref={modalRef}
            title={<Title>{editing ? "编辑脚本" : "添加脚本"}</Title>}
            onOk={form.submit}
            width={compact ? 600 : 640}
            modalProps={{
                rootClassName: styles.modal,
                confirmLoading: submitting || infoLoading,
                forceRender: true,
                okText: editing ? "保存" : "添加",
                cancelText: "取消",
                style: {top: "clamp(12px, 7vh, 72px)", paddingBottom: 24},
                mask: {closable: !submitting},
                keyboard: !submitting,
                closable: !submitting,
                focusable: {focusTriggerAfterClose: false},
                afterClose: reset,
            } as any}>
            <AntForm<ScriptFormValues> form={form} layout="vertical" onFinish={submit}>
                {active && (infoLoading ? (
                    <div className={styles.loading}>
                        <Spin description="加载脚本详情..."/>
                    </div>
                ) : (
                    <>
                        <Row gutter={[16, 0]}>
                            <Col xs={24} sm={14}>
                                <AntForm.Item
                                    name="name"
                                    label="脚本名称"
                                    rules={[
                                        {required: true, whitespace: true, message: "请输入脚本名称"},
                                        {max: 100, message: "脚本名称不能超过 100 个字符"},
                                    ]}>
                                    <Input maxLength={100} placeholder="请输入脚本名称" autoComplete="off"/>
                                </AntForm.Item>
                            </Col>
                            <Col xs={24} sm={10}>
                                <AntForm.Item name="type_id" label="脚本分类">
                                    <Select
                                        showSearch
                                        optionFilterProp="label"
                                        options={typeOptions}/>
                                </AntForm.Item>
                            </Col>
                        </Row>
                        <AntForm.Item
                            name="script"
                            label="脚本内容"
                            rules={[
                                {required: true, whitespace: true, message: "请输入脚本内容"},
                                {max: 131072, message: "脚本内容不能超过 131072 个字符"},
                            ]}>
                            <AceEditor
                                mode="sh"
                                width="100%"
                                height={compact
                                    ? "clamp(170px, 34dvh, 250px)"
                                    : "clamp(180px, 36dvh, 280px)"}
                                fontSize={12}
                                showPrintMargin={false}
                                wrapEnabled
                                acePath={acePath}
                                onLoad={(editor) => editor.textInput?.getElement?.()
                                    ?.setAttribute("aria-label", "脚本内容")}
                                className={styles.editor}/>
                        </AntForm.Item>
                    </>
                ))}
            </AntForm>
        </ProModal>
    );
});

ScriptForm.displayName = "ScriptForm";

export default memo(ScriptForm);
