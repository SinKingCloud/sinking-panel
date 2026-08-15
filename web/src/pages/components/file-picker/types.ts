import type {InputProps} from "antd";

export type FilePickerMode = "directory" | "file";

export interface FilePickerProps extends Omit<
    InputProps,
    "addonAfter" | "addonBefore" | "onChange" | "suffix" | "value"
> {
    mode?: FilePickerMode;
    value?: string;
    onChange?: (value: string) => void;
}

export interface SelectedFile {
    name: string;
    path: string;
}
