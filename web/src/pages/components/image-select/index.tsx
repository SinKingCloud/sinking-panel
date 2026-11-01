import {forwardRef} from "react";
import type {InputRef} from "antd";
import ResourceSelect from "@/pages/components/ui/resource-select";
import type {ResourceSelectProps} from "@/pages/components/ui/resource-select";
import ImagesManager from "@/pages/software/components/container/images";
import {containerImages} from "@/service/api/container";

interface ImageSelectProps extends Omit<ResourceSelectProps, "renderModal"> {
    zIndex?: number;
}

// 镜像管理复用选择模式，表单只保存镜像 ID。
const ImageSelect = forwardRef<InputRef, ImageSelectProps>(({zIndex = 1150, ...props}, ref) => <ResourceSelect
    ref={ref}
    placeholder="请选择镜像"
    recordLabel={(record) => [...new Set([
        record?.name || record?.id,
        ...(Array.isArray(record?.tags) ? record.tags : []),
    ].filter(Boolean))].join(" / ")}
    resolveRecord={async (id) => {
        const response = await containerImages({body: {action: "info", id}});
        return response?.code === 200 ? response?.data : undefined;
    }}
    {...props}
    renderModal={({open, onCancel, value, onSelect}) => <ImagesManager
        open={open}
        onCancel={onCancel}
        zIndex={zIndex}
        selection={{value, onSelect}}/>}/>
);

ImageSelect.displayName = "ImageSelect";

export default ImageSelect;
