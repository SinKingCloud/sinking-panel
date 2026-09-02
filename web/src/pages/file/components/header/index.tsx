import React from "react";
import {useTheme} from "sinking-antd";
import HeaderCommandBar from "../header-command-bar";
import HeaderHero from "../header-hero";
import HeaderPathBar from "../header-path-bar";
import type {HeaderProps} from "./types";
import useStyles from "./styles";

const Header = ({
    path,
    disks,
    keyword,
    uploading,
    refreshing,
    clipboardCount,
    clipboardMode,
    pasting,
    pasteDisabled,
    directoryActionsDisabled,
    selectedCount,
    editorMinimized,
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
    onRestoreEditor,
    onRemoteDownload,
    onOpenTerminal,
    onOpenRecycle,
    onRefresh,
}: HeaderProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const {styles} = useStyles({compact});

    return (
        <>
            <HeaderHero
                disks={disks}
                onNavigate={onNavigate}/>
            <HeaderCommandBar
                keyword={keyword}
                uploading={uploading}
                refreshing={refreshing}
                onKeywordChange={onKeywordChange}
                onCreate={onCreate}
                onUpload={onUpload}
                onRemoteDownload={onRemoteDownload}
                onOpenTerminal={onOpenTerminal}
                onOpenRecycle={onOpenRecycle}
                onRefresh={onRefresh}/>
            <HeaderPathBar
                path={path}
                disks={disks}
                clipboardCount={clipboardCount}
                clipboardMode={clipboardMode}
                pasting={pasting}
                pasteDisabled={pasteDisabled}
                directoryActionsDisabled={directoryActionsDisabled}
                selectedCount={selectedCount}
                editorMinimized={editorMinimized}
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
                onRestoreEditor={onRestoreEditor}
                styles={styles}/>
        </>
    );
};

export type {HeaderProps} from "./types";
export default React.memo(Header);
