import {createStyles} from "antd-style";

/*
 * Database icons: https://github.com/devicons/devicon
 * The MIT License (MIT)
 *
 * Copyright (c) 2015 konpa
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */
const mysqlIcon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128"><path fill="#00618A" d="M117.688 98.242c-6.973-.191-12.297.461-16.852 2.379-1.293.547-3.355.559-3.566 2.18.711.746.82 1.859 1.387 2.777 1.086 1.754 2.922 4.113 4.559 5.352 1.789 1.348 3.633 2.793 5.551 3.961 3.414 2.082 7.223 3.27 10.504 5.352 1.938 1.23 3.859 2.777 5.75 4.164.934.684 1.563 1.75 2.773 2.18v-.195c-.637-.812-.801-1.93-1.387-2.777l-2.578-2.578c-2.52-3.344-5.719-6.281-9.117-8.719-2.711-1.949-8.781-4.578-9.91-7.73l-.199-.199c1.922-.219 4.172-.914 5.949-1.391 2.98-.797 5.645-.59 8.719-1.387l4.164-1.187v-.793c-1.555-1.594-2.664-3.707-4.359-5.152-4.441-3.781-9.285-7.555-14.273-10.703-2.766-1.746-6.184-2.883-9.117-4.363-.988-.496-2.719-.758-3.371-1.586-1.539-1.961-2.379-4.449-3.566-6.738-2.488-4.793-4.93-10.023-7.137-15.066-1.504-3.437-2.484-6.828-4.359-9.91-9-14.797-18.687-23.73-33.695-32.508-3.195-1.867-7.039-2.605-11.102-3.57l-6.543-.395c-1.332-.555-2.715-2.184-3.965-2.977C16.977 3.52 4.223-3.312.539 5.672-1.785 11.34 4.016 16.871 6.09 19.746c1.457 2.012 3.32 4.273 4.359 6.539.688 1.492.805 2.984 1.391 4.559 1.438 3.883 2.695 8.109 4.559 11.695.941 1.816 1.98 3.727 3.172 5.352.727.996 1.98 1.438 2.18 2.973-1.227 1.715-1.297 4.375-1.984 6.543-3.098 9.77-1.926 21.91 2.578 29.137 1.383 2.223 4.641 6.98 9.117 5.156 3.918-1.598 3.043-6.539 4.164-10.902.254-.988.098-1.715.594-2.379v.199l3.57 7.133c2.641 4.254 7.324 8.699 11.297 11.699 2.059 1.555 3.68 4.242 6.344 5.152v-.199h-.199c-.516-.805-1.324-1.137-1.98-1.781-1.551-1.523-3.277-3.414-4.559-5.156-3.613-4.902-6.805-10.27-9.711-15.855-1.391-2.668-2.598-5.609-3.77-8.324-.453-1.047-.445-2.633-1.387-3.172-1.281 1.988-3.172 3.598-4.164 5.945-1.582 3.754-1.789 8.336-2.375 13.082-.348.125-.195.039-.398.199-2.762-.668-3.73-3.508-4.758-5.949-2.594-6.164-3.078-16.09-.793-23.191.59-1.836 3.262-7.617 2.18-9.316-.516-1.691-2.219-2.672-3.172-3.965-1.18-1.598-2.355-3.703-3.172-5.551-2.125-4.805-3.113-10.203-5.352-15.062-1.07-2.324-2.875-4.676-4.359-6.738-1.645-2.289-3.484-3.977-4.758-6.742-.453-.984-1.066-2.559-.398-3.566.215-.684.516-.969 1.191-1.191 1.148-.887 4.352.297 5.547.793 3.18 1.32 5.832 2.578 8.527 4.363 1.289.855 2.598 2.512 4.16 2.973h1.785c2.789.641 5.914.195 8.523.988 4.609 1.402 8.738 3.582 12.488 5.949 11.422 7.215 20.766 17.48 27.156 29.734 1.027 1.973 1.473 3.852 2.379 5.945 1.824 4.219 4.125 8.559 5.941 12.688 1.816 4.113 3.582 8.27 6.148 11.695 1.348 1.801 6.551 2.766 8.918 3.766 1.66.699 4.379 1.43 5.949 2.379 3 1.809 5.906 3.965 8.723 5.945 1.402.992 5.73 3.168 5.945 4.957zm-88.605-75.52c-1.453-.027-2.48.156-3.566.395v.199h.195c.695 1.422 1.918 2.34 2.777 3.566l1.98 4.164.199-.195c1.227-.867 1.789-2.25 1.781-4.363-.492-.52-.562-1.164-.992-1.785-.562-.824-1.66-1.289-2.375-1.98zm0 0"/></svg>`;

