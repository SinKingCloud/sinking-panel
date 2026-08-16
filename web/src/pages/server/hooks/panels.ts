import {useCallback, useEffect, useState} from "react";

interface PanelState {
    server: boolean;
    scripts: boolean;
}

const storageKey = "terminal";
const defaultState: PanelState = {
    server: false,
    scripts: true,
};

const loadState = (): PanelState => {
    if (typeof window === "undefined") {
        return {...defaultState};
    }
    try {
        const value = JSON.parse(window.localStorage.getItem(storageKey) || "{}") as Partial<PanelState> & {
            commands?: boolean;
        };
        return {
            server: typeof value.server === "boolean" ? value.server : defaultState.server,
            scripts: typeof value.scripts === "boolean"
                ? value.scripts
                : typeof value.commands === "boolean" ? value.commands : defaultState.scripts,
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
    const setScriptsCollapsed = useCallback((collapsed: boolean) => {
        setState((current) => ({...current, scripts: collapsed}));
    }, []);

    return {
        serverCollapsed: state.server,
        scriptsCollapsed: state.scripts,
        setServerCollapsed,
        setScriptsCollapsed,
    };
};

export default usePanels;
