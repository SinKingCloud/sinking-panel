import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Col, Empty, Row, Spin} from "antd";
import {Body, useTheme} from "sinking-antd";
import useEnum from "@/utils/enum";
import TypeManager, {
    TypeManagerRef,
    TypeRecord,
} from "@/pages/components/type-manager";
import {getAllTypes} from "@/service/api/type";
import Form, {FormRef} from "./components/form";
import Header from "./components/header";
import Log, {LogRef} from "./components/log";
import Pagination from "./components/pagination";
import Table from "./components/table";
import useList from "./hooks/list";
import useStyles from "./styles";

const emptyEnum = {};
const defaultExecTypeEnum = {
    "0": "系统脚本",
    "1": "HTTP请求",
};

export default (): React.ReactNode => {
    const theme = useTheme();
    const isCompactMode = theme?.isCompactTheme?.() || false;
    const isDarkMode = Boolean(theme?.isDarkMode?.() || theme?.isDarkTheme?.());
    const {styles} = useStyles({isCompactMode, isDarkMode});
    const [enumData, enumLoading] = useEnum("task");
    const formRef = useRef<FormRef>({} as FormRef);
    const logRef = useRef<LogRef>({} as LogRef);
    const typeManagerRef = useRef<TypeManagerRef>({} as TypeManagerRef);
    const typeVersionRef = useRef(0);
    const [typeItems, setTypeItems] = useState<TypeRecord[] | null>(null);
    const list = useList();
    const execTypeData = enumData?.exec_type && Object.keys(enumData.exec_type).length > 0
        ? enumData.exec_type
        : defaultExecTypeEnum;
    const typeData = useMemo(() => typeItems === null
        ? enumData?.type ?? emptyEnum
        : Object.fromEntries(typeItems.map((item) => [String(item.id), item.name])), [enumData?.type, typeItems]);
    const statusData = enumData?.status || emptyEnum;

    useEffect(() => {
        let active = true;
        const version = typeVersionRef.current;
        void getAllTypes("task").then((response) => {
            if (!active || typeVersionRef.current !== version || response?.code !== 200) {
                return;
            }
            setTypeItems(Array.isArray(response.data?.list) ? response.data.list : []);
        });
        return () => {
            active = false;
        };
    }, []);

    const openCreate = useCallback(() => formRef.current?.open(), []);
    const openEdit = useCallback((record: any) => formRef.current?.open(record), []);
    const openLog = useCallback((record: any) => logRef.current?.open(record), []);
    const openTypeManager = useCallback(() => typeManagerRef.current?.open(), []);
    const handleTypesChange = useCallback((items: TypeRecord[]) => {
        typeVersionRef.current += 1;
        const nextTypes = Object.fromEntries(items.map((item) => [String(item.id), item.name]));
        setTypeItems(items);
        if (list.typeId !== "0" && !nextTypes[list.typeId]) {
            list.changeTypeId("0");
        }
    }, [list.changeTypeId, list.typeId]);

    return (
        <Body space={false}>
            <Row className={styles.page} gutter={[0, isCompactMode ? 10 : 12]}>
                <Col span={24}>
                    <section className={styles.workspace}>
                        <Header
                            styles={styles}
                            keyword={list.keyword}
                            typeId={list.typeId}
                            execType={list.execType}
                            status={list.status}
                            typeData={typeData}
                            typeItems={typeItems || undefined}
                            execTypeData={execTypeData}
                            statusData={statusData}
                            onKeywordChange={list.changeKeyword}
                            onSearch={list.search}
                            onTypeIdChange={list.changeTypeId}
                            onManageTypes={openTypeManager}
                            onExecTypeChange={list.changeExecType}
                            onStatusChange={list.changeStatus}
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
                                    execTypeData={execTypeData}
                                    statusData={statusData}
                                    sort={list.sort}
                                    order={list.order}
                                    operating={list.operating}
                                    onSortChange={list.changeSort}
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
            <Form
                ref={formRef}
                typeData={typeData}
                typeItems={typeItems || undefined}
                execTypeData={execTypeData}
                onSuccess={list.reload}/>
            <Log ref={logRef}/>
            <TypeManager
                ref={typeManagerRef}
                module="task"
                onChange={handleTypesChange}
                onMutation={list.reload}/>
        </Body>
    );
};
