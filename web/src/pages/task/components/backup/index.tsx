import {useRef} from "react";
import {Button, Col, Form, InputNumber, Row, Select} from "antd";
import {createStyles} from "antd-style";
import {Icon, useTheme} from "sinking-antd";
import FileSelect from "@/pages/components/file-select";
import SiteSelect from "@/pages/components/site-select";
import OssSelect from "@/pages/components/oss-select";
import {validateBackupPaths} from "./config";
import type {BackupType} from "./config";

const useStyles = createStyles<{compact: boolean}>(({css, token}, {compact}) => ({
    directoryRow: css`
        display: flex;
        align-items: flex-start;
        gap: ${compact ? 6 : 8}px;
        min-width: 0;

        > .ant-form-item {
            flex: 1;
            min-width: 0;
            margin-bottom: ${compact ? 6 : 8}px;
        }

        > .ant-btn { flex: none; }
    `,
    storage: css`
        > .ant-col:last-child > .ant-form-item {
            margin-bottom: 0;
        }
    `,
    lastRowItem: css`
        @media (min-width: 576px) {
            && {
                margin-bottom: 0;
            }
        }
    `,
    note: css`
        margin: ${compact ? 6 : 8}px 0 0;
        color: ${token.colorTextTertiary};
        font-size: ${token.fontSizeSM}px;
        line-height: ${token.lineHeight};
    `,
}));

