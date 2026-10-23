import {createContext, useContext} from "react";
import type {RedisClientContext} from "./entities";

export const RedisContext = createContext<RedisClientContext | undefined>(undefined);

export function useRedisClient() {
    const context = useContext(RedisContext);
    if (!context) throw new Error("此组件需要放在 Redis 管理器中");
    return context;
}
