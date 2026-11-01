import {memo, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react";
import {Button, Input} from "antd";
import type {InputRef} from "antd";
import {Icon, useTheme} from "sinking-antd";
import useStyles from "../path-bar/styles";

// OSS 路径是对象前缀，保留重复斜杠、反斜杠和点目录，不使用本地路径规范化。
const OssPathBar = ({path, disabled, onNavigate}: {
    path: string;
    disabled: boolean;
    onNavigate: (path: string) => void;
}) => {
    const theme = useTheme();
    const {styles} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});
    const breadcrumb = useRef<HTMLOListElement>(null);
    const input = useRef<InputRef>(null);
    const displayPath = `/${path}`;
    const [editing, setEditing] = useState(false);
    const [draft, setDraft] = useState(displayPath);
    const breadcrumbs = useMemo(() => {
        const items = [{path: "", label: "/"}];
        let start = 0;
        for (let index = 0; index < path.length; index++) {
            if (path[index] !== "/") continue;
            items.push({path: path.slice(0, index + 1), label: path.slice(start, index) || "/"});
            start = index + 1;
        }
        if (start < path.length) items.push({path, label: path.slice(start)});
        return items;
    }, [path]);

    useEffect(() => { setEditing(false); setDraft(displayPath); }, [displayPath, disabled]);
    useLayoutEffect(() => {
        if (editing) { input.current?.focus(); input.current?.select(); }
    }, [editing]);
    useLayoutEffect(() => {
        const element = breadcrumb.current;
        if (editing || !element) return;
        const reveal = () => { element.scrollLeft = element.scrollWidth; };
        reveal();
        if (typeof ResizeObserver === "undefined") return;
        const observer = new ResizeObserver(reveal);
        observer.observe(element);
        return () => observer.disconnect();
    }, [breadcrumbs, editing]);

    const submit = useCallback(() => {
        if (disabled) return;
        setEditing(false);
        // 首个斜杠表示存储桶根目录，其余字符保留为实际对象前缀。
        const value = draft.startsWith("/") ? draft.slice(1) : draft;
        onNavigate(value && !value.endsWith("/") ? `${value}/` : value);
    }, [disabled, draft, onNavigate]);

    return <nav className={styles.pathBar} aria-label={`当前对象目录：${displayPath}`}>
        <Button className="path-back" type="text" icon={<Icon type="LeftOutlined"/>}
            aria-label="返回上一级目录" disabled={disabled || !path}
            onClick={() => onNavigate(breadcrumbs[breadcrumbs.length - 2]?.path || "")}/>
        {editing ? <Input ref={input} className="path-input" value={draft}
            placeholder="/ 表示存储桶根目录" aria-label="编辑对象目录"
            onChange={(event) => setDraft(event.target.value)} onBlur={() => setEditing(false)}
            onKeyDown={(event) => {
                if (event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229) return;
                if (event.key === "Enter") { event.preventDefault(); submit(); }
                if (event.key === "Escape") { event.preventDefault(); setEditing(false); }
            }}/> : <ol ref={breadcrumb} className="path-breadcrumb">
            {breadcrumbs.map((item, index) => {
                const current = index === breadcrumbs.length - 1;
                return <li key={item.path}>
                    {index > 0 && <Icon type="RightOutlined" className="path-separator" aria-hidden="true"/>}
                    <button className={current ? "path-current" : "path-segment"} type="button" disabled={disabled}
                        aria-current={current ? "page" : undefined} title={current ? "点击编辑路径" : `/${item.path}`}
                        aria-label={current ? "编辑当前对象目录" : `前往 /${item.path}`}
                        onClick={() => {
                            if (current) { setDraft(displayPath); setEditing(true); }
                            else onNavigate(item.path);
                        }}><span>{item.label}</span></button>
                </li>;
            })}
        </ol>}
    </nav>;
};

export default memo(OssPathBar);
