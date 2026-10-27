import {Form, Input, InputNumber, Select, Switch} from "antd";
import FilePicker from "@/pages/components/file-picker";

// 插件页面可复用的通用表单项，字段、默认值和配置流程由插件维护。
export default function SoftwareField({field}: any) {
    const label = `${field.label || field.name}${field.unit ? `（${field.unit}）` : ""}`;
    const placeholder = field.placeholder || `${field.type === "select" || field.type === "file" ? "请选择" : "请输入"}${field.label || field.name}`;
    const booleanSelect = field.type === "select" && field.options?.some((option: any) => typeof option.value === "boolean");
    return <Form.Item name={Array.isArray(field.name) ? field.name : String(field.name).split(".")}
        label={label} valuePropName={field.type === "switch" ? "checked" : "value"}
        {...(booleanSelect ? {getValueProps: (value: any) => ({value: value === undefined ? undefined : String(value)}),
            getValueFromEvent: (value: string) => value === "true"} : {})}
        rules={[...(field.required ? [{required: true, message: placeholder}] : []), ...(field.must_true ? [{validator: (_: any, value: any) => value === true ? Promise.resolve() : Promise.reject(new Error(`请确认${field.label || field.name}`))}] : [])]}>
        {field.type === "number" ? <InputNumber min={field.min} max={field.max} step={field.step} placeholder={placeholder} style={{width: "100%"}}/>
            : field.type === "select" ? <Select placeholder={placeholder} options={booleanSelect ? field.options.map((option: any) => ({...option, value: String(option.value)})) : field.options || []}/>
                : field.type === "password" ? <Input.Password placeholder={placeholder} autoComplete="new-password"/>
                    : field.type === "switch" ? <Switch/>
                        : field.type === "file" ? <FilePicker mode="file" placeholder={placeholder}/>
                            : <Input placeholder={placeholder} autoComplete="off" maxLength={field.max_length}/>}
    </Form.Item>;
}
