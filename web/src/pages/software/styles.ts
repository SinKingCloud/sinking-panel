import {createStyles} from "antd-style";

export default createStyles(({css, token}) => ({
    loading: css`min-height: 300px; display: grid; place-items: center;`,
    icon: css`
        display: inline-flex; align-items: center; justify-content: center; flex: none;
        width: 1em; height: 1em; color: ${token.colorPrimary};
        img { display: block; width: 100%; height: 100%; object-fit: contain; }
    `,
    title: css`display: inline-flex; align-items: center; gap: 9px; min-width: 0;`,
    name: css`
        display: flex; align-items: center; gap: 9px; min-width: 0;
        > span:first-child { color: ${token.colorPrimary}; flex: none; }
        button { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; border: 0; padding: 0; background: none; color: inherit; font: inherit; cursor: pointer; }
        button:hover { color: ${token.colorPrimary}; }
    `,
}));
