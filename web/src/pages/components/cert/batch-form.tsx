import {memo} from "react";
import {Button, Collapse, Form, Select, Spin, Tooltip} from "antd";
import type {FormInstance} from "antd";
import {createStyles} from "antd-style";
import {Icon, useTheme} from "sinking-antd";

export interface BatchFormProps {
    index: number;
    domain: string;
    renew?: boolean;
    form: FormInstance;
    disabled: boolean;
    expanded: boolean;
    onToggle: (index: number) => void;
    secretOptions: {value: number; label: string}[];
    secretLoading: boolean;
    secretError: boolean;
    onReloadSecrets: () => void;
    onManageSecrets: (index: number, create?: boolean) => void;
    error?: string;
}

const useStyles = createStyles(({css, token}: any, props: {compact: boolean}) => ({
    item: css`
        min-width: 0;
        scroll-margin-block: 8px;
        & + & { margin-top: 4px; }
        .ant-form-item { min-width: 0; margin-bottom: 0; }
        .ant-form-item-control-input { min-height: 0; }
        .ant-select { width: 100%; }
        .ant-form-item-explain { font-size: 12px; overflow-wrap: anywhere; }
    `,
    summary: css`
        &.ant-collapse-header {
            gap: 10px;
            min-height: ${props.compact ? 48 : 52}px;
            padding: 8px 10px;
            border-radius: ${token.borderRadius}px;
            transition: background ${token.motionDurationMid};
            &:hover, &[aria-expanded="true"] { background: ${token.colorFillQuaternary}; }
            &:focus-visible { outline: 2px solid ${token.colorPrimaryBorder}; outline-offset: -2px; }
        }
        .batch-title { display: flex; align-items: center; gap: 8px; }
        .batch-domain { min-width: 0; color: ${token.colorText}; font-size: ${props.compact ? 12 : 13}px; font-weight: 500; line-height: 20px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .batch-kind { flex: none; color: ${token.colorTextTertiary}; font-size: 12px; }
        .ant-collapse-arrow { color: ${token.colorTextTertiary}; }
        .batch-meta { display: block; color: ${token.colorTextTertiary}; font-size: 12px; line-height: 18px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .batch-missing { color: ${token.colorWarningText}; font-size: 12px; white-space: nowrap; }
    `,
    fields: css`
        display: grid;
        gap: ${props.compact ? 8 : 10}px;
        min-width: 0;
        padding: ${props.compact ? 10 : 12}px 10px;
        .ant-form-item-row { flex-wrap: nowrap; }
        .ant-form-item-label { flex: 0 0 ${props.compact ? 72 : 80}px; padding: 0; text-align: left; }
        .ant-form-item-label > label { height: ${props.compact ? 24 : 28}px; color: ${token.colorTextSecondary}; font-size: ${props.compact ? 12 : 13}px; font-weight: 400; }
        .ant-form-item-control { flex: 1; min-width: 0; }
    `,
    secret: css`
        display: flex;
        align-items: flex-start;
        gap: 6px;
        min-width: 0;
        > .ant-form-item { flex: 1; }
        > .ant-btn { flex: none; }
    `,
    error: css`
        padding: 4px 10px;
        color: ${token.colorError};
        font-size: 12px;
        line-height: 1.6;
        overflow-wrap: anywhere;
    `,
}));

