import {forwardRef, useMemo} from "react";
import type {ComponentProps} from "react";
import type {InputRef} from "antd";
import type {ModalTableProps} from "sinking-antd";
import ResourceSelect from "@/pages/components/ui/resource-select";
import ResourceSelectModal from "@/pages/components/ui/resource-select/modal";
import type {ResourceSelectModalProps} from "@/pages/components/ui/resource-select/modal";
import {getOssInfo, getOssList} from "@/service/api/oss";
import useEnum from "@/utils/enum";
import OssPathPicker from "./picker";
import type {OssPathPickerProps} from "./picker";

interface OssSelectModalProps extends Pick<ResourceSelectModalProps, "open" | "onCancel" | "value" | "onSelect" | "zIndex" | "toolbarActions"> {
    allowLocal?: boolean;
}

const request: NonNullable<ModalTableProps<any>["request"]> = async ({current, pageSize, keyword}) => {
    const response = await getOssList({body: {page: current, page_size: pageSize, keyword}});
    if (response?.code !== 200) throw new Error(response?.message || "获取存储列表失败");
    return {data: response?.data?.list || [], total: Number(response?.data?.total || 0), success: true};
};
const resolveRecord = async (id: string | number) => {
    const response = await getOssInfo({body: {id}});
    return response?.code === 200 ? response?.data : undefined;
};
export function OssSelectModal({allowLocal, toolbarActions, ...props}: OssSelectModalProps) {
    const [enumData] = useEnum("secret");
    const columns = useMemo<ModalTableProps<any>["columns"]>(() => [
        {title: "存储名称", dataIndex: "name", width: 190, ellipsis: true},
        {title: "密钥名称", dataIndex: "secret_name", width: 160, ellipsis: true,
            render: (name: string) => name || "—"},
        {title: "厂商", dataIndex: "provider", width: 110, ellipsis: true,
            render: (provider: number | null) => provider == null ? "—" : enumData?.provider?.[provider] || "未知厂商"},
        {title: "更新时间", dataIndex: "update_time", width: 170, ellipsis: true},
    ], [enumData]);
    const actions = useMemo<ResourceSelectModalProps["toolbarActions"]>(() => [
        ...(allowLocal ? [{key: "local", type: "button" as const, label: "本地存储", onClick: () => props.onSelect([{id: 0, name: "本地存储"}])}] : []),
        ...(toolbarActions || []),
    ], [allowLocal, toolbarActions, props.onSelect]);
    return <ResourceSelectModal {...props} title="选择存储" placeholder="搜索存储名称" columns={columns}
        request={request} toolbarActions={actions}/>;
}

export type OssSelectProps = OssPathPickerProps | (Omit<ComponentProps<typeof ResourceSelect>, "resolveRecord" | "renderModal"> & {
    mode?: "storage";
    zIndex?: number;
});

/** 默认选择 OSS 连接；文件、目录模式的单选和多选均使用路径数组。 */
const OssSelect = forwardRef<InputRef, OssSelectProps>((props, ref) => {
    if (props.mode === "file" || props.mode === "directory") {
        return <OssPathPicker {...props} ref={ref}/>;
    }
    const {zIndex, mode, ...inputProps} = props;
    return <ResourceSelect {...inputProps} ref={ref} resolveRecord={resolveRecord}
        renderModal={(modal) => <OssSelectModal {...modal} zIndex={zIndex}/>}/>;
});

OssSelect.displayName = "OssSelect";

export default OssSelect;
