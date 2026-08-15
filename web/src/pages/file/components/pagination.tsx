import React from "react";
import {Card, Pagination as AntPagination} from "antd";
import {useTheme} from "sinking-antd";
import useStyles from "./pagination.styles";

interface PaginationProps {
    page: number;
    pageSize: number;
    total: number;
    disabled?: boolean;
    onChange: (page: number, pageSize: number) => void;
}

const Pagination = ({page, pageSize, total, disabled, onChange}: PaginationProps) => {
    const theme = useTheme();
    const compact = Boolean(theme?.isCompactTheme?.());
    const dark = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles({compact, dark});
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
                disabled={disabled}
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
                }}
                pageSizeOptions={[10, 20, 50, 100]}
                showTotal={(nextTotal, range) => `第 ${range[0]}-${range[1]} 项 / 共 ${nextTotal} 项`}
                onChange={onChange}
            />
        </Card>
    );
};

export default React.memo(Pagination);
