import {Button, Select} from "antd";
import {createStyles} from "antd-style";

const useStyles = createStyles(({css, token}) => ({
    select: css`
        width: 100%; min-width: 0;
        .ant-select-selection-overflow { flex-wrap: nowrap; }
        .permission-summary { display: block; box-sizing: border-box; min-width: 0; max-width: 100%; padding-inline: ${token.paddingXS}px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    `,
    popup: css`
        max-width: calc(100vw - 32px);
        .permission-presets { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: ${token.marginXXS}px; padding: ${token.paddingXXS}px; margin-bottom: ${token.marginXXS}px; }
        .permission-presets:has(> .ant-btn:nth-child(3)):not(:has(> .ant-btn:nth-child(4))) { grid-template-columns: repeat(3, minmax(0, 1fr)); }
        .permission-presets .ant-btn { padding-inline: ${token.paddingXXS}px; font-size: ${token.fontSizeSM}px; }
    `,
}));

// 权限值直接受 Form 控制，快捷方案和逐项选择共用同一份数据。
export default function PermissionSelect({profiles, options, value = [], onChange, disabled, allowEmpty = false, ...props}: {
    profiles: {value: string; label: string; privileges: string[]; description?: string}[];
    options: {value: string; label: string}[];
    value?: string[];
    onChange?: (value: string[]) => void;
    disabled?: boolean;
    allowEmpty?: boolean;
    id?: string;
}) {
    const {styles} = useStyles();
    const selected = profiles.find((profile) => profile.privileges.length === value.length
        && profile.privileges.every((privilege) => value.includes(privilege)));
    return <Select {...props} className={styles.select} mode="multiple" value={value} onChange={onChange} disabled={disabled}
        aria-label="权限方案" options={options} showSearch={false} maxTagCount={0}
        placeholder={allowEmpty ? "不授予权限" : "请选择权限"}
        maxTagPlaceholder={() => selected?.label || `自定义 · ${value.length} 项`}
        tagRender={({label}) => <span className="permission-summary">{label}</span>}
        popupMatchSelectWidth={280} listHeight={240} classNames={{popup: {root: styles.popup}}}
        popupRender={(menu) => <>
            <div className="permission-presets" onMouseDown={(event) => event.preventDefault()}>
                {profiles.map((profile) => <Button key={profile.value} size="small" color={selected?.value === profile.value ? "primary" : "default"}
                    variant="filled" disabled={disabled || profile.privileges.some((privilege) => !options.some((option) => option.value === privilege))}
                    onClick={() => onChange?.([...profile.privileges])}>{profile.label}</Button>)}
            </div>
            {menu}
        </>}/>;
}