// Backup 与任务表单共用保存及校验，按内容选择来源，按存储方式显示保留策略。
const Backup = () => {
    const form = Form.useFormInstance();
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const optionDrafts = useRef<Partial<Record<BackupType, any>>>({});
    const type = Form.useWatch(["backup", "type"], form) ?? form.getFieldValue(["backup", "type"]);
    const allSites = Form.useWatch(["backup", "options", "all_sites"], form);
    const storage = Form.useWatch(["backup", "storage"], form);
    const ossId = Form.useWatch(["backup", "oss_id"], form);
    const cloud = storage === "oss";

    const changeType = (nextType: BackupType) => {
        optionDrafts.current[type as BackupType] = form.getFieldValue(["backup", "options"]);
        const options = optionDrafts.current[nextType]
            || (nextType === "site" ? {all_sites: true, site_ids: []} : {paths: nextType === "directory" ? [""] : []});
        form.setFieldValue(["backup", "options"], options);
    };

    return <>
        <Row gutter={[16, 0]}>
            <Col xs={24} sm={type === "site" ? 12 : 24}>
                <Form.Item name={["backup", "type"]} label="备份内容" rules={[{required: true, message: "请选择备份内容"}]}>
                    <Select placeholder="选择备份内容" onChange={changeType} options={[
                        {label: "网站目录", value: "site"},
                        {label: "指定目录", value: "directory"},
                        {label: "指定文件", value: "file"},
                    ]}/>
                </Form.Item>
            </Col>
            {type === "site" && <Col xs={24} sm={12}>
                <Form.Item name={["backup", "options", "all_sites"]} label="备份范围"
                    rules={[{required: true, message: "请选择备份范围"}]}
                    getValueProps={(value) => ({value: value === undefined ? undefined : value ? "all" : "selected"})}
                    getValueFromEvent={(value) => value === "all"}>
                    <Select placeholder="选择备份范围" options={[{label: "全部网站", value: "all"}, {label: "指定网站", value: "selected"}]}/>
                </Form.Item>
            </Col>}
        </Row>
        {type === "site" && <Form.Item name={["backup", "options", "site_ids"]} label="选择网站" hidden={allSites !== false}
            rules={allSites === false ? [{required: true, type: "array", min: 1, message: "请至少选择一个网站"}] : []}>
            <SiteSelect multiple requireRoot allowClear placeholder="选择需要备份的网站"/>
        </Form.Item>}
        {type === "directory" && <Form.Item label="选择目录" required>
            <Form.List name={["backup", "options", "paths"]} rules={[{
                validator: async (_, paths: string[] = []) => {
                    if (paths.length === 0) throw new Error("请至少选择一个目录");
                    if (paths.length > 1000) throw new Error("最多选择 1000 个目录");
                },
            }]}>
                {(fields, {add, remove}, {errors}) => <>
                    {fields.map((field) => <div className={styles.directoryRow} key={field.key}>
                        <Form.Item name={field.name} rules={[{
                            validator: async (_, path) => { validateBackupPaths([path], "directory"); },
                        }]}
                            getValueProps={(value) => ({value: value ? [value] : []})}
                            getValueFromEvent={(paths: string[]) => paths[0] || ""}>
                            <FileSelect mode="directory" allowClear placeholder="请输入或选择需要备份的目录"
                                aria-label={`备份目录 ${field.name + 1}`}
                                buttonProps={{color: "default", variant: "filled", style: {borderColor: "transparent"}}}/>
                        </Form.Item>
                        <Button color="danger" variant="filled" icon={<Icon type="DeleteOutlined"/>}
                            aria-label={`移除备份目录 ${field.name + 1}`} onClick={() => remove(field.name)}/>
                    </div>)}
                    <Button type="dashed" block icon={<Icon type="PlusOutlined"/>}
                        disabled={fields.length >= 1000} onClick={() => add("")}>
                        添加目录
                    </Button>
                    <Form.ErrorList errors={errors}/>
                </>}
            </Form.List>
        </Form.Item>}
        {type === "file" && <Form.Item name={["backup", "options", "paths"]} label="选择文件" required rules={[{
            validator: async (_, paths) => { validateBackupPaths(paths, "file"); },
        }]}>
            <FileSelect mode="file" multiple placeholder="请输入或选择需要备份的文件"
                buttonProps={{color: "default", variant: "filled", style: {borderColor: "transparent"}}}/>
        </Form.Item>}
        <Row gutter={[16, 0]} className={styles.storage}>
            <Col xs={24} sm={12}>
                <Form.Item name={["backup", "storage"]} label="存储方式" className={cloud ? undefined : styles.lastRowItem}
                    rules={[{required: true, message: "请选择存储方式"}]}>
                    <Select placeholder="选择存储方式" options={[{label: "本地存储", value: "local"}, {label: "云端存储", value: "oss"}]}
                        onChange={(value) => {
                            if (value === "local" && form.getFieldValue(["backup", "local_keep"]) === 0) {
                                form.setFieldValue(["backup", "local_keep"], 7);
                            }
                            if (value === "oss" && !form.getFieldValue(["backup", "oss_keep"])) {
                                form.setFieldValue(["backup", "oss_keep"], 7);
                            }
                        }}/>
                </Form.Item>
            </Col>
            {cloud && <Col xs={24} sm={12}>
                <Form.Item name={["backup", "oss_id"]} label="存储实例" rules={[{
                    required: true,
                    type: "number",
                    min: 1,
                    message: "请选择存储实例",
                }]}>
                    <OssSelect emptyValue={0} placeholder="选择存储实例" onChange={(id) => {
                        if (String(id) !== String(ossId)) form.setFieldValue(["backup", "prefix"], "backups");
                    }}/>
                </Form.Item>
            </Col>}
            {cloud && <Col span={24}>
                <Form.Item name={["backup", "prefix"]} label="备份目录" rules={[
                    {required: true, whitespace: true, transform: (value) => value === "" ? "/" : value, message: "请选择备份目录"},
                    {max: 512, message: "备份目录不能超过 512 个字符"},
                ]}
                    getValueProps={(value) => ({value: value === undefined || value === null ? [] : [value]})}
                    getValueFromEvent={(paths: string[]) => paths[0]}>
                    <OssSelect mode="directory" ossId={ossId} placeholder={Number(ossId) > 0 ? "/" : "请先选择存储实例"} maxLength={512}/>
                </Form.Item>
            </Col>}
            <Col xs={24} sm={12}>
                <Form.Item name={["backup", "local_keep"]} label="本地保留副本" className={styles.lastRowItem} rules={[{
                    required: true,
                    type: "integer",
                    min: cloud ? 0 : 1,
                    max: 1000,
                    message: cloud ? "请输入 0 到 1000 份" : "请输入 1 到 1000 份",
                }]}>
                    <InputNumber min={cloud ? 0 : 1} max={1000} precision={0} suffix="份" placeholder="输入保留份数" style={{width: "100%"}}/>
                </Form.Item>
            </Col>
            {cloud && <Col xs={24} sm={12}>
                <Form.Item name={["backup", "oss_keep"]} label="云端保留副本" rules={[{
                    required: true,
                    type: "integer",
                    min: 1,
                    max: 1000,
                    message: "请输入 1 到 1000 份",
                }]}>
                    <InputNumber min={1} max={1000} precision={0} suffix="份" placeholder="输入保留份数" style={{width: "100%"}}/>
                </Form.Item>
            </Col>}
        </Row>
        <p className={styles.note}>{type === "site"
            ? "按网站分别保留，备份成功后清理旧副本。网站目录备份不包含数据库。"
            : `按${type === "directory" ? "目录" : "文件"}分别保留，备份成功后清理旧副本。`}</p>
    </>;
};

export default Backup;
