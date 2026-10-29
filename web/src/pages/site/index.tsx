import {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Body, PageTable} from "sinking-antd";
import TypeManager, {TypeManagerRef} from "@/pages/components/type-manager";
import {getTypeList} from "@/service/api/type";
import useEnum from "@/utils/enum";
import useSiteHeader from "./components/header";
import useSiteTable from "./components/table";
import SiteForm from "./components/site-form";
import Settings from "./components/settings";
import CertManager, {CertManagerRef} from "@/pages/components/cert";
import SecretManager, {SecretManagerRef} from "@/pages/components/secret";
import HTTPService from "./components/http-service";
import useList from "./hooks/list";
import useStyles from "./styles";

const defaultTypeData = {"0": "静态网站", "1": "反向代理", "2": "通用网站", "3": "PHP网站"};
const defaultStatusData = {"0": "已启用", "1": "已停用"};

export default (): React.ReactNode | any => {
    const {styles} = useStyles();
    const [enumData, enumLoading] = useEnum("site");
    const formRef = useRef<any>({});
    const settingsRef = useRef<any>({});
    const certRef = useRef<CertManagerRef | null>(null);
    const secretRef = useRef<SecretManagerRef | null>(null);
    const httpRef = useRef<any>({});
    const typeManagerRef = useRef<TypeManagerRef>({} as TypeManagerRef);
    const typeVersionRef = useRef(0);
    const list = useList();
    const [certVersion, setCertVersion] = useState(0);
    const [secretVersion, setSecretVersion] = useState(0);
    const [typeItems, setTypeItems] = useState<any[] | null>(null);
    const typeData = enumData?.kind || defaultTypeData;
    const statusData = enumData?.status || defaultStatusData;
    const categoryData = useMemo(() => typeItems === null
        ? enumData?.type || {}
        : Object.fromEntries(typeItems.map((item) => [String(item.id), String(item.name)])), [enumData?.type, typeItems]);

    useEffect(() => {
        let active = true;
        const version = typeVersionRef.current;
        void getTypeList({body: {module: "site", order_by_field: "sort", order_by_type: "asc"}}).then((response) => {
            if (!active || typeVersionRef.current !== version || response?.code !== 200) {
                return;
            }
            setTypeItems(Array.isArray(response?.data?.list) ? response?.data?.list.filter(Boolean) : []);
        });
        return () => {
            active = false;
        };
    }, []);

    const openSettings = useCallback((record: any, section: string) => settingsRef.current?.open(record, section), []);
    const openCertManager = useCallback(() => certRef.current?.open(), []);
    const obtainCert = useCallback((domain: string | string[], onObtained: (cert: any) => void) => certRef.current?.openObtain(domain, onObtained), []);
    const deployCert = useCallback((domain: string, onSelected: (cert: any) => void) => certRef.current?.openDeploy(domain, onSelected), []);
    const renewCert = useCallback((certificate: any, onRenewed: (cert: any) => void, domain: string) => certRef.current?.openRenew(certificate, onRenewed, domain), []);
    const openTypeManager = useCallback(() => typeManagerRef.current?.open(), []);
    const handleCertChange = useCallback(() => setCertVersion((value) => value + 1), []);
    const handleSecretChange = useCallback(() => {
        setSecretVersion((value) => value + 1);
        setCertVersion((value) => value + 1);
    }, []);
    const handleTypesChange = useCallback((items: any[]) => {
        typeVersionRef.current += 1;
        const nextTypes = Object.fromEntries(items.map((item) => [String(item.id), item.name]));
        setTypeItems(items);
        if (list.typeId !== "0" && !nextTypes[list.typeId]) {
            list.changeTypeId("0");
        }
    }, [list.changeTypeId, list.typeId]);

    const header = useSiteHeader({
        keyword: list.keyword,
        searchField: list.searchField,
        typeId: list.typeId,
        type: list.type,
        status: list.status,
        typeItems: typeItems || undefined,
        typeData,
        statusData,
        refreshing: list.loading,
        onKeywordChange: list.changeKeyword,
        onSearchFieldChange: list.changeSearchField,
        onSearch: list.search,
        onTypeIdChange: list.changeTypeId,
        onManageTypes: openTypeManager,
        onTypeChange: list.changeType,
        onStatusChange: list.changeStatus,
        onRefresh: list.reload,
        onCreate: () => formRef.current?.open(),
        onHTTPSettings: () => httpRef.current?.open(),
        onManageCert: openCertManager,
        onManageSecret: () => secretRef.current?.open(),
    });
    const table = useSiteTable({
        className: styles.siteTable,
        sites: list.sites,
        loading: list.loading,
        categoryData,
        typeData,
        statusData,
        sort: list.sort,
        order: list.order,
        operating: list.operating,
        onSortChange: list.changeSort,
        onSettings: openSettings,
        onAction: list.confirmAction,
    });

    return (
        <Body loading={enumLoading}>
            <PageTable
                {...header}
                {...table}
                ariaLabel="网站管理"
                paginationAffix={{offsetBottom: 15}}
                empty={list.initialized && list.sites.length === 0}
                pagination={{
                    page: list.page,
                    pageSize: list.pageSize,
                    total: list.total,
                    unit: "条",
                    onChange: list.changePage,
                }}/>
            <SiteForm ref={formRef} typeItems={typeItems || undefined} onSuccess={list.reload}/>
            <Settings
                ref={settingsRef}
                certVersion={certVersion}
                typeItems={typeItems || undefined}
                onSuccess={list.reload}
                onObtainCert={obtainCert}
                onDeployCert={deployCert}
                onRenewCert={renewCert}
                onManageCert={openCertManager}/>
            <CertManager ref={certRef} secretVersion={secretVersion} onChange={handleCertChange}/>
            <SecretManager ref={secretRef} onChange={handleSecretChange}/>
            <HTTPService ref={httpRef}/>
            <TypeManager
                ref={typeManagerRef}
                module="site"
                onChange={handleTypesChange}
                onMutation={list.reload}/>
        </Body>
    );
};
