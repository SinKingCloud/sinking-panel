import {forwardRef, memo, useCallback, useImperativeHandle, useMemo, useRef, useState} from "react";
import {App, Col, Form as AntForm, Input, Row, Select, Spin} from "antd";
import {createStyles} from "antd-style";
import {ProModal, ProModalRef, Title, useTheme} from "sinking-antd";
import defaultSettings from "@/../config/defaultSettings";
import AceEditor from "@/components/ace-editor";
import {getTypeList} from "@/service/api/type";
import {
    createScript,
    getScriptInfo,
    updateScript,
} from "@/service/api/script";

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

const ScriptForm = forwardRef<any, any>(({instanceModule, onSuccess}, ref): any => {
    const {message} = App.useApp();
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles();
    const modalRef = useRef<ProModalRef>({} as ProModalRef);
    const requestRef = useRef(0);
    const [form] = AntForm.useForm<any>();
    const [scriptId, setScriptId] = useState<number>();
    const [active, setActive] = useState(false);
    const [infoLoading, setInfoLoading] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const [scopeModule, setScopeModule] = useState<string>();
    const editing = scriptId !== undefined;
    const [categories, setCategories] = useState<Record<string, any[]>>({});
    const selectedModule = AntForm.useWatch("module", form) || "script_ssh";
    const typeOptions = useMemo(() => [{label: "全部分类", value: 0}, ...(categories[selectedModule] || [])
        .map((item) => ({label: item.name, value: Number(item.id)}))], [categories, selectedModule]);

    const loadCategories = useCallback(async () => {
        const response = await getTypeList({body: {module: instanceModule || "script_ssh", order_by_field: "sort", order_by_type: "asc"}});
        if (response?.code !== 200) {
            throw new Error(response?.message || "获取脚本分类失败");
        }
        const items = Array.isArray(response?.data?.list) ? response?.data?.list.filter(Boolean) : [];
        const modules = instanceModule ? ["script_ssh", instanceModule] : ["script_ssh"];
        return Object.fromEntries(modules.map((module) => [module,
            module === "script_ssh" ? items.filter((item: any) => item.module === "script_ssh") : items,
        ]));
    }, [instanceModule]);

    const reset = useCallback(() => {
        requestRef.current += 1;
        form.resetFields();
        setScriptId(undefined);
        setActive(false);
        setInfoLoading(false);
        setSubmitting(false);
    }, [form]);

    const openCreate = useCallback((defaultTypeId = 0) => {
        const requestId = ++requestRef.current;
        setScriptId(undefined);
        setActive(true);
        setInfoLoading(true);
        setScopeModule(instanceModule);
        form.resetFields();
        modalRef.current?.show();
        void loadCategories().then((data) => {
            if (requestRef.current !== requestId) return;
            const module = defaultTypeId && !data.script_ssh?.some((item: any) => Number(item.id) === defaultTypeId) && instanceModule
                ? instanceModule : "script_ssh";
            setCategories(data);
            form.setFieldsValue({
                name: "",
                type_id: data[module]?.some((item: any) => Number(item.id) === defaultTypeId) ? defaultTypeId : 0,
                module,
                script: "",
            });
        }).catch((error) => {
            if (requestRef.current !== requestId) return;
            modalRef.current?.hide();
            message.error(error?.message || "获取脚本分类失败");
        }).finally(() => {
            if (requestRef.current === requestId) setInfoLoading(false);
        });
    }, [form, instanceModule, loadCategories, message]);

    const openEdit = useCallback((record: any) => {
        const requestId = ++requestRef.current;
        setScriptId(record.id);
        setActive(true);
        setInfoLoading(true);
        form.resetFields();
        modalRef.current?.show();

        void getScriptInfo({body: {id: record.id}}).then((response) => {
            if (requestRef.current !== requestId) return;
            if (response?.code !== 200 || !response?.data) {
                throw new Error(response?.message || "获取脚本详情失败");
            }
            form.setFieldsValue({
                name: String(response?.data?.name || ""),
                script: String(response?.data?.script || ""),
            });
        }).catch((error) => {
            if (requestRef.current !== requestId) return;
            modalRef.current?.hide();
            message.error(error?.message || "获取脚本详情失败");
        }).finally(() => {
            if (requestRef.current === requestId) setInfoLoading(false);
        });
    }, [form, message]);

    useImperativeHandle(ref, () => ({
        open: (record?: any, defaultTypeId = 0) => {
            if (record) {
                openEdit(record);
            } else {
                openCreate(defaultTypeId);
            }
        },
    }), [openCreate, openEdit]);

    const close = useCallback(() => {
        requestRef.current += 1;
        modalRef.current?.hide();
    }, []);

    const submit = useCallback(async (values: any) => {
        const requestId = ++requestRef.current;
        const name = String(values.name || "").trim();
        const script = String(values.script || "");
        setSubmitting(true);
        try {
            const response = editing
                ? await updateScript({
                    body: {
                        ids: [scriptId as number],
                        name,
                        script,
                    }
                })
                : await createScript({
                    body: {
                        type_id: Number(values.type_id || 0),
                        module: values.module,
                        name,
                        script,
                    }
                });
            const current = requestRef.current === requestId;
            if (response?.code !== 200) {
                if (current && response) {
                    message.error(response?.message || (editing ? "修改脚本失败" : "添加脚本失败"));
                }
                return;
            }
            if (!current) {
                return;
            }
            modalRef.current?.hide();
            message.success(response?.message || (editing ? "修改成功" : "添加成功"));
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
            onCancel={close}
            width={compact ? 600 : 640}
            modalProps={{
                rootClassName: styles.modal,
                confirmLoading: submitting || infoLoading,
                okText: editing ? "保存" : "添加",
                cancelText: "取消",
                style: {top: "clamp(12px, 7vh, 72px)", paddingBottom: 24},
                mask: {closable: true},
                keyboard: true,
                closable: true,
                focusable: {focusTriggerAfterClose: false},
                afterClose: reset,
            } as any}>
            <AntForm<any> form={form} layout="vertical" variant="filled" onFinish={submit}>
                {active && (infoLoading ? (
                    <div className={styles.loading}>
                        <Spin/>
                    </div>
                ) : (
                    <>
                        <AntForm.Item
                            name="name"
                            label="脚本名称"
                            rules={[
                                {required: true, whitespace: true, message: "请输入脚本名称"},
                                {max: 100, message: "脚本名称不能超过 100 个字符"},
                            ]}>
                            <Input maxLength={100} placeholder="请输入脚本名称" autoComplete="off"/>
                        </AntForm.Item>
                        {!editing && <Row gutter={[16, 0]}>
                            <Col xs={24} sm={12}>
                                <AntForm.Item name="module" label="保存范围" rules={[{required: true, message: "请选择保存范围"}]}>
                                    <Select
                                        placeholder="请选择保存范围"
                                        options={[
                                            {label: "全局脚本", value: "script_ssh"},
                                            ...(scopeModule ? [{label: "实例脚本", value: scopeModule}] : []),
                                        ]}
                                        onChange={(module) => {
                                            if (!categories[module]?.some((item) => Number(item.id) === Number(form.getFieldValue("type_id")))) {
                                                form.setFieldValue("type_id", 0);
                                            }
                                        }}/>
                                </AntForm.Item>
                            </Col>
                            <Col xs={24} sm={12}>
                                <AntForm.Item name="type_id" label="脚本分类" rules={[{required: true, message: "请选择脚本分类"}]}>
                                    <Select options={typeOptions} placeholder="请选择脚本分类"/>
                                </AntForm.Item>
                            </Col>
                        </Row>}
                        <AntForm.Item
                            name="script"
                            label="脚本内容"
                            rules={[
                                {required: true, whitespace: true, message: "请输入脚本内容"},
                                {max: 131072, message: "脚本内容不能超过 131072 个字符"},
                            ]}>
                            <AceEditor
                                mode="sh"
                                theme={theme?.isDarkTheme?.() ? "monokai" : "chrome"}
                                width="100%"
                                height={compact
                                    ? "clamp(170px, 34dvh, 250px)"
                                    : "clamp(180px, 36dvh, 280px)"}
                                fontSize={compact ? 12 : 14}
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
