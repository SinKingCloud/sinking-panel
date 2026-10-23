import type {DatabasePanelRecord} from "../entities";

export interface SqlEngine {
    query: string;
    objects: {key: string; label: string; icon: string}[];
}

export interface SqlHostProps {
    record: DatabasePanelRecord;
    initialTab?: string;
    engine: SqlEngine;
}

export interface SqlTarget {
    schema?: string;
    table: string;
    isView: boolean;
    schemas: string[];
    tables: any[];
    loading?: boolean;
    error?: string;
}

export interface SqlPanelProps extends SqlHostProps {
    section?: "browse" | "query" | "objects";
    hidden?: boolean;
    onTargetChange?: (target: SqlTarget) => void;
    onBrowse?: () => void;
    reloadKey?: number;
    tableValue?: string;
}
