import {useCallback, useMemo} from "react";
import type {ComponentProps} from "react";
import type {ModalTableProps} from "sinking-antd";
import ResourceSelect from "@/pages/components/ui/resource-select";
import ResourceSelectModal from "@/pages/components/ui/resource-select/modal";
import {getSiteInfo, getSiteList} from "@/service/api/site";

type SiteSelectProps = Omit<ComponentProps<typeof ResourceSelect>, "resolveRecord" | "renderModal"> & {
    stringValue?: boolean;
    zIndex?: number;
};

export default function SiteSelect({stringValue, zIndex, ...props}: SiteSelectProps) {
    const normalize = useCallback((record: any) => ({...record, id: stringValue ? String(record?.id) : Number(record?.id)}), [stringValue]);
    const request = useCallback<NonNullable<ModalTableProps<any>["request"]>>(async ({current, pageSize, keyword}) => {
        const response = await getSiteList({body: {page: current, page_size: pageSize, keyword}});
        if (response?.code !== 200) throw new Error(response?.message || "获取网站列表失败");
        return {data: (response?.data?.list || []).map(normalize), total: Number(response?.data?.total || 0), success: true};
    }, [normalize]);
    const resolveRecord = useCallback(async (id: string | number) => {
        const response = await getSiteInfo({body: {id}});
        if (response?.code !== 200 || !response?.data) return undefined;
        return normalize(response.data);
    }, [normalize]);
    const columns = useMemo<ModalTableProps<any>["columns"]>(() => [
        {title: "网站名称", dataIndex: "name", width: 190, ellipsis: true},
        {title: "网站目录", dataIndex: "root", width: 260, ellipsis: true, render: (root: string) => root || "未设置目录"},
    ], []);

    return <ResourceSelect {...props} resolveRecord={resolveRecord}
        renderModal={(modal) => <ResourceSelectModal {...modal} title="选择网站" placeholder="搜索网站名称"
            multiple={props.multiple} zIndex={zIndex} columns={columns} request={request}
            onSelect={(records) => modal.onSelect(records.map(normalize))}/>}/>;
}
