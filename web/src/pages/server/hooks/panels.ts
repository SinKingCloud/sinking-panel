import {useCallback, useEffect, useState} from "react";

interface PanelState {
    server: boolean;
    commands: boolean;
}

const storageKey = "terminal.panel.collapsed";
const defaultState: PanelState = {
    server: false,
    commands: true,
};

const loadState = (): PanelState => {
    if (typeof window === "undefined") {
        return {...defaultState};
    }
    try {
        const value = JSON.parse(window.localStorage.getItem(storageKey) || "{}") as Partial<PanelState>;
        return {
            server: typeof value.server === "boolean" ? value.server : defaultState.server,
            commands: typeof value.commands === "boolean" ? value.commands : defaultState.commands,
        };
    } catch {
        return {...defaultState};
    }
};

const usePanels = () => {
    const [state, setState] = useState<PanelState>(loadState);

    useEffect(() => {
        try {
            window.localStorage.setItem(storageKey, JSON.stringify(state));
        } catch {
            // Storage can be unavailable in privacy-restricted browser contexts.
        }
    }, [state]);

    const setServerCollapsed = useCallback((collapsed: boolean) => {
        setState((current) => ({...current, server: collapsed}));
    }, []);
    const setCommandsCollapsed = useCallback((collapsed: boolean) => {
        setState((current) => ({...current, commands: collapsed}));
    }, []);

    return {
        serverCollapsed: state.server,
        commandsCollapsed: state.commands,
        setServerCollapsed,
        setCommandsCollapsed,
    };
};

export default usePanels;
