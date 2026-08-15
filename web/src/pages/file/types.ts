import type {FileRecord} from "@/service/api/file";

export type FileCreateMode = "directory" | "file";
export type FileFormMode = FileCreateMode | "rename";

export interface FileFormResult {
    mode: FileFormMode;
    name: string;
    sourcePath?: string;
    targetPath?: string;
}

export type FileOperationMode = "compress" | "extract" | "remote-download";

export interface FileOperationContext {
    path: string;
    records?: FileRecord[];
    clearSelectionOnSuccess?: boolean;
}

export interface FileClipboardItem {
    sourcePath: string;
    sourceDirectory: string;
    name: string;
    isDir: boolean;
}

export type FileClipboardMode = "copy" | "move";

export interface FileClipboardState {
    mode: FileClipboardMode;
    items: FileClipboardItem[];
}
