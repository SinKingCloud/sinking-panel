import {useMemo, useState} from "react";
import {App, Button} from "antd";
import {createStyles} from "antd-style";
import {ModalTable, Title, useTheme} from "sinking-antd";
import dayjs from "dayjs";
import {copyTextToClipboard} from "../shared/clipboard";

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    dialog: css`
        &.modal-table-loading { min-height: ${compact ? 200 : 240}px; }
        && .ui-table-data > [role="status"] { min-height: ${compact ? 140 : 160}px; padding: 12px; }
        && .ui-table-toolbar { display: flex; gap: 8px; }
        && .ui-table-toolbar > .ui-table-search { width: 240px; flex: 0 1 240px; }
        && .ui-table-toolbar .command-actions { width: auto; flex: none; margin-left: auto; }
        .request-history-entry { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 12px; }
        .request-history-content { display: grid; min-width: 0; gap: 5px; }
        .request-history-meta { display: flex; align-items: center; flex-wrap: wrap; gap: 4px 10px; min-width: 0; color: ${token.colorTextTertiary}; font-size: 11px; line-height: 1.5; }
        .request-history-time { white-space: nowrap; font-variant-numeric: tabular-nums; }
        .request-history-result { display: inline-flex; align-items: center; flex-wrap: wrap; gap: 4px 8px; }
        .request-history-status { display: inline-flex; align-items: center; gap: 5px; white-space: nowrap; }
        .request-history-status::before { content: ""; width: 5px; height: 5px; border-radius: 50%; background: ${token.colorSuccess}; }
        .request-history-status.is-failed::before { background: ${token.colorError}; }
        .request-history-actions { display: flex; align-items: center; flex: none; gap: 6px; }
        .request-history-actions .ant-btn { font-size: ${compact ? 11 : 12}px; }
        .request-history-request { margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: ${token.colorTextSecondary}; font-family: ${token.fontFamilyCode}; font-size: ${compact ? 11 : 12}px; font-weight: 400; line-height: 1.65; }
        @media (max-width: 575px) {
            .request-history-entry { gap: 8px; }
            .request-history-meta { gap: 4px 8px; }
        }
    `,
}));

export default ({open, history, loading, onCancel, onUse, onRefresh}: any) => {
    const compact = Boolean(useTheme()?.isCompactTheme?.());
    const {styles} = useStyles({compact});
    const {message} = App.useApp();
    const [keyword, setKeyword] = useState("");
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(5);
    const records = useMemo(() => {
        const search = keyword.trim().toLowerCase();
        return search ? history.filter((item: any) => `${item.method} ${item.path}\n${item.body}`.toLowerCase().includes(search)) : history;
    }, [history, keyword]);
    const currentPage = Math.min(page, Math.max(1, Math.ceil(records.length / pageSize)));

    return <ModalTable open={open} title={<Title>请求历史</Title>} rootClassName={styles.dialog}
        width={compact ? 520 : 560} onCancel={onCancel} scrollMode="modal" rowKey="time" showHeader={false} scroll={{x: undefined}}
        modalProps={{
            style: {maxWidth: "calc(100vw - 24px)"},
            afterClose: () => { setKeyword(""); setPage(1); setPageSize(5); },
        }}
        toolbar={{
            search: {value: keyword, placeholder: "搜索方法、路径或内容", ariaLabel: "搜索请求历史", onChange: (value) => { setKeyword(value); setPage(1); }},
            refresh: {ariaLabel: "刷新请求历史", onClick: onRefresh},
        }}
        pagination={{page: currentPage, pageSize, total: records.length, pageSizeOptions: [5, 10, 20, 50, 100],
            onChange: (nextPage, nextSize) => { setPage(nextSize === pageSize ? nextPage : 1); setPageSize(nextSize); }}}
        dataSource={records.slice((currentPage - 1) * pageSize, currentPage * pageSize)}
        columns={[{
            key: "history",
            render: (_: unknown, item: any) => <article className="request-history-entry">
                <div className="request-history-content">
                    <p className="request-history-request" title={`${item.method} ${item.path}`}>{item.method} {item.path}</p>
                    <div className="request-history-meta">
                        <time className="request-history-time" dateTime={dayjs(item.time).format()} title={dayjs(item.time).format("YYYY-MM-DD HH:mm:ss")}>{dayjs(item.time).format("MM-DD HH:mm:ss")}</time>
                        <span className="request-history-result"><span className={`request-history-status${item.success ? "" : " is-failed"}`}>{item.success ? "成功" : "失败"}</span>
                            {Number.isFinite(item.status) && <span>{item.status}</span>}
                            {Number.isFinite(item.elapsed_ms) && <span>{item.elapsed_ms} ms</span>}</span>
                    </div>
                </div>
                <div className="request-history-actions">
                    <Button size="small" color="default" variant="filled" autoInsertSpace={false} onClick={async () => {
                        if (await copyTextToClipboard(`${item.method} ${item.path}${item.body ? `\n${item.body}` : ""}`)) message.success("复制成功");
                        else message.error("复制失败，请手动复制");
                    }}>复制</Button>
                    <Button size="small" color="default" variant="filled" autoInsertSpace={false} disabled={loading}
                        onClick={() => onUse(item)}>填入</Button>
                </div>
            </article>,
        }]}/>;
};
