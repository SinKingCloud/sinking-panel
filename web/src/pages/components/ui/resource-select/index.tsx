import React, {forwardRef, useContext, useEffect, useRef, useState} from "react";
import {Button, Input, Space} from "antd";
import type {InputProps, InputRef} from "antd";
import DisabledContext from "antd/es/config-provider/DisabledContext";
import useVariant from "antd/es/form/hooks/useVariants";
import {createStyles} from "antd-style";
import {Icon} from "sinking-antd";

type Key = string | number;

export interface ResourceSelectDialogProps {
    open: boolean;
    value: Key[];
    onCancel: () => void;
    onSelect: (records: any[]) => void;
}

export interface ResourceSelectProps extends Omit<InputProps, "value" | "defaultValue" | "onChange" | "multiple"> {
    value?: Key | Key[];
    defaultValue?: Key | Key[];
    onChange?: (value: Key | Key[] | undefined, records: any[]) => void;
    multiple?: boolean;
    emptyValue?: Key;
    /** 已有的名称直接回显，只为缺少名称的已选 ID 请求详情。 */
    records?: any[];
    resolveRecord?: (id: Key) => Promise<any>;
    recordLabel?: (record: any) => string;
    refreshKey?: string | number;
    renderModal: (props: ResourceSelectDialogProps) => React.ReactNode;
}

const useStyles = createStyles(({css, token}) => ({
    field: css`
        min-width: 0;
        > .ant-input, > .ant-input-affix-wrapper { min-width: 0; flex: 1; }
        > .ant-btn { flex: none; }
        > .ant-btn:not(:disabled) .ant-btn-icon { color: ${token.colorTextTertiary}; }
        input[readonly] { text-overflow: ellipsis; }
    `,
    clear: css`
        display: inline-flex;
        align-items: center;
        justify-content: center;
        padding: 0;
        border: 0;
        background: transparent;
        color: ${token.colorTextQuaternary};
        cursor: pointer;
        font-size: ${token.fontSizeSM}px;
        &:hover { color: ${token.colorTextSecondary}; }
    `,
}));

/** 资源名称与选择按钮组成一个表单控件，列表交给对应资源的分页弹窗。 */
const ResourceSelect = forwardRef<InputRef, ResourceSelectProps>((props, ref) => {
    const {
        value, defaultValue, onChange, multiple, emptyValue, records = [], resolveRecord,
        recordLabel, refreshKey, renderModal, allowClear, disabled, size, className, style,
        onKeyDown, onClick, ...inputProps
    } = props;
    const {styles, cx} = useStyles();
    const contextDisabled = useContext(DisabledContext);
    const isDisabled = disabled ?? contextDisabled;
    const [variant] = useVariant("input", inputProps.variant);
    const [localValue, setLocalValue] = useState(defaultValue);
    const selectedValue = Object.prototype.hasOwnProperty.call(props, "value") ? value : localValue;
    const [open, setOpen] = useState(false);
    const [visited, setVisited] = useState(false);
    const [resolved, setResolved] = useState<any[]>([]);
    const resolvedKey = useRef(refreshKey);
    const requestRef = useRef(0);
    const latest = useRef({records, resolveRecord, resolved});
    latest.current = {records, resolveRecord, resolved};
    const keys = (Array.isArray(selectedValue) ? selectedValue : [selectedValue])
        .filter((key): key is Key => key !== undefined && key !== null && key !== "" && key !== emptyValue);
    const selectionKey = JSON.stringify(keys);

    useEffect(() => {
        let active = true;
        const requestId = ++requestRef.current;
        const source = latest.current;
        const cache = resolvedKey.current === refreshKey ? source.resolved : [];
        resolvedKey.current = refreshKey;
        const known = new Map<string, any>();
        [...cache, ...source.records].filter(Boolean).forEach((item) => known.set(String(item.id), {...known.get(String(item.id)), ...item}));
        const selected: Key[] = JSON.parse(selectionKey);
        const retained = selected.flatMap((id) => known.has(String(id)) ? [known.get(String(id))] : []);
        setResolved(retained);
        const missing = selected.filter((id) => !known.get(String(id))?.name);
        if (source.resolveRecord && missing.length) {
            void Promise.all(missing.map(async (id) => {
                try { return await source.resolveRecord!(id); } catch { return undefined; }
            })).then((items) => {
                if (active && requestRef.current === requestId) setResolved([...retained, ...items.filter((item) => item?.id !== undefined)]);
            });
        }
        return () => { active = false; };
        // 请求函数允许内联；父表单渲染不会重复请求详情。
    }, [selectionKey, refreshKey]);

    const known = new Map<string, any>();
    [...resolved, ...records].filter(Boolean).forEach((item) => known.set(String(item.id), {...known.get(String(item.id)), ...item}));
    const label = keys.map((id) => {
        const record = known.get(String(id));
        return record ? recordLabel?.(record) || record.name || String(id) : String(id);
    }).join("、");
    const select = (items: any[]) => {
        requestRef.current++;
        const selected = items.filter((item) => item?.id !== undefined).map((item) => ({...known.get(String(item.id)), ...item}));
        const next = multiple ? selected.map((item) => item.id) : selected[0]?.id ?? emptyValue;
        setResolved(selected);
        setLocalValue(next);
        setOpen(false);
        onChange?.(next, selected);
    };
    const show = () => {
        setVisited(true);
        setOpen(true);
    };

    return <>
        <Space.Compact block size={size} className={cx(styles.field, className)} style={style}>
            <Input {...inputProps} ref={ref} size={size} readOnly disabled={isDisabled}
                value={label} title={label || undefined}
                onClick={(event) => { onClick?.(event); if (!isDisabled && !event.defaultPrevented) show(); }}
                onKeyDown={(event) => {
                    onKeyDown?.(event);
                    if (!isDisabled && !event.defaultPrevented && (event.key === "Enter" || event.key === "ArrowDown")) {
                        event.preventDefault();
                        show();
                    }
                }}
                suffix={allowClear && keys.length > 0 && !isDisabled
                    ? <button type="button" className={styles.clear} aria-label="清除选择"
                        onClick={(event) => { event.stopPropagation(); select([]); }}><Icon type="CloseCircleFilled"/></button>
                    : inputProps.suffix}/>
            <Button color="default" variant={variant === "filled" ? "filled" : variant === "outlined" ? "outlined" : "text"}
                size={size} disabled={isDisabled} aria-label="选择" aria-haspopup="dialog" aria-expanded={open && !isDisabled}
                icon={<Icon type="UnorderedListOutlined"/>} onClick={show}/>
        </Space.Compact>
        {visited && renderModal({open: open && !isDisabled, value: keys, onCancel: () => setOpen(false), onSelect: select})}
    </>;
});

ResourceSelect.displayName = "ResourceSelect";

export default ResourceSelect;
