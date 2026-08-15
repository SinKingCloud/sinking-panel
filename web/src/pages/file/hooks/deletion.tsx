import {useCallback, useEffect, useRef} from "react";
import {App, Checkbox} from "antd";
import {deleteFile} from "@/service/api/file";
import type {FileRecord} from "@/service/api/file";
import type {FileOperationLock} from "./operation-lock";
import {joinFilePath} from "../utils";

interface DeleteConfirmState {
    id: symbol;
    destroy: () => void;
}

export interface UseFileDeletionOptions {
    path: string;
    selectedRecords: readonly FileRecord[];
    navigationVersionRef: {current: number};
    operationLock: FileOperationLock;
    reload: () => void;
    removeSelection: (paths: Iterable<string>) => void;
}

const useFileDeletion = ({
    path,
    selectedRecords,
    navigationVersionRef,
    operationLock,
    reload,
    removeSelection,
}: UseFileDeletionOptions) => {
    const {message, modal} = App.useApp();
    const confirmRef = useRef<DeleteConfirmState | undefined>(undefined);
    const {begin, finish, beginMany, finishMany} = operationLock;

    const closeConfirm = useCallback(() => {
        const current = confirmRef.current;
        confirmRef.current = undefined;
        current?.destroy();
    }, []);

    const removeRecord = useCallback(async (record: FileRecord, permanentlyDelete: boolean) => {
        const target = joinFilePath(path, record.name);
        if (!begin(target)) {
            return;
        }
        try {
            const response = await deleteFile({body: {paths: [target], recycle: !permanentlyDelete}});
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || "删除失败");
                return;
            }
            message.success(response.message || (permanentlyDelete ? "已彻底删除" : "已移至回收站"));
            reload();
        } finally {
            finish(target);
        }
    }, [begin, finish, message, path, reload]);

    const confirmRecord = useCallback((record: FileRecord) => {
        const navigationVersion = navigationVersionRef.current;
        closeConfirm();
        const confirmId = Symbol("delete");
        let permanentlyDelete = false;
        const instance = modal.confirm({
            title: "删除",
            content: (
                <div>
                    <div>确定删除“{record.name}”吗？</div>
                    <Checkbox
                        style={{marginTop: 12}}
                        onChange={(event) => {
                            permanentlyDelete = event.target.checked;
                        }}>
                        彻底删除
                    </Checkbox>
                </div>
            ),
            okText: "确认",
            cancelText: "取消",
            okButtonProps: {danger: true, type: "default"},
            mask: {closable: true},
            onCancel: () => {
                if (confirmRef.current?.id === confirmId) {
                    confirmRef.current = undefined;
                }
            },
            onOk: () => {
                if (confirmRef.current?.id === confirmId) {
                    confirmRef.current = undefined;
                }
                if (navigationVersionRef.current !== navigationVersion) {
                    message.warning("目录已切换，请重新操作");
                    return;
                }
                return removeRecord(record, permanentlyDelete);
            },
        });
        confirmRef.current = {id: confirmId, destroy: instance.destroy};
    }, [closeConfirm, message, modal, navigationVersionRef, removeRecord]);

    const removeMany = useCallback(async (records: readonly FileRecord[], permanentlyDelete: boolean) => {
        const targets = records.map((record) => joinFilePath(path, record.name));
        if (targets.length === 0) {
            return;
        }
        if (!beginMany(targets)) {
            message.warning("选中的文件正在执行其他操作");
            return;
        }
        try {
            const response = await deleteFile({body: {paths: targets, recycle: !permanentlyDelete}});
            if (!response) {
                return;
            }
            if (response.code !== 200) {
                message.error(response.message || "删除失败");
                reload();
                return;
            }
            removeSelection(targets);
            reload();
            message.success(permanentlyDelete
                ? `已彻底删除 ${targets.length} 项`
                : `已将 ${targets.length} 项移至回收站`);
        } finally {
            finishMany(targets);
        }
    }, [beginMany, finishMany, message, path, reload, removeSelection]);

    const confirmSelection = useCallback(() => {
        if (selectedRecords.length === 0) {
            return;
        }
        const records = [...selectedRecords];
        const navigationVersion = navigationVersionRef.current;
        closeConfirm();
        const confirmId = Symbol("batch-delete");
        let permanentlyDelete = false;
        const instance = modal.confirm({
            title: "批量删除",
            content: (
                <div>
                    <div>确定删除选中的 {records.length} 项吗？</div>
                    <Checkbox
                        style={{marginTop: 12}}
                        onChange={(event) => {
                            permanentlyDelete = event.target.checked;
                        }}>
                        彻底删除
                    </Checkbox>
                </div>
            ),
            okText: "确认",
            cancelText: "取消",
            okButtonProps: {danger: true, type: "default"},
            mask: {closable: true},
            onCancel: () => {
                if (confirmRef.current?.id === confirmId) {
                    confirmRef.current = undefined;
                }
            },
            onOk: () => {
                if (confirmRef.current?.id === confirmId) {
                    confirmRef.current = undefined;
                }
                if (navigationVersionRef.current !== navigationVersion) {
                    message.warning("目录已切换，请重新选择");
                    return;
                }
                return removeMany(records, permanentlyDelete);
            },
        });
        confirmRef.current = {id: confirmId, destroy: instance.destroy};
    }, [closeConfirm, message, modal, navigationVersionRef, removeMany, selectedRecords]);

    useEffect(() => () => closeConfirm(), [closeConfirm]);

    return {confirmRecord, confirmSelection, closeConfirm};
};

export default useFileDeletion;
