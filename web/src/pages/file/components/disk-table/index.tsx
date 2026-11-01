import {useEffect, useMemo, useState} from "react";
import type {Key} from "react";
import {Tooltip} from "antd";
import {Icon} from "sinking-antd";
import type {TableColumnsType} from "antd";
import {getFileDiskName} from "../../utils";
import useStyles from "../table/styles";

// useDiskTable 将磁盘作为目录展示，点击后进入对应的实际路径。
const useDiskTable = ({disks, keyword, active, loading, onOpen}: {
    disks: string[];
    keyword: string;
    active: boolean;
    loading: boolean;
    onOpen: (path: string) => void;
}) => {
    const {styles} = useStyles();
    const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([]);
    const items = useMemo(() => {
        const search = keyword.toLowerCase();
        return disks.map((path) => ({path, name: getFileDiskName(path)}))
            .filter((item) => `${item.name} ${item.path}`.toLowerCase().includes(search));
    }, [disks, keyword]);
    useEffect(() => {
        const visible = new Set(items.map((item) => item.path));
        setSelectedRowKeys((current) => {
            const next = active ? current.filter((key) => visible.has(String(key))) : [];
            return next.length === current.length ? current : next;
        });
    }, [active, items]);
    const columns: TableColumnsType<{path: string; name: string}> = [
        {
            title: "名称", dataIndex: "name", key: "name", width: 300,
            render: (_, record) => <button
                type="button"
                className={styles.fileNameButton}
                disabled={loading}
                aria-label={`打开磁盘 ${record.path}`}
                onClick={() => onOpen(record.path)}>
                <span className={styles.fileNameContent}>
                    <Icon className={`${styles.fileIcon} folder`} type="FolderOutlined"/>
                    <Tooltip title={record.name}>
                        <span className={`${styles.fileName} file-name`}>{record.name}</span>
                    </Tooltip>
                </span>
            </button>,
        },
        {
            title: "路径", dataIndex: "path", ellipsis: true,
            render: (value) => <span className={styles.fileMeta}>{value}</span>,
        },
    ];
    return {
        columns,
        dataSource: items,
        rowKey: "path",
        loading,
        rowSelection: {
            selectedRowKeys,
            columnWidth: 44,
            fixed: true,
            onChange: setSelectedRowKeys,
            getCheckboxProps: () => ({disabled: loading}),
            clearDisabled: loading,
        },
    };
};

export default useDiskTable;
