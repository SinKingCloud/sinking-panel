
export interface HeaderProps {
    path: string;
    disks: string[];
    keyword: string;
    uploading: boolean;
    clipboardCount: number;
    clipboardMode: any;
    pasting: boolean;
    pasteDisabled: boolean;
    directoryActionsDisabled: boolean;
    selectedCount: number;
    selectionClipboardDisabled: boolean;
    selectionOperationDisabled: boolean;
    selectionClearDisabled: boolean;
    onKeywordChange: (value: string) => void;
    onNavigate: (path: string) => void;
    onCreate: (type: any) => void;
    onUpload: () => void;
    onPaste: () => void;
    onCopySelected: () => void;
    onMoveSelected: () => void;
    onCompressSelected: () => void;
    onDeleteSelected: () => void;
    onClearSelection: () => void;
    onRemoteDownload: () => void;
    onOpenRecycle: () => void;
}

export interface HeaderStyles {
    hero: string;
    commandBar: string;
    pathSection: string;
    pathBar: string;
    searchBox: string;
    toolbarTrigger: string;
    toolbarAction: string;
    toolbarDropdown: string;
}
