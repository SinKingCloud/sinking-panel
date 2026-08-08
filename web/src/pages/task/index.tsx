import React, {useCallback, useRef} from "react";
import {Col, Empty, Row, Spin} from "antd";
import {Body, useTheme} from "sinking-antd";
import useEnum from "@/utils/enum";
import Form, {FormRef} from "./components/form";
import Header from "./components/header";
import Log, {LogRef} from "./components/log";
import Pagination from "./components/pagination";
import Table from "./components/table";
import useList from "./hooks/list";
import useStyles from "./styles";

const emptyEnum = {};

export default (): React.ReactNode => {
    const theme = useTheme();
    const isCompactMode = theme?.isCompactTheme?.() || false;
    const isDarkMode = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles({isCompactMode, isDarkMode});
    const [enumData, enumLoading] = useEnum("task");
    const formRef = useRef<FormRef>({} as FormRef);
    const logRef = useRef<LogRef>({} as LogRef);
    const list = useList();
    const typeData = enumData?.type || emptyEnum;
    const statusData = enumData?.status || emptyEnum;

    const openCreate = useCallback(() => formRef.current?.open(), []);
    const openEdit = useCallback((record: any) => formRef.current?.open(record), []);
    const openLog = useCallback((record: any) => logRef.current?.open(record), []);

    return (
        <Body space={false}>
            <Row className={styles.page} gutter={[0, isCompactMode ? 10 : 12]}>
                <Col span={24}>
                    <section className={styles.workspace}>
                        <Header
                            styles={styles}
                            keyword={list.keyword}
                            status={list.status}
                            sort={list.sort}
                            onKeywordChange={list.changeKeyword}
                            onSearch={list.search}
                            onStatusChange={list.changeStatus}
                            onSortChange={list.changeSort}
                            onCreate={openCreate}/>

                        <main className={styles.dataPanel}>
                            {enumLoading ? (
                                <div className={styles.loadingState}><Spin/></div>
                            ) : list.loading || list.tasks.length > 0 ? (
                                <Table
                                    className={styles.taskTable}
                                    tasks={list.tasks}
                                    loading={list.loading}
                                    typeData={typeData}
                                    statusData={statusData}
                                    operating={list.operating}
                                    onAction={list.confirmAction}
                                    onEdit={openEdit}
                                    onLog={openLog}/>
                            ) : (
                                <div className={styles.state}>
                                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE}/>
                                </div>
                            )}
                        </main>
                    </section>

                    <Pagination
                        styles={styles}
                        page={list.page}
                        pageSize={list.pageSize}
                        total={list.total}
                        onChange={list.changePage}/>
                </Col>
            </Row>
            <Form ref={formRef} typeData={typeData} onSuccess={list.reload}/>
            <Log ref={logRef}/>
        </Body>
    );
};
