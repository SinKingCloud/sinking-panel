import React from "react";
import {useTheme} from "sinking-antd";
import HeaderCommandBar from "./header-command-bar";
import HeaderHero from "./header-hero";
import HeaderPathBar from "./header-path-bar";
import type {HeaderProps} from "./header.types";
import useStyles from "./header.styles";

const Header = ({
    path,
    disks,
    keyword,
    uploading,
    clipboardCount,
    clipboardMode,
    pasting,
    pasteDisabled,
    directoryActionsDisabled,
    selectedCount,
    selectionClipboardDisabled,
    selectionOperationDisabled,
    selectionClearDisabled,
    onKeywordChange,
    onNavigate,
    onCreate,
    onUpload,
    onPaste,
    onCopySelected,
    onMoveSelected,
    onCompressSelected,
    onDeleteSelected,
    onClearSelection,
    onRemoteDownload,
    onOpenRecycle,
}: HeaderProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles({compact, dark});

    return (
        <>
            <HeaderHero
                disks={disks}
                onNavigate={onNavigate}
                styles={styles}/>
            <HeaderCommandBar
                keyword={keyword}
                uploading={uploading}
                onKeywordChange={onKeywordChange}
                onCreate={onCreate}
                onUpload={onUpload}
                onRemoteDownload={onRemoteDownload}
                onOpenRecycle={onOpenRecycle}
                styles={styles}/>
            <HeaderPathBar
                path={path}
                disks={disks}
                clipboardCount={clipboardCount}
                clipboardMode={clipboardMode}
                pasting={pasting}
                pasteDisabled={pasteDisabled}
                directoryActionsDisabled={directoryActionsDisabled}
                selectedCount={selectedCount}
                selectionClipboardDisabled={selectionClipboardDisabled}
                selectionOperationDisabled={selectionOperationDisabled}
                selectionClearDisabled={selectionClearDisabled}
                onNavigate={onNavigate}
                onPaste={onPaste}
                onCopySelected={onCopySelected}
                onMoveSelected={onMoveSelected}
                onCompressSelected={onCompressSelected}
                onDeleteSelected={onDeleteSelected}
                onClearSelection={onClearSelection}
                styles={styles}/>
        </>
    );
};

export type {HeaderProps} from "./header.types";
export default React.memo(Header);