const BatchForm = ({index, domain, renew = false, form, disabled, expanded, onToggle, secretOptions, secretLoading, secretError, onReloadSecrets, onManageSecrets, error}: BatchFormProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const values = Form.useWatch(["items", index], {form, preserve: true}) || form.getFieldValue(["items", index]) || {};
    const wildcard = domain.startsWith("*.");
    const address = domain.replace(/^\[([^\]]+)\]$/, "$1");
    const ipAddress = address.includes(":") || (/^(?:\d{1,3}\.){3}\d{1,3}$/.test(address) && address.split(".").every((part) => Number(part) <= 255));
    const manual = Number(values.type ?? 2) === 1;
    const automaticDNS = !manual && (wildcard || (!ipAddress && values.challenge === "dns"));
    const noSecrets = !secretLoading && !secretError && secretOptions.length === 0;
    const secretName = secretOptions.find((option) => option.value === Number(values.secret_id))?.label;
    const summary = [
        wildcard ? "DNS" : ipAddress ? "HTTP" : String(values.challenge || "http").toUpperCase(),
        manual ? "手动验证" : "自动验证",
        renew && !values.ca ? "沿用原环境" : values.ca === "staging" ? "测试环境" : "正式环境",
        ...(!manual ? [Number(values.auto_renew) === 1 ? "自动续签" : "不自动续签"] : []),
        ...(automaticDNS && secretName ? [secretName] : []),
    ].join(" · ");

    return <section className={styles.item} role="listitem" aria-label={`${domain} 的${renew ? "续签" : "申请"}设置`}>
        <Collapse ghost size="small" expandIconPlacement="end" activeKey={expanded ? ["settings"] : []} onChange={() => onToggle(index)}
            collapsible={disabled ? "disabled" : undefined} destroyOnHidden={false}
            classNames={{header: styles.summary}}
            styles={{header: {alignItems: "center"}, title: {minWidth: 0}, body: {padding: 0}, icon: {paddingInline: 0}}}
            items={[{key: "settings", forceRender: true,
                onKeyDown: (event) => {
                    if (!disabled && event.key === " " && event.target === event.currentTarget.querySelector(".ant-collapse-header")) {
                        event.preventDefault();
                        onToggle(index);
                    }
                },
                label: <>
                    <span className="batch-title"><span className="batch-domain" title={domain}>{domain}</span>{(wildcard || ipAddress) && <span className="batch-kind">{wildcard ? "泛域名" : "IP"}</span>}{renew && <span className="batch-kind">续签</span>}</span>
                    <span className="batch-meta" title={summary}>{summary}</span>
                </>,
                extra: automaticDNS && !values.secret_id ? <span className="batch-missing">选择密钥</span> : undefined,
                children: <div className={styles.fields}>
                    <Form.Item name={["items", index, "challenge"]} label="验证类型" rules={[{required: true, message: "请选择验证类型"}]}>
                        <Select size="small" variant="filled" aria-label={`${domain} 的验证类型`} disabled={disabled || wildcard || ipAddress} options={[
                            {label: "HTTP", value: "http", disabled: wildcard},
                            {label: "DNS", value: "dns", disabled: ipAddress},
                        ]} onChange={(value) => {
                            if (value !== "dns" || manual) form.setFieldValue(["items", index, "secret_id"], undefined);
                        }}/>
                    </Form.Item>
                    <Form.Item name={["items", index, "type"]} label="验证模式" rules={[{required: true, message: "请选择验证模式"}]}>
                        <Select size="small" variant="filled" aria-label={`${domain} 的验证模式`} disabled={disabled} options={[
                            {label: "自动", value: 2},
                            {label: "手动", value: 1},
                        ]} onChange={(value) => {
                            form.setFieldValue(["items", index, "auto_renew"], Number(value) === 1 ? 0 : 1);
                            if (Number(value) === 1 || values.challenge !== "dns") form.setFieldValue(["items", index, "secret_id"], undefined);
                        }}/>
                    </Form.Item>
                    {!manual && <Form.Item name={["items", index, "auto_renew"]} label="自动续签" rules={[{required: true, message: "请选择是否自动续签"}]}>
                        <Select size="small" variant="filled" aria-label={`${domain} 的自动续签`} disabled={disabled} options={[
                            {label: "关闭", value: 0},
                            {label: "开启", value: 1},
                        ]}/>
                    </Form.Item>}
                    <Form.Item name={["items", index, "ca"]} label="签发环境" rules={[{required: !renew, message: "请选择签发环境"}]}>
                        <Select size="small" variant="filled" aria-label={`${domain} 的签发环境`} disabled={disabled} allowClear={renew} placeholder={renew ? "沿用原签发环境" : "选择签发环境"} options={[
                            {label: "正式环境", value: "production"},
                            {label: "测试环境", value: "staging"},
                        ]}/>
                    </Form.Item>
                    {automaticDNS && <Form.Item label="关联密钥">
                        <div className={styles.secret}>
                            <Form.Item name={["items", index, "secret_id"]} rules={[{required: true, message: "请选择云密钥"}]}>
                                <Select size="small" variant="filled" aria-label={`${domain} 的关联密钥`} showSearch={false} loading={secretLoading} disabled={disabled || secretLoading}
                                    placeholder={secretError ? "密钥加载失败" : !secretLoading && !secretOptions.length ? "暂无密钥" : "选择云密钥"} options={secretOptions}
                                    notFoundContent={secretLoading ? <Spin size="small"/> : "暂无可用密钥"}/>
                            </Form.Item>
                            {secretError && <Tooltip title="重新加载密钥"><Button size="small" color="default" variant="filled" htmlType="button" aria-label={`${domain} 重新加载密钥`} disabled={disabled || secretLoading} icon={<Icon type="ReloadOutlined"/>} onClick={onReloadSecrets}/></Tooltip>}
                            <Tooltip title={noSecrets ? "添加密钥" : "管理密钥"}>
                                <Button size="small" color="default" variant="filled" htmlType="button" aria-label={`${domain} ${noSecrets ? "添加密钥" : "管理密钥"}`} disabled={disabled}
                                    icon={<Icon type={noSecrets ? "PlusOutlined" : "KeyOutlined"}/>}
                                    onClick={() => onManageSecrets(index, noSecrets)}/>
                            </Tooltip>
                        </div>
                    </Form.Item>}
                </div>,
            }]}/>
        {error && <div className={styles.error} role="alert">{error}</div>}
    </section>;
};

export default memo(BatchForm);
