import {forwardRef, useCallback, useEffect, useRef, useState} from "react";
import type {DropdownProps} from "antd";
import {createStyles} from "antd-style";
import {useTheme} from "sinking-antd";
import Dropdown from "@/pages/components/ui/stable-dropdown";

const useStyles = createStyles(({css, token}, {compact}: {compact: boolean}) => ({
    menu: css`
        && {
            min-width: ${compact ? 94 : 104}px;
            max-width: calc(100vw - 24px);
            padding: 0;
            border-radius: ${token.borderRadius}px;
        }

        && .ant-dropdown-menu {
            padding: 3px !important;
            border-radius: ${token.borderRadius}px !important;
            background: ${token.colorBgElevated};
            box-shadow: ${token.boxShadowSecondary};
        }

        && .ant-dropdown-menu-item {
            min-height: ${compact ? 28 : 30}px;
            padding: 0 ${compact ? 7 : 8}px !important;
            border-radius: ${token.borderRadiusSM}px !important;
            color: ${token.colorTextSecondary};
            font-size: ${compact ? 11 : 12}px;
        }

        && .ant-dropdown-menu-item-danger {
            color: ${token.colorError};
        }

        && .ant-dropdown-menu-item-disabled {
            color: ${token.colorTextDisabled};
        }
    `,
}));

const ScrollDropdown = forwardRef<HTMLElement, DropdownProps>(({open, onOpenChange, ...props}, ref) => {
    const anchor = useRef<HTMLElement | null>(null);
    const setAnchor = useCallback((node: HTMLElement | null) => {
        anchor.current = node;
        if (typeof ref === "function") return ref(node);
        if (ref) ref.current = node;
    }, [ref]);
    const [localOpen, setLocalOpen] = useState(false);
    const visible = open ?? localOpen;
    const change = useCallback<NonNullable<DropdownProps["onOpenChange"]>>((next, info) => {
        if (open === undefined) setLocalOpen(next);
        onOpenChange?.(next, info);
    }, [open, onOpenChange]);
    useEffect(() => {
        if (!visible) return;
        const onScroll = ({target}: Event) => {
            if (target === window || (target instanceof Node && target.contains(anchor.current))) {
                change(false, {source: "trigger"});
            }
        };
        window.addEventListener("scroll", onScroll, {capture: true, passive: true});
        return () => window.removeEventListener("scroll", onScroll, true);
    }, [visible, change]);
    return <Dropdown {...props} ref={setAnchor} open={visible} onOpenChange={change}/>;
});

// 统一操作菜单外观，列表内的按钮可选在所属滚动区移动时收起。
const ActionDropdown = forwardRef<HTMLElement, DropdownProps & {closeOnScroll?: boolean}>(({rootClassName, closeOnScroll,
    trigger = ["click"], placement = "bottom", arrow = {pointAtCenter: true}, ...props}, ref) => {
    const theme = useTheme();
    const {styles, cx} = useStyles({compact: Boolean(theme?.isCompactTheme?.())});

    const Component = closeOnScroll ? ScrollDropdown : Dropdown;
    return <Component {...props} ref={ref}
        trigger={trigger} placement={placement} arrow={arrow}
        rootClassName={cx(styles.menu, rootClassName)}/>;
});

ActionDropdown.displayName = "ActionDropdown";

export default ActionDropdown;
