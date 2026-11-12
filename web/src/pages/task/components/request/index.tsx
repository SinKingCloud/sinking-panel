import {Button, Col, Form, Input, Row, Select} from "antd";
import {createStyles} from "antd-style";
import {Icon, useTheme} from "sinking-antd";
import {requestBodyMethods} from "../../utils";

const methodOptions = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"].map((value) => ({
    label: value,
    value,
}));

const useStyles = createStyles<{compact: boolean}>(({css}, {compact}) => ({
    headerRow: css`
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
        align-items: start;
        gap: ${compact ? 6 : 8}px;
        margin-bottom: ${compact ? 6 : 8}px;

        > .ant-form-item {
            min-width: 0;
            margin-bottom: 0;
        }

        @media (max-width: 575px) {
            grid-template-columns: minmax(0, 1fr) auto;

            > .ant-form-item:nth-child(2) {
                grid-column: 1;
            }

            > .ant-btn {
                grid-column: 2;
                grid-row: 1;
            }
        }
    `,
}));

const Request = () => {
    const form = Form.useFormInstance();
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const method = Form.useWatch(["request", "method"], form);
    const hasBody = requestBodyMethods.includes(method);

    return (
        <>
            <Row gutter={[16, 0]}>
                <Col xs={24} sm={7}>
                    <Form.Item
                        name={["request", "method"]}
                        label="请求方式"
                        rules={[
                            {required: true, whitespace: true, message: "请输入请求方式"},
                            {pattern: /^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/, message: "请求方式格式不合法"},
                        ]}>
                        <Select options={methodOptions} placeholder="选择请求方式"/>
                    </Form.Item>
                </Col>
                <Col xs={24} sm={17}>
                    <Form.Item
                        name={["request", "url"]}
                        label="请求地址"
                        rules={[
                            {required: true, whitespace: true, message: "请输入请求地址"},
                            {
                                validator: async (_, value) => {
                                    if (!String(value || "").trim()) {
                                        return;
                                    }
                                    try {
                                        const url = new URL(String(value).trim());
                                        if ((url.protocol === "http:" || url.protocol === "https:") && url.hostname) {
                                            return;
                                        }
                                    } catch {
                                        throw new Error("请输入有效的HTTP或HTTPS地址");
                                    }
                                    throw new Error("请输入有效的HTTP或HTTPS地址");
                                },
                            },
                        ]}>
                        <Input placeholder="https://example.com/api"/>
                    </Form.Item>
                </Col>
            </Row>
            <Form.Item label="请求头" style={hasBody ? undefined : {marginBottom: 0}}>
                <Form.List
                    name={["request", "headers"]}
                    rules={[{
                        validator: async (_, headers: any[] = []) => {
                            if (headers.length > 100) {
                                throw new Error("请求头数量不能超过100个");
                            }
                            const names = headers.map((header) => String(header?.name || "").trim().toLowerCase()).filter(Boolean);
                            if (new Set(names).size !== names.length) {
                                throw new Error("请求头名称不能重复");
                            }
                            if (names.some((name) => name === "content-length" || name === "transfer-encoding" || name === "trailer")) {
                                throw new Error("不支持自定义传输层请求头");
                            }
                            if (headers.some((header) => String(header?.name || "").trim().toLowerCase() === "host" && !String(header?.value || "").trim())) {
                                throw new Error("Host请求头不能为空");
                            }
                        },
                    }]}>
                    {(fields, {add, remove}, {errors}) => (
                        <div>
                            {fields.map((field) => (
                                <div className={styles.headerRow} key={field.key}>
                                    <Form.Item
                                        name={[field.name, "name"]}
                                        rules={[
                                            {required: true, whitespace: true, message: "请输入名称"},
                                            {pattern: /^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/, message: "名称格式不合法"},
                                        ]}>
                                        <Input placeholder="名称"/>
                                    </Form.Item>
                                    <Form.Item
                                        name={[field.name, "value"]}
                                        rules={[{pattern: /^[^\u0000-\u0008\u000A-\u001F\u007F]*$/, message: "请求头值包含非法字符"}]}>
                                        <Input placeholder="值"/>
                                    </Form.Item>
                                    <Button
                                        color="danger"
                                        variant="filled"
                                        icon={<Icon type="DeleteOutlined"/>}
                                        aria-label="删除请求头"
                                        onClick={() => remove(field.name)}/>
                                </div>
                            ))}
                            <Button type="dashed" block icon={<Icon type="PlusOutlined"/>} onClick={() => add({name: "", value: ""})}>
                                添加请求头
                            </Button>
                            <Form.ErrorList errors={errors}/>
                        </div>
                    )}
                </Form.List>
            </Form.Item>
            <Form.Item name={["request", "body"]} label="请求体" hidden={!hasBody} style={{marginBottom: 0}}>
                <Input.TextArea autoSize={{minRows: 3, maxRows: 8}} placeholder="请输入请求内容"/>
            </Form.Item>
        </>
    );
};

export default Request;
