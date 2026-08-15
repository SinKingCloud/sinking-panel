export type FilePreviewKind = "image" | "video" | "audio";

export interface FilePreviewItem {
    name: string;
    path: string;
    size?: number;
    isDirectory?: boolean;
}

export interface FilePreviewRef {
    open: (files: readonly FilePreviewItem[], active?: string | number) => void;
    close: () => void;
}
