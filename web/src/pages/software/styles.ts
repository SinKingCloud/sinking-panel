import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    loading: css`min-height: 300px; display: grid; place-items: center;`,
    name: css`
        display: flex; align-items: center; gap: 9px; min-width: 0;
        > span:first-child { color: ${token.colorPrimary}; flex: none; }
        button { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; border: 0; padding: 0; background: none; color: inherit; font: inherit; cursor: pointer; }
        button:hover { color: ${token.colorPrimary}; }
    `,
}));
