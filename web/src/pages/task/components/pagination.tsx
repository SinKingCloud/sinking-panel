import React from "react";
import {Card, Pagination as AntPagination} from "antd";

const Pagination = ({styles, page, pageSize, total, onChange}: any) => {
    if (total <= 0) {
        return null;
    }

    return (
        <Card className={styles.paginationCard} variant="borderless">
            <AntPagination
                className={styles.pagination}
                size="small"
                current={page}
                pageSize={pageSize}
                total={total}
                align="center"
                responsive
                showQuickJumper
                showLessItems
                showSizeChanger={{
                    showSearch: false,
                    variant: "filled",
                    size: "small",
                    className: styles.pageSizeSelect,
                    classNames: {popup: {root: styles.pageSizeDropdown}},
                } as any}
                pageSizeOptions={[10, 20, 50, 100]}
                showTotal={(nextTotal, range) => `第 ${range[0]}-${range[1]} 条 / 共 ${nextTotal} 条`}
                onChange={onChange}
            />
        </Card>
    );
};

export default React.memo(Pagination);
