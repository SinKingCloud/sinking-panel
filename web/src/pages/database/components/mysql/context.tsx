import {createContext, useContext} from "react";
import type {MysqlClientContext} from "./entities";

export const MysqlContext = createContext<MysqlClientContext | undefined>(undefined);

export function useMysqlClient() {
    const context = useContext(MysqlContext);
    if (!context) throw new Error("此组件需要放在 MySQL 管理器中");
    return context;
}