export default createStyles(({css, token}) => ({
    databaseIcon: css`
        display: inline-block; flex: none; width: ${token.fontSize + 2}px; height: ${token.fontSize + 2}px; background: center / contain no-repeat;
        background-image: url("data:image/svg+xml;charset=utf-8,${encodeURIComponent(mysqlIcon)}");
    `,
    panel: css`
        && {
            min-width: 0;
            font-size: ${token.fontSize}px;
            line-height: 1.6;
        }
    `,
    form: css`
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form-item:last-child { margin-bottom: 0; }
        .ant-form-item-extra { font-size: ${token.fontSizeSM}px; }
    `,
    loading: css`
        display: flex;
        justify-content: center;
        align-items: center;
    `,
    modal: css`
        .ant-modal-body { min-width: 0; }
    `,
    codeEditor: css`
        border: 1px solid ${token.colorBorderSecondary};
        border-radius: ${token.borderRadius}px;

        &:not(textarea) { overflow: hidden; }
        .ace_editor { border-radius: ${token.borderRadius}px !important; }
    `,
    objectDefinition: css`
        display: flex; flex-direction: column; gap: ${token.marginSM}px; min-width: 0;
        .object-definition-toolbar { display: flex; align-items: center; justify-content: space-between; gap: ${token.marginSM}px; }
        .object-definition-context { display: flex; align-items: baseline; gap: ${token.marginXS}px; min-width: 0; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .object-definition-context > span:not(:nth-child(2)) { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .object-definition-context > span:nth-child(2) { flex: none; color: ${token.colorTextQuaternary}; }
        .object-definition-context > span:last-child { color: ${token.colorText}; }
        .object-definition-toolbar > .ant-btn { flex: none; font-size: ${token.fontSizeSM}px; }
        pre { margin: 0; padding: ${token.padding}px; border-radius: ${token.borderRadius}px; background: ${token.colorFillQuaternary}; color: ${token.colorText}; font-family: ${token.fontFamilyCode}; font-size: ${token.fontSizeSM}px; line-height: 1.8; white-space: pre-wrap; overflow-wrap: anywhere; tab-size: 4; }
    `,
    editor: css`
        .ant-modal-body { font-size: ${token.fontSize}px; line-height: 1.6; }
        .ant-form-item-label > label, .ant-input, .ant-input-number, .ant-select, .ant-radio-wrapper, .ant-checkbox-wrapper, .ant-collapse-header { font-size: ${token.fontSize}px; }
        .ant-form-item-explain, .ant-form-item-extra { font-size: ${token.fontSizeSM}px; }
        .ant-form-item { margin-bottom: ${token.margin}px; }
        .ant-form-item:last-child { margin-bottom: 0; }
        .ant-form-item-control, .ant-form-item-control-input-content { min-width: 0; }
        .ant-alert, .ant-typography { overflow-wrap: anywhere; }
        .ant-form-item-extra { font-size: ${token.fontSizeSM}px; }
        .transfer-result { display: flex; flex-wrap: wrap; gap: 8px 18px; padding: 12px; margin-top: 12px; background: ${token.colorFillQuaternary}; border-radius: ${token.borderRadius}px; color: ${token.colorTextSecondary}; font-size: ${token.fontSizeSM}px; }
        .transfer-result b { color: ${token.colorText}; font-weight: 500; margin-left: 5px; }
        .transfer-result > .ant-typography { width: 100%; font-size: inherit; overflow-wrap: anywhere; }
    `,
    workspace: css`
        &, .ant-modal-body { font-size: ${token.fontSize}px; line-height: 1.6; }
        .ant-btn, .ant-btn-sm, .ant-input, .ant-input-number, .ant-select, .ant-form-item-label > label, .ant-checkbox-wrapper, .ant-radio-wrapper, .ant-tabs-tab { font-size: ${token.fontSize}px; }
        .ant-form-item-explain, .ant-form-item-extra { font-size: ${token.fontSizeSM}px; }
        .ant-form-item-control, .ant-form-item-control-input-content, .ant-tabs-content-holder, .ant-tabs-tabpane { min-width: 0; }
        .ant-collapse-title, .ant-alert-description, .ant-empty-description { font-size: ${token.fontSize}px; overflow-wrap: anywhere; }
        .ant-btn > span { line-height: 1.4; }
        .ant-space { max-width: 100%; }
        .ant-modal-body { max-height: calc(100dvh - 112px); overflow-y: auto; overscroll-behavior: contain; }
        .client-content { min-width: 0; }
        .client-content .ant-tabs-nav { margin-bottom: 14px; }
        .client-content .ant-tabs-tab { font-size: ${token.fontSize}px; padding-block: 8px; }
        .client-content .ant-tabs-tab + .ant-tabs-tab { margin-left: 20px; }
        .client-content .ant-table, .client-content .ant-descriptions, .client-content .ant-alert, .client-content .ant-empty { font-size: ${token.fontSize}px; }
        .client-toolbar { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
        .client-toolbar { margin-bottom: 12px; }
        .client-toolbar > .ant-input-affix-wrapper, .client-toolbar > .ant-input { flex: 1; min-width: 130px; }
        .client-toolbar-spacer { flex: 1; }
        .client-query { display: flex; flex-direction: column; gap: 12px; min-width: 0; }
        .client-query .client-toolbar { margin-bottom: 0; }
        .client-cell { display: block; max-width: 270px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: ${token.fontSize}px; }
        .client-column-title { display: flex; align-items: center; gap: 6px; min-width: 0; max-width: 100%; }
        .client-column-title > .ant-typography { flex: 0 1 auto; min-width: 0; max-width: 100%; font: inherit; color: inherit; }
        .client-column-comment { display: inline-flex; flex: none; color: ${token.colorTextTertiary}; cursor: help; }
        .client-column-comment:focus-visible { outline: 2px solid ${token.colorPrimaryBorder}; outline-offset: 2px; border-radius: 2px; }
        .client-row-number { color: ${token.colorTextTertiary}; font-size: ${token.fontSizeSM}px; font-variant-numeric: tabular-nums; }
        .client-null { color: ${token.colorTextTertiary}; font-style: italic; }
        .client-footer { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; padding-top: 12px; font-size: ${token.fontSizeSM}px; }
        .client-footer .ant-typography, .client-hint { font-size: ${token.fontSizeSM}px; }
        .client-hint { display: block; margin-bottom: 10px; line-height: 1.65; }
        .client-result-status { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 10px; font-size: ${token.fontSizeSM}px; color: ${token.colorTextSecondary}; }
        .client-empty { min-height: 320px; display: grid; align-content: center; justify-items: center; gap: 12px; }
        .maintenance-panel { min-width: 0; }
        .maintenance-list { border: 1px solid ${token.colorBorderSecondary}; border-radius: ${token.borderRadius}px; overflow: hidden; }
        .maintenance-row { display: flex; align-items: center; justify-content: space-between; gap: 14px; padding: 14px; }
        .maintenance-row + .maintenance-row { border-top: 1px solid ${token.colorBorderSecondary}; }
        .maintenance-row > div { min-width: 0; display: flex; flex-direction: column; gap: 4px; }
        .maintenance-tool-title { color: ${token.colorText}; font-size: ${token.fontSize}px; font-weight: 500; }
        .maintenance-row .ant-typography { font-size: ${token.fontSizeSM}px; overflow-wrap: anywhere; }
        .maintenance-row > .ant-btn { flex: none; }
        @media (max-width: 767px) {
            .ant-modal-body { max-height: calc(100dvh - 94px); }
            .client-toolbar { gap: 6px; }
            .client-toolbar > .ant-input-affix-wrapper, .client-toolbar > .ant-input { min-width: 100px; }
            .client-content .ant-tabs-tab + .ant-tabs-tab { margin-left: 16px; }
            .client-footer { gap: 8px; }
            .client-footer .ant-space { flex-wrap: wrap; }
            .client-empty { min-height: 220px; }
            .maintenance-row { padding: 12px; gap: 10px; }
            .maintenance-row > .ant-btn { padding-inline: 10px; }
        }
    `,
}));
