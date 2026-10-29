import React, {forwardRef} from "react";
import {Dropdown as AntDropdown} from "antd";
import type {DropdownProps} from "antd";

const StableDropdown = forwardRef<HTMLElement, DropdownProps>((props, ref) => (
    <AntDropdown
        {...props}
        ref={ref}/>
));

StableDropdown.displayName = "StableDropdown";

export default StableDropdown;
