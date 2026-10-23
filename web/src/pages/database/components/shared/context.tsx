import {createContext, useContext} from "react";
import type {DatabaseClientContext} from "./entities";

export const DatabaseContext = createContext<DatabaseClientContext | undefined>(undefined);

export function useDatabaseClient() {
    const context = useContext(DatabaseContext);
    if (!context) throw new Error("数据库操作组件需要放在对应的数据库面板中");
    return context;
}
